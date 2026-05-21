package key_pair

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// keyPairModel maps the resource schema to Go types.
type keyPairModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	PublicKey   types.String `tfsdk:"public_key"`
	PrivateKey  types.String `tfsdk:"private_key"`
	Fingerprint types.String `tfsdk:"fingerprint"`
}

func (r *keyPairResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Ace Cloud SSH key pair.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the key pair.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the key pair. Allowed characters: letters, numbers, hyphens.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"public_key": schema.StringAttribute{
				Description: "SSH public key in OpenSSH format (e.g. `ssh-rsa AAAA...` or `ssh-ed25519 AAAA...`). " +
					"Must be at least 100 characters; shorter values are almost always the result of a copy-paste " +
					"truncation or an HCL composition error, and the backend silently auto-generates a fresh keypair " +
					"in that case which is rarely what the user intended. If omitted entirely, the provider will " +
					"generate a keypair server-side and return the private key on create.",
				Optional: true,
				Validators: []validator.String{
					// 100 chars is a conservative floor: a minimal valid OpenSSH
					// public key (ssh-ed25519, no comment) is ~85 chars; ssh-rsa
					// 2048-bit starts around 380 chars. 100 catches the common
					// truncation / wrong-field-pasted cases at plan time.
					stringvalidator.LengthAtLeast(100),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"private_key": schema.StringAttribute{
				Description: "Generated private key in OpenSSH PEM format. Returned only on create when `public_key` is not supplied. Empty when the caller provides their own `public_key`.",
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"fingerprint": schema.StringAttribute{
				Description: "Fingerprint of the SSH public key.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}
