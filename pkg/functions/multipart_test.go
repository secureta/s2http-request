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
		"values": map[string]interface{}{
			"name":  "John",
			"email": "john@example.com",
		},
		"boundary": "----WebKitFormBoundary7MA4YWxkTrZu0gW",
	}
	
	result, err := fn.Execute(ctx, []interface{}{input})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	resultStr, ok := result.(string)
	if !ok {
		t.Fatalf("Expected string result, got %T", result)
	}
	
	// マルチパート形式になっているかチェック
	if !strings.Contains(resultStr, "----WebKitFormBoundary7MA4YWxkTrZu0gW") {
		t.Errorf("Expected boundary in result, got %s", resultStr)
	}
	
	if !strings.Contains(resultStr, "Content-Disposition: form-data") {
		t.Errorf("Expected multipart headers, got %s", resultStr)
	}
	
	// valuesフィールドが欠けている場合
	invalidInput := map[string]interface{}{
		"boundary": "----WebKitFormBoundary7MA4YWxkTrZu0gW",
	}
	_, err = fn.Execute(ctx, []interface{}{invalidInput})
	if err == nil {
		t.Error("Expected error for missing values field")
	}
	
	// boundaryフィールドが欠けている場合
	invalidInput2 := map[string]interface{}{
		"values": map[string]interface{}{
			"name": "John",
		},
	}
	_, err = fn.Execute(ctx, []interface{}{invalidInput2})
	if err == nil {
		t.Error("Expected error for missing boundary field")
	}
}