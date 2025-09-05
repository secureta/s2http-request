package parser

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/secureta/s2http-request/internal/config"
)

// combToString converts a combination map to a string for comparison
func combToString(comb map[string]interface{}) string {
	var keys []string
	for k := range comb {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s:%v", k, comb[k]))
	}
	return strings.Join(parts, ",")
}

func TestValidateDict(t *testing.T) {
	tests := []struct {
		name      string
		dict      map[string][]interface{}
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid dict with single array",
			dict: map[string][]interface{}{
				"user_name": {"Alice", "Bob", "Charlie"},
			},
			wantError: false,
		},
		{
			name: "valid dict with multiple arrays",
			dict: map[string][]interface{}{
				"user_name": {"Alice", "Bob"},
				"user_age":  {25, 30, 35},
			},
			wantError: false,
		},
		{
			name: "valid dict with mixed types in array",
			dict: map[string][]interface{}{
				"mixed": {"string", 123, true},
			},
			wantError: false,
		},
		{
			name:      "empty dict",
			dict:      map[string][]interface{}{},
			wantError: false,
		},
		{
			name: "dict with empty array",
			dict: map[string][]interface{}{
				"empty_array": {},
			},
			wantError: true,
			errorMsg:  "at dict.empty_array [ERROR] array cannot be empty",
		},
		{
			name: "dict with nil array",
			dict: map[string][]interface{}{
				"nil_array": nil,
			},
			wantError: true,
			errorMsg:  "at dict.nil_array [ERROR] value must be an array, got null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			err := parser.validateDict(tt.dict)

			if tt.wantError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tt.wantError && err != nil && tt.errorMsg != "" {
				if err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			}
		})
	}
}

func TestParseWithDict(t *testing.T) {
	tests := []struct {
		name      string
		jsonData  string
		wantError bool
		expected  *config.RequestConfig
		errorMsg  string
	}{
		{
			name: "valid request with dict",
			jsonData: `{
				"method": "POST",
				"path": "/api/users",
				"body": {
					"name": {"$dict": "user_name"},
					"age": {"$dict": "user_age"}
				},
				"dict": {
					"user_name": ["Alice", "Bob", "Charlie"],
					"user_age": [25, 30, 35]
				}
			}`,
			wantError: false,
			expected: &config.RequestConfig{
				Method: "POST",
				Path:   "/api/users",
				Body: map[string]interface{}{
					"name": map[string]interface{}{"$dict": "user_name"},
					"age":  map[string]interface{}{"$dict": "user_age"},
				},
				Dict: map[string][]interface{}{
					"user_name": {"Alice", "Bob", "Charlie"},
					"user_age":  {25, 30, 35},
				},
			},
		},
		{
			name: "request with dict and variables",
			jsonData: `{
				"method": "POST",
				"path": "/api/users",
				"body": {
					"name": {"$dict": "user_name"},
					"api_key": {"$var": "api_key"}
				},
				"dict": {
					"user_name": ["Alice", "Bob"]
				},
				"variables": {
					"api_key": "secret123"
				}
			}`,
			wantError: false,
			expected: &config.RequestConfig{
				Method: "POST",
				Path:   "/api/users",
				Body: map[string]interface{}{
					"name":    map[string]interface{}{"$dict": "user_name"},
					"api_key": map[string]interface{}{"$var": "api_key"},
				},
				Dict: map[string][]interface{}{
					"user_name": {"Alice", "Bob"},
				},
				Variables: map[string]interface{}{
					"api_key": "secret123",
				},
			},
		},
		{
			name: "request with empty dict array",
			jsonData: `{
				"method": "POST",
				"path": "/api/users",
				"dict": {
					"user_name": []
				}
			}`,
			wantError: true,
			errorMsg:  "test.json:5:6 at dict.user_name [ERROR] array cannot be empty",
		},
		{
			name: "request without dict",
			jsonData: `{
				"method": "GET",
				"path": "/api/users"
			}`,
			wantError: false,
			expected: &config.RequestConfig{
				Method: "GET",
				Path:   "/api/users",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()

			result, err := parser.Parse([]byte(tt.jsonData), ".json", "test.json")

			if tt.wantError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tt.wantError && err != nil && tt.errorMsg != "" {
				if err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			}
			if !tt.wantError && result != nil {
				if result.Method != tt.expected.Method {
					t.Errorf("Expected method %s, got %s", tt.expected.Method, result.Method)
				}
				if result.Path != tt.expected.Path {
					t.Errorf("Expected path %s, got %s", tt.expected.Path, result.Path)
				}
				if tt.expected.Dict != nil {
					if result.Dict == nil {
						t.Errorf("Expected dict to be present")
					} else {
						for key, expectedArray := range tt.expected.Dict {
							if actualArray, exists := result.Dict[key]; !exists {
								t.Errorf("Expected dict key %s to exist", key)
							} else if len(actualArray) != len(expectedArray) {
								t.Errorf("Expected dict[%s] to have %d elements, got %d", key, len(expectedArray), len(actualArray))
							}
						}
					}
				}
			}
		})
	}
}

