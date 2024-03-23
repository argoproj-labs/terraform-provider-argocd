package provider

import (
	"bytes"
	"context"
	"fmt"
	"net/url"

	"github.com/argoproj/argo-cd/v2/cmd/argocd/commands/headless"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/session"
	"github.com/argoproj/argo-cd/v2/util/cache"
	"github.com/argoproj/argo-cd/v2/util/io"
	"github.com/argoproj/argo-cd/v2/util/localconfig"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/oboukili/terraform-provider-argocd/internal/diagnostics"
	apimachineryschema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

type ArgoCDProviderConfig struct {
	// Configuration for standard login using either with username/password or auth_token
	AuthToken types.String `tfsdk:"auth_token"`
	Username  types.String `tfsdk:"username"`
	Password  types.String `tfsdk:"password"`

	// When using standard login either server address or port forwarding must be used
	ServerAddr               types.String `tfsdk:"server_addr"`
	PortForward              types.Bool   `tfsdk:"port_forward"`
	PortForwardWithNamespace types.String `tfsdk:"port_forward_with_namespace"`
	Kubernetes               []Kubernetes `tfsdk:"kubernetes"`

	// Run ArgoCD API server locally
	Core types.Bool `tfsdk:"core"`

	// Login using credentials from local ArgoCD config file
	UseLocalConfig types.Bool   `tfsdk:"use_local_config"`
	ConfigPath     types.String `tfsdk:"config_path"`
	Context        types.String `tfsdk:"context"`

	// Other configuration
	CertFile        types.String `tfsdk:"cert_file"`
	ClientCertFile  types.String `tfsdk:"client_cert_file"`
	ClientCertKey   types.String `tfsdk:"client_cert_key"`
	GRPCWeb         types.Bool   `tfsdk:"grpc_web"`
	GRPCWebRootPath types.String `tfsdk:"grpc_web_root_path"`
	Headers         types.Set    `tfsdk:"headers"`
	Insecure        types.Bool   `tfsdk:"insecure"`
	PlainText       types.Bool   `tfsdk:"plain_text"`
	UserAgent       types.String `tfsdk:"user_agent"`
}

func (p ArgoCDProviderConfig) getApiClientOptions(ctx context.Context) (*apiclient.ClientOptions, diag.Diagnostics) {
	var diags diag.Diagnostics

	opts := &apiclient.ClientOptions{
		AuthToken:            getDefaultString(p.AuthToken, "ARGOCD_AUTH_TOKEN"),
		CertFile:             p.CertFile.ValueString(),
		ClientCertFile:       p.ClientCertFile.ValueString(),
		ClientCertKeyFile:    p.ClientCertKey.ValueString(),
		GRPCWeb:              p.GRPCWeb.ValueBool(),
		GRPCWebRootPath:      p.GRPCWebRootPath.ValueString(),
		Insecure:             getDefaultBool(ctx, p.Insecure, "ARGOCD_INSECURE"),
		PlainText:            p.PlainText.ValueBool(),
		PortForward:          p.PortForward.ValueBool(),
		PortForwardNamespace: p.PortForwardWithNamespace.ValueString(),
		ServerAddr:           getDefaultString(p.ServerAddr, "ARGOCD_SERVER"),
		UserAgent:            p.Username.ValueString(),
	}

	if !p.Headers.IsNull() {
		var h []string

		diags.Append(p.Headers.ElementsAs(ctx, &h, false)...)

		opts.Headers = h
	}

	coreEnabled, d := p.setCoreOpts(opts)

	diags.Append(d...)

	localConfigEnabled, d := p.setLocalConfigOpts(opts)

	diags.Append(d...)

	portForwardingEnabled, d := p.setPortForwardingOpts(ctx, opts)

	diags.Append(d...)

	username := getDefaultString(p.Username, "ARGOCD_AUTH_USERNAME")
	password := getDefaultString(p.Password, "ARGOCD_AUTH_PASSWORD")

	usernameAndPasswordSet := username != "" && password != ""

	switch {
	// Provider configuration errors
	case !coreEnabled && !portForwardingEnabled && !localConfigEnabled && opts.ServerAddr == "":
		diags.Append(diagnostics.Error("invalid provider configuration: one of `core,port_forward,port_forward_with_namespace,use_local_config,server_addr` must be specified", nil)...)
	case portForwardingEnabled && opts.AuthToken == "" && !usernameAndPasswordSet:
		diags.Append(diagnostics.Error("invalid provider configuration: either `username/password` or `auth_token` must be specified when port forwarding is enabled", nil)...)
	case opts.ServerAddr != "" && !coreEnabled && opts.AuthToken == "" && !usernameAndPasswordSet:
		diags.Append(diagnostics.Error("invalid provider configuration: either `username/password` or `auth_token` must be specified if `server_addr` is specified", nil)...)
	}

	if diags.HasError() {
		return nil, diags
	}

	switch {
	// Handle "special" configuration use-cases
	case coreEnabled:
		// HACK: `headless.StartLocalServer` manipulates this global variable
		// when starting the local server without checking it's length/contents
		// which leads to a panic if called multiple times. So, we need to
		// ensure we "reset" it before calling the method.
		if runtimeErrorHandlers == nil {
			runtimeErrorHandlers = runtime.ErrorHandlers
		} else {
			runtime.ErrorHandlers = runtimeErrorHandlers
		}

		err := headless.MaybeStartLocalServer(ctx, opts, "", nil, nil, cache.RedisCompressionNone, nil)
		if err != nil {
			diags.Append(diagnostics.Error("failed to start local server", err)...)
			return nil, diags
		}
	case opts.ServerAddr != "" && opts.AuthToken == "" && usernameAndPasswordSet:
		apiClient, err := apiclient.NewClient(opts)
		if err != nil {
			diags.Append(diagnostics.Error("failed to create new API client", err)...)
			return nil, diags
		}

		closer, sc, err := apiClient.NewSessionClient()
		if err != nil {
			diags.Append(diagnostics.Error("failed to create new session client", err)...)
			return nil, diags
		}

		defer io.Close(closer)

		sessionOpts := session.SessionCreateRequest{
			Username: username,
			Password: password,
		}

		resp, err := sc.Create(ctx, &sessionOpts)
		if err != nil {
			diags.Append(diagnostics.Error("failed to create new session", err)...)
			return nil, diags
		}

		opts.AuthToken = resp.Token
	}

	return opts, diags
}

func (p ArgoCDProviderConfig) setCoreOpts(opts *apiclient.ClientOptions) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	coreEnabled := p.Core.ValueBool()
	if coreEnabled {
		if opts.ServerAddr != "" {
			diags.AddWarning("`server_addr` is ignored by the provider and overwritten when `core = true`.", "")
		}

		opts.ServerAddr = "kubernetes"
		opts.Core = true

		if !p.Username.IsNull() {
			diags.AddWarning("`username` is ignored when `core = true`.", "")
		}
	}

	return coreEnabled, diags
}

