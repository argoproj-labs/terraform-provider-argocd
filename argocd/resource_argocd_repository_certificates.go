package argocd

import (
	"context"
	"fmt"
	"strings"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient/certificate"
	application "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceArgoCDRepositoryCertificates() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceArgoCDRepositoryCertificatesCreate,
		ReadContext:   resourceArgoCDRepositoryCertificatesRead,
		DeleteContext: resourceArgoCDRepositoryCertificatesDelete,
		Schema:        certificatesSchema(),
	}
}

func resourceArgoCDRepositoryCertificatesCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	repoCertificate, err := expandRepositoryCertificate(d)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("certificate %s could not be created", d.Id()),
				Detail:   err.Error(),
			},
		}
	}

	server := meta.(*ServerInterface)
	if err := server.initClients(); err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Failed to init clients"),
				Detail:   err.Error(),
			},
		}
	}

	featureRepositoryCertificateSupported, err := server.isFeatureSupported(featureRepositoryCertificates)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "feature not supported",
				Detail:   err.Error(),
			},
		}
	}

	if !featureRepositoryCertificateSupported {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary: fmt.Sprintf(
					"repository certificate is only supported from ArgoCD %s onwards",
					featureVersionConstraintsMap[featureRepositoryCertificates].String()),
			},
		}
	}
	c := *server.CertificateClient

	// Not doing a RLock here because we can have a race-condition between the ListCertificates & CreateCertificate
	tokenMutexConfiguration.Lock()

	if repoCertificate.CertType == "https" {
		rcl, err := c.ListCertificates(ctx, &certificate.RepositoryCertificateQuery{
			HostNamePattern: repoCertificate.ServerName,
			CertType:        repoCertificate.CertType,
			CertSubType:     repoCertificate.CertSubType,
		})

		if err != nil {
			tokenMutexConfiguration.Unlock()
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("certificates for host %s could not be listed", repoCertificate.ServerName),
					Detail:   err.Error(),
				},
			}
		}

		if len(rcl.Items) > 0 {
			tokenMutexConfiguration.Unlock()
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("https certificate for '%s' already exist.", repoCertificate.ServerName),
				},
			}
		}
	}

	certs := application.RepositoryCertificateList{
		Items: []application.RepositoryCertificate{
			*repoCertificate,
		},
	}

	rc, err := c.CreateCertificate(
		ctx,
		&certificate.RepositoryCertificateCreateRequest{
			Certificates: &certs,
			Upsert:       false,
		},
	)
	tokenMutexConfiguration.Unlock()

	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("certificate for repository %s could not be created", repoCertificate.ServerName),
				Detail:   err.Error(),
			},
		}
	}

	// TODO: upstream bug : if https certificate already exists, the response will be empty
	// instead of erroring about missing upsert flag but since the call is success
	// we assume everything went fine and get the id from the request
	var resourceId string
	if len(rc.Items) > 0 {
		err, resourceId = getId(&rc.Items[0])
	} else {
		err, resourceId = getId(repoCertificate)
	}

	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("certificate for repository %s created but could not be handled", repoCertificate.ServerName),
				Detail:   err.Error(),
			},
		}
	}
	d.SetId(resourceId)
	return resourceArgoCDRepositoryCertificatesRead(context.WithValue(ctx, "cert_data", repoCertificate.CertData), d, meta)
}

// Compute resource's id as :
// for ssh -> certType/certSubType/serverName
// for https -> certType/serverName
func getId(rc *application.RepositoryCertificate) (error, string) {
	if rc.CertType == "ssh" {
		if rc.CertSubType == "" || rc.ServerName == "" {
			return fmt.Errorf("invalid certificate: %s %s %s", rc.CertType, rc.CertSubType, rc.ServerName), ""
		}
		return nil, fmt.Sprintf("%s/%s/%s", rc.CertType, rc.CertSubType, rc.ServerName)
	} else {
		if rc.ServerName == "" {
			return fmt.Errorf("invalid certificate: %s %s", rc.CertType, rc.ServerName), ""
		}
		return nil, fmt.Sprintf("%s/%s", rc.CertType, rc.ServerName)
	}
}

