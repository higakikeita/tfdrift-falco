package e2e

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestDriftDetection_EC2Termination tests EC2 termination protection drift
func TestDriftDetection_EC2Termination(t *testing.T) {
	skipIfShort(t)
	skipIfNoAWS(t)
	skipIfNoFalco(t)

	ctx := NewTestContext(t)
	defer ctx.Cleanup()

	t.Log("=== E2E Test: EC2 Termination Protection Drift ===")

	// Start detector
	ctx.StartDetector()

	// Modify EC2 termination protection (disable it)
	originalValue := true
	newValue := false
	ctx.ModifyEC2Termination(ctx.testInstanceID, newValue)
	ctx.AddCleanup(func() {
		// Restore original value
		ctx.ModifyEC2Termination(ctx.testInstanceID, originalValue)
	})

	// Wait for alert (CloudTrail can take 5-15 minutes)
	// For E2E tests, we use a longer timeout
	alert := ctx.WaitForAlert(20 * time.Minute)

	// Verify alert
	AssertAlertReceived(t, alert, "aws_instance", ctx.testInstanceID)
	assert.Contains(t, alert.Changes, "disable_api_termination", "Expected termination protection change")

	t.Log("✅ EC2 termination protection drift detected successfully")
}

// TestDriftDetection_IAMAssumeRolePolicy tests IAM assume role policy drift
func TestDriftDetection_IAMAssumeRolePolicy(t *testing.T) {
	skipIfShort(t)
	skipIfNoAWS(t)
	skipIfNoFalco(t)

	ctx := NewTestContext(t)
	defer ctx.Cleanup()

	t.Log("=== E2E Test: IAM Assume Role Policy Drift ===")

	// Start detector
	ctx.StartDetector()

	// Modify IAM assume role policy
	originalPolicy := CreateTestAssumeRolePolicy("ec2.amazonaws.com")
	newPolicy := CreateTestAssumeRolePolicy("lambda.amazonaws.com")

	ctx.ModifyIAMAssumeRolePolicy(ctx.testRoleName, newPolicy)
	ctx.AddCleanup(func() {
		// Restore original policy
		ctx.ModifyIAMAssumeRolePolicy(ctx.testRoleName, originalPolicy)
	})

	// Wait for alert
	alert := ctx.WaitForAlert(20 * time.Minute)

	// Verify alert
	AssertAlertReceived(t, alert, "aws_iam_role", ctx.testRoleName)
	assert.Contains(t, alert.Changes, "assume_role_policy", "Expected assume role policy change")
	assert.Equal(t, "critical", alert.Severity, "IAM changes should be critical")

	t.Log("✅ IAM assume role policy drift detected successfully")
}

// TestDriftDetection_S3BucketEncryption tests S3 bucket encryption drift
func TestDriftDetection_S3BucketEncryption(t *testing.T) {
	skipIfShort(t)
	skipIfNoAWS(t)
	skipIfNoFalco(t)

	ctx := NewTestContext(t)
	defer ctx.Cleanup()

	t.Log("=== E2E Test: S3 Bucket Encryption Drift ===")

	// Start detector
	ctx.StartDetector()

	// Disable S3 bucket encryption
	ctx.DisableS3BucketEncryption(ctx.testBucketName)
	ctx.AddCleanup(func() {
		// Re-enable encryption
		ctx.EnableS3BucketEncryption(ctx.testBucketName)
	})

	// Wait for alert
	alert := ctx.WaitForAlert(20 * time.Minute)

	// Verify alert
	AssertAlertReceived(t, alert, "aws_s3_bucket", ctx.testBucketName)
	assert.Contains(t, alert.Changes, "server_side_encryption_configuration", "Expected encryption change")
	assert.Equal(t, "critical", alert.Severity, "Encryption changes should be critical")

	t.Log("✅ S3 bucket encryption drift detected successfully")
}

// TestDriftDetection_MultipleChanges tests multiple drift events
func TestDriftDetection_MultipleChanges(t *testing.T) {
	skipIfShort(t)
	skipIfNoAWS(t)
	skipIfNoFalco(t)

	ctx := NewTestContext(t)
	defer ctx.Cleanup()

	t.Log("=== E2E Test: Multiple Drift Changes ===")

	// Start detector
	ctx.StartDetector()

	// Make multiple changes
	ctx.ModifyEC2Termination(ctx.testInstanceID, false)
	ctx.AddCleanup(func() {
		ctx.ModifyEC2Termination(ctx.testInstanceID, true)
	})

	// Wait a bit between changes to avoid rate limiting
	time.Sleep(5 * time.Second)

	ctx.DisableS3BucketEncryption(ctx.testBucketName)
	ctx.AddCleanup(func() {
		ctx.EnableS3BucketEncryption(ctx.testBucketName)
	})

	// We should receive at least 2 alerts
	alertCount := 0
	timeout := 25 * time.Minute
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) && alertCount < 2 {
		remaining := time.Until(deadline)
		alert := ctx.WaitForAlert(remaining)
		if alert != nil {
			alertCount++
			t.Logf("Received alert %d/%d: %s:%s", alertCount, 2, alert.ResourceType, alert.ResourceID)
		}
	}

	assert.GreaterOrEqual(t, alertCount, 2, "Expected at least 2 alerts")

	t.Logf("✅ Multiple drift changes detected successfully (%d alerts)", alertCount)
}

// TestDriftDetection_NoFalsePositives tests that legitimate Terraform changes don't trigger alerts
func TestDriftDetection_NoFalsePositives(t *testing.T) {
	t.Skip("TODO: Implement false positive test with Terraform apply")

	// This test would:
	// 1. Make a change via Terraform (not manual)
	// 2. Verify NO alert is generated
	// 3. This requires tagging Terraform changes somehow
}

// TestDriftDetection_UserContext tests that user identity is captured
func TestDriftDetection_UserContext(t *testing.T) {
	skipIfShort(t)
	skipIfNoAWS(t)
	skipIfNoFalco(t)

	ctx := NewTestContext(t)
	defer ctx.Cleanup()

	t.Log("=== E2E Test: User Context Capture ===")

	// Start detector
	ctx.StartDetector()

	// Make a change
	ctx.ModifyEC2Termination(ctx.testInstanceID, false)
	ctx.AddCleanup(func() {
		ctx.ModifyEC2Termination(ctx.testInstanceID, true)
	})

	// Wait for alert
	alert := ctx.WaitForAlert(20 * time.Minute)

	// Verify user context
	require := assert.New(t)
	require.NotNil(alert, "Expected alert")
	require.NotEmpty(alert.UserIdentity.UserName, "Expected username")
	require.NotEmpty(alert.UserIdentity.Type, "Expected user type")

	t.Logf("✅ User context captured: %s (%s)", alert.UserIdentity.UserName, alert.UserIdentity.Type)
}
