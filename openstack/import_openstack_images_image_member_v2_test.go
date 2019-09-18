package openstack

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccImagesImageMemberV2_importBasic(t *testing.T) {
	memberName := "data.openstack_identity_auth_scope_v3.scope"
	imageName := "openstack_images_image_v2.image_1"
	resourceName := "openstack_images_image_member_v2.image_member_1"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckImagesImageMemberV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImagesImageMemberV2_basic,
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccImagesImageMemberV2ImportID(imageName, memberName),
			},
		},
	})
}

func testAccImagesImageMemberV2ImportID(imageName, memberName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		image, ok := s.RootModule().Resources[imageName]
		if !ok {
			return "", fmt.Errorf("Image not found: %s", imageName)
		}

		member, ok := s.RootModule().Resources[memberName]
		if !ok {
			return "", fmt.Errorf("Member not found: %s", memberName)
		}

		return fmt.Sprintf("%s/%s", image.Primary.ID, member.Primary.Attributes["project_id"]), nil
	}
}
