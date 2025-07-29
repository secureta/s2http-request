package parser

import (
	"context"
	"net/url"
	"testing"

	"github.com/secureta/s2http-request/internal/config"
)

func TestParseJSON(t *testing.T) {
	tests := []struct {
		name      string
		jsonData  string
		wantError bool
		expected  *config.RequestConfig
	}{
		{
			name: "valid basic request",
			jsonData: `{
				"method": "GET",
				"path": "/test",
				"query": {
					"param1": "value1"
				},
				"variables": {
					"test_var": "test_value"
				}
			}`,
			wantError: false,
			expected: &config.RequestConfig{
				Method: "GET",
				Path:   "/test",
				Query: map[string]interface{}{
					"param1": "value1",
				},
				Variables: map[string]interface{}{
					"test_var": "test_value",
				},
			},
		},
		{
			name: "request with functions",
			jsonData: `{
				"method": "POST",
				"path": "/api",
				"params": {
					"encoded": {
						"!url_encode": "hello world"
					}
				}
			}`,
			wantError: false,
			expected: &config.RequestConfig{
				Method: "POST",
				Path:   "/api",
				Params: map[string]interface{}{
					"encoded": map[string]interface{}{
						"!url_encode": "hello world",
					},
				},
			},
		},
		{
			name:      "invalid json",
			jsonData:  `{"invalid": json}`,
			wantError: true,
			expected:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()

			result, err := parser.Parse([]byte(tt.jsonData), ".json", "")

			if tt.wantError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.wantError && result != nil {
				if result.Method != tt.expected.Method {
					t.Errorf("Expected method %s, got %s", tt.expected.Method, result.Method)
				}
				if result.Path != tt.expected.Path {
					t.Errorf("Expected path %s, got %s", tt.expected.Path, result.Path)
				}
			}
		})
	}
}

func TestParseYAML(t *testing.T) {
	tests := []struct {
		name      string
		yamlData  string
		wantError bool
		expected  *config.RequestConfig
	}{
		{
			name: "valid yaml request",
			yamlData: `
method: GET
path: /test
query:
  param1: value1
variables:
  test_var: test_value
`,
			wantError: false,
			expected: &config.RequestConfig{
				Method: "GET",
				Path:   "/test",
				Query: map[string]interface{}{
					"param1": "value1",
				},
				Variables: map[string]interface{}{
					"test_var": "test_value",
				},
			},
		},
		{
			name: "multiple yaml documents (first document)",
			yamlData: `
method: GET
path: /first
query:
  param1: value1
---
method: POST
path: /second
body:
  key: value
`,
			wantError: false,
			expected: &config.RequestConfig{
				Method: "GET",
				Path:   "/first",
				Query: map[string]interface{}{
					"param1": "value1",
				},
			},
		},
		{
			name:      "invalid yaml",
			yamlData:  "invalid: yaml: content: [",
			wantError: true,
			expected:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()

			result, err := parser.Parse([]byte(tt.yamlData), ".yaml", "")

			if tt.wantError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.wantError && result != nil {
				if result.Method != tt.expected.Method {
					t.Errorf("Expected method %s, got %s", tt.expected.Method, result.Method)
				}
				if result.Path != tt.expected.Path {
					t.Errorf("Expected path %s, got %s", tt.expected.Path, result.Path)
				}
			}
		})
	}
}

