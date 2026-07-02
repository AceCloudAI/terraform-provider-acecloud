package security_group_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/AceCloudAI/terraform-provider-acecloud/internal/acctest"
	"github.com/AceCloudAI/terraform-provider-acecloud/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccSecurityGroup_basic(t *testing.T) {
	rName := acctest.RandomName("sg")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists("acecloud_security_group.test"),
					resource.TestCheckResourceAttr("acecloud_security_group.test", "name", rName),
					resource.TestCheckResourceAttrSet("acecloud_security_group.test", "id"),
				),
			},
		},
	})
}

func TestAccSecurityGroup_withRules(t *testing.T) {
	rName := acctest.RandomName("sg")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupConfig_withRules(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists("acecloud_security_group.test"),
					resource.TestCheckResourceAttr("acecloud_security_group.test", "name", rName),
					resource.TestCheckResourceAttr("acecloud_security_group.test", "rules.#", "2"),
				),
			},
		},
	})
}

func TestAccSecurityGroup_update(t *testing.T) {
	rName := acctest.RandomName("sg")
	rNameUpdated := acctest.RandomName("sg-upd")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupConfig_withRules(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists("acecloud_security_group.test"),
					resource.TestCheckResourceAttr("acecloud_security_group.test", "name", rName),
					resource.TestCheckResourceAttr("acecloud_security_group.test", "rules.#", "2"),
				),
			},
			{
				Config: testAccSecurityGroupConfig_updatedRules(rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists("acecloud_security_group.test"),
					resource.TestCheckResourceAttr("acecloud_security_group.test", "name", rNameUpdated),
					resource.TestCheckResourceAttr("acecloud_security_group.test", "rules.#", "1"),
				),
			},
		},
	})
}

func TestAccSecurityGroup_withDescription(t *testing.T) {
	rName := acctest.RandomName("sg")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupConfig_withDescription(rName, "test security group"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists("acecloud_security_group.test"),
					resource.TestCheckResourceAttr("acecloud_security_group.test", "description", "test security group"),
				),
			},
			{
				Config: testAccSecurityGroupConfig_withDescription(rName, "updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("acecloud_security_group.test", "description", "updated description"),
				),
			},
		},
	})
}

func TestAccSecurityGroup_disappears(t *testing.T) {
	rName := acctest.RandomName("sg")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists("acecloud_security_group.test"),
					testAccDeleteSecurityGroupOutOfBand("acecloud_security_group.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// --- Check functions ---

func testAccCheckSecurityGroupExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID is empty")
		}

		c := acctest.TestClient()
		path := fmt.Sprintf("/cloud/security-groups/%s", rs.Primary.ID)
		_, err := c.Get(context.Background(), path, nil)
		if err != nil {
			return fmt.Errorf("security group %s not found via API: %w", rs.Primary.ID, err)
		}
		return nil
	}
}

func testAccCheckSecurityGroupDestroy(s *terraform.State) error {
	c := acctest.TestClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "acecloud_security_group" {
			continue
		}
		path := fmt.Sprintf("/cloud/security-groups/%s", rs.Primary.ID)
		_, err := c.Get(context.Background(), path, nil)
		if err == nil {
			return fmt.Errorf("security group %s still exists after destroy", rs.Primary.ID)
		}
		if !client.IsNotFound(err) {
			return fmt.Errorf("unexpected error checking security group %s: %w", rs.Primary.ID, err)
		}
	}
	return nil
}

func testAccDeleteSecurityGroupOutOfBand(resourceName string) resource.TestCheckFunc {
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
		_, err := c.Delete(context.Background(), "/cloud/security-groups", body)
		if err != nil {
			return fmt.Errorf("failed to delete security group out-of-band: %w", err)
		}
		return nil
	}
}

// --- Config functions ---

func testAccSecurityGroupConfig_basic(name string) string {
	return acctest.ProviderConfig() + fmt.Sprintf(`
resource "acecloud_security_group" "test" {
  name = %[1]q
}
`, name)
}

func testAccSecurityGroupConfig_withRules(name string) string {
	return acctest.ProviderConfig() + fmt.Sprintf(`
resource "acecloud_security_group" "test" {
  name = %[1]q

  rules {
    direction        = "ingress"
    protocol         = "tcp"
    port_range_min   = 22
    port_range_max   = 22
    remote_ip_prefix = "0.0.0.0/0"
    ethertype        = "IPv4"
  }

  rules {
    direction        = "ingress"
    protocol         = "tcp"
    port_range_min   = 443
    port_range_max   = 443
    remote_ip_prefix = "0.0.0.0/0"
    ethertype        = "IPv4"
  }
}
`, name)
}

func testAccSecurityGroupConfig_updatedRules(name string) string {
	return acctest.ProviderConfig() + fmt.Sprintf(`
resource "acecloud_security_group" "test" {
  name = %[1]q

  rules {
    direction        = "ingress"
    protocol         = "tcp"
    port_range_min   = 80
    port_range_max   = 80
    remote_ip_prefix = "0.0.0.0/0"
    ethertype        = "IPv4"
  }
}
`, name)
}

func testAccSecurityGroupConfig_withDescription(name, description string) string {
	return acctest.ProviderConfig() + fmt.Sprintf(`
resource "acecloud_security_group" "test" {
  name        = %[1]q
  description = %[2]q
}
`, name, description)
}
