package openstack

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/lbaas_v2/listeners"
)

func resourceListenerV2() *schema.Resource {
	return &schema.Resource{
		Create: resourceListenerV2Create,
		Read:   resourceListenerV2Read,
		Update: resourceListenerV2Update,
		Delete: resourceListenerV2Delete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"region": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"protocol": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					if value != "TCP" && value != "HTTP" && value != "HTTPS" {
						errors = append(errors, fmt.Errorf(
							"Only 'TCP', 'HTTP', and 'HTTPS' are supported values for 'protocol'"))
					}
					return
				},
			},

			"protocol_port": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},

			"tenant_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"loadbalancer_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"default_pool_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"connection_limit": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},

			"default_tls_container_ref": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"sni_container_refs": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"admin_state_up": &schema.Schema{
				Type:     schema.TypeBool,
				Default:  true,
				Optional: true,
			},

			"id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceListenerV2Create(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	networkingClient, err := config.networkingV2Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack networking client: %s", err)
	}

	adminStateUp := d.Get("admin_state_up").(bool)
	connLimit := d.Get("connection_limit").(int)
	var sniContainerRefs []string
	if raw, ok := d.GetOk("sni_container_refs"); ok {
		for _, v := range raw.([]interface{}) {
			sniContainerRefs = append(sniContainerRefs, v.(string))
		}
	}
	createOpts := listeners.CreateOpts{
		Protocol:               listeners.Protocol(d.Get("protocol").(string)),
		ProtocolPort:           d.Get("protocol_port").(int),
		TenantID:               d.Get("tenant_id").(string),
		LoadbalancerID:         d.Get("loadbalancer_id").(string),
		Name:                   d.Get("name").(string),
		DefaultPoolID:          d.Get("default_pool_id").(string),
		Description:            d.Get("description").(string),
		ConnLimit:              &connLimit,
		DefaultTlsContainerRef: d.Get("default_tls_container_ref").(string),
		SniContainerRefs:       sniContainerRefs,
		AdminStateUp:           &adminStateUp,
	}

	log.Printf("[DEBUG] Create Options: %#v", createOpts)

	// Wait for LoadBalancer to become active before continuing
	lbID := createOpts.LoadbalancerID
	timeout := d.Timeout(schema.TimeoutCreate)
	err = waitForLBV2LoadBalancer(networkingClient, lbID, "ACTIVE", nil, timeout)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Attempting to create listener")
	var listener *listeners.Listener
	err = resource.Retry(timeout, func() *resource.RetryError {
		listener, err = listeners.Create(networkingClient, createOpts).Extract()
		if err != nil {
			return checkForRetryableError(err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Error creating listener: %s", err)
	}

	// Wait for Listener to become active
	err = waitForLBV2Listener(networkingClient, listener.ID, "ACTIVE", nil, timeout)
	if err != nil {
		return err
	}

	// Wait for LoadBalancer to become active again before continuing
	err = waitForLBV2LoadBalancer(networkingClient, lbID, "ACTIVE", nil, timeout)
	if err != nil {
		return err
	}

	d.SetId(listener.ID)

	return resourceListenerV2Read(d, meta)
}

func resourceListenerV2Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	networkingClient, err := config.networkingV2Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack networking client: %s", err)
	}

	listener, err := listeners.Get(networkingClient, d.Id()).Extract()
	if err != nil {
		return CheckDeleted(d, err, "listener")
	}

	log.Printf("[DEBUG] Retrieved listener %s: %#v", d.Id(), listener)

	d.Set("id", listener.ID)
	d.Set("name", listener.Name)
	d.Set("protocol", listener.Protocol)
	d.Set("tenant_id", listener.TenantID)
	d.Set("description", listener.Description)
	d.Set("protocol_port", listener.ProtocolPort)
	d.Set("admin_state_up", listener.AdminStateUp)
	d.Set("default_pool_id", listener.DefaultPoolID)
	d.Set("connection_limit", listener.ConnLimit)
	d.Set("sni_container_refs", listener.SniContainerRefs)
	d.Set("default_tls_container_ref", listener.DefaultTlsContainerRef)
	d.Set("region", GetRegion(d, config))

	return nil
}

