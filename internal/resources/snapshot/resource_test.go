package snapshot_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/AceCloudAI/terraform-provider-acecloud/internal/acctest"
	"github.com/AceCloudAI/terraform-provider-acecloud/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccSnapshot_basic(t *testing.T) {
	rName := acctest.RandomName("snap")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSnapshotExists("acecloud_snapshot.test"),
					resource.TestCheckResourceAttr("acecloud_snapshot.test", "name", rName+"-snap"),
					resource.TestCheckResourceAttrSet("acecloud_snapshot.test", "id"),
					resource.TestCheckResourceAttrSet("acecloud_snapshot.test", "status"),
					resource.TestCheckResourceAttrSet("acecloud_snapshot.test", "size"),
					resource.TestCheckResourceAttrPair("acecloud_snapshot.test", "volume_id", "acecloud_volume.test", "id"),
				),
			},
		},
	})
}

func TestAccSnapshot_updateName(t *testing.T) {
	rName := acctest.RandomName("snap")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotConfig_named(rName, rName+"-snap"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSnapshotExists("acecloud_snapshot.test"),
					resource.TestCheckResourceAttr("acecloud_snapshot.test", "name", rName+"-snap"),
				),
			},
			{
				Config: testAccSnapshotConfig_named(rName, rName+"-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSnapshotExists("acecloud_snapshot.test"),
					resource.TestCheckResourceAttr("acecloud_snapshot.test", "name", rName+"-updated"),
				),
			},
		},
	})
}

func TestAccSnapshot_disappears(t *testing.T) {
	rName := acctest.RandomName("snap")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSnapshotExists("acecloud_snapshot.test"),
					testAccDeleteSnapshotOutOfBand("acecloud_snapshot.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// --- Check functions ---

func testAccCheckSnapshotExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID is empty")
		}

		c := acctest.TestClient()
		path := fmt.Sprintf("/os/cinder/%s/snapshots/%s", os.Getenv("ACECLOUD_PROJECT_ID"), rs.Primary.ID)
		_, err := c.Get(context.Background(), path, nil)
		if err != nil {
			return fmt.Errorf("snapshot %s not found via API: %w", rs.Primary.ID, err)
		}
		return nil
	}
}

func testAccCheckSnapshotDestroy(s *terraform.State) error {
	c := acctest.TestClient()
	projectID := os.Getenv("ACECLOUD_PROJECT_ID")
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "acecloud_snapshot" {
			continue
		}
		path := fmt.Sprintf("/os/cinder/%s/snapshots/%s", projectID, rs.Primary.ID)
		_, err := c.Get(context.Background(), path, nil)
		if err == nil {
			return fmt.Errorf("snapshot %s still exists after destroy", rs.Primary.ID)
		}
		if !client.IsNotFound(err) {
			return fmt.Errorf("unexpected error checking snapshot %s: %w", rs.Primary.ID, err)
		}
	}
	return nil
}

func testAccDeleteSnapshotOutOfBand(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		c := acctest.TestClient()
		projectID := os.Getenv("ACECLOUD_PROJECT_ID")
		deletePath := fmt.Sprintf("/os/cinder/%s/snapshots", projectID)
		body := map[string]interface{}{
			"key":    "id",
			"values": []string{rs.Primary.ID},
		}
		_, err := c.Delete(context.Background(), deletePath, body)
		if err != nil {
			return fmt.Errorf("failed to delete snapshot out-of-band: %w", err)
		}
		return nil
	}
}

// --- Config functions ---

func testAccSnapshotConfig_base(name string) string {
	return acctest.ProviderConfig() + fmt.Sprintf(`
resource "acecloud_volume" "test" {
  name        = "%[1]s-vol"
  size        = 10
  volume_type = "ssd"
}
`, name)
}

func testAccSnapshotConfig_basic(name string) string {
	return testAccSnapshotConfig_base(name) + fmt.Sprintf(`
resource "acecloud_snapshot" "test" {
  name      = "%[1]s-snap"
  volume_id = acecloud_volume.test.id
}
`, name)
}

func testAccSnapshotConfig_named(baseName, snapName string) string {
	return testAccSnapshotConfig_base(baseName) + fmt.Sprintf(`
resource "acecloud_snapshot" "test" {
  name      = %[1]q
  volume_id = acecloud_volume.test.id
}
`, snapName)
}
