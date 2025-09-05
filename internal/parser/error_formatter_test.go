package parser

import (
	"strings"
	"testing"
)

func TestErrorFormatter_FormatError(t *testing.T) {
	formatter := NewErrorFormatter()

	tests := []struct {
		name           string
		error          error
		expectContains []string
	}{
		{
			name: "parse error with all fields",
			error: &ParseError{
				FilePath:     "/test/config.yaml",
				LineNumber:   10,
				ColumnNumber: 5,
				PropertyPath: "dict.user_id",
				Message:      "array cannot be empty",
				Level:        ErrorLevelError,
				SourceLine:   "  user_id: []",
			},
			expectContains: []string{
				"📁 /test/config.yaml:10:5",
				"🎯 Property: dict.user_id",
				"❌ ERROR: array cannot be empty",
				"📝 Source: user_id: []",
			},
		},
		{
			name: "dict validation error",
			error: &DictValidationError{
				ParseError: &ParseError{
					FilePath:     "/test/config.yaml",
					LineNumber:   15,
					PropertyPath: "dict.name",
					Message:      "value must be an array",
					Level:        ErrorLevelError,
				},
				DictKey: "name",
			},
			expectContains: []string{
				"🔧 Dict Configuration Error",
				"📁 /test/config.yaml:15",
				"🎯 Property: dict.name",
				"❌ ERROR: value must be an array",
				"🔑 Dict Key: name",
			},
		},
		{
			name: "error collection with multiple errors",
			error: &ErrorCollection{
				Errors: []error{
					&ParseError{
						FilePath:     "/test.yaml",
						LineNumber:   5,
						PropertyPath: "dict.user_id",
						Message:      "array cannot be empty",
						Level:        ErrorLevelError,
					},
					&ParseError{
						FilePath:     "/test.yaml",
						LineNumber:   10,
						PropertyPath: "body.id",
						Message:      "$dict reference 'missing_var': no dict variables are defined",
						Level:        ErrorLevelError,
					},
				},
			},
			expectContains: []string{
				"❌ Configuration Validation Failed (2 errors)",
				"🔧 Dict Validation (1)",
				"🔗 Dict References (1)",
				"array cannot be empty",
				"$dict reference 'missing_var'",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.FormatError(tt.error)

			for _, expected := range tt.expectContains {
				if !strings.Contains(result, expected) {
					t.Errorf("Formatted error should contain %q, got:\n%s", expected, result)
				}
			}
		})
	}
}

func TestErrorFormatter_categorizeErrors(t *testing.T) {
	formatter := NewErrorFormatter()

	errors := []error{
		&DictValidationError{
			ParseError: &ParseError{
				PropertyPath: "dict.user_id",
				Message:      "array cannot be empty",
			},
			DictKey: "user_id",
		},
		&ParseError{
			PropertyPath: "body.id",
			Message:      "$dict reference 'missing_var': not found",
		},
		&ParseError{
			PropertyPath: "dict.name",
			Message:      "invalid value",
		},
		&ParseError{
			PropertyPath: "headers.auth",
			Message:      "some other error",
		},
	}

	categories := formatter.categorizeErrors(errors)

	// Check Dict Validation category
	if len(categories["Dict Validation"]) != 2 {
		t.Errorf("Expected 2 dict validation errors, got %d", len(categories["Dict Validation"]))
	}

	// Check Dict References category
	if len(categories["Dict References"]) != 1 {
		t.Errorf("Expected 1 dict reference error, got %d", len(categories["Dict References"]))
	}

	// Check Other category
	if len(categories["Other"]) != 1 {
		t.Errorf("Expected 1 other error, got %d", len(categories["Other"]))
	}
}

