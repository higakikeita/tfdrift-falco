//go:build e2e
// +build e2e

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	falcoHTTPEndpoint  = "http://localhost:8081"
	tfdriftAPIEndpoint = "http://localhost:8080"
	testTimeout        = 30 * time.Second
	healthCheckTimeout = 60 * time.Second
)

// TestLocalE2E runs the full local E2E test suite
func TestLocalE2E(t *testing.T) {
	// Skip if not running E2E tests
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Get the local directory path
	localDir := getLocalTestDir(t)

	// Start docker-compose
	t.Log("Starting docker-compose services...")
	startCmd := exec.CommandContext(context.Background(), "docker-compose", "-f",
		filepath.Join(localDir, "docker-compose.yml"), "up", "-d")
	startCmd.Dir = localDir
	if output, err := startCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to start docker-compose: %v\n%s", err, string(output))
	}

	// Cleanup: stop services after test
	t.Cleanup(func() {
		t.Log("Stopping docker-compose services...")
		downCmd := exec.CommandContext(context.Background(), "docker-compose", "-f",
			filepath.Join(localDir, "docker-compose.yml"), "down")
		downCmd.Dir = localDir
		if output, err := downCmd.CombinedOutput(); err != nil {
			t.Logf("Warning: Failed to stop docker-compose: %v\n%s", err, string(output))
		}
	})

	// Wait for services to be healthy
	t.Log("Waiting for Falco service to be healthy...")
	require.NoError(t, waitForHealthy(falcoHTTPEndpoint, healthCheckTimeout))

	t.Log("Waiting for TFDrift API service to be healthy...")
	require.NoError(t, waitForHealthy(tfdriftAPIEndpoint, healthCheckTimeout))

	// Run test scenarios
	t.Run("TriggerEC2DriftEvent", func(t *testing.T) {
		testEC2DriftDetection(t)
	})

	t.Run("TriggerSecurityGroupDriftEvent", func(t *testing.T) {
		testSecurityGroupDriftDetection(t)
	})

	t.Run("TriggerS3DriftEvent", func(t *testing.T) {
		testS3DriftDetection(t)
	})

	t.Run("APIHealthCheck", func(t *testing.T) {
		testAPIHealthCheck(t)
	})
}

// testEC2DriftDetection triggers an EC2 change and verifies drift detection
func testEC2DriftDetection(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	// Trigger EC2 event
	t.Log("Triggering EC2 instance modification event...")
	resp, err := http.Post(
		falcoHTTPEndpoint+"/trigger-ec2-change",
		"application/json",
		nil,
	)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Give the system time to process the event
	time.Sleep(2 * time.Second)

	// Poll TFDrift API for drift detection
	t.Log("Checking for drift detection...")
	driftDetected := pollForDrift(ctx, tfdriftAPIEndpoint, "aws_instance")
	assert.True(t, driftDetected, "Expected drift to be detected for aws_instance")
}

// testSecurityGroupDriftDetection triggers a security group change
func testSecurityGroupDriftDetection(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	// Trigger security group event
	t.Log("Triggering security group modification event...")
	resp, err := http.Post(
		falcoHTTPEndpoint+"/trigger-sg-change",
		"application/json",
		nil,
	)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Give the system time to process the event
	time.Sleep(2 * time.Second)

	// Poll TFDrift API for drift detection
	t.Log("Checking for drift detection...")
	driftDetected := pollForDrift(ctx, tfdriftAPIEndpoint, "aws_security_group")
	assert.True(t, driftDetected, "Expected drift to be detected for aws_security_group")
}

// testS3DriftDetection triggers an S3 bucket change
func testS3DriftDetection(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	// Trigger S3 event
	t.Log("Triggering S3 bucket modification event...")
	resp, err := http.Post(
		falcoHTTPEndpoint+"/trigger-s3-change",
		"application/json",
		nil,
	)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Give the system time to process the event
	time.Sleep(2 * time.Second)

	// Poll TFDrift API for drift detection
	t.Log("Checking for drift detection...")
	driftDetected := pollForDrift(ctx, tfdriftAPIEndpoint, "aws_s3_bucket")
	assert.True(t, driftDetected, "Expected drift to be detected for aws_s3_bucket")
}

// testAPIHealthCheck verifies the TFDrift API is responsive
func testAPIHealthCheck(t *testing.T) {
	resp, err := http.Get(tfdriftAPIEndpoint + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, "healthy", result["status"])
}

// waitForHealthy polls an endpoint until it returns healthy or timeout
func waitForHealthy(endpoint string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		resp, err := http.Get(endpoint + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}

		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("service at %s did not become healthy within %v", endpoint, timeout)
}

// pollForDrift checks the TFDrift API for detected drift
func pollForDrift(ctx context.Context, apiEndpoint, resourceType string) bool {
	pollDeadline := time.Now().Add(10 * time.Second)

	for time.Now().Before(pollDeadline) {
		select {
		case <-ctx.Done():
			return false
		default:
		}

		// Try to fetch drift from API endpoint
		// This assumes the API has a drift status or events endpoint
		resp, err := http.Get(apiEndpoint + "/api/v1/drift")
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		if bytes.Contains(body, []byte(resourceType)) {
			return true
		}

		time.Sleep(500 * time.Millisecond)
	}

	return false
}

// getLocalTestDir returns the absolute path to the tests/e2e/local directory
func getLocalTestDir(t *testing.T) string {
	// Try to find the local test directory from current working directory
	// This handles both direct execution and go test execution
	paths := []string{
		".", // Current directory
		"tests/e2e/local",
		"./tests/e2e/local",
		"../../../tests/e2e/local",
	}

	for _, p := range paths {
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			abs, err := filepath.Abs(p)
			if err == nil {
				// Check if docker-compose.yml exists
				if _, err := os.Stat(filepath.Join(abs, "docker-compose.yml")); err == nil {
					return abs
				}
			}
		}
	}

	// If not found, try to get it from the test file location
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(file)
}
