package parser

import (
	"strings"
	"testing"
)

func TestParseError_Error(t *testing.T) {
	tests := []struct {
		name     string
		error    *ParseError
		expected string
	}{
		{
			name: "complete error with all fields",
			error: &ParseError{
				FilePath:     "/path/to/file.yaml",
				LineNumber:   10,
				ColumnNumber: 5,
				PropertyPath: "dict.user_id",
				Message:      "array cannot be empty",
				Level:        ErrorLevelError,
			},
			expected: "/path/to/file.yaml:10:5 at dict.user_id [ERROR] array cannot be empty",
		},
		{
			name: "error without column number",
			error: &ParseError{
				FilePath:     "/path/to/file.json",
				LineNumber:   5,
				PropertyPath: "dict.name",
				Message:      "value must be an array",
				Level:        ErrorLevelWarning,
			},
			expected: "/path/to/file.json:5 at dict.name [WARNING] value must be an array",
		},
		{
			name: "error without line number",
			error: &ParseError{
				FilePath:     "/path/to/file.jsonl",
				PropertyPath: "dict.age",
				Message:      "invalid type",
				Level:        ErrorLevelError,
			},
			expected: "/path/to/file.jsonl at dict.age [ERROR] invalid type",
		},
		{
			name: "minimal error",
			error: &ParseError{
				Message: "something went wrong",
				Level:   ErrorLevelError,
			},
			expected: "[ERROR] something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.error.Error()
			if result != tt.expected {
				t.Errorf("ParseError.Error() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestDictValidationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		error    *DictValidationError
		expected string
	}{
		{
			name: "dict validation error with parse error",
			error: &DictValidationError{
				ParseError: &ParseError{
					FilePath:     "/path/to/file.yaml",
					LineNumber:   15,
					PropertyPath: "dict.user_type",
					Message:      "array cannot be empty",
					Level:        ErrorLevelError,
				},
				DictKey: "user_type",
			},
			expected: "/path/to/file.yaml:15 at dict.user_type [ERROR] array cannot be empty",
		},
		{
			name: "dict validation error without parse error",
			error: &DictValidationError{
				DictKey: "user_id",
				ParseError: &ParseError{
					Message: "invalid value",
					Level:   ErrorLevelError,
				},
			},
			expected: "[ERROR] invalid value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.error.Error()
			if result != tt.expected {
				t.Errorf("DictValidationError.Error() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestErrorCollection(t *testing.T) {
	t.Run("empty collection", func(t *testing.T) {
		collection := NewErrorCollection()

		if collection.HasErrors() {
			t.Error("Empty collection should not have errors")
		}

		if err := collection.ToError(); err != nil {
			t.Errorf("Empty collection ToError() should return nil, got %v", err)
		}
	})

	t.Run("single error", func(t *testing.T) {
		collection := NewErrorCollection()
		testErr := NewParseError("/test.yaml", 5, "dict.test", "test error")
		collection.Add(testErr)

		if !collection.HasErrors() {
			t.Error("Collection with error should have errors")
		}

		err := collection.ToError()
		if err == nil {
			t.Error("Collection with error ToError() should not return nil")
		}

		expected := "/test.yaml:5 at dict.test [ERROR] test error"
		if err.Error() != expected {
			t.Errorf("Single error message = %q, expected %q", err.Error(), expected)
		}
	})

	t.Run("multiple errors", func(t *testing.T) {
		collection := NewErrorCollection()
		err1 := NewParseError("/test.yaml", 5, "dict.test1", "first error")
		err2 := NewParseError("/test.yaml", 10, "dict.test2", "second error")

		collection.Add(err1)
		collection.Add(err2)

		if !collection.HasErrors() {
			t.Error("Collection with errors should have errors")
		}

		err := collection.ToError()
		if err == nil {
			t.Error("Collection with errors ToError() should not return nil")
		}

		errorMsg := err.Error()
		if !strings.Contains(errorMsg, "Multiple errors (2):") {
			t.Errorf("Multiple errors message should contain count, got: %s", errorMsg)
		}

		if !strings.Contains(errorMsg, "first error") {
			t.Errorf("Multiple errors message should contain first error, got: %s", errorMsg)
		}

		if !strings.Contains(errorMsg, "second error") {
			t.Errorf("Multiple errors message should contain second error, got: %s", errorMsg)
		}
	})

	t.Run("add nil error", func(t *testing.T) {
		collection := NewErrorCollection()
		collection.Add(nil)

		if collection.HasErrors() {
			t.Error("Collection with nil error should not have errors")
		}
	})
}

func TestNewParseError(t *testing.T) {
	err := NewParseError("/test.yaml", 10, "dict.test", "test message")

	if err.FilePath != "/test.yaml" {
		t.Errorf("FilePath = %q, expected %q", err.FilePath, "/test.yaml")
	}

	if err.LineNumber != 10 {
		t.Errorf("LineNumber = %d, expected %d", err.LineNumber, 10)
	}

	if err.PropertyPath != "dict.test" {
		t.Errorf("PropertyPath = %q, expected %q", err.PropertyPath, "dict.test")
	}

	if err.Message != "test message" {
		t.Errorf("Message = %q, expected %q", err.Message, "test message")
	}

	if err.Level != ErrorLevelError {
		t.Errorf("Level = %v, expected %v", err.Level, ErrorLevelError)
	}
}

func TestNewDictValidationError(t *testing.T) {
	err := NewDictValidationError("/test.yaml", 15, "dict.user_id", "user_id", "array cannot be empty")

	if err.DictKey != "user_id" {
		t.Errorf("DictKey = %q, expected %q", err.DictKey, "user_id")
	}

	if err.ParseError == nil {
		t.Error("ParseError should not be nil")
		return
	}

	if err.ParseError.FilePath != "/test.yaml" {
		t.Errorf("ParseError.FilePath = %q, expected %q", err.ParseError.FilePath, "/test.yaml")
	}

	if err.ParseError.LineNumber != 15 {
		t.Errorf("ParseError.LineNumber = %d, expected %d", err.ParseError.LineNumber, 15)
	}
}

func TestErrorLevel_String(t *testing.T) {
	tests := []struct {
		level    ErrorLevel
		expected string
	}{
		{ErrorLevelError, "ERROR"},
		{ErrorLevelWarning, "WARNING"},
		{ErrorLevelInfo, "INFO"},
		{ErrorLevel(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.level.String()
			if result != tt.expected {
				t.Errorf("ErrorLevel.String() = %q, expected %q", result, tt.expected)
			}
		})
	}
}
