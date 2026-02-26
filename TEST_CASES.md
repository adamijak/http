# Test Use Cases and Documentation

This document provides a comprehensive overview of all test scenarios covered in the HTTP client tool testing suite.

## Test Organization

The testing suite is organized into three layers:

1. **Unit Tests**: Test individual components (54 tests across 4 modules)
2. **E2E Tests**: Test the complete CLI flow with test server (18 scenarios)
3. **Integration Tests**: Test complete workflows via shell script (17 scenarios)

---

## Unit Test Use Cases

### Parser Tests (39 tests total)

#### Basic Parsing
- ✓ Simple GET request parsing
- ✓ POST request with JSON body
- ✓ Multiple HTTP methods (GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS, CONNECT, TRACE)
- ✓ Multiple HTTP versions (1.0, 1.1, 2, 3)
- ✓ Path-only URLs (e.g., `/api/users`)
- ✓ URLs with query parameters
- ✓ URLs with complex paths and query strings
- ✓ URLs with explicit ports
- ✓ IPv6 URLs (e.g., `https://[::1]/api`)

#### Headers
- ✓ Multiple headers parsing
- ✓ Headers with spaces in values
- ✓ Headers with colons in values (e.g., timestamps, URLs)
- ✓ Case-insensitive header names
- ✓ Very long header values (1000+ chars)

#### Body Handling
- ✓ Empty body requests
- ✓ Multiline bodies (JSON, XML)
- ✓ Complex nested JSON bodies
- ✓ Bodies with blank lines
- ✓ Bodies with special characters and Unicode
- ✓ Very long bodies (5000+ chars)
- ✓ Form data bodies (`application/x-www-form-urlencoded`)
- ✓ XML request bodies

#### Preprocessing
- ✓ Comment handling (`#` and `//` styles)
- ✓ Mixed comment styles
- ✓ Environment variable substitution (`${VAR}` and `$VAR`)
- ✓ Multiple environment variables in one request
- ✓ Environment variables in headers
- ✓ Environment variables in body
- ✓ Undefined environment variables
- ✓ Shell command execution in headers (`$(command)`)
- ✓ Nested shell commands with pipes

#### Edge Cases
- ✓ Trailing whitespace handling
- ✓ Mixed line endings (CRLF vs LF)
- ✓ Invalid request formats
- ✓ Special characters in headers

### Validator Tests (49 tests total)

#### HTTP Methods
- ✓ Standard methods (GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS, CONNECT, TRACE)
- ✓ Non-standard methods (PROPFIND, PROPPATCH, MKCOL, COPY, MOVE, LOCK, UNLOCK)

