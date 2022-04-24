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
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: certificatesSchema(),
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
	c := *server.CertificateClient
	certs := application.RepositoryCertificateList{
		Items: []application.RepositoryCertificate{
			*repoCertificate,
		},
	}

	tokenMutexConfiguration.Lock()
	rc, err := c.CreateCertificate(
		ctx,
		&certificate.RepositoryCertificateCreateRequest{
			Certificates: &certs,
			Upsert:       false,
		},
	)
	_ = rc
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
	// If certificate already exists and didn't change, the response will be empty but since the call is success
	// we assume everything went fine and get the id from the request
	// if len(rc.Items) > 0 {
	// d.SetId(getId(&rc.Items[0])) // for https, certType is not returned from create call, but properly return on list call
	// } else {
	d.SetId(getId(repoCertificate))
	// d.Set("ssh.0.cert_data", repoCertificate.CertData)
	// }
	return resourceArgoCDRepositoryCertificatesRead(context.WithValue(ctx, "cert_data", repoCertificate.CertData), d, meta)
}

// Compute resource's id as : serverName/certType/certSubType
func getId(rc *application.RepositoryCertificate) string {
	if rc.CertType == "ssh" {
		return fmt.Sprintf("%s/%s/%s", rc.CertType, rc.CertSubType, rc.ServerName)
	} else {
		return fmt.Sprintf("%s/%s", rc.CertType, rc.ServerName)
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
	c := *server.CertificateClient
	repoCertificate := application.RepositoryCertificate{}
	err, certType, certSubType, serverName := fromId(d.Id())
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Failed to parse state"),
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
		if getId(&_rc) == d.Id() {
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
	c := *server.CertificateClient
	err, certType, certSubType, serverName := fromId(d.Id())
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Failed to parse state"),
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
