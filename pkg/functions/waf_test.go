package functions

import (
	"context"
	"net/url"
	"strings"
	"testing"
	"unicode"
)

func TestDoubleEncodeFunction(t *testing.T) {
	tests := []struct {
		name      string
		args      []interface{}
		expected  string
		wantError bool
	}{
		{
			name:      "simple string",
			args:      []interface{}{"hello world"},
			expected:  url.QueryEscape(url.QueryEscape("hello world")),
			wantError: false,
		},
		{
			name:      "special characters",
			args:      []interface{}{"' OR 1=1 --"},
			expected:  url.QueryEscape(url.QueryEscape("' OR 1=1 --")),
			wantError: false,
		},
		{
			name:      "html tags",
			args:      []interface{}{"<script>alert('xss')</script>"},
			expected:  url.QueryEscape(url.QueryEscape("<script>alert('xss')</script>")),
			wantError: false,
		},
		{
			name:      "empty string",
			args:      []interface{}{""},
			expected:  "",
			wantError: false,
		},
		{
			name:      "wrong number of arguments",
			args:      []interface{}{"arg1", "arg2"},
			expected:  "",
			wantError: true,
		},
		{
			name:      "non-string argument",
			args:      []interface{}{123},
			expected:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := &DoubleEncodeFunction{}
			ctx := context.Background()

			result, err := fn.Execute(ctx, tt.args)

			if tt.wantError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.wantError && result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestUnicodeEncodeFunction(t *testing.T) {
	tests := []struct {
		name      string
		args      []interface{}
		expected  string
		wantError bool
	}{
		{
			name:      "ascii only",
			args:      []interface{}{"hello"},
			expected:  "hello",
			wantError: false,
		},
		{
			name:      "unicode characters",
			args:      []interface{}{"こんにちは"},
			expected:  "\\u3053\\u3093\\u306b\\u3061\\u306f",
			wantError: false,
		},
		{
			name:      "mixed ascii and unicode",
			args:      []interface{}{"hello世界"},
			expected:  "hello\\u4e16\\u754c",
			wantError: false,
		},
		{
			name:      "control characters",
			args:      []interface{}{"hello\n\t"},
			expected:  "hello\\u000a\\u0009",
			wantError: false,
		},
		{
			name:      "high ascii characters",
			args:      []interface{}{"café"},
			expected:  "caf\\u00e9",
			wantError: false,
		},
		{
			name:      "empty string",
			args:      []interface{}{""},
			expected:  "",
			wantError: false,
		},
		{
			name:      "wrong number of arguments",
			args:      []interface{}{"arg1", "arg2"},
			expected:  "",
			wantError: true,
		},
		{
			name:      "non-string argument",
			args:      []interface{}{123},
			expected:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := &UnicodeEncodeFunction{}
			ctx := context.Background()

			result, err := fn.Execute(ctx, tt.args)

			if tt.wantError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.wantError && result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestCaseVariationFunction(t *testing.T) {
	tests := []struct {
		name      string
		args      []interface{}
		checkFunc func(string, string) bool
		wantError bool
	}{
		{
			name: "alphabetic string",
			args: []interface{}{"admin"},
			checkFunc: func(input, result string) bool {
				// 結果の長さが同じであることを確認
				if len(input) != len(result) {
					return false
				}
				// 各文字が元の文字または大文字小文字が変換された文字であることを確認
				inputRunes := []rune(input)
				resultRunes := []rune(result)
				for i, r := range inputRunes {
					if unicode.IsLetter(r) {
						// 文字または大文字小文字変換された文字である必要がある
						if resultRunes[i] != r && 
						   resultRunes[i] != unicode.ToUpper(r) && 
						   resultRunes[i] != unicode.ToLower(r) {
							return false
						}
					} else {
						// 非文字は変更されないはず
						if resultRunes[i] != r {
							return false
						}
					}
				}
				return true
			},
			wantError: false,
		},
		{
			name: "mixed case string",
			args: []interface{}{"AdMiN"},
			checkFunc: func(input, result string) bool {
				return len(input) == len(result)
			},
			wantError: false,
		},
		{
			name: "string with numbers and symbols",
			args: []interface{}{"admin123!@#"},
			checkFunc: func(input, result string) bool {
				// 数字と記号は変更されないことを確認
				inputRunes := []rune(input)
				resultRunes := []rune(result)
				for i, r := range inputRunes {
					if !unicode.IsLetter(r) {
						if resultRunes[i] != r {
							return false
						}
					}
				}
				return len(input) == len(result)
			},
			wantError: false,
		},
		{
			name: "empty string",
			args: []interface{}{""},
			checkFunc: func(input, result string) bool {
				return result == ""
			},
			wantError: false,
		},
		{
			name: "non-alphabetic string",
			args: []interface{}{"123!@#"},
			checkFunc: func(input, result string) bool {
				return input == result // 変更されないはず
			},
			wantError: false,
		},
		{
			name:      "wrong number of arguments",
			args:      []interface{}{"arg1", "arg2"},
			checkFunc: nil,
			wantError: true,
		},
		{
			name:      "non-string argument",
			args:      []interface{}{123},
			checkFunc: nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := &CaseVariationFunction{}
			ctx := context.Background()

			result, err := fn.Execute(ctx, tt.args)

			if tt.wantError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.wantError && tt.checkFunc != nil {
				resultStr, ok := result.(string)
				if !ok {
					t.Errorf("Expected string result, got %T", result)
				}
				inputStr := tt.args[0].(string)
				if !tt.checkFunc(inputStr, resultStr) {
					t.Errorf("Result validation failed for input %q, got %q", inputStr, resultStr)
				}
			}
		})
	}
}

func TestCaseVariationFunctionRandomness(t *testing.T) {
	fn := &CaseVariationFunction{}
	ctx := context.Background()
	input := "abcdefghijklmnopqrstuvwxyz"

	// 複数回実行して、結果が異なることを確認（ランダム性のテスト）
	results := make(map[string]bool)
	for i := 0; i < 10; i++ {
		result, err := fn.Execute(ctx, []interface{}{input})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		resultStr := result.(string)
		results[resultStr] = true
	}

	// 少なくとも2つ以上の異なる結果があることを期待
	// （確率的に同じ結果が複数回出る可能性もあるが、10回中全て同じは非常に稀）
	if len(results) < 2 {
		t.Logf("Warning: Only %d unique results in 10 attempts. This might indicate low randomness.", len(results))
		// これは警告レベルで、テスト失敗とはしない（確率的な問題のため）
	}
}

func TestDoubleEncodeFunctionName(t *testing.T) {
	fn := &DoubleEncodeFunction{}
	if fn.Name() != "double_encode" {
		t.Errorf("Expected function name 'double_encode', got %q", fn.Name())
	}
}

func TestUnicodeEncodeFunctionName(t *testing.T) {
	fn := &UnicodeEncodeFunction{}
	if fn.Name() != "unicode_encode" {
		t.Errorf("Expected function name 'unicode_encode', got %q", fn.Name())
	}
}

func TestCaseVariationFunctionName(t *testing.T) {
	fn := &CaseVariationFunction{}
	if fn.Name() != "case_variation" {
		t.Errorf("Expected function name 'case_variation', got %q", fn.Name())
	}
}

// WAF回避テスト用の実際のペイロードをテスト
func TestWAFFunctionWithRealPayloads(t *testing.T) {
	ctx := context.Background()

	// SQLインジェクションペイロード
	sqlPayload := "' OR 1=1 --"
	
	// 二重エンコーディングテスト
	doubleEncodeFn := &DoubleEncodeFunction{}
	doubleEncoded, err := doubleEncodeFn.Execute(ctx, []interface{}{sqlPayload})
	if err != nil {
		t.Fatalf("Double encode failed: %v", err)
	}
	
	// 結果が元の文字列と異なることを確認
	if doubleEncoded == sqlPayload {
		t.Errorf("Double encoding should change the payload")
	}

	// XSSペイロード
	xssPayload := "<script>alert('xss')</script>こんにちは"
	
	// Unicodeエンコーディングテスト
	unicodeEncodeFn := &UnicodeEncodeFunction{}
	unicodeEncoded, err := unicodeEncodeFn.Execute(ctx, []interface{}{xssPayload})
	if err != nil {
		t.Fatalf("Unicode encode failed: %v", err)
	}
	
	// 特殊文字が含まれていることを確認
	unicodeStr := unicodeEncoded.(string)
	if !strings.Contains(unicodeStr, "\\u") {
		t.Errorf("Unicode encoding should contain \\u sequences for special characters")
	}

	// 大文字小文字変換テスト
	adminPayload := "admin"
	caseVarFn := &CaseVariationFunction{}
	caseVaried, err := caseVarFn.Execute(ctx, []interface{}{adminPayload})
	if err != nil {
		t.Fatalf("Case variation failed: %v", err)
	}
	
	// 結果が文字列であることを確認
	if _, ok := caseVaried.(string); !ok {
		t.Errorf("Case variation should return a string")
	}
}