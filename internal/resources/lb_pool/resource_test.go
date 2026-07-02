package lb_pool_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/AceCloudAI/terraform-provider-acecloud/internal/acctest"
	"github.com/AceCloudAI/terraform-provider-acecloud/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccLBPool_basic(t *testing.T) {
	rName := acctest.RandomName("pool")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckLBPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBPoolConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLBPoolExists("acecloud_lb_pool.test"),
					resource.TestCheckResourceAttr("acecloud_lb_pool.test", "name", rName+"-pool"),
					resource.TestCheckResourceAttrSet("acecloud_lb_pool.test", "id"),
					resource.TestCheckResourceAttr("acecloud_lb_pool.test", "protocol", "HTTP"),
					resource.TestCheckResourceAttr("acecloud_lb_pool.test", "lb_algorithm", "ROUND_ROBIN"),
					resource.TestCheckResourceAttrPair(
						"acecloud_lb_pool.test", "listener_id",
						"acecloud_lb_listener.test", "id",
					),
				),
			},
		},
	})
}

func TestAccLBPool_leastConnections(t *testing.T) {
	rName := acctest.RandomName("pool")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckLBPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBPoolConfig_leastConn(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLBPoolExists("acecloud_lb_pool.test"),
					resource.TestCheckResourceAttr("acecloud_lb_pool.test", "lb_algorithm", "LEAST_CONNECTIONS"),
				),
			},
		},
	})
}

func TestAccLBPool_recreate(t *testing.T) {
	rName1 := acctest.RandomName("pool")
	rName2 := acctest.RandomName("pool")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckLBPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBPoolConfig_basic(rName1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLBPoolExists("acecloud_lb_pool.test"),
					resource.TestCheckResourceAttr("acecloud_lb_pool.test", "name", rName1+"-pool"),
				),
			},
			{
				Config: testAccLBPoolConfig_basic(rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLBPoolExists("acecloud_lb_pool.test"),
					resource.TestCheckResourceAttr("acecloud_lb_pool.test", "name", rName2+"-pool"),
				),
			},
		},
	})
}

func TestAccLBPool_disappears(t *testing.T) {
	rName := acctest.RandomName("pool")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckLBPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBPoolConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLBPoolExists("acecloud_lb_pool.test"),
					testAccDeleteLBPoolOutOfBand("acecloud_lb_pool.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// --- Check functions ---

func testAccCheckLBPoolExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID is empty")
		}

		c := acctest.TestClient()
		path := fmt.Sprintf("/cloud/loadbalancers/pools/%s", rs.Primary.ID)
		_, err := c.Get(context.Background(), path, nil)
		if err != nil {
			return fmt.Errorf("LB pool %s not found via API: %w", rs.Primary.ID, err)
		}
		return nil
	}
}

func testAccCheckLBPoolDestroy(s *terraform.State) error {
	c := acctest.TestClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "acecloud_lb_pool" {
			continue
		}
		path := fmt.Sprintf("/cloud/loadbalancers/pools/%s", rs.Primary.ID)
		_, err := c.Get(context.Background(), path, nil)
		if err == nil {
			return fmt.Errorf("LB pool %s still exists after destroy", rs.Primary.ID)
		}
		if !client.IsNotFound(err) {
			return fmt.Errorf("unexpected error checking LB pool %s: %w", rs.Primary.ID, err)
		}
	}
	return nil
}

func testAccDeleteLBPoolOutOfBand(resourceName string) resource.TestCheckFunc {
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
		_, err := c.Delete(context.Background(), "/cloud/loadbalancers/pools", body)
		if err != nil {
			return fmt.Errorf("failed to delete LB pool out-of-band: %w", err)
		}
		return nil
	}
}

// --- Config functions ---

// testAccLBPoolConfig_baseLBWithListener returns the base infrastructure
// shared by all pool tests: VPC + subnet + LB + listener.
func testAccLBPoolConfig_baseLBWithListener(name string) string {
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
`, name),
	)
}

func testAccLBPoolConfig_basic(name string) string {
	return acctest.ConfigCompose(
		testAccLBPoolConfig_baseLBWithListener(name),
		fmt.Sprintf(`
resource "acecloud_lb_pool" "test" {
  name         = "%[1]s-pool"
  protocol     = "HTTP"
  lb_algorithm = "ROUND_ROBIN"
  listener_id  = acecloud_lb_listener.test.id
}
`, name),
	)
}

func testAccLBPoolConfig_leastConn(name string) string {
	return acctest.ConfigCompose(
		testAccLBPoolConfig_baseLBWithListener(name),
		fmt.Sprintf(`
resource "acecloud_lb_pool" "test" {
  name         = "%[1]s-pool"
  protocol     = "HTTP"
  lb_algorithm = "LEAST_CONNECTIONS"
  listener_id  = acecloud_lb_listener.test.id
}
`, name),
	)
}
