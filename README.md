# HTTP Client Tool

[![Test](https://github.com/adamijak/http/workflows/Test/badge.svg)](https://github.com/adamijak/http/actions/workflows/test.yml)
[![Release](https://github.com/adamijak/http/workflows/Release/badge.svg)](https://github.com/adamijak/http/actions/workflows/release.yml)

A command-line HTTP client written in Go that processes `.http` files, validates requests against HTTP standards, and sends them over TCP with colored output.

## File Formats

This tool supports two HTTP request formats and **automatically detects** which one you're using:

### HTP Format (HTTP Template Protocol)

Our friendly, human-readable format with preprocessing support:
- **Comments**: Lines starting with `#` or `//`
- **Environment variables**: `${VAR_NAME}` or `$VAR_NAME`
- **Shell commands**: `$(command)`
- **Line endings**: Standard newlines (`\n`)
- **Detection**: Automatically used when file doesn't contain `\r\n`
- **Use case**: Writing request templates with dynamic content

**Example HTP file:**
```http
# API Request Template
POST https://api.example.com/data HTTP/1.1
Host: api.example.com
Authorization: Bearer ${API_TOKEN}
X-Request-ID: $(uuidgen)

{"user": "$(whoami)"}
```

### RFC Compliant Format

Standard HTTP protocol format ready to send over the wire:
- **No preprocessing**: Pure HTTP request format
- **Line endings**: CRLF (`\r\n`) as per RFC specification
- **Detection**: Automatically used when file contains `\r\n`
- **Use case**: Saved/exported requests, sharing exact requests, production templates

**Example RFC compliant file:**
```http
POST https://api.example.com/data HTTP/1.1
Host: api.example.com
Authorization: Bearer abc123token
Content-Length: 15

{"user": "john"}
```
*(Note: In actual file, line endings would be \r\n)*

### Format Comparison

| Feature | HTP Format | RFC Compliant Format |
|---------|-----------|---------------------|
| Line endings | `\n` | `\r\n` |
| Comments | Yes (`#`, `//`) | No |
| Env variables | Yes (`${VAR}`, `$VAR`) | No |
| Shell commands | Yes (`$(cmd)`) | No |
| Preprocessing | Yes | No |
| Auto-detection | No `\r\n` in file | Contains `\r\n` |
| Use with `-f` | ✅ Yes | ✅ Yes |
| Use with stdin | ✅ Yes | ✅ Yes |

## Features

- 🚀 **HTTP & HTTPS Support**: Send requests over plain HTTP or secure HTTPS
- ✅ **Request Validation**: Validates requests against HTTP RFC standards with colored error/warning output
- 📝 **Dual Format Support**: HTP format (human-friendly with preprocessing) and RFC compliant format
- 🔧 **Preprocessing**: 
  - Comments (# or //)
  - Environment variable substitution (${VAR} or $VAR)
  - Shell command execution $(command)
- 💾 **Save/Load Requests**: Save preprocessed RFC compliant requests to files and load them later
- 🔒 **Strict Mode**: Enforce full RFC compliance by failing on validation warnings
- 🎨 **Colored Output**: Beautiful colored output for requests, responses, and validation
- 📥 **stdin/stdout**: Reads from stdin and writes to stdout for easy piping
- 👁️ **Dry Run Mode**: Preview preprocessed and validated requests without sending

## Installation

### Pre-built Binaries

Download pre-built binaries from the [Releases page](https://github.com/adamijak/http/releases).

```bash
# Example: Linux AMD64
wget https://github.com/adamijak/http/releases/latest/download/http-linux-amd64
chmod +x http-linux-amd64
sudo mv http-linux-amd64 /usr/local/bin/http
```

### Build from Source

```bash
go build -o http
```

Or install directly:

```bash
go install github.com/adamijak/http@latest
```

## Usage

### Basic Usage

```bash
# From stdin
cat request.http | ./http

# From file (supports both HTP and RFC compliant formats)
./http -f request.http

# From heredoc
./http <<EOF
GET https://api.github.com/users/octocat HTTP/1.1
User-Agent: http-client/1.0
EOF
```

### Command-Line Options

```bash
./http [OPTIONS]

Options:
  -f FILE             Read request from FILE (auto-detects HTP or RFC compliant format)
  -dry-run            Show preprocessed and validated request without sending
  -no-color           Disable colored output
  -no-secure          Send request in plain HTTP instead of HTTPS
  -save-request FILE  Save the preprocessed RFC compliant request to FILE instead of sending
  -strict             Strict mode: fail on any validation warnings (RFC compliance enforcement)
  -v                  Verbose output
  -version            Show version information
```

### HTP Format (HTTP Template Protocol)

HTP is our friendly, human-readable request format with preprocessing support. Files in this format use standard newlines (`\n`) and can contain:

#### Simple GET Request

```http
GET https://api.example.com/users HTTP/1.1
Host: api.example.com
User-Agent: MyClient/1.0
```

#### Path-Only Request (RFC Standard)

You can use just the path and specify the host in the Host header:

```http
GET /users HTTP/1.1
Host: api.example.com
User-Agent: MyClient/1.0
```

By default, this will use HTTPS. To force plain HTTP, use the `--no-secure` flag:

```bash
cat request.http | ./http --no-secure
```

#### POST Request with Body

```http
POST https://api.example.com/users HTTP/1.1
Host: api.example.com
Content-Type: application/json

{
  "name": "John Doe",
  "email": "john@example.com"
}
```

#### Using Comments

```http
# This is a comment
// This is also a comment
GET https://api.example.com/status HTTP/1.1
# Comments can appear anywhere
Host: api.example.com
```

#### Environment Variables

```http
# Using ${VAR} syntax
GET https://api.example.com/users HTTP/1.1
Host: api.example.com
Authorization: Bearer ${API_TOKEN}

# Using $VAR syntax
X-API-Key: $API_KEY
```

Set environment variables:
```bash
export API_TOKEN="your-token-here"
export API_KEY="your-key-here"
cat request.http | ./http
```

#### Shell Command Execution

```http
GET https://api.example.com/request HTTP/1.1
Host: api.example.com
X-Request-ID: $(uuidgen)
X-Timestamp: $(date +%s)
```

#### Complete Example

```http
# API Request Example
# Make sure to set API_TOKEN environment variable

POST https://api.example.com/data HTTP/1.1
Host: api.example.com
Content-Type: application/json
Authorization: Bearer ${API_TOKEN}
X-Request-ID: $(uuidgen)
User-Agent: http-client/1.0

{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "message": "Hello from http client"
}
```

## Examples

### Check Request Before Sending

```bash
cat request.http | ./http -dry-run
```

Output:
```
✓ Validation passed

--- Preprocessed Request ---
POST https://api.example.com/data HTTP/1.1
Host: api.example.com
Content-Type: application/json
Authorization: Bearer abc123token
X-Request-ID: 550e8400-e29b-41d4-a716-446655440000
...
```

### Working with Both Formats

The tool automatically detects whether your input is in HTP or RFC compliant format:

```bash
# Read HTP format from file (will preprocess)
./http -f template.http

# Read RFC compliant format from file (no preprocessing needed)
./http -f saved-request.http

# Read from stdin (supports both formats)
cat request.http | ./http
```

### Save Preprocessed RFC Compliant Request

Convert HTP format to RFC compliant format and save for reuse:

```bash
# Save HTP request as RFC compliant (env vars substituted, comments removed)
cat template.http | ./http -save-request saved-request.http

# Or from file
./http -f template.http -save-request saved-request.http
```

Output:
```
✓ Validation passed
✓ RFC compliant request saved to: saved-request.http
```

The saved file is in **RFC compliant format** with:
- Environment variables substituted
- Shell commands executed
- Comments removed
- Proper CRLF (`\r\n`) line endings
- Auto-added headers (Content-Length, Host)
- Ready to send over the wire

### Load and Send Any Format

Load requests from file in either format:

```bash
# Load HTP format (will preprocess)
./http -f template.http

# Load RFC compliant format (no preprocessing)
./http -f saved-request.http

# Preview before sending
./http -f request.http -dry-run

# Send with strict validation
./http -f request.http -strict
```

**Format Detection**: The tool automatically detects the format:
- **HTP format**: Contains standard newlines (`\n`) - will preprocess
- **RFC compliant**: Contains CRLF (`\r\n`) - used as-is

This is useful for:
- Storing request templates with dynamic content (HTP format)
- Sharing exact requests between team members (RFC compliant)
- Debugging by saving intermediate states
- Version controlling request templates (both formats)

### Strict RFC Compliance Mode

Use strict mode to enforce full RFC compliance (fails on warnings):

```bash
# This will fail if request has any validation warnings
cat request.http | ./http -strict
```

Example output:
```
Validation Warnings:
  [WARN] Content-Type header is recommended when sending a body

Strict mode: Request has validation warnings and cannot be sent
```

Strict mode is useful for:
- Ensuring production requests are fully RFC compliant
- CI/CD pipelines where warnings should be treated as errors
- Testing request templates for compliance

### Send Request with Validation

```bash
cat request.http | ./http
```

Output with colors showing:
- ✓ Green for successful validation
- ⚠️ Yellow for warnings
- ❌ Red for errors
- Cyan for request/response metadata
- Yellow for headers

### Disable Colors

```bash
cat request.http | ./http -no-color
```

### Verbose Mode

```bash
cat request.http | ./http -v
```

### Send Plain HTTP Request

Use the `--no-secure` flag to force plain HTTP instead of HTTPS:

```bash
# Force HTTP for a path-only request
cat <<EOF | ./http --no-secure
GET /api/users HTTP/1.1
Host: example.com
EOF

# Force HTTP even when URL has https://
cat <<EOF | ./http --no-secure
GET https://example.com/api/users HTTP/1.1
Host: example.com
EOF
```

## Validation Rules

The tool validates requests against HTTP standards:

### Errors (Must Fix)
- Missing or invalid URL
- Missing scheme (http:// or https://) when using full URL format
- Missing Host header when using path-only URL (e.g., `/api/users`)
- Missing host in URL
- Invalid URL format
- Missing Host header for HTTP/1.1

### Warnings (Should Review)
- Non-standard HTTP methods
- Non-standard HTTP versions
- Non-standard URL schemes
- Methods with unexpected body (e.g., GET with body)
- Missing Content-Type when body is present
- Missing Content-Length (auto-added)

## Architecture

The project is designed to be AI-agent-friendly with clear separation of concerns:

```
.
├── main.go                 # Entry point and CLI handling
└── internal/
    ├── models/            # Data structures
    │   ├── request.go    # HTTPRequest model
    │   └── response.go   # HTTPResponse model
    ├── parser/           # .http file parsing
    │   └── parser.go     # Parsing and preprocessing logic
    ├── validator/        # Request validation
    │   └── validator.go  # RFC compliance validation
    └── client/           # HTTP client
        └── client.go     # TCP-based request sending
```

### AI Agent Maintenance Notes

Each package is self-contained with:
- Clear comments explaining purpose and functionality
- Simple, linear logic flow
- Separated concerns (parsing, validation, sending)
- No complex abstractions or inheritance
- Standard library only (no external dependencies for core functionality)

## HTTP Standards Supported

- HTTP/1.0, HTTP/1.1, HTTP/2, HTTP/3
- Standard HTTP methods: GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS, CONNECT, TRACE
- Custom HTTP methods (with warnings)
- Both HTTP and HTTPS protocols
- Chunked transfer encoding
- Content-Length based body reading

## Development

### Build

```bash
go build -o http
```

### Test

Run the test suite:

```bash
./test.sh
```

Or use Make:

```bash
# Run all checks (format, lint, test)
make check

# Individual commands
make format    # Format code with gofmt
make lint      # Run go vet
make test      # Run example tests
```

### Contributing

Before submitting a pull request:

1. **Format your code**: `make format` or `gofmt -w .`
2. **Lint your code**: `make lint` or `go vet ./...`
3. **Run all tests**: `./test.sh`
4. **Verify everything**: `make check`

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

### Manual Testing

```bash
# Create a test request
cat > test.http <<EOF
GET https://httpbin.org/get HTTP/1.1
Host: httpbin.org
User-Agent: http-client-test/1.0
EOF

# Test it
cat test.http | ./http
```

### Add to PATH

```bash
# Build
go build -o http

# Move to PATH location
sudo mv http /usr/local/bin/

# Now use from anywhere
echo "GET https://httpbin.org/get HTTP/1.1
Host: httpbin.org" | http
```

## License

See LICENSE file for details.