package instance

import (
	"context"
	"encoding/json"

	"github.com/AceCloudAI/terraform-provider-acecloud/internal/client"
	tfpath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// ─────────────────────────────────────────────────────────────────────────
// Plan-time rejection: GPU flavor combined with metered billing.
//
// GPU flavors do not support hourly or spot billing. The validator below
// looks up the configured `flavor_id` against `/cloud/flavors` (the same
// endpoint the `acecloud_flavors` data source uses) and flags the
// combination at `terraform plan` time so no resources are created.
//
// Edge cases:
//   - flavor_id unknown at plan time (e.g. from a data-source filter or
//     locals expression that can't be resolved yet): validator no-ops; the
//     apply-time check still fires as a last line of defence.
//   - flavor_id refers to a flavor the lookup can't find: no-op, same as
//     above.
//   - flavor lookup itself fails (network blip / auth issue): no-op — we
//     don't want to block plan on a transient API issue; the apply-time
//     check still catches the bad config.
// ─────────────────────────────────────────────────────────────────────────

// meteredBillingTypes lists the billing types that are not allowed on GPU
// flavors.
var meteredBillingTypes = map[string]struct{}{
	"hourly": {},
	"spot":   {},
}

// flavorIsGPU looks up the given flavor ID against the live flavors list and
// reports whether it is a GPU flavor. The bool `known` indicates whether the
// lookup produced a definitive answer; on any uncertainty (flavor not in
// list, network error, API rejection) it returns known=false and the caller
// should skip the plan-time check rather than block on uncertain data.
func flavorIsGPU(ctx context.Context, c *client.Client, flavorID string) (isGPU, known bool) {
	if flavorID == "" {
		return false, false
	}
	apiResp, err := c.Get(ctx, "/cloud/flavors", nil)
	if err != nil {
		return false, false
	}
	var raw []struct {
		ID         string `json:"id"`
		ExtraSpecs struct {
			GPU bool `json:"gpu"`
		} `json:"extra_specs"`
	}
	if err := json.Unmarshal(apiResp.Data, &raw); err != nil {
		return false, false
	}
	for _, f := range raw {
		if f.ID == flavorID {
			return f.ExtraSpecs.GPU, true
		}
	}
	return false, false
}

// ModifyPlan implements resource.ResourceWithModifyPlan and runs the
// GPU + metered-billing pre-flight check described above.
func (r *instanceResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Destroy plan: Plan.Raw is null. Skip — nothing to validate.
	if req.Plan.Raw.IsNull() {
		return
	}
	// Provider hasn't been configured (e.g. validate-only run with no
	// credentials). The client is nil; skip the live lookup.
	if r.client == nil {
		return
	}

	var billingType, flavorID string
	if diags := req.Plan.GetAttribute(ctx, tfpath.Root("billing_type"), &billingType); diags.HasError() {
		// Schema-level validation will catch type mismatches; we just bail.
		return
	}
	if _, metered := meteredBillingTypes[billingType]; !metered {
		// Only metered billing types are restricted on GPU flavors. Anything
		// else is fine.
		return
	}
	if diags := req.Plan.GetAttribute(ctx, tfpath.Root("flavor_id"), &flavorID); diags.HasError() {
		return
	}
	if flavorID == "" {
		// flavor_id is unknown / not yet resolvable. The apply-time check
		// is the last line of defence.
		return
	}

	isGPU, known := flavorIsGPU(ctx, r.client, flavorID)
	if !known {
		// Lookup didn't give a definitive answer. Don't block plan on
		// uncertainty — the apply-time check still fires.
		return
	}
	if !isGPU {
		return
	}

	resp.Diagnostics.AddAttributeError(
		tfpath.Root("billing_type"),
		"GPU flavor cannot use metered billing",
		"GPU flavors do not support hourly or spot billing. "+
			"Use monthly, quarterly, half-yearly, or yearly billing for GPU instances.",
	)
}

