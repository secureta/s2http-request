package parser

import (
	"strings"
	"testing"

	"github.com/secureta/s2http-request/internal/config"
)

func TestParser_validateAllConfigurations(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name          string
		configs       []*config.RequestConfig
		filePath      string
		fileExt       string
		content       string
		expectError   bool
		errorContains []string
	}{
		{
			name: "valid configurations",
			configs: []*config.RequestConfig{
				{
					Method: "POST",
					Path:   "/api/users",
					Dict: map[string][]interface{}{
						"user_id": {1, 2, 3},
					},
					Body: map[string]interface{}{
						"id": map[string]interface{}{
							"$dict": "user_id",
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "dict validation errors",
			configs: []*config.RequestConfig{
				{
					Method: "POST",
					Path:   "/api/users",
					Dict: map[string][]interface{}{
						"empty_array": {},
						"nil_array":   nil,
					},
				},
			},
			filePath:    "/test.yaml",
			fileExt:     ".yaml",
			content:     "dict:\n  empty_array: []\n  nil_array: null",
			expectError: true,
			errorContains: []string{
				"Multiple errors",
				"empty_array",
				"array cannot be empty",
				"nil_array",
				"value must be an array, got null",
			},
		},
		{
			name: "dict reference errors",
			configs: []*config.RequestConfig{
				{
					Method: "POST",
					Path:   "/api/users",
					Body: map[string]interface{}{
						"id": map[string]interface{}{
							"$dict": "nonexistent_var",
						},
						"name": map[string]interface{}{
							"$dict": "another_missing_var",
						},
					},
				},
			},
			filePath:    "/test.json",
			fileExt:     ".json",
			content:     `{"method": "POST", "body": {"id": {"$dict": "nonexistent_var"}}}`,
			expectError: true,
			errorContains: []string{
				"Multiple errors",
				"$dict reference 'nonexistent_var'",
				"no dict variables are defined",
				"$dict reference 'another_missing_var'",
				"no dict variables are defined",
			},
		},
		{
			name: "dict reference with available variables",
			configs: []*config.RequestConfig{
				{
					Method: "POST",
					Path:   "/api/users",
					Dict: map[string][]interface{}{
						"user_id": {1, 2, 3},
						"name":    {"Alice", "Bob"},
					},
					Body: map[string]interface{}{
						"id": map[string]interface{}{
							"$dict": "missing_var",
						},
					},
				},
			},
			filePath:    "/test.yaml",
			fileExt:     ".yaml",
			content:     "dict:\n  user_id: [1, 2, 3]\nbody:\n  id:\n    $dict: missing_var",
			expectError: true,
			errorContains: []string{
				"$dict reference 'missing_var'",
				"dict variable 'missing_var' not found",
				"Available variables:",
				"user_id",
				"name",
			},
		},
		{
			name: "mixed validation and reference errors",
			configs: []*config.RequestConfig{
				{
					Method: "POST",
					Path:   "/api/users",
					Dict: map[string][]interface{}{
						"empty_array": {},
						"valid_array": {1, 2, 3},
					},
					Body: map[string]interface{}{
						"id": map[string]interface{}{
							"$dict": "missing_var",
						},
						"valid_id": map[string]interface{}{
							"$dict": "valid_array",
						},
					},
				},
			},
			filePath:    "/test.yaml",
			fileExt:     ".yaml",
			content:     "dict:\n  empty_array: []\n  valid_array: [1, 2, 3]",
			expectError: true,
			errorContains: []string{
				"Multiple errors",
				"empty_array",
				"array cannot be empty",
				"$dict reference 'missing_var'",
				"dict variable 'missing_var' not found",
				"Available variables:",
				"valid_array",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parser.validateAllConfigurations(tt.configs, tt.filePath, tt.fileExt, tt.content)

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

func TestParser_findAllDictReferences(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name     string
		config   *config.RequestConfig
		expected []DictReference
	}{
		{
			name: "no dict references",
			config: &config.RequestConfig{
				Method: "GET",
				Path:   "/api/users",
				Body:   "static content",
			},
			expected: []DictReference{},
		},
		{
			name: "single dict reference in body",
			config: &config.RequestConfig{
				Method: "POST",
				Path:   "/api/users",
				Body: map[string]interface{}{
					"id": map[string]interface{}{
						"$dict": "user_id",
					},
				},
			},
			expected: []DictReference{
				{PropertyPath: "body.id", VariableName: "user_id"},
			},
		},
		{
			name: "multiple dict references",
			config: &config.RequestConfig{
				Method: "POST",
				Path: map[string]interface{}{
					"$dict": "path_var",
				},
				Headers: map[string]interface{}{
					"Authorization": map[string]interface{}{
						"$dict": "auth_token",
					},
				},
				Body: map[string]interface{}{
					"user": map[string]interface{}{
						"id": map[string]interface{}{
							"$dict": "user_id",
						},
						"name": map[string]interface{}{
							"$dict": "user_name",
						},
					},
				},
			},
			expected: []DictReference{
				{PropertyPath: "path", VariableName: "path_var"},
				{PropertyPath: "headers.Authorization", VariableName: "auth_token"},
				{PropertyPath: "body.user.id", VariableName: "user_id"},
				{PropertyPath: "body.user.name", VariableName: "user_name"},
			},
		},
		{
			name: "dict references in arrays",
			config: &config.RequestConfig{
				Method: "POST",
				Path:   "/api/users",
				Body: []interface{}{
					map[string]interface{}{
						"$dict": "first_item",
					},
					map[string]interface{}{
						"id": map[string]interface{}{
							"$dict": "second_item_id",
						},
					},
				},
			},
			expected: []DictReference{
				{PropertyPath: "body[0]", VariableName: "first_item"},
				{PropertyPath: "body[1].id", VariableName: "second_item_id"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			refs := parser.findAllDictReferences(tt.config)

			if len(refs) != len(tt.expected) {
				t.Errorf("Expected %d references, got %d", len(tt.expected), len(refs))
				return
			}

			// Convert to maps for easier comparison
			expectedMap := make(map[string]string)
			actualMap := make(map[string]string)

			for _, ref := range tt.expected {
				expectedMap[ref.PropertyPath] = ref.VariableName
			}

			for _, ref := range refs {
				actualMap[ref.PropertyPath] = ref.VariableName
			}

			for path, expectedVar := range expectedMap {
				if actualVar, exists := actualMap[path]; !exists {
					t.Errorf("Expected reference at path %q not found", path)
				} else if actualVar != expectedVar {
					t.Errorf("At path %q: expected variable %q, got %q", path, expectedVar, actualVar)
				}
			}

			for path := range actualMap {
				if _, exists := expectedMap[path]; !exists {
					t.Errorf("Unexpected reference found at path %q", path)
				}
			}
		})
	}
}

func TestParser_validateDictReferences(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name          string
		config        *config.RequestConfig
		filePath      string
		expectError   bool
		errorContains []string
	}{
		{
			name: "valid references",
			config: &config.RequestConfig{
				Dict: map[string][]interface{}{
					"user_id": {1, 2, 3},
					"name":    {"Alice", "Bob"},
				},
				Body: map[string]interface{}{
					"id": map[string]interface{}{
						"$dict": "user_id",
					},
					"name": map[string]interface{}{
						"$dict": "name",
					},
				},
			},
			expectError: false,
		},
		{
			name: "no dict defined but references exist",
			config: &config.RequestConfig{
				Body: map[string]interface{}{
					"id": map[string]interface{}{
						"$dict": "user_id",
					},
				},
			},
			filePath:    "/test.yaml",
			expectError: true,
			errorContains: []string{
				"$dict reference 'user_id'",
				"no dict variables are defined",
			},
		},
		{
			name: "reference to undefined variable",
			config: &config.RequestConfig{
				Dict: map[string][]interface{}{
					"user_id": {1, 2, 3},
				},
				Body: map[string]interface{}{
					"id": map[string]interface{}{
						"$dict": "missing_var",
					},
				},
			},
			filePath:    "/test.yaml",
			expectError: true,
			errorContains: []string{
				"$dict reference 'missing_var'",
				"dict variable 'missing_var' not found",
				"Available variables:",
				"user_id",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parser.validateDictReferences(tt.config, tt.filePath, ".yaml", "", 0)

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
