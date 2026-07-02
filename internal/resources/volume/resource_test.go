package volume_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/AceCloudAI/terraform-provider-acecloud/internal/acctest"
	"github.com/AceCloudAI/terraform-provider-acecloud/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccVolume_basic(t *testing.T) {
	rName := acctest.RandomName("vol")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVolumeExists("acecloud_volume.test"),
					resource.TestCheckResourceAttr("acecloud_volume.test", "name", rName),
					resource.TestCheckResourceAttr("acecloud_volume.test", "size", "10"),
					resource.TestCheckResourceAttrSet("acecloud_volume.test", "id"),
					resource.TestCheckResourceAttrSet("acecloud_volume.test", "status"),
				),
			},
		},
	})
}

func TestAccVolume_updateName(t *testing.T) {
	rName := acctest.RandomName("vol")
	rNameUpdated := acctest.RandomName("vol-upd")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVolumeExists("acecloud_volume.test"),
					resource.TestCheckResourceAttr("acecloud_volume.test", "name", rName),
				),
			},
			{
				Config: testAccVolumeConfig_basic(rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVolumeExists("acecloud_volume.test"),
					resource.TestCheckResourceAttr("acecloud_volume.test", "name", rNameUpdated),
				),
			},
		},
	})
}

func TestAccVolume_extend(t *testing.T) {
	rName := acctest.RandomName("vol")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeConfig_size(rName, 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVolumeExists("acecloud_volume.test"),
					resource.TestCheckResourceAttr("acecloud_volume.test", "size", "10"),
				),
			},
			{
				Config: testAccVolumeConfig_size(rName, 20),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVolumeExists("acecloud_volume.test"),
					resource.TestCheckResourceAttr("acecloud_volume.test", "size", "20"),
				),
			},
		},
	})
}

func TestAccVolume_withDescription(t *testing.T) {
	rName := acctest.RandomName("vol")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeConfig_withDescription(rName, "test volume"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVolumeExists("acecloud_volume.test"),
					resource.TestCheckResourceAttr("acecloud_volume.test", "description", "test volume"),
				),
			},
			{
				Config: testAccVolumeConfig_withDescription(rName, "updated volume"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("acecloud_volume.test", "description", "updated volume"),
				),
			},
		},
	})
}

func TestAccVolume_disappears(t *testing.T) {
	rName := acctest.RandomName("vol")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVolumeExists("acecloud_volume.test"),
					testAccDeleteVolumeOutOfBand("acecloud_volume.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// --- Check functions ---

func testAccCheckVolumeExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID is empty")
		}

		c := acctest.TestClient()
		path := fmt.Sprintf("/cloud/volumes/%s", rs.Primary.ID)
		_, err := c.Get(context.Background(), path, nil)
		if err != nil {
			return fmt.Errorf("volume %s not found via API: %w", rs.Primary.ID, err)
		}
		return nil
	}
}

func testAccCheckVolumeDestroy(s *terraform.State) error {
	c := acctest.TestClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "acecloud_volume" {
			continue
		}
		path := fmt.Sprintf("/cloud/volumes/%s", rs.Primary.ID)
		_, err := c.Get(context.Background(), path, nil)
		if err == nil {
			return fmt.Errorf("volume %s still exists after destroy", rs.Primary.ID)
		}
		if !client.IsNotFound(err) {
			return fmt.Errorf("unexpected error checking volume %s: %w", rs.Primary.ID, err)
		}
	}
	return nil
}

func testAccDeleteVolumeOutOfBand(resourceName string) resource.TestCheckFunc {
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
		_, err := c.Delete(context.Background(), "/cloud/volumes", body)
		if err != nil {
			return fmt.Errorf("failed to delete volume out-of-band: %w", err)
		}
		return nil
	}
}

// --- Config functions ---

func testAccVolumeConfig_basic(name string) string {
	return acctest.ProviderConfig() + fmt.Sprintf(`
resource "acecloud_volume" "test" {
  name        = %[1]q
  size        = 10
  volume_type = "ssd"
}
`, name)
}

func testAccVolumeConfig_size(name string, size int) string {
	return acctest.ProviderConfig() + fmt.Sprintf(`
resource "acecloud_volume" "test" {
  name        = %[1]q
  size        = %[2]d
  volume_type = "ssd"
}
`, name, size)
}

func testAccVolumeConfig_withDescription(name, description string) string {
	return acctest.ProviderConfig() + fmt.Sprintf(`
resource "acecloud_volume" "test" {
  name        = %[1]q
  size        = 10
  volume_type = "ssd"
  description = %[2]q
}
`, name, description)
}
