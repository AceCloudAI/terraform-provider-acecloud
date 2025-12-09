package resources

import (
	"context"
	"fmt"

	"github.com/AceCloudAI/terraform-provider-acecloud/acecloud/internal/client"
	"github.com/AceCloudAI/terraform-provider-acecloud/acecloud/internal/client/types"
	"github.com/AceCloudAI/terraform-provider-acecloud/acecloud/internal/helpers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ResourceKeyPair returns the Terraform resource for managing key-pairs.
func ResourceKeyPair() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeyPairCreate,
		ReadContext:   resourceKeyPairRead,
		UpdateContext: resourceKeyPairUpdate,
		DeleteContext: resourceKeyPairDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name for the key pair",
			},
			"public_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Public key string (if provided or returned)",
			},
			"private_key": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				// Sensitive:   true,
				Description: "Private key returned at creation (sensitive)",
			},
			"fingerprint": {
				//*this is byte stream technically but we represent as string
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Fingerprint of the key",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Type of the key pair",
			},
		},
	}
}

func resourceKeyPairCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.AceCloudClient)

	//*Building request from schema
	req := types.KeyPairCreateRequest{}
	if v, ok := d.GetOk("name"); ok {
		req.Name = v.(string)
	}

	//*Call client to create keypair
	kp, err := c.CreateKeyPair(ctx, &req)
	if err != nil {
		return diag.FromErr(err)
	}

	//*Set Terraform ID and state fields based on API response
	d.SetId(kp.ID)
	_ = d.Set("name", kp.Name)
	_ = d.Set("public_key", kp.PublicKey)
	_ = d.Set("private_key", kp.PrivateKey)
	_ = d.Set("fingerprint", kp.Fingerprint)
	_ = d.Set("type", kp.Type)

	// Refresh state using Read

	// return resourceKeyPairRead(ctx, d, meta)
	return nil
}

func resourceKeyPairRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.AceCloudClient)

	// Read identifying info from state/config

	id := d.Id()
	if id == "" {
		// nothing to do
		return nil
	}

	kp, err := c.GetKeyPair(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}
	if kp == nil {
		// Remote resource missing -> remove from state
		d.SetId("")
		return nil
	}

	// Update state with remote values
	_ = d.Set("name", kp.Name)
	_ = d.Set("public_key", kp.PublicKey)
	_ = d.Set("private_key", kp.PrivateKey)
	_ = d.Set("fingerprint", kp.Fingerprint)

	// Best practice: do not persist private key on subsequent reads unless API returns it.
	// _ = d.Set("private_key", "")

	return nil
}

func resourceKeyPairDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.AceCloudClient)

	id := d.Id()
	if id == "" {
		return nil
	}

	//*Building delete request from schema

	req := types.KeyPairDeleteRequestFromIDs()
	if v, ok := d.GetOk("name"); ok {
		req.Values = append(req.Values, v.(string))
	}

	if err := c.DeleteKeyPair(ctx, req, id); err != nil {
		// If already deleted on backend, treat as success and remove from state
		if helpers.IsNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(fmt.Errorf("delete keypair: %w", err))
	}

	d.SetId("")
	return nil
}

func resourceKeyPairUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Key pairs are typically immutable; no update operation is supported.
	// To change a key pair, it must be deleted and recreated.
	return resourceKeyPairRead(ctx, d, meta)
}
