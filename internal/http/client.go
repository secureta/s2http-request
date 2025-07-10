package http

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/secureta/s2http-request/internal/config"
)

// Client はHTTPクライアント
type Client struct {
	httpClient *http.Client
	timeout    time.Duration
	proxy      string
}

// fragmentTransport はフラグメントを含むリクエストを送信するためのカスタムトランスポート
type fragmentTransport struct {
	base http.RoundTripper
}

func (t *fragmentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// フラグメントが含まれているかチェック
	if !strings.Contains(req.URL.String(), "#") {
		// フラグメントがない場合は通常の処理
		if t.base != nil {
			return t.base.RoundTrip(req)
		}
		return http.DefaultTransport.RoundTrip(req)
	}

	// フラグメントがある場合は手動でリクエストを構築
	return t.sendRequestWithFragment(req)
}

func (t *fragmentTransport) sendRequestWithFragment(req *http.Request) (resp *http.Response, err error) {
	// URLを解析
	parsedURL := req.URL
	host := parsedURL.Host
	if parsedURL.Port() == "" {
		if parsedURL.Scheme == "https" {
			host += ":443"
		} else {
			host += ":80"
		}
	}

	// TCP接続を確立
	conn, err := net.DialTimeout("tcp", host, 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close connection: %w", closeErr)
		}
	}()

	// リクエストラインを構築（フラグメントを含む）
	requestURI := parsedURL.Path
	if parsedURL.RawQuery != "" {
		requestURI += "?" + parsedURL.RawQuery
	}
	if parsedURL.Fragment != "" {
		requestURI += "#" + parsedURL.Fragment
	}
	if requestURI == "" {
		requestURI = "/"
	}

	// HTTPリクエストを手動で構築
	requestLine := fmt.Sprintf("%s %s HTTP/1.1\r\n", req.Method, requestURI)

	// ヘッダーを構築
	headers := ""
	headers += fmt.Sprintf("Host: %s\r\n", parsedURL.Host)

	for key, values := range req.Header {
		for _, value := range values {
			headers += fmt.Sprintf("%s: %s\r\n", key, value)
		}
	}

	// Connection: close を追加（シンプルにするため）
	headers += "Connection: close\r\n"

	// リクエストボディの処理
	var body string
	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
		body = string(bodyBytes)
		if body != "" {
			headers += fmt.Sprintf("Content-Length: %d\r\n", len(body))
		}
	}

	// 完全なHTTPリクエストを構築
	fullRequest := requestLine + headers + "\r\n" + body

	// リクエストを送信
	_, err = conn.Write([]byte(fullRequest))
	if err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	// レスポンスを読み取り
	reader := bufio.NewReader(conn)
	resp, err = http.ReadResponse(reader, req)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return resp, nil
}

// NewClient は新しいHTTPクライアントを作成
func NewClient(timeout time.Duration, proxy string) (*Client, error) {
	transport := &fragmentTransport{
		base: http.DefaultTransport,
	}

	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}

	// プロキシ設定
	if proxy != "" {
		proxyURL, err := url.Parse(proxy)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %w", err)
		}

		baseTransport := &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
		transport.base = baseTransport
	}

	return &Client{
		httpClient: client,
		timeout:    timeout,
		proxy:      proxy,
	}, nil
}

// SendRequest はHTTPリクエストを送信
func (c *Client) SendRequest(ctx context.Context, processedRequest *config.ProcessedRequest) (responseData *config.ResponseData, err error) {
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
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close response body: %w", closeErr)
		}
	}()

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
	responseData = &config.ResponseData{
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
			lastErr = err
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
