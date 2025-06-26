package functions

import (
	"context"
	"testing"
	"time"
)

func TestTimestampFunction(t *testing.T) {
	tests := []struct {
		name      string
		args      []interface{}
		wantError bool
		checkFunc func(interface{}) bool
	}{
		{
			name:      "no arguments",
			args:      []interface{}{},
			wantError: false,
			checkFunc: func(result interface{}) bool {
				timestamp, ok := result.(int64)
				if !ok {
					return false
				}
				// タイムスタンプが現在時刻の近く（±10秒以内）であることを確認
				now := time.Now().Unix()
				return timestamp >= now-10 && timestamp <= now+10
			},
		},
		{
			name:      "with arguments",
			args:      []interface{}{"unexpected"},
			wantError: true,
			checkFunc: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := &TimestampFunction{}
			ctx := context.Background()

			result, err := fn.Execute(ctx, tt.args)

			if tt.wantError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.wantError && tt.checkFunc != nil && !tt.checkFunc(result) {
				t.Errorf("Result validation failed for %v", result)
			}
		})
	}
}

func TestDateFunction(t *testing.T) {
	tests := []struct {
		name      string
		args      []interface{}
		expected  string
		wantError bool
		checkFunc func(string) bool
	}{
		{
			name:      "default format",
			args:      []interface{}{},
			expected:  "",
			wantError: false,
			checkFunc: func(result string) bool {
				// YYYY-MM-DD形式であることを確認
				if len(result) != 10 {
					return false
				}
				if result[4] != '-' || result[7] != '-' {
					return false
				}
				return true
			},
		},
		{
			name:      "custom format",
			args:      []interface{}{"2006/01/02"},
			expected:  "",
			wantError: false,
			checkFunc: func(result string) bool {
				// YYYY/MM/DD形式であることを確認
				if len(result) != 10 {
					return false
				}
				if result[4] != '/' || result[7] != '/' {
					return false
				}
				return true
			},
		},
		{
			name:      "time format",
			args:      []interface{}{"15:04:05"},
			expected:  "",
			wantError: false,
			checkFunc: func(result string) bool {
				// HH:MM:SS形式であることを確認
				if len(result) != 8 {
					return false
				}
				if result[2] != ':' || result[5] != ':' {
					return false
				}
				return true
			},
		},
		{
			name:      "too many arguments",
			args:      []interface{}{"format1", "format2"},
			expected:  "",
			wantError: true,
			checkFunc: nil,
		},
		{
			name:      "non-string format",
			args:      []interface{}{123},
			expected:  "",
			wantError: true,
			checkFunc: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := &DateFunction{}
			ctx := context.Background()

			result, err := fn.Execute(ctx, tt.args)

			if tt.wantError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.wantError && tt.checkFunc != nil {
				resultStr, ok := result.(string)
				if !ok {
					t.Errorf("Expected string result, got %T", result)
				}
				if !tt.checkFunc(resultStr) {
					t.Errorf("Result validation failed for %q", resultStr)
				}
			}
		})
	}
}

func TestTimeFunction(t *testing.T) {
	tests := []struct {
		name      string
		args      []interface{}
		expected  string
		wantError bool
		checkFunc func(string) bool
	}{
		{
			name:      "default format",
			args:      []interface{}{},
			expected:  "",
			wantError: false,
			checkFunc: func(result string) bool {
				// HH:MM:SS形式であることを確認
				if len(result) != 8 {
					return false
				}
				if result[2] != ':' || result[5] != ':' {
					return false
				}
				return true
			},
		},
		{
			name:      "custom format",
			args:      []interface{}{"15:04"},
			expected:  "",
			wantError: false,
			checkFunc: func(result string) bool {
				// HH:MM形式であることを確認
				if len(result) != 5 {
					return false
				}
				if result[2] != ':' {
					return false
				}
				return true
			},
		},
		{
			name:      "12-hour format",
			args:      []interface{}{"03:04:05 PM"},
			expected:  "",
			wantError: false,
			checkFunc: func(result string) bool {
				// HH:MM:SS AM/PM形式であることを確認
				if len(result) != 11 {
					return false
				}
				if result[2] != ':' || result[5] != ':' {
					return false
				}
				return result[9:] == "AM" || result[9:] == "PM"
			},
		},
		{
			name:      "too many arguments",
			args:      []interface{}{"format1", "format2"},
			expected:  "",
			wantError: true,
			checkFunc: nil,
		},
		{
			name:      "non-string format",
			args:      []interface{}{123},
			expected:  "",
			wantError: true,
			checkFunc: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := &TimeFunction{}
			ctx := context.Background()

			result, err := fn.Execute(ctx, tt.args)

			if tt.wantError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.wantError && tt.checkFunc != nil {
				resultStr, ok := result.(string)
				if !ok {
					t.Errorf("Expected string result, got %T", result)
				}
				if !tt.checkFunc(resultStr) {
					t.Errorf("Result validation failed for %q", resultStr)
				}
			}
		})
	}
}

