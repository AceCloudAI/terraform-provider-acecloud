package volume_attachment_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/AceCloudAI/terraform-provider-acecloud/internal/acctest"
	"github.com/AceCloudAI/terraform-provider-acecloud/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccVolumeAttachment_basic(t *testing.T) {
	rName := acctest.RandomName("va")
	flavorID := acctest.FlavorID(t)
	imageID := acctest.ImageID(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckVolumeAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeAttachmentConfig_basic(rName, flavorID, imageID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVolumeAttachmentExists("acecloud_volume_attachment.test"),
					resource.TestCheckResourceAttrSet("acecloud_volume_attachment.test", "id"),
					resource.TestCheckResourceAttrSet("acecloud_volume_attachment.test", "device"),
					resource.TestCheckResourceAttrPair("acecloud_volume_attachment.test", "instance_id", "acecloud_instance.test", "id"),
					resource.TestCheckResourceAttrPair("acecloud_volume_attachment.test", "volume_id", "acecloud_volume.test-data", "id"),
				),
			},
		},
	})
}

// --- Check functions ---

func testAccCheckVolumeAttachmentExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID is empty")
		}

		volumeID := rs.Primary.Attributes["volume_id"]
		instanceID := rs.Primary.Attributes["instance_id"]

		c := acctest.TestClient()
		path := fmt.Sprintf("/cloud/volumes/%s", volumeID)
		apiResp, err := c.Get(context.Background(), path, nil)
		if err != nil {
			return fmt.Errorf("volume %s not found: %w", volumeID, err)
		}

		var vol struct {
			Attachments []struct {
				ServerID string `json:"server_id"`
			} `json:"attachments"`
		}
		if err := json.Unmarshal(apiResp.Data, &vol); err != nil {
			return fmt.Errorf("failed to parse volume response: %w", err)
		}

		for _, a := range vol.Attachments {
			if a.ServerID == instanceID {
				return nil
			}
		}
		return fmt.Errorf("volume %s not attached to instance %s", volumeID, instanceID)
	}
}

func testAccCheckVolumeAttachmentDestroy(s *terraform.State) error {
	c := acctest.TestClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "acecloud_volume_attachment" {
			continue
		}
		volumeID := rs.Primary.Attributes["volume_id"]
		instanceID := rs.Primary.Attributes["instance_id"]

		path := fmt.Sprintf("/cloud/volumes/%s", volumeID)
		apiResp, err := c.Get(context.Background(), path, nil)
		if err != nil {
			if client.IsNotFound(err) {
				continue
			}
			return fmt.Errorf("unexpected error checking volume %s: %w", volumeID, err)
		}

		var vol struct {
			Attachments []struct {
				ServerID string `json:"server_id"`
			} `json:"attachments"`
		}
		if err := json.Unmarshal(apiResp.Data, &vol); err != nil {
			continue
		}

		for _, a := range vol.Attachments {
			if a.ServerID == instanceID {
				return fmt.Errorf("volume %s still attached to instance %s after destroy", volumeID, instanceID)
			}
		}
	}
	return nil
}

// --- Config functions ---

func testAccVolumeAttachmentConfig_basic(name, flavorID, imageID string) string {
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

resource "acecloud_instance" "test" {
  name                  = "%[1]s-inst"
  flavor_id             = %[2]q
  boot_uuid             = %[3]q
  source_type           = "image"
  delete_on_termination = true
  network_ids           = [acecloud_vpc.test.id]
  security_group_ids    = [acecloud_security_group.test.id]

  volumes {
    size        = 20
    boot        = true
    volume_type = "ssd"
  }
}

resource "acecloud_volume" "test-data" {
  name        = "%[1]s-data"
  size        = 10
  volume_type = "ssd"
}

resource "acecloud_volume_attachment" "test" {
  instance_id = acecloud_instance.test.id
  volume_id   = acecloud_volume.test-data.id
}
`, name, flavorID, imageID)
}
