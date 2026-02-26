package parser

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// TestParseComplexBody tests parsing with complex multiline body
func TestParseComplexBody(t *testing.T) {
	input := `POST https://api.example.com/data HTTP/1.1
Host: api.example.com
Content-Type: application/json

{
  "nested": {
    "field1": "value1",
    "field2": [1, 2, 3]
  },
  "array": ["a", "b", "c"],
  "bool": true,
  "null": null
}`

	req, err := Parse(input)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Validate body is preserved correctly
	if !strings.Contains(req.Body, "nested") {
		t.Errorf("Expected body to contain 'nested'")
	}
	if !strings.Contains(req.Body, "array") {
		t.Errorf("Expected body to contain 'array'")
	}
}

// TestParseCaseInsensitiveHeaders tests that header names are case-insensitive
func TestParseCaseInsensitiveHeaders(t *testing.T) {
	input := `GET https://example.com HTTP/1.1
host: example.com
content-type: application/json
USER-AGENT: test`

	req, err := Parse(input)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Headers should be accessible (case may be normalized)
	if req.Headers["host"] == "" && req.Headers["Host"] == "" {
		t.Errorf("Expected Host header to be present")
	}
}

// TestParseTrailingWhitespace tests handling of trailing whitespace
func TestParseTrailingWhitespace(t *testing.T) {
	input := `GET https://example.com HTTP/1.1  
Host: example.com  
User-Agent: test  `

	req, err := Parse(input)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if req.Method != "GET" {
		t.Errorf("Expected method GET")
	}
}

// TestParseURLWithComplexPath tests parsing URLs with complex paths
func TestParseURLWithComplexPath(t *testing.T) {
	input := `GET https://api.example.com/v1/users/123/posts?sort=date&limit=10#section HTTP/1.1
Host: api.example.com`

	req, err := Parse(input)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if !strings.Contains(req.URL, "/v1/users/123/posts") {
		t.Errorf("Expected complex path in URL")
	}

	if !strings.Contains(req.URL, "sort=date") {
		t.Errorf("Expected query parameters in URL")
	}
}

// TestParseURLWithPort tests parsing URLs with explicit port
func TestParseURLWithPort(t *testing.T) {
	input := `GET https://example.com:8443/api HTTP/1.1
Host: example.com:8443`

	req, err := Parse(input)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if !strings.Contains(req.URL, ":8443") {
		t.Errorf("Expected port in URL")
	}
}

// TestParseHeadersWithColons tests headers with colons in values
func TestParseHeadersWithColons(t *testing.T) {
	input := `GET https://example.com HTTP/1.1
Host: example.com
X-Time: 12:34:56
X-URL: https://other.com/path`

	req, err := Parse(input)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if req.Headers["X-Time"] != "12:34:56" {
		t.Errorf("Expected X-Time header with colons")
	}

	if req.Headers["X-URL"] != "https://other.com/path" {
		t.Errorf("Expected X-URL header with URL value")
	}
}

// TestParseMultipleEnvVars tests multiple environment variables
func TestParseMultipleEnvVars(t *testing.T) {
	// Set test environment variables
	_ = os.Setenv("TEST_VAR1", "value1")
	_ = os.Setenv("TEST_VAR2", "value2")
	_ = os.Setenv("TEST_VAR3", "value3")
	defer func() {
		_ = os.Unsetenv("TEST_VAR1")
		_ = os.Unsetenv("TEST_VAR2")
		_ = os.Unsetenv("TEST_VAR3")
	}()

	input := `GET https://example.com HTTP/1.1
Host: example.com
X-Header1: ${TEST_VAR1}
X-Header2: $TEST_VAR2
X-Header3: prefix-${TEST_VAR3}-suffix`

	req, err := Parse(input)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if req.Headers["X-Header1"] != "value1" {
		t.Errorf("Expected X-Header1 to be 'value1', got %s", req.Headers["X-Header1"])
	}

	if req.Headers["X-Header2"] != "value2" {
		t.Errorf("Expected X-Header2 to be 'value2', got %s", req.Headers["X-Header2"])
	}

	if req.Headers["X-Header3"] != "prefix-value3-suffix" {
		t.Errorf("Expected X-Header3 to be 'prefix-value3-suffix', got %s", req.Headers["X-Header3"])
	}
}

// TestParseEnvVarInBody tests environment variable substitution in body
func TestParseEnvVarInBody(t *testing.T) {
	_ = os.Setenv("TEST_USER", "alice")
	_ = os.Setenv("TEST_ID", "12345")
	defer func() {
		_ = os.Unsetenv("TEST_USER")
		_ = os.Unsetenv("TEST_ID")
	}()

	input := `POST https://example.com HTTP/1.1
Host: example.com
Content-Type: application/json

{"user":"${TEST_USER}","id":"$TEST_ID"}`

	req, err := Parse(input)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if !strings.Contains(req.Body, "alice") {
		t.Errorf("Expected body to contain substituted user 'alice'")
	}

	if !strings.Contains(req.Body, "12345") {
		t.Errorf("Expected body to contain substituted id '12345'")
	}
}

// TestParseUndefinedEnvVar tests handling of undefined environment variables
func TestParseUndefinedEnvVar(t *testing.T) {
	// Ensure the variable is not set
	_ = os.Unsetenv("UNDEFINED_VAR_TEST")

	input := `GET https://example.com HTTP/1.1
Host: example.com
X-Header: ${UNDEFINED_VAR_TEST}`

	req, err := Parse(input)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Undefined variables should be left as empty or as-is
	// The behavior may vary, but parsing should not fail
	if req.Method != "GET" {
		t.Errorf("Expected request to parse despite undefined env var")
	}
}

