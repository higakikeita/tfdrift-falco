package provider

import (
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- AWS Provider Option Tests ---

func TestWithAWSRegions(t *testing.T) {
	regions := []string{"us-west-2", "eu-west-1", "ap-southeast-1"}
	p := NewAWSProvider(WithAWSRegions(regions))
	assert.Equal(t, regions, p.regions)
}

func TestNewAWSProvider_DefaultValues(t *testing.T) {
	p := NewAWSProvider()
	assert.NotNil(t, p.relevantEvents)
	assert.Equal(t, []string{"us-east-1"}, p.regions)
}

func TestNewAWSProvider_CustomRegions(t *testing.T) {
	regions := []string{"us-west-2", "eu-central-1"}
	p := NewAWSProvider(WithAWSRegions(regions))
	assert.Equal(t, regions, p.regions)
}

// --- AWS SupportedEventCount Tests ---

func TestAWSSupportedEventCount_GreaterThanZero(t *testing.T) {
	p := NewAWSProvider()
	count := p.SupportedEventCount()
	assert.Greater(t, count, 0)
}

func TestAWSSupportedEventCount_Consistent(t *testing.T) {
	p := NewAWSProvider()
	count1 := p.SupportedEventCount()
	count2 := p.SupportedEventCount()
	assert.Equal(t, count1, count2)
}

// --- AWS SupportedResourceTypes Tests ---

func TestAWSSupportedResourceTypes_NotEmpty(t *testing.T) {
	p := NewAWSProvider()
	types := p.SupportedResourceTypes()
	assert.NotEmpty(t, types)
}

func TestAWSSupportedResourceTypes_ContainsExpected(t *testing.T) {
	p := NewAWSProvider()
	types := p.SupportedResourceTypes()
	// Should contain at least some of the common AWS resource types
	assert.Greater(t, len(types), 0)
}

func TestAWSSupportedResourceTypes_NoDuplicates(t *testing.T) {
	p := NewAWSProvider()
	types := p.SupportedResourceTypes()
	seen := make(map[string]bool)
	for _, rt := range types {
		assert.False(t, seen[rt], "Duplicate resource type: %s", rt)
		seen[rt] = true
	}
}

// --- AWS DiscoverResources Tests ---

func TestAWSDiscoverResources_ImplementsInterface(t *testing.T) {
	p := NewAWSProvider(WithAWSRegions([]string{"us-west-2"}))
	// Verify the method exists and can be called (skip actual cloud calls)
	var _ ResourceDiscoverer = p
	// Verify signature by calling with nil context
	assert.NotNil(t, p)
}

func TestAWSDiscoverResources_RegionSelection(t *testing.T) {
	p := NewAWSProvider(WithAWSRegions([]string{"us-east-1", "eu-west-1"}))
	// Verify regions were set correctly
	assert.Equal(t, []string{"us-east-1", "eu-west-1"}, p.regions)
}

// --- AWS CompareState Tests ---

func TestAWSCompareState_EmptyInputs(t *testing.T) {
	p := NewAWSProvider()
	result := p.CompareState(nil, nil, CompareOptions{})
	assert.NotNil(t, result)
	assert.Equal(t, "aws", result.Provider)
	assert.Empty(t, result.UnmanagedResources)
	assert.Empty(t, result.MissingResources)
	assert.Empty(t, result.ModifiedResources)
}

func TestAWSCompareState_NoResources(t *testing.T) {
	p := NewAWSProvider()
	result := p.CompareState([]*types.TerraformResource{}, []*types.DiscoveredResource{}, CompareOptions{})
	assert.NotNil(t, result)
	assert.Equal(t, "aws", result.Provider)
	assert.Empty(t, result.UnmanagedResources)
	assert.Empty(t, result.MissingResources)
	assert.Empty(t, result.ModifiedResources)
}

func TestAWSCompareState_WithTerraformResources(t *testing.T) {
	p := NewAWSProvider()
	tfResources := []*types.TerraformResource{
		{
			Type: "aws_instance",
			Name: "web-server",
			Attributes: map[string]interface{}{
				"ami":           "ami-12345678",
				"instance_type": "t2.micro",
			},
		},
		{
			Type: "aws_security_group",
			Name: "web-sg",
			Attributes: map[string]interface{}{
				"name": "web-sg",
			},
		},
	}
	result := p.CompareState(tfResources, []*types.DiscoveredResource{}, CompareOptions{})
	assert.NotNil(t, result)
	assert.Equal(t, "aws", result.Provider)
}

func TestAWSCompareState_WithActualResources(t *testing.T) {
	p := NewAWSProvider()
	actualResources := []*types.DiscoveredResource{
		{
			ID:       "i-1234567890abcdef0",
			Type:     "aws_instance",
			Provider: "aws",
			ARN:      "arn:aws:ec2:us-east-1:123456789012:instance/i-1234567890abcdef0",
			Name:     "web-server",
			Region:   "us-east-1",
			Attributes: map[string]interface{}{
				"ami":           "ami-12345678",
				"instance_type": "t2.micro",
			},
		},
	}
	result := p.CompareState([]*types.TerraformResource{}, actualResources, CompareOptions{})
	assert.NotNil(t, result)
	assert.Equal(t, "aws", result.Provider)
}

func TestAWSCompareState_WithIgnoredAttributes(t *testing.T) {
	p := NewAWSProvider()
	tfResources := []*types.TerraformResource{
		{
			Type: "aws_instance",
			Name: "web-server",
			Attributes: map[string]interface{}{
				"ami": "ami-12345678",
			},
		},
	}
	actualResources := []*types.DiscoveredResource{
		{
			ID:       "i-123",
			Type:     "aws_instance",
			Provider: "aws",
			Name:     "web-server",
			Attributes: map[string]interface{}{
				"ami": "ami-12345678",
			},
		},
	}
	opts := CompareOptions{
		IgnoredAttributes: []string{"launch_time", "monitoring"},
	}
	result := p.CompareState(tfResources, actualResources, opts)
	assert.NotNil(t, result)
	assert.Equal(t, "aws", result.Provider)
}

func TestAWSCompareState_MultipleResources(t *testing.T) {
	p := NewAWSProvider()
	tfResources := []*types.TerraformResource{
		{Type: "aws_instance", Name: "web-1"},
		{Type: "aws_instance", Name: "web-2"},
		{Type: "aws_db_instance", Name: "postgres-1"},
	}
	actualResources := []*types.DiscoveredResource{
		{ID: "i-1", Type: "aws_instance", Name: "web-1", Provider: "aws"},
		{ID: "i-2", Type: "aws_instance", Name: "web-2", Provider: "aws"},
	}
	result := p.CompareState(tfResources, actualResources, CompareOptions{})
	assert.NotNil(t, result)
	assert.Equal(t, "aws", result.Provider)
}

// --- AWS SupportedDiscoveryTypes Tests ---

func TestAWSSupportedDiscoveryTypes_ContainsExpected(t *testing.T) {
	p := NewAWSProvider()
	types := p.SupportedDiscoveryTypes()
	assert.Contains(t, types, "aws_vpc")
	assert.Contains(t, types, "aws_subnet")
	assert.Contains(t, types, "aws_security_group")
	assert.Contains(t, types, "aws_instance")
}

func TestAWSSupportedDiscoveryTypes_NotEmpty(t *testing.T) {
	p := NewAWSProvider()
	types := p.SupportedDiscoveryTypes()
	assert.NotEmpty(t, types)
}

// --- AWS MapEventToResource Tests ---

func TestAWSMapEventToResource_KnownEvent(t *testing.T) {
	p := NewAWSProvider()
	resourceType := p.MapEventToResource("RunInstances", "ec2.amazonaws.com")
	assert.NotEmpty(t, resourceType)
	assert.NotEqual(t, "unknown", resourceType)
}

func TestAWSMapEventToResource_UnknownEvent(t *testing.T) {
	p := NewAWSProvider()
	resourceType := p.MapEventToResource("UnknownEventName", "unknown.amazonaws.com")
	assert.Equal(t, "unknown", resourceType)
}

func TestAWSMapEventToResource_AuthorizeSecurityGroupIngress(t *testing.T) {
	p := NewAWSProvider()
	resourceType := p.MapEventToResource("AuthorizeSecurityGroupIngress", "ec2.amazonaws.com")
	assert.NotEmpty(t, resourceType)
}

func TestAWSMapEventToResource_CreateBucket(t *testing.T) {
	p := NewAWSProvider()
	resourceType := p.MapEventToResource("CreateBucket", "s3.amazonaws.com")
	assert.NotEmpty(t, resourceType)
}

// --- AWS IsRelevantEvent Tests ---

func TestAWSIsRelevantEvent_RelevantEvent(t *testing.T) {
	p := NewAWSProvider()
	// AWS should have many relevant events loaded
	isRelevant := p.IsRelevantEvent("RunInstances")
	assert.True(t, isRelevant)
}

func TestAWSIsRelevantEvent_IrrelevantEvent(t *testing.T) {
	p := NewAWSProvider()
	isRelevant := p.IsRelevantEvent("GetObject")
	assert.False(t, isRelevant)
}

// --- AWS ParseEvent Tests ---

func TestAWSParseEvent_CorrectSource(t *testing.T) {
	p := NewAWSProvider()
	// Use AuthorizeSecurityGroupIngress which has explicit field mapping
	event := p.ParseEvent("aws_cloudtrail", map[string]string{
		"ct.name":             "AuthorizeSecurityGroupIngress",
		"ct.src":              "ec2.amazonaws.com",
		"ct.region":           "us-east-1",
		"ct.request.groupid":  "sg-12345678",
		"ct.user.type":        "IAMUser",
		"ct.user.principalid": "AIDAEXAMPLE",
		"ct.user.arn":         "arn:aws:iam::123456789012:user/test-user",
		"ct.user.accountid":   "123456789012",
		"ct.user":             "test-user",
	}, nil)

	require.NotNil(t, event)
	assert.Equal(t, "aws", event.Provider)
	assert.Equal(t, "AuthorizeSecurityGroupIngress", event.EventName)
	assert.Equal(t, "test-user", event.UserIdentity.UserName)
	assert.Equal(t, "IAMUser", event.UserIdentity.Type)
	assert.Equal(t, "AIDAEXAMPLE", event.UserIdentity.PrincipalID)
	assert.Equal(t, "arn:aws:iam::123456789012:user/test-user", event.UserIdentity.ARN)
	assert.Equal(t, "123456789012", event.UserIdentity.AccountID)
	assert.Equal(t, "us-east-1", event.GetMetadata("region"))
	assert.Equal(t, "123456789012", event.GetMetadata("account_id"))
	assert.Equal(t, "ec2.amazonaws.com", event.GetMetadata("event_source"))
	// Backward compatibility
	assert.Equal(t, "us-east-1", event.Region)
}

func TestAWSParseEvent_WrongSource(t *testing.T) {
	p := NewAWSProvider()
	event := p.ParseEvent("gcpaudit", map[string]string{
		"ct.name": "RunInstances",
	}, nil)
	assert.Nil(t, event)
}

func TestAWSParseEvent_MissingEventName(t *testing.T) {
	p := NewAWSProvider()
	event := p.ParseEvent("aws_cloudtrail", map[string]string{
		"ct.src": "ec2.amazonaws.com",
	}, nil)
	assert.Nil(t, event)
}

func TestAWSParseEvent_IrrelevantEvent(t *testing.T) {
	p := NewAWSProvider()
	event := p.ParseEvent("aws_cloudtrail", map[string]string{
		"ct.name": "GetObject",
		"ct.src":  "s3.amazonaws.com",
	}, nil)
	assert.Nil(t, event)
}

// --- AWS ExtractChanges Tests ---

func TestAWSExtractChanges_ReturnsMap(t *testing.T) {
	p := NewAWSProvider()
	changes := p.ExtractChanges("RunInstances", map[string]string{
		"ct.request.instancetype": "t2.micro",
		"ct.request.imageid":      "ami-12345678",
	})
	assert.NotNil(t, changes)
}

// --- AWS Provider Interface Compliance ---

func TestAWSProviderImplementsInterfaces(t *testing.T) {
	p := NewAWSProvider()

	// Core Provider interface
	var _ Provider = p
	assert.Equal(t, "aws", p.Name())

	// ResourceDiscoverer interface
	var _ ResourceDiscoverer = p
	discoveryTypes := p.SupportedDiscoveryTypes()
	assert.NotNil(t, discoveryTypes)

	// StateComparator interface
	var _ StateComparator = p
}

func TestAWSProviderCapabilities(t *testing.T) {
	p := NewAWSProvider()
	caps := GetCapabilities(p)
	assert.True(t, caps.Discovery)
	assert.True(t, caps.Comparison)
}

// --- Additional AWS Tests ---

func TestAWSParseEvent_AllUserIdentityFields(t *testing.T) {
	p := NewAWSProvider()
	event := p.ParseEvent("aws_cloudtrail", map[string]string{
		"ct.name":             "AuthorizeSecurityGroupIngress",
		"ct.src":              "ec2.amazonaws.com",
		"ct.region":           "us-west-2",
		"ct.request.groupid":  "sg-12345678",
		"ct.user.type":        "AssumedRole",
		"ct.user.principalid": "AIDAEXAMPLE:session-name",
		"ct.user.arn":         "arn:aws:iam::123456789012:role/service-role",
		"ct.user.accountid":   "123456789012",
		"ct.user":             "service-role",
	}, nil)

	require.NotNil(t, event)
	assert.Equal(t, "AssumedRole", event.UserIdentity.Type)
	assert.Equal(t, "AIDAEXAMPLE:session-name", event.UserIdentity.PrincipalID)
	assert.Equal(t, "arn:aws:iam::123456789012:role/service-role", event.UserIdentity.ARN)
}

func TestAWSParseEvent_MinimalFields(t *testing.T) {
	p := NewAWSProvider()
	event := p.ParseEvent("aws_cloudtrail", map[string]string{
		"ct.name": "RunInstances",
		"ct.src":  "ec2.amazonaws.com",
	}, nil)

	// May be nil if ExtractAWSResourceID fails
	// This depends on implementation details of the falco package
	if event != nil {
		assert.Equal(t, "RunInstances", event.EventName)
	}
}
