package load_balancer_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/AceCloudAI/terraform-provider-acecloud/internal/acctest"
	"github.com/AceCloudAI/terraform-provider-acecloud/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccLoadBalancer_basic(t *testing.T) {
	rName := acctest.RandomName("lb")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists("acecloud_load_balancer.test"),
					resource.TestCheckResourceAttr("acecloud_load_balancer.test", "name", rName),
					resource.TestCheckResourceAttrSet("acecloud_load_balancer.test", "id"),
					resource.TestCheckResourceAttrSet("acecloud_load_balancer.test", "vip_address"),
					resource.TestCheckResourceAttr("acecloud_load_balancer.test", "provisioning_status", "ACTIVE"),
					resource.TestCheckResourceAttr("acecloud_load_balancer.test", "tags.#", "1"),
					resource.TestCheckResourceAttr("acecloud_load_balancer.test", "tags.0", "ALB"),
				),
			},
		},
	})
}

func TestAccLoadBalancer_nlb(t *testing.T) {
	rName := acctest.RandomName("lb")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_nlb(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists("acecloud_load_balancer.test"),
					resource.TestCheckResourceAttr("acecloud_load_balancer.test", "tags.0", "NLB"),
				),
			},
		},
	})
}

func TestAccLoadBalancer_update(t *testing.T) {
	rName := acctest.RandomName("lb")
	rNameUpdated := acctest.RandomName("lb-upd")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists("acecloud_load_balancer.test"),
					resource.TestCheckResourceAttr("acecloud_load_balancer.test", "name", rName),
				),
			},
			{
				Config: testAccLoadBalancerConfig_basic(rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists("acecloud_load_balancer.test"),
					resource.TestCheckResourceAttr("acecloud_load_balancer.test", "name", rNameUpdated),
				),
			},
		},
	})
}

func TestAccLoadBalancer_withDescription(t *testing.T) {
	rName := acctest.RandomName("lb")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_withDescription(rName, "initial desc"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists("acecloud_load_balancer.test"),
					resource.TestCheckResourceAttr("acecloud_load_balancer.test", "description", "initial desc"),
				),
			},
			{
				Config: testAccLoadBalancerConfig_withDescription(rName, "updated desc"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("acecloud_load_balancer.test", "description", "updated desc"),
				),
			},
		},
	})
}

func TestAccLoadBalancer_disappears(t *testing.T) {
	rName := acctest.RandomName("lb")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerExists("acecloud_load_balancer.test"),
					testAccDeleteLoadBalancerOutOfBand("acecloud_load_balancer.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// --- Check functions ---

func testAccCheckLoadBalancerExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID is empty")
		}

		c := acctest.TestClient()
		path := fmt.Sprintf("/cloud/loadbalancers/%s", rs.Primary.ID)
		_, err := c.Get(context.Background(), path, nil)
		if err != nil {
			return fmt.Errorf("load balancer %s not found via API: %w", rs.Primary.ID, err)
		}
		return nil
	}
}

func testAccCheckLoadBalancerDestroy(s *terraform.State) error {
	c := acctest.TestClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "acecloud_load_balancer" {
			continue
		}
		path := fmt.Sprintf("/cloud/loadbalancers/%s", rs.Primary.ID)
		_, err := c.Get(context.Background(), path, nil)
		if err == nil {
			return fmt.Errorf("load balancer %s still exists after destroy", rs.Primary.ID)
		}
		if !client.IsNotFound(err) {
			return fmt.Errorf("unexpected error checking load balancer %s: %w", rs.Primary.ID, err)
		}
	}
	return nil
}

func testAccDeleteLoadBalancerOutOfBand(resourceName string) resource.TestCheckFunc {
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
		_, err := c.Delete(context.Background(), "/cloud/loadbalancers", body)
		if err != nil {
			return fmt.Errorf("failed to delete load balancer out-of-band: %w", err)
		}
		return nil
	}
}

// --- Config functions ---

func testAccLoadBalancerConfig_basic(name string) string {
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

func testAccLoadBalancerConfig_nlb(name string) string {
	return acctest.ConfigCompose(
		acctest.TestAccBaseVPCConfig(name),
		fmt.Sprintf(`
resource "acecloud_load_balancer" "test" {
  name      = %[1]q
  subnet_id = acecloud_vpc.test-base.subnet_id
  tags      = ["NLB"]
}
`, name),
	)
}

func testAccLoadBalancerConfig_withDescription(name, description string) string {
	return acctest.ConfigCompose(
		acctest.TestAccBaseVPCConfig(name),
		fmt.Sprintf(`
resource "acecloud_load_balancer" "test" {
  name        = %[1]q
  subnet_id   = acecloud_vpc.test-base.subnet_id
  description = %[2]q
  tags        = ["ALB"]
}
`, name, description),
	)
}
