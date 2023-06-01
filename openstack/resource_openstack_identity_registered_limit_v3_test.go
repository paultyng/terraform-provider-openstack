package openstack

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/gophercloud/gophercloud/openstack/identity/v3/registeredlimits"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/services"
)

func TestAccIdentityV3RegisteredLimit_basic(t *testing.T) {
	var service services.Service
	_ = os.Setenv("OS_SYSTEM_SCOPE", "true")
	defer os.Unsetenv("OS_SYSTEM_SCOPE")
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAdminOnly(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckIdentityV3RegisteredLimitDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIdentityV3RegisteredLimitBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdentityV3RegisteredLimitExists("openstack_identity_registered_limit_v3.limit_1", &service),
					resource.TestCheckResourceAttrPtr(
						"openstack_identity_registered_limit_v3.limit_1", "service_id", &service.ID),
					resource.TestCheckResourceAttr(
						"openstack_identity_registered_limit_v3.limit_1", "resource_name", "image_count_total"),
					resource.TestCheckResourceAttr(
						"openstack_identity_registered_limit_v3.limit_1", "default_limit", "10"),
					resource.TestCheckResourceAttr(
						"openstack_identity_registered_limit_v3.limit_1", "description", "foo"),
				),
			},
			{
				Config: testAccIdentityV3RegisteredLimitUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdentityV3RegisteredLimitExists("openstack_identity_registered_limit_v3.limit_1", &service),
					resource.TestCheckResourceAttrPtr(
						"openstack_identity_registered_limit_v3.limit_1", "service_id", &service.ID),
					resource.TestCheckResourceAttr(
						"openstack_identity_registered_limit_v3.limit_1", "resource_name", "image_count_total"),
					resource.TestCheckResourceAttr(
						"openstack_identity_registered_limit_v3.limit_1", "default_limit", "10"),
					resource.TestCheckResourceAttr(
						"openstack_identity_registered_limit_v3.limit_1", "description", "bar"),
				),
			},
		},
	})
}

func testAccCheckIdentityV3RegisteredLimitDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	identityClient, err := config.IdentityV3Client(osRegionName)
	if err != nil {
		return fmt.Errorf("Error creating OpenStack identity client: %s", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "openstack_identity_registered_limit_v3" {
			continue
		}

		_, err := registeredlimits.Get(identityClient, rs.Primary.ID).Extract()
		if err == nil {
			return fmt.Errorf("Registered limit still exists")
		}
	}

	return nil
}

func testAccCheckIdentityV3RegisteredLimitExists(n string, service *services.Service) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)
		identityClient, err := config.IdentityV3Client(osRegionName)
		if err != nil {
			return fmt.Errorf("Error creating OpenStack identity client: %s", err)
		}

		found, err := registeredlimits.Get(identityClient, rs.Primary.ID).Extract()
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("Registered limit not found")
		}

		service, err = services.Get(identityClient, found.ServiceID).Extract()
		if err != nil {
			return fmt.Errorf("Error retrieving OpenStack service %s: %s", found.ServiceID, err)
		}

		return nil
	}
}

const testAccIdentityV3RegisteredLimitBasic = `
data "openstack_identity_service_v3" "glance" {
	name = "glance"
}

resource "openstack_identity_registered_limit_v3" "limit_1" {
	service_id = data.openstack_identity_service_v3.glance.id
	resource_name = "image_count_total"
	default_limit = 10 
	description = "foo"
}
`

const testAccIdentityV3RegisteredLimitUpdate = `
data "openstack_identity_service_v3" "glance" {
	name = "glance"
}

resource "openstack_identity_registered_limit_v3" "limit_1" {
	service_id = data.openstack_identity_service_v3.glance.id
	resource_name = "image_count_total"
	default_limit = 100 
	description = "bar"
}
`
