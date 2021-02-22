package openstack

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/quotas"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceNetworkingQuotaV2() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetworkingQuotaV2Create,
		Read:   resourceNetworkingQuotaV2Read,
		Update: resourceNetworkingQuotaV2Update,
		Delete: schema.RemoveFromState,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"project_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"floatingip": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			"network": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			"port": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			"rbac_policy": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			"router": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			"security_group": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			"security_group_rule": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			"subnet": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			"subnetpool": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceNetworkingQuotaV2Create(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	region := GetRegion(d, config)
	networkingClient, err := config.NetworkingV2Client(region)
	if err != nil {
		return fmt.Errorf("Error creating OpenStack networking client: %s", err)
	}

	projectID := d.Get("project_id").(string)
	floatingIP := d.Get("floatingip").(int)
	network := d.Get("network").(int)
	port := d.Get("port").(int)
	rbacPolicy := d.Get("rbac_policy").(int)
	router := d.Get("router").(int)
	securityGroup := d.Get("security_group").(int)
	securityGroupRule := d.Get("security_group_rule").(int)
	subnet := d.Get("subnet").(int)
	subnetPool := d.Get("subnetpool").(int)

	updateOpts := quotas.UpdateOpts{
		FloatingIP:        &floatingIP,
		Network:           &network,
		Port:              &port,
		RBACPolicy:        &rbacPolicy,
		Router:            &router,
		SecurityGroup:     &securityGroup,
		SecurityGroupRule: &securityGroupRule,
		Subnet:            &subnet,
		SubnetPool:        &subnetPool,
	}

	q, err := quotas.Update(networkingClient, projectID, updateOpts).Extract()
	if err != nil {
		return fmt.Errorf("Error creating openstack_networking_quota_v2: %s", err)
	}

	id := fmt.Sprintf("%s/%s", projectID, region)
	d.SetId(id)

	log.Printf("[DEBUG] Created openstack_networking_quota_v2 %#v", q)

	return resourceNetworkingQuotaV2Read(d, meta)
}

func resourceNetworkingQuotaV2Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	region := GetRegion(d, config)
	networkingClient, err := config.NetworkingV2Client(region)
	if err != nil {
		return fmt.Errorf("Error creating OpenStack networking client: %s", err)
	}

	// update resource ID with region to allow multi-region quota management
	if !strings.Contains(d.Id(), region) {
		id := fmt.Sprintf("%s/%s", d.Id(), region)
		d.SetId(id)
		log.Printf("[DEBUG] Updated ID of openstack_networking_quota_v2 to: %s", d.Id())
	}

	projectID, region, err := parseNetworkingQuotaID(d.Id())
	if err != nil {
		return CheckDeleted(d, err, "Error parsing ID of openstack_networking_quota_v2")
	}

	q, err := quotas.Get(networkingClient, projectID).Extract()
	if err != nil {
		return CheckDeleted(d, err, "Error retrieving openstack_networking_quota_v2")
	}

	log.Printf("[DEBUG] Retrieved openstack_networking_quota_v2 %s: %#v", d.Id(), q)

	d.Set("project_id", projectID)
	d.Set("region", region)
	d.Set("floatingip", q.FloatingIP)
	d.Set("network", q.Network)
	d.Set("port", q.Port)
	d.Set("rbac_policy", q.RBACPolicy)
	d.Set("router", q.Router)
	d.Set("security_group", q.SecurityGroup)
	d.Set("security_group_rule", q.SecurityGroupRule)
	d.Set("subnet", q.Subnet)
	d.Set("subnetpool", q.SubnetPool)

	return nil
}

func resourceNetworkingQuotaV2Update(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	networkingClient, err := config.NetworkingV2Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack networking client: %s", err)
	}

	var (
		hasChange  bool
		updateOpts quotas.UpdateOpts
	)

	if d.HasChange("floatingip") {
		hasChange = true
		floatingIP := d.Get("floatingip").(int)
		updateOpts.FloatingIP = &floatingIP
	}

	if d.HasChange("network") {
		hasChange = true
		network := d.Get("network").(int)
		updateOpts.Network = &network
	}

	if d.HasChange("port") {
		hasChange = true
		port := d.Get("port").(int)
		updateOpts.Port = &port
	}

	if d.HasChange("rbac_policy") {
		hasChange = true
		rbacPolicy := d.Get("rbac_policy").(int)
		updateOpts.RBACPolicy = &rbacPolicy
	}

	if d.HasChange("router") {
		hasChange = true
		router := d.Get("router").(int)
		updateOpts.Router = &router
	}

	if d.HasChange("security_group") {
		hasChange = true
		securityGroup := d.Get("security_group").(int)
		updateOpts.SecurityGroup = &securityGroup
	}

	if d.HasChange("security_group_rule") {
		hasChange = true
		securityGroupRule := d.Get("security_group_rule").(int)
		updateOpts.SecurityGroupRule = &securityGroupRule
	}

	if d.HasChange("subnet") {
		hasChange = true
		subnet := d.Get("subnet").(int)
		updateOpts.Subnet = &subnet
	}

	if d.HasChange("subnetpool") {
		hasChange = true
		subnetPool := d.Get("subnetpool").(int)
		updateOpts.SubnetPool = &subnetPool
	}

	if hasChange {
		log.Printf("[DEBUG] openstack_networking_quota_v2 %s update options: %#v", d.Id(), updateOpts)
		projectID, _, err := parseNetworkingQuotaID(d.Id())
		if err != nil {
			return CheckDeleted(d, err, "Error parsing ID of openstack_networking_quota_v2")
		}
		_, err = quotas.Update(networkingClient, projectID, updateOpts).Extract()
		if err != nil {
			return fmt.Errorf("Error updating openstack_networking_quota_v2: %s", err)
		}
	}

	return resourceNetworkingQuotaV2Read(d, meta)
}
