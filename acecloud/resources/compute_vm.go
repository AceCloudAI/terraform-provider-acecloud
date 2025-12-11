package resources

import (
	"context"
	"fmt"

	"github.com/AceCloudAI/terraform-provider-acecloud/acecloud/internal/client"
	"github.com/AceCloudAI/terraform-provider-acecloud/acecloud/internal/client/types"
	"github.com/AceCloudAI/terraform-provider-acecloud/acecloud/internal/helpers"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceAceCloudVM() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAceCloudVMCreate,

		ReadContext:   resourceAceCloudVMRead,
		UpdateContext: resourceAceCloudVMUpdate,
		DeleteContext: resourceVMDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the virtual machine instance",
			},
			"flavor": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Flavor ID for the VM instance",
			},
			"boot_uuid": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Boot image UUID",
			},
			"delete_on_termination": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether to delete volumes on VM termination",
			},
			"network": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of network IDs to attach",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"key": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "SSH key for accessing the VM",
				// Elem: &schema.Schema{
				//     Type: schema.TypeString,
				// },
			},
			"security_group": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of security group IDs to apply",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"source_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "image",
				Description: "Source type for boot device",
			},
			"availability_zone": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "nova",
				Description: "Availability zone for the VM",
			},
			"billing_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "hourly",
				Description: "Billing type for the VM",
			},
			"custom_update": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(types.PauseInstance),
								string(types.ResumeInstance),
								string(types.SoftRebootInstance),
								string(types.HardRebootInstance),
								string(types.LockInstance),
								string(types.UnlockInstance),
								string(types.CreateSnapshot),
								string(types.DetachInterface),
								string(types.AttachInterface),
								string(types.SuspendInstance),
								string(types.UnsuspendInstance),
								string(types.AttachVolume),
								string(types.DetachVolume),
								string(types.ShutOffInstance),
								string(types.StartInstance),
								string(types.RebuildInstance),
							}, false),
							Description: "The action to update something on the VM instance",
						},
						"snapshot_name": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Name of the snapshot to create (required if custom_update is create-snapshot)",
						},
						"interface_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The specific interface ID to remove from the VM. Detaches the selected NIC from the instance.(required if custom_update is detach-interface)",
						},

						"network_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Network ID to attach the interface to the VM. Attaches a new NIC to the instance. (required if custom_update is attach-interface)",
						},

						"volume_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: " Volume ID to attach to the VM. Attaches a new volume to the instance. (required if custom_update is attach-volume)",
						},

						"delete_on_termination": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether to delete the volume on VM termination. (used with attach-volume action)",
						},
						"disk_config": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Disk configuration for rebuilding the instance. (used with rebuild-instance action)",
						},

						"description": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Description for the VM instance. (used with rebuild-instance action)",
						},

						"image_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Image id that will be rebuilt (used with rebuild-instance action) ",
						},
					},
				},
			},

			"volumes": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of volumes to attach",

				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"boot": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether this is the boot volume",
						},
						"volume_type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Type of the volume",
						},
						"size": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Size of the volume in GB",
						},
						"billing_type": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "hourly",
							Description: "Billing type for the volume",
						},
					},
				},
			},
			"vm_count": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "Number of VM instances to create",
			},
			"instance_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The created VM instance ID",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Current status of the VM instance",
			},
			"ip_address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "IP address of the VM instance",
			},
		},
	}
}

func resourceAceCloudVMCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.AceCloudClient)

	req := &types.VMCreateRequest{
		Name:                d.Get("name").(string),
		Flavor:              d.Get("flavor").(string),
		BootUUID:            d.Get("boot_uuid").(string),
		DeleteOnTermination: d.Get("delete_on_termination").(bool),
		SourceType:          d.Get("source_type").(string),
		AvailabilityZone:    d.Get("availability_zone").(string),
		BillingType:         d.Get("billing_type").(string),
		Key:                 d.Get("key").(string),
		Count:               d.Get("vm_count").(int),
	}

	if v, ok := d.GetOk("network"); ok && v != nil {
		req.Networks = helpers.InterfaceSliceToStringSlice(v.([]interface{}))
	}
	if v, ok := d.GetOk("security_group"); ok && v != nil {
		req.SecurityGroups = helpers.InterfaceSliceToStringSlice(v.([]interface{}))
	}
	if v, ok := d.GetOk("volumes"); ok && v != nil {
		raw := v.([]interface{})
		vols := make([]types.VolumeRequest, 0, len(raw))
		for _, it := range raw {
			m := it.(map[string]interface{})
			size, _ := helpers.ConvertToInt(m["size"])
			boot := false
			if b, ok := m["boot"]; ok && b != nil {
				if bb, ok := b.(bool); ok {
					boot = bb
				}
			}
			billing := "hourly"
			if bt, ok := m["billing_type"]; ok && bt != nil {
				billing = bt.(string)
			}
			vols = append(vols, types.VolumeRequest{
				Boot:        boot,
				VolumeType:  m["volume_type"].(string),
				Size:        size,
				BillingType: billing,
			})
		}
		req.Volumes = vols
	}

	resp, err := c.CreateVM(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	id := resp.Data.ID
	d.SetId(id)
	_ = d.Set("instance_id", id)
	// Backend response doesn’t include status/ip yet; leave unset.

	// No read for now.
	return nil
}

func resourceAceCloudVMRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.AceCloudClient)

	id := d.Id()
	resp, err := c.GetVM(ctx, id)
	if err != nil {
		if helpers.IsNotFoundError(err) {

			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("instance_id", resp.Data.ID)
	_ = d.Set("status", resp.Data.Status)

	// Set IP address if available (first public address)
	if len(resp.Data.Addresses.Public) > 0 {
		_ = d.Set("ip_address", resp.Data.Addresses.Public[0].Addr)
	} else {
		_ = d.Set("ip_address", "")
	}

	return nil
}

func resourceVMDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.AceCloudClient)

	id := d.Id()
	if id == "" {
		return nil
	}

	_, err := c.DeleteVMs(ctx, []string{id})
	if err != nil {
		if helpers.IsNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceAceCloudVMUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.AceCloudClient)

	id := d.Id()
	if id == "" {
		return diag.Errorf("resource ID is empty")
	}

	if d.HasChange("name") {
		name := d.Get("name").(string)
		req := &types.VMUpdateRequest{
			Name: name,
		}

		_, err := c.UpdateVM(ctx, d, id, req, "")
		if err != nil {
			if helpers.IsNotFoundError(err) {
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
	}
	if d.HasChange("custom_update") {

		updates := d.Get("custom_update").([]interface{})
		if len(updates) == 0 {
			return nil
		}

		updateBlock := updates[0].(map[string]interface{})

		rawAction := updateBlock["action"].(string)
		action := types.VMupdateAction(rawAction)

		tflog.Debug(ctx, fmt.Sprintf("Action: %s", action))

		switch action {

		case types.PauseInstance,
			types.ResumeInstance,
			types.SoftRebootInstance,
			types.HardRebootInstance,
			types.LockInstance,
			types.UnlockInstance,
			types.DetachInterface,
			types.SuspendInstance,
			types.UnsuspendInstance,
			types.DetachVolume,
			types.ShutOffInstance,
			types.StartInstance:
			emptyBody := map[string]interface{}{}
			_, err := c.UpdateVM(ctx, d, id, emptyBody, action)
			if err != nil {
				if helpers.IsNotFoundError(err) {
					d.SetId("")
					return nil
				}
				return diag.FromErr(err)
			}

		case types.CreateSnapshot:
			snapshotName := ""
			if v, ok := updateBlock["snapshot_name"]; ok {
				snapshotName = v.(string)
			}

			if snapshotName == "" {
				return diag.Errorf("snapshot_name is required when action=create-snapshot")
			}

			// Extract billing type from boot volume
			var billingType string
			vols := d.Get("volumes").([]interface{})
			for _, v := range vols {
				vol := v.(map[string]interface{})
				if vol["boot"].(bool) {
					billingType = vol["billing_type"].(string)
					break
				}
			}

			if billingType == "" {
				billingType = "hourly" // fallback if needed
			}

			req := &types.VMUpdateRequest{
				Name:        snapshotName,
				BillingType: billingType,
			}

			tflog.Debug(ctx, fmt.Sprintf("Creating snapshot '%s' with billing_type '%s'", snapshotName, billingType))

			_, err := c.UpdateVM(ctx, d, id, req, action)
			if err != nil {
				if helpers.IsNotFoundError(err) {
					d.SetId("")
					return nil
				}
				return diag.FromErr(err)
			}

		case types.AttachInterface:
			networkId := ""
			if v, ok := updateBlock["network_id"]; ok {
				networkId = v.(string)
			}

			if networkId == "" {
				return diag.Errorf("network_id is required when action = attach-interface")
			}

			// Extract billing type from boot volume
			var billingType string
			vols := d.Get("volumes").([]interface{})
			tflog.Debug(ctx, fmt.Sprintf("Volume array'%s'", vols))

			for _, v := range vols {
				vol := v.(map[string]interface{})
				if vol["boot"].(bool) {
					billingType = vol["billing_type"].(string)
					break
				}
			}

			if billingType == "" {
				billingType = "hourly" // fallback if needed
			}

			req := &types.VMUpdateRequest{
				NetworkId:   networkId,
				BillingType: billingType,
			}

			_, err := c.UpdateVM(ctx, d, id, req, action)
			if err != nil {
				if helpers.IsNotFoundError(err) {
					d.SetId("")
					return nil
				}
				return diag.FromErr(err)
			}

		case types.AttachVolume:
			vID, ok1 := updateBlock["volume_id"]
			if !ok1 || vID == "" {
				return diag.Errorf("volume_id is required when action = attach-volume")
			}
			volumeId := vID.(string)

			dotVal, ok2 := updateBlock["delete_on_termination"]
			if !ok2 {
				return diag.Errorf("delete_on_termination is required when action = attach-volume")
			}
			deleteOnTermination := dotVal.(bool)

			req := &types.VMUpdateRequest{
				VolumeId:            volumeId,
				DeleteOnTermination: deleteOnTermination,
			}

			_, err := c.UpdateVM(ctx, d, id, req, action)
			if err != nil {
				if helpers.IsNotFoundError(err) {
					d.SetId("")
					return nil
				}
				return diag.FromErr(err)
			}

		case types.RebuildInstance:
			diskConfig, ok := updateBlock["disk_config"].(string)
			if !ok || diskConfig == "" {
				return diag.Errorf("disk_config is required for rebuild-instance")
			}

			imageId, ok := updateBlock["image_id"].(string)
			if !ok || imageId == "" {
				return diag.Errorf("image_id is required for rebuild-instance")
			}

			description := ""
			if v, ok := updateBlock["description"]; ok {
				description = v.(string)
			}

			body := map[string]interface{}{
				"disk_config": diskConfig,
				"image_id":    imageId,
			}

			// only include description if user provided it
			if description != "" {
				body["description"] = description
			}

			_, err := c.UpdateVM(ctx, d, id, body, action)
			if err != nil {
				if helpers.IsNotFoundError(err) {
					d.SetId("")
					return nil
				}
				return diag.FromErr(err)
			}

		}
	}

	return resourceAceCloudVMRead(ctx, d, meta)
}
