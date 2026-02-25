# HTTP Client Tool

A command-line HTTP client written in Go that processes `.http` files, validates requests against HTTP standards, and sends them over TCP with colored output.

## Features

- 🚀 **HTTP & HTTPS Support**: Send requests over plain HTTP or secure HTTPS
- ✅ **Request Validation**: Validates requests against HTTP RFC standards with colored error/warning output
- 📝 **.http File Format**: Human-readable request format with preprocessing support
- 🔧 **Preprocessing**: 
  - Comments (# or //)
  - Environment variable substitution (${VAR} or $VAR)
  - Shell command execution $(command)
- 🎨 **Colored Output**: Beautiful colored output for requests, responses, and validation
- 📥 **stdin/stdout**: Reads from stdin and writes to stdout for easy piping
- 👁️ **Dry Run Mode**: Preview preprocessed and validated requests without sending

## Installation

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
# From file
cat request.http | ./http

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
  -dry-run       Show preprocessed and validated request without sending
  -no-color      Disable colored output
  -v             Verbose output
```

### .http File Format

#### Simple GET Request

```http
GET https://api.example.com/users HTTP/1.1
Host: api.example.com
User-Agent: MyClient/1.0
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

## Validation Rules

The tool validates requests against HTTP standards:

### Errors (Must Fix)
- Missing or invalid URL
- Missing scheme (http:// or https://)
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