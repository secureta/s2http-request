package functions

import (
	"context"
	"encoding/base64"
	"fmt"
	"html"
	"net/url"
	"strings"
)

// URLEncodeFunction はURLエンコーディング関数
type URLEncodeFunction struct{}

func (f *URLEncodeFunction) Name() string {
	return "url_encode"
}

func (f *URLEncodeFunction) Signature() string {
	return "!url_encode <string> [times] [chars_to_not_encode]"
}

func (f *URLEncodeFunction) Description() string {
	return "文字列をURLエンコーディングします。エンコード回数と、オプションでエンコードしない文字を指定できます"
}

func (f *URLEncodeFunction) Execute(_ context.Context, args []interface{}) (interface{}, error) {
	if len(args) < 1 || len(args) > 3 {
		return nil, fmt.Errorf("url_encode function expects 1 to 3 arguments, got %d", len(args))
	}

	input, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("url_encode function expects string argument for the first argument")
	}

	encodeTimes := 1
	notToEncode := ""

	if len(args) >= 2 {
		if times, ok := args[1].(float64); ok {
			encodeTimes = int(times)
			if len(args) == 3 {
				if chars, ok := args[2].(string); ok {
					notToEncode = chars
				} else {
					return nil, fmt.Errorf("the third argument for url_encode must be a string (chars_to_not_encode)")
				}
			}
		} else if chars, ok := args[1].(string); ok {
			notToEncode = chars
		} else {
			return nil, fmt.Errorf("the second argument for url_encode must be an integer (times) or a string (chars_to_not_encode)")
		}
	}

	encoded := input
	for i := 0; i < encodeTimes; i++ {
		if notToEncode != "" {
			var result strings.Builder
			for _, r := range encoded {
				if strings.ContainsRune(notToEncode, r) {
					result.WriteRune(r)
				} else {
					result.WriteString(url.QueryEscape(string(r)))
				}
			}
			encoded = result.String()
		} else {
			encoded = url.QueryEscape(encoded)
		}
	}

	return encoded, nil
}

// URLDecodeFunction はURLデコーディング関数
type URLDecodeFunction struct{}

func (f *URLDecodeFunction) Name() string {
	return "url_decode"
}

func (f *URLDecodeFunction) Signature() string {
	return "!url_decode <encoded_string>"
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

// Base64EncodeFunction はBase64エンコーディング関数
type Base64EncodeFunction struct{}

func (f *Base64EncodeFunction) Name() string {
	return "base64_encode"
}

func (f *Base64EncodeFunction) Signature() string {
	return "!base64_encode <string>"
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

// Base64DecodeFunction はBase64デコーディング関数
type Base64DecodeFunction struct{}

func (f *Base64DecodeFunction) Name() string {
	return "base64_decode"
}

func (f *Base64DecodeFunction) Signature() string {
	return "!base64_decode <encoded_string>"
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

// HTMLEncodeFunction はHTMLエンコーディング関数
type HTMLEncodeFunction struct{}

func (f *HTMLEncodeFunction) Name() string {
	return "html_encode"
}

func (f *HTMLEncodeFunction) Signature() string {
	return "!html_encode <string>"
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

// HTMLDecodeFunction はHTMLデコーディング関数
type HTMLDecodeFunction struct{}

func (f *HTMLDecodeFunction) Name() string {
	return "html_decode"
}

func (f *HTMLDecodeFunction) Signature() string {
	return "!html_decode <encoded_string>"
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
