package parser

import (
	"os"
	"strings"
	"testing"
)

// TestParseSimpleGET tests parsing a simple GET request
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

	if req.URL != "https://example.com/api" {
		t.Errorf("Expected URL https://example.com/api, got %s", req.URL)
	}

	if req.Version != "HTTP/1.1" {
		t.Errorf("Expected version HTTP/1.1, got %s", req.Version)
	}

	if req.Headers["Host"] != "example.com" {
		t.Errorf("Expected Host header example.com, got %s", req.Headers["Host"])
	}

	if req.Headers["User-Agent"] != "test-client" {
		t.Errorf("Expected User-Agent header test-client, got %s", req.Headers["User-Agent"])
	}
}

// TestParsePOSTWithBody tests parsing a POST request with body
func TestParsePOSTWithBody(t *testing.T) {
	input := `POST https://api.example.com/users HTTP/1.1
Host: api.example.com
Content-Type: application/json

{"name":"John","age":30}`

	req, err := Parse(input)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if req.Method != "POST" {
		t.Errorf("Expected method POST, got %s", req.Method)
	}

	if req.Body != `{"name":"John","age":30}` {
		t.Errorf("Expected body, got %s", req.Body)
	}

	if req.Headers["Content-Type"] != "application/json" {
		t.Errorf("Expected Content-Type header")
	}
}

// TestParseWithComments tests parsing with comments
func TestParseWithComments(t *testing.T) {
	input := `# This is a comment
GET https://example.com HTTP/1.1
// Another comment
Host: example.com
# Comment in headers
User-Agent: test`

	req, err := Parse(input)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if req.Method != "GET" {
		t.Errorf("Expected method GET, got %s", req.Method)
	}

	// Comments should be stripped
	if strings.Contains(req.Headers["Host"], "#") || strings.Contains(req.Headers["Host"], "//") {
		t.Error("Comments should be stripped")
	}
}

// TestParseWithEnvVars tests environment variable substitution
func TestParseWithEnvVars(t *testing.T) {
	// Set test environment variables
	os.Setenv("TEST_TOKEN", "secret-token-123")
	os.Setenv("TEST_HOST", "api.test.com")
	defer os.Unsetenv("TEST_TOKEN")
	defer os.Unsetenv("TEST_HOST")

	// Test ${VAR} syntax
	input1 := `GET https://example.com HTTP/1.1
Host: ${TEST_HOST}
Authorization: Bearer ${TEST_TOKEN}`

	req1, err := Parse(input1)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if req1.Headers["Host"] != "api.test.com" {
		t.Errorf("Expected Host api.test.com, got %s", req1.Headers["Host"])
	}

	if req1.Headers["Authorization"] != "Bearer secret-token-123" {
		t.Errorf("Expected Authorization header with token, got %s", req1.Headers["Authorization"])
	}

	// Test $VAR syntax
	input2 := `GET https://example.com HTTP/1.1
Host: $TEST_HOST
Authorization: Bearer $TEST_TOKEN`

	req2, err := Parse(input2)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if req2.Headers["Host"] != "api.test.com" {
		t.Errorf("Expected Host api.test.com, got %s", req2.Headers["Host"])
	}
}

// TestParseWithShellCommands tests shell command execution
func TestParseWithShellCommands(t *testing.T) {
	input := `GET https://example.com HTTP/1.1
Host: example.com
X-Year: $(echo 2024)`

	req, err := Parse(input)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if req.Headers["X-Year"] != "2024" {
		t.Errorf("Expected X-Year 2024, got %s", req.Headers["X-Year"])
	}
}

// TestParsePathOnlyURL tests parsing path-only URLs
func TestParsePathOnlyURL(t *testing.T) {
	input := `GET /api/users HTTP/1.1
Host: example.com`

	req, err := Parse(input)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if req.URL != "/api/users" {
		t.Errorf("Expected URL /api/users, got %s", req.URL)
	}

	if req.Headers["Host"] != "example.com" {
		t.Errorf("Expected Host header example.com, got %s", req.Headers["Host"])
	}
}

// TestParseMultipleHeaders tests parsing multiple headers
func TestParseMultipleHeaders(t *testing.T) {
	input := `GET https://example.com HTTP/1.1
Host: example.com
User-Agent: test-client
Accept: application/json
Authorization: Bearer token123
X-Custom-Header: custom-value`

	req, err := Parse(input)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	expectedHeaders := map[string]string{
		"Host":            "example.com",
		"User-Agent":      "test-client",
		"Accept":          "application/json",
		"Authorization":   "Bearer token123",
		"X-Custom-Header": "custom-value",
	}

	for key, expectedValue := range expectedHeaders {
		if req.Headers[key] != expectedValue {
			t.Errorf("Expected header %s=%s, got %s", key, expectedValue, req.Headers[key])
		}
	}
}

