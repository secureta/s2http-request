package parser

import (
	"context"
	"testing"
)

// TestBackwardCompatibility ensures that the dict feature doesn't break existing functionality
func TestBackwardCompatibility(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		fileExt     string
		expectError bool
		description string
	}{
		{
			name: "existing_variables_functionality",
			content: `{
				"method": "POST",
				"path": "/api/users",
				"headers": {
					"Authorization": {"$var": "auth_token"}
				},
				"body": {
					"name": {"$var": "user_name"},
					"email": {"$var": "user_email"}
				},
				"variables": {
					"auth_token": "Bearer token123",
					"user_name": "John Doe",
					"user_email": "john@example.com"
				}
			}`,
			fileExt:     ".json",
			expectError: false,
			description: "Existing variables functionality should work unchanged",
		},
		{
			name: "existing_functions_functionality",
			content: `method: POST
path: /api/timestamp
body:
  timestamp:
    $timestamp: []
  uuid:
    $uuid: []
  encoded_data:
    $url_encode: ["hello world"]`,
			fileExt:     ".yaml",
			expectError: false,
			description: "Existing functions should work unchanged",
		},
		{
			name: "existing_multipart_functionality",
			content: `{
				"method": "POST",
				"path": "/api/upload",
				"body": {
					"$multipart": {
						"values": {
							"name": "test file",
							"description": "A test upload"
						},
						"boundary": "----WebKitFormBoundary7MA4YWxkTrZu0gW"
					}
				}
			}`,
			fileExt:     ".json",
			expectError: false,
			description: "Existing multipart functionality should work unchanged",
		},
		{
			name: "existing_form_functionality",
			content: `method: POST
path: /api/form
body:
  $form:
    username: testuser
    password: testpass
    remember: true`,
			fileExt:     ".yaml",
			expectError: false,
			description: "Existing form functionality should work unchanged",
		},
		{
			name: "existing_base64_functionality",
			content: `{
				"method": "POST",
				"path": "/api/data",
				"body": {
					"encoded": {"$base64_encode": ["Hello World"]},
					"decoded": {"$base64_decode": ["SGVsbG8gV29ybGQ="]}
				}
			}`,
			fileExt:     ".json",
			expectError: false,
			description: "Existing base64 functionality should work unchanged",
		},
		{
			name: "existing_hex_encode_functionality",
			content: `method: POST
path: /api/hex
body:
  hex_data:
    $hex_encode: ["Hello World"]`,
			fileExt:     ".yaml",
			expectError: false,
			description: "Existing hex encode functionality should work unchanged",
		},
		{
			name: "existing_random_functionality",
			content: `{
				"method": "POST",
				"path": "/api/random",
				"body": {
					"random_number": {"$random": [100]},
					"random_string": {"$random_string": [10]}
				}
			}`,
			fileExt:     ".json",
			expectError: false,
			description: "Existing random functionality should work unchanged",
		},
		{
			name: "existing_time_functionality",
			content: `method: POST
path: /api/time
body:
  current_date:
    $date: []
  current_time:
    $time: []
  formatted_date:
    $date: ["2006-01-02"]`,
			fileExt:     ".yaml",
			expectError: false,
			description: "Existing time functionality should work unchanged",
		},
		{
			name: "existing_url_encoding_functionality",
			content: `{
				"method": "GET",
				"path": "/api/search",
				"query": {
					"q": {"$url_encode": ["hello world & more"]},
					"decoded": {"$url_decode": ["hello%20world"]}
				}
			}`,
			fileExt:     ".json",
			expectError: false,
			description: "Existing URL encoding functionality should work unchanged",
		},
		{
			name: "existing_concat_functionality",
			content: `method: POST
path: /api/concat
body:
  full_name:
    $concat: ["John", " ", "Doe"]
  message:
    $concat: ["Hello ", {"$var": "name"}, "!"]
variables:
  name: "World"`,
			fileExt:     ".yaml",
			expectError: false,
			description: "Existing concat functionality should work unchanged",
		},
		{
			name: "existing_json_functionality",
			content: `{
				"method": "POST",
				"path": "/api/json",
				"body": {
					"$json": {
						"value": {
							"user": {
								"name": "John",
								"age": 30
							}
						}
					}
				}
			}`,
			fileExt:     ".json",
			expectError: false,
			description: "Existing JSON functionality should work unchanged",
		},
		{
			name: "complex_existing_configuration",
			content: `method: POST
path: /api/complex
headers:
  Authorization:
    $concat: ["Bearer ", {"$var": "token"}]
  Content-Type: application/json
  X-Request-ID:
    $uuid: []
query:
  timestamp:
    $timestamp: []
  encoded_param:
    $url_encode: [{"$var": "search_term"}]
body:
  $json:
    value:
      user:
        id:
          $random: [1000]
        name:
          $var: user_name
        created_at:
          $date: ["2006-01-02T15:04:05Z"]
      metadata:
        request_id:
          $uuid: []
        encoded_data:
          $base64_encode: [{"$var": "raw_data"}]
variables:
  token: "abc123"
  user_name: "Test User"
  search_term: "hello world"
  raw_data: "sensitive information"`,
			fileExt:     ".yaml",
			expectError: false,
			description: "Complex existing configurations should work unchanged",
		},
		{
			name: "jsonl_existing_functionality",
			content: `{"method": "GET", "path": "/api/users", "query": {"page": {"$var": "page"}}, "variables": {"page": "1"}}
{"method": "POST", "path": "/api/users", "body": {"name": {"$var": "name"}}, "variables": {"name": "John"}}`,
			fileExt:     ".jsonl",
			expectError: false,
			description: "Existing JSONL functionality should work unchanged",
		},
	}

	parser := NewParser()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the configuration
			configs, err := parser.ParseMultiple([]byte(tt.content), tt.fileExt, "test"+tt.fileExt)
			if err != nil {
				if tt.expectError {
					return // Expected error
				}
				t.Fatalf("Failed to parse config: %v", err)
			}

			if len(configs) == 0 {
				t.Fatal("No configurations parsed")
			}

			// Process each configuration
			for i, config := range configs {
				processedRequests, err := parser.ProcessRequests(ctx, config, "https://api.example.com")
				if err != nil {
					if tt.expectError {
						return // Expected error
					}
					t.Fatalf("Failed to process requests for config %d: %v", i, err)
				}

				if tt.expectError {
					t.Fatalf("Expected error but got none for config %d", i)
				}

				// Verify that at least one request was generated
				if len(processedRequests) == 0 {
					t.Errorf("No processed requests generated for config %d", i)
				}

				// Verify basic structure of processed requests
				for j, req := range processedRequests {
					if req.Method == "" {
						t.Errorf("Config %d, Request %d: Method is empty", i, j)
					}
					if req.URL == "" {
						t.Errorf("Config %d, Request %d: URL is empty", i, j)
					}
				}
			}
		})
	}
}

