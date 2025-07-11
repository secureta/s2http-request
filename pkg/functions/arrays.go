package functions

import (
	"context"
	"fmt"
)

// ConcatArraysFunction は複数の配列を結合する関数
type ConcatArraysFunction struct{}

func (f *ConcatArraysFunction) Name() string {
	return "concat_arrays"
}

func (f *ConcatArraysFunction) Signature() string {
	return "$concat_arrays <arrays...>"
}

func (f *ConcatArraysFunction) Description() string {
	return "複数の配列を結合して1つの配列にします"
}

func (f *ConcatArraysFunction) Execute(_ context.Context, args []interface{}) (interface{}, error) {
	if len(args) == 0 {
		return []string{}, nil
	}

	var result []string
	for _, arg := range args {
		switch arr := arg.(type) {
		case []interface{}:
			for _, item := range arr {
				result = append(result, fmt.Sprintf("%v", item))
			}
		case []string:
			result = append(result, arr...)
		case string:
			result = append(result, arr)
		default:
			result = append(result, fmt.Sprintf("%v", arr))
		}
	}

	return result, nil
}
