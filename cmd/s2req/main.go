package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/secureta/s2http-request/internal/config"
	"github.com/secureta/s2http-request/internal/http"
	"github.com/secureta/s2http-request/internal/parser"
	"gopkg.in/yaml.v3"
)

var (
	version = "dev"
)

// parseRequestIDOption はRequest IDオプションをパースする
func parseRequestIDOption(option string) (*config.RequestIDConfig, error) {
	parts := strings.SplitN(option, "=", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid format, expected 'type=value'")
	}

	switch parts[0] {
	case "path":
		switch parts[1] {
		case "head":
			return &config.RequestIDConfig{
				Location: config.RequestIDLocationPathHead,
			}, nil
		case "tail":
			return &config.RequestIDConfig{
				Location: config.RequestIDLocationPathTail,
			}, nil
		default:
			return nil, fmt.Errorf("invalid path value, expected 'head' or 'tail'")
		}
	case "query":
		if parts[1] == "" {
			return nil, fmt.Errorf("query key cannot be empty")
		}
		return &config.RequestIDConfig{
			Location: config.RequestIDLocationQuery,
			Key:      parts[1],
		}, nil
	case "header":
		if parts[1] == "" {
			return nil, fmt.Errorf("header key cannot be empty")
		}
		return &config.RequestIDConfig{
			Location: config.RequestIDLocationHeader,
			Key:      parts[1],
		}, nil
	default:
		return nil, fmt.Errorf("invalid type, expected 'path', 'query', or 'header'")
	}
}

// getDefaultUserAgent returns the default User-Agent string
func getDefaultUserAgent() string {
	return fmt.Sprintf("s2req/%s (https://github.com/secureta/s2http-request)", version)
}

// processStdin reads from stdin and processes the input
func processStdin(p *parser.Parser, client *http.Client, cliConfig *config.CLIConfig, userAgent string) ([]*config.Result, error) {
	// Read all data from stdin
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("failed to read from stdin: %w", err)
	}

	// Detect the format of the input
	format := detectFormat(data)

	// Parse the input
	requestConfigs, err := p.ParseMultiple(data, format, "stdin")
	if err != nil {
		return nil, fmt.Errorf("failed to parse stdin input: %w", err)
	}

	var allResults []*config.Result

	// Process each request config
	for _, requestConfig := range requestConfigs {
		// Process requests with Request ID
		processedRequests, err := p.ProcessRequestsWithRequestID(context.Background(), requestConfig, cliConfig.Host, cliConfig.RequestID)
		if err != nil {
			return nil, fmt.Errorf("failed to process requests: %w", err)
		}

		// Send each processed request
		for _, processedRequest := range processedRequests {
			// Set User-Agent if not specified
			if _, exists := processedRequest.Headers["User-Agent"]; !exists {
				if userAgent != "" {
					processedRequest.Headers["User-Agent"] = userAgent
				} else {
					processedRequest.Headers["User-Agent"] = getDefaultUserAgent()
				}
			}

			// Send the request
			ctx, cancel := context.WithTimeout(context.Background(), cliConfig.Timeout)
			defer cancel()

			var response *config.ResponseData
			if cliConfig.Retry > 0 {
				response, err = client.SendRequestWithRetry(ctx, processedRequest, cliConfig.Retry)
			} else {
				response, err = client.SendRequest(ctx, processedRequest)
			}

			if err != nil {
				log.Printf("Failed to send request: %v", err)
				continue
			}

			// Create result
			result := &config.Result{
				Request:  *processedRequest,
				Response: *response,
				Metadata: map[string]interface{}{
					"file":       "stdin",
					"timestamp":  time.Now().Format(time.RFC3339),
					"request_id": processedRequest.RequestID,
				},
			}

			allResults = append(allResults, result)

			// Verbose output
			if cliConfig.Verbose {
				fmt.Printf("Request: %s %s\n", processedRequest.Method, processedRequest.URL)
				if processedRequest.RequestID != "" {
					fmt.Printf("Request ID: %s\n", processedRequest.RequestID)
				}
				fmt.Printf("Response: %d\n", response.StatusCode)
			}
		}
	}

	return allResults, nil
}

