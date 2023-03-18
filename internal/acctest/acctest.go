package acctest

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/wordcollector/terraform-provider-fauna/internal/provider"
)

var TestAccProviders map[string]*schema.Provider
var TestAccProvider *schema.Provider

func init() {
	TestAccProvider = provider.Provider()
	TestAccProviders = map[string]*schema.Provider{
		"fauna": TestAccProvider,
	}
}

func TestAccPreCheck(t *testing.T) {
	if v := os.Getenv("FAUNA_SECRET"); v == "" {
		t.Fatal("'FAUNA_SECRET' must be set for acceptance tests.")
	}
}
