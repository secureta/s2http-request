package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/secureta/s2http-request/internal/config"
	"github.com/secureta/s2http-request/pkg/functions"
	"gopkg.in/yaml.v3"
)

// Parser はリクエスト設定を解析し処理するためのパーサー
type Parser struct {
	registry *functions.Registry
}

// NewParser は新しいパーサーインスタンスを作成
func NewParser() *Parser {
	return &Parser{
		registry: functions.NewRegistry(),
	}
}

// ParseMultiple はファイル内容を解析して複数のRequestConfigを返す
func (p *Parser) ParseMultiple(data []byte, fileExt string, filePath string) ([]*config.RequestConfig, error) {
	var configs []*config.RequestConfig

	switch strings.ToLower(fileExt) {
	case ".json":
		var requestConfig config.RequestConfig
		if err := json.Unmarshal(data, &requestConfig); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
		requestConfig.FilePath = filePath

		// Dict設定の検証は後でまとめて行う

		configs = append(configs, &requestConfig)
	case ".jsonl":
		// JSONLファイルの場合、各行を個別のJSONとして解析
		lines := strings.Split(string(data), "\n")

		// 最初の有効なJSONオブジェクトをベースとして使用
		var requestConfig config.RequestConfig
		foundBase := false
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
				continue // 空行やコメント行をスキップ
			}

			// 最初の有効なJSONオブジェクトを解析
			if err := json.Unmarshal([]byte(line), &requestConfig); err != nil {
				continue // 解析エラーの場合はスキップ
			}

			foundBase = true
			break
		}

		if !foundBase {
			return nil, fmt.Errorf("no valid JSON object found in JSONL file")
		}

		requestConfig.FilePath = filePath

		// Dict設定の検証は後でまとめて行う

		configs = append(configs, &requestConfig)
	case ".yaml", ".yml":
		// YAMLファイルの場合、複数のドキュメントを処理
		decoder := yaml.NewDecoder(strings.NewReader(string(data)))

		for {
			var requestConfig config.RequestConfig
			err := decoder.Decode(&requestConfig)
			if err != nil {
				// EOFはエラーではなく、ドキュメントの終わりを示す
				break
			}

			requestConfig.FilePath = filePath

			// Dict設定の検証は後でまとめて行う

			configs = append(configs, &requestConfig)
		}

		if len(configs) == 0 {
			return nil, fmt.Errorf("no valid YAML documents found in file")
		}
	default:
		return nil, fmt.Errorf("unsupported file format: %s", fileExt)
	}

	// Validate all configurations and aggregate errors
	if err := p.validateAllConfigurations(configs, filePath, fileExt, string(data)); err != nil {
		return nil, err
	}

	return configs, nil
}

// Parse はファイル内容を解析してRequestConfigを返す
// 複数のドキュメントがある場合は最初のものを返す
func (p *Parser) Parse(data []byte, fileExt string, filePath string) (*config.RequestConfig, error) {
	configs, err := p.ParseMultiple(data, fileExt, filePath)
	if err != nil {
		return nil, err
	}

	if len(configs) == 0 {
		return nil, fmt.Errorf("no valid configuration found in file")
	}

	return configs[0], nil
}

