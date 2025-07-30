package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
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

// varFlags is a custom flag type for handling multiple --var flags
type varFlags map[string]interface{}

func (v varFlags) String() string {
	return fmt.Sprintf("%v", map[string]interface{}(v))
}

func (v varFlags) Set(value string) error {
	parts := strings.SplitN(value, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid format, expected 'key=value'")
	}
	
	key := strings.TrimSpace(parts[0])
	val := strings.TrimSpace(parts[1])
	
	if key == "" {
		return fmt.Errorf("variable key cannot be empty")
	}
	
	// Try to parse the value as JSON first for complex values
	var parsedValue interface{}
	if err := json.Unmarshal([]byte(val), &parsedValue); err == nil {
		v[key] = parsedValue
	} else {
		// If JSON parsing fails, try to parse as number or boolean
		if intVal, err := strconv.Atoi(val); err == nil {
			v[key] = intVal
		} else if floatVal, err := strconv.ParseFloat(val, 64); err == nil {
			v[key] = floatVal
		} else if boolVal, err := strconv.ParseBool(val); err == nil {
			v[key] = boolVal
		} else {
			// Fallback to string
			v[key] = val
		}
	}
	
	return nil
}

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
func processStdin(p *parser.Parser, client *http.Client, cliConfig *config.CLIConfig, userAgent string, variables map[string]interface{}) ([]*config.Result, error) {
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
		// Create context with variables
		ctx := context.Background()
		if len(variables) > 0 {
			ctx = context.WithValue(ctx, "variables", variables)
		}
		
		// Process requests with Request ID
		processedRequests, err := p.ProcessRequestsWithRequestID(ctx, requestConfig, cliConfig.Host, cliConfig.RequestID)
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

			var response *config.ResponseData
			if cliConfig.Retry > 0 {
				response, err = client.SendRequestWithRetry(ctx, processedRequest, cliConfig.Retry)
			} else {
				response, err = client.SendRequest(ctx, processedRequest)
			}

			// Always cancel the context when done with this request
			cancel()

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

func handleGenerateCommand() {
	// Create a new flag set for the generate subcommand
	generateCmd := flag.NewFlagSet("generate", flag.ExitOnError)
	
	var (
		output        = generateCmd.String("output", "", "Output file path")
		format       = generateCmd.String("format", "yaml", "Output format (yaml, json, jsonl)")
		showVersion  = generateCmd.Bool("version", false, "Show version")
		vars         = make(varFlags)
	)
	
	generateCmd.Var(&vars, "var", "Override variable value (key=value). Can be specified multiple times.")
	
	// Parse arguments starting from position 2 (after "generate")
	generateCmd.Parse(os.Args[2:])
	
	if *showVersion {
		fmt.Printf("s2req version %s\n", version)
		return
	}
	
	files := generateCmd.Args()
	
	// Check if we should read from stdin
	readFromStdin := len(files) == 0 || (len(files) == 1 && files[0] == "-")
	
	if !readFromStdin && len(files) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: %s generate [options] <request-file>... or provide input via stdin\n", os.Args[0])
		generateCmd.PrintDefaults()
		os.Exit(1)
	}
	
	// パーサーの作成
	p := parser.NewParser()
	
	var generatedRequests []*config.RequestConfig
	
	if readFromStdin {
		// Process stdin input
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalf("Failed to read from stdin: %v", err)
		}
		
		// Detect the format of the input
		fileFormat := detectFormat(data)
		
		// Parse the input
		requestConfigs, err := p.ParseMultiple(data, fileFormat, "stdin")
		if err != nil {
			log.Fatalf("Failed to parse stdin input: %v", err)
		}
		
		generatedRequests = append(generatedRequests, requestConfigs...)
	} else {
		// 各ファイルを処理
		for _, filePath := range files {
			// ファイルの読み込み
			data, err := os.ReadFile(filePath)
			if err != nil {
				log.Printf("Failed to read file %s: %v", filePath, err)
				continue
			}
			
			// ファイル拡張子の取得
			ext := filepath.Ext(filePath)
			
			// リクエスト設定の解析（複数のドキュメントに対応）
			requestConfigs, err := p.ParseMultiple(data, ext, filePath)
			if err != nil {
				log.Printf("Failed to parse request config in %s: %v", filePath, err)
				continue
			}
			
			generatedRequests = append(generatedRequests, requestConfigs...)
		}
	}
	
	// Generate all variable combinations and process requests
	variableCombinations := generateVariableCombinations(map[string]interface{}(vars))
	
	processedRequests := make([]*config.RequestConfig, 0, len(generatedRequests)*len(variableCombinations))
	
	for _, requestConfig := range generatedRequests {
		for _, variableSet := range variableCombinations {
			// Create context with this specific variable combination
			ctx := context.Background()
			if len(variableSet) > 0 {
				ctx = context.WithValue(ctx, "variables", variableSet)
			}
			
			// Process the request to resolve all variables and functions
			processedRequest, err := processRequestForGenerate(ctx, p, requestConfig)
			if err != nil {
				log.Printf("Failed to process request: %v", err)
				continue
			}
			
			processedRequests = append(processedRequests, processedRequest)
		}
	}
	
	// Output the generated requests
	var output_data []byte
	var err error
	
	switch *format {
	case "json":
		output_data, err = json.MarshalIndent(processedRequests, "", "  ")
	case "jsonl":
		var jsonlOutput strings.Builder
		for _, req := range processedRequests {
			jsonBytes, jsonErr := json.Marshal(req)
			if jsonErr != nil {
				err = jsonErr
				break
			}
			jsonlOutput.Write(jsonBytes)
			jsonlOutput.WriteString("\n")
		}
		output_data = []byte(jsonlOutput.String())
	case "yaml":
		if len(processedRequests) == 1 {
			output_data, err = yaml.Marshal(processedRequests[0])
		} else {
			// For multiple requests, output as YAML stream
			var yamlOutput strings.Builder
			for i, req := range processedRequests {
				if i > 0 {
					yamlOutput.WriteString("---\n")
				}
				yamlBytes, yamlErr := yaml.Marshal(req)
				if yamlErr != nil {
					err = yamlErr
					break
				}
				yamlOutput.Write(yamlBytes)
			}
			output_data = []byte(yamlOutput.String())
		}
	default:
		// Default to yaml as specified in requirements
		if len(processedRequests) == 1 {
			output_data, err = yaml.Marshal(processedRequests[0])
		} else {
			var yamlOutput strings.Builder
			for i, req := range processedRequests {
				if i > 0 {
					yamlOutput.WriteString("---\n")
				}
				yamlBytes, yamlErr := yaml.Marshal(req)
				if yamlErr != nil {
					err = yamlErr
					break
				}
				yamlOutput.Write(yamlBytes)
			}
			output_data = []byte(yamlOutput.String())
		}
	}
	
	if err != nil {
		log.Fatalf("Failed to format output: %v", err)
	}
	
	if *output != "" {
		err = os.WriteFile(*output, output_data, 0644)
		if err != nil {
			log.Fatalf("Failed to write output file: %v", err)
		}
	} else {
		fmt.Print(string(output_data))
	}
}

