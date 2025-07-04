package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/secureta/s2http-request/internal/config"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name      string
		timeout   time.Duration
		proxy     string
		wantError bool
	}{
		{
			name:      "valid client without proxy",
			timeout:   30 * time.Second,
			proxy:     "",
			wantError: false,
		},
		{
			name:      "valid client with proxy",
			timeout:   30 * time.Second,
			proxy:     "http://proxy.example.com:8080",
			wantError: false,
		},
		{
			name:      "invalid proxy URL",
			timeout:   30 * time.Second,
			proxy:     "ht!tp://invalid",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.timeout, tt.proxy)

			if tt.wantError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.wantError && client == nil {
				t.Errorf("Expected client but got nil")
			}
			if !tt.wantError && client != nil {
				if client.timeout != tt.timeout {
					t.Errorf("Expected timeout %v, got %v", tt.timeout, client.timeout)
				}
				if client.proxy != tt.proxy {
					t.Errorf("Expected proxy %s, got %s", tt.proxy, client.proxy)
				}
			}
		})
	}
}

func TestSendRequest(t *testing.T) {
	// テスト用HTTPサーバーの作成
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// リクエストメソッドとパスの確認
		if r.Method == "GET" && r.URL.Path == "/test" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message": "success"}`))
		} else if r.Method == "POST" && r.URL.Path == "/api" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"status": "created"}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Not Found"))
		}
	}))
	defer server.Close()

	tests := []struct {
		name            string
		processedRequest *config.ProcessedRequest
		expectedStatus  int
		expectedBody    string
		wantError       bool
	}{
		{
			name: "successful GET request",
			processedRequest: &config.ProcessedRequest{
				Method:  "GET",
				URL:     server.URL + "/test",
				Headers: map[string]string{},
				Body:    "",
			},
			expectedStatus: 200,
			expectedBody:   `{"message": "success"}`,
			wantError:      false,
		},
		{
			name: "successful POST request",
			processedRequest: &config.ProcessedRequest{
				Method:  "POST",
				URL:     server.URL + "/api",
				Headers: map[string]string{"Content-Type": "application/json"},
				Body:    `{"data": "test"}`,
			},
			expectedStatus: 201,
			expectedBody:   `{"status": "created"}`,
			wantError:      false,
		},
		{
			name: "404 not found",
			processedRequest: &config.ProcessedRequest{
				Method:  "GET",
				URL:     server.URL + "/nonexistent",
				Headers: map[string]string{},
				Body:    "",
			},
			expectedStatus: 404,
			expectedBody:   "Not Found",
			wantError:      false,
		},
		{
			name: "invalid URL",
			processedRequest: &config.ProcessedRequest{
				Method:  "GET",
				URL:     "invalid-url",
				Headers: map[string]string{},
				Body:    "",
			},
			expectedStatus: 0,
			expectedBody:   "",
			wantError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(30*time.Second, "")
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			ctx := context.Background()
			response, err := client.SendRequest(ctx, tt.processedRequest)

			if tt.wantError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.wantError && response != nil {
				if response.StatusCode != tt.expectedStatus {
					t.Errorf("Expected status %d, got %d", tt.expectedStatus, response.StatusCode)
				}
				if response.Body != tt.expectedBody {
					t.Errorf("Expected body %q, got %q", tt.expectedBody, response.Body)
				}
				// タイミング情報の基本チェック
				if response.Time.Total <= 0 {
					t.Errorf("Expected positive total time, got %f", response.Time.Total)
				}
			}
		})
	}
}

func TestSendRequestWithTimeout(t *testing.T) {
	// 遅いレスポンスを返すテストサーバー
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // 2秒待機
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("slow response"))
	}))
	defer server.Close()

	client, err := NewClient(1*time.Second, "") // 1秒のタイムアウト
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	processedRequest := &config.ProcessedRequest{
		Method:  "GET",
		URL:     server.URL,
		Headers: map[string]string{},
		Body:    "",
	}

	ctx := context.Background()
	_, err = client.SendRequest(ctx, processedRequest)

	if err == nil {
		t.Errorf("Expected timeout error but got none")
	}
}