// ProcessRequest はリクエスト設定を処理してProcessedRequestを返す
func (p *Parser) ProcessRequest(ctx context.Context, requestConfig *config.RequestConfig, baseURL string) (*config.ProcessedRequest, error) {
	// Pathの処理
	processedPath, err := p.processValue(ctx, requestConfig.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to process path: %w", err)
	}
	pathStr := fmt.Sprintf("%v", processedPath)

	// URLの構築
	fullURL := baseURL + pathStr

	// コンテキストにリクエストファイルのパスを設定
	ctx = context.WithValue(ctx, "requestFilePath", requestConfig.FilePath)

	// クエリパラメータの処理
	if requestConfig.Query != nil {
		if queryMap, ok := requestConfig.Query.(map[string]interface{}); ok {
			processedQuery, err := p.processMap(ctx, queryMap)
			if err != nil {
				return nil, fmt.Errorf("failed to process query: %w", err)
			}
			if len(processedQuery) > 0 {
				queryString := p.mapToQueryString(processedQuery)
				if queryString != "" {
					fullURL += "?" + queryString
				}
			}
		}
	}

	// ヘッダーの処理
	headers := make(map[string]string)
	if requestConfig.Headers != nil {
		if headersMap, ok := requestConfig.Headers.(map[string]interface{}); ok {
			processedHeaders, err := p.processMap(ctx, headersMap)
			if err != nil {
				return nil, fmt.Errorf("failed to process headers: %w", err)
			}
			for k, v := range processedHeaders {
				headers[k] = fmt.Sprintf("%v", v)
			}
		}
	}

	// ボディの処理
	var body string
	if requestConfig.Params != nil {
		if paramsMap, ok := requestConfig.Params.(map[string]interface{}); ok {
			processedParams, err := p.processMap(ctx, paramsMap)
			if err != nil {
				return nil, fmt.Errorf("failed to process params: %w", err)
			}
			body = p.mapToQueryString(processedParams)
		}
	} else if requestConfig.Body != nil {
		processedBody, err := p.processValue(ctx, requestConfig.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to process body: %w", err)
		}

		// マップ型かどうかを確認し、適切に処理
		switch v := processedBody.(type) {
		case map[string]interface{}:
			// JSONに変換
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal body to JSON: %w", err)
			}
			body = string(jsonBytes)
		case []interface{}:
			// JSONに変換
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal body to JSON: %w", err)
			}
			body = string(jsonBytes)
		default:
			// その他の型はそのまま文字列化
			body = fmt.Sprintf("%v", processedBody)
		}
	}

	return &config.ProcessedRequest{
		Method:  requestConfig.Method,
		URL:     fullURL,
		Headers: headers,
		Body:    body,
	}, nil
}

// processMap はマップの値を再帰的に処理
func (p *Parser) processMap(ctx context.Context, m map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for k, v := range m {
		processedValue, err := p.processValue(ctx, v)
		if err != nil {
			return nil, err
		}
		if processedValue != nil {
			result[k] = processedValue
		}
	}
	return result, nil
}

// processValue は値を再帰的に処理（関数呼び出しや変数参照を解決）
func (p *Parser) processValue(ctx context.Context, value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case map[string]interface{}:
		// マップに単一のキーがあり、それが関数呼び出しかどうかをチェック
		if len(v) == 1 {
			for key, args := range v {
				if strings.HasPrefix(key, "$") {
					funcName := key[1:] // "$" を除去
					fn, exists := p.registry.Get(funcName)
					if !exists {
						return nil, fmt.Errorf("unknown function: %s", funcName)
					}

					// 引数を処理
					var processedArgs []interface{}
					switch a := args.(type) {
					case []interface{}:
						for _, arg := range a {
							processedArg, err := p.processValue(ctx, arg)
							if err != nil {
								return nil, err
							}
							processedArgs = append(processedArgs, processedArg)
						}
					default:
						processedArg, err := p.processValue(ctx, a)
						if err != nil {
							return nil, err
						}
						processedArgs = []interface{}{processedArg}
					}

					// 関数を実行して結果を返す
					return fn.Execute(ctx, processedArgs)
				}
			}
		}

		// 通常のマップとして処理（各値に対して再帰的に処理）
		result := make(map[string]interface{})
		for k, val := range v {
			processedVal, err := p.processValue(ctx, val)
			if err != nil {
				return nil, err
			}
			result[k] = processedVal
		}
		return result, nil

	case []interface{}:
		var result []interface{}
		for _, item := range v {
			processedItem, err := p.processValue(ctx, item)
			if err != nil {
				return nil, err
			}
			result = append(result, processedItem)
		}
		return result, nil

	default:
		return value, nil
	}
}

