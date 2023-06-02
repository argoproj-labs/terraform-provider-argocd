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

type (
	certDataKey struct{}
)

func resourceArgoCDRepositoryCertificates() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages [custom TLS certificates](https://argo-cd.readthedocs.io/en/stable/user-guide/private-repositories/#self-signed-untrusted-tls-certificates) used by ArgoCD for connecting Git repositories.",
		CreateContext: resourceArgoCDRepositoryCertificatesCreate,
		ReadContext:   resourceArgoCDRepositoryCertificatesRead,
		DeleteContext: resourceArgoCDRepositoryCertificatesDelete,
		Schema:        certificatesSchema(),
	}
}

func resourceArgoCDRepositoryCertificatesCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	si := meta.(*ServerInterface)
	if err := si.initClients(ctx); err != nil {
		return errorToDiagnostics("failed to init clients", err)
	}

	// Not doing a RLock here because we can have a race-condition between the ListCertificates & CreateCertificate
	tokenMutexConfiguration.Lock()

	repoCertificate := expandRepositoryCertificate(d)

	if repoCertificate.CertType == "https" {
		rcl, err := si.CertificateClient.ListCertificates(ctx, &certificate.RepositoryCertificateQuery{
			HostNamePattern: repoCertificate.ServerName,
			CertType:        repoCertificate.CertType,
			CertSubType:     repoCertificate.CertSubType,
		})
		if err != nil {
			tokenMutexConfiguration.Unlock()

			return errorToDiagnostics(fmt.Sprintf("failed to list existing repository certificates when creating certificate for %s", repoCertificate.ServerName), err)
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

	rc, err := si.CertificateClient.CreateCertificate(
		ctx,
		&certificate.RepositoryCertificateCreateRequest{
			Certificates: &certs,
			Upsert:       false,
		},
	)
	tokenMutexConfiguration.Unlock()

	if err != nil {
		return argoCDAPIError("create", "repository certificate", repoCertificate.ServerName, err)
	}

	// TODO: upstream bug : if https certificate already exists, the response will be empty
	// instead of erroring about missing upsert flag but since the call is success
	// we assume everything went fine and get the id from the request
	var resourceId string
	if len(rc.Items) > 0 {
		resourceId, err = getId(&rc.Items[0])
	} else {
		resourceId, err = getId(repoCertificate)
	}

	if err != nil {
		return errorToDiagnostics(fmt.Sprintf("certificate for repository %s created but could not be handled", repoCertificate.ServerName), err)
	}

	d.SetId(resourceId)

	return resourceArgoCDRepositoryCertificatesRead(context.WithValue(ctx, certDataKey{}, repoCertificate.CertData), d, meta)
}

// Compute resource's id as :
// for ssh -> certType/certSubType/serverName
// for https -> certType/serverName
func getId(rc *application.RepositoryCertificate) (string, error) {
	if rc.CertType == "ssh" {
		if rc.CertSubType == "" || rc.ServerName == "" {
			return "", fmt.Errorf("invalid certificate: %s %s %s", rc.CertType, rc.CertSubType, rc.ServerName)
		}

		return fmt.Sprintf("%s/%s/%s", rc.CertType, rc.CertSubType, rc.ServerName), nil
	}

	if rc.ServerName == "" {
		return "", fmt.Errorf("invalid certificate: %s %s", rc.CertType, rc.ServerName)
	}

	return fmt.Sprintf("%s/%s", rc.CertType, rc.ServerName), nil
}

// Get serverName/certType/certSubType from resource's id
func fromId(id string) (string, string, string, error) {
	parts := strings.Split(id, "/")
	if len(parts) < 2 {
		return "", "", "", fmt.Errorf("unknown certificate %s in state", id)
	}

	certType := parts[0]
	if certType == "ssh" {
		if len(parts) < 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
			return "", "", "", fmt.Errorf("unknown certificate %s in state", id)
		}

		return parts[0], parts[1], parts[2], nil
	}

	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", "", "", fmt.Errorf("unknown certificate %s in state", id)
	}

	return parts[0], parts[1], "", nil
}

func resourceArgoCDRepositoryCertificatesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	si := meta.(*ServerInterface)
	if err := si.initClients(ctx); err != nil {
		return errorToDiagnostics("failed to init clients", err)
	}

	certType, certSubType, serverName, err := fromId(d.Id())
	if err != nil {
		return errorToDiagnostics("failed to parse certificate state", err)
	}

	tokenMutexConfiguration.RLock()
	rcl, err := si.CertificateClient.ListCertificates(ctx, &certificate.RepositoryCertificateQuery{
		HostNamePattern: serverName,
		CertType:        certType,
		CertSubType:     certSubType,
	})
	tokenMutexConfiguration.RUnlock()

	if err != nil {
		return argoCDAPIError("read", "repository certificate", serverName, err)
	} else if rcl == nil || len(rcl.Items) == 0 {
		// Certificate have already been deleted in an out-of-band fashion
		d.SetId("")
		return nil
	}

	repoCertificate := application.RepositoryCertificate{}

	for i, _rc := range rcl.Items {
		var resourceId string

		resourceId, err = getId(&_rc)
		if err != nil {
			return errorToDiagnostics(fmt.Sprintf("certificate for repository %s could not be handled", repoCertificate.ServerName), err)
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
		return errorToDiagnostics(fmt.Sprintf("failed to flatten repository certificate %s", d.Id()), err)
	}

	return nil
}

func resourceArgoCDRepositoryCertificatesDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	si := meta.(*ServerInterface)
	if err := si.initClients(ctx); err != nil {
		return errorToDiagnostics("failed to init clients", err)
	}

	certType, certSubType, serverName, err := fromId(d.Id())
	if err != nil {
		return errorToDiagnostics("failed to parse certificate state", err)
	}

	query := certificate.RepositoryCertificateQuery{
		HostNamePattern: serverName,
		CertType:        certType,
		CertSubType:     certSubType,
	}

	tokenMutexConfiguration.Lock()
	_, err = si.CertificateClient.DeleteCertificate(
		ctx,
		&query,
	)
	tokenMutexConfiguration.Unlock()

	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			// Certificate have already been deleted in an out-of-band fashion
			d.SetId("")
			return nil
		}

		return argoCDAPIError("delete", "repository certificate", serverName, err)
	}

	d.SetId("")

	return nil
}
