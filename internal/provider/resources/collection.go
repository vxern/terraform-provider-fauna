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

func ResourceCollection() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCollectionCreate,
		ReadContext:   resourceCollectionRead,
		UpdateContext: resourceCollectionUpdate,
		DeleteContext: resourceCollectionDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: fmt.Sprintf("The name of this collection. Cannot be one of: %s.", strings.Join(BlacklistedResourceNames, ", ")),
				Type:        schema.TypeString,
				Required:    true,
			},
			"data": {
				Description: "Developer-defined metadata for this collection.",
				Type:        schema.TypeMap,
				Optional:    true,
			},
			"history_days": {
				Description: "The number of days that document history is to be retained for in this collection.",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
			},
			"ttl": {
				Description: "A timestamp of when this collection is to be removed.",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     nil,
			},
			"ttl_days": {
				Description: "The number of days documents are to be retained for in this collection.",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     nil,
			},
			"ts": {
				Description: "A timestamp of when this collection was created.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
		},
	}
}

func synchroniseCollectionResourceData(res f.Value, data *schema.ResourceData) error {
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

	if historyDays, ok := GetProperty(obj, "history_days", 0); ok {
		data.Set("history_days", historyDays)
	}

	if ttl, ok := GetProperty(obj, "ttl", 0); ok {
		data.Set("ttl", ttl)
	}

	if ttlDays, ok := GetProperty(obj, "ttl_days", 0); ok {
		data.Set("ttl_days", ttlDays)
	}

	if ts, ok := GetProperty[int64](obj, "ts", 0); ok {
		data.SetId(strconv.FormatInt(ts, 10))
		data.Set("ts", ts)
	}

	return nil
}

func resourceCollectionCreate(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*f.FaunaClient)

	name := data.Get("name").(string)

	if err := CheckNameNotBlacklisted(name, "collection"); err != nil {
		return diag.FromErr(err)
	}

	res, err := conn.Query(f.CreateCollection(f.Obj{
		"name":         name,
		"data":         data.Get("data"),
		"history_days": data.Get("history_days"),
		"ttl":          data.Get("ttl"),
		"ttl_days":     data.Get("ttl_days"),
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	if err := synchroniseCollectionResourceData(res, data); err != nil {
		return diag.FromErr(err)
	}

	return resourceCollectionRead(ctx, data, meta)
}

func resourceCollectionRead(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*f.FaunaClient)

	res, err := conn.Query(f.Get(f.Collection(data.Get("name"))))
	if err != nil {
		return diag.FromErr(err)
	}

	if err := synchroniseCollectionResourceData(res, data); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

var collectionPropertiesToCheck = []string{"name", "data", "history_days", "ttl", "ttl_days"}

func resourceCollectionUpdate(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*f.FaunaClient)

	object := make(map[string]any)
	for _, property := range collectionPropertiesToCheck {
		if !data.HasChange(property) {
			continue
		}

		object[property] = data.Get(property)
	}

	if len(object) != 0 {
		_, err := conn.Query(f.Update(f.Collection(data.Get("name")), object))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceCollectionRead(ctx, data, meta)
}

func resourceCollectionDelete(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*f.FaunaClient)

	_, err := conn.Query(f.Delete(f.Collection(data.Get("name"))))
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId("")

	return diags
}