// mapToQueryString はマップをクエリ文字列に変換
func (p *Parser) mapToQueryString(m map[string]interface{}) string {
	values := url.Values{}
	for k, v := range m {
		if v != nil {
			values.Add(k, fmt.Sprintf("%v", v))
		}
	}
	return values.Encode()
}

// validateAllConfigurations validates all aspects of the request configuration and aggregates errors
func (p *Parser) validateAllConfigurations(configs []*config.RequestConfig, filePath string, fileExt string, content string) error {
	errorCollection := NewErrorCollection()

	for i, requestConfig := range configs {
		// Validate dict configuration
		if requestConfig.Dict != nil {
			if err := p.validateDictWithPosition(requestConfig.Dict, filePath, fileExt, content); err != nil {
				errorCollection.Add(err)
			}
		}

		// Validate dict references in the configuration
		if err := p.validateDictReferences(requestConfig, filePath, fileExt, content, i); err != nil {
			errorCollection.Add(err)
		}
	}

	return errorCollection.ToError()
}

// validateDictReferences validates that all $dict references have corresponding dict definitions
func (p *Parser) validateDictReferences(requestConfig *config.RequestConfig, filePath string, fileExt string, content string, configIndex int) error {
	errorCollection := NewErrorCollection()

	// Get all dict references in the configuration
	dictRefs := p.findAllDictReferences(requestConfig)

	// Check if each reference has a corresponding definition
	for _, ref := range dictRefs {
		if requestConfig.Dict == nil {
			err := p.createDictReferenceError(filePath, ref.PropertyPath, ref.VariableName, "no dict variables are defined")
			errorCollection.Add(err)
			continue
		}

		if _, exists := requestConfig.Dict[ref.VariableName]; !exists {
			var availableVars []string
			for key := range requestConfig.Dict {
				availableVars = append(availableVars, key)
			}

			var message string
			if len(availableVars) > 0 {
				message = fmt.Sprintf("dict variable '%s' not found. Available variables: %v", ref.VariableName, availableVars)
			} else {
				message = fmt.Sprintf("dict variable '%s' not found. No dict variables are defined", ref.VariableName)
			}

			err := p.createDictReferenceError(filePath, ref.PropertyPath, ref.VariableName, message)
			errorCollection.Add(err)
		}
	}

	return errorCollection.ToError()
}

// DictReference represents a reference to a dict variable
type DictReference struct {
	PropertyPath string
	VariableName string
}

// findAllDictReferences finds all $dict references in a request configuration
func (p *Parser) findAllDictReferences(requestConfig *config.RequestConfig) []DictReference {
	var refs []DictReference

	// Check different parts of the configuration
	refs = append(refs, p.findDictReferencesInValue(requestConfig.Path, "path")...)
	refs = append(refs, p.findDictReferencesInValue(requestConfig.Query, "query")...)
	refs = append(refs, p.findDictReferencesInValue(requestConfig.Headers, "headers")...)
	refs = append(refs, p.findDictReferencesInValue(requestConfig.Params, "params")...)
	refs = append(refs, p.findDictReferencesInValue(requestConfig.Body, "body")...)

	return refs
}

// findDictReferencesInValue recursively finds dict references in a value
func (p *Parser) findDictReferencesInValue(value interface{}, basePath string) []DictReference {
	var refs []DictReference

	switch v := value.(type) {
	case map[string]interface{}:
		// Check if this is a dict function call
		if len(v) == 1 {
			for key, args := range v {
				if key == "$dict" {
					if varName, ok := args.(string); ok {
						refs = append(refs, DictReference{
							PropertyPath: basePath,
							VariableName: varName,
						})
					}
					return refs
				}
			}
		}

		// Recursively check all values in the map
		for k, val := range v {
			childPath := fmt.Sprintf("%s.%s", basePath, k)
			refs = append(refs, p.findDictReferencesInValue(val, childPath)...)
		}

	case []interface{}:
		// Recursively check all items in the array
		for i, item := range v {
			childPath := fmt.Sprintf("%s[%d]", basePath, i)
			refs = append(refs, p.findDictReferencesInValue(item, childPath)...)
		}
	}

	return refs
}

