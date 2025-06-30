# s2http-request

A **Simple and Structured HTTP Request** dispatching tool.

## Table of Contents

- [Overview](#overview)
- [Features]
- [Request Definition Format]
- [Built-in Functions]
- [Dictionary Feature for Fuzzing]
- [Payload Management using Dictionary Files]
- [Usage]
- [Output Format]
- [Directory Structure]
- [License]
- [Contributing]

## Overview

s2http-request is a versatile and lightweight HTTP request dispatching tool designed for various testing and automation scenarios. It allows users to define complex HTTP requests using pure JSON/YAML, eliminating the need for domain-specific languages (DSLs) or templating engines. This approach ensures clarity, simplicity, and ease of integration with existing workflows.

## Features

- 🚀 **Lightweight & Portable:** Distributed as a single, self-contained binary.
- 📝 **Declarative Configuration:** Define requests using intuitive JSON/YAML files.
- 🔧 **Dynamic Value Generation:** Leverage powerful built-in functions for on-the-fly data manipulation.
- 📊 **Comprehensive Output:** Get detailed response information for thorough analysis.
- ⚡ **High Performance:** Built with Go for speed and efficiency.
- 📋 **JSON Schema Support:** Benefit from IDE auto-completion and validation for request definitions.

## Request Definition Format

### Basic Request Definition

```json
{
  "method": "GET",
  "path": "/",
  "query": {
    "q": "hello",
    "r": "",
    "s": null,
    "t": {
      "!var": "malicious_key"
    }
  },
  "headers": {
    "Content-Type": "application/www-form-urlencoded"
  },
  "params": {
    "id": "1",
    "name": {
      "!url_encode": "John Doe"
    }
  },
  "variables": {
    "malicious_key": {
      "!concat": [
        {"!var": "v"},
        {"!var": "k"}
      ]
    },
    "v": "1",
    "k": "2"
  }
}
```

### Array-based Parameter Definition

```json
{
  "method": "GET",
  "path": "/",
  "query": [
    {
      "key": "q",
      "value": "v"
    },
    {
      "key": "q",
      "value": "v2"
    },
    {
      "key": {
        "!join": [
          ",",
          {"!var": "v"},
          {"!var": "k"}
        ]
      },
      "value": "v"
    }
  ],
  "headers": [
    {
      "key": "Content-Type",
      "value": "multipart/form-data"
    }
  ],
  "params": [
    {
      "key": "p",
      "value": "v"
    },
    {
      "key": "p",
      "value": "v2"
    },
    {
      "key": {
        "!join": [
          ",",
          {"!var": "v"},
          {"!var": "k"}
        ]
      },
      "value": "v"
    }
  ],
  "variables": {
    "v": {
      "!random": [10]
    },
    "k": "2"
  }
}
```

## Built-in Functions

The tool provides a set of built-in functions for dynamic value generation.

### Variable Operations
- `!var`: Reference a variable
- `!concat`: Concatenate strings
- `!join`: Join array elements

### Encoding/Decoding
- `!url_encode`: URL encode a string
- `!url_decode`: URL decode a string
- `!base64_encode`: Base64 encode a string
- `!base64_decode`: Base64 decode a string
- `!html_encode`: HTML encode a string
- `!html_decode`: HTML decode a string

### Random Generation
- `!random`: Generate a random number (0 to N-1)
- `!random_string`: Generate a random string
- `!uuid`: Generate a UUID

### Time Functions
- `!timestamp`: Current Unix timestamp
- `!date`: Current date (YYYY-MM-DD)
- `!time`: Current time (HH:MM:SS)

### Dictionary Operations
- `!dict_load`: Load dictionary data from an external file
- `!dict_random`: Select a random value from a dictionary
- `!dict_get`: Get a value from a dictionary at a specified index

### Array Operations
- `!concat_arrays`: Concatenate multiple arrays

### Dictionary Feature for Fuzzing

This feature allows you to define a list of values (payloads) within your request, and the dispatcher will send a separate request for each value in the list. This is particularly useful for fuzzing or iterating through different inputs.

```json
{
  "method": "POST",
  "path": "/test",
  "params": {
    "input": { "!var": "payload" }
  },
  "dictionary": {
    "payload": [
      "<script>alert(1)</script>",
      "' OR '1'='1",
      "; cat /etc/passwd"
    ]
  }
}
```

In this example, a request will be sent for each element (3 elements) of the `payload`. You can refer to each element using `!var: payload`.

---

## Payload Management using Dictionary Files

You can load and use external dictionary files containing lists of values (e.g., common strings, test data, or attack payloads). This allows for efficient management and reuse of data across multiple requests.

### Supported File Formats

- **JSON**: Define payloads as an array
- **YAML**: Define payloads as an array or map
- **Text**: One payload per line (empty lines are skipped)

### Example Dictionary Files

```json
// examples/dictionaries/xss_payloads.json
[
  "<script>alert('XSS')</script>",
  "<img src=x onerror=alert('XSS')>",
  "<svg onload=alert('XSS')>",
  "javascript:alert('XSS')"
]
```

```yaml
# examples/dictionaries/injection_payloads.yaml
sql_injection:
  - "' OR '1'='1"
  - "' OR '1'='1' --"
  - "admin' --"
command_injection:
  - "; ls"
  - "| whoami"
  - "&& id"
```

```text
# examples/dictionaries/common_payloads.txt
# XSS Payloads
<script>alert('XSS')</script>
<img src=x onerror=alert('XSS')>

# SQL Injection Payloads
' OR '1'='1
admin' --

# Hash-based payloads (these start with # but are not comments)
#hashtag_injection
#social_media_payload
```

### Example Usage of Dictionary Functions

```json
{
  "method": "POST",
  "path": "/search",
  "params": {
    "query": {
      "!dict_random": {
        "!dict_load": "examples/dictionaries/xss_payloads.json"
      }
    },
    "category": {
      "!dict_get": [
        {
          "!dict_load": "examples/dictionaries/injection_payloads.yaml"
        },
        2
      ]
    }
  },
  "variables": {
    "all_payloads": {
      "!concat_arrays": [
        {
          "!dict_load": "examples/dictionaries/xss_payloads.json"
        },
        {
          "!dict_load": "examples/dictionaries/injection_payloads.yaml"
        }
      ]
    }
  }
}
```

## Usage

### Installation

```bash
# Install the latest version
go install github.com/secureta/s2http-request/cmd/s2req@latest
go install github.com/secureta/s2http-request/cmd/s2req-schema@latest

# Or, clone the repository and install from the local directory
git clone https://github.com/secureta/s2http-request.git
cd s2http-request
go install ./cmd/s2req
go install ./cmd/s2req-schema
```

The `install` command builds the `s2req` binary and installs it into your `$GOPATH/bin` or `$GOBIN` directory.

Alternatively, you can download a pre-built binary from the [Releases](https://github.com/your-username/s2http-request/releases) page.

```bash
# Example for Linux
curl -L https://github.com/your-username/s2http-request/releases/latest/download/s2req-linux-amd64 -o s2req
chmod +x s2req
# You can move it to a directory in your PATH, e.g., /usr/local/bin
```

### JSON Schema Generation

```bash
# Generate JSON Schema
go tool mage schema

# IDE auto-completion settings (VS Code example):
# Add to .vscode/settings.json:
{
  "json.schemas": [
    {
      "fileMatch": ["**/requests/*.json", "**/examples/*.json"],
      "url": "./request-schema.json"
    }
  ]
}
```

### Basic Usage

```bash
# Execute request from a JSON file
s2req request.json

# Execute request from a YAML file
s2req request.yaml

# Execute multiple request files
s2req requests/*.json

# Verbose output mode
s2req --verbose request.json

# Save results to a file
s2req --output results.json request.json
```

### Configuration Options

```bash
# Specify target host
s2req --host https://example.com request.json

# Set timeout
s2req --timeout 30 request.json

# Set retry count
s2req --retry 3 request.json

# Use proxy
s2req --proxy http://proxy.example.com:8080 request.json
```

## Output Format

```json
{
  "request": {
    "method": "GET",
    "url": "https://example.com/?q=hello&t=12",
    "headers": {
      "Content-Type": "application/www-form-urlencoded"
    },
    "body": "id=1&name=John%20Doe"
  },
  "response": {
    "status_code": 200,
    "headers": {
      "Content-Type": "text/html",
      "Content-Length": "1234"
    },
    "body": "...",
    "time": {
      "total": 0.123,
      "dns": 0.001,
      "connect": 0.010,
      "ssl": 0.050,
      "send": 0.001,
      "wait": 0.060,
      "receive": 0.001
    }
  },
  "metadata": {
    "timestamp": "2025-06-09T04:44:46.758Z",
    "variables": {
      "v": "1",
      "k": "2",
      "malicious_key": "12"
    }
  }
}
```

## Directory Structure

```
s2http-request/
├── examples/
│   └── dictionaries/
├── internal/
│   ├── config/
│   ├── http/
│   └── parser/
└── pkg/
    └── functions/
```

## License

MIT License

## Contributing

Pull requests and issue reports are welcome.