// detectFormat detects the format of the input data
func detectFormat(data []byte) string {
	// Try to parse as JSON
	var jsonObj interface{}
	if err := json.Unmarshal(data, &jsonObj); err == nil {
		// Check if it's JSONL by looking for newlines with valid JSON objects
		lines := strings.Split(string(data), "\n")
		if len(lines) > 1 {
			// Check if at least one more line is valid JSON
			for _, line := range lines[1:] {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
					continue // Skip empty lines and comments
				}
				var lineObj interface{}
				if json.Unmarshal([]byte(line), &lineObj) == nil {
					return ".jsonl" // It's JSONL
				}
				break
			}
		}
		return ".json" // It's JSON
	}

	// Try to parse as YAML
	var yamlObj interface{}
	if err := yaml.Unmarshal(data, &yamlObj); err == nil {
		return ".yaml" // It's YAML
	}

	// Default to JSON if we can't determine the format
	return ".json"
}

func main() {
	var (
		host      = flag.String("host", "http://localhost", "Target host URL")
		timeout   = flag.Duration("timeout", 30*time.Second, "Request timeout")
		retry     = flag.Int("retry", 0, "Number of retries")
		proxy     = flag.String("proxy", "", "Proxy URL")
		verbose   = flag.Bool("verbose", false, "Verbose output")
		output    = flag.String("output", "", "Output file path")
		format    = flag.String("format", "json", "Output format (json, csv, table)")
		userAgent = flag.String("user-agent", "", "Override User-Agent header")
		requestID = flag.String("request-id", "", "Enable Request ID (path=head|tail, query=<key>, header=<key>)")
		showVersion = flag.Bool("version", false, "Show version")
	)

	flag.Parse()

	if *showVersion {
		fmt.Printf("s2req version %s\n", version)
		return
	}

	files := flag.Args()

	// Check if we should read from stdin
	readFromStdin := len(files) == 0 || (len(files) == 1 && files[0] == "-")

	if !readFromStdin && len(files) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <request-file>... or provide input via stdin\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Request IDの設定をパース
	var requestIDConfig *config.RequestIDConfig
	if *requestID != "" {
		var err error
		requestIDConfig, err = parseRequestIDOption(*requestID)
		if err != nil {
			log.Fatalf("Invalid request-id option: %v", err)
		}
	}

	// CLI設定の作成
	cliConfig := &config.CLIConfig{
		Host:      *host,
		Timeout:   *timeout,
		Retry:     *retry,
		Proxy:     *proxy,
		Verbose:   *verbose,
		Output:    *output,
		Format:    config.OutputFormat(*format),
		Files:     files,
		RequestID: requestIDConfig,
	}

	// If reading from stdin, update the Files field
	if readFromStdin {
		cliConfig.Files = []string{"stdin"}
	}

	// HTTPクライアントの作成
	client, err := http.NewClient(cliConfig.Timeout, cliConfig.Proxy)
	if err != nil {
		log.Fatalf("Failed to create HTTP client: %v", err)
	}

	// パーサーの作成
	p := parser.NewParser()

	var results []*config.Result

	if readFromStdin {
		// Process stdin input
		stdinResults, err := processStdin(p, client, cliConfig, *userAgent)
		if err != nil {
			log.Fatalf("Error processing stdin: %v", err)
		}
		results = append(results, stdinResults...)
	} else {
		// 各ファイルを処理
		for _, filePath := range files {
			fileResults, err := processFile(p, client, cliConfig, filePath, *userAgent)
			if err != nil {
				log.Printf("Error processing file %s: %v", filePath, err)
				continue
			}
			results = append(results, fileResults...)
		}
	}

	// 結果の出力
	if err := outputResults(results, cliConfig); err != nil {
		log.Fatalf("Failed to output results: %v", err)
	}
}