// createDictReferenceError creates an error for dict reference issues
func (p *Parser) createDictReferenceError(filePath string, propertyPath string, variableName string, message string) *ParseError {
	return &ParseError{
		FilePath:     filePath,
		PropertyPath: propertyPath,
		Message:      fmt.Sprintf("$dict reference '%s': %s", variableName, message),
		Level:        ErrorLevelError,
	}
}

// ProcessRequestsWithConfig はCLI設定を使用してリクエストを処理する
func (p *Parser) ProcessRequestsWithConfig(ctx context.Context, requestConfig *config.RequestConfig, baseURL string, cliConfig *config.CLIConfig) ([]*config.ProcessedRequest, error) {
	// Request IDの設定を決定（リクエスト定義ファイル > CLI設定）
	var requestIDConfig *config.RequestIDConfig
	if requestConfig.Meta != nil && requestConfig.Meta.RequestID != nil {
		requestIDConfig = requestConfig.Meta.RequestID
	} else if cliConfig != nil && cliConfig.RequestID != nil {
		requestIDConfig = cliConfig.RequestID
	}

	// 変数をコンテキストに設定（変数を事前に処理）
	ctxWithVars := ctx
	var finalVars map[string]interface{}

	// CLI変数を取得
	cliVars, hasCLIVars := ctx.Value("variables").(map[string]interface{})

	if requestConfig.Variables != nil {
		processedVars, err := p.processVariables(ctx, requestConfig.Variables)
		if err != nil {
			return nil, fmt.Errorf("failed to process variables: %w", err)
		}
		finalVars = processedVars
	} else {
		finalVars = make(map[string]interface{})
	}

	// CLI変数でファイル変数を上書き（CLI変数が優先）
	if hasCLIVars {
		for key, value := range cliVars {
			finalVars[key] = value
		}
	}

	// 最終的な変数をコンテキストに設定
	if len(finalVars) > 0 {
		ctxWithVars = context.WithValue(ctx, "variables", finalVars)
	}

	// Dict変数が存在し、実際に使用されている場合は複数リクエストを処理
	if requestConfig.Dict != nil && len(requestConfig.Dict) > 0 {
		// Dict変数が実際に使用されているかチェック
		if p.hasDictReferences(requestConfig) {
			maxCombinations := 1000 // デフォルト値
			if cliConfig != nil {
				maxCombinations = cliConfig.MaxCombinations
			}
			return p.processRequestsWithDictCombinationsAndLimit(ctxWithVars, requestConfig, baseURL, requestIDConfig, maxCombinations)
		}
	}

	// 単一リクエストの処理
	pr, err := p.ProcessRequestWithRequestID(ctxWithVars, requestConfig, baseURL, requestIDConfig)
	if err != nil {
		return nil, err
	}
	return []*config.ProcessedRequest{pr}, nil
}

// ProcessRequestsWithRequestID はRequest ID機能付きでリクエストを処理する
func (p *Parser) ProcessRequestsWithRequestID(ctx context.Context, requestConfig *config.RequestConfig, baseURL string, cliRequestIDConfig *config.RequestIDConfig) ([]*config.ProcessedRequest, error) {
	// Request IDの設定を決定（リクエスト定義ファイル > CLI設定）
	var requestIDConfig *config.RequestIDConfig
	if requestConfig.Meta != nil && requestConfig.Meta.RequestID != nil {
		requestIDConfig = requestConfig.Meta.RequestID
	} else if cliRequestIDConfig != nil {
		requestIDConfig = cliRequestIDConfig
	}

	// 変数をコンテキストに設定（変数を事前に処理）
	ctxWithVars := ctx
	var finalVars map[string]interface{}

	// CLI変数を取得
	cliVars, hasCLIVars := ctx.Value("variables").(map[string]interface{})

	if requestConfig.Variables != nil {
		processedVars, err := p.processVariables(ctx, requestConfig.Variables)
		if err != nil {
			return nil, fmt.Errorf("failed to process variables: %w", err)
		}
		finalVars = processedVars
	} else {
		finalVars = make(map[string]interface{})
	}

	// CLI変数でファイル変数を上書き（CLI変数が優先）
	if hasCLIVars {
		for key, value := range cliVars {
			finalVars[key] = value
		}
	}

	// 最終的な変数をコンテキストに設定
	if len(finalVars) > 0 {
		ctxWithVars = context.WithValue(ctx, "variables", finalVars)
	}
	pr, err := p.ProcessRequestWithRequestID(ctxWithVars, requestConfig, baseURL, requestIDConfig)
	if err != nil {
		return nil, err
	}
	return []*config.ProcessedRequest{pr}, nil
}

