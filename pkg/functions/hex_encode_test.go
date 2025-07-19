package functions

import (
	"context"
	"testing"
)

func TestHexEncodeFunction(t *testing.T) {
	fn := &HexEncodeFunction{}

	tests := []struct {
		name     string
		args     []interface{}
		expected string
		hasError bool
	}{
		{
			name:     "basic hex encoding",
			args:     []interface{}{"hello"},
			expected: "68656c6c6f",
			hasError: false,
		},
		{
			name:     "empty string",
			args:     []interface{}{""},
			expected: "",
			hasError: false,
		},
		{
			name:     "special characters",
			args:     []interface{}{"Hello, World!"},
			expected: "48656c6c6f2c20576f726c6421",
			hasError: false,
		},
		{
			name:     "unicode characters",
			args:     []interface{}{"こんにちは"},
			expected: "e38193e38293e381abe381a1e381af",
			hasError: false,
		},
		{
			name:     "no arguments",
			args:     []interface{}{},
			expected: "",
			hasError: true,
		},
		{
			name:     "too many arguments",
			args:     []interface{}{"hello", "world"},
			expected: "",
			hasError: true,
		},
		{
			name:     "non-string argument",
			args:     []interface{}{123},
			expected: "",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fn.Execute(context.Background(), tt.args)

			if tt.hasError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestHexEncodeFunctionMetadata(t *testing.T) {
	fn := &HexEncodeFunction{}

	if fn.Name() != "hex_encode" {
		t.Errorf("expected name 'hex_encode', got %q", fn.Name())
	}

	if fn.Signature() != "$hex_encode <string>" {
		t.Errorf("expected signature '$hex_encode <string>', got %q", fn.Signature())
	}

	if fn.Description() == "" {
		t.Error("expected non-empty description")
	}
}
