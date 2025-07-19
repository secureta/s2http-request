# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Core Commands

All project tasks must be executed using `go tool mage`:

- `go tool mage build` - Build s2req and s2req-schema binaries
- `go tool mage test` - Run all tests  
- `go tool mage testCoverage` - Run tests with HTML coverage report
- `go tool mage lint` - Run golangci-lint
- `go tool mage fmt` - Format all Go code
- `go tool mage security` - Run gosec security scanner
- `go tool mage schema` - Generate request-schema.json file
- `go tool mage install` - Install binaries to GOPATH/bin
- `go tool mage clean` - Remove build artifacts

To run a single test: `go test -v -run TestFunctionName ./path/to/package`

## Architecture Overview

**s2http-request** is a WAF testing tool that sends HTTP requests defined in pure JSON/YAML files with dynamic value generation via `!function` tags.

### Key Components

1. **cmd/s2req** - Main HTTP request dispatcher
2. **cmd/s2req-schema** - JSON schema generator for request validation
3. **internal/parser** - Parses and processes request definitions with function evaluation
4. **internal/http** - HTTP client that builds and sends requests
5. **pkg/functions/registry.go** - Central registry for all dynamic functions

### Function System

Dynamic functions use `!function_name` tags in JSON/YAML:
- Variable: `!var`, `!concat`, `!join`
- Encoding: `!url_encode`, `!base64_encode`, `!html_encode`, `!hex_encode`
- Decoding: `!url_decode`, `!base64_decode`, `!html_decode`
- Random: `!random`, `!random_string`, `!uuid`
- Time: `!timestamp`, `!date`, `!time`
- Dictionary: `!dict_load`, `!dict_random`, `!dict_get`
- Arrays: `!concat_arrays`

New functions must:
1. Be implemented in `pkg/functions/`
2. Be registered in `pkg/functions/registry.go`
3. Have comprehensive tests following TDD

## Development Requirements

### Test-Driven Development (TDD)

**MANDATORY**: Follow Takuto Wada's TDD methodology for all changes:

1. **Red** - Write a failing test first
2. **Green** - Write minimal code to pass
3. **Refactor** - Improve design with tests as safety net

Test naming: `Test<FunctionName>_<Condition>_<ExpectedBehavior>`

### Critical Rules

1. **No DSLs or templating** - Keep request definitions pure JSON/YAML
2. **Update README.md** when modifying magefile.go commands
3. **All responses in English** - Think in English regardless of instruction language
4. **Function registry** - All new functions must be registered
5. **Use mage** - Never use raw `go` commands for project tasks

### Request Definition Structure

Requests support two formats:

1. **Map-based** (simple key-value):
```json
{
    "method": "GET",
    "path": "/api/test",
    "query": {"key": "value"},
    "headers": {"Content-Type": "application/json"},
    "params": {"id": "123"}
}
```

2. **Array-based** (duplicate keys):
```json
{
    "query": [
        {"key": "q", "value": "v1"},
        {"key": "q", "value": "v2"}
    ]
}
```

Variables and functions can be used anywhere in the request definition.