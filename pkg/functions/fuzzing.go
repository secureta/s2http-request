package functions

import (
	"context"
	"fmt"
)

// DictionaryFuzzFunction は辞書の各要素に対してテンプレートを展開する関数
type DictionaryFuzzFunction struct{}

func (f *DictionaryFuzzFunction) Name() string {
	return "dict_fuzz"
}

func (f *DictionaryFuzzFunction) Signature() string {
	return "dict_fuzz(dictionary: []string, template: string, placeholder: string) -> []string"
}

func (f *DictionaryFuzzFunction) Description() string {
	return "辞書の各要素をテンプレートのプレースホルダに置換してリクエストのバリエーションを生成します"
}

func (f *DictionaryFuzzFunction) Execute(ctx context.Context, args []interface{}) (interface{}, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, fmt.Errorf("dict_fuzz requires 2 or 3 arguments, got %d", len(args))
	}

	// 辞書の取得
	var dict []string
	switch d := args[0].(type) {
	case []interface{}:
		dict = make([]string, len(d))
		for i, item := range d {
			dict[i] = fmt.Sprintf("%v", item)
		}
	case []string:
		dict = d
	default:
		return nil, fmt.Errorf("dict_fuzz: first argument must be an array")
	}

	// テンプレートの取得
	template, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("dict_fuzz: second argument must be a string template")
	}

	// プレースホルダの取得（デフォルトは "{{FUZZ}}"）
	placeholder := "{{FUZZ}}"
	if len(args) == 3 {
		if ph, ok := args[2].(string); ok {
			placeholder = ph
		} else {
			return nil, fmt.Errorf("dict_fuzz: third argument must be a string placeholder")
		}
	}

	// 各辞書要素でテンプレートを展開
	var results []string
	for _, item := range dict {
		// プレースホルダを辞書の値で置換
		expanded := replacePlaceholder(template, placeholder, item)
		results = append(results, expanded)
	}

	return results, nil
}

// replacePlaceholder はテンプレート内のプレースホルダを値で置換する
func replacePlaceholder(template, placeholder, value string) string {
	// 簡単な文字列置換（より高度な実装も可能）
	result := template
	for i := 0; i < len(result); i++ {
		if i+len(placeholder) <= len(result) && result[i:i+len(placeholder)] == placeholder {
			result = result[:i] + value + result[i+len(placeholder):]
			i += len(value) - 1
		}
	}
	return result
}