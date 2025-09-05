package parser

import (
	"strings"
	"testing"
)

func TestParser_validateDictWithPosition(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name          string
		dict          map[string][]interface{}
		filePath      string
		fileExt       string
		content       string
		expectError   bool
		errorContains []string
	}{
		{
			name: "valid dict",
			dict: map[string][]interface{}{
				"user_id": {1, 2, 3},
				"name":    {"Alice", "Bob"},
			},
			expectError: false,
		},
		{
			name: "empty array",
			dict: map[string][]interface{}{
				"user_id": {},
			},
			filePath:    "/test.yaml",
			fileExt:     ".yaml",
			content:     "dict:\n  user_id: []",
			expectError: true,
			errorContains: []string{
				"/test.yaml",
				"dict.user_id",
				"array cannot be empty",
			},
		},
		{
			name: "nil array",
			dict: map[string][]interface{}{
				"user_id": nil,
			},
			filePath:    "/test.json",
			fileExt:     ".json",
			content:     `{"dict": {"user_id": null}}`,
			expectError: true,
			errorContains: []string{
				"/test.json",
				"dict.user_id",
				"value must be an array, got null",
			},
		},
		{
			name: "non-primitive array element",
			dict: map[string][]interface{}{
				"user_data": {
					map[string]interface{}{"id": 1, "name": "Alice"},
					"Bob",
				},
			},
			filePath:    "/test.yaml",
			fileExt:     ".yaml",
			content:     "dict:\n  user_data:\n    - id: 1\n      name: Alice\n    - Bob",
			expectError: true,
			errorContains: []string{
				"/test.yaml",
				"dict.user_data[0]",
				"must be a primitive value",
				"got map[string]interface {}",
			},
		},
		{
			name: "multiple errors",
			dict: map[string][]interface{}{
				"empty_array": {},
				"nil_array":   nil,
				"invalid_elements": {
					[]interface{}{1, 2, 3},
					"valid_string",
				},
			},
			filePath:    "/test.json",
			fileExt:     ".json",
			content:     `{"dict": {"empty_array": [], "nil_array": null, "invalid_elements": [[1,2,3], "valid_string"]}}`,
			expectError: true,
			errorContains: []string{
				"Multiple errors",
				"empty_array",
				"array cannot be empty",
				"nil_array",
				"value must be an array, got null",
				"invalid_elements[0]",
				"must be a primitive value",
			},
		},
		{
			name:        "nil dict",
			dict:        nil,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parser.validateDictWithPosition(tt.dict, tt.filePath, tt.fileExt, tt.content)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
					return
				}

				errorMsg := err.Error()
				for _, expectedText := range tt.errorContains {
					if !strings.Contains(errorMsg, expectedText) {
						t.Errorf("Error message should contain %q, got: %s", expectedText, errorMsg)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestParser_isPrimitiveValue(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		{"string", "hello", true},
		{"int", 42, true},
		{"int8", int8(42), true},
		{"int16", int16(42), true},
		{"int32", int32(42), true},
		{"int64", int64(42), true},
		{"uint", uint(42), true},
		{"uint8", uint8(42), true},
		{"uint16", uint16(42), true},
		{"uint32", uint32(42), true},
		{"uint64", uint64(42), true},
		{"float32", float32(3.14), true},
		{"float64", 3.14, true},
		{"bool true", true, true},
		{"bool false", false, true},
		{"map", map[string]interface{}{"key": "value"}, false},
		{"slice", []interface{}{1, 2, 3}, false},
		{"array", [3]int{1, 2, 3}, false},
		{"nil", nil, false},
		{"struct", struct{ Name string }{Name: "test"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.isPrimitiveValue(tt.value)
			if result != tt.expected {
				t.Errorf("isPrimitiveValue(%v) = %v, expected %v", tt.value, result, tt.expected)
			}
		})
	}
}

func TestParser_createDictValidationError(t *testing.T) {
	parser := NewParser()

	position := &PositionInfo{
		Line:   10,
		Column: 5,
	}

	err := parser.createDictValidationError("/test.yaml", position, "dict.user_id", "user_id", "test message")

	if err.DictKey != "user_id" {
		t.Errorf("DictKey = %q, expected %q", err.DictKey, "user_id")
	}

	if err.ParseError == nil {
		t.Error("ParseError should not be nil")
		return
	}

	if err.ParseError.FilePath != "/test.yaml" {
		t.Errorf("FilePath = %q, expected %q", err.ParseError.FilePath, "/test.yaml")
	}

	if err.ParseError.LineNumber != 10 {
		t.Errorf("LineNumber = %d, expected %d", err.ParseError.LineNumber, 10)
	}

	if err.ParseError.ColumnNumber != 5 {
		t.Errorf("ColumnNumber = %d, expected %d", err.ParseError.ColumnNumber, 5)
	}

	if err.ParseError.PropertyPath != "dict.user_id" {
		t.Errorf("PropertyPath = %q, expected %q", err.ParseError.PropertyPath, "dict.user_id")
	}

	if err.ParseError.Message != "test message" {
		t.Errorf("Message = %q, expected %q", err.ParseError.Message, "test message")
	}

	if err.ParseError.Level != ErrorLevelError {
		t.Errorf("Level = %v, expected %v", err.ParseError.Level, ErrorLevelError)
	}
}

func TestParser_createDictValidationError_nilPosition(t *testing.T) {
	parser := NewParser()

	err := parser.createDictValidationError("/test.yaml", nil, "dict.user_id", "user_id", "test message")

	if err.ParseError.LineNumber != 0 {
		t.Errorf("LineNumber = %d, expected %d", err.ParseError.LineNumber, 0)
	}

	if err.ParseError.ColumnNumber != 0 {
		t.Errorf("ColumnNumber = %d, expected %d", err.ParseError.ColumnNumber, 0)
	}
}
