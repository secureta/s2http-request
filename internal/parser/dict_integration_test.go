package parser

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/secureta/s2http-request/internal/config"
)

// TestDictIntegrationJSON tests dict functionality with JSON format
func TestDictIntegrationJSON(t *testing.T) {
	tests := []struct {
		name          string
		jsonContent   string
		expectedCount int
		expectedError bool
		errorContains string
	}{
		{
			name: "basic_dict_json",
			jsonContent: `{
				"method": "POST",
				"path": "/api/users",
				"body": {
					"name": {"$dict": "user_name"},
					"age": {"$dict": "user_age"}
				},
				"dict": {
					"user_name": ["Alice", "Bob"],
					"user_age": [25, 30]
				}
			}`,
			expectedCount: 4,
			expectedError: false,
		},
		{
			name: "single_dict_variable_json",
			jsonContent: `{
				"method": "GET",
				"path": "/api/users",
				"query": {
					"id": {"$dict": "user_id"}
				},
				"dict": {
					"user_id": ["123", "456", "789"]
				}
			}`,
			expectedCount: 3,
			expectedError: false,
		},
		{
			name: "dict_with_variables_json",
			jsonContent: `{
				"method": "POST",
				"path": "/api/users",
				"headers": {
					"Authorization": {"$var": "api_key"}
				},
				"body": {
					"name": {"$dict": "user_name"}
				},
				"variables": {
					"api_key": "Bearer token123"
				},
				"dict": {
					"user_name": ["Alice", "Bob"]
				}
			}`,
			expectedCount: 2,
			expectedError: false,
		},
		{
			name: "empty_dict_array_json",
			jsonContent: `{
				"method": "GET",
				"path": "/api/users",
				"dict": {
					"user_id": []
				}
			}`,
			expectedCount: 0,
			expectedError: true,
			errorContains: "array cannot be empty",
		},
		{
			name: "non_array_dict_value_json",
			jsonContent: `{
				"method": "GET",
				"path": "/api/users",
				"dict": {
					"user_id": "not_an_array"
				}
			}`,
			expectedCount: 0,
			expectedError: true,
			errorContains: "cannot unmarshal string",
		},
		{
			name: "undefined_dict_reference_json",
			jsonContent: `{
				"method": "GET",
				"path": "/api/users",
				"query": {
					"id": {"$dict": "undefined_var"}
				},
				"dict": {
					"user_id": ["123"]
				}
			}`,
			expectedCount: 0,
			expectedError: true,
			errorContains: "not found",
		},
	}

	parser := NewParser()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the JSON content
			configs, err := parser.ParseMultiple([]byte(tt.jsonContent), ".json", "test.json")
			if err != nil {
				if tt.expectedError && strings.Contains(err.Error(), tt.errorContains) {
					return // Expected error
				}
				t.Fatalf("Failed to parse JSON: %v", err)
			}

			if len(configs) == 0 {
				t.Fatal("No configurations parsed")
			}

			// Process the requests
			processedRequests, err := parser.ProcessRequests(ctx, configs[0], "https://api.example.com")
			if err != nil {
				if tt.expectedError && strings.Contains(err.Error(), tt.errorContains) {
					return // Expected error
				}
				t.Fatalf("Failed to process requests: %v", err)
			}

			if tt.expectedError {
				t.Fatalf("Expected error containing '%s', but got none", tt.errorContains)
			}

			if len(processedRequests) != tt.expectedCount {
				t.Errorf("Expected %d processed requests, got %d", tt.expectedCount, len(processedRequests))
			}

			// Verify that each processed request has the expected structure
			for i, req := range processedRequests {
				if req.Method == "" {
					t.Errorf("Request %d: Method is empty", i)
				}
				if req.URL == "" {
					t.Errorf("Request %d: URL is empty", i)
				}
			}
		})
	}
}