func processRequestForGenerate(ctx context.Context, p *parser.Parser, requestConfig *config.RequestConfig) (*config.RequestConfig, error) {
	// Use the existing ProcessRequestsWithRequestID method to handle the variable processing
	// This method already handles variable overrides correctly
	processedRequests, err := p.ProcessRequestsWithRequestID(ctx, requestConfig, "", nil)
	if err != nil {
		return nil, err
	}
	
	if len(processedRequests) != 1 {
		return nil, fmt.Errorf("expected 1 processed request, got %d", len(processedRequests))
	}
	
	processedRequest := processedRequests[0]
	
	// Convert back to RequestConfig format for output
	// Parse the URL to extract path and query parts
	parsedURL, err := url.Parse(processedRequest.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse processed URL: %w", err)
	}
	
	// Create the processed config
	processedConfig := &config.RequestConfig{
		Method:   processedRequest.Method,
		Path:     parsedURL.Path,
		Headers:  convertHeadersToInterface(processedRequest.Headers),
		Body:     convertBodyToInterface(processedRequest.Body),
		FilePath: requestConfig.FilePath,
	}
	
	// Convert query parameters back to interface{}
	if len(parsedURL.RawQuery) > 0 {
		queryParams := make(map[string]interface{})
		for key, values := range parsedURL.Query() {
			if len(values) == 1 {
				queryParams[key] = values[0]
			} else {
				queryParams[key] = values
			}
		}
		processedConfig.Query = queryParams
	}
	
	return processedConfig, nil
}

