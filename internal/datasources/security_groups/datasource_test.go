package security_groups_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/AceCloudAI/terraform-provider-acecloud/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceSecurityGroups_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSecurityGroupsConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.acecloud_security_groups.all", "security_groups.#"),
					resource.TestMatchResourceAttr("data.acecloud_security_groups.all", "security_groups.#", regexp.MustCompile(`^[1-9]\d*$`)),
				),
			},
		},
	})
}

func TestAccDataSourceSecurityGroups_withResource(t *testing.T) {
	rName := acctest.RandomName("dssg")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSecurityGroupsConfig_withSG(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("data.acecloud_security_groups.all", "security_groups.#", regexp.MustCompile(`^[1-9]\d*$`)),
				),
			},
		},
	})
}

func testAccDataSourceSecurityGroupsConfig() string {
	return acctest.ProviderConfig() + `
data "acecloud_security_groups" "all" {}
`
}

func testAccDataSourceSecurityGroupsConfig_withSG(name string) string {
	return acctest.ProviderConfig() + fmt.Sprintf(`
resource "acecloud_security_group" "test" {
  name = %[1]q

  rules {
    direction        = "ingress"
    protocol         = "tcp"
    port_range_min   = 22
    port_range_max   = 22
    remote_ip_prefix = "0.0.0.0/0"
    ethertype        = "IPv4"
  }
}

data "acecloud_security_groups" "all" {
  depends_on = [acecloud_security_group.test]
}
`, name)
}