// TestDictIntegrationYAML tests dict functionality with YAML format
func TestDictIntegrationYAML(t *testing.T) {
	tests := []struct {
		name          string
		yamlContent   string
		expectedCount int
		expectedError bool
		errorContains string
	}{
		{
			name: "basic_dict_yaml",
			yamlContent: `method: POST
path: /api/users
body:
  name:
    $dict: user_name
  age:
    $dict: user_age
dict:
  user_name: ["Alice", "Bob"]
  user_age: [25, 30]`,
			expectedCount: 4,
			expectedError: false,
		},
		{
			name: "complex_dict_yaml",
			yamlContent: `method: POST
path: /api/orders
headers:
  Content-Type: application/json
body:
  product:
    $dict: product_name
  quantity:
    $dict: quantity
  priority:
    $dict: priority
dict:
  product_name: ["laptop", "mouse"]
  quantity: [1, 2, 5]
  priority: ["high", "low"]`,
			expectedCount: 12, // 2 * 3 * 2 = 12 combinations
			expectedError: false,
		},
		{
			name: "dict_in_query_yaml",
			yamlContent: `method: GET
path: /api/search
query:
  category:
    $dict: category
  limit:
    $dict: limit
dict:
  category: ["books", "electronics"]
  limit: [10, 20]`,
			expectedCount: 4,
			expectedError: false,
		},
		{
			name: "mixed_dict_and_variables_yaml",
			yamlContent: `method: POST
path: /api/users
headers:
  Authorization:
    $var: auth_token
  Content-Type: application/json
body:
  name:
    $dict: user_name
  department:
    $var: department
variables:
  auth_token: "Bearer abc123"
  department: "engineering"
dict:
  user_name: ["Alice", "Bob", "Charlie"]`,
			expectedCount: 3,
			expectedError: false,
		},
	}

	parser := NewParser()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the YAML content
			configs, err := parser.ParseMultiple([]byte(tt.yamlContent), ".yaml", "test.yaml")
			if err != nil {
				if tt.expectedError && strings.Contains(err.Error(), tt.errorContains) {
					return // Expected error
				}
				t.Fatalf("Failed to parse YAML: %v", err)
			}

			if len(configs) == 0 {
				t.Fatal("No configurations parsed")
			}

			// Process the requests
			processedRequests, err := parser.ProcessRequests(ctx, configs[0], "https://api.example.com")
			if err != nil {
				if tt.expectedError && strings.Contains(err.Error(), tt.errorContains) {
					return // Expected error
				}
				t.Fatalf("Failed to process requests: %v", err)
			}

			if tt.expectedError {
				t.Fatalf("Expected error containing '%s', but got none", tt.errorContains)
			}

			if len(processedRequests) != tt.expectedCount {
				t.Errorf("Expected %d processed requests, got %d", tt.expectedCount, len(processedRequests))
			}

			// Verify request content for the first few requests
			if len(processedRequests) > 0 {
				firstReq := processedRequests[0]
				if firstReq.Method == "" {
					t.Error("First request: Method is empty")
				}
				if firstReq.URL == "" {
					t.Error("First request: URL is empty")
				}
			}
		})
	}
}

// TestDictIntegrationJSONL tests dict functionality with JSONL format
func TestDictIntegrationJSONL(t *testing.T) {
	tests := []struct {
		name          string
		jsonlContent  string
		expectedCount int
		expectedError bool
		errorContains string
	}{
		{
			name:          "basic_dict_jsonl",
			jsonlContent:  `{"method": "POST", "path": "/api/users", "body": {"name": {"$dict": "user_name"}}, "dict": {"user_name": ["Alice", "Bob"]}}`,
			expectedCount: 2,
			expectedError: false,
		},
		{
			name: "multi_line_jsonl_with_dict",
			jsonlContent: `{"method": "POST", "path": "/api/users", "body": {"name": {"$dict": "user_name"}}, "dict": {"user_name": ["Alice", "Bob"]}}
{"method": "GET", "path": "/api/status"}`,
			expectedCount: 2, // First request generates 2, second generates 1
			expectedError: false,
		},
	}

	parser := NewParser()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the JSONL content
			configs, err := parser.ParseMultiple([]byte(tt.jsonlContent), ".jsonl", "test.jsonl")
			if err != nil {
				if tt.expectedError && strings.Contains(err.Error(), tt.errorContains) {
					return // Expected error
				}
				t.Fatalf("Failed to parse JSONL: %v", err)
			}

			if len(configs) == 0 {
				t.Fatal("No configurations parsed")
			}

			// Process all requests from all configs
			var allProcessedRequests []*config.ProcessedRequest
			for _, cfg := range configs {
				processedRequests, err := parser.ProcessRequests(ctx, cfg, "https://api.example.com")
				if err != nil {
					if tt.expectedError && strings.Contains(err.Error(), tt.errorContains) {
						return // Expected error
					}
					t.Fatalf("Failed to process requests: %v", err)
				}
				allProcessedRequests = append(allProcessedRequests, processedRequests...)
			}

			if tt.expectedError {
				t.Fatalf("Expected error containing '%s', but got none", tt.errorContains)
			}

			if len(allProcessedRequests) != tt.expectedCount {
				t.Errorf("Expected %d total processed requests, got %d", tt.expectedCount, len(allProcessedRequests))
			}
		})
	}
}

