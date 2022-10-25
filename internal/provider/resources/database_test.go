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

func TestAccDatabase(t *testing.T) {
	rColName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.TestAccPreCheck(t) },
		Providers:    acctest.TestAccProviders,
		CheckDestroy: testAccCheckDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfiguration(rColName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists("fauna_database.database"),
				),
			},
			{
				Config: testAccDatabaseConfiguration_addedProperties(rColName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists("fauna_database.database"),
					resource.TestCheckResourceAttr("fauna_database.database", "data.sample_key", "sample_value"),
					resource.TestCheckResourceAttr("fauna_database.database", "data.sample_key_2", "false"),
					resource.TestCheckResourceAttr("fauna_database.database", "data.sample_key_3", "65"),
					resource.TestCheckResourceAttr("fauna_database.database", "ttl", "7"),
				),
			},
		},
	})
}

func testAccDatabaseConfiguration(rColName string) string {
	return fmt.Sprintf(`
resource "fauna_database" "database" {
	name = "%s"
}`, rColName)
}

func testAccDatabaseConfiguration_addedProperties(rColName string) string {
	return fmt.Sprintf(`
resource "fauna_database" "database" {
	name         = "%s"
	data = {
		sample_key = "sample_value"
		sample_key_2 = false
		sample_key_3 = 65
	}
	ttl = 7
}`, rColName)
}

func testAccCheckDatabaseExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var name string

		{
			res, ok := s.RootModule().Resources[resourceName]
			if !ok {
				return fmt.Errorf("Not found: %s", resourceName)
			}

			name = res.Primary.Attributes["name"]
			if name == "" {
				return fmt.Errorf("Database ID is not set.")
			}
		}

		client := acctest.TestAccProvider.Meta().(*f.FaunaClient)

		if _, err := client.Query(f.Get(f.Database(name))); err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckDatabaseDestroy(s *terraform.State) error {
	client := acctest.TestAccProvider.Meta().(*f.FaunaClient)

	for _, res := range s.RootModule().Resources {
		if res.Type != "fauna_database" {
			continue
		}

		name := res.Primary.Attributes["name"]

		_, err := client.Query(f.Get(f.Database(name)))
		if err == nil {
			return fmt.Errorf("Database '%s' still exists.", name)
		}

		if !strings.Contains(err.Error(), "Ref refers to undefined database") {
			return err
		}
	}

	return nil
}
