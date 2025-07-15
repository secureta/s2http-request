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
	return "$multipart {values: <map>, boundary: <string>}"
}

func (f *MultipartFunction) Description() string {
	return "マップをmultipart/form-dataフォーマットに変換します。{values: <map>, boundary: <string>}の形式で指定します"
}

func (f *MultipartFunction) Execute(_ context.Context, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("multipart function expects 1 argument, got %d", len(args))
	}

	config, ok := args[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("multipart function expects map[string]interface{} as argument")
	}

	// valuesフィールドを取得
	valuesInterface, ok := config["values"]
	if !ok {
		return nil, fmt.Errorf("multipart function requires 'values' field")
	}

	values, ok := valuesInterface.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("multipart function expects 'values' to be map[string]interface{}")
	}

	// boundaryフィールドを取得
	boundaryInterface, ok := config["boundary"]
	if !ok {
		return nil, fmt.Errorf("multipart function requires 'boundary' field")
	}

	boundary, ok := boundaryInterface.(string)
	if !ok {
		return nil, fmt.Errorf("multipart function expects 'boundary' to be string")
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// カスタムバウンダリを設定
	if err := writer.SetBoundary(boundary); err != nil {
		return nil, fmt.Errorf("multipart function failed to set boundary: %w", err)
	}

	for key, value := range values {
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