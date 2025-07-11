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
	return "$var <variable_name>"
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
	return "$concat [value1, value2, ...]"
}

func (f *ConcatFunction) Description() string {
	return "複数の値を文字列として連結します"
}

func (f *ConcatFunction) Execute(_ context.Context, args []interface{}) (interface{}, error) {
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
	return "$join {values: [value1, value2, ...], delimiter: optional_delimiter}"
}

func (f *JoinFunction) Description() string {
	return "文字列の配列を指定した区切り文字で結合します。区切り文字が指定されない場合は区切り文字なしで結合します。"
}

func (f *JoinFunction) Execute(_ context.Context, args []interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("join function expects at least 1 argument (values), got 0")
	}

	// デフォルトの区切り文字は空文字
	delimiter := ""

	// 引数が2つある場合、2番目は区切り文字
	if len(args) > 1 {
		var ok bool
		delimiter, ok = args[1].(string)
		if !ok {
			return nil, fmt.Errorf("join function expects string delimiter as second argument")
		}
	}

	// 最初の引数は文字列の配列
	var strValues []string

	// 配列の型に応じて処理
	switch values := args[0].(type) {
	case []interface{}:
		for _, v := range values {
			strValues = append(strValues, fmt.Sprintf("%v", v))
		}
	case []string:
		strValues = values
	case string:
		// 単一の文字列の場合はそのまま返す
		return values, nil
	default:
		return nil, fmt.Errorf("join function expects array of values as first argument")
	}

	return strings.Join(strValues, delimiter), nil
}
