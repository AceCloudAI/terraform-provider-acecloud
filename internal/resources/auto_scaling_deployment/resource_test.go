package auto_scaling_deployment_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/AceCloudAI/terraform-provider-acecloud/internal/acctest"
	"github.com/AceCloudAI/terraform-provider-acecloud/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccAutoScalingDeployment_basic(t *testing.T) {
	flavorID := acctest.FlavorID(t)
	imageID := acctest.ImageID(t)
	rName := acctest.RandomName("asd")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckAutoScalingDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAutoScalingDeploymentConfig_basic(rName, flavorID, imageID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutoScalingDeploymentExists("acecloud_auto_scaling_deployment.test"),
					resource.TestCheckResourceAttr("acecloud_auto_scaling_deployment.test", "name", rName+"-deploy"),
					resource.TestCheckResourceAttrSet("acecloud_auto_scaling_deployment.test", "id"),
					resource.TestCheckResourceAttr("acecloud_auto_scaling_deployment.test", "desired_capacity", "1"),
					resource.TestCheckResourceAttr("acecloud_auto_scaling_deployment.test", "max_capacity", "3"),
					resource.TestCheckResourceAttr("acecloud_auto_scaling_deployment.test", "nodes_scale_count", "1"),
					resource.TestCheckResourceAttr("acecloud_auto_scaling_deployment.test", "scaling_parameter", "cpu"),
					resource.TestCheckResourceAttr("acecloud_auto_scaling_deployment.test", "min_threshold", "30"),
					resource.TestCheckResourceAttr("acecloud_auto_scaling_deployment.test", "max_threshold", "80"),
					resource.TestCheckResourceAttr("acecloud_auto_scaling_deployment.test", "cool_down_time", "300"),
					resource.TestCheckResourceAttr("acecloud_auto_scaling_deployment.test", "is_integrated_with_lb", "false"),
					resource.TestCheckResourceAttrPair(
						"acecloud_auto_scaling_deployment.test", "template_id",
						"acecloud_auto_scaling_template.test", "id",
					),
				),
			},
		},
	})
}

func TestAccAutoScalingDeployment_update(t *testing.T) {
	flavorID := acctest.FlavorID(t)
	imageID := acctest.ImageID(t)
	rName := acctest.RandomName("asd")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckAutoScalingDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAutoScalingDeploymentConfig_basic(rName, flavorID, imageID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutoScalingDeploymentExists("acecloud_auto_scaling_deployment.test"),
					resource.TestCheckResourceAttr("acecloud_auto_scaling_deployment.test", "desired_capacity", "1"),
					resource.TestCheckResourceAttr("acecloud_auto_scaling_deployment.test", "max_capacity", "3"),
					resource.TestCheckResourceAttr("acecloud_auto_scaling_deployment.test", "cool_down_time", "300"),
				),
			},
			{
				Config: testAccAutoScalingDeploymentConfig_updated(rName, flavorID, imageID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutoScalingDeploymentExists("acecloud_auto_scaling_deployment.test"),
					resource.TestCheckResourceAttr("acecloud_auto_scaling_deployment.test", "desired_capacity", "2"),
					resource.TestCheckResourceAttr("acecloud_auto_scaling_deployment.test", "max_capacity", "5"),
					resource.TestCheckResourceAttr("acecloud_auto_scaling_deployment.test", "cool_down_time", "600"),
					resource.TestCheckResourceAttr("acecloud_auto_scaling_deployment.test", "scaling_parameter", "ram"),
					resource.TestCheckResourceAttr("acecloud_auto_scaling_deployment.test", "description", "updated deployment"),
				),
			},
		},
	})
}

func TestAccAutoScalingDeployment_disappears(t *testing.T) {
	flavorID := acctest.FlavorID(t)
	imageID := acctest.ImageID(t)
	rName := acctest.RandomName("asd")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckAutoScalingDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAutoScalingDeploymentConfig_basic(rName, flavorID, imageID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutoScalingDeploymentExists("acecloud_auto_scaling_deployment.test"),
					testAccDeleteAutoScalingDeploymentOutOfBand("acecloud_auto_scaling_deployment.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// --- Check functions ---

func testAccCheckAutoScalingDeploymentExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID is empty")
		}

		c := acctest.TestClient()
		path := fmt.Sprintf("/auto-scaling/deployments/%s", rs.Primary.ID)
		_, err := c.Get(context.Background(), path, nil)
		if err != nil {
			return fmt.Errorf("auto scaling deployment %s not found via API: %w", rs.Primary.ID, err)
		}
		return nil
	}
}

