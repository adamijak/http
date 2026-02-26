# Testing Documentation

This document describes the testing strategy, test organization, and how to run tests for the HTTP client tool.

## Table of Contents

1. [Testing Philosophy](#testing-philosophy)
2. [Test Organization](#test-organization)
3. [Running Tests](#running-tests)
4. [Test Coverage](#test-coverage)
5. [Writing Tests](#writing-tests)
6. [E2E Testing](#e2e-testing)
7. [CI/CD Integration](#cicd-integration)

## Testing Philosophy

The HTTP client tool follows a comprehensive testing approach with multiple layers:

- **Unit Tests**: Test individual components in isolation (parser, validator, client, models)
- **Integration Tests**: Test the CLI tool with the internal test server (E2E tests)
- **Shell Script Tests**: Test complete workflows and edge cases via `test.sh`

Our testing principles:
- Test the behavior, not the implementation
- Use the internal test server for realistic E2E scenarios
- Validate both success and failure paths
- Test edge cases and error handling
- Keep tests fast and deterministic

## Test Organization

```
http/
├── test.sh                          # Main integration test suite (17 tests)
├── e2e_test.go                      # End-to-end CLI tests with test server
├── internal/
│   ├── client/client_test.go        # HTTP client unit tests (9 tests)
│   ├── models/models_test.go        # Data model tests (12 tests)
│   ├── parser/parser_test.go        # Parser unit tests (15 tests)
│   ├── validator/validator_test.go  # Validator unit tests (18 tests)
│   └── testserver/testserver.go     # Test HTTP/HTTPS server
└── examples/                        # Sample .http files for manual testing
    ├── simple-get.http
    ├── post-json.http
    ├── with-env-vars.http
    └── with-shell-commands.http
```

### Test Categories

#### 1. Unit Tests (Go)
Located in `internal/*/` directories, these tests verify individual components:

- **Parser Tests** (`parser_test.go`): Request parsing, preprocessing, format detection
- **Validator Tests** (`validator_test.go`): RFC compliance, validation rules
- **Client Tests** (`client_test.go`): HTTP/HTTPS sending, connection handling
- **Models Tests** (`models_test.go`): Data structures, serialization, output formatting

#### 2. E2E Tests (Go)
Located in `e2e_test.go` at the root, these tests:
- Use the actual CLI binary
- Spin up internal test servers
- Validate complete request/response flows
- Test real-world scenarios

#### 3. Integration Tests (Bash)
Located in `test.sh`, these tests verify:
- CLI flags and options
- File I/O and piping
- Environment variable substitution
- Shell command execution
- Error handling and validation
- Pipeline workflows

## Running Tests

### Quick Start
```bash
# Run all tests (format, lint, unit, integration)
make check

# Run only unit tests
make unittest

# Run only integration tests
./test.sh

# Run E2E tests
go test -v e2e_test.go

# Run specific package tests
go test -v ./internal/parser
go test -v ./internal/client
```

### Individual Test Commands

```bash
# Format check
gofmt -l .

# Linting
go vet ./...

# All unit tests with verbose output
go test ./... -v

# Run specific test
go test ./internal/parser -run TestParseWithEnvVars -v

# Run tests with coverage
go test ./... -cover

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Test Coverage

### Current Coverage by Component

| Component | Unit Tests | E2E Tests | Integration Tests |
|-----------|-----------|-----------|-------------------|
| Parser | 15 tests | ✓ | ✓ (env vars, shell commands) |
| Validator | 18 tests | ✓ | ✓ (validation errors, strict mode) |
| Client | 9 tests | ✓ | - |
| Models | 12 tests | ✓ | - |
| CLI | - | ✓ | 17 scenarios |

### Use Cases Covered

#### Basic Functionality
- ✓ Simple GET requests
- ✓ POST requests with JSON bodies
- ✓ Multiple HTTP methods (GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS, CONNECT, TRACE)
- ✓ Custom headers
- ✓ Query parameters
- ✓ Request body handling

#### Advanced Features
- ✓ Environment variable substitution (`${VAR}` and `$VAR`)
- ✓ Shell command execution (`$(command)`)
- ✓ Comment handling (`#` and `//`)
- ✓ Auto-detection of format (HTP vs RFC)
- ✓ Automatic Host header addition
- ✓ Automatic Content-Length calculation
- ✓ HTTPS/TLS support (with self-signed cert handling)

#### CLI Features
- ✓ File input (`-f` flag)
- ✓ Stdin input (piping)
- ✓ Output RFC format without sending (`--no-send`)
- ✓ Colored output control (`--no-color`)
- ✓ Verbose mode (`-v`, `--verbose`)
- ✓ Strict validation mode (`--strict`)
- ✓ Port override (`--port`)
- ✓ HTTP/HTTPS toggle (`--no-secure`)
- ✓ Version display (`--version`)
- ✓ Help display (`-h`)

#### Error Handling
- ✓ Invalid URL detection
- ✓ Connection refused handling
- ✓ Certificate validation errors
- ✓ Parse errors
- ✓ Validation errors and warnings
- ✓ Missing required headers
- ✓ Strict mode enforcement

#### Workflow Tests
- ✓ Pipeline flow (HTP → RFC → tool)
- ✓ File loading and processing
- ✓ Environment variable preprocessing
- ✓ Shell command execution in requests

## Writing Tests

### Unit Test Example

```go
func TestParseSimpleGET(t *testing.T) {
    input := `GET https://example.com/api HTTP/1.1
Host: example.com
User-Agent: test-client`

    req, err := Parse(input)
    if err != nil {
        t.Fatalf("Failed to parse: %v", err)
    }

    if req.Method != "GET" {
        t.Errorf("Expected method GET, got %s", req.Method)
    }
}
```

### E2E Test Example

```go
func TestE2ESimpleGET(t *testing.T) {
    // Create test server
    ts, err := testserver.New()
    if err != nil {
        t.Fatalf("Failed to create test server: %v", err)
    }
    defer ts.Close()

    // Build the CLI binary
    buildCmd := exec.Command("go", "build", "-o", "http-test")
    if err := buildCmd.Run(); err != nil {
        t.Fatalf("Failed to build: %v", err)
    }
    defer os.Remove("http-test")

    // Create test request
    request := fmt.Sprintf("GET %s/test HTTP/1.1\nHost: %s\n",
        ts.URL, strings.TrimPrefix(ts.URL, "http://"))

    // Run CLI
    cmd := exec.Command("./http-test")
    cmd.Stdin = strings.NewReader(request)
    output, err := cmd.CombinedOutput()

    // Validate results
    if err != nil {
        t.Fatalf("CLI failed: %v\nOutput: %s", err, output)
    }
}
```

### Integration Test Example (Bash)

```bash
# Test environment variable substitution
echo "Test: Environment variable substitution"
export TEST_TOKEN="token123"
echo "GET https://example.com HTTP/1.1
Host: example.com
Authorization: Bearer \${TEST_TOKEN}" | ./http --no-send | grep -q "token123"
echo "✓ Environment variables substituted"
```

## E2E Testing

### Test Server

The tool includes an internal test server (`internal/testserver/testserver.go`) that provides:

- **HTTP Server**: Plain HTTP with configurable responses
- **HTTPS Server**: TLS with self-signed certificates
- **Request Capture**: Records all received requests for validation
- **Configurable Responses**: Custom status codes, headers, bodies, delays

### E2E Test Scenarios

The E2E tests validate the complete flow from CLI input to HTTP request/response:

1. **Simple Requests**: Basic GET/POST/PUT/DELETE operations
2. **Headers**: Custom headers, authorization, content-type
3. **Request Bodies**: JSON, form data, text
4. **Response Handling**: Status codes, headers, body parsing
5. **Error Cases**: Invalid requests, connection failures, timeouts
6. **Format Handling**: HTP format and RFC format inputs
7. **Preprocessing**: Environment variables and shell commands
8. **CLI Options**: All command-line flags and combinations

### Running E2E Tests

```bash
# Run all E2E tests
go test -v e2e_test.go

# Run specific E2E test
go test -v e2e_test.go -run TestE2ESimpleGET

# Run with verbose server logs
go test -v e2e_test.go -test.v
```

## CI/CD Integration

### Pre-commit Checks

Before committing code, always run:
```bash
make check
```

This runs:
1. Code formatting check (`gofmt`)
2. Linting (`go vet`)
3. Unit tests (`go test ./...`)
4. Integration tests (`./test.sh`)

### CI Pipeline

The GitHub Actions workflow should run:
1. **Lint**: Format and static analysis
2. **Unit Tests**: All Go unit tests
3. **Integration Tests**: Shell script test suite
4. **E2E Tests**: CLI tests with test server
5. **Build**: Multi-platform binary builds
6. **Coverage**: Code coverage reporting

### Test Requirements for PRs

All pull requests must:
- ✓ Pass all unit tests
- ✓ Pass all integration tests
- ✓ Pass all E2E tests
- ✓ Maintain or improve code coverage
- ✓ Pass formatting and linting checks
- ✓ Include tests for new features
- ✓ Update documentation for behavior changes

## Test Maintenance

### Adding New Tests

When adding new features:
1. Write unit tests for the component
2. Add E2E test for the complete flow
3. Add integration test in `test.sh` if needed
4. Update this documentation
5. Add example `.http` file if applicable

### Updating Existing Tests

When fixing bugs:
1. Add test case that reproduces the bug
2. Fix the bug
3. Verify test passes
4. Ensure no regressions in other tests

### Test Best Practices

- Keep tests focused and independent
- Use descriptive test names
- Test both success and failure paths
- Use table-driven tests for multiple cases
- Clean up resources (defer server.Close())
- Use test fixtures for complex inputs
- Validate both behavior and error messages
- Keep tests fast (< 1 second per test)

## Troubleshooting Tests

### Common Issues

**Tests fail with "connection refused"**
- The test server may not have started
- Check port availability
- Increase server startup delay in tests

**TLS certificate errors**
- Expected for self-signed certs in tests
- Tests should handle this gracefully
- Use `--no-secure` flag for test servers

**Environment variable tests fail**
- Ensure proper setup and cleanup
- Use `defer os.Unsetenv()` in tests
- Check for variable name conflicts

**Flaky integration tests**
- Add delays for async operations
- Use proper synchronization
- Check for race conditions

### Debugging Tests

```bash
# Run with verbose output
go test -v ./internal/parser

# Run specific test
go test -run TestParseWithEnvVars ./internal/parser -v

# Enable race detector
go test -race ./...

# Show all output (including t.Log)
go test -v ./... | grep -A 10 "TestName"

# Debug with print statements
go test -v ./internal/client -run TestSendHTTP
```

## Performance Testing

While not automated, consider these manual performance checks:

```bash
# Time a request
time cat examples/simple-get.http | ./http --no-send

# Memory usage
/usr/bin/time -v ./http < examples/post-json.http

# Load testing (with external tool)
# Use wrk, ab, or hey against test server
```

## Future Testing Enhancements

Planned improvements:
- [ ] Increase unit test coverage to 90%+
- [ ] Add benchmark tests for parser performance
- [ ] Add mutation testing
- [ ] Integration with external test APIs
- [ ] Performance regression tests
- [ ] Security vulnerability scanning
- [ ] Fuzz testing for parser
- [ ] Property-based testing for validator
- [ ] Docker-based E2E tests
- [ ] Automated coverage reporting in CI

## Resources

- [Go Testing Documentation](https://pkg.go.dev/testing)
- [Table Driven Tests in Go](https://github.com/golang/go/wiki/TableDrivenTests)
- [HTTP Testing in Go](https://golang.org/pkg/net/http/httptest/)
- [RFC 7230 - HTTP/1.1 Message Syntax](https://tools.ietf.org/html/rfc7230)

---

*Last updated: 2026-02-26*
*Version: 1.0.0*
