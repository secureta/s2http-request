# LLM Instructions

## Core Principles

- Regardless of the language of the instruction, your thinking process must be in English. The final response should be in the language of the original instruction.
- When implementing new features or fixing bugs, you must follow the Test-Driven Development (TDD) methodology as advocated by Takuto Wada. This means writing a failing test before writing the implementation code.

## TDD Methodology (based on Takuto Wada's teachings)

Takuto Wada emphasizes a pragmatic and sustainable approach to TDD. The core of his philosophy is the "Red-Green-Refactor" cycle, but with a strong focus on the "why" behind each step.

### The Red-Green-Refactor Cycle

1.  **Red - Write a small test that fails.**
    *   Start by writing a test for a single, small piece of functionality that does not yet exist.
    *   The test should fail because the implementation code hasn't been written yet. This confirms that the test is working correctly and that the feature is indeed missing.
    *   **Key principle:** The test should clearly state the expected behavior of the system.

2.  **Green - Write the simplest possible code to make the test pass.**
    *   Write the absolute minimum amount of code required to make the test pass.
    *   It's acceptable for this initial code to be "ugly" or inefficient. The goal is to get to a passing state as quickly as possible.
    *   **Key principle:** This step ensures that you have a working system, even if the implementation is not perfect.

3.  **Refactor - Improve the code's design without changing its behavior.**
    *   Now that you have a passing test, you can refactor the implementation code to improve its clarity, efficiency, and overall design.
    *   The existing tests act as a safety net, ensuring that your refactoring does not break any existing functionality.
    *   **Key principle:** This is where you focus on creating clean, maintainable code.

### Practical Tips from Takuto Wada's Approach

*   **Test Naming:** Tests should be named descriptively, clearly stating what they are testing. A good practice is to use a format like `Test<FunctionName>_<Condition>_<ExpectedBehavior>`.
*   **Small Steps:** Take small, incremental steps. Each test should cover a tiny piece of functionality. This makes the process more manageable and reduces the risk of introducing bugs.
*   **Listen to Your Tests:** If a test is difficult to write, it might be a sign that your design is too complex. Use this as an opportunity to rethink your approach.
*   **Focus on Behavior, Not Implementation:** Tests should focus on the "what" (the behavior of the system) rather than the "how" (the implementation details). This makes your tests more resilient to refactoring.

By following these principles, you can ensure that the codebase remains clean, maintainable, and well-tested, which is a key goal of this project.

## Project Directory Structure

- **`/` (root)**: Contains project configuration files (`.gitignore`, `go.mod`, `go.sum`), build scripts (`magefile.go`), documentation (`README.md`), and schemas (`request-schema.json`).
- **`examples/`**: Contains various example configuration files (JSON, YAML) demonstrating how to use the project.
    - `dictionaries/`: Stores dictionary files used for payloads, fuzzing, etc. (e.g., `common_payloads.txt`, `injection_payloads.yaml`).
- **`internal/`**: Holds internal Go packages that are not intended for external use.
    - `config/`: Configuration-related code.
    - `http/`: HTTP client implementation.
    - `parser/`: Parsing logic.
- **`pkg/`**: Contains Go packages intended to be used by external applications.
    - `functions/`: A collection of utility functions for arrays, dictionaries, encoding, fuzzing, random data generation, time manipulation, variables, and WAF (Web Application Firewall) testing.

## Project Overview

The goal of this project is to create a simple tool for sending HTTP requests to test the effectiveness of a Web Application Firewall (WAF).

A key requirement is that request definitions should be written in pure JSON or YAML, without using any Domain-Specific Languages (DSLs) or template engines. This allows for clear and straightforward request structures.

The tool supports dynamic value generation within the request definitions using special `!function` tags. This enables features like variable substitution, string manipulation, and random data generation.

### Example Request Definitions:

Here are examples of how requests can be defined.

**Example 1: Using variables and functions in a map structure**

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

**Example 2: Using functions and lists for query parameters and headers**

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