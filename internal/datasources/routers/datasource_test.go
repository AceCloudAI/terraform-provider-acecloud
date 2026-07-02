package routers_test

import (
	"fmt"
	"testing"

	"github.com/AceCloudAI/terraform-provider-acecloud/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceRouters_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceRoutersConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.acecloud_routers.all", "routers.#"),
				),
			},
		},
	})
}

func TestAccDataSourceRouters_withResource(t *testing.T) {
	rName := acctest.RandomName("dsrtr")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceRoutersConfig_withRouter(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.acecloud_routers.all", "routers.#"),
				),
			},
		},
	})
}

func testAccDataSourceRoutersConfig() string {
	return acctest.ProviderConfig() + `
data "acecloud_routers" "all" {}
`
}

func testAccDataSourceRoutersConfig_withRouter(name string) string {
	return acctest.ProviderConfig() + fmt.Sprintf(`
resource "acecloud_router" "test" {
  name = %[1]q
}

data "acecloud_routers" "all" {
  depends_on = [acecloud_router.test]
}
`, name)
}