// ProcessRequestWithRequestID はRequest ID機能付きでリクエストを処理する
func (p *Parser) ProcessRequestWithRequestID(ctx context.Context, requestConfig *config.RequestConfig, baseURL string, requestIDConfig *config.RequestIDConfig) (*config.ProcessedRequest, error) {
	// Request IDを生成
	var requestID string
	if requestIDConfig != nil {
		requestID = uuid.New().String()
	}

	// Pathの処理
	processedPath, err := p.processValue(ctx, requestConfig.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to process path: %w", err)
	}
	pathStr := fmt.Sprintf("%v", processedPath)

	// URLの構築
	fullURL := baseURL + pathStr

	// Request IDをパスに追加
	if requestIDConfig != nil && requestIDConfig.Location == config.RequestIDLocationPathHead {
		fullURL = baseURL + "/" + requestID + pathStr
	} else if requestIDConfig != nil && requestIDConfig.Location == config.RequestIDLocationPathTail {
		fullURL = baseURL + pathStr + "/" + requestID
	}

	// コンテキストにリクエストファイルのパスを設定
	ctx = context.WithValue(ctx, "requestFilePath", requestConfig.FilePath)

	// クエリパラメータの処理
	queryParams := make(map[string]interface{})
	if requestConfig.Query != nil {
		if queryMap, ok := requestConfig.Query.(map[string]interface{}); ok {
			for k, v := range queryMap {
				queryParams[k] = v
			}
		}
	}

	// Request IDをクエリパラメータに追加
	if requestIDConfig != nil && requestIDConfig.Location == config.RequestIDLocationQuery {
		queryParams[requestIDConfig.Key] = requestID
	}

	if len(queryParams) > 0 {
		processedQuery, err := p.processMap(ctx, queryParams)
		if err != nil {
			return nil, fmt.Errorf("failed to process query: %w", err)
		}
		if len(processedQuery) > 0 {
			queryString := p.mapToQueryString(processedQuery)
			if queryString != "" {
				fullURL += "?" + queryString
			}
		}
	}

	// ヘッダーの処理
	headers := make(map[string]string)
	headerParams := make(map[string]interface{})
	if requestConfig.Headers != nil {
		if headersMap, ok := requestConfig.Headers.(map[string]interface{}); ok {
			for k, v := range headersMap {
				headerParams[k] = v
			}
		}
	}

	// Request IDをヘッダーに追加
	if requestIDConfig != nil && requestIDConfig.Location == config.RequestIDLocationHeader {
		headerParams[requestIDConfig.Key] = requestID
	}

	if len(headerParams) > 0 {
		processedHeaders, err := p.processMap(ctx, headerParams)
		if err != nil {
			return nil, fmt.Errorf("failed to process headers: %w", err)
		}
		for k, v := range processedHeaders {
			headers[k] = fmt.Sprintf("%v", v)
		}
	}

	// ボディの処理
	var body string
	if requestConfig.Params != nil {
		if paramsMap, ok := requestConfig.Params.(map[string]interface{}); ok {
			processedParams, err := p.processMap(ctx, paramsMap)
			if err != nil {
				return nil, fmt.Errorf("failed to process params: %w", err)
			}
			body = p.mapToQueryString(processedParams)
		}
	} else if requestConfig.Body != nil {
		processedBody, err := p.processValue(ctx, requestConfig.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to process body: %w", err)
		}

		// マップ型かどうかを確認し、適切に処理
		switch v := processedBody.(type) {
		case map[string]interface{}:
			// JSONに変換
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal body to JSON: %w", err)
			}
			body = string(jsonBytes)
		case []interface{}:
			// JSONに変換
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal body to JSON: %w", err)
			}
			body = string(jsonBytes)
		default:
			// その他の型はそのまま文字列化
			body = fmt.Sprintf("%v", processedBody)
		}
	}

	return &config.ProcessedRequest{
		Method:    requestConfig.Method,
		URL:       fullURL,
		Headers:   headers,
		Body:      body,
		RequestID: requestID,
	}, nil
}

