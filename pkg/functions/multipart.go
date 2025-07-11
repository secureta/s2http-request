package functions

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
)

// MultipartFunction はマルチパートデータエンコーディング関数
type MultipartFunction struct{}

func (f *MultipartFunction) Name() string {
	return "multipart"
}

func (f *MultipartFunction) Signature() string {
	return "$multipart <value> <boundary>"
}

func (f *MultipartFunction) Description() string {
	return "マップをmultipart/form-dataフォーマットに変換します。boundaryパラメータが必要です"
}

func (f *MultipartFunction) Execute(_ context.Context, args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("multipart function expects 2 arguments, got %d", len(args))
	}

	input, ok := args[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("multipart function expects map[string]interface{} as first argument")
	}

	boundary, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("multipart function expects string as boundary parameter")
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// カスタムバウンダリを設定
	if err := writer.SetBoundary(boundary); err != nil {
		return nil, fmt.Errorf("multipart function failed to set boundary: %w", err)
	}

	for key, value := range input {
		if value != nil {
			field, err := writer.CreateFormField(key)
			if err != nil {
				return nil, fmt.Errorf("multipart function failed to create field %s: %w", key, err)
			}

			_, err = field.Write([]byte(fmt.Sprintf("%v", value)))
			if err != nil {
				return nil, fmt.Errorf("multipart function failed to write field %s: %w", key, err)
			}
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("multipart function failed to close writer: %w", err)
	}

	return buf.String(), nil
}