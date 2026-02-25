# HTTP Client Architecture

## Overview

This is a command-line HTTP client tool built in Go that processes `.http` files, validates them against HTTP standards, and sends requests over TCP with colored output.

## Design Goals

1. **AI-Agent Friendly**: Simple, linear code with clear separation of concerns
2. **Standards Compliant**: Validates against HTTP RFC standards
3. **Developer Friendly**: Human-readable .http file format with preprocessing
4. **Zero Dependencies**: Uses only Go standard library
5. **Extensible**: Easy to add new features and validations

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                         main.go                             │
│  (CLI Interface - Flags, stdin/stdout handling)             │
└─────────────────────────────────────────────────────────────┘
                              │
                    ┌─────────┴─────────┐
                    │                   │
                    ▼                   ▼
        ┌───────────────────┐   ┌──────────────────┐
        │   parser.Parse    │   │   Other CLI      │
        │                   │   │   Functions      │
        └───────────────────┘   └──────────────────┘
                    │
                    ▼
        ┌───────────────────────────┐
        │  internal/models          │
        │  - HTTPRequest            │
        │  - HTTPResponse           │
        └───────────────────────────┘
                    │
        ┌───────────┴────────────┐
        │                        │
        ▼                        ▼
┌──────────────────┐    ┌────────────────────┐
│ validator.       │    │ client.Send        │
│ Validate         │    │                    │
│                  │    │                    │
│ - validateMethod │    │ - Parse URL        │
│ - validateURL    │    │ - TCP Connection   │
│ - validateVer    │    │ - TLS (HTTPS)      │
│ - validateHdrs   │    │ - Send Request     │
│ - validateBody   │    │ - Read Response    │
└──────────────────┘    └────────────────────┘
        │                        │
        ▼                        ▼
┌──────────────────┐    ┌────────────────────┐
│ ValidationResult │    │   HTTPResponse     │
└──────────────────┘    └────────────────────┘
        │                        │
        └────────────┬───────────┘
                     │
                     ▼
            ┌─────────────────┐
            │  stdout Output  │
            └─────────────────┘
```

## Component Details

### 1. Main Entry Point (`main.go`)

**Purpose**: CLI interface and coordination

**Responsibilities**:
- Parse command-line flags
- Read from stdin
- Coordinate parser → validator → client flow
- Output results to stdout
- Handle errors

**Key Functions**:
- `main()`: Entry point, orchestrates all components

### 2. Models (`internal/models/`)

**Purpose**: Data structures for HTTP requests and responses

**Files**:
- `request.go`: HTTPRequest structure and methods
- `response.go`: HTTPResponse structure and methods

**Key Structures**:
```go
type HTTPRequest struct {
    Method  string            // GET, POST, etc.
    URL     string            // Full URL with scheme
    Version string            // HTTP/1.1, etc.
    Headers map[string]string // Request headers
    Body    string            // Request body
}

type HTTPResponse struct {
    Version    string            // HTTP version
    StatusCode int               // 200, 404, etc.
    Status     string            // OK, Not Found, etc.
    Headers    map[string]string // Response headers
    Body       string            // Response body
}
```

**Key Methods**:
- `Print()`: Output with/without colors
- `ToRawRequest()`: Convert to raw HTTP format

### 3. Parser (`internal/parser/parser.go`)

**Purpose**: Parse .http files and preprocess them

**Key Functions**:
- `Parse(content)`: Main entry point
- `preprocess(content)`: Handle comments, vars, commands
- `expandEnvVars(line)`: Substitute ${VAR} and $VAR
- `executeShellCommands(line)`: Run $(command)
- `parseHTTP(content)`: Parse HTTP request format

**Processing Pipeline**:
```
Raw .http content
    ↓
