package openstack

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/gophercloud/gophercloud/openstack/identity/v3/limits"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/projects"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/services"
)

func TestAccIdentityV3Limit_basic(t *testing.T) {
	var project projects.Project
	var service services.Service
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAdminOnly(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckIdentityV3LimitDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIdentityV3LimitBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdentityV3LimitExists("openstack_identity_limit_v3.limit_1", &project, &service),
					resource.TestCheckResourceAttrPtr(
						"openstack_identity_limit_v3.limit_1", "service_id", &service.ID),
					resource.TestCheckResourceAttrPtr(
						"openstack_identity_limit_v3.limit_1", "project_id", &project.ID),
					resource.TestCheckResourceAttr(
						"openstack_identity_limit_v3.limit_1", "resource_name", "image_count_total"),
					resource.TestCheckResourceAttr(
						"openstack_identity_limit_v3.limit_1", "resource_limit", "10"),
					resource.TestCheckResourceAttr(
						"openstack_identity_limit_v3.limit_1", "description", "foo"),
				),
			},
			{
				Config: testAccIdentityV3LimitUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdentityV3LimitExists("openstack_identity_limit_v3.limit_1", &project, &service),
					resource.TestCheckResourceAttrPtr(
						"openstack_identity_limit_v3.limit_1", "service_id", &service.ID),
					resource.TestCheckResourceAttrPtr(
						"openstack_identity_limit_v3.limit_1", "project_id", &project.ID),
					resource.TestCheckResourceAttr(
						"openstack_identity_limit_v3.limit_1", "resource_name", "image_count_total"),
					resource.TestCheckResourceAttr(
						"openstack_identity_limit_v3.limit_1", "resource_limit", "10"),
					resource.TestCheckResourceAttr(
						"openstack_identity_limit_v3.limit_1", "description", "bar"),
				),
			},
		},
	})
}

func testAccCheckIdentityV3LimitDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	identityClient, err := config.IdentityV3Client(osRegionName)
	if err != nil {
		return fmt.Errorf("Error creating OpenStack identity client: %s", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "openstack_identity_limit_v3" {
			continue
		}

		_, err := limits.Get(identityClient, rs.Primary.ID).Extract()
		if err == nil {
			return fmt.Errorf("Limit still exists")
		}
	}

	return nil
}

func testAccCheckIdentityV3LimitExists(n string, project *projects.Project, service *services.Service) resource.TestCheckFunc {
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

		found, err := limits.Get(identityClient, rs.Primary.ID).Extract()
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("Limit not found")
		}

		project, err = projects.Get(identityClient, found.ProjectID).Extract()
		if err != nil {
			return fmt.Errorf("Error retrieving OpenStack project %s: %s", found.ProjectID, err)
		}

		service, err = services.Get(identityClient, found.ServiceID).Extract()
		if err != nil {
			return fmt.Errorf("Error retrieving OpenStack service %s: %s", found.ServiceID, err)
		}

		return nil
	}
}

const testAccIdentityV3LimitBasic = `
data "openstack_identity_service_v3" "glance" {
	name = "glance"
}

resource "openstack_identity_project_v3" "project_1" {
	name = "project_1"
}

resource "openstack_identity_limit_v3" "limit_1" {
	project_id = openstack_identity_project_v3.project_1.id
	service_id = data.openstack_identity_service_v3.glance.id
	resource_name = "image_count_total"
	resource_limit = 10 
	description = "foo"
}
`

const testAccIdentityV3LimitUpdate = `
data "openstack_identity_service_v3" "glance" {
	name = "glance"
}

resource "openstack_identity_project_v3" "project_1" {
	name = "project_1"
}

resource "openstack_identity_limit_v3" "limit_1" {
	project_id = openstack_identity_project_v3.project_1.id
	service_id = data.openstack_identity_service_v3.glance.id
	resource_name = "image_count_total"
	resource_limit = 100 
	description = "bar"
}
`
