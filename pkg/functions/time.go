package functions

import (
	"context"
	"fmt"
	"time"
)

// TimestampFunction は現在のタイムスタンプ取得関数
type TimestampFunction struct{}

func (f *TimestampFunction) Name() string {
	return "timestamp"
}

func (f *TimestampFunction) Signature() string {
	return "!timestamp"
}

func (f *TimestampFunction) Description() string {
	return "現在のUnixタイムスタンプ（秒）を取得します"
}

func (f *TimestampFunction) Execute(_ context.Context, args []interface{}) (interface{}, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("timestamp function expects no arguments, got %d", len(args))
	}

	return time.Now().Unix(), nil
}

// DateFunction は現在の日付取得関数
type DateFunction struct{}

func (f *DateFunction) Name() string {
	return "date"
}

func (f *DateFunction) Signature() string {
	return "!date [format]"
}

func (f *DateFunction) Description() string {
	return "現在の日付を取得します（デフォルト: YYYY-MM-DD）"
}

func (f *DateFunction) Execute(_ context.Context, args []interface{}) (interface{}, error) {
	if len(args) > 1 {
		return nil, fmt.Errorf("date function expects 0 or 1 argument, got %d", len(args))
	}

	format := "2006-01-02" // デフォルトフォーマット (YYYY-MM-DD)

	if len(args) == 1 {
		if customFormat, ok := args[0].(string); ok {
			format = customFormat
		} else {
			return nil, fmt.Errorf("date function expects string format argument")
		}
	}

	return time.Now().Format(format), nil
}

// TimeFunction は現在の時刻取得関数
type TimeFunction struct{}

func (f *TimeFunction) Name() string {
	return "time"
}

func (f *TimeFunction) Signature() string {
	return "!time [format]"
}

func (f *TimeFunction) Description() string {
	return "現在の時刻を取得します（デフォルト: HH:MM:SS）"
}

func (f *TimeFunction) Execute(_ context.Context, args []interface{}) (interface{}, error) {
	if len(args) > 1 {
		return nil, fmt.Errorf("time function expects 0 or 1 argument, got %d", len(args))
	}

	format := "15:04:05" // デフォルトフォーマット (HH:MM:SS)

	if len(args) == 1 {
		if customFormat, ok := args[0].(string); ok {
			format = customFormat
		} else {
			return nil, fmt.Errorf("time function expects string format argument")
		}
	}

	return time.Now().Format(format), nil
}
