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

func TestJSONFunction(t *testing.T) {
	fn := &JSONFunction{}
	ctx := context.Background()

	// 基本的なJSON変換
	input := map[string]interface{}{
		"name": "John",
		"age":  30,
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
	
	// スペース指定でのJSON変換
	result, err = fn.Execute(ctx, []interface{}{input, 2})
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
}

func TestMultipartFunction(t *testing.T) {
	fn := &MultipartFunction{}
	ctx := context.Background()

	input := map[string]interface{}{
		"name":  "John",
		"email": "john@example.com",
	}
	boundary := "----WebKitFormBoundary7MA4YWxkTrZu0gW"
	
	result, err := fn.Execute(ctx, []interface{}{input, boundary})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	resultStr, ok := result.(string)
	if !ok {
		t.Fatalf("Expected string result, got %T", result)
	}
	
	// マルチパート形式になっているかチェック
	if !strings.Contains(resultStr, boundary) {
		t.Errorf("Expected boundary in result, got %s", resultStr)
	}
	
	if !strings.Contains(resultStr, "Content-Disposition: form-data") {
		t.Errorf("Expected multipart headers, got %s", resultStr)
	}
	
	// 引数の数が間違っている場合
	_, err = fn.Execute(ctx, []interface{}{input})
	if err == nil {
		t.Error("Expected error for missing boundary")
	}
}