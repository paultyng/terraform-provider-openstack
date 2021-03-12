package openstack

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gophercloud/gophercloud/openstack/blockstorage/extensions/quotasets"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceBlockStorageQuotasetV2() *schema.Resource {
	return &schema.Resource{
		Create: resourceBlockStorageQuotasetV2Create,
		Read:   resourceBlockStorageQuotasetV2Read,
		Update: resourceBlockStorageQuotasetV2Update,
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

			"volumes": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			"snapshots": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			"gigabytes": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			"per_volume_gigabytes": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			"backups": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			"backup_gigabytes": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			"groups": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			"volume_type_quota": {
				Type:     schema.TypeMap,
				Elem:     &schema.Schema{Type: schema.TypeInt},
				Optional: true,
			},
		},
	}
}

func resourceBlockStorageQuotasetV2Create(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	region := GetRegion(d, config)
	blockStorageClient, err := config.BlockStorageV2Client(region)
	if err != nil {
		return fmt.Errorf("Error creating OpenStack block storage client: %s", err)
	}

	projectID := d.Get("project_id").(string)
	volumes := d.Get("volumes").(int)
	snapshots := d.Get("snapshots").(int)
	gigabytes := d.Get("gigabytes").(int)
	perVolumeGigabytes := d.Get("per_volume_gigabytes").(int)
	backups := d.Get("backups").(int)
	backupGigabytes := d.Get("backup_gigabytes").(int)
	groups := d.Get("groups").(int)
	volumeTypeQuota := d.Get("volume_type_quota").(map[string]interface{})

	updateOpts := quotasets.UpdateOpts{
		Volumes:            &volumes,
		Snapshots:          &snapshots,
		Gigabytes:          &gigabytes,
		PerVolumeGigabytes: &perVolumeGigabytes,
		Backups:            &backups,
		BackupGigabytes:    &backupGigabytes,
		Groups:             &groups,
		Extra:              volumeTypeQuota,
	}

	q, err := quotasets.Update(blockStorageClient, projectID, updateOpts).Extract()
	if err != nil {
		return fmt.Errorf("Error creating openstack_blockstorage_quotaset_v2: %s", err)
	}

	id := fmt.Sprintf("%s/%s", projectID, region)
	d.SetId(id)

	log.Printf("[DEBUG] Created openstack_blockstorage_quotaset_v2 %#v", q)

	return resourceBlockStorageQuotasetV2Read(d, meta)
}

func resourceBlockStorageQuotasetV2Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	region := GetRegion(d, config)
	blockStorageClient, err := config.BlockStorageV2Client(region)
	if err != nil {
		return fmt.Errorf("Error creating OpenStack block storage client: %s", err)
	}

	// Depending on the provider version the resource was created, the resource id
	// can be either <project_id> or <project_id>/<region>. This parses the project_id
	// in both cases
	projectID := strings.Split(d.Id(), "/")[0]

	q, err := quotasets.Get(blockStorageClient, projectID).Extract()
	if err != nil {
		return CheckDeleted(d, err, "Error retrieving openstack_blockstorage_quotaset_v2")
	}

	log.Printf("[DEBUG] Retrieved openstack_blockstorage_quotaset_v2 %s: %#v", d.Id(), q)

	d.Set("project_id", projectID)
	d.Set("region", region)
	d.Set("volumes", q.Volumes)
	d.Set("snapshots", q.Snapshots)
	d.Set("gigabytes", q.Gigabytes)
	d.Set("per_volume_gigabytes", q.PerVolumeGigabytes)
	d.Set("backups", q.Backups)
	d.Set("backup_gigabytes", q.BackupGigabytes)
	d.Set("groups", q.Groups)

	// We only volume_type_quota when user is defining them to avoid unnecessary updates on every run
	// and not introduce breaking changes
	volumeTypeQuota := d.Get("volume_type_quota").(map[string]interface{})
	if len(volumeTypeQuota) > 0 {
		if err := d.Set("volume_type_quota", q.Extra); err != nil {
			log.Printf(
				"[WARN] Unable to set openstack_blockstorage_quotaset_v3 %s volume_type_quotas: %s", d.Id(), err)
		}
	}

	return nil
}

func resourceBlockStorageQuotasetV2Update(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	blockStorageClient, err := config.BlockStorageV2Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack block storage client: %s", err)
	}

	var (
		hasChange  bool
		updateOpts quotasets.UpdateOpts
	)

	if d.HasChange("volumes") {
		hasChange = true
		volumes := d.Get("volumes").(int)
		updateOpts.Volumes = &volumes
	}

	if d.HasChange("snapshots") {
		hasChange = true
		snapshots := d.Get("snapshots").(int)
		updateOpts.Snapshots = &snapshots
	}

	if d.HasChange("gigabytes") {
		hasChange = true
		gigabytes := d.Get("gigabytes").(int)
		updateOpts.Gigabytes = &gigabytes
	}

	if d.HasChange("per_volume_gigabytes") {
		hasChange = true
		perVolumeGigabytes := d.Get("per_volume_gigabytes").(int)
		updateOpts.PerVolumeGigabytes = &perVolumeGigabytes
	}

	if d.HasChange("backups") {
		hasChange = true
		backups := d.Get("backups").(int)
		updateOpts.Backups = &backups
	}

	if d.HasChange("backup_gigabytes") {
		hasChange = true
		backupGigabytes := d.Get("backup_gigabytes").(int)
		updateOpts.BackupGigabytes = &backupGigabytes
	}

	if d.HasChange("groups") {
		hasChange = true
		groups := d.Get("groups").(int)
		updateOpts.Groups = &groups
	}

	if d.HasChange("volume_type_quota") {
		_, newVTQRaw := d.GetChange("volume_type_quota")
		newVTQ := newVTQRaw.(map[string]interface{})

		// if len(newVTQ) == 0 it can lead to error when trying to do an update with
		// zero attributes. Not updating when a user removes all attributes is acceptable
		// as this attributes are not removed anyways
		if len(newVTQ) > 0 {
			hasChange = true
			volumeTypeQuota := d.Get("volume_type_quota").(map[string]interface{})
			updateOpts.Extra = volumeTypeQuota
		}
	}

	if hasChange {
		log.Printf("[DEBUG] openstack_blockstorage_quotaset_v2 %s update options: %#v", d.Id(), updateOpts)
		projectID := d.Get("project_id").(string)
		_, err := quotasets.Update(blockStorageClient, projectID, updateOpts).Extract()
		if err != nil {
			return fmt.Errorf("Error updating openstack_blockstorage_quotaset_v2: %s", err)
		}
	}

	return resourceBlockStorageQuotasetV2Read(d, meta)
}
