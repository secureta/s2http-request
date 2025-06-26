package functions

import (
	"context"
	"testing"
)

func TestConcatArraysFunction(t *testing.T) {
	fn := &ConcatArraysFunction{}
	ctx := context.Background()

	// 複数の配列を結合
	arr1 := []string{"payload1", "payload2"}
	arr2 := []string{"payload3", "payload4"}
	arr3 := []string{"payload5"}

	result, err := fn.Execute(ctx, []interface{}{arr1, arr2, arr3})
	if err != nil {
		t.Fatalf("concat_arrays failed: %v", err)
	}

	resultSlice, ok := result.([]string)
	if !ok {
		t.Fatalf("Expected []string, got %T", result)
	}

	if len(resultSlice) != 5 {
		t.Fatalf("Expected 5 items, got %d", len(resultSlice))
	}

	expected := []string{"payload1", "payload2", "payload3", "payload4", "payload5"}
	for i, item := range expected {
		if resultSlice[i] != item {
			t.Errorf("Expected '%s' at index %d, got '%s'", item, i, resultSlice[i])
		}
	}
}

func TestConcatArraysFunction_MixedTypes(t *testing.T) {
	fn := &ConcatArraysFunction{}
	ctx := context.Background()

	// 異なる型の配列を結合
	arr1 := []string{"string1", "string2"}
	arr2 := []interface{}{"interface1", 123, true}
	singleString := "single"

	result, err := fn.Execute(ctx, []interface{}{arr1, arr2, singleString})
	if err != nil {
		t.Fatalf("concat_arrays failed: %v", err)
	}

	resultSlice, ok := result.([]string)
	if !ok {
		t.Fatalf("Expected []string, got %T", result)
	}

	if len(resultSlice) != 6 {
		t.Fatalf("Expected 6 items, got %d", len(resultSlice))
	}

	expected := []string{"string1", "string2", "interface1", "123", "true", "single"}
	for i, item := range expected {
		if resultSlice[i] != item {
			t.Errorf("Expected '%s' at index %d, got '%s'", item, i, resultSlice[i])
		}
	}
}

func TestConcatArraysFunction_EmptyArgs(t *testing.T) {
	fn := &ConcatArraysFunction{}
	ctx := context.Background()

	result, err := fn.Execute(ctx, []interface{}{})
	if err != nil {
		t.Fatalf("concat_arrays failed: %v", err)
	}

	resultSlice, ok := result.([]string)
	if !ok {
		t.Fatalf("Expected []string, got %T", result)
	}

	if len(resultSlice) != 0 {
		t.Fatalf("Expected 0 items, got %d", len(resultSlice))
	}
}

func TestConcatArraysFunction_SingleArray(t *testing.T) {
	fn := &ConcatArraysFunction{}
	ctx := context.Background()

	arr := []string{"only", "one", "array"}

	result, err := fn.Execute(ctx, []interface{}{arr})
	if err != nil {
		t.Fatalf("concat_arrays failed: %v", err)
	}

	resultSlice, ok := result.([]string)
	if !ok {
		t.Fatalf("Expected []string, got %T", result)
	}

	if len(resultSlice) != 3 {
		t.Fatalf("Expected 3 items, got %d", len(resultSlice))
	}

	for i, expected := range arr {
		if resultSlice[i] != expected {
			t.Errorf("Expected '%s' at index %d, got '%s'", expected, i, resultSlice[i])
		}
	}
}