func TestTimestampFunctionName(t *testing.T) {
	fn := &TimestampFunction{}
	if fn.Name() != "timestamp" {
		t.Errorf("Expected function name 'timestamp', got %q", fn.Name())
	}
}

func TestDateFunctionName(t *testing.T) {
	fn := &DateFunction{}
	if fn.Name() != "date" {
		t.Errorf("Expected function name 'date', got %q", fn.Name())
	}
}

func TestTimeFunctionName(t *testing.T) {
	fn := &TimeFunction{}
	if fn.Name() != "time" {
		t.Errorf("Expected function name 'time', got %q", fn.Name())
	}
}

func TestTimeFunctionConsistency(t *testing.T) {
	// 短時間内での複数回実行で一貫性をテスト
	fn := &TimestampFunction{}
	ctx := context.Background()

	// 最初のタイムスタンプ取得
	result1, err := fn.Execute(ctx, []interface{}{})
	if err != nil {
		t.Fatalf("First timestamp failed: %v", err)
	}
	timestamp1 := result1.(int64)

	// 少し待機
	time.Sleep(1 * time.Second)

	// 2回目のタイムスタンプ取得
	result2, err := fn.Execute(ctx, []interface{}{})
	if err != nil {
		t.Fatalf("Second timestamp failed: %v", err)
	}
	timestamp2 := result2.(int64)

	// 2回目のタイムスタンプが1回目より大きいことを確認
	if timestamp2 <= timestamp1 {
		t.Errorf("Expected second timestamp (%d) to be greater than first (%d)", timestamp2, timestamp1)
	}

	// 差が妥当な範囲内であることを確認（1-3秒程度）
	diff := timestamp2 - timestamp1
	if diff < 1 || diff > 3 {
		t.Errorf("Expected timestamp difference to be 1-3 seconds, got %d", diff)
	}
}

func TestDateTimeFormatConsistency(t *testing.T) {
	// 日付と時刻関数が同じ時刻を参照していることを確認
	dateFn := &DateFunction{}
	timeFn := &TimeFunction{}
	ctx := context.Background()

	// 同じフォーマットで日付と時刻を取得
	dateResult, err := dateFn.Execute(ctx, []interface{}{"2006-01-02 15:04:05"})
	if err != nil {
		t.Fatalf("Date function failed: %v", err)
	}

	timeResult, err := timeFn.Execute(ctx, []interface{}{"2006-01-02 15:04:05"})
	if err != nil {
		t.Fatalf("Time function failed: %v", err)
	}

	dateStr := dateResult.(string)
	timeStr := timeResult.(string)

	// 両方とも同じフォーマットで返されることを確認
	if len(dateStr) != 19 || len(timeStr) != 19 {
		t.Errorf("Expected both results to be 19 characters long, got date: %d, time: %d", len(dateStr), len(timeStr))
	}

	// 秒単位での差が小さいことを確認（ほぼ同時実行のため）
	dateTime, err := time.Parse("2006-01-02 15:04:05", dateStr)
	if err != nil {
		t.Fatalf("Failed to parse date result: %v", err)
	}

	timeTime, err := time.Parse("2006-01-02 15:04:05", timeStr)
	if err != nil {
		t.Fatalf("Failed to parse time result: %v", err)
	}

	diff := timeTime.Sub(dateTime)
	if diff < 0 {
		diff = -diff
	}

	// 1秒以内の差であることを確認
	if diff > time.Second {
		t.Errorf("Expected time difference to be less than 1 second, got %v", diff)
	}
}

func TestTimeFormatEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		function Function
		format   string
		validate func(string) bool
	}{
		{
			name:     "date with year only",
			function: &DateFunction{},
			format:   "2006",
			validate: func(result string) bool {
				return len(result) == 4
			},
		},
		{
			name:     "time with milliseconds",
			function: &TimeFunction{},
			format:   "15:04:05.000",
			validate: func(result string) bool {
				return len(result) == 12 && result[8] == '.'
			},
		},
		{
			name:     "date with weekday",
			function: &DateFunction{},
			format:   "Monday, 2006-01-02",
			validate: func(result string) bool {
				return len(result) > 10 // 曜日が含まれるので10文字より長い
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := tt.function.Execute(ctx, []interface{}{tt.format})

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			resultStr, ok := result.(string)
			if !ok {
				t.Errorf("Expected string result, got %T", result)
			}

			if !tt.validate(resultStr) {
				t.Errorf("Validation failed for result %q with format %q", resultStr, tt.format)
			}
		})
	}
}