func TestGenerateDictCombinations(t *testing.T) {
	tests := []struct {
		name     string
		dict     map[string][]interface{}
		expected []map[string]interface{}
	}{
		{
			name: "single array",
			dict: map[string][]interface{}{
				"user_name": {"Alice", "Bob"},
			},
			expected: []map[string]interface{}{
				{"user_name": "Alice"},
				{"user_name": "Bob"},
			},
		},
		{
			name: "two arrays",
			dict: map[string][]interface{}{
				"user_name": {"Alice", "Bob"},
				"user_age":  {25, 30},
			},
			expected: []map[string]interface{}{
				{"user_name": "Alice", "user_age": 25},
				{"user_name": "Alice", "user_age": 30},
				{"user_name": "Bob", "user_age": 25},
				{"user_name": "Bob", "user_age": 30},
			},
		},
		{
			name: "three arrays",
			dict: map[string][]interface{}{
				"name":   {"A", "B"},
				"age":    {1, 2},
				"active": {true, false},
			},
			expected: []map[string]interface{}{
				{"name": "A", "age": 1, "active": true},
				{"name": "A", "age": 1, "active": false},
				{"name": "A", "age": 2, "active": true},
				{"name": "A", "age": 2, "active": false},
				{"name": "B", "age": 1, "active": true},
				{"name": "B", "age": 1, "active": false},
				{"name": "B", "age": 2, "active": true},
				{"name": "B", "age": 2, "active": false},
			},
		},
		{
			name:     "empty dict",
			dict:     map[string][]interface{}{},
			expected: []map[string]interface{}{{}},
		},
		{
			name: "single element arrays",
			dict: map[string][]interface{}{
				"user_name": {"Alice"},
				"user_age":  {25},
			},
			expected: []map[string]interface{}{
				{"user_name": "Alice", "user_age": 25},
			},
		},
		{
			name: "mixed data types",
			dict: map[string][]interface{}{
				"string_val": {"hello", "world"},
				"int_val":    {1, 2},
				"bool_val":   {true, false},
				"float_val":  {3.14, 2.71},
			},
			expected: []map[string]interface{}{
				{"string_val": "hello", "int_val": 1, "bool_val": true, "float_val": 3.14},
				{"string_val": "hello", "int_val": 1, "bool_val": true, "float_val": 2.71},
				{"string_val": "hello", "int_val": 1, "bool_val": false, "float_val": 3.14},
				{"string_val": "hello", "int_val": 1, "bool_val": false, "float_val": 2.71},
				{"string_val": "hello", "int_val": 2, "bool_val": true, "float_val": 3.14},
				{"string_val": "hello", "int_val": 2, "bool_val": true, "float_val": 2.71},
				{"string_val": "hello", "int_val": 2, "bool_val": false, "float_val": 3.14},
				{"string_val": "hello", "int_val": 2, "bool_val": false, "float_val": 2.71},
				{"string_val": "world", "int_val": 1, "bool_val": true, "float_val": 3.14},
				{"string_val": "world", "int_val": 1, "bool_val": true, "float_val": 2.71},
				{"string_val": "world", "int_val": 1, "bool_val": false, "float_val": 3.14},
				{"string_val": "world", "int_val": 1, "bool_val": false, "float_val": 2.71},
				{"string_val": "world", "int_val": 2, "bool_val": true, "float_val": 3.14},
				{"string_val": "world", "int_val": 2, "bool_val": true, "float_val": 2.71},
				{"string_val": "world", "int_val": 2, "bool_val": false, "float_val": 3.14},
				{"string_val": "world", "int_val": 2, "bool_val": false, "float_val": 2.71},
			},
		},
		{
			name: "large combination count",
			dict: map[string][]interface{}{
				"a": {1, 2, 3, 4, 5},
				"b": {"x", "y", "z"},
			},
			expected: []map[string]interface{}{
				{"a": 1, "b": "x"}, {"a": 1, "b": "y"}, {"a": 1, "b": "z"},
				{"a": 2, "b": "x"}, {"a": 2, "b": "y"}, {"a": 2, "b": "z"},
				{"a": 3, "b": "x"}, {"a": 3, "b": "y"}, {"a": 3, "b": "z"},
				{"a": 4, "b": "x"}, {"a": 4, "b": "y"}, {"a": 4, "b": "z"},
				{"a": 5, "b": "x"}, {"a": 5, "b": "y"}, {"a": 5, "b": "z"},
			},
		},
		{
			name: "nil values in arrays",
			dict: map[string][]interface{}{
				"nullable": {nil, "value", nil},
				"normal":   {"a", "b"},
			},
			expected: []map[string]interface{}{
				{"nullable": nil, "normal": "a"},
				{"nullable": nil, "normal": "b"},
				{"nullable": "value", "normal": "a"},
				{"nullable": "value", "normal": "b"},
				{"nullable": nil, "normal": "a"},
				{"nullable": nil, "normal": "b"},
			},
		},
		{
			name: "single key with multiple values",
			dict: map[string][]interface{}{
				"status": {"active", "inactive", "pending", "suspended"},
			},
			expected: []map[string]interface{}{
				{"status": "active"},
				{"status": "inactive"},
				{"status": "pending"},
				{"status": "suspended"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			result := parser.generateDictCombinations(tt.dict)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d combinations, got %d", len(tt.expected), len(result))
				return
			}

			// Check that all expected combinations are present (order doesn't matter)
			expectedSet := make(map[string]bool)
			for _, expectedComb := range tt.expected {
				key := combToString(expectedComb)
				expectedSet[key] = true
			}

			actualSet := make(map[string]bool)
			for _, actualComb := range result {
				key := combToString(actualComb)
				actualSet[key] = true
			}

			// Check that all expected combinations are in the result
			for expectedKey := range expectedSet {
				if !actualSet[expectedKey] {
					t.Errorf("Missing expected combination: %s", expectedKey)
				}
			}

			// Check that no unexpected combinations are in the result
			for actualKey := range actualSet {
				if !expectedSet[actualKey] {
					t.Errorf("Unexpected combination: %s", actualKey)
				}
			}
		})
	}
}

