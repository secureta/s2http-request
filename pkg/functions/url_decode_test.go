package functions

import (
	"context"
	"testing"
)

func TestURLDecodeFunction(t *testing.T) {
	tests := []struct {
		name      string
		args      []interface{}
		expected  string
		wantError bool
	}{
		{
			name:      "encoded string",
			args:      []interface{}{"hello+world"},
			expected:  "hello world",
			wantError: false,
		},
		{
			name:      "encoded special characters",
			args:      []interface{}{"hello%40world.com"},
			expected:  "hello@world.com",
			wantError: false,
		},
		{
			name:      "invalid encoding",
			args:      []interface{}{"hello%ZZ"},
			expected:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := &URLDecodeFunction{}
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