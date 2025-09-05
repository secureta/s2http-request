package functions

import (
	"context"
	"testing"
)

func TestDictFunction(t *testing.T) {
	tests := []struct {
		name      string
		dictVars  map[string]interface{}
		args      []interface{}
		expected  interface{}
		wantError bool
	}{
		{
			name:      "existing dict variable",
			dictVars:  map[string]interface{}{"user_name": "Alice"},
			args:      []interface{}{"user_name"},
			expected:  "Alice",
			wantError: false,
		},
		{
			name:      "dict variable with number",
			dictVars:  map[string]interface{}{"user_age": 25},
			args:      []interface{}{"user_age"},
			expected:  25,
			wantError: false,
		},
		{
			name:      "dict variable with boolean",
			dictVars:  map[string]interface{}{"is_active": true},
			args:      []interface{}{"is_active"},
			expected:  true,
			wantError: false,
		},
		{
			name:      "dict variable with array",
			dictVars:  map[string]interface{}{"tags": []interface{}{"tag1", "tag2"}},
			args:      []interface{}{"tags"},
			expected:  []interface{}{"tag1", "tag2"},
			wantError: false,
		},
		{
			name:      "dict variable with object",
			dictVars:  map[string]interface{}{"config": map[string]interface{}{"key": "value"}},
			args:      []interface{}{"config"},
			expected:  map[string]interface{}{"key": "value"},
			wantError: false,
		},
		{
			name:      "non-existing dict variable",
			dictVars:  map[string]interface{}{},
			args:      []interface{}{"missing"},
			expected:  nil,
			wantError: true,
		},
		{
			name:      "empty dict variables",
			dictVars:  map[string]interface{}{},
			args:      []interface{}{"any_var"},
			expected:  nil,
			wantError: true,
		},
		{
			name:      "wrong number of arguments - no args",
			dictVars:  map[string]interface{}{"test": "value"},
			args:      []interface{}{},
			expected:  nil,
			wantError: true,
		},
		{
			name:      "wrong number of arguments - multiple args",
			dictVars:  map[string]interface{}{"test": "value"},
			args:      []interface{}{"arg1", "arg2"},
			expected:  nil,
			wantError: true,
		},
		{
			name:      "non-string argument",
			dictVars:  map[string]interface{}{"test": "value"},
			args:      []interface{}{123},
			expected:  nil,
			wantError: true,
		},
		{
			name:      "nil dict variables in context",
			dictVars:  nil,
			args:      []interface{}{"test"},
			expected:  nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := &DictFunction{}

			// Create context with dict variables
			ctx := context.Background()
			if tt.dictVars != nil {
				ctx = context.WithValue(ctx, "dict", tt.dictVars)
			}

			result, err := fn.Execute(ctx, tt.args)

			if tt.wantError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.wantError {
				// For complex types, use deep comparison
				if !deepEqual(result, tt.expected) {
					t.Errorf("Expected %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

func TestDictFunctionMetadata(t *testing.T) {
	fn := &DictFunction{}

	if fn.Name() != "dict" {
		t.Errorf("Expected name 'dict', got '%s'", fn.Name())
	}

	signature := fn.Signature()
	if signature == "" {
		t.Error("Signature should not be empty")
	}

	description := fn.Description()
	if description == "" {
		t.Error("Description should not be empty")
	}
}

func TestDictFunctionWithoutContext(t *testing.T) {
	fn := &DictFunction{}
	ctx := context.Background()

	_, err := fn.Execute(ctx, []interface{}{"test"})
	if err == nil {
		t.Error("Expected error when dict variables not in context")
	}
}

func TestDictFunctionWithWrongContextType(t *testing.T) {
	fn := &DictFunction{}
	ctx := context.WithValue(context.Background(), "dict", "not_a_map")

	_, err := fn.Execute(ctx, []interface{}{"test"})
	if err == nil {
		t.Error("Expected error when dict context is not a map")
	}
}

// deepEqual performs deep comparison for complex types
func deepEqual(a, b interface{}) bool {
	switch va := a.(type) {
	case []interface{}:
		vb, ok := b.([]interface{})
		if !ok || len(va) != len(vb) {
			return false
		}
		for i := range va {
			if !deepEqual(va[i], vb[i]) {
				return false
			}
		}
		return true
	case map[string]interface{}:
		vb, ok := b.(map[string]interface{})
		if !ok || len(va) != len(vb) {
			return false
		}
		for k, v := range va {
			if !deepEqual(v, vb[k]) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}