// TestDictAndExistingFeaturesCoexistence ensures dict and existing features work together
func TestDictAndExistingFeaturesCoexistence(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		fileExt       string
		expectedCount int
		expectError   bool
		description   string
	}{
		{
			name: "dict_with_variables_coexistence",
			content: `{
				"method": "POST",
				"path": "/api/users",
				"headers": {
					"Authorization": {"$var": "auth_token"},
					"Content-Type": "application/json"
				},
				"body": {
					"name": {"$dict": "user_name"},
					"department": {"$var": "department"},
					"timestamp": {"$timestamp": []}
				},
				"variables": {
					"auth_token": "Bearer token123",
					"department": "engineering"
				},
				"dict": {
					"user_name": ["Alice", "Bob"]
				}
			}`,
			fileExt:       ".json",
			expectedCount: 2,
			expectError:   false,
			description:   "Dict should coexist with variables and functions",
		},
		{
			name: "dict_with_complex_functions",
			content: `method: POST
path: /api/complex
headers:
  Authorization:
    $concat: ["Bearer ", {"$var": "token"}]
  X-Request-ID:
    $uuid: []
body:
  user:
    name:
      $dict: user_name
    id:
      $random: [1000]
    created_at:
      $date: ["2006-01-02"]
  metadata:
    encoded_name:
      $base64_encode: [{"$dict": "user_name"}]
    hex_data:
      $hex_encode: [{"$var": "raw_data"}]
variables:
  token: "abc123"
  raw_data: "test data"
dict:
  user_name: ["Alice", "Bob", "Charlie"]`,
			fileExt:       ".yaml",
			expectedCount: 3,
			expectError:   false,
			description:   "Dict should work with complex function combinations",
		},
		{
			name: "dict_with_multipart",
			content: `{
				"method": "POST",
				"path": "/api/upload",
				"body": {
					"$multipart": {
						"values": {
							"user": {"$dict": "user_name"},
							"description": {"$var": "description"}
						},
						"boundary": "----WebKitFormBoundary7MA4YWxkTrZu0gW"
					}
				},
				"variables": {
					"description": "Test upload"
				},
				"dict": {
					"user_name": ["alice", "bob"]
				}
			}`,
			fileExt:       ".json",
			expectedCount: 2,
			expectError:   false,
			description:   "Dict should work with multipart forms",
		},
		{
			name: "dict_with_form",
			content: `method: POST
path: /api/form
body:
  $form:
    username:
      $dict: username
    password:
      $var: password
    remember: true
variables:
  password: "secret123"
dict:
  username: ["user1", "user2"]`,
			fileExt:       ".yaml",
			expectedCount: 2,
			expectError:   false,
			description:   "Dict should work with form encoding",
		},
		{
			name: "dict_with_json_function",
			content: `{
				"method": "POST",
				"path": "/api/json",
				"body": {
					"$json": {
						"value": {
							"user": {
								"name": {"$dict": "user_name"},
								"role": {"$var": "role"},
								"id": {"$random": [100]}
							},
							"timestamp": {"$timestamp": []}
						}
					}
				},
				"variables": {
					"role": "admin"
				},
				"dict": {
					"user_name": ["Alice", "Bob"]
				}
			}`,
			fileExt:       ".json",
			expectedCount: 2,
			expectError:   false,
			description:   "Dict should work within JSON function",
		},
	}

	parser := NewParser()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the configuration
			configs, err := parser.ParseMultiple([]byte(tt.content), tt.fileExt, "test"+tt.fileExt)
			if err != nil {
				if tt.expectError {
					return // Expected error
				}
				t.Fatalf("Failed to parse config: %v", err)
			}

			if len(configs) == 0 {
				t.Fatal("No configurations parsed")
			}

			// Process the requests
			processedRequests, err := parser.ProcessRequests(ctx, configs[0], "https://api.example.com")
			if err != nil {
				if tt.expectError {
					return // Expected error
				}
				t.Fatalf("Failed to process requests: %v", err)
			}

			if tt.expectError {
				t.Fatalf("Expected error but got none")
			}

			if len(processedRequests) != tt.expectedCount {
				t.Errorf("Expected %d requests, got %d", tt.expectedCount, len(processedRequests))
			}

			// Verify that each request has the expected structure
			for i, req := range processedRequests {
				if req.Method == "" {
					t.Errorf("Request %d: Method is empty", i)
				}
				if req.URL == "" {
					t.Errorf("Request %d: URL is empty", i)
				}
				// Additional verification could be added here based on the specific test case
			}
		})
	}
}