// ProcessRequests はリクエストを処理して返す
func (p *Parser) ProcessRequests(ctx context.Context, requestConfig *config.RequestConfig, baseURL string) ([]*config.ProcessedRequest, error) {
	// 変数をコンテキストに設定（変数を事前に処理）
	ctxWithVars := ctx
	var finalVars map[string]interface{}

	// CLI変数を取得
	cliVars, hasCLIVars := ctx.Value("variables").(map[string]interface{})

	if requestConfig.Variables != nil {
		processedVars, err := p.processVariables(ctx, requestConfig.Variables)
		if err != nil {
			return nil, fmt.Errorf("failed to process variables: %w", err)
		}
		finalVars = processedVars
	} else {
		finalVars = make(map[string]interface{})
	}

	// CLI変数でファイル変数を上書き（CLI変数が優先）
	if hasCLIVars {
		for key, value := range cliVars {
			finalVars[key] = value
		}
	}

	// 最終的な変数をコンテキストに設定
	if len(finalVars) > 0 {
		ctxWithVars = context.WithValue(ctx, "variables", finalVars)
	}

	// Dict組み合わせの検出と処理
	if requestConfig.Dict != nil && len(requestConfig.Dict) > 0 {
		// Dict変数が実際に使用されているかチェック
		if p.hasDictReferences(requestConfig) {
			return p.processRequestsWithDictCombinations(ctxWithVars, requestConfig, baseURL)
		}
	}

	// Dict組み合わせがない場合は単一リクエストを処理
	pr, err := p.ProcessRequest(ctxWithVars, requestConfig, baseURL)
	if err != nil {
		return nil, err
	}
	return []*config.ProcessedRequest{pr}, nil
}

// validateDict はdict設定を検証する
func (p *Parser) validateDict(dict map[string][]interface{}) error {
	return p.validateDictWithPosition(dict, "", "", "")
}

// validateDictWithPosition はdict設定を位置情報付きで検証する
func (p *Parser) validateDictWithPosition(dict map[string][]interface{}, filePath string, fileExt string, content string) error {
	if dict == nil {
		return nil
	}

	var tracker *PositionTracker
	if content != "" {
		tracker = NewPositionTracker(filePath, []byte(content))
	}

	errorCollection := NewErrorCollection()

	for key, array := range dict {
		propertyPath := fmt.Sprintf("dict.%s", key)

		// Get position information
		var position *PositionInfo
		if tracker != nil {
			position = tracker.GetPosition(propertyPath, fileExt)
		}

		// Validate array is not nil or empty
		if array == nil {
			err := p.createDictValidationError(filePath, position, propertyPath, key, "value must be an array, got null")
			errorCollection.Add(err)
			continue
		}

		if len(array) == 0 {
			err := p.createDictValidationError(filePath, position, propertyPath, key, "array cannot be empty")
			errorCollection.Add(err)
			continue
		}

		// Validate array elements are primitive values
		for i, element := range array {
			elementPath := fmt.Sprintf("%s[%d]", propertyPath, i)
			if !p.isPrimitiveValue(element) {
				err := p.createDictValidationError(filePath, position, elementPath, key,
					fmt.Sprintf("array element at index %d must be a primitive value (string, number, boolean), got %T", i, element))
				errorCollection.Add(err)
			}
		}
	}

	return errorCollection.ToError()
}

