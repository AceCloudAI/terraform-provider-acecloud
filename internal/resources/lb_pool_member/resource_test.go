package lb_pool_member_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/AceCloudAI/terraform-provider-acecloud/internal/acctest"
	"github.com/AceCloudAI/terraform-provider-acecloud/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccLBPoolMember_basic(t *testing.T) {
	rName := acctest.RandomName("mbr")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckLBPoolMemberDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBPoolMemberConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLBPoolMemberExists("acecloud_lb_pool_member.test"),
					resource.TestCheckResourceAttrSet("acecloud_lb_pool_member.test", "id"),
					resource.TestCheckResourceAttr("acecloud_lb_pool_member.test", "address", "10.0.0.10"),
					resource.TestCheckResourceAttr("acecloud_lb_pool_member.test", "protocol_port", "8080"),
					resource.TestCheckResourceAttr("acecloud_lb_pool_member.test", "weight", "1"),
					resource.TestCheckResourceAttrPair(
						"acecloud_lb_pool_member.test", "pool_id",
						"acecloud_lb_pool.test", "id",
					),
				),
			},
		},
	})
}

func TestAccLBPoolMember_updateWeight(t *testing.T) {
	rName := acctest.RandomName("mbr")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckLBPoolMemberDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBPoolMemberConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLBPoolMemberExists("acecloud_lb_pool_member.test"),
					resource.TestCheckResourceAttr("acecloud_lb_pool_member.test", "weight", "1"),
				),
			},
			{
				Config: testAccLBPoolMemberConfig_weight(rName, 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLBPoolMemberExists("acecloud_lb_pool_member.test"),
					resource.TestCheckResourceAttr("acecloud_lb_pool_member.test", "weight", "5"),
				),
			},
		},
	})
}

func TestAccLBPoolMember_disappears(t *testing.T) {
	rName := acctest.RandomName("mbr")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckLBPoolMemberDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBPoolMemberConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLBPoolMemberExists("acecloud_lb_pool_member.test"),
					testAccDeleteLBPoolMemberOutOfBand("acecloud_lb_pool_member.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// --- Check functions ---

func testAccCheckLBPoolMemberExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID is empty")
		}

		poolID := rs.Primary.Attributes["pool_id"]
		c := acctest.TestClient()
		path := fmt.Sprintf("/cloud/loadbalancers/pools/%s/backend-servers/%s", poolID, rs.Primary.ID)
		_, err := c.Get(context.Background(), path, nil)
		if err != nil {
			return fmt.Errorf("LB pool member %s not found via API: %w", rs.Primary.ID, err)
		}
		return nil
	}
}

func testAccCheckLBPoolMemberDestroy(s *terraform.State) error {
	c := acctest.TestClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "acecloud_lb_pool_member" {
			continue
		}
		poolID := rs.Primary.Attributes["pool_id"]
		path := fmt.Sprintf("/cloud/loadbalancers/pools/%s/backend-servers/%s", poolID, rs.Primary.ID)
		_, err := c.Get(context.Background(), path, nil)
		if err == nil {
			return fmt.Errorf("LB pool member %s still exists after destroy", rs.Primary.ID)
		}
		if !client.IsNotFound(err) {
			return fmt.Errorf("unexpected error checking LB pool member %s: %w", rs.Primary.ID, err)
		}
	}
	return nil
}

func testAccDeleteLBPoolMemberOutOfBand(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		poolID := rs.Primary.Attributes["pool_id"]
		c := acctest.TestClient()
		path := fmt.Sprintf("/cloud/loadbalancers/pools/%s/backend-servers", poolID)
		body := map[string]interface{}{
			"key":    "id",
			"values": []string{rs.Primary.ID},
		}
		_, err := c.Delete(context.Background(), path, body)
		if err != nil {
			return fmt.Errorf("failed to delete LB pool member out-of-band: %w", err)
		}
		return nil
	}
}

// --- Config functions ---

// testAccLBPoolMemberConfig_baseLBStack returns the full LB stack
// infrastructure: VPC + subnet + LB + listener + pool.
func testAccLBPoolMemberConfig_baseLBStack(name string) string {
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

func testAccLBPoolMemberConfig_basic(name string) string {
	return acctest.ConfigCompose(
		testAccLBPoolMemberConfig_baseLBStack(name),
		`
resource "acecloud_lb_pool_member" "test" {
  pool_id       = acecloud_lb_pool.test.id
  address       = "10.0.0.10"
  protocol_port = 8080
}
`,
	)
}

func testAccLBPoolMemberConfig_weight(name string, weight int) string {
	return acctest.ConfigCompose(
		testAccLBPoolMemberConfig_baseLBStack(name),
		fmt.Sprintf(`
resource "acecloud_lb_pool_member" "test" {
  pool_id       = acecloud_lb_pool.test.id
  address       = "10.0.0.10"
  protocol_port = 8080
  weight        = %[1]d
}
`, weight),
	)
}