func resourceListenerV2Update(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	networkingClient, err := config.networkingV2Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack networking client: %s", err)
	}

	var updateOpts listeners.UpdateOpts
	if d.HasChange("name") {
		updateOpts.Name = d.Get("name").(string)
	}
	if d.HasChange("description") {
		updateOpts.Description = d.Get("description").(string)
	}
	if d.HasChange("connection_limit") {
		connLimit := d.Get("connection_limit").(int)
		updateOpts.ConnLimit = &connLimit
	}
	if d.HasChange("default_tls_container_ref") {
		updateOpts.DefaultTlsContainerRef = d.Get("default_tls_container_ref").(string)
	}
	if d.HasChange("sni_container_refs") {
		var sniContainerRefs []string
		if raw, ok := d.GetOk("sni_container_refs"); ok {
			for _, v := range raw.([]interface{}) {
				sniContainerRefs = append(sniContainerRefs, v.(string))
			}
		}
		updateOpts.SniContainerRefs = sniContainerRefs
	}
	if d.HasChange("admin_state_up") {
		asu := d.Get("admin_state_up").(bool)
		updateOpts.AdminStateUp = &asu
	}

	// Wait for LoadBalancer to become active before continuing
	lbID := d.Get("loadbalancer_id").(string)
	timeout := d.Timeout(schema.TimeoutUpdate)
	err = waitForLBV2LoadBalancer(networkingClient, lbID, "ACTIVE", nil, timeout)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Updating listener %s with options: %#v", d.Id(), updateOpts)
	err = resource.Retry(timeout, func() *resource.RetryError {
		_, err = listeners.Update(networkingClient, d.Id(), updateOpts).Extract()
		if err != nil {
			return checkForRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("Error updating listener %s: %s", d.Id(), err)
	}

	// Wait for Listener to become active before continuing
	err = waitForLBV2Listener(networkingClient, d.Id(), "ACTIVE", nil, timeout)
	if err != nil {
		return err
	}

	// Wait for LoadBalancer to become active again before continuing
	err = waitForLBV2LoadBalancer(networkingClient, lbID, "ACTIVE", nil, timeout)
	if err != nil {
		return err
	}

	return resourceListenerV2Read(d, meta)

}

func resourceListenerV2Delete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	networkingClient, err := config.networkingV2Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack networking client: %s", err)
	}

	// Wait for LoadBalancer to become active before continuing
	lbID := d.Get("loadbalancer_id").(string)
	timeout := d.Timeout(schema.TimeoutDelete)
	err = waitForLBV2LoadBalancer(networkingClient, lbID, "ACTIVE", nil, timeout)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Deleting listener %s", d.Id())
	err = resource.Retry(timeout, func() *resource.RetryError {
		err = listeners.Delete(networkingClient, d.Id()).ExtractErr()
		if err != nil {
			return checkForRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("Error deleting listener %s: %s", d.Id(), err)
	}

	// Wait for Listener to delete
	err = waitForLBV2Listener(networkingClient, d.Id(), "DELETED", nil, timeout)
	if err != nil {
		return err
	}

	return nil
}

func resourceLBV2ListenerRefreshFunc(networkingClient *gophercloud.ServiceClient, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		listener, err := listeners.Get(networkingClient, id).Extract()
		if err != nil {
			return nil, "", err
		}

		// The listener resource has no Status attribute, so a successful Get is the best we can do
		return listener, "ACTIVE", nil
	}
}

func waitForLBV2Listener(networkingClient *gophercloud.ServiceClient, id string, target string, pending []string, timeout time.Duration) error {
	log.Printf("[DEBUG] Waiting for listener %s to become %s.", id, target)

	stateConf := &resource.StateChangeConf{
		Target:     []string{target},
		Pending:    pending,
		Refresh:    resourceLBV2ListenerRefreshFunc(networkingClient, id),
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 1 * time.Second,
	}

	_, err := stateConf.WaitForState()
	if err != nil {
		if _, ok := err.(gophercloud.ErrDefault404); ok {
			switch target {
			case "DELETED":
				return nil
			default:
				return fmt.Errorf("Error: listener %s not found: %s", id, err)
			}
		}
		return fmt.Errorf("Error waiting for listener %s to become %s: %s", id, target, err)
	}

	return nil
}
