package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/secureta/s2http-request/internal/config"
	"github.com/secureta/s2http-request/internal/http"
	"github.com/secureta/s2http-request/internal/parser"
)

var (
	version = "dev"
)

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
		showVersion = flag.Bool("version", false, "Show version")
	)

	flag.Parse()

	if *showVersion {
		fmt.Printf("s2req version %s\n", version)
		return
	}

	files := flag.Args()
	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <request-file>...\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	// CLI設定の作成
	cliConfig := &config.CLIConfig{
		Host:    *host,
		Timeout: *timeout,
		Retry:   *retry,
		Proxy:   *proxy,
		Verbose: *verbose,
		Output:  *output,
		Format:  config.OutputFormat(*format),
		Files:   files,
	}

	// HTTPクライアントの作成
	client, err := http.NewClient(cliConfig.Timeout, cliConfig.Proxy)
	if err != nil {
		log.Fatalf("Failed to create HTTP client: %v", err)
	}

	// パーサーの作成
	p := parser.NewParser()

	var results []*config.Result

	// 各ファイルを処理
	for _, filePath := range files {
		fileResults, err := processFile(p, client, cliConfig, filePath, *userAgent)
		if err != nil {
			log.Printf("Error processing file %s: %v", filePath, err)
			continue
		}
		results = append(results, fileResults...)
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

	// リクエスト設定の解析
	requestConfig, err := p.Parse(data, ext, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse request config: %w", err)
	}

	// リクエストの処理（辞書展開を含む）
	processedRequests, err := p.ProcessRequests(context.Background(), requestConfig, cliConfig.Host)
	if err != nil {
		return nil, fmt.Errorf("failed to process requests: %w", err)
	}

	var results []*config.Result

	// 各処理済みリクエストを送信
	for _, processedRequest := range processedRequests {
		// User-Agentの上書き処理
		// JSONやYAMLファイルでUser-Agentが指定されていない場合のみ、コマンドライン引数を適用
		if userAgent != "" {
			if _, exists := processedRequest.Headers["User-Agent"]; !exists {
				processedRequest.Headers["User-Agent"] = userAgent
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
				"file":      filePath,
				"timestamp": time.Now().Format(time.RFC3339),
			},
		}

		results = append(results, result)

		// Verbose出力
		if cliConfig.Verbose {
			fmt.Printf("Request: %s %s\n", processedRequest.Method, processedRequest.URL)
			fmt.Printf("Response: %d\n", response.StatusCode)
		}
	}

	return results, nil
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
	lines = append(lines, "Method,URL,StatusCode,ResponseTime,BodyLength")

	for _, result := range results {
		line := fmt.Sprintf("%s,%s,%d,%.3f,%d",
			result.Request.Method,
			result.Request.URL,
			result.Response.StatusCode,
			result.Response.Time.Total,
			len(result.Response.Body))
		lines = append(lines, line)
	}

	return []byte(strings.Join(lines, "\n")), nil
}

func formatAsTable(results []*config.Result) ([]byte, error) {
	var lines []string
	lines = append(lines, "METHOD\tURL\tSTATUS\tTIME\tSIZE")
	lines = append(lines, strings.Repeat("-", 80))

	for _, result := range results {
		line := fmt.Sprintf("%s\t%s\t%d\t%.3fs\t%d",
			result.Request.Method,
			result.Request.URL,
			result.Response.StatusCode,
			result.Response.Time.Total,
			len(result.Response.Body))
		lines = append(lines, line)
	}

	return []byte(strings.Join(lines, "\n")), nil
}
