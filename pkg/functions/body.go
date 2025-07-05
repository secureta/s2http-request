package functions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/url"
	"strings"
)

// FormFunction はフォームデータエンコーディング関数
type FormFunction struct{}

func (f *FormFunction) Name() string {
	return "form"
}

func (f *FormFunction) Signature() string {
	return "!form <map>"
}

func (f *FormFunction) Description() string {
	return "マップをapplication/x-www-form-urlencodedフォーマットに変換します"
}

func (f *FormFunction) Execute(ctx context.Context, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("form function expects 1 argument, got %d", len(args))
	}

	input, ok := args[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("form function expects map[string]interface{} argument")
	}

	values := url.Values{}
	for k, v := range input {
		if v != nil {
			values.Add(k, fmt.Sprintf("%v", v))
		}
	}

	return values.Encode(), nil
}

// JSONFunction はJSONエンコーディング関数
type JSONFunction struct{}

func (f *JSONFunction) Name() string {
	return "json"
}

func (f *JSONFunction) Signature() string {
	return "!json {value: <value>, space?: <space>}"
}

func (f *JSONFunction) Description() string {
	return "値をJSON文字列に変換します。{value: <value>, space?: <space>} の形式で引数を指定します"
}

func (f *JSONFunction) Execute(ctx context.Context, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("json function expects 1 argument, got %d", len(args))
	}

	// 引数はマップである必要がある
	argMap, ok := args[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("json function expects map[string]interface{} argument")
	}

	// valueフィールドが必須
	value, hasValue := argMap["value"]
	if !hasValue {
		return nil, fmt.Errorf("json function requires 'value' field")
	}

	// spaceフィールドはオプション
	space, hasSpace := argMap["space"]
	if !hasSpace {
		// スペースなしでJSON化
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("json function failed to marshal: %w", err)
		}
		return string(jsonBytes), nil
	}

	// スペースありでJSON化
	var indent string
	switch spaceVal := space.(type) {
	case string:
		indent = spaceVal
	case float64:
		// 数値の場合はスペースの数として扱う
		indent = strings.Repeat(" ", int(spaceVal))
	default:
		return nil, fmt.Errorf("json function expects string or number for space parameter")
	}

	jsonBytes, err := json.MarshalIndent(value, "", indent)
	if err != nil {
		return nil, fmt.Errorf("json function failed to marshal with indent: %w", err)
	}

	return string(jsonBytes), nil
}

// MultipartFunction はマルチパートデータエンコーディング関数
type MultipartFunction struct{}

func (f *MultipartFunction) Name() string {
	return "multipart"
}

func (f *MultipartFunction) Signature() string {
	return "!multipart <value> <boundary>"
}

func (f *MultipartFunction) Description() string {
	return "マップをmultipart/form-dataフォーマットに変換します。boundaryパラメータが必要です"
}

func (f *MultipartFunction) Execute(ctx context.Context, args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("multipart function expects 2 arguments, got %d", len(args))
	}

	input, ok := args[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("multipart function expects map[string]interface{} as first argument")
	}

	boundary, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("multipart function expects string as boundary parameter")
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	
	// カスタムバウンダリを設定
	if err := writer.SetBoundary(boundary); err != nil {
		return nil, fmt.Errorf("multipart function failed to set boundary: %w", err)
	}

	for key, value := range input {
		if value != nil {
			field, err := writer.CreateFormField(key)
			if err != nil {
				return nil, fmt.Errorf("multipart function failed to create field %s: %w", key, err)
			}
			
			_, err = field.Write([]byte(fmt.Sprintf("%v", value)))
			if err != nil {
				return nil, fmt.Errorf("multipart function failed to write field %s: %w", key, err)
			}
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("multipart function failed to close writer: %w", err)
	}

	return buf.String(), nil
}