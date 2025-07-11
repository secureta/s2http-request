package functions

import (
	"context"
	"fmt"
	"html"
)

// HTMLEncodeFunction はHTMLエンコーディング関数
type HTMLEncodeFunction struct{}

func (f *HTMLEncodeFunction) Name() string {
	return "html_encode"
}

func (f *HTMLEncodeFunction) Signature() string {
	return "$html_encode <string>"
}

func (f *HTMLEncodeFunction) Description() string {
	return "文字列をHTMLエンコーディングします"
}

func (f *HTMLEncodeFunction) Execute(_ context.Context, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("html_encode function expects 1 argument, got %d", len(args))
	}

	input, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("html_encode function expects string argument")
	}

	return html.EscapeString(input), nil
}