// Get serverName/certType/certSubType from resource's id
func fromId(id string) (error, string, string, string) {
	parts := strings.Split(id, "/")
	if len(parts) < 2 {
		return fmt.Errorf("Unknown certificate %s in state", id), "", "", ""
	}
	certType := parts[0]
	if certType == "ssh" {
		if len(parts) < 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
			return fmt.Errorf("Unknown certificate %s in state", id), "", "", ""
		}
		return nil, parts[0], parts[1], parts[2]
	} else {
		if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
			return fmt.Errorf("Unknown certificate %s in state", id), "", "", ""
		}
		return nil, parts[0], parts[1], ""
	}
}

func resourceArgoCDRepositoryCertificatesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	server := meta.(*ServerInterface)
	if err := server.initClients(); err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Failed to init clients"),
				Detail:   err.Error(),
			},
		}
	}

	featureRepositoryCertificateSupported, err := server.isFeatureSupported(featureRepositoryCertificates)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "feature not supported",
				Detail:   err.Error(),
			},
		}
	}

	if !featureRepositoryCertificateSupported {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary: fmt.Sprintf(
					"repository certificate is only supported from ArgoCD %s onwards",
					featureVersionConstraintsMap[featureRepositoryCertificates].String()),
			},
		}
	}

	c := *server.CertificateClient
	repoCertificate := application.RepositoryCertificate{}
	err, certType, certSubType, serverName := fromId(d.Id())
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Failed to parse certificate state"),
				Detail:   err.Error(),
			},
		}
	}

	tokenMutexConfiguration.RLock()
	rcl, err := c.ListCertificates(ctx, &certificate.RepositoryCertificateQuery{
		HostNamePattern: serverName,
		CertType:        certType,
		CertSubType:     certSubType,
	})
	tokenMutexConfiguration.RUnlock()

	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("certificates for host %s could not be listed", d.Id()),
				Detail:   err.Error(),
			},
		}
	}
	if rcl == nil || len(rcl.Items) == 0 {
		// Certificate have already been deleted in an out-of-band fashion
		d.SetId("")
		return nil
	}
	for i, _rc := range rcl.Items {
		err, resourceId := getId(&_rc)
		if err != nil {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("certificate for repository %s could not be handled", repoCertificate.ServerName),
					Detail:   err.Error(),
				},
			}
		}
		if resourceId == d.Id() {
			repoCertificate = _rc
			break
		}
		// Certificate have already been deleted in an out-of-band fashion
		if i == len(rcl.Items)-1 {
			d.SetId("")
			return nil
		}
	}

	err = flattenRepositoryCertificate(&repoCertificate, d, ctx)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("certificate %s could not be flattened", d.Id()),
				Detail:   err.Error(),
			},
		}
	}
	return nil
}

func resourceArgoCDRepositoryCertificatesDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	server := meta.(*ServerInterface)
	if err := server.initClients(); err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Failed to init clients"),
				Detail:   err.Error(),
			},
		}
	}

	featureRepositoryCertificateSupported, err := server.isFeatureSupported(featureRepositoryCertificates)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "feature not supported",
				Detail:   err.Error(),
			},
		}
	}

	if !featureRepositoryCertificateSupported {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary: fmt.Sprintf(
					"repository certificate is only supported from ArgoCD %s onwards",
					featureVersionConstraintsMap[featureRepositoryCertificates].String()),
			},
		}
	}

	c := *server.CertificateClient
	err, certType, certSubType, serverName := fromId(d.Id())
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Failed to parse certificate state"),
				Detail:   err.Error(),
			},
		}
	}

	query := certificate.RepositoryCertificateQuery{
		HostNamePattern: serverName,
		CertType:        certType,
		CertSubType:     certSubType,
	}

	tokenMutexConfiguration.Lock()
	_, err = c.DeleteCertificate(
		ctx,
		&query,
	)
	tokenMutexConfiguration.Unlock()

	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			// Certificate have already been deleted in an out-of-band fashion
			d.SetId("")
			return nil
		} else {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("certificate for repository %s could not be deleted", d.Id()),
					Detail:   err.Error(),
				},
			}
		}
	}
	d.SetId("")
	return nil
}
