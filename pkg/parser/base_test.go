package parser

import (
	"testing"

	"github.com/falcosecurity/client-go/pkg/api/outputs"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	"github.com/stretchr/testify/assert"
)

// createTestConfig creates a minimal test configuration
func createTestConfig() EventParserConfig {
	return EventParserConfig{
		Provider:       "test",
		ExpectedSource: "test_source",
		ExtractEventName: func(fields map[string]string) string {
			return GetStringField(fields, "event.name")
		},
		IsRelevantEvent: func(eventName string) bool {
			return eventName == "test.create" || eventName == "test.delete"
		},
		ExtractResourceID: func(eventName string, fields map[string]string) string {
			return GetStringField(fields, "resource.id")
		},
		MapResourceType: func(eventName string, fields map[string]string) string {
			return "test_resource"
		},
		ExtractUserIdentity: func(fields map[string]string) types.UserIdentity {
			return types.UserIdentity{
				Type:      "TestUser",
				UserName:  GetStringField(fields, "user.name"),
				AccountID: GetStringField(fields, "account.id"),
			}
		},
		ExtractChanges: func(eventName string, fields map[string]string) map[string]interface{} {
			changes := make(map[string]interface{})
			if eventName == "test.create" {
				changes["action"] = "create"
			}
			return changes
		},
		ExtractMetadata: func(eventName string, fields map[string]string) map[string]string {
			metadata := make(map[string]string)
			if region := GetStringField(fields, "region"); region != "" {
				metadata["region"] = region
			}
			return metadata
		},
	}
}

func TestNewBaseEventParser(t *testing.T) {
	config := createTestConfig()
	parser := NewBaseEventParser(config)
	assert.NotNil(t, parser)
	assert.Equal(t, "test", parser.config.Provider)
}

func TestBaseEventParser_Parse_NilResponse(t *testing.T) {
	config := createTestConfig()
	parser := NewBaseEventParser(config)

	event := parser.Parse(nil)
	assert.Nil(t, event, "Should return nil for nil response")
}

func TestBaseEventParser_Parse_WrongSource(t *testing.T) {
	config := createTestConfig()
	parser := NewBaseEventParser(config)

	res := &outputs.Response{
		Source:       "wrong_source",
		OutputFields: map[string]string{},
	}

	event := parser.Parse(res)
	assert.Nil(t, event, "Should return nil for wrong source")
}

func TestBaseEventParser_Parse_MissingEventName(t *testing.T) {
	config := createTestConfig()
	parser := NewBaseEventParser(config)

	res := &outputs.Response{
		Source:       "test_source",
		OutputFields: map[string]string{},
	}

	event := parser.Parse(res)
	assert.Nil(t, event, "Should return nil when event name is missing")
}

func TestBaseEventParser_Parse_IrrelevantEvent(t *testing.T) {
	config := createTestConfig()
	parser := NewBaseEventParser(config)

	res := &outputs.Response{
		Source: "test_source",
		OutputFields: map[string]string{
			"event.name": "test.read",
		},
	}

	event := parser.Parse(res)
	assert.Nil(t, event, "Should return nil for irrelevant event")
}

func TestBaseEventParser_Parse_MissingResourceID(t *testing.T) {
	config := createTestConfig()
	parser := NewBaseEventParser(config)

	res := &outputs.Response{
		Source: "test_source",
		OutputFields: map[string]string{
			"event.name": "test.create",
		},
	}

	event := parser.Parse(res)
	assert.Nil(t, event, "Should return nil when resource ID is missing")
}

func TestBaseEventParser_Parse_MissingResourceType(t *testing.T) {
	config := createTestConfig()
	config.MapResourceType = func(eventName string, fields map[string]string) string {
		return "" // No mapping
	}
	parser := NewBaseEventParser(config)

	res := &outputs.Response{
		Source: "test_source",
		OutputFields: map[string]string{
			"event.name":  "test.create",
			"resource.id": "res-123",
		},
	}

	event := parser.Parse(res)
	assert.Nil(t, event, "Should return nil when resource type mapping fails")
}

