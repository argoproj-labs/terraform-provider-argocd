package testhelpers

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"
)

var (
	// Global test environment - shared across all tests
	globalTestEnv *K3sTestEnvironment
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

// SetupTestSuite sets up a shared test environment for all acceptance tests
func SetupTestSuite(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	// Setup the test environment once
	var setupErr error
	testEnvOnce.Do(func() {
		argoCDVersion := os.Getenv("ARGOCD_VERSION")
		if argoCDVersion == "" {
			argoCDVersion = "v3.0.0"
		}

		k8sVersion := "v1.31.6"

		globalTestEnv, setupErr = SetupK3sWithArgoCD(ctx, argoCDVersion, k8sVersion)
		if setupErr != nil {
			return
		}

		// Set environment variables for tests
		envVars := globalTestEnv.GetEnvironmentVariables()
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
	if globalTestEnv != nil {
		globalTestEnv.Cleanup()
	}

	os.Exit(code)
}
