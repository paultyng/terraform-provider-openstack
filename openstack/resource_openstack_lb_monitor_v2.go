package openstack

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/lbaas_v2/monitors"
)

func resourceMonitorV2() *schema.Resource {
	return &schema.Resource{
		Create: resourceMonitorV2Create,
		Read:   resourceMonitorV2Read,
		Update: resourceMonitorV2Update,
		Delete: resourceMonitorV2Delete,

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

			"pool_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"tenant_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"delay": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"timeout": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"max_retries": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"url_path": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"http_method": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"expected_codes": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
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

func resourceMonitorV2Create(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	networkingClient, err := config.networkingV2Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack networking client: %s", err)
	}

	adminStateUp := d.Get("admin_state_up").(bool)
	createOpts := monitors.CreateOpts{
		PoolID:        d.Get("pool_id").(string),
		TenantID:      d.Get("tenant_id").(string),
		Type:          d.Get("type").(string),
		Delay:         d.Get("delay").(int),
		Timeout:       d.Get("timeout").(int),
		MaxRetries:    d.Get("max_retries").(int),
		URLPath:       d.Get("url_path").(string),
		HTTPMethod:    d.Get("http_method").(string),
		ExpectedCodes: d.Get("expected_codes").(string),
		Name:          d.Get("name").(string),
		AdminStateUp:  &adminStateUp,
	}

	log.Printf("[DEBUG] Create Options: %#v", createOpts)
	log.Printf("[DEBUG] Attempting to create monitor")
	var monitor *monitors.Monitor
	timeout := d.Timeout(schema.TimeoutCreate)
	err = resource.Retry(timeout, func() *resource.RetryError {
		monitor, err = monitors.Create(networkingClient, createOpts).Extract()
		if err != nil {
			return checkForRetryableError(err)
		}
		return nil
	})

	// Wait for monitor to become active before continuing
	err = waitForLBV2Monitor(networkingClient, monitor.ID, "ACTIVE", nil, timeout)
	if err != nil {
		return err
	}

	d.SetId(monitor.ID)

	return resourceMonitorV2Read(d, meta)
}

func resourceMonitorV2Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	networkingClient, err := config.networkingV2Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack networking client: %s", err)
	}

	monitor, err := monitors.Get(networkingClient, d.Id()).Extract()
	if err != nil {
		return CheckDeleted(d, err, "monitor")
	}

	log.Printf("[DEBUG] Retrieved monitor %s: %#v", d.Id(), monitor)

	d.Set("id", monitor.ID)
	d.Set("tenant_id", monitor.TenantID)
	d.Set("type", monitor.Type)
	d.Set("delay", monitor.Delay)
	d.Set("timeout", monitor.Timeout)
	d.Set("max_retries", monitor.MaxRetries)
	d.Set("url_path", monitor.URLPath)
	d.Set("http_method", monitor.HTTPMethod)
	d.Set("expected_codes", monitor.ExpectedCodes)
	d.Set("admin_state_up", monitor.AdminStateUp)
	d.Set("name", monitor.Name)
	d.Set("region", GetRegion(d, config))

	return nil
}

func resourceMonitorV2Update(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	networkingClient, err := config.networkingV2Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack networking client: %s", err)
	}

	var updateOpts monitors.UpdateOpts
	if d.HasChange("url_path") {
		updateOpts.URLPath = d.Get("url_path").(string)
	}
	if d.HasChange("expected_codes") {
		updateOpts.ExpectedCodes = d.Get("expected_codes").(string)
	}
	if d.HasChange("delay") {
		updateOpts.Delay = d.Get("delay").(int)
	}
	if d.HasChange("timeout") {
		updateOpts.Timeout = d.Get("timeout").(int)
	}
	if d.HasChange("max_retries") {
		updateOpts.MaxRetries = d.Get("max_retries").(int)
	}
	if d.HasChange("admin_state_up") {
		asu := d.Get("admin_state_up").(bool)
		updateOpts.AdminStateUp = &asu
	}
	if d.HasChange("name") {
		updateOpts.Name = d.Get("name").(string)
	}
	if d.HasChange("http_method") {
		updateOpts.HTTPMethod = d.Get("http_method").(string)
	}

	log.Printf("[DEBUG] Updating monitor %s with options: %#v", d.Id(), updateOpts)
	timeout := d.Timeout(schema.TimeoutUpdate)
	err = resource.Retry(timeout, func() *resource.RetryError {
		_, err = monitors.Update(networkingClient, d.Id(), updateOpts).Extract()
		if err != nil {
			return checkForRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("Unable to update monitor %s: %s", d.Id(), err)
	}

	// Wait for monitor to become active before continuing
	err = waitForLBV2Monitor(networkingClient, d.Id(), "ACTIVE", nil, timeout)
	if err != nil {
		return err
	}

	return resourceMonitorV2Read(d, meta)
}

func resourceMonitorV2Delete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	networkingClient, err := config.networkingV2Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack networking client: %s", err)
	}

	log.Printf("[DEBUG] Deleting monitor %s", d.Id())
	timeout := d.Timeout(schema.TimeoutUpdate)
	err = resource.Retry(timeout, func() *resource.RetryError {
		err = monitors.Delete(networkingClient, d.Id()).ExtractErr()
		if err != nil {
			return checkForRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("Unable to delete monitor %s: %s", d.Id(), err)
	}

	// Wait for monitor to delete
	err = waitForLBV2Monitor(networkingClient, d.Id(), "DELETED", nil, timeout)
	if err != nil {
		return err
	}

	return nil
}

func resourceLBV2MonitorRefreshFunc(networkingClient *gophercloud.ServiceClient, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		monitor, err := monitors.Get(networkingClient, id).Extract()
		if err != nil {
			return nil, "", err
		}

		// The monitor resource has no Status attribute, so a successful Get is the best we can do
		return monitor, "ACTIVE", nil
	}
}

func waitForLBV2Monitor(networkingClient *gophercloud.ServiceClient, id string, target string, pending []string, timeout time.Duration) error {
	log.Printf("[DEBUG] Waiting for monitor %s to become %s.", id, target)

	stateConf := &resource.StateChangeConf{
		Target:     []string{target},
		Pending:    pending,
		Refresh:    resourceLBV2MonitorRefreshFunc(networkingClient, id),
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
				return fmt.Errorf("Error: monitor %s not found: %s", id, err)
			}
		}
		return fmt.Errorf("Error waiting for monitor %s to become %s: %s", id, target, err)
	}

	return nil
}
