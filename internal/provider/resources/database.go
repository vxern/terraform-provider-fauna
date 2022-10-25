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

func ResourceDatabase() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDatabaseCreate,
		ReadContext:   resourceDatabaseRead,
		UpdateContext: resourceDatabaseUpdate,
		DeleteContext: resourceDatabaseDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: fmt.Sprintf("The name of this database. Cannot be one of: %s.", strings.Join(BlacklistedResourceNames, ", ")),
				Type:        schema.TypeString,
				Required:    true,
			},
			"data": {
				Description: "Developer-defined metadata for this database.",
				Type:        schema.TypeMap,
				Optional:    true,
			},
			"ttl": {
				Description: "A timestamp of when this database is to be removed.",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     nil,
			},
			"global_id": {
				Description: "A globally unique identifier for this database.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"ts": {
				Description: "A timestamp of when this database was created.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
		},
	}
}

func synchroniseDatabaseResourceData(res f.Value, data *schema.ResourceData) error {
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

	if ttl, ok := GetProperty(obj, "ttl", 0); ok {
		data.Set("ttl", ttl)
	}

	if globalId, ok := GetProperty(obj, "global_id", ""); ok {
		data.Set("global_id", globalId)
	}

	if ts, ok := GetProperty[int64](obj, "ts", 0); ok {
		data.SetId(strconv.FormatInt(ts, 10))
		data.Set("ts", ts)
	}

	return nil
}

func resourceDatabaseCreate(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*f.FaunaClient)

	name := data.Get("name").(string)

	if err := CheckNameNotBlacklisted(name, "database"); err != nil {
		return diag.FromErr(err)
	}

	res, err := conn.Query(f.CreateDatabase(f.Obj{
		"name": name,
		"data": data.Get("data"),
		"ttl":  data.Get("ttl"),
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	if err := synchroniseDatabaseResourceData(res, data); err != nil {
		return diag.FromErr(err)
	}

	return resourceDatabaseRead(ctx, data, meta)
}

func resourceDatabaseRead(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*f.FaunaClient)

	res, err := conn.Query(f.Get(f.Database(data.Get("name"))))
	if err != nil {
		return diag.FromErr(err)
	}

	if err := synchroniseDatabaseResourceData(res, data); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

var databasePropertiesToCheck = []string{"name", "data", "ttl"}

func resourceDatabaseUpdate(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*f.FaunaClient)

	object := make(map[string]any)
	for _, property := range databasePropertiesToCheck {
		if !data.HasChange(property) {
			continue
		}

		object[property] = data.Get(property)
	}

	if data.HasChange("name") {
		data.SetId(data.Get("name").(string))
	}

	if len(object) != 0 {
		_, err := conn.Query(f.Update(f.Database(data.Get("name")), object))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceDatabaseRead(ctx, data, meta)
}

func resourceDatabaseDelete(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*f.FaunaClient)

	_, err := conn.Query(f.Delete(f.Database(data.Get("name"))))
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId("")

	return diags
}
