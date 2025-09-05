package parser

import (
	"fmt"
	"strings"
)

// ErrorLevel represents the severity level of an error
type ErrorLevel int

const (
	ErrorLevelError ErrorLevel = iota
	ErrorLevelWarning
	ErrorLevelInfo
)

// String returns the string representation of ErrorLevel
func (e ErrorLevel) String() string {
	switch e {
	case ErrorLevelError:
		return "ERROR"
	case ErrorLevelWarning:
		return "WARNING"
	case ErrorLevelInfo:
		return "INFO"
	default:
		return "UNKNOWN"
	}
}

// ParseError represents a detailed parsing error with location information
type ParseError struct {
	FilePath     string     `json:"file_path"`
	LineNumber   int        `json:"line_number"`
	ColumnNumber int        `json:"column_number,omitempty"`
	PropertyPath string     `json:"property_path"`
	Message      string     `json:"message"`
	Level        ErrorLevel `json:"level"`
	SourceLine   string     `json:"source_line,omitempty"`
}

// Error implements the error interface
func (e *ParseError) Error() string {
	var parts []string

	// Add file location
	if e.FilePath != "" {
		location := e.FilePath
		if e.LineNumber > 0 {
			location += fmt.Sprintf(":%d", e.LineNumber)
			if e.ColumnNumber > 0 {
				location += fmt.Sprintf(":%d", e.ColumnNumber)
			}
		}
		parts = append(parts, location)
	}

	// Add property path if available
	if e.PropertyPath != "" {
		parts = append(parts, fmt.Sprintf("at %s", e.PropertyPath))
	}

	// Add error level and message
	parts = append(parts, fmt.Sprintf("[%s] %s", e.Level.String(), e.Message))

	return strings.Join(parts, " ")
}

// DictValidationError represents a dict-specific validation error
type DictValidationError struct {
	*ParseError
	DictKey   string      `json:"dict_key"`
	DictValue interface{} `json:"dict_value,omitempty"`
}

// Error implements the error interface for DictValidationError
func (e *DictValidationError) Error() string {
	if e.ParseError != nil {
		return e.ParseError.Error()
	}
	return fmt.Sprintf("dict validation error for key '%s': %s", e.DictKey, e.Message)
}

// ErrorCollection represents a collection of multiple errors
type ErrorCollection struct {
	Errors []error `json:"errors"`
}

// Error implements the error interface for ErrorCollection
func (e *ErrorCollection) Error() string {
	if len(e.Errors) == 0 {
		return "no errors"
	}

	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}

	var messages []string
	messages = append(messages, fmt.Sprintf("Multiple errors (%d):", len(e.Errors)))

	for i, err := range e.Errors {
		messages = append(messages, fmt.Sprintf("  %d. %s", i+1, err.Error()))
	}

	return strings.Join(messages, "\n")
}

// Add adds an error to the collection
func (e *ErrorCollection) Add(err error) {
	if err != nil {
		e.Errors = append(e.Errors, err)
	}
}

// HasErrors returns true if the collection contains any errors
func (e *ErrorCollection) HasErrors() bool {
	return len(e.Errors) > 0
}

// ToError returns the ErrorCollection as an error if it has errors, otherwise nil
func (e *ErrorCollection) ToError() error {
	if e.HasErrors() {
		return e
	}
	return nil
}

// NewParseError creates a new ParseError with the given parameters
func NewParseError(filePath string, lineNumber int, propertyPath string, message string) *ParseError {
	return &ParseError{
		FilePath:     filePath,
		LineNumber:   lineNumber,
		PropertyPath: propertyPath,
		Message:      message,
		Level:        ErrorLevelError,
	}
}

// NewDictValidationError creates a new DictValidationError
func NewDictValidationError(filePath string, lineNumber int, propertyPath string, dictKey string, message string) *DictValidationError {
	return &DictValidationError{
		ParseError: &ParseError{
			FilePath:     filePath,
			LineNumber:   lineNumber,
			PropertyPath: propertyPath,
			Message:      message,
			Level:        ErrorLevelError,
		},
		DictKey: dictKey,
	}
}

// NewErrorCollection creates a new ErrorCollection
func NewErrorCollection() *ErrorCollection {
	return &ErrorCollection{
		Errors: make([]error, 0),
	}
}
