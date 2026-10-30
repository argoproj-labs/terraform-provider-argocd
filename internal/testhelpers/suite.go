package testhelpers

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"
)

var (
	GlobalTestEnv *K3sTestEnvironment
	testEnvOnce   sync.Once
)

// TestMain is a helper function to be used in test files' TestMain functions. The test fixture (K3s with Argo CD) is only
// set up for acceptance tests (TF_ACC=1); otherwise only unit tests are run, which don't need a cluster.
func TestMain(m *testing.M) {
	if os.Getenv("TF_ACC") == "" {
		os.Exit(m.Run())
	}

	os.Exit(runTestSuite(m))
}

const (
	// DefaultTestTimeout is the default timeout for test setup
	DefaultTestTimeout = 15 * time.Minute
)

func runTestSuite(m *testing.M) int {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTestTimeout)
	defer cancel()

	// Setup the test environment once
	var setupErr error

	testEnvOnce.Do(func() {
		// The versions are defined in the GNUmakefile (and overridden by CI), so there is deliberately no default here
		argoCDVersion := os.Getenv("ARGOCD_VERSION")
		k3sVersion := os.Getenv("K3S_VERSION")

		if argoCDVersion == "" || k3sVersion == "" {
			setupErr = errors.New("ARGOCD_VERSION and K3S_VERSION must be set; run the tests via 'make testacc'")

			return
		}

		GlobalTestEnv, setupErr = SetupK3sWithArgoCD(ctx, argoCDVersion, k3sVersion)
		if setupErr != nil {
			return
		}

		// Set the environment variables the provider and tests use to connect to the freshly created Argo CD installation
		envVars := GlobalTestEnv.GetEnvironmentVariables()
		for key, value := range envVars {
			os.Setenv(key, value)
		}
	})

	if setupErr != nil {
		panic("Failed to setup test environment: " + setupErr.Error())
	}

	// Run tests
	code := m.Run()

	// Cleanup
	if GlobalTestEnv != nil {
		GlobalTestEnv.Cleanup(ctx)
	}

	return code
}