func testAccCheckAutoScalingDeploymentDestroy(s *terraform.State) error {
	c := acctest.TestClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "acecloud_auto_scaling_deployment" {
			continue
		}
		path := fmt.Sprintf("/auto-scaling/deployments/%s", rs.Primary.ID)
		_, err := c.Get(context.Background(), path, nil)
		if err == nil {
			return fmt.Errorf("auto scaling deployment %s still exists after destroy", rs.Primary.ID)
		}
		if !client.IsNotFound(err) {
			return fmt.Errorf("unexpected error checking auto scaling deployment %s: %w", rs.Primary.ID, err)
		}
	}
	return nil
}

func testAccDeleteAutoScalingDeploymentOutOfBand(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		c := acctest.TestClient()
		path := fmt.Sprintf("/auto-scaling/deployments/%s", rs.Primary.ID)
		_, err := c.Delete(context.Background(), path, nil)
		if err != nil {
			return fmt.Errorf("failed to delete auto scaling deployment out-of-band: %w", err)
		}
		return nil
	}
}

// --- Config functions ---

// testAccAutoScalingDeploymentConfig_baseTemplate returns the shared
// infrastructure for deployment tests: VPC + subnet + SG + key pair + template.
func testAccAutoScalingDeploymentConfig_baseTemplate(name, flavorID, imageID string) string {
	return acctest.ConfigCompose(
		acctest.TestAccBaseVPCConfig(name),
		acctest.TestAccBaseSecurityGroupConfig(name+"-sg"),
		fmt.Sprintf(`
resource "acecloud_key_pair" "test" {
  name = %[1]q
}

resource "acecloud_auto_scaling_template" "test" {
  name                   = %[1]q
  type                   = "linux"
  volume_type            = "ssd"
  volume_size            = 20
  vol_del_on_termination = true
  flavor_id              = %[2]q
  image_id               = %[3]q
  key_name               = acecloud_key_pair.test.name
  network_id             = acecloud_vpc.test-base.id
  subnet_id              = acecloud_vpc.test-base.subnet_id
  security_groups        = [acecloud_security_group.test-base.id]
  is_instance_snapshot   = false
}
`, name, flavorID, imageID),
	)
}

func testAccAutoScalingDeploymentConfig_basic(name, flavorID, imageID string) string {
	return acctest.ConfigCompose(
		testAccAutoScalingDeploymentConfig_baseTemplate(name, flavorID, imageID),
		fmt.Sprintf(`
resource "acecloud_auto_scaling_deployment" "test" {
  name                  = "%[1]s-deploy"
  template_id           = acecloud_auto_scaling_template.test.id
  desired_capacity      = 1
  max_capacity          = 3
  nodes_scale_count     = 1
  scaling_parameter     = "cpu"
  min_threshold         = 30
  max_threshold         = 80
  cool_down_time        = 300
  user_email            = ["test@example.com"]
  is_integrated_with_lb = false
}
`, name),
	)
}

func testAccAutoScalingDeploymentConfig_updated(name, flavorID, imageID string) string {
	return acctest.ConfigCompose(
		testAccAutoScalingDeploymentConfig_baseTemplate(name, flavorID, imageID),
		fmt.Sprintf(`
resource "acecloud_auto_scaling_deployment" "test" {
  name                  = "%[1]s-deploy"
  description           = "updated deployment"
  template_id           = acecloud_auto_scaling_template.test.id
  desired_capacity      = 2
  max_capacity          = 5
  nodes_scale_count     = 1
  scaling_parameter     = "ram"
  min_threshold         = 40
  max_threshold         = 85
  cool_down_time        = 600
  user_email            = ["test@example.com"]
  is_integrated_with_lb = false
}
`, name),
	)
}