func TestBaseEventParser_Parse_SuccessfulParse(t *testing.T) {
	config := createTestConfig()
	parser := NewBaseEventParser(config)

	res := &outputs.Response{
		Source: "test_source",
		OutputFields: map[string]string{
			"event.name":  "test.create",
			"resource.id": "res-123",
			"user.name":   "testuser",
			"account.id":  "acc-456",
			"region":      "us-west-2",
		},
	}

	event := parser.Parse(res)
	assert.NotNil(t, event)
	assert.Equal(t, "test", event.Provider)
	assert.Equal(t, "test.create", event.EventName)
	assert.Equal(t, "test_resource", event.ResourceType)
	assert.Equal(t, "res-123", event.ResourceID)
	assert.Equal(t, "testuser", event.UserIdentity.UserName)
	assert.Equal(t, "acc-456", event.UserIdentity.AccountID)
	assert.NotNil(t, event.Changes)
	assert.Equal(t, "create", event.Changes["action"])
	assert.Equal(t, "us-west-2", event.Metadata["region"])
}

func TestBaseEventParser_Parse_WithoutMetadata(t *testing.T) {
	config := createTestConfig()
	config.ExtractMetadata = nil
	parser := NewBaseEventParser(config)

	res := &outputs.Response{
		Source: "test_source",
		OutputFields: map[string]string{
			"event.name":  "test.create",
			"resource.id": "res-123",
			"user.name":   "testuser",
			"account.id":  "acc-456",
		},
	}

	event := parser.Parse(res)
	assert.NotNil(t, event)
	assert.Nil(t, event.Metadata)
}

func TestBaseEventParser_Parse_EmptyMetadata(t *testing.T) {
	config := createTestConfig()
	parser := NewBaseEventParser(config)

	res := &outputs.Response{
		Source: "test_source",
		OutputFields: map[string]string{
			"event.name":  "test.create",
			"resource.id": "res-123",
			"user.name":   "testuser",
			"account.id":  "acc-456",
		},
	}

	event := parser.Parse(res)
	assert.NotNil(t, event)
	assert.Empty(t, event.Metadata)
}

func TestBaseEventParser_Parse_PreservesRawEvent(t *testing.T) {
	config := createTestConfig()
	parser := NewBaseEventParser(config)

	res := &outputs.Response{
		Source: "test_source",
		OutputFields: map[string]string{
			"event.name":  "test.create",
			"resource.id": "res-123",
		},
	}

	event := parser.Parse(res)
	assert.NotNil(t, event)
	assert.Equal(t, res, event.RawEvent)
}

func TestGetStringField_ExistingKey(t *testing.T) {
	fields := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	result := GetStringField(fields, "key1")
	assert.Equal(t, "value1", result)
}

func TestGetStringField_MissingKey(t *testing.T) {
	fields := map[string]string{
		"key1": "value1",
	}

	result := GetStringField(fields, "missing")
	assert.Equal(t, "", result)
}

func TestGetStringField_EmptyMap(t *testing.T) {
	fields := map[string]string{}

	result := GetStringField(fields, "key1")
	assert.Equal(t, "", result)
}

func TestBaseEventParser_Parse_AllFieldsPresent(t *testing.T) {
	config := createTestConfig()
	parser := NewBaseEventParser(config)

	res := &outputs.Response{
		Source: "test_source",
		OutputFields: map[string]string{
			"event.name":  "test.delete",
			"resource.id": "res-789",
			"user.name":   "admin",
			"account.id":  "acc-999",
			"region":      "eu-west-1",
		},
	}

	event := parser.Parse(res)
	assert.NotNil(t, event)
	assert.Equal(t, "test.delete", event.EventName)
	assert.Equal(t, "res-789", event.ResourceID)
	assert.Equal(t, "admin", event.UserIdentity.UserName)
	assert.Equal(t, "acc-999", event.UserIdentity.AccountID)
	assert.Equal(t, "eu-west-1", event.Metadata["region"])
}

func TestBaseEventParser_MultipleParses(t *testing.T) {
	config := createTestConfig()
	parser := NewBaseEventParser(config)

	// First event
	res1 := &outputs.Response{
		Source: "test_source",
		OutputFields: map[string]string{
			"event.name":  "test.create",
			"resource.id": "res-1",
			"user.name":   "user1",
		},
	}

	// Second event
	res2 := &outputs.Response{
		Source: "test_source",
		OutputFields: map[string]string{
			"event.name":  "test.delete",
			"resource.id": "res-2",
			"user.name":   "user2",
		},
	}

	event1 := parser.Parse(res1)
	event2 := parser.Parse(res2)

	assert.NotNil(t, event1)
	assert.NotNil(t, event2)
	assert.Equal(t, "res-1", event1.ResourceID)
	assert.Equal(t, "res-2", event2.ResourceID)
	assert.Equal(t, "user1", event1.UserIdentity.UserName)
	assert.Equal(t, "user2", event2.UserIdentity.UserName)
}
