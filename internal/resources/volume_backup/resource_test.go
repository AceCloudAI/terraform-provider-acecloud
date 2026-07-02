package volume_backup_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/AceCloudAI/terraform-provider-acecloud/internal/acctest"
	"github.com/AceCloudAI/terraform-provider-acecloud/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccVolumeBackup_basic(t *testing.T) {
	rName := acctest.RandomName("bkp")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckVolumeBackupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeBackupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVolumeBackupExists("acecloud_volume_backup.test"),
					resource.TestCheckResourceAttr("acecloud_volume_backup.test", "name", rName+"-bkp"),
					resource.TestCheckResourceAttrSet("acecloud_volume_backup.test", "id"),
					resource.TestCheckResourceAttrSet("acecloud_volume_backup.test", "status"),
					resource.TestCheckResourceAttrSet("acecloud_volume_backup.test", "size"),
					resource.TestCheckResourceAttrPair("acecloud_volume_backup.test", "volume_id", "acecloud_volume.test", "id"),
				),
			},
		},
	})
}

func TestAccVolumeBackup_updateName(t *testing.T) {
	rName := acctest.RandomName("bkp")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckVolumeBackupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeBackupConfig_named(rName, rName+"-bkp"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVolumeBackupExists("acecloud_volume_backup.test"),
					resource.TestCheckResourceAttr("acecloud_volume_backup.test", "name", rName+"-bkp"),
				),
			},
			{
				Config: testAccVolumeBackupConfig_named(rName, rName+"-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVolumeBackupExists("acecloud_volume_backup.test"),
					resource.TestCheckResourceAttr("acecloud_volume_backup.test", "name", rName+"-updated"),
				),
			},
		},
	})
}

func TestAccVolumeBackup_disappears(t *testing.T) {
	rName := acctest.RandomName("bkp")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckVolumeBackupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeBackupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVolumeBackupExists("acecloud_volume_backup.test"),
					testAccDeleteVolumeBackupOutOfBand("acecloud_volume_backup.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// --- Check functions ---

func testAccCheckVolumeBackupExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID is empty")
		}

		c := acctest.TestClient()
		path := fmt.Sprintf("/cloud/volume-backups/%s", rs.Primary.ID)
		_, err := c.Get(context.Background(), path, nil)
		if err != nil {
			return fmt.Errorf("volume backup %s not found via API: %w", rs.Primary.ID, err)
		}
		return nil
	}
}

func testAccCheckVolumeBackupDestroy(s *terraform.State) error {
	c := acctest.TestClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "acecloud_volume_backup" {
			continue
		}
		path := fmt.Sprintf("/cloud/volume-backups/%s", rs.Primary.ID)
		_, err := c.Get(context.Background(), path, nil)
		if err == nil {
			return fmt.Errorf("volume backup %s still exists after destroy", rs.Primary.ID)
		}
		if !client.IsNotFound(err) {
			return fmt.Errorf("unexpected error checking volume backup %s: %w", rs.Primary.ID, err)
		}
	}
	return nil
}

func testAccDeleteVolumeBackupOutOfBand(resourceName string) resource.TestCheckFunc {
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
		_, err := c.Delete(context.Background(), "/backups", body)
		if err != nil {
			return fmt.Errorf("failed to delete volume backup out-of-band: %w", err)
		}
		return nil
	}
}

// --- Config functions ---

func testAccVolumeBackupConfig_base(name string) string {
	return acctest.ProviderConfig() + fmt.Sprintf(`
resource "acecloud_volume" "test" {
  name        = "%[1]s-vol"
  size        = 10
  volume_type = "ssd"
}
`, name)
}

func testAccVolumeBackupConfig_basic(name string) string {
	return testAccVolumeBackupConfig_base(name) + fmt.Sprintf(`
resource "acecloud_volume_backup" "test" {
  name      = "%[1]s-bkp"
  volume_id = acecloud_volume.test.id
}
`, name)
}

func testAccVolumeBackupConfig_named(baseName, backupName string) string {
	return testAccVolumeBackupConfig_base(baseName) + fmt.Sprintf(`
resource "acecloud_volume_backup" "test" {
  name      = %[1]q
  volume_id = acecloud_volume.test.id
}
`, backupName)
}
