package parser

import (
	"testing"
)

func TestPositionTracker_GetLineContent(t *testing.T) {
	content := `line 1
line 2
line 3`

	tracker := NewPositionTracker("/test.txt", []byte(content))

	tests := []struct {
		line     int
		expected string
	}{
		{1, "line 1"},
		{2, "line 2"},
		{3, "line 3"},
		{0, ""},  // Invalid line number
		{4, ""},  // Line doesn't exist
		{-1, ""}, // Negative line number
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := tracker.GetLineContent(tt.line)
			if result != tt.expected {
				t.Errorf("GetLineContent(%d) = %q, expected %q", tt.line, result, tt.expected)
			}
		})
	}
}

func TestPositionTracker_FindJSONPosition(t *testing.T) {
	content := `{
  "method": "POST",
  "dict": {
    "user_id": [1, 2, 3],
    "name": ["Alice", "Bob"]
  }
}`

	tracker := NewPositionTracker("/test.json", []byte(content))

	tests := []struct {
		propertyPath string
		expectedLine int
		minColumn    int // Minimum expected column (exact position may vary)
	}{
		{"dict.user_id", 4, 1},
		{"dict.name", 5, 1},
		{"method", 2, 1},
		{"nonexistent", 1, 1}, // Should return default position
	}

	for _, tt := range tests {
		t.Run(tt.propertyPath, func(t *testing.T) {
			pos := tracker.FindJSONPosition(tt.propertyPath)
			if pos.Line != tt.expectedLine {
				t.Errorf("FindJSONPosition(%q).Line = %d, expected %d", tt.propertyPath, pos.Line, tt.expectedLine)
			}
			if pos.Column < tt.minColumn {
				t.Errorf("FindJSONPosition(%q).Column = %d, expected >= %d", tt.propertyPath, pos.Column, tt.minColumn)
			}
		})
	}
}

func TestPositionTracker_FindYAMLPosition(t *testing.T) {
	content := `method: POST
dict:
  user_id: [1, 2, 3]
  name: ["Alice", "Bob"]
body:
  test: value`

	tracker := NewPositionTracker("/test.yaml", []byte(content))

	tests := []struct {
		propertyPath string
		expectedLine int
	}{
		{"method", 1},
		{"dict", 3},
		{"body", 6},
	}

	for _, tt := range tests {
		t.Run(tt.propertyPath, func(t *testing.T) {
			pos := tracker.FindYAMLPosition(tt.propertyPath)
			if pos.Line != tt.expectedLine {
				t.Errorf("FindYAMLPosition(%q).Line = %d, expected %d", tt.propertyPath, pos.Line, tt.expectedLine)
			}
		})
	}
}

func TestPositionTracker_FindJSONLPosition(t *testing.T) {
	content := `# Comment line
{"method": "POST", "dict": {"user_id": [1, 2, 3]}}
// Another comment
{"method": "GET", "path": "/test"}`

	tracker := NewPositionTracker("/test.jsonl", []byte(content))

	tests := []struct {
		propertyPath string
		expectedLine int
		minColumn    int
	}{
		{"dict.user_id", 2, 1},
		{"method", 2, 1},
		{"nonexistent", 2, 1}, // Should return first valid JSON line
	}

	for _, tt := range tests {
		t.Run(tt.propertyPath, func(t *testing.T) {
			pos := tracker.FindJSONLPosition(tt.propertyPath)
			if pos.Line != tt.expectedLine {
				t.Errorf("FindJSONLPosition(%q).Line = %d, expected %d", tt.propertyPath, pos.Line, tt.expectedLine)
			}
			if pos.Column < tt.minColumn {
				t.Errorf("FindJSONLPosition(%q).Column = %d, expected >= %d", tt.propertyPath, pos.Column, tt.minColumn)
			}
		})
	}
}

func TestPositionTracker_GetPosition(t *testing.T) {
	content := `{"dict": {"user_id": [1, 2, 3]}}`
	tracker := NewPositionTracker("/test", []byte(content))

	tests := []struct {
		fileExt      string
		propertyPath string
		expectedLine int
	}{
		{".json", "dict.user_id", 1},
		{".yaml", "dict", 1},
		{".yml", "dict", 1},
		{".jsonl", "dict.user_id", 1},
		{".unknown", "dict", 1}, // Should return default position
	}

	for _, tt := range tests {
		t.Run(tt.fileExt, func(t *testing.T) {
			pos := tracker.GetPosition(tt.propertyPath, tt.fileExt)
			if pos.Line != tt.expectedLine {
				t.Errorf("GetPosition(%q, %q).Line = %d, expected %d", tt.propertyPath, tt.fileExt, pos.Line, tt.expectedLine)
			}
		})
	}
}

func TestNewPositionTracker(t *testing.T) {
	content := []byte("test content\nline 2")
	tracker := NewPositionTracker("/test.txt", content)

	if tracker.filePath != "/test.txt" {
		t.Errorf("filePath = %q, expected %q", tracker.filePath, "/test.txt")
	}

	if tracker.content != string(content) {
		t.Errorf("content = %q, expected %q", tracker.content, string(content))
	}

	expectedLines := []string{"test content", "line 2"}
	if len(tracker.lines) != len(expectedLines) {
		t.Errorf("lines length = %d, expected %d", len(tracker.lines), len(expectedLines))
		return
	}

	for i, expected := range expectedLines {
		if tracker.lines[i] != expected {
			t.Errorf("lines[%d] = %q, expected %q", i, tracker.lines[i], expected)
		}
	}
}
