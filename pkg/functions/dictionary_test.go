package functions

import (
	"context"
	"encoding/json"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"testing"
)

func TestDictionaryLoadFunction_JSON(t *testing.T) {
	// テスト用のJSONファイルを作成
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "test.json")

	testData := []string{"payload1", "payload2", "payload3"}
	jsonData, _ := json.Marshal(testData)

	err := os.WriteFile(jsonFile, jsonData, 0644)
	if err != nil {
		t.Fatalf("Failed to create test JSON file: %v", err)
	}

	// 関数をテスト
	fn := &DictionaryLoadFunction{}
	ctx := context.Background()

	result, err := fn.Execute(ctx, []interface{}{jsonFile})
	if err != nil {
		t.Fatalf("dict_load failed: %v", err)
	}

	resultSlice, ok := result.([]string)
	if !ok {
		t.Fatalf("Expected []string, got %T", result)
	}

	if len(resultSlice) != 3 {
		t.Fatalf("Expected 3 items, got %d", len(resultSlice))
	}

	if resultSlice[0] != "payload1" {
		t.Errorf("Expected 'payload1', got '%s'", resultSlice[0])
	}
}

func TestDictionaryLoadFunction_YAML(t *testing.T) {
	// テスト用のYAMLファイルを作成
	tmpDir := t.TempDir()
	yamlFile := filepath.Join(tmpDir, "test.yaml")

	testData := map[string][]string{
		"payloads": {"<script>alert(1)</script>", "'; DROP TABLE users; --", "admin' OR '1'='1"},
	}
	yamlData, _ := yaml.Marshal(testData)

	err := os.WriteFile(yamlFile, yamlData, 0644)
	if err != nil {
		t.Fatalf("Failed to create test YAML file: %v", err)
	}

	// 関数をテスト
	fn := &DictionaryLoadFunction{}
	ctx := context.Background()

	result, err := fn.Execute(ctx, []interface{}{yamlFile})
	if err != nil {
		t.Fatalf("dict_load failed: %v", err)
	}

	resultSlice, ok := result.([]string)
	if !ok {
		t.Fatalf("Expected []string, got %T", result)
	}

	if len(resultSlice) != 3 {
		t.Fatalf("Expected 3 items, got %d", len(resultSlice))
	}
}

func TestDictionaryLoadFunction_Text(t *testing.T) {
	// テスト用のテキストファイルを作成
	tmpDir := t.TempDir()
	textFile := filepath.Join(tmpDir, "test.txt")

	content := `<script>alert(1)</script>
<img src=x onerror=alert(1)>
javascript:alert(1)

'; DROP TABLE users; --
admin' OR '1'='1
' UNION SELECT * FROM users --
#hashtag_payload
#another_payload`

	err := os.WriteFile(textFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test text file: %v", err)
	}

	// 関数をテスト
	fn := &DictionaryLoadFunction{}
	ctx := context.Background()

	result, err := fn.Execute(ctx, []interface{}{textFile})
	if err != nil {
		t.Fatalf("dict_load failed: %v", err)
	}

	resultSlice, ok := result.([]string)
	if !ok {
		t.Fatalf("Expected []string, got %T", result)
	}

	// 空行を除いて8行になるはず（#で始まる行も含む）
	if len(resultSlice) != 8 {
		t.Fatalf("Expected 8 items, got %d", len(resultSlice))
	}

	if resultSlice[0] != "<script>alert(1)</script>" {
		t.Errorf("Expected '<script>alert(1)</script>', got '%s'", resultSlice[0])
	}
}

func TestDictionaryRandomFunction(t *testing.T) {
	fn := &DictionaryRandomFunction{}
	ctx := context.Background()

	testDict := []string{"payload1", "payload2", "payload3"}

	result, err := fn.Execute(ctx, []interface{}{testDict})
	if err != nil {
		t.Fatalf("dict_random failed: %v", err)
	}

	resultStr, ok := result.(string)
	if !ok {
		t.Fatalf("Expected string, got %T", result)
	}

	// 結果が辞書の中の値のいずれかであることを確認
	found := false
	for _, item := range testDict {
		if resultStr == item {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Result '%s' not found in dictionary", resultStr)
	}
}

func TestDictionaryGetFunction(t *testing.T) {
	fn := &DictionaryGetFunction{}
	ctx := context.Background()

	testDict := []string{"payload1", "payload2", "payload3"}

	// インデックス1の値を取得
	result, err := fn.Execute(ctx, []interface{}{testDict, 1})
	if err != nil {
		t.Fatalf("dict_get failed: %v", err)
	}

	resultStr, ok := result.(string)
	if !ok {
		t.Fatalf("Expected string, got %T", result)
	}

	if resultStr != "payload2" {
		t.Errorf("Expected 'payload2', got '%s'", resultStr)
	}

	// 範囲外のインデックスをテスト
	_, err = fn.Execute(ctx, []interface{}{testDict, 10})
	if err == nil {
		t.Error("Expected error for out of range index")
	}
}

func TestDictionaryLoadFunction_FileNotFound(t *testing.T) {
	fn := &DictionaryLoadFunction{}
	ctx := context.Background()

	_, err := fn.Execute(ctx, []interface{}{"non_existent_file.txt"})
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestDictionaryLoadFunction_UnsupportedFormat(t *testing.T) {
	// テスト用のサポートされていない形式のファイルを作成
	tmpDir := t.TempDir()
	unsupportedFile := filepath.Join(tmpDir, "test.doc")

	err := os.WriteFile(unsupportedFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	fn := &DictionaryLoadFunction{}
	ctx := context.Background()

	_, err = fn.Execute(ctx, []interface{}{unsupportedFile})
	if err == nil {
		t.Error("Expected error for unsupported file format")
	}
}

func TestDictionaryLoadFunction_RelativePath(t *testing.T) {
	// テスト用のディレクトリとファイルを作成
	tmpDir := t.TempDir()

	// リクエストファイルが置かれるディレクトリ
	reqDir := filepath.Join(tmpDir, "requests")
	err := os.Mkdir(reqDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// 辞書ファイル
	dictFile := filepath.Join(reqDir, "payloads.txt")
	err = os.WriteFile(dictFile, []byte("test_payload"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test dictionary file: %v", err)
	}

	// 関数をテスト
	fn := &DictionaryLoadFunction{}

	// コンテキストにリクエストファイルのパスを設定
	reqFilePath := filepath.Join(reqDir, "request.json")
	ctx := context.WithValue(context.Background(), "requestFilePath", reqFilePath)

	// 相対パスで辞書ファイルを指定
	result, err := fn.Execute(ctx, []interface{}{"payloads.txt"})
	if err != nil {
		t.Fatalf("dict_load with relative path failed: %v", err)
	}

	resultSlice, ok := result.([]string)
	if !ok {
		t.Fatalf("Expected []string, got %T", result)
	}

	if len(resultSlice) != 1 || resultSlice[0] != "test_payload" {
		t.Errorf("Expected ['test_payload'], got %v", resultSlice)
	}
}