func (p ArgoCDProviderConfig) setLocalConfigOpts(opts *apiclient.ClientOptions) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	useLocalConfig := p.UseLocalConfig.ValueBool()
	switch useLocalConfig {
	case true:
		if opts.ServerAddr != "" {
			diags.AddWarning("setting `server_addr` alongside `use_local_config = true` is unnecessary and not recommended as this will overwrite the address retrieved from the local ArgoCD context.", "")
		}

		if !p.Username.IsNull() {
			diags.AddWarning("`username` is ignored when `use_local_config = true`.", "")
		}

		opts.Context = getDefaultString(p.Context, "ARGOCD_CONTEXT")

		cp := getDefaultString(p.ConfigPath, "ARGOCD_CONFIG_PATH")

		if cp != "" {
			opts.ConfigPath = p.ConfigPath.ValueString()
			break
		}

		cp, err := localconfig.DefaultLocalConfigPath()
		if err == nil {
			opts.ConfigPath = cp
			break
		}

		diags.Append(diagnostics.Error("failed to find default ArgoCD config path", err)...)
	case false:
		// Log warnings if explicit configuration has been provided for local config when `use_local_config` is not enabled.
		if !p.ConfigPath.IsNull() {
			diags.AddWarning("`config_path` is ignored by provider unless `use_local_config = true`.", "")
		}

		if !p.Context.IsNull() {
			diags.AddWarning("`context` is ignored by provider unless `use_local_config = true`.", "")
		}
	}

	return useLocalConfig, diags
}

