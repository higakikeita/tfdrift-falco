// Package e2e provides end-to-end testing utilities for TFDrift-Falco.
package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	// TODO: Migrate to aws-sdk-go-v2 (aws-sdk-go-v1 deprecated, EOL July 31, 2025)
	// See: https://aws.amazon.com/blogs/developer/announcing-end-of-support-for-aws-sdk-for-go-v1-on-july-31-2025/
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/keitahigaki/tfdrift-falco/pkg/detector"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	"github.com/stretchr/testify/require"
)

// TestContext holds test resources and clients
type TestContext struct {
	t              *testing.T
	cfg            *config.Config
	detector       *detector.Detector
	awsSession     *session.Session
	ec2Client      *ec2.EC2
	iamClient      *iam.IAM
	s3Client       *s3.S3
	alertChan      chan types.DriftAlert
	cleanupFuncs   []func()
	testInstanceID string
	testRoleName   string
	testBucketName string
	testPolicyName string
}

// NewTestContext creates a new test context
func NewTestContext(t *testing.T) *TestContext {
	t.Helper()

	// Load test outputs from Terraform
	outputs := loadTerraformOutputs(t)

	// Create AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(getEnvOrDefault("AWS_REGION", "us-east-1")),
	})
	require.NoError(t, err, "Failed to create AWS session")

	ctx := &TestContext{
		t:              t,
		awsSession:     sess,
		ec2Client:      ec2.New(sess),
		iamClient:      iam.New(sess),
		s3Client:       s3.New(sess),
		alertChan:      make(chan types.DriftAlert, 100),
		cleanupFuncs:   []func(){},
		testInstanceID: outputs["test_instance_id"].(string),
		testRoleName:   outputs["test_role_name"].(string),
		testBucketName: outputs["test_bucket_name"].(string),
		testPolicyName: outputs["test_policy_name"].(string),
	}

	return ctx
}

// StartDetector starts the TFDrift detector
func (ctx *TestContext) StartDetector() {
	ctx.t.Helper()

	// Load configuration
	cfg := ctx.loadConfig()
	ctx.cfg = cfg

	// Create detector
	det, err := detector.New(cfg)
	require.NoError(ctx.t, err, "Failed to create detector")
	ctx.detector = det

	// Start detector in background
	detectorCtx, cancel := context.WithCancel(context.Background())
	ctx.AddCleanup(cancel)

	go func() {
		if err := det.Start(detectorCtx); err != nil && err != context.Canceled {
			ctx.t.Logf("Detector error: %v", err)
		}
	}()

	// Wait for detector to be ready
	ctx.WaitForReady(10 * time.Second)
}

// WaitForReady waits for the detector to be ready
func (ctx *TestContext) WaitForReady(timeout time.Duration) {
	ctx.t.Helper()
	ctx.t.Logf("Waiting for detector to be ready...")
	time.Sleep(timeout)
	ctx.t.Logf("Detector ready")
}

// WaitForAlert waits for a drift alert with timeout
func (ctx *TestContext) WaitForAlert(timeout time.Duration) *types.DriftAlert {
	ctx.t.Helper()

	ctx.t.Logf("Waiting for alert (timeout: %v)...", timeout)

	select {
	case alert := <-ctx.alertChan:
		ctx.t.Logf("Received alert: %+v", alert)
		return &alert
	case <-time.After(timeout):
		ctx.t.Logf("Alert timeout after %v", timeout)
		return nil
	}
}

// ModifyEC2Termination modifies EC2 termination protection
func (ctx *TestContext) ModifyEC2Termination(instanceID string, enabled bool) {
	ctx.t.Helper()

	ctx.t.Logf("Modifying EC2 termination protection: %s -> %v", instanceID, enabled)

	input := &ec2.ModifyInstanceAttributeInput{
		InstanceId: aws.String(instanceID),
		DisableApiTermination: &ec2.AttributeBooleanValue{
			Value: aws.Bool(enabled),
		},
	}

	_, err := ctx.ec2Client.ModifyInstanceAttribute(input)
	require.NoError(ctx.t, err, "Failed to modify instance attribute")

	ctx.t.Logf("Modified EC2 termination protection successfully")
}

// ModifyIAMAssumeRolePolicy modifies IAM role assume role policy
func (ctx *TestContext) ModifyIAMAssumeRolePolicy(roleName string, newPolicy string) {
	ctx.t.Helper()

	ctx.t.Logf("Modifying IAM assume role policy: %s", roleName)

	input := &iam.UpdateAssumeRolePolicyInput{
		RoleName:       aws.String(roleName),
		PolicyDocument: aws.String(newPolicy),
	}

	_, err := ctx.iamClient.UpdateAssumeRolePolicy(input)
	require.NoError(ctx.t, err, "Failed to update assume role policy")

	ctx.t.Logf("Modified IAM assume role policy successfully")
}