func processFile(p *parser.Parser, client *http.Client, cliConfig *config.CLIConfig, filePath string, userAgent string) ([]*config.Result, error) {
	// ファイルの読み込み
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// ファイル拡張子の取得
	ext := filepath.Ext(filePath)

	// リクエスト設定の解析（複数のドキュメントに対応）
	requestConfigs, err := p.ParseMultiple(data, ext, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse request config: %w", err)
	}

	var allResults []*config.Result

	// 各リクエスト設定を処理
	for _, requestConfig := range requestConfigs {
		// リクエストの処理（辞書展開を含む）
		processedRequests, err := p.ProcessRequestsWithRequestID(context.Background(), requestConfig, cliConfig.Host, cliConfig.RequestID)
		if err != nil {
			return nil, fmt.Errorf("failed to process requests: %w", err)
		}

		// 各処理済みリクエストを送信
		for _, processedRequest := range processedRequests {
			// User-Agentの設定処理
			if _, exists := processedRequest.Headers["User-Agent"]; !exists {
				// JSONやYAMLファイルでUser-Agentが指定されていない場合
				if userAgent != "" {
					// コマンドライン引数が指定されている場合はそれを使用
					processedRequest.Headers["User-Agent"] = userAgent
				} else {
					// コマンドライン引数も指定されていない場合はデフォルト値を使用
					processedRequest.Headers["User-Agent"] = getDefaultUserAgent()
				}
			}

			// リクエストの送信
			ctx, cancel := context.WithTimeout(context.Background(), cliConfig.Timeout)
			defer cancel()

			var response *config.ResponseData
			if cliConfig.Retry > 0 {
				response, err = client.SendRequestWithRetry(ctx, processedRequest, cliConfig.Retry)
			} else {
				response, err = client.SendRequest(ctx, processedRequest)
			}

			if err != nil {
				log.Printf("Failed to send request: %v", err)
				continue
			}

			// 結果の作成
			result := &config.Result{
				Request:  *processedRequest,
				Response: *response,
				Metadata: map[string]interface{}{
					"file":       filePath,
					"timestamp":  time.Now().Format(time.RFC3339),
					"request_id": processedRequest.RequestID,
				},
			}

			allResults = append(allResults, result)

			// Verbose出力
			if cliConfig.Verbose {
				fmt.Printf("Request: %s %s\n", processedRequest.Method, processedRequest.URL)
				if processedRequest.RequestID != "" {
					fmt.Printf("Request ID: %s\n", processedRequest.RequestID)
				}
				fmt.Printf("Response: %d\n", response.StatusCode)
			}
		}
	}

	return allResults, nil
}

func outputResults(results []*config.Result, cliConfig *config.CLIConfig) error {
	var output []byte
	var err error

	switch cliConfig.Format {
	case config.OutputFormatJSON:
		output, err = json.MarshalIndent(results, "", "  ")
	case config.OutputFormatCSV:
		output, err = formatAsCSV(results)
	case config.OutputFormatTable:
		output, err = formatAsTable(results)
	default:
		output, err = json.MarshalIndent(results, "", "  ")
	}

	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	if cliConfig.Output != "" {
		return ioutil.WriteFile(cliConfig.Output, output, 0644)
	}

	fmt.Print(string(output))
	return nil
}

func formatAsCSV(results []*config.Result) ([]byte, error) {
	var lines []string
	lines = append(lines, "Method,URL,StatusCode,ResponseTime,BodyLength,RequestID")

	for _, result := range results {
		line := fmt.Sprintf("%s,%s,%d,%.3f,%d,%s",
			result.Request.Method,
			result.Request.URL,
			result.Response.StatusCode,
			result.Response.Time.Total,
			len(result.Response.Body),
			result.Request.RequestID)
		lines = append(lines, line)
	}

	return []byte(strings.Join(lines, "\n")), nil
}

func formatAsTable(results []*config.Result) ([]byte, error) {
	var lines []string
	lines = append(lines, "METHOD\tURL\tSTATUS\tTIME\tSIZE\tREQUEST_ID")
	lines = append(lines, strings.Repeat("-", 100))

	for _, result := range results {
		line := fmt.Sprintf("%s\t%s\t%d\t%.3fs\t%d\t%s",
			result.Request.Method,
			result.Request.URL,
			result.Response.StatusCode,
			result.Response.Time.Total,
			len(result.Response.Body),
			result.Request.RequestID)
		lines = append(lines, line)
	}

	return []byte(strings.Join(lines, "\n")), nil
}
