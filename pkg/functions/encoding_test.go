package functions

import (
	"context"
	"testing"
)

func TestURLEncodeFunction(t *testing.T) {
	tests := []struct {
		name      string
		args      []interface{}
		expected  string
		wantError bool
	}{
		{
			name:      "simple string",
			args:      []interface{}{"hello world"},
			expected:  "hello+world",
			wantError: false,
		},
		{
			name:      "special characters",
			args:      []interface{}{"hello@world.com"},
			expected:  "hello%40world.com",
			wantError: false,
		},
		{
			name:      "empty string",
			args:      []interface{}{""},
			expected:  "",
			wantError: false,
		},
		{
			name:      "wrong number of arguments",
			args:      []interface{}{"arg1", "arg2", "arg3", "arg4"},
			expected:  "",
			wantError: true,
		},
		{
			name:      "non-string argument",
			args:      []interface{}{123},
			expected:  "",
			wantError: true,
		},
		{
			name:      "with custom unescaped characters",
			args:      []interface{}{"hello/world", "/"},
			expected:  "hello/world",
			wantError: false,
		},
		{
			name:      "with multiple custom unescaped characters",
			args:      []interface{}{"a=b&c=d", "&="},
			expected:  "a=b&c=d",
			wantError: false,
		},
		{
			name:      "double encode",
			args:      []interface{}{"hello world", 2.0},
			expected:  "hello%2Bworld",
			wantError: false,
		},
		{
			name:      "triple encode",
			args:      []interface{}{"hello world", 3.0},
			expected:  "hello%252Bworld",
			wantError: false,
		},
		{
			name:      "double encode with chars to not encode",
			args:      []interface{}{"hello/world", 2.0, "/"},
			expected:  "hello/world",
			wantError: false,
		},
		{
			name:      "invalid times argument type",
			args:      []interface{}{"hello", true},
			expected:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := &URLEncodeFunction{}
			ctx := context.Background()

			result, err := fn.Execute(ctx, tt.args)

			if tt.wantError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.wantError && result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
