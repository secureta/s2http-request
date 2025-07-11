package functions

import (
	"context"
	"encoding/base64"
	"fmt"
)

// Base64DecodeFunction はBase64デコーディング関数
type Base64DecodeFunction struct{}

func (f *Base64DecodeFunction) Name() string {
	return "base64_decode"
}

func (f *Base64DecodeFunction) Signature() string {
	return "$base64_decode <encoded_string>"
}

func (f *Base64DecodeFunction) Description() string {
	return "Base64エンコードされた文字列をデコードします"
}

func (f *Base64DecodeFunction) Execute(_ context.Context, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("base64_decode function expects 1 argument, got %d", len(args))
	}

	input, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("base64_decode function expects string argument")
	}

	decoded, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return nil, fmt.Errorf("base64_decode failed: %w", err)
	}

	return string(decoded), nil
}