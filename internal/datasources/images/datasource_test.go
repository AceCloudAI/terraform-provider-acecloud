package images_test

import (
	"regexp"
	"testing"

	"github.com/AceCloudAI/terraform-provider-acecloud/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceImages_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceImagesConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.acecloud_images.all", "images.#"),
					resource.TestMatchResourceAttr("data.acecloud_images.all", "images.#", regexp.MustCompile(`^[1-9]\d*$`)),
				),
			},
		},
	})
}

func TestAccDataSourceImages_attributes(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceImagesConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.acecloud_images.all", "images.0.id"),
					resource.TestCheckResourceAttrSet("data.acecloud_images.all", "images.0.name"),
					resource.TestCheckResourceAttrSet("data.acecloud_images.all", "images.0.status"),
				),
			},
		},
	})
}

func testAccDataSourceImagesConfig() string {
	return acctest.ProviderConfig() + `
data "acecloud_images" "all" {}
`
}
