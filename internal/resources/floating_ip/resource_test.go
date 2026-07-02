package floating_ip_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/AceCloudAI/terraform-provider-acecloud/internal/acctest"
	"github.com/AceCloudAI/terraform-provider-acecloud/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccFloatingIP_basic(t *testing.T) {
	extNetID := acctest.ExternalNetworkID(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckFloatingIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFloatingIPConfig_basic(extNetID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFloatingIPExists("acecloud_floating_ip.test"),
					resource.TestCheckResourceAttrSet("acecloud_floating_ip.test", "id"),
					resource.TestCheckResourceAttrSet("acecloud_floating_ip.test", "floating_ip_address"),
					resource.TestCheckResourceAttr("acecloud_floating_ip.test", "floating_network_id", extNetID),
					resource.TestCheckResourceAttr("acecloud_floating_ip.test", "status", "ACTIVE"),
				),
			},
		},
	})
}

func TestAccFloatingIP_withDescription(t *testing.T) {
	extNetID := acctest.ExternalNetworkID(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckFloatingIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFloatingIPConfig_withDescription(extNetID, "test floating ip"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFloatingIPExists("acecloud_floating_ip.test"),
					resource.TestCheckResourceAttr("acecloud_floating_ip.test", "description", "test floating ip"),
					resource.TestCheckResourceAttrSet("acecloud_floating_ip.test", "floating_ip_address"),
				),
			},
		},
	})
}

func TestAccFloatingIP_recreateOnDescriptionChange(t *testing.T) {
	extNetID := acctest.ExternalNetworkID(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckFloatingIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFloatingIPConfig_withDescription(extNetID, "initial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFloatingIPExists("acecloud_floating_ip.test"),
					resource.TestCheckResourceAttr("acecloud_floating_ip.test", "description", "initial"),
				),
			},
			{
				Config: testAccFloatingIPConfig_withDescription(extNetID, "updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFloatingIPExists("acecloud_floating_ip.test"),
					resource.TestCheckResourceAttr("acecloud_floating_ip.test", "description", "updated"),
				),
			},
		},
	})
}

func TestAccFloatingIP_disappears(t *testing.T) {
	extNetID := acctest.ExternalNetworkID(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckFloatingIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFloatingIPConfig_basic(extNetID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFloatingIPExists("acecloud_floating_ip.test"),
					testAccDeleteFloatingIPOutOfBand("acecloud_floating_ip.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// --- Check functions ---

func testAccCheckFloatingIPExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID is empty")
		}

		c := acctest.TestClient()
		found, err := findFloatingIPByID(c, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error listing floating IPs: %w", err)
		}
		if !found {
			return fmt.Errorf("floating IP %s not found via API", rs.Primary.ID)
		}
		return nil
	}
}

func testAccCheckFloatingIPDestroy(s *terraform.State) error {
	c := acctest.TestClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "acecloud_floating_ip" {
			continue
		}
		found, err := findFloatingIPByID(c, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("unexpected error checking floating IP %s: %w", rs.Primary.ID, err)
		}
		if found {
			return fmt.Errorf("floating IP %s still exists after destroy", rs.Primary.ID)
		}
	}
	return nil
}

// findFloatingIPByID lists all floating IPs and checks for the given ID,
// since the API has no GET-by-ID endpoint.
func findFloatingIPByID(c *client.Client, id string) (bool, error) {
	apiResp, err := c.Get(context.Background(), "/cloud/floating-ips", nil)
	if err != nil {
		if client.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	var items []map[string]interface{}
	if err := json.Unmarshal(apiResp.Data, &items); err != nil {
		return false, fmt.Errorf("failed to parse floating IPs list: %w", err)
	}

	for _, item := range items {
		if itemID, ok := item["id"].(string); ok && itemID == id {
			return true, nil
		}
	}
	return false, nil
}

func testAccDeleteFloatingIPOutOfBand(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		c := acctest.TestClient()
		body := map[string]interface{}{
			"key":    "id",
			"values": []string{rs.Primary.ID},
		}
		_, err := c.Delete(context.Background(), "/cloud/floating-ips", body)
		if err != nil {
			return fmt.Errorf("failed to delete floating IP out-of-band: %w", err)
		}
		return nil
	}
}

// --- Config functions ---

func testAccFloatingIPConfig_basic(extNetID string) string {
	return acctest.ProviderConfig() + fmt.Sprintf(`
resource "acecloud_floating_ip" "test" {
  floating_network_id = %[1]q
}
`, extNetID)
}

func testAccFloatingIPConfig_withDescription(extNetID, description string) string {
	return acctest.ProviderConfig() + fmt.Sprintf(`
resource "acecloud_floating_ip" "test" {
  floating_network_id = %[1]q
  description         = %[2]q
}
`, extNetID, description)
}

// testAccFloatingIPConfig_forAssociation creates a floating IP without the
// provider config prefix (intended to be composed with other configs).
func testAccFloatingIPConfig_forAssociation() string {
	return fmt.Sprintf(`
resource "acecloud_floating_ip" "test" {
  floating_network_id = %[1]q
}
`, os.Getenv("ACECLOUD_EXTERNAL_NETWORK_ID"))
}
