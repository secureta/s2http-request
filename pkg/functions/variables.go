package functions

import (
	"context"
	"fmt"
	"strings"
)

// VarFunction は変数参照関数
type VarFunction struct {
	variables map[string]interface{}
}

func (f *VarFunction) Name() string {
	return "var"
}

func (f *VarFunction) Signature() string {
	return "!var <variable_name>"
}

func (f *VarFunction) Description() string {
	return "変数の値を参照します"
}

func (f *VarFunction) Execute(ctx context.Context, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("var function expects 1 argument, got %d", len(args))
	}

	varName, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("var function expects string argument")
	}

	// コンテキストから変数を取得 (dictionary優先→variables)
	if variables, ok := ctx.Value("variables").(map[string]interface{}); ok {
		if value, exists := variables[varName]; exists {
			return value, nil
		}
	}
	return nil, fmt.Errorf("variable '%s' not found", varName)
}

// ConcatFunction は文字列連結関数
type ConcatFunction struct{}

func (f *ConcatFunction) Name() string {
	return "concat"
}

func (f *ConcatFunction) Signature() string {
	return "!concat [value1, value2, ...]"
}

func (f *ConcatFunction) Description() string {
	return "複数の値を文字列として連結します"
}

func (f *ConcatFunction) Execute(ctx context.Context, args []interface{}) (interface{}, error) {
	if len(args) == 0 {
		return "", nil
	}
	
	var result strings.Builder
	for _, arg := range args {
		result.WriteString(fmt.Sprintf("%v", arg))
	}
	
	return result.String(), nil
}

// JoinFunction は配列要素結合関数
type JoinFunction struct{}

func (f *JoinFunction) Name() string {
	return "join"
}

func (f *JoinFunction) Signature() string {
	return "!join [separator, value1, value2, ...]"
}

func (f *JoinFunction) Description() string {
	return "指定した区切り文字で複数の値を結合します"
}

func (f *JoinFunction) Execute(ctx context.Context, args []interface{}) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("join function expects at least 2 arguments, got %d", len(args))
	}
	
	separator, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("join function expects string separator as first argument")
	}
	
	var parts []string
	for _, arg := range args[1:] {
		parts = append(parts, fmt.Sprintf("%v", arg))
	}
	
	return strings.Join(parts, separator), nil
}