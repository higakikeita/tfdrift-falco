package testutil

import (
	"errors"
	"testing"
)

func TestAssertJSONEqual(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		actual   string
	}{
		{
			name:     "Equal JSON objects",
			expected: `{"name": "test", "value": 123}`,
			actual:   `{"value": 123, "name": "test"}`,
		},
		{
			name:     "Equal arrays",
			expected: `[1, 2, 3]`,
			actual:   `[1, 2, 3]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use actual test T to execute the function properly
			AssertJSONEqual(t, tt.expected, tt.actual)
		})
	}
}

func TestAssertValidJSON(t *testing.T) {
	tests := []struct {
		name    string
		jsonStr string
	}{
		{
			name:    "Valid JSON object",
			jsonStr: `{"name": "test", "value": 123}`,
		},
		{
			name:    "Valid JSON array",
			jsonStr: `["a", "b", "c"]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AssertValidJSON(t, tt.jsonStr)
		})
	}
}

func TestAssertContainsAll(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		expected []string
	}{
		{
			name:     "All elements present",
			slice:    []string{"quick", "brown", "fox", "dog"},
			expected: []string{"quick", "fox", "dog"},
		},
		{
			name:     "Empty expected",
			slice:    []string{"any", "string"},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AssertContainsAll(t, tt.slice, tt.expected)
		})
	}
}

func TestAssertMapContainsKeys(t *testing.T) {
	tests := []struct {
		name string
		m    map[string]interface{}
		keys []string
	}{
		{
			name: "All keys present",
			m:    map[string]interface{}{"a": 1, "b": 2, "c": 3},
			keys: []string{"a", "b", "c"},
		},
		{
			name: "Empty keys",
			m:    map[string]interface{}{"a": 1},
			keys: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AssertMapContainsKeys(t, tt.m, tt.keys)
		})
	}
}

func TestAssertEventually(t *testing.T) {
	tests := []struct {
		name       string
		condition  func() bool
		maxRetries int
		message    string
	}{
		{
			name: "Condition becomes true quickly",
			condition: func() bool {
				return true
			},
			maxRetries: 5,
			message:    "should pass",
		},
		{
			name: "Condition becomes true after retries",
			condition: func() func() bool {
				counter := 0
				return func() bool {
					counter++
					return counter >= 3
				}
			}(),
			maxRetries: 5,
			message:    "should pass after 3 tries",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AssertEventually(t, tt.condition, tt.maxRetries, tt.message)
		})
	}
}

func TestAssertNoError(t *testing.T) {
	AssertNoError(t, nil)
}

func TestAssertError(t *testing.T) {
	AssertError(t, errors.New("test error"))
}

func TestAssertEqual(t *testing.T) {
	tests := []struct {
		name     string
		expected interface{}
		actual   interface{}
	}{
		{
			name:     "Equal integers",
			expected: 42,
			actual:   42,
		},
		{
			name:     "Equal strings",
			expected: "hello",
			actual:   "hello",
		},
		{
			name:     "Nil values",
			expected: nil,
			actual:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AssertEqual(t, tt.expected, tt.actual)
		})
	}
}

func TestAssertNotEqual(t *testing.T) {
	AssertNotEqual(t, 42, 43)
	AssertNotEqual(t, "hello", "world")
}

func TestAssertNil(t *testing.T) {
	AssertNil(t, nil)
}

func TestAssertNotNil(t *testing.T) {
	AssertNotNil(t, "not nil")
	AssertNotNil(t, 42)
}

func TestAssertTrue(t *testing.T) {
	AssertTrue(t, true)
}

func TestAssertFalse(t *testing.T) {
	AssertFalse(t, false)
}

func TestAssertContains(t *testing.T) {
	AssertContains(t, "The quick brown fox", "quick")
	AssertContains(t, "hello world", "world")
}

func TestAssertNotContains(t *testing.T) {
	AssertNotContains(t, "The quick brown fox", "elephant")
	AssertNotContains(t, "hello", "world")
}

func TestAssertLen(t *testing.T) {
	AssertLen(t, []int{1, 2, 3}, 3)
	AssertLen(t, "hello", 5)
}

func TestAssertEmpty(t *testing.T) {
	AssertEmpty(t, []int{})
	AssertEmpty(t, "")
}

func TestAssertNotEmpty(t *testing.T) {
	AssertNotEmpty(t, []int{1, 2, 3})
	AssertNotEmpty(t, "hello")
}
