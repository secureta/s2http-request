package functions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// JSONFunction はJSONエンコーディング関数
type JSONFunction struct{}

func (f *JSONFunction) Name() string {
	return "json"
}

func (f *JSONFunction) Signature() string {
	return "$json {value: <value>, space?: <space>}"
}

func (f *JSONFunction) Description() string {
	return "値をJSON文字列に変換します。{value: <value>, space?: <space>} の形式で引数を指定します"
}

func (f *JSONFunction) Execute(_ context.Context, args []interface{}) (interface{}, error) {
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
