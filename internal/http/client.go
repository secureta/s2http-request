package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/user/simple-request-dispatcher/internal/config"
)

// Client はHTTPクライアント
type Client struct {
	httpClient *http.Client
	timeout    time.Duration
	proxy      string
}

// NewClient は新しいHTTPクライアントを作成
func NewClient(timeout time.Duration, proxy string) (*Client, error) {
	client := &http.Client{
		Timeout: timeout,
	}

	// プロキシ設定
	if proxy != "" {
		proxyURL, err := url.Parse(proxy)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %w", err)
		}
		
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
		client.Transport = transport
	}

	return &Client{
		httpClient: client,
		timeout:    timeout,
		proxy:      proxy,
	}, nil
}

// SendRequest はHTTPリクエストを送信
func (c *Client) SendRequest(ctx context.Context, processedRequest *config.ProcessedRequest) (*config.ResponseData, error) {
	// タイミング測定用
	startTime := time.Now()
	var dnsTime, connectTime, sslTime, sendTime, waitTime, receiveTime time.Duration

	// リクエストボディの準備
	var bodyReader io.Reader
	if processedRequest.Body != "" {
		bodyReader = strings.NewReader(processedRequest.Body)
	}

	// HTTPリクエストの作成
	req, err := http.NewRequestWithContext(ctx, processedRequest.Method, processedRequest.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// ヘッダーの設定
	for key, value := range processedRequest.Headers {
		req.Header.Set(key, value)
	}

	// Content-Typeが設定されていない場合のデフォルト設定
	if processedRequest.Body != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	// リクエスト送信時刻
	sendTime = time.Since(startTime)

	// HTTPリクエストの送信
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// レスポンス受信時刻
	waitTime = time.Since(startTime) - sendTime

	// レスポンスボディの読み取り
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// レスポンス処理完了時刻
	receiveTime = time.Since(startTime) - waitTime - sendTime
	totalTime := time.Since(startTime)

	// レスポンスデータの構築
	responseData := &config.ResponseData{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       string(bodyBytes),
		Time: config.ResponseTiming{
			Total:   totalTime.Seconds(),
			DNS:     dnsTime.Seconds(),
			Connect: connectTime.Seconds(),
			SSL:     sslTime.Seconds(),
			Send:    sendTime.Seconds(),
			Wait:    waitTime.Seconds(),
			Receive: receiveTime.Seconds(),
		},
	}

	return responseData, nil
}

// SendRequestWithRetry はリトライ機能付きでHTTPリクエストを送信
func (c *Client) SendRequestWithRetry(ctx context.Context, processedRequest *config.ProcessedRequest, maxRetries int) (*config.ResponseData, error) {
	var lastErr error
	
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// リトライ前の待機時間（指数バックオフ）
			waitTime := time.Duration(attempt) * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(waitTime):
			}
		}

		response, err := c.SendRequest(ctx, processedRequest)
		if err == nil && response.StatusCode < 500 {
			return response, nil
		}
		if err != nil {
		} else if response.StatusCode >= 500 {
			lastErr = fmt.Errorf("server error: %d", response.StatusCode)
		} else {
			return response, nil
		}

		
		// コンテキストがキャンセルされた場合はリトライしない
		if ctx.Err() != nil {
			break
		}
	}

	return nil, fmt.Errorf("request failed after %d retries: %w", maxRetries, lastErr)
}