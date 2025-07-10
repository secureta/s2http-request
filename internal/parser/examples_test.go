package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseExamples(t *testing.T) {
	exampleDir := "../../examples"
	entries, err := os.ReadDir(exampleDir)
	if err != nil {
		t.Fatalf("Failed to read examples directory: %v", err)
	}

	parser := NewParser()

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()
		filePath := filepath.Join(exampleDir, fileName)
		if strings.HasPrefix(fileName, ".") {
			continue
		}

		t.Run(fileName, func(t *testing.T) {
			data, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("Failed to read example file: %v", err)
			}

			ext := filepath.Ext(fileName)
			_, err = parser.Parse(data, ext, filePath)
			if err != nil {
				t.Errorf("Failed to parse example file %s: %v", fileName, err)
			}
		})
	}
}
