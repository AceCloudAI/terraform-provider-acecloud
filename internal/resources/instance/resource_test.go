package instance_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/AceCloudAI/terraform-provider-acecloud/internal/acctest"
	"github.com/AceCloudAI/terraform-provider-acecloud/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccInstance_basic(t *testing.T) {
	rName := acctest.RandomName("inst")
	flavorID := acctest.FlavorID(t)
	imageID := acctest.ImageID(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_basic(rName, flavorID, imageID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists("acecloud_instance.test"),
					resource.TestCheckResourceAttr("acecloud_instance.test", "name", rName),
					resource.TestCheckResourceAttrSet("acecloud_instance.test", "id"),
					resource.TestCheckResourceAttr("acecloud_instance.test", "status", "ACTIVE"),
					resource.TestCheckResourceAttr("acecloud_instance.test", "power_state", "ON"),
					resource.TestCheckResourceAttrSet("acecloud_instance.test", "private_ip"),
				),
			},
		},
	})
}

func TestAccInstance_updateName(t *testing.T) {
	rName := acctest.RandomName("inst")
	rNameUpdated := acctest.RandomName("inst-upd")
	flavorID := acctest.FlavorID(t)
	imageID := acctest.ImageID(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_basic(rName, flavorID, imageID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists("acecloud_instance.test"),
					resource.TestCheckResourceAttr("acecloud_instance.test", "name", rName),
				),
			},
			{
				Config: testAccInstanceConfig_basic(rNameUpdated, flavorID, imageID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists("acecloud_instance.test"),
					resource.TestCheckResourceAttr("acecloud_instance.test", "name", rNameUpdated),
				),
			},
		},
	})
}

func TestAccInstance_powerState(t *testing.T) {
	rName := acctest.RandomName("inst")
	flavorID := acctest.FlavorID(t)
	imageID := acctest.ImageID(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_powerState(rName, flavorID, imageID, "ON"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists("acecloud_instance.test"),
					resource.TestCheckResourceAttr("acecloud_instance.test", "power_state", "ON"),
				),
			},
			{
				Config: testAccInstanceConfig_powerState(rName, flavorID, imageID, "OFF"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists("acecloud_instance.test"),
					resource.TestCheckResourceAttr("acecloud_instance.test", "power_state", "OFF"),
				),
			},
			{
				Config: testAccInstanceConfig_powerState(rName, flavorID, imageID, "ON"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("acecloud_instance.test", "power_state", "ON"),
				),
			},
		},
	})
}

func TestAccInstance_disappears(t *testing.T) {
	rName := acctest.RandomName("inst")
	flavorID := acctest.FlavorID(t)
	imageID := acctest.ImageID(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_basic(rName, flavorID, imageID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists("acecloud_instance.test"),
					testAccDeleteInstanceOutOfBand("acecloud_instance.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// --- Check functions ---

func testAccCheckInstanceExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID is empty")
		}

		c := acctest.TestClient()
		path := fmt.Sprintf("/cloud/instances/%s", rs.Primary.ID)
		_, err := c.Get(context.Background(), path, nil)
		if err != nil {
			return fmt.Errorf("instance %s not found via API: %w", rs.Primary.ID, err)
		}
		return nil
	}
}

func testAccCheckInstanceDestroy(s *terraform.State) error {
	c := acctest.TestClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "acecloud_instance" {
			continue
		}
		path := fmt.Sprintf("/cloud/instances/%s", rs.Primary.ID)
		_, err := c.Get(context.Background(), path, nil)
		if err == nil {
			return fmt.Errorf("instance %s still exists after destroy", rs.Primary.ID)
		}
		if !client.IsNotFound(err) {
			return fmt.Errorf("unexpected error checking instance %s: %w", rs.Primary.ID, err)
		}
	}
	return nil
}

func testAccDeleteInstanceOutOfBand(resourceName string) resource.TestCheckFunc {
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
		_, err := c.Delete(context.Background(), "/cloud/instances", body)
		if err != nil {
			return fmt.Errorf("failed to delete instance out-of-band: %w", err)
		}
		return nil
	}
}

// --- Config functions ---

func testAccInstanceConfig_base(name string) string {
	return acctest.ProviderConfig() + fmt.Sprintf(`
resource "acecloud_vpc" "test" {
  name              = "%[1]s-vpc"
  subnet_name       = "%[1]s-subnet"
  subnet_cidr       = "10.0.0.0/16"
  subnet_ip_version = 4
}

resource "acecloud_security_group" "test" {
  name = "%[1]s-sg"

  rules {
    direction        = "ingress"
    protocol         = "tcp"
    port_range_min   = 22
    port_range_max   = 22
    remote_ip_prefix = "0.0.0.0/0"
    ethertype        = "IPv4"
  }
}
`, name)
}

func testAccInstanceConfig_basic(name, flavorID, imageID string) string {
	return testAccInstanceConfig_base(name) + fmt.Sprintf(`
resource "acecloud_instance" "test" {
  name                = %[1]q
  flavor_id           = %[2]q
  boot_uuid           = %[3]q
  source_type         = "image"
  delete_on_termination = true
  network_ids         = [acecloud_vpc.test.id]
  security_group_ids  = [acecloud_security_group.test.id]

  volumes {
    size        = 20
    boot        = true
    volume_type = "ssd"
  }
}
`, name, flavorID, imageID)
}

func testAccInstanceConfig_powerState(name, flavorID, imageID, powerState string) string {
	return testAccInstanceConfig_base(name) + fmt.Sprintf(`
resource "acecloud_instance" "test" {
  name                = %[1]q
  flavor_id           = %[2]q
  boot_uuid           = %[3]q
  source_type         = "image"
  delete_on_termination = true
  network_ids         = [acecloud_vpc.test.id]
  security_group_ids  = [acecloud_security_group.test.id]
  power_state         = %[4]q

  volumes {
    size        = 20
    boot        = true
    volume_type = "ssd"
  }
}
`, name, flavorID, imageID, powerState)
}
