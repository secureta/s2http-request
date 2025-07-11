package functions

import (
	"context"
	"strings"
	"testing"
)

func TestFormFunction(t *testing.T) {
	fn := &FormFunction{}
	ctx := context.Background()

	// 正常なケース
	input := map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	}
	
	result, err := fn.Execute(ctx, []interface{}{input})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	resultStr, ok := result.(string)
	if !ok {
		t.Fatalf("Expected string result, got %T", result)
	}
	
	// URLエンコードされた形式になっているかチェック
	if !strings.Contains(resultStr, "name=John+Doe") {
		t.Errorf("Expected encoded name, got %s", resultStr)
	}
	
	// 引数の数が間違っている場合
	_, err = fn.Execute(ctx, []interface{}{})
	if err == nil {
		t.Error("Expected error for no arguments")
	}
}