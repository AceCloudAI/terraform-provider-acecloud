package lb_health_monitor_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/AceCloudAI/terraform-provider-acecloud/internal/acctest"
	"github.com/AceCloudAI/terraform-provider-acecloud/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccLBHealthMonitor_basic(t *testing.T) {
	rName := acctest.RandomName("hm")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckLBHealthMonitorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBHealthMonitorConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLBHealthMonitorExists("acecloud_lb_health_monitor.test"),
					resource.TestCheckResourceAttr("acecloud_lb_health_monitor.test", "name", rName+"-monitor"),
					resource.TestCheckResourceAttrSet("acecloud_lb_health_monitor.test", "id"),
					resource.TestCheckResourceAttr("acecloud_lb_health_monitor.test", "type", "HTTP"),
					resource.TestCheckResourceAttr("acecloud_lb_health_monitor.test", "delay", "5"),
					resource.TestCheckResourceAttr("acecloud_lb_health_monitor.test", "timeout", "3"),
					resource.TestCheckResourceAttr("acecloud_lb_health_monitor.test", "max_retries", "3"),
					resource.TestCheckResourceAttrPair(
						"acecloud_lb_health_monitor.test", "pool_id",
						"acecloud_lb_pool.test", "id",
					),
				),
			},
		},
	})
}

func TestAccLBHealthMonitor_tcp(t *testing.T) {
	rName := acctest.RandomName("hm")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckLBHealthMonitorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBHealthMonitorConfig_tcp(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLBHealthMonitorExists("acecloud_lb_health_monitor.test"),
					resource.TestCheckResourceAttr("acecloud_lb_health_monitor.test", "type", "TCP"),
				),
			},
		},
	})
}

func TestAccLBHealthMonitor_update(t *testing.T) {
	rName := acctest.RandomName("hm")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckLBHealthMonitorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBHealthMonitorConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLBHealthMonitorExists("acecloud_lb_health_monitor.test"),
					resource.TestCheckResourceAttr("acecloud_lb_health_monitor.test", "delay", "5"),
					resource.TestCheckResourceAttr("acecloud_lb_health_monitor.test", "timeout", "3"),
					resource.TestCheckResourceAttr("acecloud_lb_health_monitor.test", "max_retries", "3"),
				),
			},
			{
				Config: testAccLBHealthMonitorConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLBHealthMonitorExists("acecloud_lb_health_monitor.test"),
					resource.TestCheckResourceAttr("acecloud_lb_health_monitor.test", "delay", "10"),
					resource.TestCheckResourceAttr("acecloud_lb_health_monitor.test", "timeout", "5"),
					resource.TestCheckResourceAttr("acecloud_lb_health_monitor.test", "max_retries", "5"),
					resource.TestCheckResourceAttr("acecloud_lb_health_monitor.test", "url_path", "/health"),
					resource.TestCheckResourceAttr("acecloud_lb_health_monitor.test", "expected_codes", "200-299"),
				),
			},
		},
	})
}

func TestAccLBHealthMonitor_disappears(t *testing.T) {
	rName := acctest.RandomName("hm")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckLBHealthMonitorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBHealthMonitorConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLBHealthMonitorExists("acecloud_lb_health_monitor.test"),
					testAccDeleteLBHealthMonitorOutOfBand("acecloud_lb_health_monitor.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// --- Check functions ---

func testAccCheckLBHealthMonitorExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID is empty")
		}

		c := acctest.TestClient()
		path := fmt.Sprintf("/cloud/loadbalancers/pools/health-monitors/%s", rs.Primary.ID)
		_, err := c.Get(context.Background(), path, nil)
		if err != nil {
			return fmt.Errorf("LB health monitor %s not found via API: %w", rs.Primary.ID, err)
		}
		return nil
	}
}

func testAccCheckLBHealthMonitorDestroy(s *terraform.State) error {
	c := acctest.TestClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "acecloud_lb_health_monitor" {
			continue
		}
		path := fmt.Sprintf("/cloud/loadbalancers/pools/health-monitors/%s", rs.Primary.ID)
		_, err := c.Get(context.Background(), path, nil)
		if err == nil {
			return fmt.Errorf("LB health monitor %s still exists after destroy", rs.Primary.ID)
		}
		if !client.IsNotFound(err) {
			return fmt.Errorf("unexpected error checking LB health monitor %s: %w", rs.Primary.ID, err)
		}
	}
	return nil
}

func testAccDeleteLBHealthMonitorOutOfBand(resourceName string) resource.TestCheckFunc {
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
		_, err := c.Delete(context.Background(), "/cloud/loadbalancers/pools/health-monitors", body)
		if err != nil {
			return fmt.Errorf("failed to delete LB health monitor out-of-band: %w", err)
		}
		return nil
	}
}

// --- Config functions ---

// testAccLBHealthMonitorConfig_baseLBStack returns the full LB stack
// infrastructure: VPC + subnet + LB + listener + pool.
func testAccLBHealthMonitorConfig_baseLBStack(name string) string {
	return acctest.ConfigCompose(
		acctest.TestAccBaseVPCConfig(name),
		fmt.Sprintf(`
resource "acecloud_load_balancer" "test" {
  name      = %[1]q
  subnet_id = acecloud_vpc.test-base.subnet_id
  tags      = ["ALB"]
}

resource "acecloud_lb_listener" "test" {
  name            = "%[1]s-listener"
  protocol        = "HTTP"
  protocol_port   = 80
  loadbalancer_id = acecloud_load_balancer.test.id
}

resource "acecloud_lb_pool" "test" {
  name         = "%[1]s-pool"
  protocol     = "HTTP"
  lb_algorithm = "ROUND_ROBIN"
  listener_id  = acecloud_lb_listener.test.id
}
`, name),
	)
}

func testAccLBHealthMonitorConfig_basic(name string) string {
	return acctest.ConfigCompose(
		testAccLBHealthMonitorConfig_baseLBStack(name),
		fmt.Sprintf(`
resource "acecloud_lb_health_monitor" "test" {
  name        = "%[1]s-monitor"
  pool_id     = acecloud_lb_pool.test.id
  type        = "HTTP"
  delay       = 5
  timeout     = 3
  max_retries = 3
}
`, name),
	)
}

func testAccLBHealthMonitorConfig_tcp(name string) string {
	return acctest.ConfigCompose(
		testAccLBHealthMonitorConfig_baseLBStack(name),
		fmt.Sprintf(`
resource "acecloud_lb_health_monitor" "test" {
  name        = "%[1]s-monitor"
  pool_id     = acecloud_lb_pool.test.id
  type        = "TCP"
  delay       = 5
  timeout     = 3
  max_retries = 3
}
`, name),
	)
}

func testAccLBHealthMonitorConfig_updated(name string) string {
	return acctest.ConfigCompose(
		testAccLBHealthMonitorConfig_baseLBStack(name),
		fmt.Sprintf(`
resource "acecloud_lb_health_monitor" "test" {
  name           = "%[1]s-monitor"
  pool_id        = acecloud_lb_pool.test.id
  type           = "HTTP"
  delay          = 10
  timeout        = 5
  max_retries    = 5
  url_path       = "/health"
  expected_codes = "200-299"
}
`, name),
	)
}
