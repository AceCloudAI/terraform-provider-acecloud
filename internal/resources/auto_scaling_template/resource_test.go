package auto_scaling_template_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/AceCloudAI/terraform-provider-acecloud/internal/acctest"
	"github.com/AceCloudAI/terraform-provider-acecloud/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccAutoScalingTemplate_basic(t *testing.T) {
	flavorID := acctest.FlavorID(t)
	imageID := acctest.ImageID(t)
	rName := acctest.RandomName("ast")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckAutoScalingTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAutoScalingTemplateConfig_basic(rName, flavorID, imageID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutoScalingTemplateExists("acecloud_auto_scaling_template.test"),
					resource.TestCheckResourceAttr("acecloud_auto_scaling_template.test", "name", rName),
					resource.TestCheckResourceAttrSet("acecloud_auto_scaling_template.test", "id"),
					resource.TestCheckResourceAttr("acecloud_auto_scaling_template.test", "type", "linux"),
					resource.TestCheckResourceAttr("acecloud_auto_scaling_template.test", "volume_type", "ssd"),
					resource.TestCheckResourceAttr("acecloud_auto_scaling_template.test", "volume_size", "20"),
					resource.TestCheckResourceAttr("acecloud_auto_scaling_template.test", "vol_del_on_termination", "true"),
					resource.TestCheckResourceAttr("acecloud_auto_scaling_template.test", "flavor_id", flavorID),
					resource.TestCheckResourceAttr("acecloud_auto_scaling_template.test", "image_id", imageID),
					resource.TestCheckResourceAttr("acecloud_auto_scaling_template.test", "is_instance_snapshot", "false"),
					resource.TestCheckResourceAttrSet("acecloud_auto_scaling_template.test", "status"),
				),
			},
		},
	})
}

func TestAccAutoScalingTemplate_update(t *testing.T) {
	flavorID := acctest.FlavorID(t)
	imageID := acctest.ImageID(t)
	rName := acctest.RandomName("ast")
	rNameUpdated := acctest.RandomName("ast-upd")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckAutoScalingTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAutoScalingTemplateConfig_basic(rName, flavorID, imageID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutoScalingTemplateExists("acecloud_auto_scaling_template.test"),
					resource.TestCheckResourceAttr("acecloud_auto_scaling_template.test", "name", rName),
					resource.TestCheckResourceAttr("acecloud_auto_scaling_template.test", "volume_size", "20"),
				),
			},
			{
				Config: testAccAutoScalingTemplateConfig_updated(rNameUpdated, flavorID, imageID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutoScalingTemplateExists("acecloud_auto_scaling_template.test"),
					resource.TestCheckResourceAttr("acecloud_auto_scaling_template.test", "name", rNameUpdated),
					resource.TestCheckResourceAttr("acecloud_auto_scaling_template.test", "volume_size", "40"),
					resource.TestCheckResourceAttr("acecloud_auto_scaling_template.test", "description", "updated template"),
				),
			},
		},
	})
}

func TestAccAutoScalingTemplate_disappears(t *testing.T) {
	flavorID := acctest.FlavorID(t)
	imageID := acctest.ImageID(t)
	rName := acctest.RandomName("ast")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckAutoScalingTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAutoScalingTemplateConfig_basic(rName, flavorID, imageID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutoScalingTemplateExists("acecloud_auto_scaling_template.test"),
					testAccDeleteAutoScalingTemplateOutOfBand("acecloud_auto_scaling_template.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// --- Check functions ---

func testAccCheckAutoScalingTemplateExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID is empty")
		}

		c := acctest.TestClient()
		path := fmt.Sprintf("/auto-scaling/templates/%s", rs.Primary.ID)
		_, err := c.Get(context.Background(), path, nil)
		if err != nil {
			return fmt.Errorf("auto scaling template %s not found via API: %w", rs.Primary.ID, err)
		}
		return nil
	}
}

func testAccCheckAutoScalingTemplateDestroy(s *terraform.State) error {
	c := acctest.TestClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "acecloud_auto_scaling_template" {
			continue
		}
		path := fmt.Sprintf("/auto-scaling/templates/%s", rs.Primary.ID)
		_, err := c.Get(context.Background(), path, nil)
		if err == nil {
			return fmt.Errorf("auto scaling template %s still exists after destroy", rs.Primary.ID)
		}
		if !client.IsNotFound(err) {
			return fmt.Errorf("unexpected error checking auto scaling template %s: %w", rs.Primary.ID, err)
		}
	}
	return nil
}

func testAccDeleteAutoScalingTemplateOutOfBand(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		c := acctest.TestClient()
		path := fmt.Sprintf("/auto-scaling/templates/%s", rs.Primary.ID)
		_, err := c.Delete(context.Background(), path, nil)
		if err != nil {
			return fmt.Errorf("failed to delete auto scaling template out-of-band: %w", err)
		}
		return nil
	}
}

// --- Config functions ---

func testAccAutoScalingTemplateConfig_basic(name, flavorID, imageID string) string {
	return acctest.ConfigCompose(
		acctest.TestAccBaseVPCConfig(name),
		acctest.TestAccBaseSecurityGroupConfig(name+"-sg"),
		fmt.Sprintf(`
resource "acecloud_key_pair" "test" {
  name = %[1]q
}

resource "acecloud_auto_scaling_template" "test" {
  name                  = %[1]q
  type                  = "linux"
  volume_type           = "ssd"
  volume_size           = 20
  vol_del_on_termination = true
  flavor_id             = %[2]q
  image_id              = %[3]q
  key_name              = acecloud_key_pair.test.name
  network_id            = acecloud_vpc.test-base.id
  subnet_id             = acecloud_vpc.test-base.subnet_id
  security_groups       = [acecloud_security_group.test-base.id]
  is_instance_snapshot  = false
}
`, name, flavorID, imageID),
	)
}

func testAccAutoScalingTemplateConfig_updated(name, flavorID, imageID string) string {
	return acctest.ConfigCompose(
		acctest.TestAccBaseVPCConfig(name),
		acctest.TestAccBaseSecurityGroupConfig(name+"-sg"),
		fmt.Sprintf(`
resource "acecloud_key_pair" "test" {
  name = %[1]q
}

resource "acecloud_auto_scaling_template" "test" {
  name                  = %[1]q
  type                  = "linux"
  description           = "updated template"
  volume_type           = "ssd"
  volume_size           = 40
  vol_del_on_termination = true
  flavor_id             = %[2]q
  image_id              = %[3]q
  key_name              = acecloud_key_pair.test.name
  network_id            = acecloud_vpc.test-base.id
  subnet_id             = acecloud_vpc.test-base.subnet_id
  security_groups       = [acecloud_security_group.test-base.id]
  is_instance_snapshot  = false
}
`, name, flavorID, imageID),
	)
}
