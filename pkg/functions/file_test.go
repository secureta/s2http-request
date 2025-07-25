package functions

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFileFunction_Execute_BasicTextFile(t *testing.T) {
	// Create a temporary directory and file for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "Hello, World!"
	
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	fn := &FileFunction{}
	ctx := context.WithValue(context.Background(), "requestFilePath", filepath.Join(tempDir, "request.json"))
	
	result, err := fn.Execute(ctx, []interface{}{"test.txt"})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	
	if result != testContent {
		t.Errorf("Expected %q, got %q", testContent, result)
	}
}

func TestFileFunction_Execute_BinaryFile(t *testing.T) {
	// Create a temporary directory and binary file for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.png")
	binaryData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A} // PNG header
	
	err := os.WriteFile(testFile, binaryData, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	fn := &FileFunction{}
	ctx := context.WithValue(context.Background(), "requestFilePath", filepath.Join(tempDir, "request.json"))
	
	result, err := fn.Execute(ctx, []interface{}{"test.png"})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	
	// Should return binary content as string (not base64 encoded)
	expected := string(binaryData)
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestFileFunction_Execute_AbsolutePathRejection(t *testing.T) {
	fn := &FileFunction{}
	ctx := context.WithValue(context.Background(), "requestFilePath", "/tmp/request.json")
	
	_, err := fn.Execute(ctx, []interface{}{"/etc/passwd"})
	if err == nil {
		t.Error("Expected error for absolute path, got none")
	}
	
	expectedError := "absolute paths are not allowed"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}

func TestFileFunction_Execute_ParentDirectoryTraversal(t *testing.T) {
	// Create a temporary directory structure
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "subdir")
	err := os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	fn := &FileFunction{}
	ctx := context.WithValue(context.Background(), "requestFilePath", filepath.Join(subDir, "request.json"))
	
	_, err = fn.Execute(ctx, []interface{}{"../../../etc/passwd"})
	if err == nil {
		t.Error("Expected error for parent directory traversal, got none")
	}
	
	expectedError := "path escapes request file directory"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}

func TestFileFunction_Execute_FileNotFound(t *testing.T) {
	tempDir := t.TempDir()
	fn := &FileFunction{}
	ctx := context.WithValue(context.Background(), "requestFilePath", filepath.Join(tempDir, "request.json"))
	
	_, err := fn.Execute(ctx, []interface{}{"nonexistent.txt"})
	if err == nil {
		t.Error("Expected error for nonexistent file, got none")
	}
}

func TestFileFunction_Execute_InvalidArguments(t *testing.T) {
	fn := &FileFunction{}
	ctx := context.Background()
	
	// Test with no arguments
	_, err := fn.Execute(ctx, []interface{}{})
	if err == nil {
		t.Error("Expected error for no arguments, got none")
	}
	
	// Test with too many arguments
	_, err = fn.Execute(ctx, []interface{}{"file1.txt", "file2.txt"})
	if err == nil {
		t.Error("Expected error for too many arguments, got none")
	}
	
	// Test with non-string argument
	_, err = fn.Execute(ctx, []interface{}{123})
	if err == nil {
		t.Error("Expected error for non-string argument, got none")
	}
}

func TestFileFunction_Execute_StdinMode(t *testing.T) {
	// Test stdin mode (when requestFilePath is empty)
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "Hello from stdin mode!"
	
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Change to temp directory to simulate current working directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(oldWd)
	
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	fn := &FileFunction{}
	ctx := context.WithValue(context.Background(), "requestFilePath", "")
	
	result, err := fn.Execute(ctx, []interface{}{"test.txt"})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	
	if result != testContent {
		t.Errorf("Expected %q, got %q", testContent, result)
	}
}