func convertHeadersToInterface(headers map[string]string) interface{} {
	if len(headers) == 0 {
		return nil
	}
	
	result := make(map[string]interface{})
	for k, v := range headers {
		result[k] = v
	}
	return result
}

func convertBodyToInterface(body string) interface{} {
	if body == "" {
		return nil
	}
	
	// Try to parse as JSON first
	var jsonObj interface{}
	if err := json.Unmarshal([]byte(body), &jsonObj); err == nil {
		return jsonObj
	}
	
	// Return as string if not JSON
	return body
}

// generateVariableCombinations creates all possible combinations of variables
// If a variable is an array, it expands to multiple combinations
func generateVariableCombinations(vars map[string]interface{}) []map[string]interface{} {
	if len(vars) == 0 {
		return []map[string]interface{}{make(map[string]interface{})}
	}
	
	// Separate array variables from scalar variables
	arrayVars := make(map[string][]interface{})
	scalarVars := make(map[string]interface{})
	
	for key, value := range vars {
		if arr, ok := value.([]interface{}); ok {
			arrayVars[key] = arr
		} else {
			scalarVars[key] = value
		}
	}
	
	// If no array variables, return single combination
	if len(arrayVars) == 0 {
		return []map[string]interface{}{vars}
	}
	
	// Generate cartesian product of array variables
	arrayKeys := make([]string, 0, len(arrayVars))
	for key := range arrayVars {
		arrayKeys = append(arrayKeys, key)
	}
	
	combinations := generateCartesianProduct(arrayKeys, arrayVars, 0, make(map[string]interface{}))
	
	// Add scalar variables to each combination
	result := make([]map[string]interface{}, 0, len(combinations))
	for _, combination := range combinations {
		finalCombination := make(map[string]interface{})
		
		// Add scalar variables
		for key, value := range scalarVars {
			finalCombination[key] = value
		}
		
		// Add array variable values for this combination
		for key, value := range combination {
			finalCombination[key] = value
		}
		
		result = append(result, finalCombination)
	}
	
	return result
}

// generateCartesianProduct recursively generates all combinations of array variables
func generateCartesianProduct(keys []string, arrays map[string][]interface{}, index int, current map[string]interface{}) []map[string]interface{} {
	if index >= len(keys) {
		// Base case: create a copy of current combination
		result := make(map[string]interface{})
		for k, v := range current {
			result[k] = v
		}
		return []map[string]interface{}{result}
	}
	
	key := keys[index]
	array := arrays[key]
	
	var allCombinations []map[string]interface{}
	
	for _, value := range array {
		// Set current key to this value
		current[key] = value
		
		// Recursively generate combinations for remaining keys
		combinations := generateCartesianProduct(keys, arrays, index+1, current)
		allCombinations = append(allCombinations, combinations...)
	}
	
	return allCombinations
}

