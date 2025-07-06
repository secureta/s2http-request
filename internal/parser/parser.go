package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"github.com/secureta/s2http-request/internal/config"
	"github.com/secureta/s2http-request/pkg/functions"
	"github.com/google/uuid"
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

// Parse はファイル内容を解析してRequestConfigを返す
func (p *Parser) Parse(data []byte, fileExt string, filePath string) (*config.RequestConfig, error) {
	var requestConfig config.RequestConfig
	
	switch strings.ToLower(fileExt) {
	case ".json":
		if err := json.Unmarshal(data, &requestConfig); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &requestConfig); err != nil {
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported file format: %s", fileExt)
	}
	
	requestConfig.FilePath = filePath
	return &requestConfig, nil
}

// ProcessRequest はリクエスト設定を処理してProcessedRequestを返す
func (p *Parser) ProcessRequest(ctx context.Context, requestConfig *config.RequestConfig, baseURL string) (*config.ProcessedRequest, error) {
	// URLの構築
	fullURL := baseURL + requestConfig.Path
	
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
		body = fmt.Sprintf("%v", processedBody)
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
		// 関数呼び出しかチェック
		for key, args := range v {
			if strings.HasPrefix(key, "!") {
				funcName := key[1:] // "!" を除去
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
				
				return fn.Execute(ctx, processedArgs)
			}
		}
		
		// 通常のマップとして処理
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

// ProcessRequestsWithRequestID はRequest ID機能付きでリクエストを処理する
func (p *Parser) ProcessRequestsWithRequestID(ctx context.Context, requestConfig *config.RequestConfig, baseURL string, cliRequestIDConfig *config.RequestIDConfig) ([]*config.ProcessedRequest, error) {
	// Request IDの設定を決定（リクエスト定義ファイル > CLI設定）
	var requestIDConfig *config.RequestIDConfig
	if requestConfig.Meta != nil && requestConfig.Meta.RequestID != nil {
		requestIDConfig = requestConfig.Meta.RequestID
	} else if cliRequestIDConfig != nil {
		requestIDConfig = cliRequestIDConfig
	}

	dict := requestConfig.Dictionary
	if dict == nil || len(dict) == 0 {
		// 変数をコンテキストに設定（変数を事前に処理）
		ctxWithVars := ctx
		if requestConfig.Variables != nil {
			processedVars, err := p.processVariables(ctx, requestConfig.Variables)
			if err != nil {
				return nil, fmt.Errorf("failed to process variables: %w", err)
			}
			ctxWithVars = context.WithValue(ctx, "variables", processedVars)
		}
		pr, err := p.ProcessRequestWithRequestID(ctxWithVars, requestConfig, baseURL, requestIDConfig)
		if err != nil {
			return nil, err
		}
		return []*config.ProcessedRequest{pr}, nil
	}

	// 1つ以上のdictionaryがある場合（zip方式: 最初の配列長で回す）
	arrLen := 0
	for _, v := range dict {
		arrLen = len(v)
		break
	}

	var results []*config.ProcessedRequest
	for i := 0; i < arrLen; i++ {
		// 各ループで context に dictionary の i番目をセット
		dictVars := make(map[string]interface{})
		for k, v := range dict {
			if i < len(v) {
				dictVars[k] = v[i]
			} else {
				dictVars[k] = nil
			}
		}
		// variablesもマージ
		mergedVars := map[string]interface{}{}
		for k, v := range requestConfig.Variables {
			mergedVars[k] = v
		}
		for k, v := range dictVars {
			mergedVars[k] = v
		}
		
		// 変数を事前に処理
		processedVars, err := p.processVariables(ctx, mergedVars)
		if err != nil {
			return nil, fmt.Errorf("failed to process variables: %w", err)
		}
		ctxWithVars := context.WithValue(ctx, "variables", processedVars)
		pr, err := p.ProcessRequestWithRequestID(ctxWithVars, requestConfig, baseURL, requestIDConfig)
		if err != nil {
			return nil, err
		}
		results = append(results, pr)
	}
	return results, nil
}

// ProcessRequestWithRequestID はRequest ID機能付きでリクエストを処理する
func (p *Parser) ProcessRequestWithRequestID(ctx context.Context, requestConfig *config.RequestConfig, baseURL string, requestIDConfig *config.RequestIDConfig) (*config.ProcessedRequest, error) {
	// Request IDを生成
	var requestID string
	if requestIDConfig != nil {
		requestID = uuid.New().String()
	}

	// URLの構築
	fullURL := baseURL + requestConfig.Path
	
	// Request IDをパスに追加
	if requestIDConfig != nil && requestIDConfig.Location == config.RequestIDLocationPathHead {
		fullURL = baseURL + "/" + requestID + requestConfig.Path
	} else if requestIDConfig != nil && requestIDConfig.Location == config.RequestIDLocationPathTail {
		fullURL = baseURL + requestConfig.Path + "/" + requestID
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
		body = fmt.Sprintf("%v", processedBody)
	}
	
	return &config.ProcessedRequest{
		Method:    requestConfig.Method,
		URL:       fullURL,
		Headers:   headers,
		Body:      body,
		RequestID: requestID,
	}, nil
}

// ProcessRequests はdictionaryの要素数ぶんリクエストを展開して返す
func (p *Parser) ProcessRequests(ctx context.Context, requestConfig *config.RequestConfig, baseURL string) ([]*config.ProcessedRequest, error) {
	dict := requestConfig.Dictionary
	if dict == nil || len(dict) == 0 {
		// 変数をコンテキストに設定（変数を事前に処理）
		ctxWithVars := ctx
		if requestConfig.Variables != nil {
			processedVars, err := p.processVariables(ctx, requestConfig.Variables)
			if err != nil {
				return nil, fmt.Errorf("failed to process variables: %w", err)
			}
			ctxWithVars = context.WithValue(ctx, "variables", processedVars)
		}
		pr, err := p.ProcessRequest(ctxWithVars, requestConfig, baseURL)
		if err != nil {
			return nil, err
		}
		return []*config.ProcessedRequest{pr}, nil
	}

	// 1つ以上のdictionaryがある場合（zip方式: 最初の配列長で回す）
	arrLen := 0
	for _, v := range dict {
		arrLen = len(v)
		break
	}

	var results []*config.ProcessedRequest
	for i := 0; i < arrLen; i++ {
		// 各ループで context に dictionary の i番目をセット
		dictVars := make(map[string]interface{})
		for k, v := range dict {
			if i < len(v) {
				dictVars[k] = v[i]
			} else {
				dictVars[k] = nil
			}
		}
		// variablesもマージ
		mergedVars := map[string]interface{}{}
		for k, v := range requestConfig.Variables {
			mergedVars[k] = v
		}
		for k, v := range dictVars {
			mergedVars[k] = v
		}
		
		// 変数を事前に処理
		processedVars, err := p.processVariables(ctx, mergedVars)
		if err != nil {
			return nil, fmt.Errorf("failed to process variables: %w", err)
		}
		ctxWithVars := context.WithValue(ctx, "variables", processedVars)
		pr, err := p.ProcessRequest(ctxWithVars, requestConfig, baseURL)
		if err != nil {
			return nil, err
		}
		results = append(results, pr)
	}
	return results, nil
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