package acecloud

import (
	"github.com/AceCloudAI/terraform-provider-acecloud/acecloud/resources"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	var descriptions = map[string]string{
		"api_endpoint": "The base URL for the AceCloud API endpoint.",
		"api_key":      "The API key used to authenticate with AceCloud services.",
		"region":       "The AceCloud region to deploy resources in.",
		"project_id":   "The project ID for organizing resources in AceCloud.",
		"client_id":    "The tenant/client ID for AceCloud account identification.",
		"user_id":      "The user ID for AceCloud account access.",
	}

	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_endpoint": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ACECLOUD_API_ENDPOINT", nil),
				Description: descriptions["api_endpoint"],
			},
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("ACECLOUD_API_KEY", nil),
				Description: descriptions["api_key"],
			},
			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "us-east-1",
				Description: descriptions["region"],
			},
			"project_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: descriptions["project_id"],
			},
			"client_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: descriptions["client_id"],
			},
			"user_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: descriptions["user_id"],
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"acecloud_vm": resources.ResourceAceCloudVM(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			// "acecloud_flavor": dataSourceAceCloudFlavor(),
			// "acecloud_image": dataSourceAceCloudImage(),
			// "acecloud_network": dataSourceAceCloudNetwork(),
			// "acecloud_security_group": dataSourceAceCloudSecurityGroup(),
			// "acecloud_volume_type": dataSourceAceCloudVolumeType(),
			// "acecloud_availability_zone": dataSourceAceCloudAvailabilityZone(),
		},
		ConfigureContextFunc: configureProvider,
	}
}
