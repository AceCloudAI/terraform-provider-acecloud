package lb_listener_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/AceCloudAI/terraform-provider-acecloud/internal/acctest"
	"github.com/AceCloudAI/terraform-provider-acecloud/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccLBListener_basic(t *testing.T) {
	rName := acctest.RandomName("lis")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckLBListenerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBListenerConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLBListenerExists("acecloud_lb_listener.test"),
					resource.TestCheckResourceAttr("acecloud_lb_listener.test", "name", rName+"-listener"),
					resource.TestCheckResourceAttrSet("acecloud_lb_listener.test", "id"),
					resource.TestCheckResourceAttr("acecloud_lb_listener.test", "protocol", "HTTP"),
					resource.TestCheckResourceAttr("acecloud_lb_listener.test", "protocol_port", "80"),
					resource.TestCheckResourceAttrPair(
						"acecloud_lb_listener.test", "loadbalancer_id",
						"acecloud_load_balancer.test", "id",
					),
				),
			},
		},
	})
}

func TestAccLBListener_tcp(t *testing.T) {
	rName := acctest.RandomName("lis")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckLBListenerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBListenerConfig_tcp(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLBListenerExists("acecloud_lb_listener.test"),
					resource.TestCheckResourceAttr("acecloud_lb_listener.test", "protocol", "TCP"),
					resource.TestCheckResourceAttr("acecloud_lb_listener.test", "protocol_port", "443"),
				),
			},
		},
	})
}

func TestAccLBListener_recreate(t *testing.T) {
	rName1 := acctest.RandomName("lis")
	rName2 := acctest.RandomName("lis")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckLBListenerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBListenerConfig_basic(rName1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLBListenerExists("acecloud_lb_listener.test"),
					resource.TestCheckResourceAttr("acecloud_lb_listener.test", "name", rName1+"-listener"),
				),
			},
			{
				Config: testAccLBListenerConfig_basic(rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLBListenerExists("acecloud_lb_listener.test"),
					resource.TestCheckResourceAttr("acecloud_lb_listener.test", "name", rName2+"-listener"),
				),
			},
		},
	})
}

func TestAccLBListener_disappears(t *testing.T) {
	rName := acctest.RandomName("lis")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckLBListenerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBListenerConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLBListenerExists("acecloud_lb_listener.test"),
					testAccDeleteLBListenerOutOfBand("acecloud_lb_listener.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// --- Check functions ---

func testAccCheckLBListenerExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID is empty")
		}

		c := acctest.TestClient()
		path := fmt.Sprintf("/cloud/loadbalancers/listeners/%s", rs.Primary.ID)
		_, err := c.Get(context.Background(), path, nil)
		if err != nil {
			return fmt.Errorf("LB listener %s not found via API: %w", rs.Primary.ID, err)
		}
		return nil
	}
}

func testAccCheckLBListenerDestroy(s *terraform.State) error {
	c := acctest.TestClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "acecloud_lb_listener" {
			continue
		}
		path := fmt.Sprintf("/cloud/loadbalancers/listeners/%s", rs.Primary.ID)
		_, err := c.Get(context.Background(), path, nil)
		if err == nil {
			return fmt.Errorf("LB listener %s still exists after destroy", rs.Primary.ID)
		}
		if !client.IsNotFound(err) {
			return fmt.Errorf("unexpected error checking LB listener %s: %w", rs.Primary.ID, err)
		}
	}
	return nil
}

func testAccDeleteLBListenerOutOfBand(resourceName string) resource.TestCheckFunc {
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
		_, err := c.Delete(context.Background(), "/cloud/loadbalancers/listeners", body)
		if err != nil {
			return fmt.Errorf("failed to delete LB listener out-of-band: %w", err)
		}
		return nil
	}
}

// --- Config functions ---

// testAccLBListenerConfig_baseLB returns the base LB infrastructure shared
// by all listener tests (VPC + subnet + load balancer).
func testAccLBListenerConfig_baseLB(name string) string {
	return acctest.ConfigCompose(
		acctest.TestAccBaseVPCConfig(name),
		fmt.Sprintf(`
resource "acecloud_load_balancer" "test" {
  name      = %[1]q
  subnet_id = acecloud_vpc.test-base.subnet_id
  tags      = ["ALB"]
}
`, name),
	)
}

func testAccLBListenerConfig_basic(name string) string {
	return acctest.ConfigCompose(
		testAccLBListenerConfig_baseLB(name),
		fmt.Sprintf(`
resource "acecloud_lb_listener" "test" {
  name            = "%[1]s-listener"
  protocol        = "HTTP"
  protocol_port   = 80
  loadbalancer_id = acecloud_load_balancer.test.id
}
`, name),
	)
}

func testAccLBListenerConfig_tcp(name string) string {
	return acctest.ConfigCompose(
		testAccLBListenerConfig_baseLB(name),
		fmt.Sprintf(`
resource "acecloud_lb_listener" "test" {
  name            = "%[1]s-listener"
  protocol        = "TCP"
  protocol_port   = 443
  loadbalancer_id = acecloud_load_balancer.test.id
}
`, name),
	)
}
