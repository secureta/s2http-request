package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/secureta/s2http-request/internal/config"
	httpClient "github.com/secureta/s2http-request/internal/http"
)

// TestDictEndToEndWithHTTPServer tests dict functionality with actual HTTP requests
func TestDictEndToEndWithHTTPServer(t *testing.T) {
	// Track received requests for verification
	var receivedRequests []ReceivedRequest
	var mu sync.Mutex

	// Create test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		// Parse request body
		var bodyMap map[string]interface{}
		if r.Body != nil {
			decoder := json.NewDecoder(r.Body)
			_ = decoder.Decode(&bodyMap) // Ignore errors for non-JSON bodies
		}

		// Store received request
		receivedRequests = append(receivedRequests, ReceivedRequest{
			Method:  r.Method,
			Path:    r.URL.Path,
			Query:   r.URL.RawQuery,
			Headers: r.Header.Clone(),
			Body:    bodyMap,
		})

		// Send response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{
			"status":  "success",
			"request": len(receivedRequests),
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	tests := []struct {
		name           string
		configContent  string
		fileExt        string
		expectedCount  int
		verifyRequests func(t *testing.T, requests []ReceivedRequest)
	}{
		{
			name: "basic_dict_combinations",
			configContent: fmt.Sprintf(`{
				"method": "POST",
				"path": "/api/users",
				"headers": {
					"Content-Type": "application/json"
				},
				"body": {
					"name": {"$dict": "user_name"},
					"age": {"$dict": "user_age"}
				},
				"dict": {
					"user_name": ["Alice", "Bob"],
					"user_age": [25, 30]
				}
			}`),
			fileExt:       ".json",
			expectedCount: 4,
			verifyRequests: func(t *testing.T, requests []ReceivedRequest) {
				expectedCombinations := []map[string]interface{}{
					{"name": "Alice", "age": float64(25)},
					{"name": "Alice", "age": float64(30)},
					{"name": "Bob", "age": float64(25)},
					{"name": "Bob", "age": float64(30)},
				}

				if len(requests) != 4 {
					t.Errorf("Expected 4 requests, got %d", len(requests))
					return
				}

				for i, req := range requests {
					if req.Method != "POST" {
						t.Errorf("Request %d: Expected POST, got %s", i, req.Method)
					}
					if req.Path != "/api/users" {
						t.Errorf("Request %d: Expected /api/users, got %s", i, req.Path)
					}

					// Check if body matches one of expected combinations
					found := false
					for _, expected := range expectedCombinations {
						if bodyMatches(req.Body, expected) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Request %d: Body %v doesn't match any expected combination", i, req.Body)
					}
				}
			},
		},
		{
			name: "dict_with_query_parameters",
			configContent: fmt.Sprintf(`method: GET
path: /api/search
query:
  category:
    $dict: category
  limit:
    $dict: limit
  format: json
dict:
  category: ["books", "electronics"]
  limit: [10, 20]`),
			fileExt:       ".yaml",
			expectedCount: 4,
			verifyRequests: func(t *testing.T, requests []ReceivedRequest) {
				expectedQueries := []map[string]string{
					{"category": "books", "limit": "10", "format": "json"},
					{"category": "books", "limit": "20", "format": "json"},
					{"category": "electronics", "limit": "10", "format": "json"},
					{"category": "electronics", "limit": "20", "format": "json"},
				}

				if len(requests) != 4 {
					t.Errorf("Expected 4 requests, got %d", len(requests))
					return
				}

				for i, req := range requests {
					if req.Method != "GET" {
						t.Errorf("Request %d: Expected GET, got %s", i, req.Method)
					}

					// Parse query parameters
					queryParams := parseQueryString(req.Query)

					// Check if query matches one of expected combinations
					found := false
					for _, expected := range expectedQueries {
						if queryMatches(queryParams, expected) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Request %d: Query %v doesn't match any expected combination", i, queryParams)
					}
				}
			},
		},
		{
			name: "dict_with_variables_mixed",
			configContent: fmt.Sprintf(`{
				"method": "POST",
				"path": "/api/orders",
				"headers": {
					"Authorization": {"$var": "auth_token"},
					"Content-Type": "application/json"
				},
				"body": {
					"product": {"$dict": "product"},
					"quantity": {"$dict": "quantity"},
					"user_id": {"$var": "user_id"}
				},
				"variables": {
					"auth_token": "Bearer token123",
					"user_id": "user_456"
				},
				"dict": {
					"product": ["laptop", "mouse"],
					"quantity": [1, 2]
				}
			}`),
			fileExt:       ".json",
			expectedCount: 4,
			verifyRequests: func(t *testing.T, requests []ReceivedRequest) {
				if len(requests) != 4 {
					t.Errorf("Expected 4 requests, got %d", len(requests))
					return
				}

				for i, req := range requests {
					// Check headers
					if auth := req.Headers.Get("Authorization"); auth != "Bearer token123" {
						t.Errorf("Request %d: Expected Authorization header 'Bearer token123', got '%s'", i, auth)
					}

					// Check body contains user_id from variables
					if userID, ok := req.Body["user_id"]; !ok || userID != "user_456" {
						t.Errorf("Request %d: Expected user_id 'user_456', got %v", i, userID)
					}

					// Check dict variables are present
					if _, ok := req.Body["product"]; !ok {
						t.Errorf("Request %d: Missing product field", i)
					}
					if _, ok := req.Body["quantity"]; !ok {
						t.Errorf("Request %d: Missing quantity field", i)
					}
				}
			},
		},
		{
			name: "complex_dict_combinations",
			configContent: fmt.Sprintf(`method: POST
path: /api/complex
body:
  user:
    name:
      $dict: user_name
    role:
      $dict: user_role
  settings:
    theme:
      $dict: theme
    notifications: true
dict:
  user_name: ["Alice", "Bob"]
  user_role: ["admin", "user"]
  theme: ["dark", "light"]`),
			fileExt:       ".yaml",
			expectedCount: 8, // 2 * 2 * 2 = 8 combinations
			verifyRequests: func(t *testing.T, requests []ReceivedRequest) {
				if len(requests) != 8 {
					t.Errorf("Expected 8 requests, got %d", len(requests))
					return
				}

				// Verify all requests have the expected structure
				for i, req := range requests {
					if req.Method != "POST" {
						t.Errorf("Request %d: Expected POST, got %s", i, req.Method)
					}

					// Check nested structure
					if user, ok := req.Body["user"].(map[string]interface{}); ok {
						if _, ok := user["name"]; !ok {
							t.Errorf("Request %d: Missing user.name", i)
						}
						if _, ok := user["role"]; !ok {
							t.Errorf("Request %d: Missing user.role", i)
						}
					} else {
						t.Errorf("Request %d: Missing or invalid user object", i)
					}

					if settings, ok := req.Body["settings"].(map[string]interface{}); ok {
						if _, ok := settings["theme"]; !ok {
							t.Errorf("Request %d: Missing settings.theme", i)
						}
						if notifications, ok := settings["notifications"]; !ok || notifications != true {
							t.Errorf("Request %d: Expected settings.notifications to be true", i)
						}
					} else {
						t.Errorf("Request %d: Missing or invalid settings object", i)
					}
				}
			},
		},
	}

	parser := NewParser()
	client, err := httpClient.NewClient(30*time.Second, "")
	if err != nil {
		t.Fatalf("Failed to create HTTP client: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset received requests
			mu.Lock()
			receivedRequests = nil
			mu.Unlock()

			// Parse configuration
			configs, err := parser.ParseMultiple([]byte(tt.configContent), tt.fileExt, "test"+tt.fileExt)
			if err != nil {
				t.Fatalf("Failed to parse config: %v", err)
			}

			if len(configs) == 0 {
				t.Fatal("No configurations parsed")
			}

			// Process requests
			ctx := context.Background()
			processedRequests, err := parser.ProcessRequests(ctx, configs[0], server.URL)
			if err != nil {
				t.Fatalf("Failed to process requests: %v", err)
			}

			if len(processedRequests) != tt.expectedCount {
				t.Errorf("Expected %d processed requests, got %d", tt.expectedCount, len(processedRequests))
			}

			// Send all requests to the server
			for i, req := range processedRequests {
				response, err := client.SendRequest(ctx, req)
				if err != nil {
					t.Errorf("Request %d failed: %v", i, err)
					continue
				}

				if response.StatusCode != 200 {
					t.Errorf("Request %d: Expected status 200, got %d", i, response.StatusCode)
				}
			}

			// Wait a bit for all requests to be processed
			time.Sleep(100 * time.Millisecond)

			// Verify received requests
			mu.Lock()
			requests := make([]ReceivedRequest, len(receivedRequests))
			copy(requests, receivedRequests)
			mu.Unlock()

			tt.verifyRequests(t, requests)
		})
	}
}