// DisableS3BucketEncryption disables S3 bucket encryption
func (ctx *TestContext) DisableS3BucketEncryption(bucketName string) {
	ctx.t.Helper()

	ctx.t.Logf("Disabling S3 bucket encryption: %s", bucketName)

	input := &s3.DeleteBucketEncryptionInput{
		Bucket: aws.String(bucketName),
	}

	_, err := ctx.s3Client.DeleteBucketEncryption(input)
	require.NoError(ctx.t, err, "Failed to delete bucket encryption")

	ctx.t.Logf("Disabled S3 bucket encryption successfully")
}

// EnableS3BucketEncryption enables S3 bucket encryption
func (ctx *TestContext) EnableS3BucketEncryption(bucketName string) {
	ctx.t.Helper()

	ctx.t.Logf("Enabling S3 bucket encryption: %s", bucketName)

	input := &s3.PutBucketEncryptionInput{
		Bucket: aws.String(bucketName),
		ServerSideEncryptionConfiguration: &s3.ServerSideEncryptionConfiguration{
			Rules: []*s3.ServerSideEncryptionRule{
				{
					ApplyServerSideEncryptionByDefault: &s3.ServerSideEncryptionByDefault{
						SSEAlgorithm: aws.String("AES256"),
					},
				},
			},
		},
	}

	_, err := ctx.s3Client.PutBucketEncryption(input)
	require.NoError(ctx.t, err, "Failed to enable bucket encryption")

	ctx.t.Logf("Enabled S3 bucket encryption successfully")
}

// AddCleanup adds a cleanup function to be called at test end
func (ctx *TestContext) AddCleanup(fn func()) {
	ctx.cleanupFuncs = append(ctx.cleanupFuncs, fn)
}

// Cleanup runs all cleanup functions
func (ctx *TestContext) Cleanup() {
	ctx.t.Helper()
	ctx.t.Logf("Running cleanup...")

	for i := len(ctx.cleanupFuncs) - 1; i >= 0; i-- {
		ctx.cleanupFuncs[i]()
	}

	ctx.t.Logf("Cleanup complete")
}

// loadConfig loads test configuration
func (ctx *TestContext) loadConfig() *config.Config {
	ctx.t.Helper()

	configPath := getEnvOrDefault("E2E_CONFIG_PATH", "fixtures/config.yaml")

	cfg, err := config.Load(configPath)
	require.NoError(ctx.t, err, "Failed to load config")

	// Override with test values
	cfg.DryRun = false

	return cfg
}

// loadTerraformOutputs loads Terraform outputs
func loadTerraformOutputs(t *testing.T) map[string]interface{} {
	t.Helper()

	outputFile := getEnvOrDefault("E2E_TF_OUTPUTS", "fixtures/terraform_outputs.json")

	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Skipf("Skipping E2E test: terraform outputs not found at %s", outputFile)
		return nil
	}

	var outputs map[string]interface{}
	err = json.Unmarshal(data, &outputs)
	require.NoError(t, err, "Failed to parse terraform outputs")

	return outputs
}

// getEnvOrDefault gets environment variable or returns default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// skipIfShort skips test if running in short mode
// Note: Used in E2E tests which are excluded from normal builds with //go:build ignore
func skipIfShort(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}
}

// skipIfNoAWS skips test if AWS credentials are not available
// Note: Used in E2E tests which are excluded from normal builds with //go:build ignore
func skipIfNoAWS(t *testing.T) {
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" {
		t.Skip("Skipping E2E test: AWS credentials not found")
	}
}

// skipIfNoFalco skips test if Falco is not available
// Note: Used in E2E tests which are excluded from normal builds with //go:build ignore
func skipIfNoFalco(t *testing.T) {
	falcoHost := getEnvOrDefault("FALCO_HOSTNAME", "localhost")
	falcoPort := getEnvOrDefault("FALCO_PORT", "5060")

	// Simple check - could be enhanced with actual gRPC probe
	if falcoHost == "" {
		t.Skip("Skipping E2E test: Falco not configured")
	}

	t.Logf("Using Falco at %s:%s", falcoHost, falcoPort)
}

// AssertAlertReceived asserts that an alert was received
func AssertAlertReceived(t *testing.T, alert *types.DriftAlert, resourceType, resourceID string) {
	t.Helper()

	require.NotNil(t, alert, "Expected alert but got nil")
	require.Equal(t, resourceType, alert.ResourceType, "Resource type mismatch")
	require.Equal(t, resourceID, alert.ResourceID, "Resource ID mismatch")
	require.NotEmpty(t, alert.Attribute, "Expected attribute but got empty")
	require.NotEmpty(t, alert.UserIdentity.UserName, "Expected user identity")
}

// CreateTestIAMPolicy returns a test IAM policy document
func CreateTestIAMPolicy(bucketARN string) string {
	return fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": [
					"s3:GetObject",
					"s3:PutObject"
				],
				"Resource": "%s/*"
			}
		]
	}`, bucketARN)
}

// CreateTestAssumeRolePolicy returns a test assume role policy
func CreateTestAssumeRolePolicy(service string) string {
	return fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": {
					"Service": "%s"
				},
				"Action": "sts:AssumeRole"
			}
		]
	}`, service)
}
