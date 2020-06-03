package openstack

import (
	"fmt"
	"log"
	"strings"

	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"

	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceImagesImageIDsV2() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceImagesImageIdsV2Read,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"visibility": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(images.ImageVisibilityPublic),
					string(images.ImageVisibilityPrivate),
					string(images.ImageVisibilityShared),
					string(images.ImageVisibilityCommunity),
				}, false),
			},

			"member_status": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(images.ImageMemberStatusAccepted),
					string(images.ImageMemberStatusPending),
					string(images.ImageMemberStatusRejected),
					string(images.ImageMemberStatusAll),
				}, false),
			},

			"owner": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"size_min": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},

			"size_max": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},

			"sort": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"sort_key": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "name",
			},

			"sort_direction": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "asc",
				ValidateFunc: validation.StringInSlice([]string{
					"asc", "desc",
				}, false),
			},

			"tag": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"properties": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
			},

			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsValidRegExp,
			},

			// Computed values
			"ids": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}

}

// dataSourceImagesImageIdsV2Read performs the image lookup.
func dataSourceImagesImageIdsV2Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	imageClient, err := config.ImageV2Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack image client: %s", err)
	}

	_, nameOk := d.GetOk("name")
	_, nameRegexOk := d.GetOk("name_regex")

	if nameOk && nameRegexOk {
		return fmt.Errorf("Attributes name and name_regexp can not be used at the same time")
	}

	_, sortOk := d.GetOk("sort")
	sortKeyValue := d.Get("sort_key")
	sortDirectionValue := d.Get("sort_direction")

	if sortOk {
		// Attribute "sort" cannot be used simultaneously with
		// "sort_key". If both are present, only "sort" will be used.
		sortKeyValue = ""
		sortDirectionValue = ""
	}

	visibility := resourceImagesImageV2VisibilityFromString(
		d.Get("visibility").(string))
	member_status := resourceImagesImageV2MemberStatusFromString(
		d.Get("member_status").(string))
	properties := resourceImagesImageV2ExpandProperties(
		d.Get("properties").(map[string]interface{}))

	var tags []string
	if tag := d.Get("tag").(string); tag != "" {
		tags = append(tags, tag)
	}

	listOpts := images.ListOpts{
		Name:         d.Get("name").(string),
		Visibility:   visibility,
		Owner:        d.Get("owner").(string),
		Status:       images.ImageStatusActive,
		SizeMin:      int64(d.Get("size_min").(int)),
		SizeMax:      int64(d.Get("size_max").(int)),
		Sort:         d.Get("sort").(string),
		SortKey:      sortKeyValue.(string),
		SortDir:      sortDirectionValue.(string),
		Tags:         tags,
		MemberStatus: member_status,
	}

	log.Printf("[DEBUG] List Options in openstack_images_image_ids_v2: %#v", listOpts)

	allPages, err := images.List(imageClient, listOpts).AllPages()
	if err != nil {
		return fmt.Errorf("Unable to list images in openstack_images_image_ids_v2: %s", err)
	}

	allImages, err := images.ExtractImages(allPages)
	if err != nil {
		return fmt.Errorf("Unable to retrieve images in openstack_images_image_ids_v2: %s", err)
	}

	log.Printf("[DEBUG] Retrieved %d images in openstack_images_image_ids_v2: %+v", len(allImages), allImages)

	allImages = imagesFilterByProperties(allImages, properties)

	log.Printf("[DEBUG] Image list filtered by properties: %#v", properties)

	if nameRegexOk {
		allImages = imagesFilterByRegex(allImages, d.Get("name_regex").(string))
		log.Printf("[DEBUG] Image list filtered by regex: %s", d.Get("name_regex"))
	}

	log.Printf("[DEBUG] Got %d images after filtering in openstack_images_image_ids_v2: %+v", len(allImages), allImages)

	imageIDs := make([]string, 0)
	for _, image := range allImages {

		imageIDs = append(imageIDs, image.ID)
	}

	d.SetId(fmt.Sprintf("%d", hashcode.String(strings.Join(imageIDs, ","))))
	d.Set("ids", imageIDs)

	return nil
}