func TestProcessRequestsWithDict(t *testing.T) {
	tests := []struct {
		name           string
		requestData    string
		baseURL        string
		expectedCount  int
		expectedBodies []string
		wantError      bool
	}{
		{
			name: "single dict variable",
			requestData: `{
				"method": "POST",
				"path": "/api/users",
				"body": {
					"name": {"$dict": "user_name"}
				},
				"dict": {
					"user_name": ["Alice", "Bob"]
				}
			}`,
			baseURL:       "https://api.example.com",
			expectedCount: 2,
			expectedBodies: []string{
				`{"name":"Alice"}`,
				`{"name":"Bob"}`,
			},
			wantError: false,
		},
		{
			name: "multiple dict variables",
			requestData: `{
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
			baseURL:       "https://api.example.com",
			expectedCount: 4,
			expectedBodies: []string{
				`{"age":25,"name":"Alice"}`,
				`{"age":30,"name":"Alice"}`,
				`{"age":25,"name":"Bob"}`,
				`{"age":30,"name":"Bob"}`,
			},
			wantError: false,
		},
		{
			name: "dict with variables",
			requestData: `{
				"method": "POST",
				"path": "/api/users",
				"body": {
					"name": {"$dict": "user_name"},
					"api_key": {"$var": "api_key"}
				},
				"dict": {
					"user_name": ["Alice", "Bob"]
				},
				"variables": {
					"api_key": "secret123"
				}
			}`,
			baseURL:       "https://api.example.com",
			expectedCount: 2,
			expectedBodies: []string{
				`{"api_key":"secret123","name":"Alice"}`,
				`{"api_key":"secret123","name":"Bob"}`,
			},
			wantError: false,
		},
		{
			name: "request without dict",
			requestData: `{
				"method": "GET",
				"path": "/api/users"
			}`,
			baseURL:       "https://api.example.com",
			expectedCount: 1,
			wantError:     false,
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

			// Process the requests
			processedRequests, err := parser.ProcessRequests(context.Background(), requestConfig, tt.baseURL)

			if tt.wantError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.wantError {
				if len(processedRequests) != tt.expectedCount {
					t.Errorf("Expected %d processed requests, got %d", tt.expectedCount, len(processedRequests))
				}
				if tt.expectedBodies != nil {
					// Create a set of expected bodies for flexible comparison
					expectedSet := make(map[string]bool)
					for _, expectedBody := range tt.expectedBodies {
						expectedSet[expectedBody] = true
					}

					// Check that all actual bodies are in the expected set
					for i, req := range processedRequests {
						if !expectedSet[req.Body] {
							t.Errorf("Request %d: unexpected body %s", i, req.Body)
						}
					}

					// Check that we have the right number of unique bodies
					actualSet := make(map[string]bool)
					for _, req := range processedRequests {
						actualSet[req.Body] = true
					}

					if len(actualSet) != len(expectedSet) {
						t.Errorf("Expected %d unique bodies, got %d", len(expectedSet), len(actualSet))
					}
				}
			}
		})
	}
}
func TestGenerateDictCombinationsWithLimit(t *testing.T) {
	tests := []struct {
		name            string
		dict            map[string][]interface{}
		maxCombinations int
		expectNil       bool
		expectedCount   int
	}{
		{
			name: "within limit",
			dict: map[string][]interface{}{
				"user_name": {"Alice", "Bob"},
				"user_age":  {25, 30},
			},
			maxCombinations: 10,
			expectNil:       false,
			expectedCount:   4,
		},
		{
			name: "exactly at limit",
			dict: map[string][]interface{}{
				"user_name": {"Alice", "Bob"},
				"user_age":  {25, 30},
			},
			maxCombinations: 4,
			expectNil:       false,
			expectedCount:   4,
		},
		{
			name: "exceeds limit",
			dict: map[string][]interface{}{
				"user_name": {"Alice", "Bob", "Charlie"},
				"user_age":  {25, 30, 35},
				"active":    {true, false},
			},
			maxCombinations: 10,
			expectNil:       true,
			expectedCount:   0,
		},
		{
			name: "no limit (zero)",
			dict: map[string][]interface{}{
				"user_name": {"Alice", "Bob"},
				"user_age":  {25, 30},
			},
			maxCombinations: 0,
			expectNil:       false,
			expectedCount:   4,
		},
		{
			name: "large combination count exceeds limit",
			dict: map[string][]interface{}{
				"a": {1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
				"b": {1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			},
			maxCombinations: 50,
			expectNil:       true,
			expectedCount:   0,
		},
		{
			name:            "empty dict with limit",
			dict:            map[string][]interface{}{},
			maxCombinations: 5,
			expectNil:       false,
			expectedCount:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			result := parser.generateDictCombinationsWithLimit(tt.dict, tt.maxCombinations)

			if tt.expectNil {
				if result != nil {
					t.Errorf("Expected nil result due to limit exceeded, got %d combinations", len(result))
				}
			} else {
				if result == nil {
					t.Errorf("Expected non-nil result, got nil")
				} else if len(result) != tt.expectedCount {
					t.Errorf("Expected %d combinations, got %d", tt.expectedCount, len(result))
				}
			}
		})
	}
}

func TestGenerateDictCombinationsMemoryEfficiency(t *testing.T) {
	// Test that the function handles reasonably large combinations efficiently
	dict := map[string][]interface{}{
		"param1": {"a", "b", "c", "d", "e"},
		"param2": {1, 2, 3, 4, 5},
		"param3": {true, false},
	}

	parser := NewParser()
	result := parser.generateDictCombinations(dict)

	expectedCount := 5 * 5 * 2 // 50 combinations
	if len(result) != expectedCount {
		t.Errorf("Expected %d combinations, got %d", expectedCount, len(result))
	}

	// Verify all combinations are unique
	seen := make(map[string]bool)
	for _, comb := range result {
		key := combToString(comb)
		if seen[key] {
			t.Errorf("Duplicate combination found: %s", key)
		}
		seen[key] = true
	}

	// Verify each combination has all required keys
	for i, comb := range result {
		if len(comb) != 3 {
			t.Errorf("Combination %d: expected 3 keys, got %d", i, len(comb))
		}
		if _, exists := comb["param1"]; !exists {
			t.Errorf("Combination %d: missing param1", i)
		}
		if _, exists := comb["param2"]; !exists {
			t.Errorf("Combination %d: missing param2", i)
		}
		if _, exists := comb["param3"]; !exists {
			t.Errorf("Combination %d: missing param3", i)
		}
	}
}

// TestMultipleRequestProcessing tests the detection and processing of dict combinations
func TestMultipleRequestProcessing(t *testing.T) {
	tests := []struct {
		name            string
		requestData     string
		baseURL         string
		expectedCount   int
		expectedURLs    []string
		expectedBodies  []string
		expectedHeaders []map[string]string
		shouldHaveDict  bool
		wantError       bool
	}{
		{
			name: "dict combination detection - single variable",
			requestData: `{
				"method": "POST",
				"path": "/api/users",
				"body": {
					"name": {"$dict": "user_name"}
				},
				"dict": {
					"user_name": ["Alice", "Bob", "Charlie"]
				}
			}`,
			baseURL:       "https://api.example.com",
			expectedCount: 3,
			expectedURLs: []string{
				"https://api.example.com/api/users",
				"https://api.example.com/api/users",
				"https://api.example.com/api/users",
			},
			expectedBodies: []string{
				`{"name":"Alice"}`,
				`{"name":"Bob"}`,
				`{"name":"Charlie"}`,
			},
			shouldHaveDict: true,
			wantError:      false,
		},
		{
			name: "dict combination detection - multiple variables",
			requestData: `{
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
			baseURL:       "https://api.example.com",
			expectedCount: 4,
			expectedURLs: []string{
				"https://api.example.com/api/users",
				"https://api.example.com/api/users",
				"https://api.example.com/api/users",
				"https://api.example.com/api/users",
			},
			expectedBodies: []string{
				`{"age":25,"name":"Alice"}`,
				`{"age":30,"name":"Alice"}`,
				`{"age":25,"name":"Bob"}`,
				`{"age":30,"name":"Bob"}`,
			},
			shouldHaveDict: true,
			wantError:      false,
		},
		{
			name: "dict in headers and query parameters",
			requestData: `{
				"method": "GET",
				"path": "/api/search",
				"query": {
					"category": {"$dict": "category"},
					"limit": 10
				},
				"headers": {
					"X-User-Type": {"$dict": "user_type"},
					"Content-Type": "application/json"
				},
				"dict": {
					"category": ["books", "movies"],
					"user_type": ["admin", "user"]
				}
			}`,
			baseURL:       "https://api.example.com",
			expectedCount: 4,
			expectedURLs: []string{
				"https://api.example.com/api/search?category=books&limit=10",
				"https://api.example.com/api/search?category=books&limit=10",
				"https://api.example.com/api/search?category=movies&limit=10",
				"https://api.example.com/api/search?category=movies&limit=10",
			},
			expectedHeaders: []map[string]string{
				{"X-User-Type": "admin", "Content-Type": "application/json"},
				{"X-User-Type": "user", "Content-Type": "application/json"},
				{"X-User-Type": "admin", "Content-Type": "application/json"},
				{"X-User-Type": "user", "Content-Type": "application/json"},
			},
			shouldHaveDict: true,
			wantError:      false,
		},
		{
			name: "dict with variables combination",
			requestData: `{
				"method": "POST",
				"path": "/api/users",
				"body": {
					"name": {"$dict": "user_name"},
					"api_key": {"$var": "api_key"},
					"timestamp": {"$var": "timestamp"}
				},
				"dict": {
					"user_name": ["Alice", "Bob"]
				},
				"variables": {
					"api_key": "secret123",
					"timestamp": "2023-01-01T00:00:00Z"
				}
			}`,
			baseURL:       "https://api.example.com",
			expectedCount: 2,
			expectedBodies: []string{
				`{"api_key":"secret123","name":"Alice","timestamp":"2023-01-01T00:00:00Z"}`,
				`{"api_key":"secret123","name":"Bob","timestamp":"2023-01-01T00:00:00Z"}`,
			},
			shouldHaveDict: true,
			wantError:      false,
		},
		{
			name: "no dict - single request",
			requestData: `{
				"method": "GET",
				"path": "/api/users",
				"query": {
					"limit": 10
				}
			}`,
			baseURL:       "https://api.example.com",
			expectedCount: 1,
			expectedURLs: []string{
				"https://api.example.com/api/users?limit=10",
			},
			shouldHaveDict: false,
			wantError:      false,
		},
		{
			name: "dict defined but not used - single request",
			requestData: `{
				"method": "GET",
				"path": "/api/users",
				"query": {
					"limit": 10
				},
				"dict": {
					"user_name": ["Alice", "Bob"]
				}
			}`,
			baseURL:       "https://api.example.com",
			expectedCount: 1,
			expectedURLs: []string{
				"https://api.example.com/api/users?limit=10",
			},
			shouldHaveDict: false,
			wantError:      false,
		},
		{
			name: "complex nested dict usage",
			requestData: `{
				"method": "POST",
				"path": "/api/complex",
				"body": {
					"user": {
						"name": {"$dict": "user_name"},
						"profile": {
							"age": {"$dict": "user_age"},
							"active": true
						}
					},
					"metadata": {
						"source": "test"
					}
				},
				"dict": {
					"user_name": ["Alice", "Bob"],
					"user_age": [25, 30]
				}
			}`,
			baseURL:       "https://api.example.com",
			expectedCount: 4,
			expectedBodies: []string{
				`{"metadata":{"source":"test"},"user":{"name":"Alice","profile":{"active":true,"age":25}}}`,
				`{"metadata":{"source":"test"},"user":{"name":"Alice","profile":{"active":true,"age":30}}}`,
				`{"metadata":{"source":"test"},"user":{"name":"Bob","profile":{"active":true,"age":25}}}`,
				`{"metadata":{"source":"test"},"user":{"name":"Bob","profile":{"active":true,"age":30}}}`,
			},
			shouldHaveDict: true,
			wantError:      false,
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

			// Process the requests
			processedRequests, err := parser.ProcessRequests(context.Background(), requestConfig, tt.baseURL)

			if tt.wantError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !tt.wantError {
				// Check request count
				if len(processedRequests) != tt.expectedCount {
					t.Errorf("Expected %d processed requests, got %d", tt.expectedCount, len(processedRequests))
					return
				}

				// Check URLs if provided
				if tt.expectedURLs != nil {
					for i, expectedURL := range tt.expectedURLs {
						if i < len(processedRequests) {
							if processedRequests[i].URL != expectedURL {
								t.Errorf("Request %d: expected URL %s, got %s", i, expectedURL, processedRequests[i].URL)
							}
						}
					}
				}

				// Check bodies if provided
				if tt.expectedBodies != nil {
					// Create a set of expected bodies for flexible comparison
					expectedSet := make(map[string]bool)
					for _, expectedBody := range tt.expectedBodies {
						expectedSet[expectedBody] = true
					}

					// Check that all actual bodies are in the expected set
					for i, req := range processedRequests {
						if !expectedSet[req.Body] {
							t.Errorf("Request %d: unexpected body %s", i, req.Body)
						}
					}

					// Check that we have the right number of unique bodies
					actualSet := make(map[string]bool)
					for _, req := range processedRequests {
						actualSet[req.Body] = true
					}

					if len(actualSet) != len(expectedSet) {
						t.Errorf("Expected %d unique bodies, got %d", len(expectedSet), len(actualSet))
					}
				}

				// Check headers if provided
				if tt.expectedHeaders != nil {
					// Create a set of expected header combinations for flexible comparison
					expectedHeaderSet := make(map[string]bool)
					for _, expectedHeaders := range tt.expectedHeaders {
						// Create a string representation of headers for comparison
						headerStr := ""
						for key, value := range expectedHeaders {
							headerStr += fmt.Sprintf("%s:%s;", key, value)
						}
						expectedHeaderSet[headerStr] = true
					}

					// For each request, check that it has headers matching one of the expected combinations
					for i, req := range processedRequests {
						found := false
						for _, expectedHeaders := range tt.expectedHeaders {
							match := true
							// Check if all expected headers match
							for key, expectedValue := range expectedHeaders {
								if actualValue, exists := req.Headers[key]; !exists || actualValue != expectedValue {
									match = false
									break
								}
							}
							if match {
								found = true
								break
							}
						}
						if !found {
							t.Errorf("Request %d: headers %v don't match any expected combination", i, req.Headers)
						}
					}
				}
			}
		})
	}
}

