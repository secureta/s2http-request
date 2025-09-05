package parser

import (
	"fmt"
	"sort"
	"strings"
)

// ErrorFormatter provides formatted error output with categorization
type ErrorFormatter struct{}

// NewErrorFormatter creates a new error formatter
func NewErrorFormatter() *ErrorFormatter {
	return &ErrorFormatter{}
}

// FormatError formats a single error with enhanced display
func (f *ErrorFormatter) FormatError(err error) string {
	switch e := err.(type) {
	case *ParseError:
		return f.formatParseError(e)
	case *DictValidationError:
		return f.formatDictValidationError(e)
	case *ErrorCollection:
		return f.formatErrorCollection(e)
	default:
		return err.Error()
	}
}

// formatParseError formats a ParseError with color coding and structure
func (f *ErrorFormatter) formatParseError(err *ParseError) string {
	var parts []string

	// File location with emphasis
	if err.FilePath != "" {
		location := fmt.Sprintf("📁 %s", err.FilePath)
		if err.LineNumber > 0 {
			location += fmt.Sprintf(":%d", err.LineNumber)
			if err.ColumnNumber > 0 {
				location += fmt.Sprintf(":%d", err.ColumnNumber)
			}
		}
		parts = append(parts, location)
	}

	// Property path with context
	if err.PropertyPath != "" {
		parts = append(parts, fmt.Sprintf("🎯 Property: %s", err.PropertyPath))
	}

	// Error level and message
	levelIcon := f.getLevelIcon(err.Level)
	parts = append(parts, fmt.Sprintf("%s %s: %s", levelIcon, err.Level.String(), err.Message))

	// Source line if available
	if err.SourceLine != "" {
		parts = append(parts, fmt.Sprintf("📝 Source: %s", strings.TrimSpace(err.SourceLine)))
	}

	return strings.Join(parts, "\n   ")
}

// formatDictValidationError formats a DictValidationError
func (f *ErrorFormatter) formatDictValidationError(err *DictValidationError) string {
	if err.ParseError != nil {
		formatted := f.formatParseError(err.ParseError)
		return fmt.Sprintf("🔧 Dict Configuration Error\n   %s\n   🔑 Dict Key: %s", formatted, err.DictKey)
	}
	return fmt.Sprintf("🔧 Dict Configuration Error for key '%s': %s", err.DictKey, err.Message)
}

// formatErrorCollection formats multiple errors with categorization
func (f *ErrorFormatter) formatErrorCollection(collection *ErrorCollection) string {
	if len(collection.Errors) == 0 {
		return "No errors"
	}

	if len(collection.Errors) == 1 {
		return f.FormatError(collection.Errors[0])
	}

	// Categorize errors
	categories := f.categorizeErrors(collection.Errors)

	var sections []string

	// Add summary
	sections = append(sections, fmt.Sprintf("❌ Configuration Validation Failed (%d errors)", len(collection.Errors)))
	sections = append(sections, "")

	// Add each category
	for _, category := range []string{"Dict Validation", "Dict References", "Other"} {
		if errors, exists := categories[category]; exists && len(errors) > 0 {
			sections = append(sections, f.formatErrorCategory(category, errors))
			sections = append(sections, "")
		}
	}

	return strings.Join(sections, "\n")
}

// categorizeErrors groups errors by type
func (f *ErrorFormatter) categorizeErrors(errors []error) map[string][]error {
	categories := make(map[string][]error)

	for _, err := range errors {
		category := f.getErrorCategory(err)
		categories[category] = append(categories[category], err)
	}

	return categories
}

// getErrorCategory determines the category of an error
func (f *ErrorFormatter) getErrorCategory(err error) string {
	switch e := err.(type) {
	case *DictValidationError:
		return "Dict Validation"
	case *ParseError:
		if strings.Contains(e.Message, "$dict reference") {
			return "Dict References"
		}
		if strings.Contains(e.PropertyPath, "dict.") {
			return "Dict Validation"
		}
		return "Other"
	case *ErrorCollection:
		// For nested error collections, categorize based on the first error
		if len(e.Errors) > 0 {
			return f.getErrorCategory(e.Errors[0])
		}
		return "Other"
	default:
		return "Other"
	}
}

// formatErrorCategory formats a category of errors
func (f *ErrorFormatter) formatErrorCategory(category string, errors []error) string {
	var lines []string

	// Category header
	icon := f.getCategoryIcon(category)
	lines = append(lines, fmt.Sprintf("%s %s (%d)", icon, category, len(errors)))

	// Sort errors by file path and line number for consistent output
	sortedErrors := f.sortErrors(errors)

	// Format each error in the category
	for i, err := range sortedErrors {
		errorText := f.FormatError(err)
		// Indent the error text
		indentedText := f.indentText(errorText, "   ")
		lines = append(lines, fmt.Sprintf("  %d. %s", i+1, indentedText))
	}

	return strings.Join(lines, "\n")
}

// sortErrors sorts errors by file path and line number
func (f *ErrorFormatter) sortErrors(errors []error) []error {
	sorted := make([]error, len(errors))
	copy(sorted, errors)

	sort.Slice(sorted, func(i, j int) bool {
		errI := f.getParseError(sorted[i])
		errJ := f.getParseError(sorted[j])

		if errI == nil && errJ == nil {
			return false
		}
		if errI == nil {
			return false
		}
		if errJ == nil {
			return true
		}

		// Sort by file path first
		if errI.FilePath != errJ.FilePath {
			return errI.FilePath < errJ.FilePath
		}

		// Then by line number
		return errI.LineNumber < errJ.LineNumber
	})

	return sorted
}

// getParseError extracts ParseError from various error types
func (f *ErrorFormatter) getParseError(err error) *ParseError {
	switch e := err.(type) {
	case *ParseError:
		return e
	case *DictValidationError:
		return e.ParseError
	default:
		return nil
	}
}

// indentText indents each line of text
func (f *ErrorFormatter) indentText(text string, indent string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if i == 0 {
			continue // Don't indent the first line
		}
		lines[i] = indent + line
	}
	return strings.Join(lines, "\n")
}

// getLevelIcon returns an icon for the error level
func (f *ErrorFormatter) getLevelIcon(level ErrorLevel) string {
	switch level {
	case ErrorLevelError:
		return "❌"
	case ErrorLevelWarning:
		return "⚠️"
	case ErrorLevelInfo:
		return "ℹ️"
	default:
		return "❓"
	}
}

// getCategoryIcon returns an icon for the error category
func (f *ErrorFormatter) getCategoryIcon(category string) string {
	switch category {
	case "Dict Validation":
		return "🔧"
	case "Dict References":
		return "🔗"
	case "Other":
		return "📋"
	default:
		return "📄"
	}
}

// FormatErrorSummary provides a brief summary of errors
func (f *ErrorFormatter) FormatErrorSummary(err error) string {
	switch e := err.(type) {
	case *ErrorCollection:
		if len(e.Errors) == 0 {
			return "No errors"
		}

		if len(e.Errors) == 1 {
			return f.FormatError(e.Errors[0])
		}

		categories := f.categorizeErrors(e.Errors)
		var parts []string

		for category, errors := range categories {
			if len(errors) > 0 {
				parts = append(parts, fmt.Sprintf("%s: %d", category, len(errors)))
			}
		}

		return fmt.Sprintf("Validation failed with %d errors (%s)", len(e.Errors), strings.Join(parts, ", "))
	default:
		return err.Error()
	}
}