// TestDictParallelExecution tests that dict requests can be executed in parallel
func TestDictParallelExecution(t *testing.T) {
	// Track request timing to verify parallel execution
	var requestTimes []time.Time
	var mu sync.Mutex

	// Create server that takes some time to respond
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestTimes = append(requestTimes, time.Now())
		mu.Unlock()

		// Simulate processing time
		time.Sleep(100 * time.Millisecond)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	configContent := fmt.Sprintf(`{
		"method": "GET",
		"path": "/api/test",
		"query": {
			"id": {"$dict": "test_id"}
		},
		"dict": {
			"test_id": ["1", "2", "3", "4", "5"]
		}
	}`)

	parser := NewParser()
	client, err := httpClient.NewClient(30*time.Second, "")
	if err != nil {
		t.Fatalf("Failed to create HTTP client: %v", err)
	}

	// Parse and process requests
	configs, err := parser.ParseMultiple([]byte(configContent), ".json", "test.json")
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	ctx := context.Background()
	processedRequests, err := parser.ProcessRequests(ctx, configs[0], server.URL)
	if err != nil {
		t.Fatalf("Failed to process requests: %v", err)
	}

	if len(processedRequests) != 5 {
		t.Fatalf("Expected 5 requests, got %d", len(processedRequests))
	}

	// Execute requests in parallel using goroutines
	startTime := time.Now()
	var wg sync.WaitGroup
	var errors []error
	var errorsMu sync.Mutex

	for i, req := range processedRequests {
		wg.Add(1)
		go func(index int, request *config.ProcessedRequest) {
			defer wg.Done()

			_, err := client.SendRequest(ctx, request)
			if err != nil {
				errorsMu.Lock()
				errors = append(errors, fmt.Errorf("request %d failed: %w", index, err))
				errorsMu.Unlock()
			}
		}(i, req)
	}

	wg.Wait()
	totalTime := time.Since(startTime)

	// Check for errors
	if len(errors) > 0 {
		for _, err := range errors {
			t.Error(err)
		}
		t.FailNow()
	}

	// Verify parallel execution
	// If executed sequentially, it would take at least 5 * 100ms = 500ms
	// With parallel execution, it should be much less
	if totalTime > 300*time.Millisecond {
		t.Errorf("Parallel execution took too long: %v (expected < 300ms)", totalTime)
	}

	// Verify all requests were received
	mu.Lock()
	receivedCount := len(requestTimes)
	mu.Unlock()

	if receivedCount != 5 {
		t.Errorf("Expected 5 requests received, got %d", receivedCount)
	}

	// Verify requests arrived within a reasonable time window (indicating parallelism)
	if len(requestTimes) >= 2 {
		mu.Lock()
		firstRequest := requestTimes[0]
		lastRequest := requestTimes[len(requestTimes)-1]
		mu.Unlock()

		requestSpread := lastRequest.Sub(firstRequest)
		// All requests should arrive within 200ms if executed in parallel
		if requestSpread > 200*time.Millisecond {
			t.Errorf("Request spread too large: %v (expected < 200ms)", requestSpread)
		}
	}
}

