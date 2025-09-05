package parser

import (
	"encoding/json"
	"strings"

	"gopkg.in/yaml.v3"
)

// PositionInfo holds position information for a parsed element
type PositionInfo struct {
	Line   int
	Column int
	Offset int
}

// PositionTracker tracks position information during parsing
type PositionTracker struct {
	filePath string
	content  string
	lines    []string
}

// NewPositionTracker creates a new position tracker
func NewPositionTracker(filePath string, content []byte) *PositionTracker {
	contentStr := string(content)
	return &PositionTracker{
		filePath: filePath,
		content:  contentStr,
		lines:    strings.Split(contentStr, "\n"),
	}
}

// GetLineContent returns the content of a specific line (1-based)
func (pt *PositionTracker) GetLineContent(lineNumber int) string {
	if lineNumber < 1 || lineNumber > len(pt.lines) {
		return ""
	}
	return pt.lines[lineNumber-1]
}

// FindJSONPosition attempts to find the position of a JSON property
func (pt *PositionTracker) FindJSONPosition(propertyPath string) *PositionInfo {
	// For JSON, we'll do a simple text search for the property
	// This is a basic implementation - more sophisticated parsing could be added
	parts := strings.Split(propertyPath, ".")
	if len(parts) == 0 {
		return nil
	}

	// Look for the property in the content
	searchKey := `"` + parts[len(parts)-1] + `"`
	lines := pt.lines

	for i, line := range lines {
		if strings.Contains(line, searchKey) {
			column := strings.Index(line, searchKey) + 1
			return &PositionInfo{
				Line:   i + 1,
				Column: column,
			}
		}
	}

	return &PositionInfo{Line: 1, Column: 1}
}

// FindYAMLPosition attempts to find the position of a YAML property using yaml.v3 Node information
func (pt *PositionTracker) FindYAMLPosition(propertyPath string) *PositionInfo {
	// Parse YAML to get node information
	var node yaml.Node
	if err := yaml.Unmarshal([]byte(pt.content), &node); err != nil {
		return &PositionInfo{Line: 1, Column: 1}
	}

	// Navigate to the property
	parts := strings.Split(propertyPath, ".")
	currentNode := &node

	// Find the document node (usually the first content node)
	if len(currentNode.Content) > 0 {
		currentNode = currentNode.Content[0]
	}

	for _, part := range parts {
		if part == "" {
			continue
		}

		found := false
		// Look for the key in mapping nodes
		if currentNode.Kind == yaml.MappingNode {
			for i := 0; i < len(currentNode.Content); i += 2 {
				keyNode := currentNode.Content[i]
				valueNode := currentNode.Content[i+1]

				if keyNode.Value == part {
					currentNode = valueNode
					found = true
					break
				}
			}
		}

		if !found {
			break
		}
	}

	return &PositionInfo{
		Line:   currentNode.Line,
		Column: currentNode.Column,
	}
}

// FindJSONLPosition finds position in JSONL format
func (pt *PositionTracker) FindJSONLPosition(propertyPath string) *PositionInfo {
	// For JSONL, we'll look at the first valid JSON line
	lines := pt.lines

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		// Try to parse as JSON to verify it's valid
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			continue
		}

		// Find the property in this line
		parts := strings.Split(propertyPath, ".")
		if len(parts) > 0 {
			searchKey := `"` + parts[len(parts)-1] + `"`
			if strings.Contains(line, searchKey) {
				column := strings.Index(line, searchKey) + 1
				return &PositionInfo{
					Line:   i + 1,
					Column: column,
				}
			}
		}

		// Return the line position even if property not found
		return &PositionInfo{
			Line:   i + 1,
			Column: 1,
		}
	}

	return &PositionInfo{Line: 1, Column: 1}
}

// GetPosition returns position information for a property path based on file extension
func (pt *PositionTracker) GetPosition(propertyPath string, fileExt string) *PositionInfo {
	switch strings.ToLower(fileExt) {
	case ".json":
		return pt.FindJSONPosition(propertyPath)
	case ".yaml", ".yml":
		return pt.FindYAMLPosition(propertyPath)
	case ".jsonl":
		return pt.FindJSONLPosition(propertyPath)
	default:
		return &PositionInfo{Line: 1, Column: 1}
	}
}
