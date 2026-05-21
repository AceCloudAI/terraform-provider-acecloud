package floating_ip_association

import (
	"context"
	"fmt"
	"strings"

	"github.com/AceCloudAI/terraform-provider-acecloud/internal/client"
	"github.com/AceCloudAI/terraform-provider-acecloud/internal/wait"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const apiPath = "/cloud/floating-ips/action"

var (
	_ resource.Resource              = &floatingIPAssociationResource{}
	_ resource.ResourceWithConfigure = &floatingIPAssociationResource{}
)

type floatingIPAssociationResource struct {
	client *client.Client
}

// --- API types ---

type associateRequest struct {
	FloatingIPAddress string `json:"floating_ip_address"`
	InstanceID        string `json:"instance_id"`
	FixedIPAddress    string `json:"fixed_ip_address,omitempty"`
}

type disassociateRequest struct {
	FloatingIPAddress string `json:"floating_ip_address"`
	InstanceID        string `json:"instance_id"`
}

func NewResource() resource.Resource {
	return &floatingIPAssociationResource{}
}

func (r *floatingIPAssociationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_floating_ip_association"
}

func (r *floatingIPAssociationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = floatingIPAssociationSchema()
}

func (r *floatingIPAssociationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData),
		)
		return
	}
	r.client = c
}

func (r *floatingIPAssociationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan floatingIPAssociationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := associateRequest{
		FloatingIPAddress: plan.FloatingIPAddress.ValueString(),
		InstanceID:        plan.InstanceID.ValueString(),
	}
	if !plan.FixedIPAddress.IsNull() && !plan.FixedIPAddress.IsUnknown() {
		body.FixedIPAddress = plan.FixedIPAddress.ValueString()
	}

	// PUT /cloud/floating-ips/action?type=attach
	_, err := r.client.PutWithParams(ctx, apiPath, body, map[string]string{
		"type": "attach",
	})
	if err != nil {
		summary, detail := classifyAssociateError(err,
			plan.FloatingIPAddress.ValueString(),
			plan.InstanceID.ValueString())
		resp.Diagnostics.AddError(summary, detail)
		return
	}

	// Composite ID: floating_ip_address/instance_id
	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", plan.FloatingIPAddress.ValueString(), plan.InstanceID.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *floatingIPAssociationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state floatingIPAssociationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// No direct read API for association — the state is maintained by Terraform.
	// We could verify by reading the floating IP and checking its status,
	// but the association is managed as a side effect.
	// Keep existing state as-is.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *floatingIPAssociationResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Associations do not support update. Changes trigger destroy and recreate.
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Floating IP association does not support in-place updates. Changes will trigger a destroy and recreate.",
	)
}

func (r *floatingIPAssociationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state floatingIPAssociationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := disassociateRequest{
		FloatingIPAddress: state.FloatingIPAddress.ValueString(),
		InstanceID:        state.InstanceID.ValueString(),
	}

	// FIP detach can fail transiently when the instance port is in a
	// transitional state during concurrent destroy operations.
	// the API returns: "Cannot perform this action on the instance in current state"
	err := wait.RetryOnConflict(ctx, wait.RetryOnConflictOpts{
		Operation: func(ctx context.Context) error {
			// PUT /cloud/floating-ips/action?type=detach
			_, err := r.client.PutWithParams(ctx, apiPath, body, map[string]string{
				"type": "detach",
			})
			return err
		},
		RetryableErrors: []string{"Cannot perform this action", "in current state"},
		RetryAuthErrors: true,
	})
	if err != nil {
		summary, detail := classifyDisassociateError(err,
			state.FloatingIPAddress.ValueString(),
			state.InstanceID.ValueString())
		resp.Diagnostics.AddError(summary, detail)
		return
	}
}

// classifyAssociateError maps known attach-time API error patterns to
// actionable messages with a remediation hint. The original API error
// text is preserved in the detail block so support can still triage.
func classifyAssociateError(err error, fipAddr, instanceID string) (summary, detail string) {
	if err == nil {
		return "", ""
	}
	msg := err.Error()
	lower := strings.ToLower(msg)
	switch {
	case strings.Contains(lower, "already associated"),
		strings.Contains(lower, "already attached"),
		strings.Contains(lower, "already has a floating ip"):
		return "Floating IP is already associated",
			fmt.Sprintf("The floating IP %s appears to already be associated, possibly to instance %s "+
				"or another instance, via an out-of-band change (console or API). "+
				"Run `terraform refresh` so state matches reality, then re-plan. "+
				"Underlying API error: %s", fipAddr, instanceID, msg)
	case strings.Contains(lower, "not found"):
		return "Floating IP or instance not found",
			fmt.Sprintf("Either floating IP %s or instance %s no longer exists. "+
				"This usually means the resource was deleted via the console. "+
				"Run `terraform refresh` to reconcile state. Underlying API error: %s",
				fipAddr, instanceID, msg)
	default:
		return "Failed to associate floating IP", msg
	}
}

// classifyDisassociateError is the symmetric helper for detach-time
// failures. Out-of-band disassociation is by far the most common cause
// and deserves a clear remediation hint.
func classifyDisassociateError(err error, fipAddr, instanceID string) (summary, detail string) {
	if err == nil {
		return "", ""
	}
	msg := err.Error()
	lower := strings.ToLower(msg)
	switch {
	case strings.Contains(lower, "not associated"),
		strings.Contains(lower, "not attached"),
		strings.Contains(lower, "no floating ip"):
		return "Floating IP is already disassociated",
			fmt.Sprintf("The floating IP %s is no longer associated with instance %s "+
				"(probably detached via the console or API). "+
				"Run `terraform state rm` on the affected association resource to drop it from state, "+
				"or `terraform refresh` to let Terraform detect the drift. "+
				"Underlying API error: %s", fipAddr, instanceID, msg)
	case strings.Contains(lower, "not found"):
		return "Floating IP or instance not found",
			fmt.Sprintf("Either floating IP %s or instance %s no longer exists. "+
				"Treat the association as already removed by running `terraform state rm` on it. "+
				"Underlying API error: %s", fipAddr, instanceID, msg)
	default:
		return "Failed to disassociate floating IP", msg
	}
}
