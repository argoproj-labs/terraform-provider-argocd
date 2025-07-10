package testhelpers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/k3s"
	"github.com/testcontainers/testcontainers-go/wait"
)

// K3sTestEnvironment represents a test environment with K3s and ArgoCD
type K3sTestEnvironment struct {
	K3sContainer *k3s.K3sContainer
	ArgoCDURL    string
}

// SetupK3sWithArgoCD sets up a K3s cluster with ArgoCD using testcontainers
func SetupK3sWithArgoCD(ctx context.Context, argoCDVersion, k8sVersion string) (*K3sTestEnvironment, error) {
	log.Println("Setting up K3s test environment...")
	k3sContainer, err := k3s.Run(ctx,
		fmt.Sprintf("rancher/k3s:%s-k3s1", k8sVersion),
		testcontainers.WithWaitStrategy(wait.ForLog("k3s is up and running")),
		testcontainers.WithExposedPorts("30124/tcp", "30123/tcp"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start K3s container: %w", err)
	}

	env := &K3sTestEnvironment{K3sContainer: k3sContainer}

	if err := env.installArgoCD(ctx, argoCDVersion); err != nil {
		env.Cleanup()
		return nil, fmt.Errorf("failed to install ArgoCD: %w", err)
	}

	log.Println("Waiting for ArgoCD to be ready...")
	if err := env.waitForArgoCD(ctx); err != nil {
		env.Cleanup()
		return nil, fmt.Errorf("failed to wait for ArgoCD: %w", err)
	}

	return env, nil
}

// installArgoCD installs ArgoCD in the K3s cluster using kustomize
func (env *K3sTestEnvironment) installArgoCD(ctx context.Context, version string) error {
	rootDir, err := env.projectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	kustomizeDir := filepath.Join(rootDir, "manifests", "overlays", version)
	log.Printf("Running 'kustomize build %s'\n", kustomizeDir)

	kustomizedManifests, err := env.runKustomizeBuild(kustomizeDir)
	if err != nil {
		return fmt.Errorf("failed to run kustomize build for version %s: %w", version, err)
	}

	log.Println("Applying manifests...")
	if err = env.applyManifestsToContainer(ctx, kustomizedManifests, "/tmp/argocd-kustomized.yaml"); err != nil {
		return fmt.Errorf("failed to copy kustomized manifests to container: %w", err)
	}

	testDataDir := filepath.Join(rootDir, "manifests/testdata")
	if _, err = os.Stat(testDataDir); os.IsNotExist(err) {
		return nil // No test data to install
	}

	if err = env.K3sContainer.CopyFileToContainer(ctx, testDataDir, "/tmp/testdata", 0644); err != nil {
		return fmt.Errorf("failed to copy testdata to container: %w", err)
	}

	if err = env.execInK3s(ctx, "kubectl", "apply", "-f", "/tmp/testdata"); err != nil {
		return err
	}

	return nil
}

func (env *K3sTestEnvironment) applyManifestsToContainer(ctx context.Context, manifests []byte, containerFilePath string) error {
	// Copy manifests to container
	if err := env.K3sContainer.CopyToContainer(ctx, manifests, containerFilePath, 0644); err != nil {
		return fmt.Errorf("failed to copy kustomized manifests to container: %w", err)
	}

	// Apply manifests
	if err := env.execInK3s(ctx, "kubectl", "apply", "-f", containerFilePath); err != nil {
		return err
	}

	return nil
}

// projectRoot gets the project root directory by checking `go env GOMOD`
func (env *K3sTestEnvironment) projectRoot() (string, error) {
	cmd := exec.Command("go", "env", "GOMOD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to find project root: %w", err)
	}

	return filepath.Dir(string(output)), nil
}

// runKustomizeBuild runs kustomize build on the temporary directory
func (env *K3sTestEnvironment) runKustomizeBuild(dir string) ([]byte, error) {
	cmd := exec.Command("kustomize", "build", dir)
	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("kustomize build failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to run kustomize: %w", err)
	}

	return output, nil
}

func (env *K3sTestEnvironment) execInK3s(ctx context.Context, args ...string) error {
	concat := strings.Join(args, " ")
	exitCode, reader, err := env.K3sContainer.Exec(ctx, args)
	if err != nil {
		return fmt.Errorf("failed to exec '%s': %w", concat, err)
	}
	output, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read kubectl output: %w", err)
	}
	if exitCode != 0 {
		return fmt.Errorf("'%s' failed with exit code %d: %s", concat, exitCode, string(output))
	}
	return nil
}

// waitForArgoCD waits for ArgoCD components to be ready
func (env *K3sTestEnvironment) waitForArgoCD(ctx context.Context) error {
	// Wait for CRDs to be established
	crds := []string{
		"applications.argoproj.io",
		"applicationsets.argoproj.io",
		"appprojects.argoproj.io",
	}

	for _, crd := range crds {
		if err := env.execInK3s(ctx, "kubectl", "wait", "--for=condition=Established", fmt.Sprintf("crd/%s", crd), "--timeout=60s"); err != nil {
			return err
		}
	}

	// Wait for deployments to be ready
	deployments := []string{"argocd-server", "argocd-repo-server", "argocd-redis"}

	timeout := "60s"
	for _, deployment := range deployments {
		if err := env.execInK3s(ctx, "kubectl", "wait", "--for=condition=available", fmt.Sprintf("deployment/%s", deployment), "-n", "argocd", "--timeout="+timeout); err != nil {
			return fmt.Errorf("failed to wait for deployment %s: %w", deployment, err)
		}
	}

	localPort, err := env.K3sContainer.MappedPort(ctx, "30123")
	if err != nil {
		return fmt.Errorf("failed to setup port forward: %w", err)
	}

	env.ArgoCDURL = fmt.Sprintf("127.0.0.1:%s", localPort.Port())

	return nil
}

// GetEnvironmentVariables returns the environment variables needed for tests
func (env *K3sTestEnvironment) GetEnvironmentVariables() map[string]string {
	return map[string]string{
		"ARGOCD_SERVER":        env.ArgoCDURL,
		"ARGOCD_AUTH_USERNAME": "admin",
		"ARGOCD_AUTH_PASSWORD": "acceptancetesting",
		"ARGOCD_INSECURE":      "true",
		"ARGOCD_VERSION":       "v3.0.0",
	}
}

// Cleanup cleans up the test environment
func (env *K3sTestEnvironment) Cleanup() {
	// Terminate container
	if env.K3sContainer != nil {
		if err := env.K3sContainer.Terminate(context.Background()); err != nil {
			fmt.Printf("Warning: failed to terminate container: %v\n", err)
		}
	}
}
