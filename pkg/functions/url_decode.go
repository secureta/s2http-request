package functions

import (
	"context"
	"fmt"
	"net/url"
)

// URLDecodeFunction はURLデコーディング関数
type URLDecodeFunction struct{}

func (f *URLDecodeFunction) Name() string {
	return "url_decode"
}

func (f *URLDecodeFunction) Signature() string {
	return "$url_decode <encoded_string>"
}

func (f *URLDecodeFunction) Description() string {
	return "URLエンコードされた文字列をデコードします"
}

func (f *URLDecodeFunction) Execute(_ context.Context, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("url_decode function expects 1 argument, got %d", len(args))
	}

	input, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("url_decode function expects string argument")
	}

	decoded, err := url.QueryUnescape(input)
	if err != nil {
		return nil, fmt.Errorf("url_decode failed: %w", err)
	}

	return decoded, nil
}