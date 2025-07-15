package functions

import (
	"context"
	"testing"
)

func TestVarFunction(t *testing.T) {
	tests := []struct {
		name      string
		variables map[string]interface{}
		args      []interface{}
		expected  interface{}
		wantError bool
	}{
		{
			name:      "existing variable",
			variables: map[string]interface{}{"test": "value"},
			args:      []interface{}{"test"},
			expected:  "value",
			wantError: false,
		},
		{
			name:      "non-existing variable",
			variables: map[string]interface{}{},
			args:      []interface{}{"missing"},
			expected:  nil,
			wantError: true,
		},
		{
			name:      "wrong number of arguments",
			variables: map[string]interface{}{},
			args:      []interface{}{"arg1", "arg2"},
			expected:  nil,
			wantError: true,
		},
		{
			name:      "non-string argument",
			variables: map[string]interface{}{},
			args:      []interface{}{123},
			expected:  nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := &VarFunction{}
			ctx := context.WithValue(context.Background(), "variables", tt.variables)

			result, err := fn.Execute(ctx, tt.args)

			if tt.wantError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.wantError && result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestConcatFunction(t *testing.T) {
	tests := []struct {
		name     string
		args     []interface{}
		expected string
	}{
		{
			name:     "empty args",
			args:     []interface{}{},
			expected: "",
		},
		{
			name:     "single string",
			args:     []interface{}{"hello"},
			expected: "hello",
		},
		{
			name:     "multiple strings",
			args:     []interface{}{"hello", " ", "world"},
			expected: "hello world",
		},
		{
			name:     "mixed types",
			args:     []interface{}{"count:", 42, " items"},
			expected: "count:42 items",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := &ConcatFunction{}
			ctx := context.Background()

			result, err := fn.Execute(ctx, tt.args)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestJoinFunction(t *testing.T) {
	tests := []struct {
		name      string
		args      []interface{}
		expected  string
		wantError bool
	}{
		{
			name: "array with no delimiter",
			args: []interface{}{map[string]interface{}{
				"values": []interface{}{"a", "b", "c"},
			}},
			expected:  "abc",
			wantError: false,
		},
		{
			name: "array with comma delimiter",
			args: []interface{}{map[string]interface{}{
				"values":    []interface{}{"a", "b", "c"},
				"delimiter": ",",
			}},
			expected:  "a,b,c",
			wantError: false,
		},
		{
			name: "array with space delimiter",
			args: []interface{}{map[string]interface{}{
				"values":    []interface{}{"hello", "world"},
				"delimiter": " ",
			}},
			expected:  "hello world",
			wantError: false,
		},
		{
			name: "string array with delimiter",
			args: []interface{}{map[string]interface{}{
				"values":    []string{"x", "y", "z"},
				"delimiter": "-",
			}},
			expected:  "x-y-z",
			wantError: false,
		},
		{
			name: "single string value",
			args: []interface{}{map[string]interface{}{
				"values": "single",
			}},
			expected:  "single",
			wantError: false,
		},
		{
			name: "empty array",
			args: []interface{}{map[string]interface{}{
				"values": []interface{}{},
			}},
			expected:  "",
			wantError: false,
		},
		{
			name:      "no arguments",
			args:      []interface{}{},
			expected:  "",
			wantError: true,
		},
		{
			name: "non-string delimiter",
			args: []interface{}{map[string]interface{}{
				"values":    []interface{}{"a", "b"},
				"delimiter": 123,
			}},
			expected:  "",
			wantError: true,
		},
		{
			name: "invalid values type",
			args: []interface{}{map[string]interface{}{
				"values": 123,
			}},
			expected:  "",
			wantError: true,
		},
		{
			name: "missing values field",
			args: []interface{}{map[string]interface{}{
				"delimiter": ",",
			}},
			expected:  "",
			wantError: true,
		},
		{
			name:      "non-map argument",
			args:      []interface{}{123},
			expected:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := &JoinFunction{}
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
