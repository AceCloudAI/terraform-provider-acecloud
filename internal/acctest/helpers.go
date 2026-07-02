package acctest

import (
	"fmt"
	"os"

	"github.com/AceCloudAI/terraform-provider-acecloud/internal/client"
)

// TestClient returns a configured *client.Client for direct API calls in
// acceptance test check functions (e.g. verifying a resource still exists
// after create, or confirming deletion). Panics if required env vars are
// missing — callers must have already passed PreCheck.
func TestClient() *client.Client {
	return client.NewClientWithAPIKey(
		os.Getenv("ACECLOUD_API_URL"),
		os.Getenv("ACECLOUD_API_KEY_ID"),
		os.Getenv("ACECLOUD_API_KEY_SECRET"),
		os.Getenv("ACECLOUD_API_KEY_SERVICE_NAME"),
		os.Getenv("ACECLOUD_REGION"),
		os.Getenv("ACECLOUD_PROJECT_ID"),
	)
}

// ConfigCompose concatenates multiple HCL config fragments into a single
// string. Useful for tests that need base infrastructure (VPC, subnet, etc.)
// alongside the resource under test.
func ConfigCompose(configs ...string) string {
	var result string
	for _, c := range configs {
		result += c + "\n"
	}
	return result
}

// TestAccBaseVPCConfig returns HCL for a VPC with an inline subnet, suitable
// as a dependency for tests that need networking infrastructure.
func TestAccBaseVPCConfig(name string) string {
	return ProviderConfig() + fmt.Sprintf(`
resource "acecloud_vpc" "test-base" {
  name              = %[1]q
  subnet_name       = "%[1]s-subnet"
  subnet_cidr       = "10.0.0.0/24"
  subnet_ip_version = 4
}
`, name)
}

// TestAccBaseSecurityGroupConfig returns HCL for a minimal security group
// with an SSH ingress rule. Use as a dependency for instance tests.
func TestAccBaseSecurityGroupConfig(name string) string {
	return fmt.Sprintf(`
resource "acecloud_security_group" "test-base" {
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
`, name)
}

// TestAccBaseRouterConfig returns HCL for a router, suitable as a
// dependency for router interface tests.
func TestAccBaseRouterConfig(name string) string {
	return fmt.Sprintf(`
resource "acecloud_router" "test-base" {
  name = %[1]q
}
`, name)
}

// TestAccBaseRouterWithGatewayConfig returns HCL for a router with an
// external gateway set. Used by tests that need egress/SNAT.
func TestAccBaseRouterWithGatewayConfig(name string) string {
	return fmt.Sprintf(`
resource "acecloud_router" "test-base" {
  name                        = %[1]q
  external_gateway_network_id = %[2]q
}
`, name, os.Getenv("ACECLOUD_EXTERNAL_NETWORK_ID"))
}
