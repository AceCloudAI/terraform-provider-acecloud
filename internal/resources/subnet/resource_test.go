package subnet_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/AceCloudAI/terraform-provider-acecloud/internal/acctest"
	"github.com/AceCloudAI/terraform-provider-acecloud/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccSubnet_basic(t *testing.T) {
	rName := acctest.RandomName("subnet")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubnetExists("acecloud_subnet.test"),
					resource.TestCheckResourceAttr("acecloud_subnet.test", "name", rName+"-extra"),
					resource.TestCheckResourceAttr("acecloud_subnet.test", "cidr", "10.0.1.0/24"),
					resource.TestCheckResourceAttr("acecloud_subnet.test", "ip_version", "4"),
					resource.TestCheckResourceAttrSet("acecloud_subnet.test", "id"),
					resource.TestCheckResourceAttrSet("acecloud_subnet.test", "vpc_id"),
				),
			},
		},
	})
}

func TestAccSubnet_update(t *testing.T) {
	rName := acctest.RandomName("subnet")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubnetExists("acecloud_subnet.test"),
					resource.TestCheckResourceAttr("acecloud_subnet.test", "name", rName+"-extra"),
				),
			},
			{
				Config: testAccSubnetConfig_updateName(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubnetExists("acecloud_subnet.test"),
					resource.TestCheckResourceAttr("acecloud_subnet.test", "name", rName+"-updated"),
				),
			},
		},
	})
}

func TestAccSubnet_withDNS(t *testing.T) {
	rName := acctest.RandomName("subnet")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetConfig_withDNS(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubnetExists("acecloud_subnet.test"),
					resource.TestCheckResourceAttr("acecloud_subnet.test", "dns_nameservers.#", "2"),
					resource.TestCheckResourceAttr("acecloud_subnet.test", "dns_nameservers.0", "8.8.8.8"),
					resource.TestCheckResourceAttr("acecloud_subnet.test", "dns_nameservers.1", "8.8.4.4"),
				),
			},
		},
	})
}

func TestAccSubnet_disappears(t *testing.T) {
	rName := acctest.RandomName("subnet")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubnetExists("acecloud_subnet.test"),
					testAccDeleteSubnetOutOfBand("acecloud_subnet.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// --- Check functions ---

func testAccCheckSubnetExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID is empty")
		}

		c := acctest.TestClient()
		path := fmt.Sprintf("/cloud/vpcs/subnets/%s", rs.Primary.ID)
		_, err := c.Get(context.Background(), path, nil)
		if err != nil {
			return fmt.Errorf("subnet %s not found via API: %w", rs.Primary.ID, err)
		}
		return nil
	}
}

func testAccCheckSubnetDestroy(s *terraform.State) error {
	c := acctest.TestClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "acecloud_subnet" {
			continue
		}
		path := fmt.Sprintf("/cloud/vpcs/subnets/%s", rs.Primary.ID)
		_, err := c.Get(context.Background(), path, nil)
		if err == nil {
			return fmt.Errorf("subnet %s still exists after destroy", rs.Primary.ID)
		}
		if !client.IsNotFound(err) {
			return fmt.Errorf("unexpected error checking subnet %s: %w", rs.Primary.ID, err)
		}
	}
	return nil
}

func testAccDeleteSubnetOutOfBand(resourceName string) resource.TestCheckFunc {
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
		_, err := c.Delete(context.Background(), "/cloud/vpcs/subnets", body)
		if err != nil {
			return fmt.Errorf("failed to delete subnet out-of-band: %w", err)
		}
		return nil
	}
}

// --- Config functions ---

func testAccSubnetConfig_base(name string) string {
	return acctest.ProviderConfig() + fmt.Sprintf(`
resource "acecloud_vpc" "test" {
  name              = %[1]q
  subnet_name       = "%[1]s-subnet"
  subnet_cidr       = "10.0.0.0/16"
  subnet_ip_version = 4
}
`, name)
}

func testAccSubnetConfig_basic(name string) string {
	return testAccSubnetConfig_base(name) + fmt.Sprintf(`
resource "acecloud_subnet" "test" {
  name       = "%[1]s-extra"
  cidr       = "10.0.1.0/24"
  vpc_id     = acecloud_vpc.test.id
  ip_version = 4
}
`, name)
}

func testAccSubnetConfig_updateName(name string) string {
	return testAccSubnetConfig_base(name) + fmt.Sprintf(`
resource "acecloud_subnet" "test" {
  name       = "%[1]s-updated"
  cidr       = "10.0.1.0/24"
  vpc_id     = acecloud_vpc.test.id
  ip_version = 4
}
`, name)
}

func testAccSubnetConfig_withDNS(name string) string {
	return testAccSubnetConfig_base(name) + fmt.Sprintf(`
resource "acecloud_subnet" "test" {
  name             = "%[1]s-dns"
  cidr             = "10.0.2.0/24"
  vpc_id           = acecloud_vpc.test.id
  ip_version       = 4
  dns_nameservers  = ["8.8.8.8", "8.8.4.4"]
}
`, name)
}
