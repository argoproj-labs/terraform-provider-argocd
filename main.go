package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/oboukili/terraform-provider-argocd/argocd"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: argocd.Provider,
	})
}
