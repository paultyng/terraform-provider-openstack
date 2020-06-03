package openstack

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccOpenStackImagesV2ImageIDsDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccOpenStackImagesV2ImageIDsDataSource_empty,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.openstack_images_image_ids_v2.images_empty", "ids.#", "0"),
				),
			},
			{
				Config: testAccOpenStackImagesV2ImageIDsDataSource_name,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.openstack_images_image_ids_v2.images_by_name", "ids.#", "1"),
					resource.TestCheckResourceAttrPair(
						"data.openstack_images_image_ids_v2.images_by_name", "ids.0",
						"openstack_images_image_v2.image_1", "id"),
				),
			},
			{
				Config: testAccOpenStackImagesV2ImageIDsDataSource_regex,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.openstack_images_image_ids_v2.images_by_name_regex", "ids.#", "1"),
					resource.TestCheckResourceAttrPair(
						"data.openstack_images_image_ids_v2.images_by_name_regex", "ids.0",
						"openstack_images_image_v2.image_2", "id"),
				),
			},
			{
				Config: testAccOpenStackImagesV2ImageIDsDataSource_tag,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.openstack_images_image_ids_v2.images_by_tag", "ids.#", "1"),
					resource.TestCheckResourceAttrPair(
						"data.openstack_images_image_ids_v2.images_by_tag", "ids.0",
						"openstack_images_image_v2.image_1", "id"),
				),
			},
			/*{
				Config: testAccOpenStackImagesV2ImageIDsDataSource_min,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.openstack_images_image_ids_v2.images_by_size_min", "ids.#", "1"),
					resource.TestCheckResourceAttrPair(
						"data.openstack_images_image_ids_v2.images_by_size_min", "ids.0",
						"openstack_images_image_v2.image_2", "id"),
				),
			},
			{
				Config: testAccOpenStackImagesV2ImageIDsDataSource_max,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.openstack_images_image_ids_v2.images_by_size_max", "ids.#", "1"),
					resource.TestCheckResourceAttrPair(
						"data.openstack_images_image_ids_v2.images_by_size_max", "ids.0",
						"openstack_images_image_v2.image_1", "id"),
				),
			},*/
			{
				Config: testAccOpenStackImagesV2ImageIDsDataSource_properties,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.openstack_images_image_ids_v2.images_by_properties", "ids.#", "1"),
					resource.TestCheckResourceAttrPair(
						"data.openstack_images_image_ids_v2.images_by_properties", "ids.0",
						"openstack_images_image_v2.image_1", "id"),
				),
			},
		},
	})
}

// Standard CirrOS image
const testAccOpenStackImagesV2ImageIDsDataSource_cirros = `
resource "openstack_images_image_v2" "image_1" {
  name = "CirrOS-tf_1"
  container_format = "bare"
  disk_format = "qcow2"
  image_source_url = "http://download.cirros-cloud.net/0.3.5/cirros-0.3.5-x86_64-disk.img"
  tags = ["cirros-tf_1"]
  properties = {
    foo = "bar"
    bar = "foo"
  }
  visibility = "private"
}

resource "openstack_images_image_v2" "image_2" {
  name = "CirrOS-tf_2"
  container_format = "bare"
  disk_format = "qcow2"
  image_source_url = "http://download.cirros-cloud.net/0.5.1/cirros-0.5.1-x86_64-disk.img"
  tags = ["cirros-tf_2"]
  properties = {
    foo = "bar"
  }
  visibility = "private"
}
`

var testAccOpenStackImagesV2ImageIDsDataSource_empty = fmt.Sprintf(`
%s

data "openstack_images_image_ids_v2" "images_empty" {
        name = "non-existed-image"
	visibility = "private"
}
`, testAccOpenStackImagesV2ImageIDsDataSource_cirros)

var testAccOpenStackImagesV2ImageIDsDataSource_name = fmt.Sprintf(`
%s

data "openstack_images_image_ids_v2" "images_by_name" {
	name = "${openstack_images_image_v2.image_1.name}"
	visibility = "private"
}
`, testAccOpenStackImagesV2ImageIDsDataSource_cirros)

var testAccOpenStackImagesV2ImageIDsDataSource_regex = fmt.Sprintf(`
%s

data "openstack_images_image_ids_v2" "images_by_name_regex" {
	name_regex = "^.+tf_2$"
	visibility = "private"
}
`, testAccOpenStackImagesV2ImageIDsDataSource_cirros)

var testAccOpenStackImagesV2ImageIDsDataSource_tag = fmt.Sprintf(`
%s

data "openstack_images_image_ids_v2" "images_by_tag" {
	tag = "cirros-tf_1"
	visibility = "private"
}
`, testAccOpenStackImagesV2ImageIDsDataSource_cirros)

var testAccOpenStackImagesV2ImageIDsDataSource_min = fmt.Sprintf(`
%s

data "openstack_images_image_ids_v2" "images_by_size_min" {
	size_min = "15000000"
	visibility = "private"
}
`, testAccOpenStackImagesV2ImageIDsDataSource_cirros)

var testAccOpenStackImagesV2ImageIDsDataSource_max = fmt.Sprintf(`
%s

data "openstack_images_image_ids_v2" "images_by_size_max" {
	size_max = "15000000"
	visibility = "private"
}
`, testAccOpenStackImagesV2ImageIDsDataSource_cirros)

var testAccOpenStackImagesV2ImageIDsDataSource_properties = fmt.Sprintf(`
%s

data "openstack_images_image_ids_v2" "images_by_properties" {
	properties = {
		foo = "bar"
		bar = "foo"
	}
	visibility = "private"
}
`, testAccOpenStackImagesV2ImageIDsDataSource_cirros)
