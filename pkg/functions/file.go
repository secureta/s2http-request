package functions

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// FileFunction implements file reading functionality
type FileFunction struct{}

// Name returns the function name
func (f *FileFunction) Name() string {
	return "file"
}

// Signature returns the function signature
func (f *FileFunction) Signature() string {
	return "$file <file_path>"
}

// Description returns the function description
func (f *FileFunction) Description() string {
	return "Reads file content from relative path and returns as string."
}

// Execute reads a file and returns its content as string
func (f *FileFunction) Execute(ctx context.Context, args []interface{}) (interface{}, error) {
	// Validate arguments
	if len(args) != 1 {
		return nil, fmt.Errorf("file function requires exactly one argument, got %d", len(args))
	}
	
	filePath, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("file function argument must be a string, got %T", args[0])
	}
	
	// Security check: reject absolute paths
	if filepath.IsAbs(filePath) {
		return nil, fmt.Errorf("absolute paths are not allowed")
	}
	
	// Determine the base directory for resolving relative paths
	var baseDir string
	if requestFilePath, ok := ctx.Value("requestFilePath").(string); ok && requestFilePath != "" {
		// Use the directory containing the request definition file
		baseDir = filepath.Dir(requestFilePath)
	} else {
		// Use current working directory (stdin mode)
		var err error
		baseDir, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current working directory: %w", err)
		}
	}
	
	// Resolve the full path
	fullPath := filepath.Join(baseDir, filePath)
	
	// Security check: ensure the resolved path doesn't escape the base directory
	cleanPath := filepath.Clean(fullPath)
	cleanBase := filepath.Clean(baseDir)
	
	// Check if the clean path is within the base directory
	relPath, err := filepath.Rel(cleanBase, cleanPath)
	if err != nil || strings.HasPrefix(relPath, "..") {
		return nil, fmt.Errorf("path escapes request file directory")
	}
	
	// Read the file
	file, err := os.Open(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	
	// Read file content and return as string
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	
	// Always return content as string
	return string(content), nil
}