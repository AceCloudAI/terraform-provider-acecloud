package main

import (
	"flag"

	"github.com/AceCloudAI/terraform-provider-acecloud/acecloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

const providerAddr = "registry.terraform.io/acecloud/acecloud"

func main() {

	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	plugin.Serve(&plugin.ServeOpts{
		Debug:        debugMode,
		ProviderAddr: providerAddr,
		ProviderFunc: acecloud.Provider,
	})
}