func (p ArgoCDProviderConfig) setPortForwardingOpts(ctx context.Context, opts *apiclient.ClientOptions) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	portForwardingEnabled := opts.PortForward || opts.PortForwardNamespace != ""
	switch portForwardingEnabled {
	case true:
		if opts.ServerAddr != "" {
			diags.AddWarning("`server_addr` is ignored by the provider and overwritten when port forwarding is enabled.", "")
		}

		opts.ServerAddr = "localhost" // will be overwritten by ArgoCD module when we initialize the API client but needs to be set here to ensure we

		if p.Kubernetes == nil {
			break
		}

		k := p.Kubernetes[0]
		opts.KubeOverrides = &clientcmd.ConfigOverrides{
			AuthInfo: api.AuthInfo{
				ClientCertificateData: bytes.NewBufferString(getDefaultString(k.ClientCertificate, "KUBE_CLIENT_CERT_DATA")).Bytes(),
				Username:              getDefaultString(k.Username, "KUBE_USER"),
				Password:              getDefaultString(k.Password, "KUBE_PASSWORD"),
				ClientKeyData:         bytes.NewBufferString(getDefaultString(k.ClientKey, "KUBE_CLIENT_KEY_DATA")).Bytes(),
				Token:                 getDefaultString(k.Token, "KUBE_TOKEN"),
			},
			ClusterInfo: api.Cluster{
				InsecureSkipTLSVerify:    getDefaultBool(ctx, k.Insecure, "KUBE_INSECURE"),
				CertificateAuthorityData: bytes.NewBufferString(getDefaultString(k.ClusterCACertificate, "KUBE_CLUSTER_CA_CERT_DATA")).Bytes(),
			},
			CurrentContext: getDefaultString(k.ConfigContext, "KUBE_CTX"),
			Context: api.Context{
				AuthInfo: getDefaultString(k.ConfigContextAuthInfo, "KUBE_CTX_AUTH_INFO"),
				Cluster:  getDefaultString(k.ConfigContextCluster, "KUBE_CTX_CLUSTER"),
			},
		}

		h := getDefaultString(k.Host, "KUBE_HOST")
		if h != "" {
			// Server has to be the complete address of the Kubernetes cluster (scheme://hostname:port), not just the hostname,
			// because `overrides` are processed too late to be taken into account by `defaultServerUrlFor()`.
			// This basically replicates what defaultServerUrlFor() does with config but for overrides,
			// see https://github.com/Kubernetes/client-go/blob/v12.0.0/rest/url_utils.go#L85-L87
			hasCA := len(opts.KubeOverrides.ClusterInfo.CertificateAuthorityData) != 0
			hasCert := len(opts.KubeOverrides.AuthInfo.ClientCertificateData) != 0
			defaultTLS := hasCA || hasCert || opts.KubeOverrides.ClusterInfo.InsecureSkipTLSVerify

			var host *url.URL

			host, _, err := rest.DefaultServerURL(h, "", apimachineryschema.GroupVersion{}, defaultTLS)
			if err == nil {
				opts.KubeOverrides.ClusterInfo.Server = host.String()
			} else {
				diags.Append(diagnostics.Error(fmt.Sprintf("failed to extract default server URL for host %s", h), err)...)
			}
		}

		if k.Exec == nil {
			break
		}

		e := k.Exec[0]
		exec := &api.ExecConfig{
			InteractiveMode: api.IfAvailableExecInteractiveMode,
			APIVersion:      e.APIVersion.ValueString(),
			Command:         e.Command.ValueString(),
		}

		var a []string

		diags.Append(e.Args.ElementsAs(ctx, &a, false)...)
		exec.Args = a

		var env map[string]string

		diags.Append(e.Env.ElementsAs(ctx, &env, false)...)

		for k, v := range env {
			exec.Env = append(exec.Env, api.ExecEnvVar{Name: k, Value: v})
		}

		opts.KubeOverrides.AuthInfo.Exec = exec
	case false:
		if p.Kubernetes != nil {
			diags.AddWarning("`Kubernetes` configuration block is ignored by provider unless `port_forward` or `port_forward_with_namespace` are configured.", "")
		}
	}

	return portForwardingEnabled, diags
}

type Kubernetes struct {
	Host                  types.String     `tfsdk:"host"`
	Username              types.String     `tfsdk:"username"`
	Password              types.String     `tfsdk:"password"`
	Insecure              types.Bool       `tfsdk:"insecure"`
	ClientCertificate     types.String     `tfsdk:"client_certificate"`
	ClientKey             types.String     `tfsdk:"client_key"`
	ClusterCACertificate  types.String     `tfsdk:"cluster_ca_certificate"`
	ConfigContext         types.String     `tfsdk:"config_context"`
	ConfigContextAuthInfo types.String     `tfsdk:"config_context_auth_info"`
	ConfigContextCluster  types.String     `tfsdk:"config_context_cluster"`
	Token                 types.String     `tfsdk:"token"`
	Exec                  []KubernetesExec `tfsdk:"exec"`
}

type KubernetesExec struct {
	APIVersion types.String `tfsdk:"api_version"`
	Command    types.String `tfsdk:"command"`
	Env        types.Map    `tfsdk:"env"`
	Args       types.List   `tfsdk:"args"`
}