// createDictValidationError creates a DictValidationError with position information
func (p *Parser) createDictValidationError(filePath string, position *PositionInfo, propertyPath string, dictKey string, message string) *DictValidationError {
	var lineNumber, columnNumber int
	if position != nil {
		lineNumber = position.Line
		columnNumber = position.Column
	}

	return &DictValidationError{
		ParseError: &ParseError{
			FilePath:     filePath,
			LineNumber:   lineNumber,
			ColumnNumber: columnNumber,
			PropertyPath: propertyPath,
			Message:      message,
			Level:        ErrorLevelError,
		},
		DictKey: dictKey,
	}
}

// isPrimitiveValue checks if a value is a primitive type (string, number, boolean)
func (p *Parser) isPrimitiveValue(value interface{}) bool {
	switch value.(type) {
	case string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		return true
	default:
		return false
	}
}

// generateDictCombinations はdict配列の直積（全組み合わせ）を生成する
func (p *Parser) generateDictCombinations(dict map[string][]interface{}) []map[string]interface{} {
	return p.generateDictCombinationsWithLimit(dict, 0) // 0 means no limit
}

// generateDictCombinationsWithLimit はdict配列の直積を制限付きで生成する
func (p *Parser) generateDictCombinationsWithLimit(dict map[string][]interface{}, maxCombinations int) []map[string]interface{} {
	if len(dict) == 0 {
		return []map[string]interface{}{{}}
	}

	// 組み合わせ数を事前に計算
	totalCombinations := 1
	for _, array := range dict {
		if len(array) == 0 {
			return []map[string]interface{}{} // 空配列があれば組み合わせなし
		}
		totalCombinations *= len(array)

		// オーバーフロー防止と早期終了
		if maxCombinations > 0 && totalCombinations > maxCombinations {
			return nil // 制限を超える場合はnilを返す
		}
	}

	// キーと配列を分離（メモリ効率のため事前にソート）
	keys := make([]string, 0, len(dict))
	for key := range dict {
		keys = append(keys, key)
	}

	// 一貫した順序のためにキーをソート
	for i := 0; i < len(keys)-1; i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	arrays := make([][]interface{}, len(keys))
	for i, key := range keys {
		arrays[i] = dict[key]
	}

	// 直積を計算
	combinations := make([]map[string]interface{}, 0, totalCombinations)
	p.generateCombinationsRecursive(keys, arrays, 0, make(map[string]interface{}), &combinations)

	return combinations
}

// generateCombinationsRecursive は再帰的に組み合わせを生成する
func (p *Parser) generateCombinationsRecursive(keys []string, arrays [][]interface{}, index int, current map[string]interface{}, result *[]map[string]interface{}) {
	if index == len(keys) {
		// 現在の組み合わせをコピーして結果に追加
		combination := make(map[string]interface{})
		for k, v := range current {
			combination[k] = v
		}
		*result = append(*result, combination)
		return
	}

	// 現在のインデックスの配列の各要素について再帰
	for _, value := range arrays[index] {
		current[keys[index]] = value
		p.generateCombinationsRecursive(keys, arrays, index+1, current, result)
	}
}

// processVariables は変数を依存関係を考慮して処理
func (p *Parser) processVariables(ctx context.Context, variables map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	processed := make(map[string]bool)

	// 依存関係を解決するために複数回処理
	maxIterations := len(variables) * 2 // 循環参照防止
	for iteration := 0; iteration < maxIterations; iteration++ {
		progressMade := false

		for varName, varValue := range variables {
			if processed[varName] {
				continue
			}

			// 一時的にこの変数以外をコンテキストに設定
			tempVars := make(map[string]interface{})
			for k, v := range result {
				tempVars[k] = v
			}
			tempCtx := context.WithValue(ctx, "variables", tempVars)

			// 変数値を処理
			processedValue, err := p.processValue(tempCtx, varValue)
			if err != nil {
				// この変数がまだ処理できない場合は次の反復で試す
				continue
			}

			result[varName] = processedValue
			processed[varName] = true
			progressMade = true
		}

		// 全ての変数が処理された場合
		if len(result) == len(variables) {
			break
		}

		// 進歩がない場合は循環参照または未定義変数
		if !progressMade {
			// 未処理の変数を特定
			var unprocessed []string
			for varName := range variables {
				if !processed[varName] {
					unprocessed = append(unprocessed, varName)
				}
			}
			return nil, fmt.Errorf("unable to resolve variables (possible circular dependency or undefined reference): %v", unprocessed)
		}
	}

	return result, nil
}