func TestErrorFormatter_FormatErrorSummary(t *testing.T) {
	formatter := NewErrorFormatter()

	tests := []struct {
		name     string
		error    error
		expected string
	}{
		{
			name: "empty error collection",
			error: &ErrorCollection{
				Errors: []error{},
			},
			expected: "No errors",
		},
		{
			name: "single error in collection",
			error: &ErrorCollection{
				Errors: []error{
					&ParseError{Message: "test error", Level: ErrorLevelError},
				},
			},
			expected: "❌ ERROR: test error",
		},
		{
			name: "multiple errors with categories",
			error: &ErrorCollection{
				Errors: []error{
					&DictValidationError{
						ParseError: &ParseError{Message: "validation error"},
						DictKey:    "test",
					},
					&ParseError{
						Message: "$dict reference 'test': not found",
					},
				},
			},
			expected: "Validation failed with 2 errors",
		},
		{
			name: "non-collection error",
			error: &ParseError{
				Message: "simple error",
				Level:   ErrorLevelError,
			},
			expected: "[ERROR] simple error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.FormatErrorSummary(tt.error)

			if !strings.Contains(result, tt.expected) {
				t.Errorf("Summary should contain %q, got: %s", tt.expected, result)
			}
		})
	}
}

func TestErrorFormatter_getLevelIcon(t *testing.T) {
	formatter := NewErrorFormatter()

	tests := []struct {
		level    ErrorLevel
		expected string
	}{
		{ErrorLevelError, "❌"},
		{ErrorLevelWarning, "⚠️"},
		{ErrorLevelInfo, "ℹ️"},
		{ErrorLevel(999), "❓"},
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			result := formatter.getLevelIcon(tt.level)
			if result != tt.expected {
				t.Errorf("getLevelIcon(%v) = %q, expected %q", tt.level, result, tt.expected)
			}
		})
	}
}

func TestErrorFormatter_getCategoryIcon(t *testing.T) {
	formatter := NewErrorFormatter()

	tests := []struct {
		category string
		expected string
	}{
		{"Dict Validation", "🔧"},
		{"Dict References", "🔗"},
		{"Other", "📋"},
		{"Unknown", "📄"},
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			result := formatter.getCategoryIcon(tt.category)
			if result != tt.expected {
				t.Errorf("getCategoryIcon(%q) = %q, expected %q", tt.category, result, tt.expected)
			}
		})
	}
}

func TestErrorFormatter_sortErrors(t *testing.T) {
	formatter := NewErrorFormatter()

	errors := []error{
		&ParseError{
			FilePath:   "/test/b.yaml",
			LineNumber: 5,
			Message:    "error b",
		},
		&ParseError{
			FilePath:   "/test/a.yaml",
			LineNumber: 10,
			Message:    "error a2",
		},
		&ParseError{
			FilePath:   "/test/a.yaml",
			LineNumber: 5,
			Message:    "error a1",
		},
	}

	sorted := formatter.sortErrors(errors)

	// Should be sorted by file path, then line number
	expected := []string{
		"error a1", // /test/a.yaml:5
		"error a2", // /test/a.yaml:10
		"error b",  // /test/b.yaml:5
	}

	for i, expectedMsg := range expected {
		if parseErr := formatter.getParseError(sorted[i]); parseErr != nil {
			if parseErr.Message != expectedMsg {
				t.Errorf("sorted[%d].Message = %q, expected %q", i, parseErr.Message, expectedMsg)
			}
		} else {
			t.Errorf("sorted[%d] is not a ParseError", i)
		}
	}
}

func TestErrorFormatter_indentText(t *testing.T) {
	formatter := NewErrorFormatter()

	text := "First line\nSecond line\nThird line"
	expected := "First line\n   Second line\n   Third line"

	result := formatter.indentText(text, "   ")

	if result != expected {
		t.Errorf("indentText() = %q, expected %q", result, expected)
	}
}

func TestNewErrorFormatter(t *testing.T) {
	formatter := NewErrorFormatter()

	if formatter == nil {
		t.Error("NewErrorFormatter() should not return nil")
	}
}
