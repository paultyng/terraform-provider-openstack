package openstack

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccOpenStackNetworkingRouterV2DataSource_name(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccOpenStackNetworkingRouterV2DataSource_router,
			},
			resource.TestStep{
				Config: testAccOpenStackNetworkingRouterV2DataSource_name,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingRouterV2DataSourceID("data.openstack_networking_router_v2.router"),
					resource.TestCheckResourceAttrSet(
						"data.openstack_networking_router_v2.router", "name"),
					resource.TestCheckResourceAttrSet(
						"data.openstack_networking_router_v2.router", "admin_state_up"),
					resource.TestCheckResourceAttrSet(
						"data.openstack_networking_router_v2.router", "status"),
				),
			},
		},
	})
}

func testAccCheckNetworkingRouterV2DataSourceID(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find router data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Router data source ID not set")
		}

		return nil
	}
}

const testAccOpenStackNetworkingRouterV2DataSource_router = `
resource "openstack_networking_router_v2" "router" {
  name           = "router_tf"
  admin_state_up = "true"
}
`

var testAccOpenStackNetworkingRouterV2DataSource_name = fmt.Sprintf(`
%s

data "openstack_networking_router_v2" "router" {
   name           = "${openstack_networking_router_v2.router.name}"
   admin_state_up = "true"
}
`, testAccOpenStackNetworkingRouterV2DataSource_router)
