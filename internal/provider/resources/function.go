package resources

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	f "github.com/fauna/faunadb-go/v5/faunadb"
)

func ResourceFunction() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFunctionCreate,
		ReadContext:   resourceFunctionRead,
		UpdateContext: resourceFunctionUpdate,
		DeleteContext: resourceFunctionDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: fmt.Sprintf("The name of this function. Cannot be one of: %s.", strings.Join(BlacklistedResourceNames, ", ")),
				Type:        schema.TypeString,
				Required:    true,
			},
			"data": {
				Description: "Developer-defined metadata for this function.",
				Type:        schema.TypeMap,
				Optional:    true,
			},
			"body": {
				Description: "The FQL instructions to be executed.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"role": {
				Description: "The role to use when calling this user-defined function.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     nil,
			},
			"ttl": {
				Description: "A timestamp of when this function is to be removed.",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     nil,
			},
			"ts": {
				Description: "A timestamp of when this function was created.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
		},
	}
}

func synchroniseFunctionResourceData(res f.Value, data *schema.ResourceData) error {
	var obj f.ObjectV
	if err := res.Get(&obj); err != nil {
		return err
	}

	if name_, ok := GetProperty(obj, "name", ""); ok {
		data.Set("name", name_)
	}

	if data_, ok := GetProperty(obj, "data", map[string]any{}); ok {
		data.Set("data", data_)
	}

	if body, ok := GetProperty(obj, "body", ""); ok {
		data.Set("body", body)
	}

	if role, ok := GetProperty(obj, "role", ""); ok {
		data.Set("role", role)
	}

	if ttl, ok := GetProperty(obj, "ttl", 0); ok {
		data.Set("ttl", ttl)
	}

	if ts, ok := GetProperty[int64](obj, "ts", 0); ok {
		data.SetId(strconv.FormatInt(ts, 10))
		data.Set("ts", ts)
	}

	return nil
}

func resourceFunctionCreate(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*f.FaunaClient)

	name := data.Get("name").(string)

	if err := CheckNameNotBlacklisted(name, "function"); err != nil {
		return diag.FromErr(err)
	}

	obj := f.Obj{
		"name": name,
		"data": data.Get("data"),
		"body": data.Get("body"),
		"ttl":  data.Get("ttl"),
	}

	if role := data.Get("role"); role != "" {
		obj["role"] = f.Role(role)
	}

	res, err := conn.Query(f.CreateFunction(obj))
	if err != nil {
		return diag.FromErr(err)
	}

	if err := synchroniseFunctionResourceData(res, data); err != nil {
		return diag.FromErr(err)
	}

	return resourceFunctionRead(ctx, data, meta)
}

func resourceFunctionRead(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*f.FaunaClient)

	res, err := conn.Query(f.Get(f.Function(data.Get("name"))))
	if err != nil {
		return diag.FromErr(err)
	}

	if err := synchroniseFunctionResourceData(res, data); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

var functionPropertiesToCheck = []string{"name", "data", "body", "ttl"}

func resourceFunctionUpdate(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*f.FaunaClient)

	object := make(map[string]any)
	for _, property := range functionPropertiesToCheck {
		if !data.HasChange(property) {
			continue
		}

		object[property] = data.Get(property)
	}

	if data.HasChange("role") {
		object["role"] = f.Role(data.Get("role"))
	}

	if len(object) != 0 {
		_, err := conn.Query(f.Update(f.Function(data.Get("name")), object))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceFunctionRead(ctx, data, meta)
}

func resourceFunctionDelete(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*f.FaunaClient)

	_, err := conn.Query(f.Delete(f.Function(data.Get("name"))))
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId("")

	return diags
}
