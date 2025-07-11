package functions

import (
	"context"
	"encoding/base64"
	"fmt"
)

// Base64EncodeFunction はBase64エンコーディング関数
type Base64EncodeFunction struct{}

func (f *Base64EncodeFunction) Name() string {
	return "base64_encode"
}

func (f *Base64EncodeFunction) Signature() string {
	return "$base64_encode <string>"
}

func (f *Base64EncodeFunction) Description() string {
	return "文字列をBase64エンコーディングします"
}

func (f *Base64EncodeFunction) Execute(_ context.Context, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("base64_encode function expects 1 argument, got %d", len(args))
	}

	input, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("base64_encode function expects string argument")
	}

	return base64.StdEncoding.EncodeToString([]byte(input)), nil
}