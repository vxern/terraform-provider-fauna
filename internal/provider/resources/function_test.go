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

func TestAccFunction(t *testing.T) {
	rColName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.TestAccPreCheck(t) },
		Providers:    acctest.TestAccProviders,
		CheckDestroy: testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfiguration(rColName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists("fauna_function.function"),
				),
			},
			{
				Config: testAccFunctionConfiguration_addedProperties(rColName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists("fauna_function.function"),
					resource.TestCheckResourceAttr("fauna_function.function", "data.sample_key", "sample_value"),
					resource.TestCheckResourceAttr("fauna_function.function", "data.sample_key_2", "false"),
					resource.TestCheckResourceAttr("fauna_function.function", "data.sample_key_3", "65"),
					resource.TestCheckResourceAttr("fauna_function.function", "ttl", "7"),
				),
			},
		},
	})
}

func testAccFunctionConfiguration(rColName string) string {
	return fmt.Sprintf(`
resource "fauna_function" "function" {
	name = "%s"
	body = "Query(Lambda(\"X\", Paginate(Collections())))"
}`, rColName)
}

func testAccFunctionConfiguration_addedProperties(rColName string) string {
	return fmt.Sprintf(`
resource "fauna_function" "function" {
	name         = "%s"
	data = {
		sample_key = "sample_value"
		sample_key_2 = false
		sample_key_3 = 65
	}
	body = "Query(Lambda(\"X\", Paginate(Collections())))"
	ttl = 7
}`, rColName)
}

func testAccCheckFunctionExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var name string

		{
			res, ok := s.RootModule().Resources[resourceName]
			if !ok {
				return fmt.Errorf("Not found: %s", resourceName)
			}

			name = res.Primary.Attributes["name"]
			if name == "" {
				return fmt.Errorf("Function ID is not set.")
			}
		}

		client := acctest.TestAccProvider.Meta().(*f.FaunaClient)

		if _, err := client.Query(f.Get(f.Function(name))); err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckFunctionDestroy(s *terraform.State) error {
	client := acctest.TestAccProvider.Meta().(*f.FaunaClient)

	for _, res := range s.RootModule().Resources {
		if res.Type != "fauna_function" {
			continue
		}

		name := res.Primary.Attributes["name"]

		_, err := client.Query(f.Get(f.Function(name)))
		if err == nil {
			return fmt.Errorf("Function '%s' still exists.", name)
		}

		if !strings.Contains(err.Error(), "Ref refers to undefined function") {
			return err
		}
	}

	return nil
}
