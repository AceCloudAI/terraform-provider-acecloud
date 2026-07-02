package api_key_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/AceCloudAI/terraform-provider-acecloud/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccAPIKey_basic(t *testing.T) {
	rName := acctest.RandomName("apikey")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckAPIKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAPIKeyExists("acecloud_api_key.test"),
					resource.TestCheckResourceAttrSet("acecloud_api_key.test", "id"),
					resource.TestCheckResourceAttr("acecloud_api_key.test", "service_name", rName),
					resource.TestCheckResourceAttrSet("acecloud_api_key.test", "secret"),
					resource.TestCheckResourceAttr("acecloud_api_key.test", "enabled", "true"),
					resource.TestCheckResourceAttrSet("acecloud_api_key.test", "created_at"),
				),
			},
		},
	})
}

func TestAccAPIKey_update(t *testing.T) {
	rName := acctest.RandomName("apikey")
	rNameUpdated := rName + "-upd"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckAPIKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("acecloud_api_key.test", "service_name", rName),
					resource.TestCheckResourceAttr("acecloud_api_key.test", "enabled", "true"),
				),
			},
			{
				Config: testAccAPIKeyConfig_updated(rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("acecloud_api_key.test", "service_name", rNameUpdated),
					resource.TestCheckResourceAttr("acecloud_api_key.test", "description", "updated description"),
					resource.TestCheckResourceAttr("acecloud_api_key.test", "enabled", "false"),
				),
			},
		},
	})
}

func TestAccAPIKey_import(t *testing.T) {
	rName := acctest.RandomName("apikey")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckAPIKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyConfig_basic(rName),
			},
			{
				ResourceName:            "acecloud_api_key.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret"},
			},
		},
	})
}

func TestAccAPIKey_description(t *testing.T) {
	rName := acctest.RandomName("apikey")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckAPIKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyConfig_withDescription(rName, "initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("acecloud_api_key.test", "description", "initial description"),
				),
			},
			{
				Config: testAccAPIKeyConfig_withDescription(rName, "changed description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("acecloud_api_key.test", "description", "changed description"),
				),
			},
		},
	})
}

func testAccCheckAPIKeyExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID is empty")
		}

		c := acctest.TestClient()
		_, err := c.Get(context.Background(), fmt.Sprintf("/iam/api-keys/%s", rs.Primary.ID), nil)
		if err != nil {
			return fmt.Errorf("API key %s not found via API: %w", rs.Primary.ID, err)
		}
		return nil
	}
}

func testAccCheckAPIKeyDestroy(s *terraform.State) error {
	c := acctest.TestClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "acecloud_api_key" {
			continue
		}
		_, err := c.Get(context.Background(), fmt.Sprintf("/iam/api-keys/%s", rs.Primary.ID), nil)
		if err == nil {
			return fmt.Errorf("API key %s still exists after destroy", rs.Primary.ID)
		}
	}
	return nil
}

func testAccAPIKeyConfig_basic(name string) string {
	return acctest.ProviderConfig() + fmt.Sprintf(`
resource "acecloud_api_key" "test" {
  service_name = %[1]q
}
`, name)
}

func testAccAPIKeyConfig_updated(name string) string {
	return acctest.ProviderConfig() + fmt.Sprintf(`
resource "acecloud_api_key" "test" {
  service_name = %[1]q
  description  = "updated description"
  enabled      = false
}
`, name)
}

func testAccAPIKeyConfig_withDescription(name, description string) string {
	return acctest.ProviderConfig() + fmt.Sprintf(`
resource "acecloud_api_key" "test" {
  service_name = %[1]q
  description  = %[2]q
}
`, name, description)
}