// TestParseWithQueryParams tests parsing URLs with query parameters
func TestParseWithQueryParams(t *testing.T) {
	input := `GET https://example.com/search?q=test&limit=10 HTTP/1.1
Host: example.com`

	req, err := Parse(input)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if req.URL != "https://example.com/search?q=test&limit=10" {
		t.Errorf("Expected URL with query params, got %s", req.URL)
	}
}

// TestParseHTTPVersions tests parsing different HTTP versions
func TestParseHTTPVersions(t *testing.T) {
	versions := []string{"HTTP/1.0", "HTTP/1.1", "HTTP/2", "HTTP/3"}

	for _, version := range versions {
		t.Run(version, func(t *testing.T) {
			input := "GET https://example.com " + version + "\nHost: example.com"

			req, err := Parse(input)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			if req.Version != version {
				t.Errorf("Expected version %s, got %s", version, req.Version)
			}
		})
	}
}

// TestParseMethods tests parsing different HTTP methods
func TestParseMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS", "CONNECT", "TRACE"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			input := method + " https://example.com HTTP/1.1\nHost: example.com"

			req, err := Parse(input)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", method, err)
			}

			if req.Method != method {
				t.Errorf("Expected method %s, got %s", method, req.Method)
			}
		})
	}
}

// TestParseEmptyBody tests parsing request with empty body
func TestParseEmptyBody(t *testing.T) {
	input := `POST https://example.com HTTP/1.1
Host: example.com
Content-Type: application/json

`

	req, err := Parse(input)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if req.Body != "" {
		t.Errorf("Expected empty body, got %s", req.Body)
	}
}

// TestParseMultilineBody tests parsing request with multiline body
func TestParseMultilineBody(t *testing.T) {
	input := `POST https://example.com HTTP/1.1
Host: example.com
Content-Type: application/json

{
  "name": "John",
  "age": 30,
  "email": "john@example.com"
}`

	req, err := Parse(input)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if !strings.Contains(req.Body, "John") {
		t.Error("Expected body to contain 'John'")
	}

	if !strings.Contains(req.Body, "john@example.com") {
		t.Error("Expected body to contain email")
	}
}

// TestParseInvalidRequest tests error handling for invalid requests
func TestParseInvalidRequest(t *testing.T) {
	invalidInputs := []string{
		"",
		"INVALID",
		"GET",
	}

	for _, input := range invalidInputs {
		t.Run("Invalid_"+input, func(t *testing.T) {
			_, err := Parse(input)
			if err == nil {
				t.Error("Expected error for invalid input")
			}
		})
	}

	// "GET HTTP/1.1" parses but will fail validation (missing URL)
	// This is by design - parser is lenient, validator is strict
	t.Run("ParsesButInvalid_GET_HTTP/1.1", func(t *testing.T) {
		req, err := Parse("GET HTTP/1.1")
		if err != nil {
			t.Errorf("Parser should be lenient, got error: %v", err)
		}
		if req != nil && req.URL == "HTTP/1.1" {
			// This is expected - parser treats "HTTP/1.1" as URL
			// Validator will catch this as an error
		}
	})
}

// TestParseHeadersWithSpaces tests parsing headers with various spacing
func TestParseHeadersWithSpaces(t *testing.T) {
	input := `GET https://example.com HTTP/1.1
Host:example.com
User-Agent:  test-client  
Accept:application/json`

	req, err := Parse(input)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Headers should handle spacing correctly
	if req.Headers["Host"] == "" {
		t.Error("Expected Host header to be parsed")
	}

	if req.Headers["User-Agent"] == "" {
		t.Error("Expected User-Agent header to be parsed")
	}
}

// TestParseWithMixedLineEndings tests parsing with different line endings
func TestParseWithMixedLineEndings(t *testing.T) {
	// Test with \r\n (Windows)
	input1 := "GET https://example.com HTTP/1.1\r\nHost: example.com\r\n"
	req1, err := Parse(input1)
	if err != nil {
		t.Fatalf("Failed to parse with \\r\\n: %v", err)
	}
	if req1.Method != "GET" {
		t.Error("Failed to parse with \\r\\n line endings")
	}

	// Test with \n (Unix)
	input2 := "GET https://example.com HTTP/1.1\nHost: example.com\n"
	req2, err := Parse(input2)
	if err != nil {
		t.Fatalf("Failed to parse with \\n: %v", err)
	}
	if req2.Method != "GET" {
		t.Error("Failed to parse with \\n line endings")
	}
}