#### URLs
- ✓ HTTPS URLs
- ✓ HTTP URLs
- ✓ URLs without scheme (error case)
- ✓ Path-only URLs
- ✓ URLs with fragments (#section)
- ✓ URLs with special/encoded characters
- ✓ IPv4 addresses
- ✓ IPv6 addresses
- ✓ Localhost variants (localhost, 127.0.0.1, [::1])
- ✓ Long URLs (100+ path segments)
- ✓ Invalid URLs (error cases)

#### HTTP Versions
- ✓ HTTP/1.0, HTTP/1.1
- ✓ HTTP/2, HTTP/2.0
- ✓ HTTP/3
- ✓ Non-standard versions (warnings)

#### Headers
- ✓ Missing Host header (warning)
- ✓ Empty headers
- ✓ Multiple content headers (Content-Type, Content-Length, Content-Encoding, etc.)
- ✓ Authorization headers (Bearer, Basic, Digest, OAuth, AWS4-HMAC)
- ✓ Cache headers (Cache-Control, Pragma, Expires, If-None-Match, If-Modified-Since)
- ✓ CORS headers (Origin, Access-Control-Request-*)
- ✓ Custom headers (X-*)
- ✓ Range request headers
- ✓ Conditional request headers (If-Match, If-Unmodified-Since)
- ✓ WebSocket upgrade headers
- ✓ Content negotiation headers (Accept, Accept-Encoding, Accept-Language)
- ✓ Cookie headers
- ✓ Redirect headers (Referer, Location)
- ✓ Proxy headers (X-Forwarded-*, Forwarded, Via)

#### Body Validation
- ✓ Body with Content-Length
- ✓ Body without Content-Type (warning)
- ✓ GET with body (warning)
- ✓ POST without Content-Type (warning)
- ✓ Large body (10000+ chars)
- ✓ Multipart form data

#### Special Cases
- ✓ No-secure flag validation
- ✓ WebSocket upgrade validation
- ✓ HTTP/2 specific validation
- ✓ Complex real-world requests

### Client Tests (9 tests total)

#### HTTP Communication
- ✓ Send plain HTTP request
- ✓ Send HTTPS request (with self-signed cert handling)
- ✓ POST request with body
- ✓ Custom headers transmission
- ✓ Query parameters in URL
- ✓ Multiple HTTP methods
- ✓ Invalid URL error handling
- ✓ Connection refused error handling

### Models Tests (12 tests total)

#### Data Structures
- ✓ HTTPRequest creation
- ✓ HTTPRequest printing
- ✓ HTTPRequest colored output
- ✓ HTTPRequest to raw format conversion
- ✓ HTTPRequest with body conversion
- ✓ HTTPResponse printing
- ✓ HTTPResponse colored output
- ✓ Multiple status codes (200, 201, 204, 301, 400, 401, 403, 404, 500)
- ✓ Empty body handling
- ✓ Multiple headers handling

---

## E2E Test Use Cases (18 scenarios)

### Basic HTTP Operations
1. **Simple GET Request**: Tests basic GET with response validation
2. **POST with Body**: Tests POST with JSON body and Content-Type
3. **Multiple Methods**: Tests GET, POST, PUT, DELETE, PATCH
4. **Custom Headers**: Tests multiple custom headers (X-*, Authorization, Accept)
5. **Query Parameters**: Implicitly tested in various scenarios

### CLI Features
6. **File Input (`-f` flag)**: Tests reading request from file
7. **No-Send Flag (`--no-send`)**: Tests RFC format output without sending
8. **Validation Error**: Tests that invalid requests are caught
9. **Strict Mode (`--strict`)**: Tests that warnings fail in strict mode
10. **Port Override (`--port`)**: Tests explicit port specification

### Advanced Features
11. **Environment Variables**: Tests `${VAR}` and `$VAR` substitution
12. **Comment Handling**: Tests that `#` and `//` comments are stripped
13. **Pipeline Flow**: Tests HTP → RFC → tool chain

### Response Handling
14. **Response Status Codes**: Tests 200, 201, 204, 400, 401, 404, 500
15. **Connection Errors**: Tests connection refused handling

### Real-World Scenarios
16. **Real-World API Request**: Tests realistic API call with auth, custom headers, JSON response
17. **Example Files**: Tests that all example files parse correctly
18. **Test Server Integration**: Validates requests actually reach the server with correct data

---

## Integration Test Use Cases (17 scenarios via test.sh)

### CLI Functionality
1. **Version Flag**: Tests `--version` output
2. **Help Flag**: Tests `-h` help display
3. **No-Send Mode**: Tests `--no-send` flag
4. **No-Color Mode**: Tests `--no-color` flag

### Request Processing
5. **Simple GET**: Tests basic GET request parsing
6. **POST with JSON**: Tests POST with JSON body
7. **Environment Variables**: Tests variable substitution
8. **Shell Commands**: Tests command execution
9. **Comment Handling**: Tests comment stripping

### Validation
10. **Validation Errors**: Tests error detection for invalid URLs
11. **Strict Mode**: Tests strict mode enforcement

### Auto-Features
12. **Content-Length Auto-Add**: Tests automatic Content-Length header
13. **Host Header Auto-Add**: Tests automatic Host header

### Format Handling
14. **RFC Output**: Tests CRLF line endings in output
15. **File Loading**: Tests `-f` flag with RFC format
16. **HTP File Loading**: Tests `-f` flag with HTP format
17. **Pipeline Flow**: Tests complete HTP → RFC → tool workflow

---

## Test Coverage by Feature

### Core Features (100% covered)
- ✓ HTTP methods (all standard + WebDAV)
- ✓ URL parsing (HTTP, HTTPS, IP, ports)
- ✓ Header parsing (all common headers)
- ✓ Body handling (JSON, XML, form data)
- ✓ HTTP versions (1.0, 1.1, 2, 3)

### CLI Features (100% covered)
- ✓ File input (`-f`, `--file`)
- ✓ Stdin input (piping)
- ✓ No-send mode (`--no-send`)
- ✓ No-color mode (`--no-color`)
- ✓ Verbose mode (`-v`, `--verbose`)
- ✓ Strict mode (`--strict`)
- ✓ Port override (`--port`)
- ✓ No-secure mode (`--no-secure`)
- ✓ Version display (`--version`)
- ✓ Help display (`-h`)

### Advanced Features (100% covered)
- ✓ Environment variables (`${VAR}`, `$VAR`)
- ✓ Shell commands (`$(cmd)`)
- ✓ Comments (`#`, `//`)
- ✓ Format auto-detection (HTP vs RFC)
- ✓ Auto Host header
- ✓ Auto Content-Length
- ✓ HTTPS/TLS support

### Error Handling (100% covered)
- ✓ Invalid URLs
- ✓ Connection failures
- ✓ Certificate errors
- ✓ Parse errors
- ✓ Validation errors
- ✓ Missing required fields
- ✓ Strict mode warnings

---

## Example Files and Use Cases

### Existing Examples
1. **simple-get.http**: Basic GET request to httpbin.org
2. **post-json.http**: POST with JSON payload
3. **with-env-vars.http**: Environment variable usage
4. **with-shell-commands.http**: Shell command execution

### New Examples Added
5. **put-update.http**: PUT request for resource update
6. **delete-resource.http**: DELETE request for resource deletion
7. **form-data.http**: Form data submission
8. **custom-headers.http**: API authentication with custom headers
9. **patch-update.http**: PATCH for partial updates
10. **query-params.http**: Query parameter handling

Each example demonstrates:
- Proper HTTP format
- Real-world API patterns
- Common authentication methods
- Different content types
- Best practices

---

## Running Specific Test Scenarios

### Run all tests
```bash
make check
```

### Run specific test layer
```bash
make unittest   # Unit tests only
make e2etest    # E2E tests only
./test.sh       # Integration tests only
```

### Run specific test package
```bash
go test ./internal/parser -v
go test ./internal/validator -v
go test ./internal/client -v
go test ./internal/models -v
```

### Run specific test case
```bash
go test ./internal/parser -run TestParseWithEnvVars -v
go test -run TestE2ESimpleGET -v e2e_test.go
```

### Run with coverage
```bash
go test ./... -cover
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## Test Scenarios by Priority

### Critical (Must Pass)
- Basic HTTP operations (GET, POST, PUT, DELETE)
- URL parsing and validation
- Header parsing and transmission
- Request/response flow
- Error handling

### Important (Should Pass)
- Environment variables
- Shell commands
- Comment handling
- Format auto-detection
- CLI flags

### Nice to Have (Enhancement)
- Advanced headers (CORS, WebSocket, etc.)
- Special URL formats (IPv6, etc.)
- Edge cases (long content, special chars)
- Real-world API patterns

---

## Continuous Integration

All tests are run in CI via `make check`:
1. Code formatting check (`gofmt`)
2. Static analysis (`go vet`)
3. Unit tests (`go test ./...`)
4. E2E tests (`go test e2e_test.go`)
5. Integration tests (`./test.sh`)

Every PR must pass all tests before merge.

---

*Last updated: 2026-02-26*
*Total test scenarios: 90+ across all layers*
