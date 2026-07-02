package vpcs_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/AceCloudAI/terraform-provider-acecloud/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceVPCs_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVPCsConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.acecloud_vpcs.all", "vpcs.#"),
				),
			},
		},
	})
}

func TestAccDataSourceVPCs_withResource(t *testing.T) {
	rName := acctest.RandomName("dsvpc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVPCsConfig_withVPC(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("data.acecloud_vpcs.all", "vpcs.#", regexp.MustCompile(`^[1-9]\d*$`)),
				),
			},
		},
	})
}

func testAccDataSourceVPCsConfig() string {
	return acctest.ProviderConfig() + `
data "acecloud_vpcs" "all" {}
`
}

func testAccDataSourceVPCsConfig_withVPC(name string) string {
	return acctest.ProviderConfig() + fmt.Sprintf(`
resource "acecloud_vpc" "test" {
  name              = %[1]q
  subnet_name       = "%[1]s-subnet"
  subnet_cidr       = "10.0.0.0/24"
  subnet_ip_version = 4
}

data "acecloud_vpcs" "all" {
  depends_on = [acecloud_vpc.test]
}
`, name)
}
