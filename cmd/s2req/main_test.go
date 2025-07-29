package main

import (
	"testing"
)

func TestVarFlags_Set(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		expected  map[string]interface{}
		wantError bool
	}{
		{
			name:      "simple string value",
			value:     "key=value",
			expected:  map[string]interface{}{"key": "value"},
			wantError: false,
		},
		{
			name:      "integer value",
			value:     "id=123",
			expected:  map[string]interface{}{"id": 123},
			wantError: false,
		},
		{
			name:      "float value",
			value:     "rate=3.14",
			expected:  map[string]interface{}{"rate": 3.14},
			wantError: false,
		},
		{
			name:      "boolean value true",
			value:     "enabled=true",
			expected:  map[string]interface{}{"enabled": true},
			wantError: false,
		},
		{
			name:      "boolean value false",
			value:     "disabled=false",
			expected:  map[string]interface{}{"disabled": false},
			wantError: false,
		},
		{
			name:      "JSON array value",
			value:     `ids=[1,2,3]`,
			expected:  map[string]interface{}{"ids": []interface{}{float64(1), float64(2), float64(3)}},
			wantError: false,
		},
		{
			name:      "JSON object value",
			value:     `config={"name":"test","count":5}`,
			expected:  map[string]interface{}{"config": map[string]interface{}{"name": "test", "count": float64(5)}},
			wantError: false,
		},
		{
			name:      "string with spaces",
			value:     `message=hello world`,
			expected:  map[string]interface{}{"message": "hello world"},
			wantError: false,
		},
		{
			name:      "value with equals sign",
			value:     `formula=x=y+z`,
			expected:  map[string]interface{}{"formula": "x=y+z"},
			wantError: false,
		},
		{
			name:      "invalid format - no equals",
			value:     "invalid",
			expected:  nil,
			wantError: true,
		},
		{
			name:      "empty key",
			value:     "=value",
			expected:  nil,
			wantError: true,
		},
		{
			name:      "key with spaces",
			value:     " key =value",
			expected:  map[string]interface{}{"key": "value"},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := make(varFlags)
			err := flags.Set(tt.value)

			if tt.wantError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if tt.wantError {
				return
			}

			// Check that all expected keys and values are present
			for key, expectedValue := range tt.expected {
				actualValue, exists := flags[key]
				if !exists {
					t.Errorf("Expected key %q not found in flags", key)
					continue
				}

				// For complex types like arrays and objects, compare using string representation
				if !compareValues(expectedValue, actualValue) {
					t.Errorf("For key %q, expected %v (%T), got %v (%T)", key, expectedValue, expectedValue, actualValue, actualValue)
				}
			}

			// Check that no unexpected keys are present
			if len(flags) != len(tt.expected) {
				t.Errorf("Expected %d keys, got %d", len(tt.expected), len(flags))
			}
		})
	}
}

func TestVarFlags_MultipleSet(t *testing.T) {
	flags := make(varFlags)
	
	err := flags.Set("name=test")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	err = flags.Set("id=123")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	err = flags.Set("items=[1,2,3]")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Check all values are present
	if flags["name"] != "test" {
		t.Errorf("Expected name=test, got %v", flags["name"])
	}
	
	if !compareValues(123, flags["id"]) {
		t.Errorf("Expected id=123, got %v", flags["id"])
	}
	
	expectedItems := []interface{}{float64(1), float64(2), float64(3)}
	if !compareValues(flags["items"], expectedItems) {
		t.Errorf("Expected items=%v, got %v", expectedItems, flags["items"])
	}
	
	if len(flags) != 3 {
		t.Errorf("Expected 3 flags, got %d", len(flags))
	}
}

func TestVarFlags_Override(t *testing.T) {
	flags := make(varFlags)
	
	// Set initial value
	err := flags.Set("key=original")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Override with new value
	err = flags.Set("key=overridden")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Check that the value was overridden
	if flags["key"] != "overridden" {
		t.Errorf("Expected key=overridden, got %v", flags["key"])
	}
	
	if len(flags) != 1 {
		t.Errorf("Expected 1 flag, got %d", len(flags))
	}
}

// compareValues compares two values, handling the JSON unmarshaling quirks
func compareValues(expected, actual interface{}) bool {
	// Handle arrays
	if expectedArr, ok := expected.([]interface{}); ok {
		if actualArr, ok := actual.([]interface{}); ok {
			if len(expectedArr) != len(actualArr) {
				return false
			}
			for i, expectedItem := range expectedArr {
				if !compareValues(expectedItem, actualArr[i]) {
					return false
				}
			}
			return true
		}
		return false
	}
	
	// Handle maps
	if expectedMap, ok := expected.(map[string]interface{}); ok {
		if actualMap, ok := actual.(map[string]interface{}); ok {
			if len(expectedMap) != len(actualMap) {
				return false
			}
			for key, expectedValue := range expectedMap {
				actualValue, exists := actualMap[key]
				if !exists || !compareValues(expectedValue, actualValue) {
					return false
				}
			}
			return true
		}
		return false
	}
	
	// Handle numeric type mismatches (int vs float64)
	if expectedInt, ok := expected.(int); ok {
		if actualFloat, ok := actual.(float64); ok {
			return float64(expectedInt) == actualFloat
		}
		if actualInt, ok := actual.(int); ok {
			return expectedInt == actualInt
		}
	}
	if expectedFloat, ok := expected.(float64); ok {
		if actualInt, ok := actual.(int); ok {
			return expectedFloat == float64(actualInt)
		}
		if actualFloat, ok := actual.(float64); ok {
			return expectedFloat == actualFloat
		}
	}
	
	// Handle primitive types
	return expected == actual
}