package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	f "github.com/fauna/faunadb-go/v5/faunadb"

	resources "github.com/linguition/terraform-provider-fauna/internal/provider/resources"
)

func init() {
	schema.DescriptionKind = schema.StringMarkdown
}

func Provider() *schema.Provider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"secret": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("FAUNA_SECRET", schema.EnvDefaultFunc("FAUNA_KEY", schema.EnvDefaultFunc("FAUNA", nil))),
			},
			"endpoint": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"fauna_collection": resources.ResourceCollection(),
			"fauna_index":      resources.ResourceIndex(),
		},
	}

	provider.ConfigureContextFunc = configure(provider)

	return provider
}

func configure(provider *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
		var diags diag.Diagnostics

		secret := data.Get("secret").(string)

		var client *f.FaunaClient

		if endpoint := data.Get("endpoint").(string); endpoint != "" {
			client = f.NewFaunaClient(secret, f.Endpoint(endpoint))
		} else {
			client = f.NewFaunaClient(secret)
		}

		return client, diags
	}
}