func handleValidateCommand() {
	// Create a new flag set for the validate subcommand
	validateCmd := flag.NewFlagSet("validate", flag.ExitOnError)
	
	var (
		verbose     = validateCmd.Bool("verbose", false, "Verbose output")
		showVersion = validateCmd.Bool("version", false, "Show version")
	)
	
	// Parse arguments starting from position 2 (after "validate")
	validateCmd.Parse(os.Args[2:])
	
	if *showVersion {
		fmt.Printf("s2req version %s\n", version)
		return
	}
	
	files := validateCmd.Args()
	
	// Check if we should read from stdin
	readFromStdin := len(files) == 0 || (len(files) == 1 && files[0] == "-")
	
	if !readFromStdin && len(files) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: %s validate [options] <request-file>... or provide input via stdin\n", os.Args[0])
		validateCmd.PrintDefaults()
		os.Exit(1)
	}
	
	// パーサーの作成
	p := parser.NewParser()
	
	var validationErrors []ValidationError
	totalFiles := 0
	
	if readFromStdin {
		// Validate stdin input
		err := validateStdinInput(p, *verbose)
		if err != nil {
			validationErrors = append(validationErrors, ValidationError{
				File:  "stdin",
				Error: err,
			})
		}
		totalFiles = 1
	} else {
		// Validate each file
		for _, filePath := range files {
			totalFiles++
			err := validateFile(p, filePath, *verbose)
			if err != nil {
				validationErrors = append(validationErrors, ValidationError{
					File:  filePath,
					Error: err,
				})
			}
		}
	}
	
	// Report results
	if len(validationErrors) == 0 {
		if *verbose {
			fmt.Printf("✓ All %d file(s) are valid\n", totalFiles)
		} else {
			fmt.Println("✓ Valid")
		}
		os.Exit(0)
	} else {
		fmt.Fprintf(os.Stderr, "✗ Found %d validation error(s) in %d file(s):\n", len(validationErrors), totalFiles)
		for _, ve := range validationErrors {
			fmt.Fprintf(os.Stderr, "  %s: %v\n", ve.File, ve.Error)
		}
		os.Exit(1)
	}
}

type ValidationError struct {
	File  string
	Error error
}

func validateFile(p *parser.Parser, filePath string, verbose bool) error {
	if verbose {
		fmt.Printf("Validating %s...\n", filePath)
	}
	
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	
	// Get file extension for format detection
	ext := filepath.Ext(filePath)
	
	// Parse the file
	_, err = p.ParseMultiple(data, ext, filePath)
	if err != nil {
		return fmt.Errorf("parsing failed: %w", err)
	}
	
	if verbose {
		fmt.Printf("  ✓ %s is valid\n", filePath)
	}
	
	return nil
}

func validateStdinInput(p *parser.Parser, verbose bool) error {
	if verbose {
		fmt.Println("Validating stdin input...")
	}
	
	// Read from stdin
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read from stdin: %w", err)
	}
	
	// Detect format
	format := detectFormat(data)
	
	// Parse the input
	_, err = p.ParseMultiple(data, format, "stdin")
	if err != nil {
		return fmt.Errorf("parsing failed: %w", err)
	}
	
	if verbose {
		fmt.Println("  ✓ stdin input is valid")
	}
	
	return nil
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "generate" {
		// Handle generate subcommand
		handleGenerateCommand()
		return
	}
	
	if len(os.Args) > 1 && os.Args[1] == "validate" {
		// Handle validate subcommand
		handleValidateCommand()
		return
	}

	// Handle main command (no variable override support)
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
		stdinResults, err := processStdin(p, client, cliConfig, *userAgent, nil)
		if err != nil {
			log.Fatalf("Error processing stdin: %v", err)
		}
		results = append(results, stdinResults...)
	} else {
		// 各ファイルを処理
		for _, filePath := range files {
			fileResults, err := processFile(p, client, cliConfig, filePath, *userAgent, nil)
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

func processFile(p *parser.Parser, client *http.Client, cliConfig *config.CLIConfig, filePath string, userAgent string, variables map[string]interface{}) ([]*config.Result, error) {
	// ファイルの読み込み
	data, err := os.ReadFile(filePath)
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
		// Create context with variables
		ctx := context.Background()
		if len(variables) > 0 {
			ctx = context.WithValue(ctx, "variables", variables)
		}
		
		// リクエストの処理（辞書展開を含む）
		processedRequests, err := p.ProcessRequestsWithRequestID(ctx, requestConfig, cliConfig.Host, cliConfig.RequestID)
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

			var response *config.ResponseData
			if cliConfig.Retry > 0 {
				response, err = client.SendRequestWithRetry(ctx, processedRequest, cliConfig.Retry)
			} else {
				response, err = client.SendRequest(ctx, processedRequest)
			}

			// Always cancel the context when done with this request
			cancel()

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
		return os.WriteFile(cliConfig.Output, output, 0644)
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
