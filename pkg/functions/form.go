package functions

import (
	"context"
	"fmt"
	"net/url"
)

// FormFunction はフォームデータエンコーディング関数
type FormFunction struct{}

func (f *FormFunction) Name() string {
	return "form"
}

func (f *FormFunction) Signature() string {
	return "$form <map>"
}

func (f *FormFunction) Description() string {
	return "マップをapplication/x-www-form-urlencodedフォーマットに変換します"
}

func (f *FormFunction) Execute(_ context.Context, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("form function expects 1 argument, got %d", len(args))
	}

	input, ok := args[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("form function expects map[string]interface{} argument")
	}

	values := url.Values{}
	for k, v := range input {
		if v != nil {
			values.Add(k, fmt.Sprintf("%v", v))
		}
	}

	return values.Encode(), nil
}