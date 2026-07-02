package key_pair_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/AceCloudAI/terraform-provider-acecloud/internal/acctest"
	"github.com/AceCloudAI/terraform-provider-acecloud/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccKeyPair_basic(t *testing.T) {
	rName := acctest.RandomName("kp")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckKeyPairDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPairConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeyPairExists("acecloud_key_pair.test"),
					resource.TestCheckResourceAttr("acecloud_key_pair.test", "name", rName),
					resource.TestCheckResourceAttrSet("acecloud_key_pair.test", "id"),
					resource.TestCheckResourceAttrSet("acecloud_key_pair.test", "fingerprint"),
					resource.TestCheckResourceAttrSet("acecloud_key_pair.test", "private_key"),
				),
			},
		},
	})
}

func TestAccKeyPair_recreate(t *testing.T) {
	rName1 := acctest.RandomName("kp")
	rName2 := acctest.RandomName("kp")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckKeyPairDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPairConfig_basic(rName1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeyPairExists("acecloud_key_pair.test"),
					resource.TestCheckResourceAttr("acecloud_key_pair.test", "name", rName1),
				),
			},
			{
				Config: testAccKeyPairConfig_basic(rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeyPairExists("acecloud_key_pair.test"),
					resource.TestCheckResourceAttr("acecloud_key_pair.test", "name", rName2),
				),
			},
		},
	})
}

func TestAccKeyPair_disappears(t *testing.T) {
	rName := acctest.RandomName("kp")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckKeyPairDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPairConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeyPairExists("acecloud_key_pair.test"),
					testAccDeleteKeyPairOutOfBand("acecloud_key_pair.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// --- Check functions ---

func testAccCheckKeyPairExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID is empty")
		}

		c := acctest.TestClient()
		path := fmt.Sprintf("/cloud/key-pairs/%s", rs.Primary.ID)
		_, err := c.Get(context.Background(), path, nil)
		if err != nil {
			return fmt.Errorf("key pair %s not found via API: %w", rs.Primary.ID, err)
		}
		return nil
	}
}

func testAccCheckKeyPairDestroy(s *terraform.State) error {
	c := acctest.TestClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "acecloud_key_pair" {
			continue
		}
		path := fmt.Sprintf("/cloud/key-pairs/%s", rs.Primary.ID)
		_, err := c.Get(context.Background(), path, nil)
		if err == nil {
			return fmt.Errorf("key pair %s still exists after destroy", rs.Primary.ID)
		}
		if !client.IsNotFound(err) {
			return fmt.Errorf("unexpected error checking key pair %s: %w", rs.Primary.ID, err)
		}
	}
	return nil
}

func testAccDeleteKeyPairOutOfBand(resourceName string) resource.TestCheckFunc {
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
		_, err := c.Delete(context.Background(), "/cloud/key-pairs", body)
		if err != nil {
			return fmt.Errorf("failed to delete key pair out-of-band: %w", err)
		}
		return nil
	}
}

// --- Config functions ---

func testAccKeyPairConfig_basic(name string) string {
	return acctest.ProviderConfig() + fmt.Sprintf(`
resource "acecloud_key_pair" "test" {
  name = %[1]q
}
`, name)
}
