package functions

import (
	"context"
	"testing"
)

func TestRandomFunction(t *testing.T) {
	tests := []struct {
		name      string
		args      []interface{}
		wantError bool
		checkFunc func(interface{}) bool
	}{
		{
			name:      "valid integer",
			args:      []interface{}{10},
			wantError: false,
			checkFunc: func(result interface{}) bool {
				val, ok := result.(int64)
				return ok && val >= 0 && val < 10
			},
		},
		{
			name:      "valid float64",
			args:      []interface{}{10.0},
			wantError: false,
			checkFunc: func(result interface{}) bool {
				val, ok := result.(int64)
				return ok && val >= 0 && val < 10
			},
		},
		{
			name:      "zero value",
			args:      []interface{}{0},
			wantError: true,
			checkFunc: nil,
		},
		{
			name:      "negative value",
			args:      []interface{}{-5},
			wantError: true,
			checkFunc: nil,
		},
		{
			name:      "wrong argument type",
			args:      []interface{}{"not_a_number"},
			wantError: true,
			checkFunc: nil,
		},
		{
			name:      "wrong number of arguments",
			args:      []interface{}{10, 20},
			wantError: true,
			checkFunc: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := &RandomFunction{}
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

func TestRandomStringFunction(t *testing.T) {
	tests := []struct {
		name      string
		args      []interface{}
		wantError bool
		checkFunc func(interface{}) bool
	}{
		{
			name:      "valid length",
			args:      []interface{}{10},
			wantError: false,
			checkFunc: func(result interface{}) bool {
				str, ok := result.(string)
				return ok && len(str) == 10
			},
		},
		{
			name:      "custom charset",
			args:      []interface{}{5, "ABC"},
			wantError: false,
			checkFunc: func(result interface{}) bool {
				str, ok := result.(string)
				if !ok || len(str) != 5 {
					return false
				}
				// Check if all characters are from the custom charset
				for _, char := range str {
					if char != 'A' && char != 'B' && char != 'C' {
						return false
					}
				}
				return true
			},
		},
		{
			name:      "zero length",
			args:      []interface{}{0},
			wantError: true,
			checkFunc: nil,
		},
		{
			name:      "empty charset",
			args:      []interface{}{5, ""},
			wantError: true,
			checkFunc: nil,
		},
		{
			name:      "wrong argument type",
			args:      []interface{}{"not_a_number"},
			wantError: true,
			checkFunc: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := &RandomStringFunction{}
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

func TestUUIDFunction(t *testing.T) {
	tests := []struct {
		name      string
		args      []interface{}
		wantError bool
	}{
		{
			name:      "no arguments",
			args:      []interface{}{},
			wantError: false,
		},
		{
			name:      "with arguments",
			args:      []interface{}{"unexpected"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := &UUIDFunction{}
			ctx := context.Background()
			
			result, err := fn.Execute(ctx, tt.args)
			
			if tt.wantError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.wantError {
				// Check if result is a valid UUID string format
				uuid, ok := result.(string)
				if !ok {
					t.Errorf("Expected string result, got %T", result)
				}
				if len(uuid) != 36 {
					t.Errorf("Expected UUID length 36, got %d", len(uuid))
				}
				// Basic UUID format check (8-4-4-4-12)
				if uuid[8] != '-' || uuid[13] != '-' || uuid[18] != '-' || uuid[23] != '-' {
					t.Errorf("Invalid UUID format: %s", uuid)
				}
			}
		})
	}
}

func TestUUIDFunctionUniqueness(t *testing.T) {
	fn := &UUIDFunction{}
	ctx := context.Background()
	
	// Generate multiple UUIDs and check they're unique
	uuids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		result, err := fn.Execute(ctx, []interface{}{})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		uuid := result.(string)
		if uuids[uuid] {
			t.Errorf("Duplicate UUID generated: %s", uuid)
		}
		uuids[uuid] = true
	}
}