# Contributing Guide for AI Agents

This document is designed to help AI agents understand and maintain this codebase effectively.

## Project Overview

This is a command-line HTTP client that:
1. Reads `.http` files from stdin
2. Preprocesses them (comments, env vars, shell commands)
3. Validates against HTTP standards
4. Sends requests over TCP (HTTP/HTTPS)
5. Outputs responses with colored formatting

## Architecture

```
main.go                 → CLI entry point, coordinates all operations
internal/
  ├── models/          → Data structures (request, response)
  ├── parser/          → Parse .http files and preprocess
  ├── validator/       → Validate requests against HTTP RFC
  └── client/          → Send requests over TCP
```

## Code Design Principles

### 1. Simplicity First
- No external dependencies for core functionality
- Linear flow, avoid complex abstractions
- One responsibility per function
- Clear function and variable names

### 2. Clear Comments
- Every package has a clear purpose statement
- Functions explain what they do and why
- Special cases are documented inline

### 3. Error Handling
- Always return errors, don't panic
- Use fmt.Errorf with %w for error wrapping
- Provide context in error messages

### 4. Testability
- Pure functions where possible
- Separate I/O from logic
- Dependencies are passed in, not imported globally

## Key Components

### Parser (`internal/parser/parser.go`)

**Purpose**: Convert .http file format to HTTPRequest struct

**Flow**:
1. `Parse()` - Entry point
2. `preprocess()` - Remove comments, expand variables, execute commands
3. `parseHTTP()` - Parse HTTP request format

**To Modify**:
- Add new preprocessing features in `preprocess()`
- Add new variable formats in `expandEnvVars()`
- Add new command formats in `executeShellCommands()`

### Validator (`internal/validator/validator.go`)

**Purpose**: Validate requests against HTTP standards

**Flow**:
1. `Validate()` - Entry point, calls all validators
2. Individual validators for: method, URL, version, headers, body
3. Returns ValidationResult with errors and warnings

**To Modify**:
- Add new validation rules as separate functions
- Call them from `Validate()`
- Add to errors (must fix) or warnings (should fix)

### Client (`internal/client/client.go`)

**Purpose**: Send HTTP requests over TCP

**Flow**:
1. `Send()` - Entry point
2. Parse URL, determine host/port
3. Open TCP connection (with TLS for HTTPS)
4. Send raw HTTP request
5. `readResponse()` - Parse response

**To Modify**:
- HTTP/2, HTTP/3 support would need new implementation
- Timeout adjustments in `conn.SetDeadline()`
- New encoding types in `readResponse()`

## Common Tasks

### Adding a New Preprocessing Feature

1. Add function in `parser/parser.go`
2. Call it from `preprocess()`
3. Add example in `examples/`
4. Update README

Example:
```go
// In preprocess()
line = myNewFeature(line)

func myNewFeature(line string) string {
    // Implementation
    return line
}
```

### Adding a New Validation Rule

1. Add function in `validator/validator.go`
2. Call it from `Validate()`
3. Add to result.Errors or result.Warnings

Example:
```go
// In Validate()
validateMyNewRule(req, result)

func validateMyNewRule(req *models.HTTPRequest, result *ValidationResult) {
    if /* check fails */ {
        result.Errors = append(result.Errors, "Error message")
    }
}
```

### Adding a New Command-Line Flag

1. Add flag in `main.go`
2. Pass to relevant function
3. Update README

Example:
```go
timeout := flag.Int("timeout", 30, "Request timeout in seconds")
flag.Parse()
// Use *timeout
```

## Testing

### Manual Testing

```bash
# Build
go build -o http

# Test basic functionality
echo "GET https://httpbin.org/get HTTP/1.1
Host: httpbin.org" | ./http -dry-run

# Test preprocessing
export TEST_VAR="value"
echo "GET https://example.com HTTP/1.1
Host: example.com
X-Test: \${TEST_VAR}" | ./http -dry-run

# Test validation
echo "GET example.com HTTP/1.1" | ./http -dry-run
# Should show error: URL must include scheme
```

### Test Files

Use files in `examples/` for testing:
```bash
cat examples/simple-get.http | ./http -dry-run
cat examples/post-json.http | ./http -dry-run
cat examples/with-env-vars.http | ./http -dry-run
cat examples/with-shell-commands.http | ./http -dry-run
```

## Making Changes

### Before Changes
1. Understand the requirement fully
2. Identify which component(s) to modify
3. Read existing code in that component

### During Changes
1. Make minimal modifications
2. Keep existing patterns and style
3. Add comments for non-obvious code
4. Update README if user-facing

### After Changes
1. Build: `go build -o http`
2. Test with examples: `cat examples/*.http | ./http -dry-run`
3. Test edge cases
4. Update CONTRIBUTING.md if architecture changes

## Code Style

### Naming
- Packages: lowercase, single word
- Files: lowercase with underscores if needed
- Functions: CamelCase (exported), camelCase (internal)
- Variables: camelCase
- Constants: CamelCase or SCREAMING_SNAKE_CASE

### Formatting
- Use `gofmt` (automatic)
- Line length: flexible, but break long lines reasonably
- Blank lines between logical sections

### Comments
- Package comment at top of package files
- Exported functions always have comments
- Complex logic has inline comments
- "AI Agent Note:" for special considerations

## Dependencies

### Current
- Standard library only for core functionality
- No external dependencies for HTTP client logic

### Adding Dependencies
- Avoid if possible
- If needed, use `go get`
- Document why it's needed
- Keep dependency tree minimal

## Common Patterns

### Error Handling
```go
if err != nil {
    return fmt.Errorf("context: %w", err)
}
```

### String Building
```go
var sb strings.Builder
sb.WriteString("text")
return sb.String()
```

### Iteration
```go
for key, value := range collection {
    // process
}
```

## Debugging

### Print Debugging
```go
fmt.Fprintf(os.Stderr, "Debug: %+v\n", variable)
```

### Trace Request Flow
1. Parser: Add print in `Parse()`
2. Validator: Add print in `Validate()`
3. Client: Add print in `Send()`

### Common Issues
- DNS resolution fails: Expected in sandboxed environments, use `-dry-run`
- TLS errors: Check URL scheme and certificate
- Parse errors: Check .http file format
- Validation errors: Read error message, fix request

## Release Process

1. Test all examples
2. Update README if needed
3. Build: `go build -o http`
4. Tag version: `git tag v1.x.x`
5. Push: `git push --tags`

## Questions?

This codebase is designed to be self-documenting. If something is unclear:
1. Read the code - it should be straightforward
2. Check comments - they explain the "why"
3. Run the code - see what happens
4. Modify carefully - make minimal changes

Remember: Simplicity and clarity over cleverness.