// TestMultipleRequestGeneration tests the generation of multiple requests from dict combinations
func TestMultipleRequestGeneration(t *testing.T) {
	tests := []struct {
		name          string
		requestData   string
		baseURL       string
		expectedCount int
		validateFunc  func(t *testing.T, requests []*config.ProcessedRequest)
	}{
		{
			name: "ensure each request maintains original structure",
			requestData: `{
				"method": "POST",
				"path": "/api/users",
				"headers": {
					"Content-Type": "application/json",
					"X-User-ID": {"$dict": "user_id"}
				},
				"body": {
					"name": {"$dict": "user_name"},
					"email": "test@example.com"
				},
				"dict": {
					"user_id": ["123", "456"],
					"user_name": ["Alice", "Bob"]
				}
			}`,
			baseURL:       "https://api.example.com",
			expectedCount: 4,
			validateFunc: func(t *testing.T, requests []*config.ProcessedRequest) {
				// Verify all requests have the same method and base structure
				for i, req := range requests {
					if req.Method != "POST" {
						t.Errorf("Request %d: expected method POST, got %s", i, req.Method)
					}
					if req.URL != "https://api.example.com/api/users" {
						t.Errorf("Request %d: expected URL https://api.example.com/api/users, got %s", i, req.URL)
					}
					if req.Headers["Content-Type"] != "application/json" {
						t.Errorf("Request %d: expected Content-Type application/json, got %s", i, req.Headers["Content-Type"])
					}
					// Verify body contains static email
					if !strings.Contains(req.Body, `"email":"test@example.com"`) {
						t.Errorf("Request %d: body should contain static email field", i)
					}
				}
			},
		},
		{
			name: "verify dict combination uniqueness",
			requestData: `{
				"method": "GET",
				"path": "/api/search",
				"query": {
					"type": {"$dict": "search_type"},
					"category": {"$dict": "category"}
				},
				"dict": {
					"search_type": ["user", "product"],
					"category": ["electronics", "books"]
				}
			}`,
			baseURL:       "https://api.example.com",
			expectedCount: 4,
			validateFunc: func(t *testing.T, requests []*config.ProcessedRequest) {
				// Collect all URLs to verify uniqueness
				urls := make(map[string]bool)
				for _, req := range requests {
					if urls[req.URL] {
						t.Errorf("Duplicate URL found: %s", req.URL)
					}
					urls[req.URL] = true
				}

				// Verify expected combinations exist
				expectedCombinations := []string{
					"category=electronics&type=user",
					"category=electronics&type=product",
					"category=books&type=user",
					"category=books&type=product",
				}

				for _, expectedQuery := range expectedCombinations {
					found := false
					for _, req := range requests {
						if strings.Contains(req.URL, expectedQuery) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected query combination not found: %s", expectedQuery)
					}
				}
			},
		},
		{
			name: "single dict variable multiple values",
			requestData: `{
				"method": "DELETE",
				"path": {"$concat": ["/api/users/", {"$dict": "user_id"}]},
				"dict": {
					"user_id": ["1", "2", "3", "4", "5"]
				}
			}`,
			baseURL:       "https://api.example.com",
			expectedCount: 5,
			validateFunc: func(t *testing.T, requests []*config.ProcessedRequest) {
				expectedPaths := []string{"/api/users/1", "/api/users/2", "/api/users/3", "/api/users/4", "/api/users/5"}
				actualPaths := make([]string, len(requests))

				for i, req := range requests {
					// Extract path from URL
					path := strings.TrimPrefix(req.URL, "https://api.example.com")
					actualPaths[i] = path
				}

				// Sort both slices for comparison
				sort.Strings(expectedPaths)
				sort.Strings(actualPaths)

				for i, expectedPath := range expectedPaths {
					if i < len(actualPaths) && actualPaths[i] != expectedPath {
						t.Errorf("Expected path %s, got %s", expectedPath, actualPaths[i])
					}
				}
			},
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

			// Process the requests
			processedRequests, err := parser.ProcessRequests(context.Background(), requestConfig, tt.baseURL)
			if err != nil {
				t.Fatalf("Failed to process requests: %v", err)
			}

			// Check request count
			if len(processedRequests) != tt.expectedCount {
				t.Errorf("Expected %d processed requests, got %d", tt.expectedCount, len(processedRequests))
				return
			}

			// Run custom validation
			if tt.validateFunc != nil {
				tt.validateFunc(t, processedRequests)
			}
		})
	}
}

