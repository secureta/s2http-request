package functions

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"time"
	"unicode"
)

// DoubleEncodeFunction は二重URLエンコーディング関数
type DoubleEncodeFunction struct{}

func (f *DoubleEncodeFunction) Name() string {
	return "double_encode"
}

func (f *DoubleEncodeFunction) Signature() string {
	return "!double_encode <string>"
}

func (f *DoubleEncodeFunction) Description() string {
	return "文字列を二重URLエンコーディングします（WAF回避用）"
}

func (f *DoubleEncodeFunction) Execute(_ context.Context, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("double_encode function expects 1 argument, got %d", len(args))
	}

	input, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("double_encode function expects string argument")
	}

	// 一回目のエンコーディング
	encoded := url.QueryEscape(input)
	// 二回目のエンコーディング
	doubleEncoded := url.QueryEscape(encoded)

	return doubleEncoded, nil
}

// UnicodeEncodeFunction はUnicode文字エンコーディング関数
type UnicodeEncodeFunction struct{}

func (f *UnicodeEncodeFunction) Name() string {
	return "unicode_encode"
}

func (f *UnicodeEncodeFunction) Signature() string {
	return "!unicode_encode <string>"
}

func (f *UnicodeEncodeFunction) Description() string {
	return "ASCII以外の文字や制御文字をUnicodeエスケープします（WAF回避用）"
}

func (f *UnicodeEncodeFunction) Execute(_ context.Context, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("unicode_encode function expects 1 argument, got %d", len(args))
	}

	input, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("unicode_encode function expects string argument")
	}

	var result strings.Builder
	for _, r := range input {
		if r > 127 || unicode.IsControl(r) {
			// ASCII以外の文字や制御文字をUnicodeエスケープ
			result.WriteString(fmt.Sprintf("\\u%04x", r))
		} else {
			result.WriteRune(r)
		}
	}

	return result.String(), nil
}

// CaseVariationFunction は大文字小文字のランダム変換関数
type CaseVariationFunction struct{}

func (f *CaseVariationFunction) Name() string {
	return "case_variation"
}

func (f *CaseVariationFunction) Signature() string {
	return "!case_variation <string>"
}

func (f *CaseVariationFunction) Description() string {
	return "文字列の大文字小文字をランダムに変換します（WAF回避用）"
}

func (f *CaseVariationFunction) Execute(_ context.Context, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("case_variation function expects 1 argument, got %d", len(args))
	}

	input, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("case_variation function expects string argument")
	}

	// シードを現在時刻で初期化
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	var result strings.Builder
	for _, r := range input {
		if unicode.IsLetter(r) {
			// 50%の確率で大文字/小文字を切り替え
			if rng.Float32() < 0.5 {
				if unicode.IsUpper(r) {
					result.WriteRune(unicode.ToLower(r))
				} else {
					result.WriteRune(unicode.ToUpper(r))
				}
			} else {
				result.WriteRune(r)
			}
		} else {
			result.WriteRune(r)
		}
	}

	return result.String(), nil
}
