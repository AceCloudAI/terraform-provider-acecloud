package router_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/AceCloudAI/terraform-provider-acecloud/internal/acctest"
	"github.com/AceCloudAI/terraform-provider-acecloud/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccRouter_basic(t *testing.T) {
	rName := acctest.RandomName("rtr")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckRouterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouterConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouterExists("acecloud_router.test"),
					resource.TestCheckResourceAttr("acecloud_router.test", "name", rName),
					resource.TestCheckResourceAttrSet("acecloud_router.test", "id"),
					resource.TestCheckResourceAttr("acecloud_router.test", "admin_state_up", "true"),
					resource.TestCheckResourceAttrSet("acecloud_router.test", "status"),
				),
			},
		},
	})
}

func TestAccRouter_update(t *testing.T) {
	rName := acctest.RandomName("rtr")
	rNameUpdated := acctest.RandomName("rtr-upd")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckRouterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouterConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouterExists("acecloud_router.test"),
					resource.TestCheckResourceAttr("acecloud_router.test", "name", rName),
				),
			},
			{
				Config: testAccRouterConfig_basic(rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouterExists("acecloud_router.test"),
					resource.TestCheckResourceAttr("acecloud_router.test", "name", rNameUpdated),
				),
			},
		},
	})
}

func TestAccRouter_withExternalGateway(t *testing.T) {
	extNetID := acctest.ExternalNetworkID(t)
	rName := acctest.RandomName("rtr")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckRouterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouterConfig_withGateway(rName, extNetID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouterExists("acecloud_router.test"),
					resource.TestCheckResourceAttr("acecloud_router.test", "name", rName),
					resource.TestCheckResourceAttr("acecloud_router.test", "external_gateway_network_id", extNetID),
				),
			},
		},
	})
}

func TestAccRouter_updateGateway(t *testing.T) {
	extNetID := acctest.ExternalNetworkID(t)
	rName := acctest.RandomName("rtr")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckRouterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouterConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouterExists("acecloud_router.test"),
				),
			},
			{
				Config: testAccRouterConfig_withGateway(rName, extNetID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouterExists("acecloud_router.test"),
					resource.TestCheckResourceAttr("acecloud_router.test", "external_gateway_network_id", extNetID),
				),
			},
		},
	})
}

func TestAccRouter_adminStateDown(t *testing.T) {
	rName := acctest.RandomName("rtr")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckRouterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouterConfig_adminState(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouterExists("acecloud_router.test"),
					resource.TestCheckResourceAttr("acecloud_router.test", "admin_state_up", "false"),
				),
			},
			{
				Config: testAccRouterConfig_adminState(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouterExists("acecloud_router.test"),
					resource.TestCheckResourceAttr("acecloud_router.test", "admin_state_up", "true"),
				),
			},
		},
	})
}

func TestAccRouter_disappears(t *testing.T) {
	rName := acctest.RandomName("rtr")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckRouterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouterConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouterExists("acecloud_router.test"),
					testAccDeleteRouterOutOfBand("acecloud_router.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// --- Check functions ---

func testAccCheckRouterExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID is empty")
		}

		c := acctest.TestClient()
		path := fmt.Sprintf("/os/neutron/routers/%s", rs.Primary.ID)
		_, err := c.Get(context.Background(), path, nil)
		if err != nil {
			return fmt.Errorf("router %s not found via API: %w", rs.Primary.ID, err)
		}
		return nil
	}
}

func testAccCheckRouterDestroy(s *terraform.State) error {
	c := acctest.TestClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "acecloud_router" {
			continue
		}
		path := fmt.Sprintf("/os/neutron/routers/%s", rs.Primary.ID)
		_, err := c.Get(context.Background(), path, nil)
		if err == nil {
			return fmt.Errorf("router %s still exists after destroy", rs.Primary.ID)
		}
		if !client.IsNotFound(err) {
			return fmt.Errorf("unexpected error checking router %s: %w", rs.Primary.ID, err)
		}
	}
	return nil
}

func testAccDeleteRouterOutOfBand(resourceName string) resource.TestCheckFunc {
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
		_, err := c.Delete(context.Background(), "/os/neutron/routers", body)
		if err != nil {
			return fmt.Errorf("failed to delete router out-of-band: %w", err)
		}
		return nil
	}
}

// --- Config functions ---

func testAccRouterConfig_basic(name string) string {
	return acctest.ProviderConfig() + fmt.Sprintf(`
resource "acecloud_router" "test" {
  name = %[1]q
}
`, name)
}

func testAccRouterConfig_withGateway(name, extNetID string) string {
	return acctest.ProviderConfig() + fmt.Sprintf(`
resource "acecloud_router" "test" {
  name                         = %[1]q
  external_gateway_network_id  = %[2]q
}
`, name, extNetID)
}

func testAccRouterConfig_adminState(name string, adminStateUp bool) string {
	return acctest.ProviderConfig() + fmt.Sprintf(`
resource "acecloud_router" "test" {
  name           = %[1]q
  admin_state_up = %[2]t
}
`, name, adminStateUp)
}
