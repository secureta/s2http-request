package functions

import (
	"context"
	"fmt"
	"html"
)

// HTMLDecodeFunction はHTMLデコーディング関数
type HTMLDecodeFunction struct{}

func (f *HTMLDecodeFunction) Name() string {
	return "html_decode"
}

func (f *HTMLDecodeFunction) Signature() string {
	return "$html_decode <encoded_string>"
}

func (f *HTMLDecodeFunction) Description() string {
	return "HTMLエンコードされた文字列をデコードします"
}

func (f *HTMLDecodeFunction) Execute(_ context.Context, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("html_decode function expects 1 argument, got %d", len(args))
	}

	input, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("html_decode function expects string argument")
	}

	return html.UnescapeString(input), nil
}