Remove comments (# and //)
    ↓
Expand environment variables
    ↓
Execute shell commands
    ↓
Parse HTTP format
    ↓
HTTPRequest struct
```

**Supported Features**:
- Comments: `#` or `//` at start of line
- Env vars: `${VAR_NAME}` or `$VAR_NAME`
- Shell commands: `$(command)`
- Standard HTTP format

### 4. Validator (`internal/validator/validator.go`)

**Purpose**: Validate HTTP requests against standards

**Key Functions**:
- `Validate(req)`: Main entry point
- `validateMethod()`: Check HTTP method
- `validateURL()`: Validate URL format and scheme
- `validateVersion()`: Check HTTP version
- `validateHeaders()`: Validate headers (Host, etc.)
- `validateBody()`: Check body requirements

**Validation Types**:
- **Errors** (must fix): Invalid URL, missing scheme, missing Host
- **Warnings** (should review): Non-standard method, missing Content-Type, etc.

**Auto-fixes**:
- Adds Host header from URL if missing
- Adds Content-Length if body present
- Normalizes method to uppercase

### 5. Client (`internal/client/client.go`)

**Purpose**: Send HTTP requests over TCP

**Key Functions**:
- `Send(req)`: Main entry point
- `buildRawRequest()`: Build raw HTTP string
- `readResponse()`: Parse HTTP response
- `readChunkedBody()`: Handle chunked encoding

**Connection Flow**:
```
Parse URL
    ↓
Determine host:port
    ↓
Open TCP connection
    ↓
(If HTTPS) Wrap with TLS
    ↓
Send raw HTTP request
    ↓
Read response (status, headers, body)
    ↓
Close connection
    ↓
Return HTTPResponse
```

**Supported**:
- HTTP and HTTPS
- Content-Length body reading
- Chunked transfer encoding
- Connection timeout (30s)

## Data Flow

### Complete Request Flow

```
User Input (.http file)
    ↓
stdin → main.go
    ↓
parser.Parse()
    ├→ preprocess (comments, vars, commands)
    └→ parseHTTP (HTTP format)
    ↓
HTTPRequest
    ↓
validator.Validate()
    ├→ validateMethod()
    ├→ validateURL()
    ├→ validateVersion()
    ├→ validateHeaders()
    └→ validateBody()
    ↓
ValidationResult → stdout (errors/warnings)
    ↓
(if --dry-run) HTTPRequest → stdout
    ↓
(else) client.Send()
    ├→ Open TCP connection
    ├→ Send request
    └→ Read response
    ↓
HTTPResponse → stdout
```

## File Organization

```
/
├── main.go              # Entry point
├── version.go           # Version info
├── go.mod              # Go module definition
├── Makefile            # Build automation
├── test.sh             # Test suite
├── demo.sh             # Feature demo
├── README.md           # User documentation
├── CONTRIBUTING.md     # AI agent guide
├── ARCHITECTURE.md     # This file
├── examples/           # Example .http files
│   ├── simple-get.http
│   ├── post-json.http
│   ├── with-env-vars.http
│   └── with-shell-commands.http
└── internal/           # Internal packages
    ├── models/
    │   ├── request.go
    │   └── response.go
    ├── parser/
    │   └── parser.go
    ├── validator/
    │   └── validator.go
    └── client/
        └── client.go
```

## Extension Points

### Adding New Preprocessing Features

Location: `internal/parser/parser.go`

Steps:
1. Add function in `parser.go`
2. Call from `preprocess()`
3. Add test in `test.sh`
4. Add example in `examples/`

Example:
```go
func preprocess(content string) string {
    // ... existing preprocessing ...
    line = myNewFeature(line)
    return processedContent
}

func myNewFeature(line string) string {
    // Implementation
    return line
}
```

### Adding New Validation Rules

Location: `internal/validator/validator.go`

Steps:
1. Add validation function
2. Call from `Validate()`
3. Add to errors or warnings
4. Add test case

Example:
```go
func Validate(req *models.HTTPRequest) *ValidationResult {
    // ... existing validations ...
    validateMyNewRule(req, result)
    return result
}

func validateMyNewRule(req *models.HTTPRequest, result *ValidationResult) {
    if /* condition */ {
        result.Errors = append(result.Errors, "Error message")
    }
}
```

### Adding New Command-Line Flags

Location: `main.go`

Steps:
1. Define flag with `flag.Bool/String/Int()`
2. Use flag value in code
3. Update README usage section

Example:
```go
timeout := flag.Int("timeout", 30, "Request timeout in seconds")
flag.Parse()
// Use *timeout
```

### Adding HTTP/2 or HTTP/3 Support

Location: `internal/client/client.go`

Current implementation uses raw TCP. For HTTP/2/3:
1. Add dependency: `golang.org/x/net/http2`
2. Modify `Send()` to detect version
3. Use appropriate connection type
4. Update validator to allow new versions

## Testing Strategy

### Manual Testing

```bash
make test              # Run automated tests
make examples          # Test all examples
./demo.sh             # Visual feature demo
```

### Test Coverage

Current tests in `test.sh`:
1. Version flag
2. Help flag
3. Simple GET parsing
4. POST with body
5. Environment variables
6. Shell commands
7. Comment handling
8. No-color mode
9. Validation errors
10. Content-Length auto-add
11. Host header auto-add

### Adding New Tests

Add to `test.sh`:
```bash
echo "Test N: Description"
# Test command with validation
echo "✓ Test passed"
```

## Performance Considerations

### Current Performance

- Parser: O(n) where n = input size
- Validator: O(h) where h = number of headers
- Client: Network-bound, 30s timeout

### Optimization Opportunities

1. **Parser**: Already efficient, single pass
2. **Validator**: Could cache compiled regexes
3. **Client**: Could reuse connections (keep-alive)

## Security Considerations

### Current Security Features

1. **Input Validation**: All inputs validated before use
2. **TLS Support**: HTTPS with proper certificate validation
3. **No Code Injection**: Shell commands are explicit, controlled by user
4. **Timeout Protection**: 30s timeout prevents hanging
5. **Error Handling**: All errors caught and reported

### Security Scan Results

- CodeQL: ✅ 0 vulnerabilities found
- No external dependencies reduces attack surface

### Security Best Practices

1. Don't commit secrets in .http files
2. Use environment variables for sensitive data
3. Review shell commands before execution
4. Use HTTPS for sensitive requests

## Future Enhancements

### Potential Features

1. **HTTP/2 and HTTP/3 Support**: Use golang.org/x/net/http2
2. **Request History**: Save sent requests
3. **Response Formatting**: JSON pretty-print, HTML rendering
4. **Authentication Helpers**: OAuth, JWT support
5. **Test Assertions**: Validate responses in .http files
6. **Variable Files**: Separate .env file support
7. **Request Collections**: Run multiple requests
8. **Performance Testing**: Benchmarking mode

### Backward Compatibility

All future enhancements should:
- Maintain existing .http file format
- Keep command-line interface compatible
- Not break existing features
- Add new features as optional flags

## Contributing Guidelines

See `CONTRIBUTING.md` for detailed contribution guidelines.

Key points:
- Make minimal changes
- Follow existing patterns
- Add tests for new features
- Update documentation
- Keep it simple for AI agents

## License

See LICENSE file for details.

---

**Document Version**: 1.0.0  
**Last Updated**: 2026-02-25  
**Maintained by**: AI Agents (GitHub Copilot)
