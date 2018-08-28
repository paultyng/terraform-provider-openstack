package openstack

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"time"

	"github.com/gophercloud/gophercloud/openstack/objectstorage/v1/objects"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceObjectstorageTempurlV1() *schema.Resource {
	return &schema.Resource{
		Create: resourceObjectstorageTempurlV1Create,
		Read:   resourceObjectstorageTempurlV1Read,
		Delete: schema.RemoveFromState,

		Schema: map[string]*schema.Schema{
			"region": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"container": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"object": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"method": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "get",
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					if value != "get" && value != "post" {
						errors = append(errors, fmt.Errorf(
							"Only 'get', and 'post' are supported values for 'method'"))
					}
					return
				},
			},

			"ttl": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},

			"split": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

// resourceObjectstorageTempurlV1Read performs the image lookup.
func resourceObjectstorageTempurlV1Read(d *schema.ResourceData, meta interface{}) error {
	turl := d.Get("url").(string)
	u, err := url.Parse(turl)
	if err != nil {
		return fmt.Errorf("Failed to read the temp url: %s", turl)
	}

	qp, _ := url.ParseQuery(u.RawQuery)
	expiry, err := strconv.ParseInt(qp.Get("temp_url_expires"), 10, 64)
	if err != nil {
		return fmt.Errorf("Failed to parse the temp url expiration time: %s", qp.Get("temp_url_expires"))
	}
	now := time.Now().Unix()
	if expiry < now {
		log.Printf("[DEBUG] URL expired, generating a new one")
		d.SetId("")
	}

	return nil
}

// resourceObjectstorageTempurlV1Create performs the image lookup.
func resourceObjectstorageTempurlV1Create(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	objectStorageClient, err := config.objectStorageV1Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack compute client: %s", err)
	}

	method := objects.GET
	switch d.Get("method") {
	case "post":
		method = objects.POST
		// gophercloud doesn't have support for PUT yet,
		// although it's a valid method for swift
		//case "put":
		//	method = objects.PUT
	}

	turlOptions := objects.CreateTempURLOpts{
		Method: method,
		TTL:    d.Get("ttl").(int),
		Split:  d.Get("split").(string),
	}

	containerName := d.Get("container").(string)
	objectName := d.Get("object").(string)

	log.Printf("[DEBUG] Create TempURL Options: %#v", turlOptions)

	url, err := objects.CreateTempURL(objectStorageClient, containerName, objectName, turlOptions)
	if err != nil {
		return fmt.Errorf("Unable to generate a TempURL for the object %s in container %s: %s",
			objectName, containerName, err)
	}

	log.Printf("[DEBUG] URL Generated: %s", url)

	// Set the URL and Id fields.
	hasher := md5.New()
	hasher.Write([]byte(url))
	d.SetId(hex.EncodeToString(hasher.Sum(nil)))
	d.Set("url", url)
	return nil
}