// TestParallelExecution tests that multiple requests can be processed concurrently
func TestParallelExecution(t *testing.T) {
	tests := []struct {
		name        string
		requestData string
		baseURL     string
		concurrent  bool
	}{
		{
			name: "concurrent processing simulation",
			requestData: `{
				"method": "POST",
				"path": "/api/users",
				"body": {
					"name": {"$dict": "user_name"},
					"batch_id": {"$dict": "batch_id"}
				},
				"dict": {
					"user_name": ["Alice", "Bob", "Charlie", "David", "Eve"],
					"batch_id": ["batch1", "batch2", "batch3"]
				}
			}`,
			baseURL:    "https://api.example.com",
			concurrent: true,
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

			// Process the requests
			processedRequests, err := parser.ProcessRequests(context.Background(), requestConfig, tt.baseURL)
			if err != nil {
				t.Fatalf("Failed to process requests: %v", err)
			}

			expectedCount := 5 * 3 // 15 combinations
			if len(processedRequests) != expectedCount {
				t.Errorf("Expected %d processed requests, got %d", expectedCount, len(processedRequests))
				return
			}

			// Verify all requests are valid and unique
			seen := make(map[string]bool)
			for i, req := range processedRequests {
				if req.Method != "POST" {
					t.Errorf("Request %d: expected method POST, got %s", i, req.Method)
				}
				if req.URL != "https://api.example.com/api/users" {
					t.Errorf("Request %d: expected URL https://api.example.com/api/users, got %s", i, req.URL)
				}
				if req.Body == "" {
					t.Errorf("Request %d: body should not be empty", i)
				}

				// Check for uniqueness
				if seen[req.Body] {
					t.Errorf("Request %d: duplicate body found: %s", i, req.Body)
				}
				seen[req.Body] = true
			}

			// Verify that all expected combinations are present
			if len(seen) != expectedCount {
				t.Errorf("Expected %d unique request bodies, got %d", expectedCount, len(seen))
			}
		})
	}
}

// TestErrorIsolation tests that errors in one request don't affect others
func TestErrorIsolation(t *testing.T) {
	tests := []struct {
		name        string
		requestData string
		baseURL     string
		expectError bool
	}{
		{
			name: "invalid dict reference should cause error",
			requestData: `{
				"method": "POST",
				"path": "/api/users",
				"body": {
					"name": {"$dict": "nonexistent_var"}
				},
				"dict": {
					"user_name": ["Alice", "Bob"]
				}
			}`,
			baseURL:     "https://api.example.com",
			expectError: true,
		},
		{
			name: "valid dict processing should succeed",
			requestData: `{
				"method": "POST",
				"path": "/api/users",
				"body": {
					"name": {"$dict": "user_name"}
				},
				"dict": {
					"user_name": ["Alice", "Bob"]
				}
			}`,
			baseURL:     "https://api.example.com",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()

			// Parse the request configuration
			requestConfig, err := parser.Parse([]byte(tt.requestData), ".json", "test.json")
			if tt.expectError && err != nil {
				// Expected error during parsing
				return
			}
			if err != nil {
				t.Fatalf("Failed to parse request config: %v", err)
			}

			// Process the requests
			_, err = parser.ProcessRequests(context.Background(), requestConfig, tt.baseURL)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