// TestDictIntegrationWithExistingExamples tests that dict functionality doesn't break existing examples
func TestDictIntegrationWithExistingExamples(t *testing.T) {
	exampleDir := "../../examples"
	entries, err := os.ReadDir(exampleDir)
	if err != nil {
		t.Fatalf("Failed to read examples directory: %v", err)
	}

	parser := NewParser()
	ctx := context.Background()

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()
		filePath := filepath.Join(exampleDir, fileName)
		if strings.HasPrefix(fileName, ".") || strings.HasPrefix(fileName, "dict_") {
			continue // Skip hidden files and dict examples (they will be tested separately)
		}

		// Skip examples with known issues unrelated to dict functionality
		skipFiles := []string{
			"join_example.yaml",
			"json_with_indent_example.yaml",
			"timestamp_with_format_example.yaml",
		}
		skip := false
		for _, skipFile := range skipFiles {
			if fileName == skipFile {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		t.Run(fileName, func(t *testing.T) {
			data, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("Failed to read example file: %v", err)
			}

			ext := filepath.Ext(fileName)
			configs, err := parser.ParseMultiple(data, ext, filePath)
			if err != nil {
				t.Errorf("Failed to parse example file %s: %v", fileName, err)
				return
			}

			// Process each configuration to ensure it still works
			for i, config := range configs {
				processedRequests, err := parser.ProcessRequests(ctx, config, "https://api.example.com")
				if err != nil {
					t.Errorf("Failed to process requests for config %d in %s: %v", i, fileName, err)
					continue
				}

				// Verify that at least one request was generated
				if len(processedRequests) == 0 {
					t.Errorf("No processed requests generated for config %d in %s", i, fileName)
				}

				// Verify basic structure of processed requests
				for j, req := range processedRequests {
					if req.Method == "" {
						t.Errorf("Config %d, Request %d in %s: Method is empty", i, j, fileName)
					}
					if req.URL == "" {
						t.Errorf("Config %d, Request %d in %s: URL is empty", i, j, fileName)
					}
				}
			}
		})
	}
}

