package testutil

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertJSONEqual asserts that two JSON strings are equal
func AssertJSONEqual(t *testing.T, expected, actual string) {
	t.Helper()

	var expectedJSON, actualJSON interface{}

	err := json.Unmarshal([]byte(expected), &expectedJSON)
	require.NoError(t, err, "Failed to unmarshal expected JSON")

	err = json.Unmarshal([]byte(actual), &actualJSON)
	require.NoError(t, err, "Failed to unmarshal actual JSON")

	assert.Equal(t, expectedJSON, actualJSON)
}

// AssertValidJSON asserts that a string is valid JSON
func AssertValidJSON(t *testing.T, jsonStr string) {
	t.Helper()

	var js interface{}
	err := json.Unmarshal([]byte(jsonStr), &js)
	require.NoError(t, err, "Invalid JSON: %s", jsonStr)
}

// AssertContainsAll asserts that a slice contains all expected elements
func AssertContainsAll(t *testing.T, slice []string, expected []string) {
	t.Helper()

	for _, exp := range expected {
		assert.Contains(t, slice, exp, "Slice does not contain expected element: %s", exp)
	}
}

// AssertMapContainsKeys asserts that a map contains all expected keys
func AssertMapContainsKeys(t *testing.T, m map[string]interface{}, expectedKeys []string) {
	t.Helper()

	for _, key := range expectedKeys {
		_, ok := m[key]
		assert.True(t, ok, "Map does not contain expected key: %s", key)
	}
}

// AssertEventually asserts that a condition becomes true within a number of retries
func AssertEventually(t *testing.T, condition func() bool, maxRetries int, message string) {
	t.Helper()

	for i := 0; i < maxRetries; i++ {
		if condition() {
			return
		}
	}

	t.Fatalf("Condition did not become true after %d retries: %s", maxRetries, message)
}

// AssertNoError is a helper that fails the test if an error is not nil
func AssertNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	require.NoError(t, err, msgAndArgs...)
}

// AssertError is a helper that fails the test if an error is nil
func AssertError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	require.Error(t, err, msgAndArgs...)
}

// AssertEqual is a helper that asserts two values are equal
func AssertEqual(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	assert.Equal(t, expected, actual, msgAndArgs...)
}

// AssertNotEqual is a helper that asserts two values are not equal
func AssertNotEqual(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	assert.NotEqual(t, expected, actual, msgAndArgs...)
}

// AssertNil is a helper that asserts a value is nil
func AssertNil(t *testing.T, value interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	assert.Nil(t, value, msgAndArgs...)
}

// AssertNotNil is a helper that asserts a value is not nil
func AssertNotNil(t *testing.T, value interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	assert.NotNil(t, value, msgAndArgs...)
}

// AssertTrue is a helper that asserts a value is true
func AssertTrue(t *testing.T, value bool, msgAndArgs ...interface{}) {
	t.Helper()
	assert.True(t, value, msgAndArgs...)
}

// AssertFalse is a helper that asserts a value is false
func AssertFalse(t *testing.T, value bool, msgAndArgs ...interface{}) {
	t.Helper()
	assert.False(t, value, msgAndArgs...)
}

// AssertContains is a helper that asserts a string contains a substring
func AssertContains(t *testing.T, str, substr string, msgAndArgs ...interface{}) {
	t.Helper()
	assert.Contains(t, str, substr, msgAndArgs...)
}

// AssertNotContains is a helper that asserts a string does not contain a substring
func AssertNotContains(t *testing.T, str, substr string, msgAndArgs ...interface{}) {
	t.Helper()
	assert.NotContains(t, str, substr, msgAndArgs...)
}

// AssertLen is a helper that asserts a collection has a specific length
func AssertLen(t *testing.T, obj interface{}, length int, msgAndArgs ...interface{}) {
	t.Helper()
	assert.Len(t, obj, length, msgAndArgs...)
}

// AssertEmpty is a helper that asserts a collection is empty
func AssertEmpty(t *testing.T, obj interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	assert.Empty(t, obj, msgAndArgs...)
}

// AssertNotEmpty is a helper that asserts a collection is not empty
func AssertNotEmpty(t *testing.T, obj interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	assert.NotEmpty(t, obj, msgAndArgs...)
}