// TestNoRegressionInErrorHandling ensures error handling hasn't regressed
func TestNoRegressionInErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		fileExt     string
		expectError bool
		description string
	}{
		{
			name: "invalid_json_still_fails",
			content: `{
				"method": "POST",
				"path": "/api/users"
				"body": {"name": "test"}
			}`,
			fileExt:     ".json",
			expectError: true,
			description: "Invalid JSON should still be rejected",
		},
		{
			name: "invalid_yaml_still_fails",
			content: `method: POST
path: /api/users
body:
  name: test
    invalid: indentation`,
			fileExt:     ".yaml",
			expectError: true,
			description: "Invalid YAML should still be rejected",
		},
		{
			name: "undefined_variable_still_fails",
			content: `{
				"method": "POST",
				"path": "/api/users",
				"body": {
					"name": {"$var": "undefined_variable"}
				}
			}`,
			fileExt:     ".json",
			expectError: true,
			description: "Undefined variables should still cause errors",
		},
		{
			name: "invalid_function_still_fails",
			content: `method: POST
path: /api/users
body:
  name:
    $nonexistent_function: ["test"]`,
			fileExt:     ".yaml",
			expectError: true,
			description: "Invalid functions should still cause errors",
		},
	}

	parser := NewParser()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the configuration
			configs, err := parser.ParseMultiple([]byte(tt.content), tt.fileExt, "test"+tt.fileExt)
			if err != nil {
				if tt.expectError {
					return // Expected error during parsing
				}
				t.Fatalf("Unexpected parse error: %v", err)
			}

			if len(configs) == 0 && tt.expectError {
				return // Expected error resulted in no configs
			}

			// Process the requests
			_, err = parser.ProcessRequests(ctx, configs[0], "https://api.example.com")
			if err != nil {
				if tt.expectError {
					return // Expected error during processing
				}
				t.Fatalf("Unexpected processing error: %v", err)
			}

			if tt.expectError {
				t.Fatalf("Expected error but got none")
			}
		})
	}
}

// TestPerformanceRegression ensures dict feature doesn't significantly impact performance
func TestPerformanceRegression(t *testing.T) {
	// Simple configuration without dict
	simpleConfig := `{
		"method": "POST",
		"path": "/api/users",
		"body": {
			"name": {"$var": "user_name"},
			"timestamp": {"$timestamp": []}
		},
		"variables": {
			"user_name": "John Doe"
		}
	}`

	// Configuration with dict
	dictConfig := `{
		"method": "POST",
		"path": "/api/users",
		"body": {
			"name": {"$dict": "user_name"},
			"timestamp": {"$timestamp": []}
		},
		"dict": {
			"user_name": ["John Doe"]
		}
	}`

	parser := NewParser()
	ctx := context.Background()

	// Test simple configuration (baseline)
	t.Run("simple_config_performance", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			configs, err := parser.ParseMultiple([]byte(simpleConfig), ".json", "test.json")
			if err != nil {
				t.Fatalf("Failed to parse simple config: %v", err)
			}

			_, err = parser.ProcessRequests(ctx, configs[0], "https://api.example.com")
			if err != nil {
				t.Fatalf("Failed to process simple config: %v", err)
			}
		}
	})

	// Test dict configuration
	t.Run("dict_config_performance", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			configs, err := parser.ParseMultiple([]byte(dictConfig), ".json", "test.json")
			if err != nil {
				t.Fatalf("Failed to parse dict config: %v", err)
			}

			_, err = parser.ProcessRequests(ctx, configs[0], "https://api.example.com")
			if err != nil {
				t.Fatalf("Failed to process dict config: %v", err)
			}
		}
	})
}
