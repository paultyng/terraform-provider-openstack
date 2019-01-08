package openstack

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"

	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
)

func resourceNetworkingPortSecGroupAssociateV2() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetworkingPortSecGroupAssociateV2Create,
		Read:   resourceNetworkingPortSecGroupAssociateV2Read,
		Update: resourceNetworkingPortSecGroupAssociateV2Update,
		Delete: resourceNetworkingPortSecGroupAssociateV2Delete,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"port_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"security_group_ids": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"enforce": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"all_security_group_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
		},
	}
}

func resourceNetworkingPortSecGroupAssociateV2Create(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	networkingClient, err := config.networkingV2Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack networking client: %s", err)
	}

	securityGroups := expandToStringSlice(d.Get("security_group_ids").(*schema.Set).List())
	portID := d.Get("port_id").(string)

	port, err := ports.Get(networkingClient, portID).Extract()
	if err != nil {
		return fmt.Errorf("Unable to get %s Port: %s", portID, err)
	}

	log.Printf("[DEBUG] Retrieved Port %s: %+v", portID, port)

	var updateOpts ports.UpdateOpts
	if v, ok := d.GetOk("enforce"); ok && v.(bool) == true {
		updateOpts = ports.UpdateOpts{SecurityGroups: &securityGroups}
	} else {
		// append security groups
		sg := sliceUnion(port.SecurityGroups, securityGroups)
		updateOpts = ports.UpdateOpts{SecurityGroups: &sg}
	}

	log.Printf("[DEBUG] Port Security Group Associate Options: %#v", updateOpts)

	_, err = ports.Update(networkingClient, portID, updateOpts).Extract()
	if err != nil {
		return fmt.Errorf("Error associating %s port with '%s' security groups: %s", portID, strings.Join(securityGroups, ","), err)
	}

	d.SetId(portID)

	log.Printf("[DEBUG] Storing old security group IDs into the 'old_security_group_ids' attribute: %#v", port.SecurityGroups)
	d.Set("old_security_group_ids", port.SecurityGroups)

	return resourceNetworkingPortSecGroupAssociateV2Read(d, meta)
}

func resourceNetworkingPortSecGroupAssociateV2Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	networkingClient, err := config.networkingV2Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack networking client: %s", err)
	}

	port, err := ports.Get(networkingClient, d.Id()).Extract()
	if err != nil {
		return CheckDeleted(d, err, "port")
	}

	if v, ok := d.GetOk("enforce"); ok && v.(bool) == true {
		d.Set("security_group_ids", port.SecurityGroups)
	}
	d.Set("all_security_group_ids", port.SecurityGroups)

	d.Set("region", GetRegion(d, config))

	return nil
}

func resourceNetworkingPortSecGroupAssociateV2Update(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	networkingClient, err := config.networkingV2Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack networking client: %s", err)
	}

	var updateOpts ports.UpdateOpts

	if d.Get("enforce").(bool) {
		securityGroups := expandToStringSlice(d.Get("security_group_ids").(*schema.Set).List())
		updateOpts = ports.UpdateOpts{SecurityGroups: &securityGroups}
	} else {
		allSet := d.Get("all_security_group_ids").(*schema.Set)
		oldIDs, newIDs := d.GetChange("security_group_ids")
		oldSet, newSet := oldIDs.(*schema.Set), newIDs.(*schema.Set)

		allWithoutOld := allSet.Difference(oldSet)

		newSecurityGroups := expandToStringSlice(allWithoutOld.Union(newSet).List())

		updateOpts = ports.UpdateOpts{SecurityGroups: &newSecurityGroups}
	}

	if d.HasChange("security_group_ids") || d.HasChange("enforce") {
		log.Printf("[DEBUG] Port Security Group Update Options: %#v", updateOpts)
		_, err = ports.Update(networkingClient, d.Id(), updateOpts).Extract()
		if err != nil {
			return fmt.Errorf("Error updating OpenStack Neutron Port: %s", err)
		}
	}

	return resourceNetworkingPortV2Read(d, meta)
}

func resourceNetworkingPortSecGroupAssociateV2Delete(d *schema.ResourceData, meta interface{}) error {
	if d.Get("enforce").(bool) == false {
		return nil
	}

	config := meta.(*Config)
	networkingClient, err := config.networkingV2Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack networking client: %s", err)
	}

	updateOpts := ports.UpdateOpts{SecurityGroups: &[]string{}}

	log.Printf("[DEBUG] Port security groups disassociation options: %#v", updateOpts)

	_, err = ports.Update(networkingClient, d.Id(), updateOpts).Extract()
	if err != nil {
		return CheckDeleted(d, err, "Error disassociating port security groups")
	}

	return nil
}
