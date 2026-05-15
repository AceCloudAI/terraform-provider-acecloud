package instance

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Validation patterns.
var (
	// Instance/resource name: alphanumeric + hyphens only.
	// Matches CLI: ResourceNameRegex = /^[a-zA-Z0-9-]+$/
	// Matches UI: INSTANCE_NAME_REGEX = /^[a-zA-Z0-9-]+$/
	resourceNameRegex = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)

	// Description: alphanumeric + underscore, hyphen, period, comma, space.
	// Matches CLI: DescriptionRegex = /^[a-zA-Z0-9_\-., ]*$/
	// Matches UI: DESCRIPTION_REGEX = /^[a-zA-Z0-9_\-., ]*$/
	descriptionRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-., ]*$`)
)

// instanceVolumeModel maps a single volume block in the Terraform plan.
type instanceVolumeModel struct {
	Size        types.Int64  `tfsdk:"size"`
	Boot        types.Bool   `tfsdk:"boot"`
	VolumeType  types.String `tfsdk:"volume_type"`
	BillingType types.String `tfsdk:"billing_type"`
}

// instanceResourceModel is the Terraform state / plan model for acecloud_instance.
type instanceResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Description         types.String `tfsdk:"description"`
	FlavorID            types.String `tfsdk:"flavor_id"`
	BootUUID            types.String `tfsdk:"boot_uuid"`
	SourceType          types.String `tfsdk:"source_type"`
	DeleteOnTermination types.Bool   `tfsdk:"delete_on_termination"`
	Volumes             types.List   `tfsdk:"volumes"`
	NetworkIDs          types.List   `tfsdk:"network_ids"`
	SecurityGroupIDs    types.List   `tfsdk:"security_group_ids"`
	Metadata            types.Map    `tfsdk:"metadata"`
	KeyName             types.String `tfsdk:"key_name"`
	UserData            types.String `tfsdk:"user_data"`
	AdminPassword       types.String `tfsdk:"admin_password"`
	BillingType         types.String `tfsdk:"billing_type"`
	PowerState          types.String `tfsdk:"power_state"`
	Status              types.String `tfsdk:"status"`
	PrivateIP           types.String `tfsdk:"private_ip"`
	PublicIP            types.String `tfsdk:"public_ip"`
}

// Schema defines the Terraform schema for the acecloud_instance resource.
func (r *instanceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Ace Cloud compute instance.",
		Attributes: map[string]schema.Attribute{
			// ── Computed identifier ───────────────────────────────────────────
			"id": schema.StringAttribute{
				Description: "Unique identifier of the instance.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// ── Required attributes ──────────────────────────────────────────
			"name": schema.StringAttribute{
				Description: "Human-readable name for the instance (1-255 characters, alphanumeric and hyphens only).",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
					stringvalidator.RegexMatches(resourceNameRegex, "must contain only letters, numbers, and hyphens"),
				},
			},
			"flavor_id": schema.StringAttribute{
				Description: "UUID of the compute flavor. Changing this resizes the instance in place — no recreation. The provider waits for the new flavor to become active before returning. Only upsize is supported; attempts to change to a smaller flavor will return an error.",
				Required:    true,
			},
			"boot_uuid": schema.StringAttribute{
				Description: "UUID of the boot source (image, snapshot, or volume).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source_type": schema.StringAttribute{
				Description: "Type of the boot source. Must be one of: image, snapshot, volume.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("image", "snapshot", "volume"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"delete_on_termination": schema.BoolAttribute{
				Description: "Whether the boot volume is deleted when the instance is terminated.",
				Required:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"network_ids": schema.ListAttribute{
				Description: "List of network UUIDs to attach. Maximum 7 networks.",
				Required:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"security_group_ids": schema.ListAttribute{
				Description: "List of security group UUIDs. Maximum 7 groups. Updated in-place when changed.",
				Required:    true,
				ElementType: types.StringType,
			},
			"billing_type": schema.StringAttribute{
				Description: "Billing type for the instance flavor. Valid: hourly, monthly, quarterly, half-yearly, yearly. Defaults to 'monthly'. Spot pricing is not exposed by this resource — it has a separate launch flow on the platform.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("monthly"),
				Validators: []validator.String{
					stringvalidator.OneOf("hourly", "monthly", "quarterly", "half-yearly", "yearly"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			// ── Optional attributes ──────────────────────────────────────────
			"description": schema.StringAttribute{
				Description: "Optional description (0-255 characters). Only letters, numbers, underscores, hyphens, periods, commas, and spaces allowed.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
					stringvalidator.RegexMatches(descriptionRegex, "can only contain letters, numbers, underscores, hyphens, periods, commas, and spaces"),
				},
			},
			"metadata": schema.MapAttribute{
				Description: "Arbitrary key/value metadata to attach to the instance.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"key_name": schema.StringAttribute{
				Description: "Name of the SSH key pair to inject into the instance.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_data": schema.StringAttribute{
				Description: "Base64-encoded cloud-init user data. The instance executes this at first boot to bootstrap itself (install packages, configure services, write files, etc.). Wrap your script with `base64encode(<<-EOT ... EOT)` in HCL. Optional; omit for instances that don't need first-boot customisation. Treated as sensitive because it commonly contains secrets or bootstrap tokens.",
				Optional:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"admin_password": schema.StringAttribute{
				Description: "Base64-encoded admin password (sensitive).",
				Optional:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"power_state": schema.StringAttribute{
				Description: "Power state of the instance: `ON` (running) or `OFF` (stopped). Defaults to `ON`. Changing this value calls the platform power action; the instance is not destroyed.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("ON"),
				Validators: []validator.String{
					stringvalidator.OneOf("ON", "OFF"),
				},
			},
			// ── Computed (read-only) ─────────────────────────────────────────
			"status": schema.StringAttribute{
				Description: "Current lifecycle status of the instance. Values: `BUILD` (provisioning), `ACTIVE` (powered on and ready), `SHUTOFF` / `STOPPED` (powered off), `REBOOT` / `HARD_REBOOT` (rebooting), `PAUSED` / `SUSPENDED` (paused), `RESIZE` / `VERIFY_RESIZE` (during a flavor change), `ERROR` (provisioning or runtime failure).",
				Computed:    true,
			},
			"private_ip": schema.StringAttribute{
				Description: "Primary private IP address on the instance's first attached network. Useful as a load-balancer pool member address. Empty until the instance reaches ACTIVE.",
				Computed:    true,
			},
			"public_ip": schema.StringAttribute{
				Description: "Public IP address of the instance, if any. Set when a floating IP is associated; empty otherwise.",
				Computed:    true,
			},
		},

		// ── Nested block: volumes ────────────────────────────────────────────
		Blocks: map[string]schema.Block{
			"volumes": schema.ListNestedBlock{
				Description: "Boot and data volumes to create with the instance.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"size": schema.Int64Attribute{
							Description: "Size of the volume in GB. Boot volume minimum 8 GB, max 16384 GB.",
							Required:    true,
							Validators: []validator.Int64{
								int64validator.Between(8, 16384),
							},
						},
						"boot": schema.BoolAttribute{
							Description: "Whether this is the boot volume.",
							Required:    true,
						},
						"volume_type": schema.StringAttribute{
							Description: "Storage tier name for this volume. The exact string must match a volume type available in the target region (for example `NVMe based High IOPS Storage`). The set of available types varies by region and account; see the AceCloud console for the current list.",
							Required:    true,
						},
						"billing_type": schema.StringAttribute{
							Description: "Billing type for the volume. Valid: hourly, monthly, quarterly, half-yearly, yearly. Defaults to 'hourly'.",
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString("hourly"),
							Validators: []validator.String{
								stringvalidator.OneOf("hourly", "monthly", "quarterly", "half-yearly", "yearly"),
							},
						},
					},
				},
			},
		},
	}
}