// hasDictReferences checks if the request configuration contains any $dict references
func (p *Parser) hasDictReferences(requestConfig *config.RequestConfig) bool {
	return p.hasDict(requestConfig.Path) ||
		p.hasDict(requestConfig.Query) ||
		p.hasDict(requestConfig.Headers) ||
		p.hasDict(requestConfig.Params) ||
		p.hasDict(requestConfig.Body)
}

// hasDict recursively checks if a value contains $dict references
func (p *Parser) hasDict(value interface{}) bool {
	switch v := value.(type) {
	case map[string]interface{}:
		// Check if this is a dict function call
		if len(v) == 1 {
			for key := range v {
				if key == "$dict" {
					return true
				}
			}
		}
		// Recursively check all values in the map
		for _, val := range v {
			if p.hasDict(val) {
				return true
			}
		}
	case []interface{}:
		// Recursively check all items in the array
		for _, item := range v {
			if p.hasDict(item) {
				return true
			}
		}
	}
	return false
}

// processRequestsWithDictCombinationsAndLimit processes requests with dict combinations and limit
func (p *Parser) processRequestsWithDictCombinationsAndLimit(ctx context.Context, requestConfig *config.RequestConfig, baseURL string, requestIDConfig *config.RequestIDConfig, maxCombinations int) ([]*config.ProcessedRequest, error) {
	// Generate dict combinations with limit
	combinations := p.generateDictCombinationsWithLimit(requestConfig.Dict, maxCombinations)
	if combinations == nil {
		return nil, fmt.Errorf("dict combinations exceed maximum limit of %d", maxCombinations)
	}
	if len(combinations) == 0 {
		// No combinations, process as single request
		pr, err := p.ProcessRequestWithRequestID(ctx, requestConfig, baseURL, requestIDConfig)
		if err != nil {
			return nil, err
		}
		return []*config.ProcessedRequest{pr}, nil
	}

	// Process each combination
	var results []*config.ProcessedRequest
	for _, combination := range combinations {
		// Create context with dict variables for this combination
		ctxWithDict := context.WithValue(ctx, "dict", combination)

		// Process the request with this combination
		pr, err := p.ProcessRequestWithRequestID(ctxWithDict, requestConfig, baseURL, requestIDConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to process request with dict combination %v: %w", combination, err)
		}
		results = append(results, pr)
	}

	return results, nil
}

// processRequestsWithDictCombinations processes requests with dict combinations
func (p *Parser) processRequestsWithDictCombinations(ctx context.Context, requestConfig *config.RequestConfig, baseURL string) ([]*config.ProcessedRequest, error) {
	// Generate dict combinations
	combinations := p.generateDictCombinations(requestConfig.Dict)
	if len(combinations) == 0 {
		// No combinations, process as single request
		pr, err := p.ProcessRequest(ctx, requestConfig, baseURL)
		if err != nil {
			return nil, err
		}
		return []*config.ProcessedRequest{pr}, nil
	}

	// Process each combination
	var results []*config.ProcessedRequest
	for _, combination := range combinations {
		// Create context with dict variables for this combination
		ctxWithDict := context.WithValue(ctx, "dict", combination)

		// Process the request with this combination
		pr, err := p.ProcessRequest(ctxWithDict, requestConfig, baseURL)
		if err != nil {
			return nil, fmt.Errorf("failed to process request with dict combination %v: %w", combination, err)
		}

		results = append(results, pr)
	}

	return results, nil
}
