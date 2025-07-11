package functions

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// URLEncodeFunction はURLエンコーディング関数
type URLEncodeFunction struct{}

func (f *URLEncodeFunction) Name() string {
	return "url_encode"
}

func (f *URLEncodeFunction) Signature() string {
	return "$url_encode <string> [times] [chars_to_not_encode]"
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
