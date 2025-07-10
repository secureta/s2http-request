# Development Guidelines for s2http-request

This document provides essential information for developers working on the s2http-request project.

## Build/Configuration Instructions

### Prerequisites

- Go 1.24 or later
- [Mage](https://github.com/magefile/mage) build tool

### Setting Up the Project

1. Clone the repository:
   ```bash
   git clone https://github.com/secureta/s2http-request.git
   cd s2http-request
   ```

2. Install dependencies:
   ```bash
   go tool mage deps
   ```

3. Build the project:
   ```bash
   go tool mage build
   ```
   This will create two binaries in the `bin` directory:
   - `s2req`: The main HTTP request dispatcher
   - `s2req-schema`: Tool for generating JSON schema

### Available Build Commands

The project uses Mage as a build tool. Here are the main commands:

- `go tool mage build`: Build all binaries
- `go tool mage clean`: Clean build artifacts
- `go tool mage install`: Install binaries to GOPATH/bin
- `go tool mage schema`: Generate JSON schema file
- `go tool mage devBuild`: Build with race detector for development
- `go tool mage buildAll`: Build for multiple platforms (Linux, macOS, Windows)
- `go tool mage version`: Show version information

For a complete list of available commands:
```bash
go tool mage help
```

## Testing Information

### Running Tests

To run all tests in the project:
```bash
go tool mage test
```

To run tests with coverage report:
```bash
go tool mage testCoverage
```
This will generate a `coverage.html` file with a visual representation of test coverage.

To run tests for a specific package:
```bash
go test -v ./path/to/package
```

### Writing Tests

Tests follow Go's standard testing conventions. Here's an example of a simple test:

```go
package functions

import (
	"testing"
)

func TestStringReverseFunction(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "simple word",
			input:    "hello",
			expected: "olleh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := reverseString(tt.input)
			
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
```

Key testing practices in this project:
- Use table-driven tests for comprehensive test coverage
- Test both success cases and error conditions
- For HTTP client tests, use `httptest` package to create mock servers
- Test edge cases and boundary conditions

## Additional Development Information

### Project Structure

- `cmd/`: Contains the main entry points for the binaries
  - `s2req/`: Main HTTP request dispatcher
  - `s2req-schema/`: JSON schema generator
- `internal/`: Internal packages not meant for external use
  - `config/`: Configuration types and parsing
  - `http/`: HTTP client implementation
  - `parser/`: Request definition parser
- `pkg/`: Public packages that can be imported by other projects
  - `functions/`: Built-in functions for request manipulation
- `examples/`: Example request definitions and dictionaries
- `bin/`: Output directory for built binaries

### Code Style

- Follow standard Go code style and conventions
- Use `go fmt` to format code before committing
- Run `go tool mage lint` to check for common issues

### Adding New Functions

When adding new functions to the `pkg/functions` package:

1. Create a new struct that implements the `Function` interface
2. Register the function in the `registry.go` file
3. Add comprehensive tests for the new function
4. Update documentation if necessary

### JSON Schema

The project uses a JSON schema for request definitions. After making changes to the request structure:

1. Update the schema generator if necessary
2. Regenerate the schema with `go tool mage schema`
3. Test the schema with example request files

### Debugging

- Use the `--verbose` flag with the `s2req` command for detailed output
- For development builds with race detection: `go tool mage devBuild`
- Check the response timing information for performance issues