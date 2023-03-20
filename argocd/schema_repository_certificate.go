package argocd

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func certificatesSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"ssh": {
			Type:          schema.TypeList,
			Optional:      true,
			ForceNew:      true,
			MaxItems:      1,
			Description:   "Defines a `ssh` certificate.",
			ConflictsWith: []string{"https"},
			AtLeastOneOf:  []string{"https", "ssh"},
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"server_name": {
						Type:        schema.TypeString,
						Description: "DNS name of the server this certificate is intended for.",
						Required:    true,
						ForceNew:    true,
					},
					"cert_subtype": {
						Type:        schema.TypeString,
						Description: "The sub type of the cert, i.e. `ssh-rsa`.",
						Required:    true,
						ForceNew:    true,
					},
					"cert_data": {
						Type:        schema.TypeString,
						Description: "The actual certificate data, dependent on the certificate type.",
						Required:    true,
						ForceNew:    true,
					},
					"cert_info": {
						Type:        schema.TypeString,
						Description: "Additional certificate info, dependent on the certificate type (e.g. SSH fingerprint, X509 CommonName).",
						Computed:    true,
					},
				},
			},
		},
		"https": {
			Type:          schema.TypeList,
			Optional:      true,
			ForceNew:      true,
			MaxItems:      1,
			ConflictsWith: []string{"ssh"},
			AtLeastOneOf:  []string{"https", "ssh"},
			Description:   "Defines a `https` certificate.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"server_name": {
						Type:        schema.TypeString,
						Description: "DNS name of the server this certificate is intended for.",
						Required:    true,
						ForceNew:    true,
					},
					"cert_data": {
						Type:        schema.TypeString,
						Description: "The actual certificate data, dependent on the certificate type.",
						Required:    true,
						ForceNew:    true,
					},
					"cert_subtype": {
						Type:        schema.TypeString,
						Description: "The sub type of the cert, i.e. `ssh-rsa`.",
						Computed:    true,
					},
					"cert_info": {
						Type:        schema.TypeString,
						Description: "Additional certificate info, dependent on the certificate type (e.g. SSH fingerprint, X509 CommonName).",
						Computed:    true,
					},
				},
			},
		},
	}
}
