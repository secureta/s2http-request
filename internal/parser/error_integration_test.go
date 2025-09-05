package parser

import (
	"strings"
	"testing"
)

func TestEnhancedErrorHandling_Integration(t *testing.T) {
	parser := NewParser()
	formatter := NewErrorFormatter()

	tests := []struct {
		name          string
		content       string
		fileExt       string
		filePath      string
		expectError   bool
		errorContains []string
	}{
		{
			name: "comprehensive error aggregation",
			content: `{
  "method": "POST",
  "path": "/api/users",
  "dict": {
    "empty_array": [],
    "nil_value": null,
    "invalid_elements": [{"nested": "object"}, "valid_string"]
  },
  "body": {
    "id": {"$dict": "missing_var"},
    "name": {"$dict": "empty_array"},
    "type": {"$dict": "another_missing"}
  }
}`,
			fileExt:     ".json",
			filePath:    "/test/config.json",
			expectError: true,
			errorContains: []string{
				"Configuration Validation Failed",
				"Dict Validation",
				"Dict References",
				"empty_array",
				"array cannot be empty",
				"nil_value",
				"value must be an array, got null",
				"invalid_elements[0]",
				"must be a primitive value",
				"$dict reference 'missing_var'",
				"$dict reference 'another_missing'",
				"Available variables:",
			},
		},
		{
			name: "yaml with position information",
			content: `method: POST
path: /api/users
dict:
  user_id: []
  name: null
body:
  id:
    $dict: nonexistent_var
  name:
    $dict: user_id`,
			fileExt:     ".yaml",
			filePath:    "/test/config.yaml",
			expectError: true,
			errorContains: []string{
				"Multiple errors",
				"/test/config.yaml",
				"dict.user_id",
				"dict.name",
				"array cannot be empty",
				"value must be an array, got null",
				"$dict reference 'nonexistent_var'",
			},
		},
		{
			name: "valid configuration with no errors",
			content: `{
  "method": "POST",
  "path": "/api/users",
  "dict": {
    "user_id": [1, 2, 3],
    "name": ["Alice", "Bob"]
  },
  "body": {
    "id": {"$dict": "user_id"},
    "name": {"$dict": "name"}
  }
}`,
			fileExt:     ".json",
			filePath:    "/test/valid.json",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configs, err := parser.ParseMultiple([]byte(tt.content), tt.fileExt, tt.filePath)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
					return
				}

				// Test both raw error and formatted error
				rawErrorMsg := err.Error()
				formattedErrorMsg := formatter.FormatError(err)

				for _, expectedText := range tt.errorContains {
					found := strings.Contains(rawErrorMsg, expectedText) || strings.Contains(formattedErrorMsg, expectedText)
					if !found {
						t.Errorf("Error message should contain %q\nRaw: %s\nFormatted: %s", expectedText, rawErrorMsg, formattedErrorMsg)
					}
				}

				// Test error summary
				summary := formatter.FormatErrorSummary(err)
				if !strings.Contains(summary, "error") {
					t.Errorf("Error summary should mention errors, got: %s", summary)
				}

			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
					return
				}

				if configs == nil || len(configs) == 0 {
					t.Error("Expected valid configs but got nil or empty")
				}
			}
		})
	}
}

func TestErrorHandling_RealWorldScenarios(t *testing.T) {
	parser := NewParser()
	formatter := NewErrorFormatter()

	t.Run("complex nested dict references", func(t *testing.T) {
		content := `{
  "method": "POST",
  "path": "/api/users",
  "dict": {
    "user_type": ["admin", "user"],
    "status": ["active", "inactive"]
  },
  "headers": {
    "Authorization": {"$dict": "auth_token"},
    "Content-Type": "application/json"
  },
  "body": {
    "user": {
      "type": {"$dict": "user_type"},
      "status": {"$dict": "status"},
      "permissions": [
        {"$dict": "permission_level"},
        {"$dict": "user_type"}
      ]
    }
  }
}`

		_, err := parser.ParseMultiple([]byte(content), ".json", "/test/complex.json")

		if err == nil {
			t.Error("Expected error for missing dict variables")
			return
		}

		formattedError := formatter.FormatError(err)

		// Should contain information about missing variables
		expectedMissing := []string{"auth_token", "permission_level"}
		for _, missing := range expectedMissing {
			if !strings.Contains(formattedError, missing) {
				t.Errorf("Formatted error should mention missing variable %q, got: %s", missing, formattedError)
			}
		}

		// Should mention available variables
		if !strings.Contains(formattedError, "user_type") || !strings.Contains(formattedError, "status") {
			t.Errorf("Formatted error should mention available variables, got: %s", formattedError)
		}
	})

	t.Run("mixed validation and reference errors with formatting", func(t *testing.T) {
		content := `method: POST
path: /api/test
dict:
  valid_array: [1, 2, 3]
  empty_array: []
  complex_array: [{"invalid": "object"}, "valid"]
body:
  valid_ref:
    $dict: valid_array
  invalid_ref:
    $dict: missing_variable
  empty_ref:
    $dict: empty_array`

		_, err := parser.ParseMultiple([]byte(content), ".yaml", "/test/mixed.yaml")

		if err == nil {
			t.Error("Expected error for validation issues")
			return
		}

		formattedError := formatter.FormatError(err)

		// Should be categorized properly
		if !strings.Contains(formattedError, "Dict Validation") {
			t.Errorf("Should categorize dict validation errors, got: %s", formattedError)
		}

		// The dict reference error should be present (may be categorized differently due to nesting)
		if !strings.Contains(formattedError, "$dict reference 'missing_variable'") {
			t.Errorf("Should contain dict reference error, got: %s", formattedError)
		}

		// Should contain file path information
		if !strings.Contains(formattedError, "/test/mixed.yaml") {
			t.Errorf("Should contain file path information, got: %s", formattedError)
		}
	})
}

func TestErrorFormatter_UserFriendlyOutput(t *testing.T) {
	formatter := NewErrorFormatter()

	// Create a realistic error collection
	errorCollection := NewErrorCollection()

	// Add dict validation errors
	errorCollection.Add(NewDictValidationError("/project/config.yaml", 10, "dict.user_ids", "user_ids", "array cannot be empty"))
	errorCollection.Add(NewDictValidationError("/project/config.yaml", 15, "dict.permissions", "permissions", "value must be an array, got null"))

	// Add dict reference errors
	errorCollection.Add(NewParseError("/project/config.yaml", 25, "body.user.id", "$dict reference 'missing_user_id': dict variable 'missing_user_id' not found. Available variables: [user_ids, permissions]"))

	formattedError := formatter.FormatError(errorCollection)

	t.Log("Formatted error output:")
	t.Log(formattedError)

	// Verify the output contains user-friendly elements
	expectedElements := []string{
		"❌ Configuration Validation Failed",
		"🔧 Dict Validation",
		"🔗 Dict References",
		"📁 /project/config.yaml",
		"🎯 Property:",
		"array cannot be empty",
		"Available variables:",
	}

	for _, element := range expectedElements {
		if !strings.Contains(formattedError, element) {
			t.Errorf("Formatted error should contain %q", element)
		}
	}

	// Test summary
	summary := formatter.FormatErrorSummary(errorCollection)
	t.Log("Error summary:")
	t.Log(summary)

	if !strings.Contains(summary, "3 errors") {
		t.Errorf("Summary should mention 3 errors, got: %s", summary)
	}
}
