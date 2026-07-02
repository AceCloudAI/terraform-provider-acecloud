package flavors_test

import (
	"regexp"
	"testing"

	"github.com/AceCloudAI/terraform-provider-acecloud/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceFlavors_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceFlavorsConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.acecloud_flavors.all", "flavors.#"),
					resource.TestMatchResourceAttr("data.acecloud_flavors.all", "flavors.#", regexp.MustCompile(`^[1-9]\d*$`)),
				),
			},
		},
	})
}

func TestAccDataSourceFlavors_attributes(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceFlavorsConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.acecloud_flavors.all", "flavors.0.id"),
					resource.TestCheckResourceAttrSet("data.acecloud_flavors.all", "flavors.0.name"),
					resource.TestCheckResourceAttrSet("data.acecloud_flavors.all", "flavors.0.vcpus"),
					resource.TestCheckResourceAttrSet("data.acecloud_flavors.all", "flavors.0.ram"),
				),
			},
		},
	})
}

func testAccDataSourceFlavorsConfig() string {
	return acctest.ProviderConfig() + `
data "acecloud_flavors" "all" {}
`
}
