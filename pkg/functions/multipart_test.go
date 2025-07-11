package functions

import (
	"context"
	"strings"
	"testing"
)

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