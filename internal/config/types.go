package config

import "time"

// RequestIDLocation はRequest IDの配置場所を表す列挙型
type RequestIDLocation string

const (
	RequestIDLocationPathHead RequestIDLocation = "path_head"
	RequestIDLocationPathTail RequestIDLocation = "path_tail"
	RequestIDLocationQuery    RequestIDLocation = "query"
	RequestIDLocationHeader   RequestIDLocation = "header"
)

// RequestIDConfig はRequest IDの設定を表す構造体
type RequestIDConfig struct {
	Location RequestIDLocation `json:"location" yaml:"location"`
	Key      string            `json:"key,omitempty" yaml:"key,omitempty"` // query/headerの場合のキー名
}

// MetaConfig はリクエストのメタデータを表す構造体
type MetaConfig struct {
	RequestID *RequestIDConfig `json:"request-id,omitempty" yaml:"request-id,omitempty"`
}

// RequestConfig はリクエスト設定を表す構造体
type RequestConfig struct {
	Method    string                 `json:"method" yaml:"method"`
	Path      interface{}            `json:"path" yaml:"path"`
	Query     interface{}            `json:"query,omitempty" yaml:"query,omitempty"`
	Headers   interface{}            `json:"headers,omitempty" yaml:"headers,omitempty"`
	Params    interface{}            `json:"params,omitempty" yaml:"params,omitempty"`
	Body      interface{}            `json:"body,omitempty" yaml:"body,omitempty"`
	Variables map[string]interface{} `json:"variables,omitempty" yaml:"variables,omitempty"`
	Meta      *MetaConfig            `json:"meta,omitempty" yaml:"meta,omitempty"`
	FilePath  string                 `json:"-" yaml:"-"`
}

// KeyValue は配列形式のパラメータを表す構造体
type KeyValue struct {
	Key   interface{} `json:"key" yaml:"key"`
	Value interface{} `json:"value" yaml:"value"`
}

// ProcessedRequest は処理済みリクエストを表す構造体
type ProcessedRequest struct {
	Method    string
	URL       string
	Headers   map[string]string
	Body      string
	RequestID string // Request IDを追加
}

// ResponseTiming はレスポンス時間の詳細を表す構造体
type ResponseTiming struct {
	Total   float64 `json:"total"`
	DNS     float64 `json:"dns"`
	Connect float64 `json:"connect"`
	SSL     float64 `json:"ssl"`
	Send    float64 `json:"send"`
	Wait    float64 `json:"wait"`
	Receive float64 `json:"receive"`
}

// ResponseData はHTTPレスポンスデータを表す構造体
type ResponseData struct {
	StatusCode int                 `json:"status_code"`
	Headers    map[string][]string `json:"headers"`
	Body       string              `json:"body"`
	Time       ResponseTiming      `json:"time"`
}

// Result は最終的な結果を表す構造体
type Result struct {
	Request  ProcessedRequest       `json:"request"`
	Response ResponseData           `json:"response"`
	Metadata map[string]interface{} `json:"metadata"`
}

// OutputFormat は出力フォーマットを表す列挙型
type OutputFormat string

const (
	OutputFormatTable OutputFormat = "table"
	OutputFormatCSV   OutputFormat = "csv"
	OutputFormatJSON  OutputFormat = "json"
)

// CLIConfig はCLIオプションを表す構造体
type CLIConfig struct {
	Host      string
	Timeout   time.Duration
	Retry     int
	Proxy     string
	Verbose   bool
	Output    string
	Format    OutputFormat
	Files     []string
	RequestID *RequestIDConfig // Request ID設定を追加
}