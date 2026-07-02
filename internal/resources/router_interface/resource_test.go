package router_interface_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/AceCloudAI/terraform-provider-acecloud/internal/acctest"
	"github.com/AceCloudAI/terraform-provider-acecloud/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccRouterInterface_basic(t *testing.T) {
	rName := acctest.RandomName("ri")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckRouterInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouterInterfaceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouterInterfaceExists("acecloud_router_interface.test"),
					resource.TestCheckResourceAttrSet("acecloud_router_interface.test", "id"),
					resource.TestCheckResourceAttrSet("acecloud_router_interface.test", "ip_address"),
					resource.TestCheckResourceAttrSet("acecloud_router_interface.test", "mac_address"),
					resource.TestCheckResourceAttrPair(
						"acecloud_router_interface.test", "router_id",
						"acecloud_router.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						"acecloud_router_interface.test", "subnet_id",
						"acecloud_vpc.test-base", "subnet_id",
					),
				),
			},
		},
	})
}

func TestAccRouterInterface_recreate(t *testing.T) {
	rName1 := acctest.RandomName("ri")
	rName2 := acctest.RandomName("ri")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckRouterInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouterInterfaceConfig_basic(rName1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouterInterfaceExists("acecloud_router_interface.test"),
				),
			},
			{
				Config: testAccRouterInterfaceConfig_basic(rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouterInterfaceExists("acecloud_router_interface.test"),
				),
			},
		},
	})
}

func TestAccRouterInterface_disappears(t *testing.T) {
	rName := acctest.RandomName("ri")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckRouterInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouterInterfaceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouterInterfaceExists("acecloud_router_interface.test"),
					testAccDeleteRouterInterfaceOutOfBand("acecloud_router_interface.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// --- Check functions ---

func testAccCheckRouterInterfaceExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID is empty")
		}

		routerID := rs.Primary.Attributes["router_id"]
		subnetID := rs.Primary.Attributes["subnet_id"]
		if routerID == "" || subnetID == "" {
			return fmt.Errorf("router_id or subnet_id is empty in state")
		}

		c := acctest.TestClient()
		found, err := findRouterInterfaceBySubnet(c, routerID, subnetID)
		if err != nil {
			return fmt.Errorf("error checking router interface: %w", err)
		}
		if !found {
			return fmt.Errorf("router interface for router %s / subnet %s not found via API", routerID, subnetID)
		}
		return nil
	}
}

func testAccCheckRouterInterfaceDestroy(s *terraform.State) error {
	c := acctest.TestClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "acecloud_router_interface" {
			continue
		}
		routerID := rs.Primary.Attributes["router_id"]
		subnetID := rs.Primary.Attributes["subnet_id"]

		found, err := findRouterInterfaceBySubnet(c, routerID, subnetID)
		if err != nil {
			if client.IsNotFound(err) {
				continue
			}
			return fmt.Errorf("unexpected error checking router interface: %w", err)
		}
		if found {
			return fmt.Errorf("router interface for router %s / subnet %s still exists after destroy", routerID, subnetID)
		}
	}
	return nil
}

type interfacePort struct {
	SubnetID string    `json:"subnet_id"`
	FixedIPs []fixedIP `json:"fixed_ips"`
}

type fixedIP struct {
	SubnetID string `json:"subnet_id"`
}

func findRouterInterfaceBySubnet(c *client.Client, routerID, subnetID string) (bool, error) {
	path := fmt.Sprintf("/os/neutron/interfaces/%s", routerID)
	apiResp, err := c.Get(context.Background(), path, nil)
	if err != nil {
		return false, err
	}

	var wrapped struct {
		Ports []interfacePort `json:"ports"`
	}
	if err := json.Unmarshal(apiResp.Data, &wrapped); err == nil {
		for _, port := range wrapped.Ports {
			if port.SubnetID == subnetID {
				return true, nil
			}
			for _, fip := range port.FixedIPs {
				if fip.SubnetID == subnetID {
					return true, nil
				}
			}
		}
		return false, nil
	}

	var ports []interfacePort
	if err := json.Unmarshal(apiResp.Data, &ports); err != nil {
		return false, fmt.Errorf("failed to parse interfaces response: %w", err)
	}
	for _, port := range ports {
		if port.SubnetID == subnetID {
			return true, nil
		}
		for _, fip := range port.FixedIPs {
			if fip.SubnetID == subnetID {
				return true, nil
			}
		}
	}
	return false, nil
}

func testAccDeleteRouterInterfaceOutOfBand(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		routerID := rs.Primary.Attributes["router_id"]
		portID := rs.Primary.ID

		c := acctest.TestClient()
		path := fmt.Sprintf("/os/neutron/interfaces/%s", routerID)
		body := map[string]interface{}{
			"key":    "id",
			"values": []string{portID},
		}
		_, err := c.Delete(context.Background(), path, body)
		if err != nil {
			return fmt.Errorf("failed to delete router interface out-of-band: %w", err)
		}
		return nil
	}
}

// --- Config functions ---

func testAccRouterInterfaceConfig_basic(name string) string {
	return acctest.ConfigCompose(
		acctest.TestAccBaseVPCConfig(name),
		fmt.Sprintf(`
resource "acecloud_router" "test" {
  name = "%[1]s-router"
}

resource "acecloud_router_interface" "test" {
  router_id = acecloud_router.test.id
  subnet_id = acecloud_vpc.test-base.subnet_id
}
`, name),
	)
}