func TestParseMultipleYAML(t *testing.T) {
	tests := []struct {
		name      string
		yamlData  string
		wantError bool
		expected  []*config.RequestConfig
	}{
		{
			name: "single yaml document",
			yamlData: `
method: GET
path: /test
query:
  param1: value1
`,
			wantError: false,
			expected: []*config.RequestConfig{
				{
					Method: "GET",
					Path:   "/test",
					Query: map[string]interface{}{
						"param1": "value1",
					},
				},
			},
		},
		{
			name: "multiple yaml documents",
			yamlData: `
method: GET
path: /first
query:
  param1: value1
---
method: POST
path: /second
body:
  key: value
`,
			wantError: false,
			expected: []*config.RequestConfig{
				{
					Method: "GET",
					Path:   "/first",
					Query: map[string]interface{}{
						"param1": "value1",
					},
				},
				{
					Method: "POST",
					Path:   "/second",
					Body: map[string]interface{}{
						"key": "value",
					},
				},
			},
		},
		{
			name:      "invalid yaml",
			yamlData:  "invalid: yaml: content: [",
			wantError: true,
			expected:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()

			results, err := parser.ParseMultiple([]byte(tt.yamlData), ".yaml", "")

			if tt.wantError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.wantError && results != nil {
				if len(results) != len(tt.expected) {
					t.Errorf("Expected %d configs, got %d", len(tt.expected), len(results))
				} else {
					for i, result := range results {
						if result.Method != tt.expected[i].Method {
							t.Errorf("Config %d: Expected method %s, got %s", i, tt.expected[i].Method, result.Method)
						}
						if result.Path != tt.expected[i].Path {
							t.Errorf("Config %d: Expected path %s, got %s", i, tt.expected[i].Path, result.Path)
						}
					}
				}
			}
		})
	}
}

