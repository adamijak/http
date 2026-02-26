#!/bin/bash
# Test script for HTTP client tool
# AI Agent Note: Simple test suite to verify functionality

set -e

echo "=== HTTP Client Test Suite ==="
echo ""

# Check formatting
echo "Checking code formatting..."
UNFORMATTED=$(gofmt -l .)
if [ -n "$UNFORMATTED" ]; then
    echo "✗ Code is not formatted. Run 'gofmt -w .' to fix:"
    echo "$UNFORMATTED"
    exit 1
fi
echo "✓ Code formatting is correct"
echo ""

# Run linting
echo "Running go vet..."
go vet ./...
echo "✓ Linting passed"
echo ""

# Run unit tests
echo "Running unit tests..."
go test ./... -v
echo "✓ Unit tests passed"
echo ""

# Build
echo "Building..."
go build -o http
echo "✓ Build successful"
echo ""

# Test 1: Version
echo "Test 1: Version flag"
./http -version
echo "✓ Version flag works"
echo ""

# Test 2: Help
echo "Test 2: Help flag"
./http -h > /dev/null
echo "✓ Help flag works"
echo ""

# Test 3: Simple GET
echo "Test 3: Simple GET request (output RFC format)"
cat examples/simple-get.http | ./http --no-send > /dev/null
echo "✓ Simple GET request parsed"
echo ""

# Test 4: POST with body
echo "Test 4: POST with JSON body (output RFC format)"
cat examples/post-json.http | ./http --no-send > /dev/null
echo "✓ POST request with body parsed"
echo ""

# Test 5: Environment variables
echo "Test 5: Environment variable substitution"
export TEST_TOKEN="token123"
export TEST_KEY="key456"
echo "GET https://example.com HTTP/1.1
Host: example.com
Authorization: Bearer \${TEST_TOKEN}
X-Key: \$TEST_KEY" | ./http --no-send | grep -q "token123"
echo "✓ Environment variables substituted"
echo ""

# Test 6: Shell commands
echo "Test 6: Shell command execution"
YEAR=$(date +%Y)
echo "GET https://example.com HTTP/1.1
Host: example.com
X-Date: \$(date +%Y)" | ./http --no-send -no-color | grep -q "X-Date: $YEAR"
echo "✓ Shell commands executed"
echo ""

# Test 7: Comments
echo "Test 7: Comment handling"
echo "# This is a comment
// This is also a comment
GET https://example.com HTTP/1.1
Host: example.com" | ./http --no-send > /dev/null
echo "✓ Comments handled"
echo ""

# Test 8: No color mode
echo "Test 8: No-color mode"
cat examples/simple-get.http | ./http --no-send -no-color > /dev/null
echo "✓ No-color mode works"
echo ""

# Test 9: Validation error
echo "Test 9: Validation error detection"
echo "GET example.com HTTP/1.1" | ./http --no-send 2>&1 | grep -q "URL must include scheme" || exit 1
echo "✓ Validation errors detected"
echo ""

# Test 10: Content-Length auto-add
echo "Test 10: Content-Length auto-addition"
echo "POST https://example.com HTTP/1.1
Host: example.com

test body" | ./http --no-send -no-color 2>&1 | grep -q "Content-Length: 9"
echo "✓ Content-Length auto-added"
echo ""

# Test 11: Host header auto-add
echo "Test 11: Host header auto-addition"
echo "GET https://example.com/path HTTP/1.1" | ./http --no-send -no-color 2>&1 | grep -q "Host: example.com"
echo "✓ Host header auto-added"
echo ""

# Test 12: Output RFC compliant request to stdout
echo "Test 12: Output RFC compliant request to stdout"
OUTPUT=$(./http --no-send < examples/simple-get.http 2>&1)
# Check for CRLF using od to ensure it works across all shells
if echo "$OUTPUT" | od -An -tx1 | grep -q "0d 0a"; then
    echo "✓ RFC compliant request output to stdout"
else
    echo "✗ Failed to output RFC compliant request"
    exit 1
fi
echo ""

# Test 13: Load request from file (with -f flag)
echo "Test 13: Load RFC compliant request from file"
TEMP_FILE="/tmp/test-load-request-$$.http"
./http --no-send < examples/simple-get.http > "$TEMP_FILE" 2>&1
./http -f "$TEMP_FILE" --no-send > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✓ Request loaded from file"
    rm -f "$TEMP_FILE"
else
    echo "✗ Failed to load request from file"
    rm -f "$TEMP_FILE"
    exit 1
fi
echo ""

# Test 14: Strict mode with warnings
echo "Test 14: Strict mode fails on warnings"
echo "POST https://example.com HTTP/1.1
Host: example.com

test body" | ./http -strict --no-send 2>&1 | grep -q "Strict mode" || { echo "✗ Strict mode not working"; exit 1; }
# Also verify the command itself failed (exit code 1)
echo "POST https://example.com HTTP/1.1
Host: example.com

test body" | ./http -strict --no-send > /dev/null 2>&1 && { echo "✗ Strict mode didn't fail the command"; exit 1; }
echo "✓ Strict mode enforced"
echo ""

# Test 15: Output preprocessed request with environment variables
echo "Test 15: Output preprocessed request with environment variables"
export TEST_VAR="test-value-123"
TEMP_FILE="/tmp/test-env-save-$$.http"
echo "GET https://example.com HTTP/1.1
Host: example.com
X-Test: \${TEST_VAR}" | ./http --no-send > "$TEMP_FILE" 2>&1
if [ $? -eq 0 ] && grep -q "test-value-123" "$TEMP_FILE"; then
    echo "✓ Environment variables preprocessed in output"
    rm -f "$TEMP_FILE"
else
    echo "✗ Environment variables not preprocessed in output"
    cat "$TEMP_FILE"
    rm -f "$TEMP_FILE"
    exit 1
fi
echo ""

# Test 16: Load HTP format file with -f flag
echo "Test 16: Load HTP format file with -f flag"
TEMP_FILE="/tmp/test-htp-format-$$.http"
echo "# Comment in HTP format
GET https://example.com HTTP/1.1
Host: example.com" > "$TEMP_FILE"
./http -f "$TEMP_FILE" --no-send > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✓ HTP format loaded from file"
    rm -f "$TEMP_FILE"
else
    echo "✗ Failed to load HTP format from file"
    rm -f "$TEMP_FILE"
    exit 1
fi
echo ""

# Test 17: Pipeline flow - HTP to RFC to tool again
echo "Test 17: Pipeline flow (--no-send output can be piped back)"
TEMP_FILE="/tmp/test-pipeline-$$.http"
cat > "$TEMP_FILE" << 'INNER_EOF'
# Comment
GET /api HTTP/1.1
Host: example.com
INNER_EOF

export TEST_PIPE_VAR="test-value"
OUTPUT1=$(./http -f "$TEMP_FILE" --no-send 2>&1)
OUTPUT2=$(echo "$OUTPUT1" | ./http --no-send 2>&1)

if echo "$OUTPUT2" | grep -q "https://example.com/api"; then
    echo "✓ Pipeline flow works"
    rm -f "$TEMP_FILE"
else
    echo "✗ Pipeline flow failed"
    echo "Output: $OUTPUT2"
    rm -f "$TEMP_FILE"
    exit 1
fi
echo ""

echo "=== All tests passed! ==="
