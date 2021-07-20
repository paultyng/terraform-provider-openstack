package openstack

import (
	"fmt"
	"log"
	"strings"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceComputeKeypairV2() *schema.Resource {
	return &schema.Resource{
		Create: resourceComputeKeypairV2Create,
		Read:   resourceComputeKeypairV2Read,
		Delete: resourceComputeKeypairV2Delete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"user_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"public_key": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"value_specs": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
			},

			// computed-only
			"private_key": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func extractComputeKeyPairNameAndUserID(fullID string) (id string, userID string) {
	id = fullID

	separatorIndex := strings.IndexRune(fullID, ':')
	if separatorIndex != -1 {
		userID = fullID[:separatorIndex]
		id = fullID[separatorIndex+1:]
	}

	return
}

func resourceComputeKeypairV2Create(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	computeClient, err := config.ComputeV2Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack compute client: %s", err)
	}
	computeClient.Microversion = computeV2KeyPairUserID

	name := d.Get("name").(string)
	createOpts := ComputeKeyPairV2CreateOpts{
		keypairs.CreateOpts{
			Name:      name,
			PublicKey: d.Get("public_key").(string),
		},
		MapValueSpecs(d),
	}

	// Check if the private key is for a specific user and in case update the creation properties
	userID, isForUser := d.GetOk("user_id")
	if isForUser {
		createOpts.CreateOpts.UserID = userID.(string)
	}

	log.Printf("[DEBUG] openstack_compute_keypair_v2 create options: %#v", createOpts)

	kp, err := keypairs.Create(computeClient, createOpts).Extract()
	if err != nil {
		return fmt.Errorf("Unable to create openstack_compute_keypair_v2 %s: %s", name, err)
	}

	id := kp.Name
	if isForUser {
		id = kp.UserID + ":" + id
	}
	d.SetId(id)

	// Private Key is only available in the response to a create.
	d.Set("private_key", kp.PrivateKey)

	return resourceComputeKeypairV2Read(d, meta)
}

func resourceComputeKeypairV2Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	computeClient, err := config.ComputeV2Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack compute client: %s", err)
	}
	computeClient.Microversion = computeV2KeyPairUserID

	// Check if the id includes a user_id
	id, userID := extractComputeKeyPairNameAndUserID(d.Id())
	opts := keypairs.GetOpts{
		UserID: userID,
	}

	kp, err := keypairs.Get(computeClient, id, opts).Extract()
	if err != nil {
		return CheckDeleted(d, err, "Error retrieving openstack_compute_keypair_v2")
	}

	log.Printf("[DEBUG] Retrieved openstack_compute_keypair_v2 %s: %#v", d.Id(), kp)

	d.Set("name", kp.Name)
	d.Set("public_key", kp.PublicKey)
	d.Set("fingerprint", kp.Fingerprint)
	d.Set("user_id", kp.UserID)
	d.Set("region", GetRegion(d, config))

	return nil
}

func resourceComputeKeypairV2Delete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	computeClient, err := config.ComputeV2Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack compute client: %s", err)
	}
	computeClient.Microversion = computeV2KeyPairUserID

	// Check if the id includes a user_id
	id, userID := extractComputeKeyPairNameAndUserID(d.Id())
	opts := keypairs.DeleteOpts{
		UserID: userID,
	}

	err = keypairs.Delete(computeClient, id, opts).ExtractErr()
	if err != nil {
		return CheckDeleted(d, err, "Error deleting openstack_compute_keypair_v2")
	}

	return nil
}
