package vpc_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/AceCloudAI/terraform-provider-acecloud/internal/acctest"
	"github.com/AceCloudAI/terraform-provider-acecloud/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccVPC_basic(t *testing.T) {
	rName := acctest.RandomName("vpc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckVPCDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCExists("acecloud_vpc.test"),
					resource.TestCheckResourceAttr("acecloud_vpc.test", "name", rName),
					resource.TestCheckResourceAttrSet("acecloud_vpc.test", "id"),
					resource.TestCheckResourceAttrSet("acecloud_vpc.test", "subnet_id"),
					resource.TestCheckResourceAttr("acecloud_vpc.test", "subnet_cidr", "10.0.0.0/16"),
					resource.TestCheckResourceAttr("acecloud_vpc.test", "subnet_ip_version", "4"),
					resource.TestCheckResourceAttr("acecloud_vpc.test", "admin_state_up", "true"),
				),
			},
		},
	})
}

func TestAccVPC_update(t *testing.T) {
	rName := acctest.RandomName("vpc")
	rNameUpdated := acctest.RandomName("vpc-upd")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckVPCDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCExists("acecloud_vpc.test"),
					resource.TestCheckResourceAttr("acecloud_vpc.test", "name", rName),
				),
			},
			{
				Config: testAccVPCConfig_basic(rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCExists("acecloud_vpc.test"),
					resource.TestCheckResourceAttr("acecloud_vpc.test", "name", rNameUpdated),
				),
			},
		},
	})
}

func TestAccVPC_withDescription(t *testing.T) {
	rName := acctest.RandomName("vpc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckVPCDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_withDescription(rName, "initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCExists("acecloud_vpc.test"),
					resource.TestCheckResourceAttr("acecloud_vpc.test", "name", rName),
					resource.TestCheckResourceAttr("acecloud_vpc.test", "description", "initial description"),
				),
			},
			{
				Config: testAccVPCConfig_withDescription(rName, "updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("acecloud_vpc.test", "description", "updated description"),
				),
			},
		},
	})
}

func TestAccVPC_disappears(t *testing.T) {
	rName := acctest.RandomName("vpc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckVPCDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCExists("acecloud_vpc.test"),
					testAccDeleteVPCOutOfBand("acecloud_vpc.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// --- Check functions ---

func testAccCheckVPCExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID is empty")
		}

		c := acctest.TestClient()
		path := fmt.Sprintf("/cloud/vpcs/%s", rs.Primary.ID)
		_, err := c.Get(context.Background(), path, nil)
		if err != nil {
			return fmt.Errorf("VPC %s not found via API: %w", rs.Primary.ID, err)
		}
		return nil
	}
}

func testAccCheckVPCDestroy(s *terraform.State) error {
	c := acctest.TestClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "acecloud_vpc" {
			continue
		}
		path := fmt.Sprintf("/cloud/vpcs/%s", rs.Primary.ID)
		_, err := c.Get(context.Background(), path, nil)
		if err == nil {
			return fmt.Errorf("VPC %s still exists after destroy", rs.Primary.ID)
		}
		if !client.IsNotFound(err) {
			return fmt.Errorf("unexpected error checking VPC %s: %w", rs.Primary.ID, err)
		}
	}
	return nil
}

func testAccDeleteVPCOutOfBand(resourceName string) resource.TestCheckFunc {
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
		_, err := c.Delete(context.Background(), "/cloud/vpcs", body)
		if err != nil {
			return fmt.Errorf("failed to delete VPC out-of-band: %w", err)
		}
		return nil
	}
}

// --- Config functions ---

func testAccVPCConfig_basic(name string) string {
	return acctest.ProviderConfig() + fmt.Sprintf(`
resource "acecloud_vpc" "test" {
  name              = %[1]q
  subnet_name       = "%[1]s-subnet"
  subnet_cidr       = "10.0.0.0/16"
  subnet_ip_version = 4
}
`, name)
}

func testAccVPCConfig_withDescription(name, description string) string {
	return acctest.ProviderConfig() + fmt.Sprintf(`
resource "acecloud_vpc" "test" {
  name              = %[1]q
  description       = %[2]q
  subnet_name       = "%[1]s-subnet"
  subnet_cidr       = "10.0.0.0/16"
  subnet_ip_version = 4
}
`, name, description)
}
