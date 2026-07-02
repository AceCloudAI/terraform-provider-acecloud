package floating_ip_association_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/AceCloudAI/terraform-provider-acecloud/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccFloatingIPAssociation_basic(t *testing.T) {
	acctest.ExternalNetworkID(t)
	flavorID := acctest.FlavorID(t)
	imageID := acctest.ImageID(t)

	rName := acctest.RandomName("fipa")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckFloatingIPAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFloatingIPAssociationConfig_basic(rName, flavorID, imageID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFloatingIPAssociationExists("acecloud_floating_ip_association.test"),
					resource.TestCheckResourceAttrSet("acecloud_floating_ip_association.test", "id"),
					resource.TestCheckResourceAttrSet("acecloud_floating_ip_association.test", "floating_ip_address"),
					resource.TestCheckResourceAttrSet("acecloud_floating_ip_association.test", "instance_id"),
				),
			},
		},
	})
}

func TestAccFloatingIPAssociation_disappears(t *testing.T) {
	acctest.ExternalNetworkID(t)
	flavorID := acctest.FlavorID(t)
	imageID := acctest.ImageID(t)

	rName := acctest.RandomName("fipa")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckFloatingIPAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFloatingIPAssociationConfig_basic(rName, flavorID, imageID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFloatingIPAssociationExists("acecloud_floating_ip_association.test"),
					testAccDetachFloatingIPOutOfBand("acecloud_floating_ip_association.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// --- Check functions ---

func testAccCheckFloatingIPAssociationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID is empty")
		}
		return nil
	}
}

func testAccCheckFloatingIPAssociationDestroy(s *terraform.State) error {
	// The association is a logical link — once the floating IP or instance
	// is destroyed, the association is implicitly gone. We verify by
	// checking that the floating IP is either deleted or no longer
	// associated with a port.
	c := acctest.TestClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "acecloud_floating_ip_association" {
			continue
		}
		fipAddr := rs.Primary.Attributes["floating_ip_address"]
		if fipAddr == "" {
			continue
		}

		apiResp, err := c.Get(context.Background(), "/cloud/floating-ips", nil)
		if err != nil {
			continue
		}

		var items []map[string]interface{}
		if err := json.Unmarshal(apiResp.Data, &items); err != nil {
			continue
		}

		for _, item := range items {
			if addr, ok := item["floating_ip_address"].(string); ok && addr == fipAddr {
				if portID, ok := item["port_id"].(string); ok && portID != "" {
					return fmt.Errorf("floating IP %s is still associated with port %s", fipAddr, portID)
				}
			}
		}
	}
	return nil
}

func testAccDetachFloatingIPOutOfBand(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		c := acctest.TestClient()
		body := map[string]interface{}{
			"floating_ip_address": rs.Primary.Attributes["floating_ip_address"],
			"instance_id":        rs.Primary.Attributes["instance_id"],
		}
		_, err := c.PutWithParams(context.Background(), "/cloud/floating-ips/action", body, map[string]string{
			"type": "detach",
		})
		if err != nil {
			return fmt.Errorf("failed to detach floating IP out-of-band: %w", err)
		}
		return nil
	}
}

// --- Config functions ---

func testAccFloatingIPAssociationConfig_basic(name, flavorID, imageID string) string {
	extNetID := os.Getenv("ACECLOUD_EXTERNAL_NETWORK_ID")

	return acctest.ConfigCompose(
		acctest.TestAccBaseVPCConfig(name),
		acctest.TestAccBaseSecurityGroupConfig(name+"-sg"),
		fmt.Sprintf(`
resource "acecloud_key_pair" "test" {
  name = %[1]q
}

resource "acecloud_instance" "test" {
  name              = %[1]q
  flavor_id         = %[2]q
  image_id          = %[3]q
  key_pair_name     = acecloud_key_pair.test.name
  security_group_id = acecloud_security_group.test-base.id
  subnet_id         = acecloud_vpc.test-base.subnet_id
}

resource "acecloud_floating_ip" "test" {
  floating_network_id = %[4]q
}

resource "acecloud_floating_ip_association" "test" {
  floating_ip_address = acecloud_floating_ip.test.floating_ip_address
  instance_id         = acecloud_instance.test.id
}
`, name, flavorID, imageID, extNetID),
	)
}
