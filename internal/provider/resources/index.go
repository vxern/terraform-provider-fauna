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

func ResourceIndex() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIndexCreate,
		ReadContext:   resourceIndexRead,
		UpdateContext: resourceIndexUpdate,
		DeleteContext: resourceIndexDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: fmt.Sprintf("The name of this index. Cannot be one of: %s.", strings.Join(BlacklistedResourceNames, ", ")),
				Type:        schema.TypeString,
				Required:    true,
			},
			"data": {
				Description: "Developer-defined metadata for this index.",
				Type:        schema.TypeMap,
				Optional:    true,
			},
			"source": {
				Description: "The source collection.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"terms": {
				Description: "The document fields whose values can be matched for the search term.",
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field": {
							Description: "The field names required to access a specific field nested within the document structure.",
							Required:    true,
							Type:        schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"values": {
				Description: "The document fields whose values are to be returned.",
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field": {
							Description: "The field names required to access a specific field nested within the document structure.",
							Required:    true,
							Type:        schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"reverse": {
							Description: "Whether this field's value should sort reversed.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
						},
					},
				},
			},
			"unique": {
				Description: "Whether to maintain a `unique` constraint on combined `terms` and `values`.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"serialized": {
				Description: "Whether to serialise concurrent reads and writes to this resource.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"ttl": {
				Description: "A timestamp of when this index is to be removed.",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     nil,
			},
			"ts": {
				Description: "A timestamp of when this index was created.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
		},
	}
}

func synchroniseIndexResourceData(res f.Value, data *schema.ResourceData) error {
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

	if source, ok := GetProperty(obj, "source", f.RefV{}); ok {
		var source_ f.RefV
		source.Get(&source_)

		data.Set("source", source_.ID)
	}

	if terms, ok := GetProperty(obj, "terms", []map[string]any{}); ok {
		data.Set("terms", terms)
	}

	if values, ok := GetProperty(obj, "values", []map[string]any{}); ok {
		data.Set("values", values)
	}

	if unique, ok := GetProperty(obj, "unique", false); ok {
		data.Set("unique", unique)
	}

	if serialized, ok := GetProperty(obj, "serialized", false); ok {
		data.Set("serialized", serialized)
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

func resourceIndexCreate(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*f.FaunaClient)

	name := data.Get("name").(string)

	if err := CheckNameNotBlacklisted(name, "index"); err != nil {
		return diag.FromErr(err)
	}

	res, err := conn.Query(f.CreateIndex(f.Obj{
		"name":       name,
		"data":       data.Get("data"),
		"source":     f.Collection(data.Get("source")),
		"terms":      data.Get("terms"),
		"values":     data.Get("values"),
		"unique":     data.Get("unique"),
		"serialized": data.Get("serialized"),
		"ttl":        data.Get("ttl"),
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	if err := synchroniseIndexResourceData(res, data); err != nil {
		return diag.FromErr(err)
	}

	return resourceIndexRead(ctx, data, meta)
}

func resourceIndexRead(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*f.FaunaClient)

	res, err := conn.Query(f.Get(f.Index(data.Get("name"))))
	if err != nil {
		return diag.FromErr(err)
	}

	if err := synchroniseIndexResourceData(res, data); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

var indexPropertiesToCheck = []string{"name", "data", "terms", "values", "unique", "serialized", "ttl", "ts"}

func resourceIndexUpdate(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*f.FaunaClient)

	object := make(map[string]any)
	for _, property := range indexPropertiesToCheck {
		if !data.HasChange(property) {
			continue
		}

		object[property] = data.Get(property)
	}

	if len(object) != 0 {
		_, err := conn.Query(f.Update(f.Index(data.Get("name")), object))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceIndexRead(ctx, data, meta)
}

func resourceIndexDelete(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*f.FaunaClient)

	_, err := conn.Query(f.Delete(f.Index(data.Get("name"))))
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId("")

	return diags
}
