package functions

import (
	"context"
	"strings"
	"testing"
)

func TestDictFunction_Execute_EnhancedErrors(t *testing.T) {
	fn := &DictFunction{}

	tests := []struct {
		name          string
		args          []interface{}
		contextDict   map[string]interface{}
		contextFile   string
		expectError   bool
		errorContains []string
	}{
		{
			name:        "invalid argument count",
			args:        []interface{}{"var1", "var2"},
			expectError: true,
			errorContains: []string{
				"dict function expects 1 argument, got 2",
			},
		},
		{
			name:        "invalid argument type",
			args:        []interface{}{123},
			expectError: true,
			errorContains: []string{
				"dict function expects string argument, got int",
			},
		},
		{
			name:        "no dict context",
			args:        []interface{}{"user_id"},
			expectError: true,
			errorContains: []string{
				"$dict reference 'user_id' found but no dict variables are defined",
			},
		},
		{
			name:        "no dict context with file path",
			args:        []interface{}{"user_id"},
			contextFile: "/test/config.yaml",
			expectError: true,
			errorContains: []string{
				"$dict reference 'user_id' found but no dict variables are defined in /test/config.yaml",
			},
		},
		{
			name: "variable not found with available variables",
			args: []interface{}{"missing_var"},
			contextDict: map[string]interface{}{
				"user_id": 123,
				"name":    "Alice",
			},
			expectError: true,
			errorContains: []string{
				"$dict reference 'missing_var' not found",
				"Available dict variables:",
				"user_id",
				"name",
			},
		},
		{
			name: "variable not found with available variables and file path",
			args: []interface{}{"missing_var"},
			contextDict: map[string]interface{}{
				"user_id": 123,
				"name":    "Alice",
			},
			contextFile: "/test/config.yaml",
			expectError: true,
			errorContains: []string{
				"$dict reference 'missing_var' not found in /test/config.yaml",
				"Available dict variables:",
				"user_id",
				"name",
			},
		},
		{
			name:        "variable not found with empty dict",
			args:        []interface{}{"user_id"},
			contextDict: map[string]interface{}{},
			expectError: true,
			errorContains: []string{
				"$dict reference 'user_id' not found",
				"No dict variables are defined",
			},
		},
		{
			name:        "variable not found with empty dict and file path",
			args:        []interface{}{"user_id"},
			contextDict: map[string]interface{}{},
			contextFile: "/test/config.yaml",
			expectError: true,
			errorContains: []string{
				"$dict reference 'user_id' not found in /test/config.yaml",
				"No dict variables are defined",
			},
		},
		{
			name: "successful variable retrieval",
			args: []interface{}{"user_id"},
			contextDict: map[string]interface{}{
				"user_id": 123,
				"name":    "Alice",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Add dict context if provided
			if tt.contextDict != nil {
				ctx = context.WithValue(ctx, "dict", tt.contextDict)
			}

			// Add file path context if provided
			if tt.contextFile != "" {
				ctx = context.WithValue(ctx, "requestFilePath", tt.contextFile)
			}

			result, err := fn.Execute(ctx, tt.args)

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
					return
				}

				// For successful cases, verify the result
				if tt.name == "successful variable retrieval" {
					expected := tt.contextDict["user_id"]
					if result != expected {
						t.Errorf("Expected result %v, got %v", expected, result)
					}
				}
			}
		})
	}
}

func TestDictFunction_Execute_ContextHandling(t *testing.T) {
	fn := &DictFunction{}

	t.Run("context without dict key", func(t *testing.T) {
		ctx := context.Background()
		// Don't add dict context

		_, err := fn.Execute(ctx, []interface{}{"user_id"})

		if err == nil {
			t.Error("Expected error when dict context is missing")
			return
		}

		if !strings.Contains(err.Error(), "no dict variables are defined") {
			t.Errorf("Error should mention missing dict variables, got: %s", err.Error())
		}
	})

	t.Run("context with wrong type for dict", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "dict", "not a map")

		_, err := fn.Execute(ctx, []interface{}{"user_id"})

		if err == nil {
			t.Error("Expected error when dict context has wrong type")
			return
		}

		if !strings.Contains(err.Error(), "no dict variables are defined") {
			t.Errorf("Error should mention missing dict variables, got: %s", err.Error())
		}
	})

	t.Run("context with file path but wrong dict type", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "dict", "not a map")
		ctx = context.WithValue(ctx, "requestFilePath", "/test.yaml")

		_, err := fn.Execute(ctx, []interface{}{"user_id"})

		if err == nil {
			t.Error("Expected error when dict context has wrong type")
			return
		}

		expectedMsg := "$dict reference 'user_id' found but no dict variables are defined in /test.yaml"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message %q, got %q", expectedMsg, err.Error())
		}
	})
}