func TestSendRequestWithRetry(t *testing.T) {
	callCount := 0
	
	// 最初の2回は失敗、3回目は成功するサーバー
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Server Error"))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Success"))
		}
	}))
	defer server.Close()

	tests := []struct {
		name         string
		maxRetries   int
		expectSuccess bool
	}{
		{
			name:         "success with retries",
			maxRetries:   3,
			expectSuccess: true,
		},
		{
			name:         "failure with insufficient retries",
			maxRetries:   1,
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount = 0 // リセット
			
			client, err := NewClient(30*time.Second, "")
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			processedRequest := &config.ProcessedRequest{
				Method:  "GET",
				URL:     server.URL,
				Headers: map[string]string{},
				Body:    "",
			}

			ctx := context.Background()
			response, err := client.SendRequestWithRetry(ctx, processedRequest, tt.maxRetries)

			if tt.expectSuccess {
				if err != nil {
					t.Errorf("Expected success but got error: %v", err)
				}
				if response == nil {
					t.Errorf("Expected response but got nil")
				}
				if response != nil && response.StatusCode != 200 {
					t.Errorf("Expected status 200, got %d", response.StatusCode)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error but got success")
				}
			}
		})
	}
}

func TestSendRequestWithContext(t *testing.T) {
	// 長時間実行されるサーバー
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response"))
	}))
	defer server.Close()

	client, err := NewClient(30*time.Second, "")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	processedRequest := &config.ProcessedRequest{
		Method:  "GET",
		URL:     server.URL,
		Headers: map[string]string{},
		Body:    "",
	}

	// 1秒後にキャンセルされるコンテキスト
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err = client.SendRequest(ctx, processedRequest)

	if err == nil {
		t.Errorf("Expected context cancellation error but got none")
	}
}

func TestSendRequestHeaders(t *testing.T) {
	var receivedHeaders http.Header
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	client, err := NewClient(30*time.Second, "")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	processedRequest := &config.ProcessedRequest{
		Method: "POST",
		URL:    server.URL,
		Headers: map[string]string{
			"Content-Type":   "application/json",
			"Authorization":  "Bearer token123",
			"X-Custom-Header": "custom-value",
		},
		Body: `{"test": "data"}`,
	}

	ctx := context.Background()
	_, err = client.SendRequest(ctx, processedRequest)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// ヘッダーの確認
	expectedHeaders := map[string]string{
		"Content-Type":    "application/json",
		"Authorization":   "Bearer token123",
		"X-Custom-Header": "custom-value",
	}

	for key, expectedValue := range expectedHeaders {
		if receivedValue := receivedHeaders.Get(key); receivedValue != expectedValue {
			t.Errorf("Expected header %s: %s, got: %s", key, expectedValue, receivedValue)
		}
	}
}

func TestSendRequestDefaultContentType(t *testing.T) {
	var receivedHeaders http.Header
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	client, err := NewClient(30*time.Second, "")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	processedRequest := &config.ProcessedRequest{
		Method:  "POST",
		URL:     server.URL,
		Headers: map[string]string{}, // Content-Typeを指定しない
		Body:    "param1=value1&param2=value2",
	}

	ctx := context.Background()
	_, err = client.SendRequest(ctx, processedRequest)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// デフォルトのContent-Typeが設定されているか確認
	contentType := receivedHeaders.Get("Content-Type")
	expectedContentType := "application/x-www-form-urlencoded"
	
	if contentType != expectedContentType {
		t.Errorf("Expected default Content-Type %s, got %s", expectedContentType, contentType)
	}
}

func TestSendRequestWithFragment(t *testing.T) {
	var receivedURL string
	var receivedRawURL string
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedURL = r.URL.String()
		receivedRawURL = r.RequestURI
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	client, err := NewClient(30*time.Second, "")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	tests := []struct {
		name         string
		url          string
		expectedPath string
		expectFragment bool
	}{
		{
			name:         "URL with fragment",
			url:          server.URL + "/test#fragment",
			expectedPath: "/test#fragment",
			expectFragment: true,
		},
		{
			name:         "URL with fragment and query",
			url:          server.URL + "/api?param=value#section",
			expectedPath: "/api?param=value#section",
			expectFragment: true,
		},
		{
			name:         "URL without fragment",
			url:          server.URL + "/normal",
			expectedPath: "/normal",
			expectFragment: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processedRequest := &config.ProcessedRequest{
				Method:  "GET",
				URL:     tt.url,
				Headers: map[string]string{},
				Body:    "",
			}

			ctx := context.Background()
			_, err := client.SendRequest(ctx, processedRequest)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			t.Logf("Received URL: %s", receivedURL)
			t.Logf("Received Raw URL: %s", receivedRawURL)

			if tt.expectFragment {
				// フラグメントがサーバーに送信されているかを確認
				if !strings.Contains(receivedRawURL, "#") {
					t.Errorf("Expected fragment in request URI, but got: %s", receivedRawURL)
				}
			} else {
				// フラグメントがないことを確認
				if strings.Contains(receivedRawURL, "#") {
					t.Errorf("Unexpected fragment in request URI: %s", receivedRawURL)
				}
			}
		})
	}
}