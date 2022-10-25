package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/linguition/terraform-provider-fauna/internal/provider"
)

func main() {
	opts := &plugin.ServeOpts{ProviderFunc: provider.Provider}

	plugin.Serve(opts)
}
