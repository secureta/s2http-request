package parser

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseExamples(t *testing.T) {
	exampleDir := "../../examples"
	files, err := ioutil.ReadDir(exampleDir)
	if err != nil {
		t.Fatalf("Failed to read examples directory: %v", err)
	}

	parser := NewParser()

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(exampleDir, file.Name())
		if strings.HasPrefix(file.Name(), ".") {
			continue
		}

		t.Run(file.Name(), func(t *testing.T) {
			data, err := ioutil.ReadFile(filePath)
			if err != nil {
				t.Fatalf("Failed to read example file: %v", err)
			}

			ext := filepath.Ext(file.Name())
			_, err = parser.Parse(data, ext, filePath)
			if err != nil {
				t.Errorf("Failed to parse example file %s: %v", file.Name(), err)
			}
		})
	}
}