// TestParseNestedShellCommands tests shell command execution
func TestParseNestedShellCommands(t *testing.T) {
	input := `GET https://example.com HTTP/1.1
Host: example.com
X-Echo: $(echo "test")
X-Date: $(date +%Y)
X-Pwd: $(pwd | head -c 10)`

	req, err := Parse(input)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Verify shell commands were executed
	if req.Headers["X-Echo"] != "test" {
		t.Errorf("Expected X-Echo to be 'test', got %s", req.Headers["X-Echo"])
	}

	// X-Date should contain a year (4 digits)
	if len(req.Headers["X-Date"]) != 4 {
		t.Logf("Warning: X-Date may not be a year: %s", req.Headers["X-Date"])
	}
}

// TestParseSpecialCharactersInHeaders tests special characters
func TestParseSpecialCharactersInHeaders(t *testing.T) {
	input := `GET https://example.com HTTP/1.1
Host: example.com
X-Special: value-with-dash_and_underscore.and.dot
X-Encoded: value%20with%20encoding
X-Symbols: value!@#$%`

	req, err := Parse(input)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if req.Headers["X-Special"] == "" {
		t.Errorf("Expected X-Special header")
	}

	if req.Headers["X-Encoded"] == "" {
		t.Errorf("Expected X-Encoded header")
	}
}

// TestParseBodyWithSpecialChars tests body with special characters
func TestParseBodyWithSpecialChars(t *testing.T) {
	input := `POST https://example.com HTTP/1.1
Host: example.com
Content-Type: application/json

{"text":"Special chars: \n\t\r\\ \"quotes\"","unicode":"émojis 😀🎉"}`

	req, err := Parse(input)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if !strings.Contains(req.Body, "Special chars") {
		t.Errorf("Expected body to contain special characters")
	}

	if !strings.Contains(req.Body, "émojis") {
		t.Errorf("Expected body to contain unicode characters")
	}
}

// TestParseVeryLongHeader tests handling of very long header values
func TestParseVeryLongHeader(t *testing.T) {
	longValue := strings.Repeat("a", 1000)
	input := fmt.Sprintf(`GET https://example.com HTTP/1.1
Host: example.com
X-Long-Header: %s`, longValue)

	req, err := Parse(input)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(req.Headers["X-Long-Header"]) != 1000 {
		t.Errorf("Expected long header to be preserved")
	}
}

// TestParseVeryLongBody tests handling of large request body
func TestParseVeryLongBody(t *testing.T) {
	longBody := strings.Repeat("x", 5000)
	input := fmt.Sprintf(`POST https://example.com HTTP/1.1
Host: example.com
Content-Type: text/plain

%s`, longBody)

	req, err := Parse(input)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(req.Body) != 5000 {
		t.Errorf("Expected body length 5000, got %d", len(req.Body))
	}
}

// TestParseMixedCommentStyles tests different comment styles
func TestParseMixedCommentStyles(t *testing.T) {
	input := `# Hash comment at start
// Double slash comment
GET https://example.com HTTP/1.1
# Comment in headers
Host: example.com
// Another comment
User-Agent: test
### Multiple hashes`

	req, err := Parse(input)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if req.Method != "GET" {
		t.Errorf("Expected method GET despite comments")
	}

	// Ensure comments don't appear in headers
	for key, value := range req.Headers {
		if strings.Contains(value, "#") || strings.Contains(value, "//") {
			t.Errorf("Comment found in header %s: %s", key, value)
		}
	}
}

// TestParseBlankLinesInBody tests blank lines within the body
func TestParseBlankLinesInBody(t *testing.T) {
	input := `POST https://example.com HTTP/1.1
Host: example.com
Content-Type: text/plain

line1

line2

line3`

	req, err := Parse(input)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Body should preserve blank lines
	if !strings.Contains(req.Body, "line1") || !strings.Contains(req.Body, "line2") {
		t.Errorf("Expected body to contain all lines")
	}

	// Count newlines in body
	newlineCount := strings.Count(req.Body, "\n")
	if newlineCount < 3 {
		t.Errorf("Expected at least 3 newlines in body (for blank lines)")
	}
}

// TestParseIPv6URL tests parsing URLs with IPv6 addresses
func TestParseIPv6URL(t *testing.T) {
	input := `GET https://[::1]/api HTTP/1.1
Host: [::1]`

	req, err := Parse(input)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if !strings.Contains(req.URL, "[::1]") {
		t.Errorf("Expected IPv6 address in URL")
	}
}

// TestParseFormDataBody tests form data body
func TestParseFormDataBody(t *testing.T) {
	input := `POST https://example.com/login HTTP/1.1
Host: example.com
Content-Type: application/x-www-form-urlencoded

username=john&password=secret123&remember=true`

	req, err := Parse(input)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if !strings.Contains(req.Body, "username=john") {
		t.Errorf("Expected form data in body")
	}

	if req.Headers["Content-Type"] != "application/x-www-form-urlencoded" {
		t.Errorf("Expected form content type")
	}
}

// TestParseXMLBody tests XML request body
func TestParseXMLBody(t *testing.T) {
	input := `POST https://api.example.com/soap HTTP/1.1
Host: api.example.com
Content-Type: application/xml

<?xml version="1.0"?>
<request>
  <user>john</user>
  <action>login</action>
</request>`

	req, err := Parse(input)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if !strings.Contains(req.Body, "<?xml") {
		t.Errorf("Expected XML declaration in body")
	}

	if !strings.Contains(req.Body, "<request>") {
		t.Errorf("Expected XML tags in body")
	}
}
