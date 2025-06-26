package functions

import (
	"context"
	"fmt"
)

// Function は組み込み関数のインターフェース
type Function interface {
	Name() string
	Execute(ctx context.Context, args []interface{}) (interface{}, error)
	Signature() string
	Description() string
}

// Registry は組み込み関数のレジストリ
type Registry struct {
	functions map[string]Function
}

// NewRegistry は新しいレジストリを作成
func NewRegistry() *Registry {
	registry := &Registry{
		functions: make(map[string]Function),
	}
	
	// 組み込み関数を登録
	registry.registerBuiltinFunctions()
	
	return registry
}

// registerBuiltinFunctions は組み込み関数を登録
func (r *Registry) registerBuiltinFunctions() {
	// 変数操作関数
	r.functions["var"] = &VarFunction{}
	r.functions["concat"] = &ConcatFunction{}
	r.functions["join"] = &JoinFunction{}
	
	// エンコーディング関数
	r.functions["url_encode"] = &URLEncodeFunction{}
	r.functions["url_decode"] = &URLDecodeFunction{}
	r.functions["base64_encode"] = &Base64EncodeFunction{}
	r.functions["base64_decode"] = &Base64DecodeFunction{}
	r.functions["html_encode"] = &HTMLEncodeFunction{}
	r.functions["html_decode"] = &HTMLDecodeFunction{}
	
	// ランダム生成関数
	r.functions["random"] = &RandomFunction{}
	r.functions["random_string"] = &RandomStringFunction{}
	r.functions["uuid"] = &UUIDFunction{}
	
	// 時間関数
	r.functions["timestamp"] = &TimestampFunction{}
	r.functions["date"] = &DateFunction{}
	r.functions["time"] = &TimeFunction{}
	
	// WAF回避関数
	r.functions["double_encode"] = &DoubleEncodeFunction{}
	r.functions["unicode_encode"] = &UnicodeEncodeFunction{}
	r.functions["case_variation"] = &CaseVariationFunction{}
	
	// 辞書操作関数
	r.functions["dict_load"] = &DictionaryLoadFunction{}
	r.functions["dict_random"] = &DictionaryRandomFunction{}
	r.functions["dict_get"] = &DictionaryGetFunction{}
	
	// 配列操作関数
	r.functions["concat_arrays"] = &ConcatArraysFunction{}
	
	// Fuzzing関数
	r.functions["dict_fuzz"] = &DictionaryFuzzFunction{}
}

// Get は関数を取得
func (r *Registry) Get(name string) (Function, bool) {
	fn, exists := r.functions[name]
	return fn, exists
}

// Execute は関数を実行
func (r *Registry) Execute(ctx context.Context, name string, args []interface{}) (interface{}, error) {
	fn, exists := r.functions[name]
	if !exists {
		return nil, fmt.Errorf("unknown function: %s", name)
	}
	
	return fn.Execute(ctx, args)
}

// List は登録されている関数名の一覧を取得
func (r *Registry) List() []string {
	var names []string
	for name := range r.functions {
		names = append(names, name)
	}
	return names
}

// GetFunctionInfo は関数の詳細情報を取得
func (r *Registry) GetFunctionInfo() []FunctionInfo {
	var infos []FunctionInfo
	for _, fn := range r.functions {
		infos = append(infos, FunctionInfo{
			Name:        fn.Name(),
			Signature:   fn.Signature(),
			Description: fn.Description(),
		})
	}
	return infos
}

// FunctionInfo は関数の情報を格納する構造体
type FunctionInfo struct {
	Name        string
	Signature   string
	Description string
}