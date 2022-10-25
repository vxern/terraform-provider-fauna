package resources_test

import (
	"fmt"
	"strings"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	f "github.com/fauna/faunadb-go/v5/faunadb"

	acctest "github.com/linguition/terraform-provider-fauna/internal/acctest"
)

func TestAccCollection(t *testing.T) {
	rColName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.TestAccPreCheck(t) },
		Providers:    acctest.TestAccProviders,
		CheckDestroy: testAccCheckCollectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionConfiguration(rColName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionExists("fauna_collection.collection"),
				),
			},
			{
				Config: testAccCollectionConfiguration_addedProperties(rColName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionExists("fauna_collection.collection"),
					resource.TestCheckResourceAttr("fauna_collection.collection", "history_days", "30"),
					resource.TestCheckResourceAttr("fauna_collection.collection", "ttl", "7"),
					resource.TestCheckResourceAttr("fauna_collection.collection", "ttl_days", "14"),
				),
			},
		},
	})
}

func testAccCollectionConfiguration(rColName string) string {
	return fmt.Sprintf(`
resource "fauna_collection" "collection" {
	name = "%s"
}`, rColName)
}

func testAccCollectionConfiguration_addedProperties(rColName string) string {
	return fmt.Sprintf(`
resource "fauna_collection" "collection" {
	name         = "%s"
	history_days = 30
	ttl = 7
	ttl_days = 14
}`, rColName)
}

func testAccCheckCollectionExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var name string

		{
			res, ok := s.RootModule().Resources[resourceName]
			if !ok {
				return fmt.Errorf("Not found: %s", resourceName)
			}

			name = res.Primary.Attributes["name"]
			if name == "" {
				return fmt.Errorf("Collection ID is not set.")
			}
		}

		client := acctest.TestAccProvider.Meta().(*f.FaunaClient)

		if _, err := client.Query(f.Get(f.Collection(name))); err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckCollectionDestroy(s *terraform.State) error {
	client := acctest.TestAccProvider.Meta().(*f.FaunaClient)

	for _, res := range s.RootModule().Resources {
		if res.Type != "fauna_collection" {
			continue
		}

		name := res.Primary.Attributes["name"]

		_, err := client.Query(f.Get(f.Collection(name)))
		if err == nil {
			return fmt.Errorf("Collection '%s' still exists.", name)
		}

		if !strings.Contains(err.Error(), "Ref refers to undefined collection") {
			return err
		}
	}

	return nil
}
