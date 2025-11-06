package acecloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	// "github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/AceCloudAI/terraform-provider-acecloud/acecloud/internal/client"
)


func configureProvider(ctx context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
	
	// terraformVersion := "1.0+" 
	var diags diag.Diagnostics

	// enableLogging := false
	// if logLevel := logging.LogLevel(); logLevel != "" {
	// 	if logLevel == "DEBUG" || logLevel == "TRACE" {
	// 		enableLogging = true
	// 	}
	// }

	apiEndpoint := d.Get("api_endpoint").(string)
	apiKey := d.Get("api_key").(string)
	region := d.Get("region").(string)
	projectID := d.Get("project_id").(string)

	c := client.NewAceCloudClient(apiEndpoint, apiKey, region, projectID)



	return c, diags
}