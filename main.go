package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/oboukili/terraform-provider-argocd/argocd"
)

func main() {
	// ArgoCD services connection pool closing channel
	var doneCh = make(chan bool, 1)

	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() terraform.ResourceProvider {
			return argocd.Provider(doneCh)
		},
	})
	doneCh <- true
}
