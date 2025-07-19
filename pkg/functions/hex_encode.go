package functions

import (
	"context"
	"encoding/hex"
	"fmt"
)

// HexEncodeFunction は16進数エンコーディング関数
type HexEncodeFunction struct{}

func (f *HexEncodeFunction) Name() string {
	return "hex_encode"
}

func (f *HexEncodeFunction) Signature() string {
	return "$hex_encode <string>"
}

func (f *HexEncodeFunction) Description() string {
	return "文字列を16進数エンコーディングします"
}

func (f *HexEncodeFunction) Execute(_ context.Context, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("hex_encode function expects 1 argument, got %d", len(args))
	}

	input, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("hex_encode function expects string argument")
	}

	return hex.EncodeToString([]byte(input)), nil
}