// TestDictPerformanceWithLargeCombinations tests performance with many combinations
func TestDictPerformanceWithLargeCombinations(t *testing.T) {
	// Create a simple server for performance testing
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	tests := []struct {
		name          string
		configContent string
		expectedCount int
		maxTime       time.Duration
	}{
		{
			name: "moderate_combinations",
			configContent: fmt.Sprintf(`{
				"method": "POST",
				"path": "/api/test",
				"body": {
					"param1": {"$dict": "values1"},
					"param2": {"$dict": "values2"},
					"param3": {"$dict": "values3"}
				},
				"dict": {
					"values1": ["a", "b", "c", "d", "e"],
					"values2": ["1", "2", "3", "4"],
					"values3": ["x", "y", "z"]
				}
			}`),
			expectedCount: 60, // 5 * 4 * 3 = 60
			maxTime:       5 * time.Second,
		},
		{
			name: "large_single_array",
			configContent: fmt.Sprintf(`{
				"method": "GET",
				"path": "/api/items",
				"query": {
					"id": {"$dict": "item_ids"}
				},
				"dict": {
					"item_ids": ["1", "2", "3", "4", "5", "6", "7", "8", "9", "10", 
					           "11", "12", "13", "14", "15", "16", "17", "18", "19", "20"]
				}
			}`),
			expectedCount: 20,
			maxTime:       3 * time.Second,
		},
	}

	parser := NewParser()
	client, err := httpClient.NewClient(30*time.Second, "")
	if err != nil {
		t.Fatalf("Failed to create HTTP client: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startTime := time.Now()

			// Parse configuration
			configs, err := parser.ParseMultiple([]byte(tt.configContent), ".json", "test.json")
			if err != nil {
				t.Fatalf("Failed to parse config: %v", err)
			}

			// Process requests
			ctx := context.Background()
			processedRequests, err := parser.ProcessRequests(ctx, configs[0], server.URL)
			if err != nil {
				t.Fatalf("Failed to process requests: %v", err)
			}

			processingTime := time.Since(startTime)

			if len(processedRequests) != tt.expectedCount {
				t.Errorf("Expected %d requests, got %d", tt.expectedCount, len(processedRequests))
			}

			// Execute a subset of requests to test performance
			maxRequests := 10
			if len(processedRequests) < maxRequests {
				maxRequests = len(processedRequests)
			}

			executionStart := time.Now()
			var wg sync.WaitGroup
			for i := 0; i < maxRequests; i++ {
				wg.Add(1)
				go func(req *config.ProcessedRequest) {
					defer wg.Done()
					client.SendRequest(ctx, req)
				}(processedRequests[i])
			}
			wg.Wait()
			executionTime := time.Since(executionStart)

			totalTime := time.Since(startTime)

			t.Logf("Processing time: %v", processingTime)
			t.Logf("Execution time (%d requests): %v", maxRequests, executionTime)
			t.Logf("Total time: %v", totalTime)

			if totalTime > tt.maxTime {
				t.Errorf("Performance test took too long: %v (expected < %v)", totalTime, tt.maxTime)
			}
		})
	}
}

// Helper types and functions

type ReceivedRequest struct {
	Method  string
	Path    string
	Query   string
	Headers http.Header
	Body    map[string]interface{}
}

func bodyMatches(actual map[string]interface{}, expected map[string]interface{}) bool {
	if len(actual) != len(expected) {
		return false
	}

	for key, expectedValue := range expected {
		actualValue, exists := actual[key]
		if !exists || actualValue != expectedValue {
			return false
		}
	}

	return true
}

func parseQueryString(query string) map[string]string {
	result := make(map[string]string)
	if query == "" {
		return result
	}

	pairs := strings.Split(query, "&")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}

	return result
}

func queryMatches(actual map[string]string, expected map[string]string) bool {
	for key, expectedValue := range expected {
		actualValue, exists := actual[key]
		if !exists || actualValue != expectedValue {
			return false
		}
	}

	return true
}