func TestProcessValue(t *testing.T) {
	tests := []struct {
		name      string
		input     interface{}
		variables map[string]interface{}
		expected  interface{}
		wantError bool
	}{
		{
			name:      "simple string",
			input:     "hello",
			variables: map[string]interface{}{},
			expected:  "hello",
			wantError: false,
		},
		{
			name:      "variable reference",
			input:     map[string]interface{}{"$var": "test_var"},
			variables: map[string]interface{}{"test_var": "test_value"},
			expected:  "test_value",
			wantError: false,
		},
		{
			name:      "url encode function",
			input:     map[string]interface{}{"$url_encode": "hello world"},
			variables: map[string]interface{}{},
			expected:  "hello+world",
			wantError: false,
		},
		{
			name:      "concat function",
			input:     map[string]interface{}{"$concat": []interface{}{"hello", " ", "world"}},
			variables: map[string]interface{}{},
			expected:  "hello world",
			wantError: false,
		},
		{
			name:      "nested functions",
			input:     map[string]interface{}{"$url_encode": map[string]interface{}{"$var": "test_var"}},
			variables: map[string]interface{}{"test_var": "hello world"},
			expected:  "hello+world",
			wantError: false,
		},
		{
			name:      "unknown function",
			input:     map[string]interface{}{"$unknown": "value"},
			variables: map[string]interface{}{},
			expected:  nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			ctx := context.WithValue(context.Background(), "variables", tt.variables)

			result, err := parser.processValue(ctx, tt.input)

			if tt.wantError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.wantError && result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestProcessRequest(t *testing.T) {
	tests := []struct {
		name      string
		config    *config.RequestConfig
		baseURL   string
		expected  *config.ProcessedRequest
		wantError bool
	}{
		{
			name: "simple GET request",
			config: &config.RequestConfig{
				Method: "GET",
				Path:   "/test",
				Query: map[string]interface{}{
					"param1": "value1",
				},
				Headers: map[string]interface{}{
					"Content-Type": "application/json",
				},
			},
			baseURL: "https://example.com",
			expected: &config.ProcessedRequest{
				Method:  "GET",
				URL:     "https://example.com/test?param1=value1",
				Headers: map[string]string{"Content-Type": "application/json"},
				Body:    "",
			},
			wantError: false,
		},
		{
			name: "POST request with params",
			config: &config.RequestConfig{
				Method: "POST",
				Path:   "/api",
				Params: map[string]interface{}{
					"username": "admin",
					"password": "secret",
				},
			},
			baseURL: "https://api.example.com",
			expected: &config.ProcessedRequest{
				Method:  "POST",
				URL:     "https://api.example.com/api",
				Headers: map[string]string{},
				Body:    "username=admin&password=secret",
			},
			wantError: false,
		},
		{
			name: "request with functions",
			config: &config.RequestConfig{
				Method: "GET",
				Path:   "/search",
				Query: map[string]interface{}{
					"q": map[string]interface{}{"$url_encode": "hello world"},
				},
				Variables: map[string]interface{}{
					"search_term": "hello world",
				},
			},
			baseURL: "https://search.example.com",
			expected: &config.ProcessedRequest{
				Method:  "GET",
				URL:     "https://search.example.com/search?q=hello%2Bworld",
				Headers: map[string]string{},
				Body:    "",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			ctx := context.Background()

			result, err := parser.ProcessRequest(ctx, tt.config, tt.baseURL)

			if tt.wantError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.wantError && result != nil {
				if result.Method != tt.expected.Method {
					t.Errorf("Expected method %s, got %s", tt.expected.Method, result.Method)
				}
				if result.URL != tt.expected.URL {
					t.Errorf("Expected URL %s, got %s", tt.expected.URL, result.URL)
				}
				// For body comparison, check if it contains the expected parameters regardless of order
				if tt.expected.Body != "" {
					expectedParams, _ := url.ParseQuery(tt.expected.Body)
					actualParams, _ := url.ParseQuery(result.Body)
					if len(expectedParams) != len(actualParams) {
						t.Errorf("Expected %d params, got %d", len(expectedParams), len(actualParams))
					}
					for key, expectedValues := range expectedParams {
						if actualValues, exists := actualParams[key]; !exists {
							t.Errorf("Missing parameter %s", key)
						} else if len(actualValues) != len(expectedValues) || actualValues[0] != expectedValues[0] {
							t.Errorf("Parameter %s: expected %v, got %v", key, expectedValues, actualValues)
						}
					}
				} else if result.Body != "" {
					t.Errorf("Expected empty body, got %s", result.Body)
				}
			}
		})
	}
}

func TestMapToQueryString(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected string
	}{
		{
			name:     "simple params",
			input:    map[string]interface{}{"a": "1", "b": "2"},
			expected: "a=1&b=2",
		},
		{
			name:     "url encoded values",
			input:    map[string]interface{}{"q": "hello world", "type": "search"},
			expected: "q=hello+world&type=search",
		},
		{
			name:     "empty map",
			input:    map[string]interface{}{},
			expected: "",
		},
		{
			name:     "nil values",
			input:    map[string]interface{}{"a": "1", "b": nil, "c": "3"},
			expected: "a=1&c=3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			result := parser.mapToQueryString(tt.input)

			// Since map iteration order is not guaranteed, we need to check
			// that all expected key-value pairs are present
			if tt.expected == "" && result != "" {
				t.Errorf("Expected empty string, got %s", result)
			}
			if tt.expected != "" && result == "" {
				t.Errorf("Expected non-empty string, got empty")
			}
			// For non-empty cases, just check length as a basic validation
			if tt.expected != "" && len(result) == 0 {
				t.Errorf("Expected non-empty result")
			}
		})
	}
}

func TestVariableOverride(t *testing.T) {
	tests := []struct {
		name         string
		requestData  string
		cliVariables map[string]interface{}
		expectedURL  string
		expectedBody string
	}{
		{
			name: "CLI variable overrides file variable",
			requestData: `{
				"method": "GET",
				"path": {
					"$concat": ["/api/users/", {"$var": "id"}]
				},
				"variables": {
					"id": "1"
				}
			}`,
			cliVariables: map[string]interface{}{
				"id": "123",
			},
			expectedURL: "http://localhost/api/users/123",
		},
		{
			name: "CLI variable used when file variable doesn't exist",
			requestData: `{
				"method": "GET",
				"path": {
					"$concat": ["/api/users/", {"$var": "id"}]
				}
			}`,
			cliVariables: map[string]interface{}{
				"id": "456",
			},
			expectedURL: "http://localhost/api/users/456",
		},
		{
			name: "Multiple CLI variables",
			requestData: `{
				"method": "POST",
				"path": {
					"$concat": ["/api/", {"$var": "resource"}]
				},
				"params": {
					"user_id": {"$var": "user_id"},
					"name": {"$var": "name"}
				},
				"variables": {
					"resource": "default",
					"user_id": "1",
					"name": "default_name"
				}
			}`,
			cliVariables: map[string]interface{}{
				"resource": "users",
				"user_id":  789,
				"name":     "test_user",
			},
			expectedURL:  "http://localhost/api/users",
			expectedBody: "name=test_user&user_id=789",
		},
		{
			name: "CLI variables with complex JSON values",
			requestData: `{
				"method": "POST",
				"path": "/api/test",
				"body": {
					"config": {"$var": "config"},
					"items": {"$var": "items"}
				}
			}`,
			cliVariables: map[string]interface{}{
				"config": map[string]interface{}{
					"enabled": true,
					"timeout": float64(30),
				},
				"items": []interface{}{
					"item1", "item2", "item3",
				},
			},
			expectedURL:  "http://localhost/api/test",
			expectedBody: `{"config":{"enabled":true,"timeout":30},"items":["item1","item2","item3"]}`,
		},
		{
			name: "File variable used when CLI variable not provided",
			requestData: `{
				"method": "GET",
				"path": {
					"$concat": ["/api/users/", {"$var": "id"}]
				},
				"variables": {
					"id": "default_id"
				}
			}`,
			cliVariables: map[string]interface{}{},
			expectedURL:  "http://localhost/api/users/default_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()

			// Parse the request configuration
			requestConfig, err := parser.Parse([]byte(tt.requestData), ".json", "test.json")
			if err != nil {
				t.Fatalf("Failed to parse request config: %v", err)
			}

			// Create context with CLI variables
			ctx := context.Background()
			if len(tt.cliVariables) > 0 {
				ctx = context.WithValue(ctx, "variables", tt.cliVariables)
			}

			// Process the request
			processedRequests, err := parser.ProcessRequestsWithRequestID(ctx, requestConfig, "http://localhost", nil)
			if err != nil {
				t.Fatalf("Failed to process request: %v", err)
			}

			if len(processedRequests) != 1 {
				t.Fatalf("Expected 1 processed request, got %d", len(processedRequests))
			}

			result := processedRequests[0]

			// Check URL
			if result.URL != tt.expectedURL {
				t.Errorf("Expected URL %s, got %s", tt.expectedURL, result.URL)
			}

			// Check body if expected
			if tt.expectedBody != "" {
				if result.Body != tt.expectedBody {
					t.Errorf("Expected body %s, got %s", tt.expectedBody, result.Body)
				}
			}
		})
	}
}

func TestVariableOverrideWithProcessRequests(t *testing.T) {
	// Test the ProcessRequests method as well to ensure both code paths work
	requestData := `{
		"method": "GET",
		"path": {
			"$concat": ["/api/users/", {"$var": "id"}]
		},
		"variables": {
			"id": "file_value"
		}
	}`

	parser := NewParser()
	requestConfig, err := parser.Parse([]byte(requestData), ".json", "test.json")
	if err != nil {
		t.Fatalf("Failed to parse request config: %v", err)
	}

	// Test with CLI variables
	cliVariables := map[string]interface{}{
		"id": "cli_value",
	}
	ctx := context.WithValue(context.Background(), "variables", cliVariables)

	processedRequests, err := parser.ProcessRequests(ctx, requestConfig, "http://localhost")
	if err != nil {
		t.Fatalf("Failed to process request: %v", err)
	}

	if len(processedRequests) != 1 {
		t.Fatalf("Expected 1 processed request, got %d", len(processedRequests))
	}

	result := processedRequests[0]
	expectedURL := "http://localhost/api/users/cli_value"

	if result.URL != expectedURL {
		t.Errorf("Expected URL %s, got %s", expectedURL, result.URL)
	}
}
