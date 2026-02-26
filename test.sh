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
echo "Test 3: Simple GET request (dry-run)"
cat examples/simple-get.http | ./http -dry-run > /dev/null
echo "✓ Simple GET request parsed"
echo ""

# Test 4: POST with body
echo "Test 4: POST with JSON body (dry-run)"
cat examples/post-json.http | ./http -dry-run > /dev/null
echo "✓ POST request with body parsed"
echo ""

# Test 5: Environment variables
echo "Test 5: Environment variable substitution"
export TEST_TOKEN="token123"
export TEST_KEY="key456"
echo "GET https://example.com HTTP/1.1
Host: example.com
Authorization: Bearer \${TEST_TOKEN}
X-Key: \$TEST_KEY" | ./http -dry-run | grep -q "token123"
echo "✓ Environment variables substituted"
echo ""

# Test 6: Shell commands
echo "Test 6: Shell command execution"
YEAR=$(date +%Y)
echo "GET https://example.com HTTP/1.1
Host: example.com
X-Date: \$(date +%Y)" | ./http -dry-run -no-color | grep -q "X-Date: $YEAR"
echo "✓ Shell commands executed"
echo ""

# Test 7: Comments
echo "Test 7: Comment handling"
echo "# This is a comment
// This is also a comment
GET https://example.com HTTP/1.1
Host: example.com" | ./http -dry-run > /dev/null
echo "✓ Comments handled"
echo ""

# Test 8: No color mode
echo "Test 8: No-color mode"
cat examples/simple-get.http | ./http -dry-run -no-color > /dev/null
echo "✓ No-color mode works"
echo ""

# Test 9: Validation error
echo "Test 9: Validation error detection"
echo "GET example.com HTTP/1.1" | ./http -dry-run 2>&1 | grep -q "URL must include scheme" || exit 1
echo "✓ Validation errors detected"
echo ""

# Test 10: Content-Length auto-add
echo "Test 10: Content-Length auto-addition"
echo "POST https://example.com HTTP/1.1
Host: example.com

test body" | ./http -dry-run -no-color 2>&1 | grep -q "Content-Length: 9"
echo "✓ Content-Length auto-added"
echo ""

# Test 11: Host header auto-add
echo "Test 11: Host header auto-addition"
echo "GET https://example.com/path HTTP/1.1" | ./http -dry-run -no-color 2>&1 | grep -q "Host: example.com"
echo "✓ Host header auto-added"
echo ""

# Test 12: Save request to file
echo "Test 12: Save RFC compliant request to file"
TEMP_FILE="/tmp/test-save-request-$$.http"
cat examples/simple-get.http | ./http -save-request "$TEMP_FILE" > /dev/null 2>&1
if [ -f "$TEMP_FILE" ]; then
    echo "✓ Request saved to file"
    rm -f "$TEMP_FILE"
else
    echo "✗ Failed to save request to file"
    exit 1
fi
echo ""

# Test 13: Load request from file
echo "Test 13: Load RFC compliant request from file"
TEMP_FILE="/tmp/test-load-request-$$.http"
cat examples/simple-get.http | ./http -save-request "$TEMP_FILE" > /dev/null 2>&1
./http -load-request "$TEMP_FILE" -dry-run > /dev/null 2>&1
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

test body" | ./http -strict -dry-run 2>&1 | grep -q "Strict mode" || { echo "✗ Strict mode not working"; exit 1; }
# Also verify the command itself failed (exit code 1)
echo "POST https://example.com HTTP/1.1
Host: example.com

test body" | ./http -strict -dry-run > /dev/null 2>&1 && { echo "✗ Strict mode didn't fail the command"; exit 1; }
echo "✓ Strict mode enforced"
echo ""

# Test 15: Save preprocessed request with environment variables
echo "Test 15: Save preprocessed request with environment variables"
export TEST_VAR="test-value-123"
TEMP_FILE="/tmp/test-env-save-$$.http"
STDERR_FILE="/tmp/test-env-save-stderr-$$.txt"
echo "GET https://example.com HTTP/1.1
Host: example.com
X-Test: \${TEST_VAR}" | ./http -save-request "$TEMP_FILE" 2>"$STDERR_FILE"
if [ $? -eq 0 ] && grep -q "test-value-123" "$TEMP_FILE"; then
    echo "✓ Environment variables preprocessed and saved"
    rm -f "$TEMP_FILE" "$STDERR_FILE"
else
    echo "✗ Environment variables not preprocessed in saved file"
    cat "$STDERR_FILE"
    rm -f "$TEMP_FILE" "$STDERR_FILE"
    exit 1
fi
echo ""

echo "=== All tests passed! ==="
