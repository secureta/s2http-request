package functions

import (
	"context"
	"fmt"
)

// DictFunction はdict変数参照関数
type DictFunction struct{}

func (f *DictFunction) Name() string {
	return "dict"
}

func (f *DictFunction) Signature() string {
	return "$dict <variable_name>"
}

func (f *DictFunction) Description() string {
	return "dict変数の値を参照します。dictプロパティで定義された配列ベースの変数を取得できます。"
}

func (f *DictFunction) Execute(ctx context.Context, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("dict function expects 1 argument, got %d", len(args))
	}

	varName, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("dict function expects string argument, got %T", args[0])
	}

	// コンテキストからdict変数を取得
	dictVars, ok := ctx.Value("dict").(map[string]interface{})
	if !ok {
		// Get file path from context for better error reporting
		filePath, _ := ctx.Value("requestFilePath").(string)
		if filePath != "" {
			return nil, fmt.Errorf("$dict reference '%s' found but no dict variables are defined in %s", varName, filePath)
		}
		return nil, fmt.Errorf("$dict reference '%s' found but no dict variables are defined", varName)
	}

	if value, exists := dictVars[varName]; exists {
		return value, nil
	}

	// Provide helpful error message with available variables
	var availableVars []string
	for key := range dictVars {
		availableVars = append(availableVars, key)
	}

	filePath, _ := ctx.Value("requestFilePath").(string)
	if len(availableVars) > 0 {
		if filePath != "" {
			return nil, fmt.Errorf("$dict reference '%s' not found in %s. Available dict variables: %v", varName, filePath, availableVars)
		}
		return nil, fmt.Errorf("$dict reference '%s' not found. Available dict variables: %v", varName, availableVars)
	}

	if filePath != "" {
		return nil, fmt.Errorf("$dict reference '%s' not found in %s. No dict variables are defined", varName, filePath)
	}
	return nil, fmt.Errorf("$dict reference '%s' not found. No dict variables are defined", varName)
}
