package functions

import (
	"context"
	"strings"
	"testing"
)

func TestJSONFunction(t *testing.T) {
	fn := &JSONFunction{}
	ctx := context.Background()

	// 新形式での基本的なJSON変換
	input := map[string]interface{}{
		"value": map[string]interface{}{
			"name": "John",
			"age":  30,
		},
	}

	result, err := fn.Execute(ctx, []interface{}{input})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	resultStr, ok := result.(string)
	if !ok {
		t.Fatalf("Expected string result, got %T", result)
	}

	if !strings.Contains(resultStr, `"name":"John"`) {
		t.Errorf("Expected JSON string, got %s", resultStr)
	}

	// 新形式でのスペース指定でのJSON変換
	inputWithSpace := map[string]interface{}{
		"value": map[string]interface{}{
			"name": "John",
			"age":  30,
		},
		"space": float64(2),
	}

	result, err = fn.Execute(ctx, []interface{}{inputWithSpace})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	resultStr, ok = result.(string)
	if !ok {
		t.Fatalf("Expected string result, got %T", result)
	}

	// インデントされているかチェック
	if !strings.Contains(resultStr, "\n") {
		t.Errorf("Expected indented JSON, got %s", resultStr)
	}

	// valueフィールドが必須であることのテスト
	invalidInput := map[string]interface{}{
		"space": float64(2),
	}

	_, err = fn.Execute(ctx, []interface{}{invalidInput})
	if err == nil {
		t.Error("Expected error for missing 'value' field")
	}

	// 引数がマップでない場合のテスト
	_, err = fn.Execute(ctx, []interface{}{"not a map"})
	if err == nil {
		t.Error("Expected error for non-map argument")
	}
}