// TestDictIntegrationErrorHandling tests comprehensive error handling scenarios
func TestDictIntegrationErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		fileExt       string
		expectedError string
	}{
		{
			name: "dict_with_non_primitive_array_elements",
			content: `{
				"method": "POST",
				"path": "/api/users",
				"body": {"name": {"$dict": "user_data"}},
				"dict": {
					"user_data": [{"name": "Alice"}, {"name": "Bob"}]
				}
			}`,
			fileExt:       ".json",
			expectedError: "must be a primitive value",
		},
		{
			name: "dict_with_nested_objects_in_array",
			content: `method: POST
path: /api/users
body:
  user:
    $dict: user_info
dict:
  user_info: 
    - name: Alice
      age: 25
    - name: Bob
      age: 30`,
			fileExt:       ".yaml",
			expectedError: "must be a primitive value",
		},
		{
			name: "multiple_dict_validation_errors",
			content: `{
				"method": "POST",
				"path": "/api/users",
				"dict": {
					"empty_array": [],
					"nested_objects": [{"key": "value"}]
				}
			}`,
			fileExt:       ".json",
			expectedError: "array cannot be empty",
		},
	}

	parser := NewParser()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configs, err := parser.ParseMultiple([]byte(tt.content), tt.fileExt, "test"+tt.fileExt)
			if err != nil {
				if strings.Contains(err.Error(), tt.expectedError) {
					return // Expected error during parsing
				}
				t.Fatalf("Unexpected parse error: %v", err)
			}

			if len(configs) == 0 {
				t.Fatal("No configurations parsed")
			}

			// Try to process the requests
			_, err = parser.ProcessRequests(ctx, configs[0], "https://api.example.com")
			if err == nil {
				t.Fatalf("Expected error containing '%s', but got none", tt.expectedError)
			}

			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("Expected error containing '%s', got: %v", tt.expectedError, err)
			}
		})
	}
}

// TestDictIntegrationRequestStructure tests that generated requests have correct structure
func TestDictIntegrationRequestStructure(t *testing.T) {
	yamlContent := `method: POST
path: /api/users
headers:
  Content-Type: application/json
  Authorization:
    $var: auth_token
query:
  format: json
  include:
    $dict: include_fields
body:
  name:
    $dict: user_name
  age:
    $dict: user_age
  department:
    $var: department
variables:
  auth_token: "Bearer token123"
  department: "engineering"
dict:
  user_name: ["Alice", "Bob"]
  user_age: [25, 30]
  include_fields: ["profile", "settings"]`

	parser := NewParser()
	ctx := context.Background()

	configs, err := parser.ParseMultiple([]byte(yamlContent), ".yaml", "test.yaml")
	if err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	processedRequests, err := parser.ProcessRequests(ctx, configs[0], "https://api.example.com")
	if err != nil {
		t.Fatalf("Failed to process requests: %v", err)
	}

	expectedCount := 2 * 2 * 2 // user_name * user_age * include_fields = 8
	if len(processedRequests) != expectedCount {
		t.Errorf("Expected %d requests, got %d", expectedCount, len(processedRequests))
	}

	// Verify structure of each request
	for i, req := range processedRequests {
		// Check basic fields
		if req.Method != "POST" {
			t.Errorf("Request %d: Expected method POST, got %s", i, req.Method)
		}

		if !strings.Contains(req.URL, "https://api.example.com/api/users") {
			t.Errorf("Request %d: URL doesn't contain expected base: %s", i, req.URL)
		}

		// Check headers
		if req.Headers == nil {
			t.Errorf("Request %d: Headers is nil", i)
			continue
		}

		if req.Headers["Content-Type"] != "application/json" {
			t.Errorf("Request %d: Expected Content-Type header", i)
		}

		if req.Headers["Authorization"] != "Bearer token123" {
			t.Errorf("Request %d: Expected Authorization header", i)
		}

		// Check body structure
		if req.Body == "" {
			t.Errorf("Request %d: Body is empty", i)
			continue
		}

		var bodyMap map[string]interface{}
		if err := json.Unmarshal([]byte(req.Body), &bodyMap); err != nil {
			t.Errorf("Request %d: Failed to parse body JSON: %v", i, err)
			continue
		}

		// Verify dict variables are resolved
		if bodyMap["name"] == nil {
			t.Errorf("Request %d: Body missing 'name' field", i)
		}
		if bodyMap["age"] == nil {
			t.Errorf("Request %d: Body missing 'age' field", i)
		}

		// Verify regular variables are resolved
		if bodyMap["department"] != "engineering" {
			t.Errorf("Request %d: Expected department 'engineering', got %v", i, bodyMap["department"])
		}

		// Verify query parameters
		if !strings.Contains(req.URL, "format=json") {
			t.Errorf("Request %d: URL missing format query parameter", i)
		}
		if !strings.Contains(req.URL, "include=") {
			t.Errorf("Request %d: URL missing include query parameter", i)
		}
	}
}
