package resources_test

import (
	"fmt"
	"strings"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	f "github.com/fauna/faunadb-go/v5/faunadb"

	acctest "github.com/wordcollector/terraform-provider-fauna/internal/acctest"
)

func TestAccIndex(t *testing.T) {
	rColName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	rIndexName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.TestAccPreCheck(t) },
		Providers:    acctest.TestAccProviders,
		CheckDestroy: testAccCheckIndexDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIndexConfiguration(rColName, rIndexName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexExists("fauna_index.index"),
					resource.TestCheckResourceAttr("fauna_index.index", "terms.0.field.1", "sample_property"),
					resource.TestCheckResourceAttr("fauna_index.index", "terms.1.field.1", "different_sample_property"),
					resource.TestCheckResourceAttr("fauna_index.index", "values.0.field.1", "sample_property"),
					resource.TestCheckResourceAttr("fauna_index.index", "values.1.field.1", "different_sample_property"),
					resource.TestCheckResourceAttr("fauna_index.index", "serialized", "true"),
				),
			},
			{
				Config: testAccIndexConfiguration_addedProperties(rColName, rIndexName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionExists("fauna_collection.collection"),
					testAccCheckIndexExists("fauna_index.index"),
					resource.TestCheckResourceAttr("fauna_index.index", "unique", "true"),
					resource.TestCheckResourceAttr("fauna_index.index", "serialized", "true"),
					resource.TestCheckResourceAttr("fauna_index.index", "ttl", "7"),
				),
			},
		},
	})
}

func testAccIndexConfiguration(rColName string, rIndexName string) string {
	return fmt.Sprintf(`
resource "fauna_collection" "collection" {
	name = "%[1]s"
}

resource "fauna_index" "index" {
	depends_on = [fauna_collection.collection]

	name   = "%[2]s"
	source = "%[1]s"
	terms {
		field = ["data", "sample_property"]
	}
	terms {
		field = ["data", "different_sample_property"]
	}
	values {
		field = ["data", "sample_property"]
	}
	values {
		field = ["data", "different_sample_property"]
	}
}
`, rColName, rIndexName)
}

func testAccIndexConfiguration_addedProperties(rColName string, rIndexName string) string {
	return fmt.Sprintf(`
resource "fauna_collection" "collection" {
	name = "%[1]s"
}

resource "fauna_index" "index" {
	depends_on = [fauna_collection.collection]

	name   = "%[2]s"
	source = "%[1]s"
	terms {
		field = ["data", "sample_property"]
	}
	terms {
		field = ["data", "different_sample_property"]
	}
	values {
		field = ["data", "sample_property"]
	}
	values {
		field = ["data", "different_sample_property"]
	}
	unique = true
	serialized = true
	ttl = 7
}
`, rColName, rIndexName)
}

func testAccCheckIndexExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var id string

		{
			res, ok := s.RootModule().Resources[resourceName]
			if !ok {
				return fmt.Errorf("Not found: %s", resourceName)
			}

			id = res.Primary.Attributes["name"]
			if id == "" {
				return fmt.Errorf("Index ID is not set.")
			}
		}

		client := acctest.TestAccProvider.Meta().(*f.FaunaClient)

		_, err := client.Query(f.Get(f.Index(id)))
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckIndexDestroy(s *terraform.State) error {
	client := acctest.TestAccProvider.Meta().(*f.FaunaClient)

	for _, res := range s.RootModule().Resources {
		if res.Type != "fauna_index" {
			continue
		}

		name := res.Primary.Attributes["name"]

		_, err := client.Query(f.Get(f.Index(name)))
		if err == nil {
			return fmt.Errorf("Index '%s' still exists.", name)
		}

		if !strings.Contains(err.Error(), "Ref refers to undefined index") {
			return err
		}
	}

	return nil
}
