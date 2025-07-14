package testhelpers

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"
)

var (
	GlobalTestEnv *K3sTestEnvironment
	testEnvOnce   sync.Once
)

// TestMain is a helper function to be used in test files' TestMain functions
func TestMain(m *testing.M) {
	if os.Getenv("USE_TESTCONTAINERS") == "true" {
		SetupTestSuite(m)
	} else {
		os.Exit(m.Run())
	}
}

const (
	// DefaultTestTimeout is the default timeout for test setup
	DefaultTestTimeout = 15 * time.Minute
)

// SetupTestSuite sets up a shared test environment for all acceptance tests
func SetupTestSuite(m *testing.M) {
	code := runTestSuite(m)
	os.Exit(code)
}

func runTestSuite(m *testing.M) int {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTestTimeout)
	defer cancel()

	// Setup the test environment once
	var setupErr error

	testEnvOnce.Do(func() {
		argoCDVersion := os.Getenv("ARGOCD_VERSION")
		k3sVersion := os.Getenv("K3S_VERSION")

		GlobalTestEnv, setupErr = SetupK3sWithArgoCD(ctx, argoCDVersion, k3sVersion)
		if setupErr != nil {
			return
		}

		// Set environment variables for tests; currently only ARGOCD_SERVER is used (since we're port-forwarding the k8s
		// service) but can be extended with more env vars if needed
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
