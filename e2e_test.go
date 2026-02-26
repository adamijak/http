package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/adamijak/http/internal/testserver"
)

// TestMain sets up and tears down for all E2E tests
func TestMain(m *testing.M) {
	// Build the binary for testing
	buildCmd := exec.Command("go", "build", "-o", "http-e2e-test")
	if err := buildCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to build test binary: %v\n", err)
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	_ = os.Remove("http-e2e-test")

	os.Exit(code)
}

// TestE2ESimpleGET tests a simple GET request end-to-end
func TestE2ESimpleGET(t *testing.T) {
	// Create test server
	ts, err := testserver.NewWithConfig(&testserver.HandlerConfig{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       `{"status":"ok"}`,
	})
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer func() { _ = ts.Close() }()

	// Create test request
	host := strings.TrimPrefix(ts.URL, "http://")
	request := fmt.Sprintf("GET %s/api/test HTTP/1.1\nHost: %s\nUser-Agent: e2e-test\n", ts.URL, host)

	// Run CLI
	cmd := exec.Command("./http-e2e-test", "--no-color")
	cmd.Stdin = strings.NewReader(request)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("CLI failed: %v\nOutput: %s", err, output)
	}

	// Validate output contains expected response
	outputStr := string(output)
	if !strings.Contains(outputStr, "200") {
		t.Errorf("Expected status 200 in output, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "ok") {
		t.Errorf("Expected 'ok' in response body, got: %s", outputStr)
	}

	// Validate server received the request
	requests := ts.GetRequests()
	if len(requests) != 1 {
		t.Fatalf("Expected 1 request, got %d", len(requests))
	}

	if requests[0].Method != "GET" {
		t.Errorf("Expected method GET, got %s", requests[0].Method)
	}

	if requests[0].Path != "/api/test" {
		t.Errorf("Expected path /api/test, got %s", requests[0].Path)
	}
}

// TestE2EPOSTWithBody tests POST request with JSON body
func TestE2EPOSTWithBody(t *testing.T) {
	// Create test server that echoes back request info
	ts, err := testserver.NewWithConfig(&testserver.HandlerConfig{
		StatusCode: 201,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       `{"created":true}`,
	})
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer func() { _ = ts.Close() }()

	// Create POST request with body
	host := strings.TrimPrefix(ts.URL, "http://")
	request := fmt.Sprintf(`POST %s/users HTTP/1.1
Host: %s
Content-Type: application/json

{"name":"Alice","age":25}`, ts.URL, host)

	// Run CLI
	cmd := exec.Command("./http-e2e-test", "--no-color")
	cmd.Stdin = strings.NewReader(request)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("CLI failed: %v\nOutput: %s", err, output)
	}

	// Validate output
	outputStr := string(output)
	if !strings.Contains(outputStr, "201") {
		t.Errorf("Expected status 201 in output, got: %s", outputStr)
	}

	// Validate server received the request with body
	requests := ts.GetRequests()
	if len(requests) != 1 {
		t.Fatalf("Expected 1 request, got %d", len(requests))
	}

	if requests[0].Method != "POST" {
		t.Errorf("Expected method POST, got %s", requests[0].Method)
	}

	if !strings.Contains(requests[0].Body, "Alice") {
		t.Errorf("Expected body to contain 'Alice', got: %s", requests[0].Body)
	}

	if requests[0].Headers.Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json")
	}
}

// TestE2EWithEnvVars tests environment variable substitution
func TestE2EWithEnvVars(t *testing.T) {
	// Set test environment variables
	_ = os.Setenv("E2E_TEST_TOKEN", "secret-token-xyz")
	_ = os.Setenv("E2E_TEST_USER", "testuser")
	defer func() {
		_ = os.Unsetenv("E2E_TEST_TOKEN")
		_ = os.Unsetenv("E2E_TEST_USER")
	}()

	// Create test server
	ts, err := testserver.New()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer func() { _ = ts.Close() }()

	// Create request with environment variables
	host := strings.TrimPrefix(ts.URL, "http://")
	request := fmt.Sprintf(`GET %s/api HTTP/1.1
Host: %s
Authorization: Bearer ${E2E_TEST_TOKEN}
X-User: $E2E_TEST_USER
`, ts.URL, host)

	// Run CLI
	cmd := exec.Command("./http-e2e-test", "--no-color")
	cmd.Stdin = strings.NewReader(request)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("CLI failed: %v\nOutput: %s", err, output)
	}

	// Validate server received substituted headers
	requests := ts.GetRequests()
	if len(requests) != 1 {
		t.Fatalf("Expected 1 request, got %d", len(requests))
	}

	authHeader := requests[0].Headers.Get("Authorization")
	if !strings.Contains(authHeader, "secret-token-xyz") {
		t.Errorf("Expected Authorization header to contain substituted token, got: %s", authHeader)
	}

	userHeader := requests[0].Headers.Get("X-User")
	if userHeader != "testuser" {
		t.Errorf("Expected X-User header to be 'testuser', got: %s", userHeader)
	}
}

// TestE2ENoSendFlag tests the --no-send flag (output RFC format without sending)
func TestE2ENoSendFlag(t *testing.T) {
	// Create request
	request := `GET https://example.com/api HTTP/1.1
Host: example.com
User-Agent: test`

	// Run CLI with --no-send
	cmd := exec.Command("./http-e2e-test", "--no-send", "--no-color")
	cmd.Stdin = strings.NewReader(request)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("CLI failed: %v\nOutput: %s", err, output)
	}

	// Validate RFC format output (CRLF line endings)
	outputStr := string(output)
	if !strings.Contains(outputStr, "GET https://example.com/api HTTP/1.1") {
		t.Errorf("Expected RFC format output, got: %s", outputStr)
	}

	// Check for CRLF (should have \r\n)
	if !strings.Contains(outputStr, "\r\n") {
		t.Errorf("Expected CRLF line endings in RFC format output")
	}
}

// TestE2EFileInput tests reading request from file with -f flag
func TestE2EFileInput(t *testing.T) {
	// Create test server
	ts, err := testserver.New()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer func() { _ = ts.Close() }()

	// Create temporary test file
	tmpfile, err := os.CreateTemp("", "e2e-test-*.http")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpfile.Name()) }()

	// Write test request to file
	host := strings.TrimPrefix(ts.URL, "http://")
	request := fmt.Sprintf("GET %s/test HTTP/1.1\nHost: %s\n", ts.URL, host)
	if _, err := tmpfile.Write([]byte(request)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	_ = tmpfile.Close()

	// Run CLI with -f flag
	cmd := exec.Command("./http-e2e-test", "-f", tmpfile.Name(), "--no-color")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("CLI failed: %v\nOutput: %s", err, output)
	}

	// Validate server received the request
	requests := ts.GetRequests()
	if len(requests) != 1 {
		t.Fatalf("Expected 1 request, got %d", len(requests))
	}

	if requests[0].Path != "/test" {
		t.Errorf("Expected path /test, got %s", requests[0].Path)
	}
}

// TestE2EValidationError tests that validation errors are caught
func TestE2EValidationError(t *testing.T) {
	// Create invalid request (missing scheme)
	request := `GET example.com/api HTTP/1.1
Host: example.com`

	// Run CLI
	cmd := exec.Command("./http-e2e-test", "--no-send")
	cmd.Stdin = strings.NewReader(request)
	output, err := cmd.CombinedOutput()

	// Should fail with validation error
	if err == nil {
		t.Errorf("Expected CLI to fail with validation error, but it succeeded")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "URL must include scheme") {
		t.Errorf("Expected validation error message about URL scheme, got: %s", outputStr)
	}
}

// TestE2EStrictMode tests that strict mode enforces warnings
func TestE2EStrictMode(t *testing.T) {
	// Create request that generates a warning (POST without Content-Type)
	request := `POST https://example.com/api HTTP/1.1
Host: example.com

test body`

	// Run CLI with --strict
	cmd := exec.Command("./http-e2e-test", "--strict", "--no-send")
	cmd.Stdin = strings.NewReader(request)
	output, err := cmd.CombinedOutput()

	// Should fail in strict mode
	if err == nil {
		t.Errorf("Expected CLI to fail in strict mode, but it succeeded")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Strict mode") {
		t.Errorf("Expected strict mode message, got: %s", outputStr)
	}
}

// TestE2EMultipleMethods tests different HTTP methods
func TestE2EMultipleMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			// Create test server
			ts, err := testserver.New()
			if err != nil {
				t.Fatalf("Failed to create test server: %v", err)
			}
			defer func() { _ = ts.Close() }()

			// Create request
			host := strings.TrimPrefix(ts.URL, "http://")
			request := fmt.Sprintf("%s %s/test HTTP/1.1\nHost: %s\n", method, ts.URL, host)

			// Run CLI
			cmd := exec.Command("./http-e2e-test", "--no-color")
			cmd.Stdin = strings.NewReader(request)
			output, err := cmd.CombinedOutput()

			if err != nil {
				t.Fatalf("CLI failed for %s: %v\nOutput: %s", method, err, output)
			}

			// Validate server received correct method
			requests := ts.GetRequests()
			if len(requests) != 1 {
				t.Fatalf("Expected 1 request, got %d", len(requests))
			}

			if requests[0].Method != method {
				t.Errorf("Expected method %s, got %s", method, requests[0].Method)
			}
		})
	}
}

// TestE2ECustomHeaders tests that custom headers are sent correctly
func TestE2ECustomHeaders(t *testing.T) {
	// Create test server
	ts, err := testserver.New()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer func() { _ = ts.Close() }()

	// Create request with custom headers
	host := strings.TrimPrefix(ts.URL, "http://")
	request := fmt.Sprintf(`GET %s/api HTTP/1.1
Host: %s
X-Custom-Header: custom-value
X-Request-ID: req-12345
Accept: application/json
Authorization: Bearer test-token
`, ts.URL, host)

	// Run CLI
	cmd := exec.Command("./http-e2e-test", "--no-color")
	cmd.Stdin = strings.NewReader(request)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("CLI failed: %v\nOutput: %s", err, output)
	}

	// Validate server received all custom headers
	requests := ts.GetRequests()
	if len(requests) != 1 {
		t.Fatalf("Expected 1 request, got %d", len(requests))
	}

	headers := requests[0].Headers
	if headers.Get("X-Custom-Header") != "custom-value" {
		t.Errorf("Expected X-Custom-Header to be 'custom-value'")
	}

	if headers.Get("X-Request-ID") != "req-12345" {
		t.Errorf("Expected X-Request-ID to be 'req-12345'")
	}

	if headers.Get("Accept") != "application/json" {
		t.Errorf("Expected Accept to be 'application/json'")
	}

	if !strings.Contains(headers.Get("Authorization"), "test-token") {
		t.Errorf("Expected Authorization header with test-token")
	}
}

// TestE2ECommentHandling tests that comments are properly stripped
func TestE2ECommentHandling(t *testing.T) {
	// Create test server
	ts, err := testserver.New()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer func() { _ = ts.Close() }()

	// Create request with comments
	host := strings.TrimPrefix(ts.URL, "http://")
	request := fmt.Sprintf(`# This is a comment
GET %s/api HTTP/1.1
// Another comment
Host: %s
# Header comment
User-Agent: test
`, ts.URL, host)

	// Run CLI
	cmd := exec.Command("./http-e2e-test", "--no-color")
	cmd.Stdin = strings.NewReader(request)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("CLI failed: %v\nOutput: %s", err, output)
	}

	// Validate server received clean request (no comments)
	requests := ts.GetRequests()
	if len(requests) != 1 {
		t.Fatalf("Expected 1 request, got %d", len(requests))
	}

	// Verify no comment characters in headers
	for key, values := range requests[0].Headers {
		for _, value := range values {
			if strings.Contains(value, "#") || strings.Contains(value, "//") {
				t.Errorf("Found comment character in header %s: %s", key, value)
			}
		}
	}
}

// TestE2EResponseStatusCodes tests handling of different response status codes
func TestE2EResponseStatusCodes(t *testing.T) {
	testCases := []struct {
		name       string
		statusCode int
		expectFail bool
	}{
		{"OK", 200, false},
		{"Created", 201, false},
		{"No Content", 204, false},
		{"Bad Request", 400, false}, // Client gets response, not a CLI error
		{"Unauthorized", 401, false},
		{"Not Found", 404, false},
		{"Internal Server Error", 500, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test server with specific status code
			ts, err := testserver.NewWithConfig(&testserver.HandlerConfig{
				StatusCode: tc.statusCode,
				Body:       fmt.Sprintf("Status: %d", tc.statusCode),
			})
			if err != nil {
				t.Fatalf("Failed to create test server: %v", err)
			}
			defer func() { _ = ts.Close() }()

			// Create request
			host := strings.TrimPrefix(ts.URL, "http://")
			request := fmt.Sprintf("GET %s/test HTTP/1.1\nHost: %s\n", ts.URL, host)

			// Run CLI
			cmd := exec.Command("./http-e2e-test", "--no-color")
			cmd.Stdin = strings.NewReader(request)
			output, err := cmd.CombinedOutput()

			if tc.expectFail && err == nil {
				t.Errorf("Expected CLI to fail, but it succeeded")
			}

			if !tc.expectFail && err != nil {
				t.Errorf("Expected CLI to succeed, but it failed: %v", err)
			}

			// Validate status code in output
			outputStr := string(output)
			statusStr := fmt.Sprintf("%d", tc.statusCode)
			if !strings.Contains(outputStr, statusStr) {
				t.Errorf("Expected status code %d in output, got: %s", tc.statusCode, outputStr)
			}
		})
	}
}

// TestE2EConnectionRefused tests handling of connection errors
func TestE2EConnectionRefused(t *testing.T) {
	// Create request to unlikely port
	request := `GET http://127.0.0.1:59999/test HTTP/1.1
Host: 127.0.0.1:59999
`

	// Run CLI
	cmd := exec.Command("./http-e2e-test", "--no-color")
	cmd.Stdin = strings.NewReader(request)
	cmd.Env = os.Environ() // Inherit environment
	output, err := cmd.CombinedOutput()

	// Should fail with connection error
	if err == nil {
		t.Errorf("Expected CLI to fail with connection error, but it succeeded")
	}

	outputStr := string(output)
	if !strings.Contains(strings.ToLower(outputStr), "connection") {
		t.Errorf("Expected connection error message, got: %s", outputStr)
	}
}

// TestE2EPortOverride tests the --port flag
func TestE2EPortOverride(t *testing.T) {
	// Create test server
	ts, err := testserver.New()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer func() { _ = ts.Close() }()

	// Extract port from test server URL
	parts := strings.Split(ts.URL, ":")
	port := parts[len(parts)-1]

	// Create request without port in URL (will use --port flag)
	request := `GET http://127.0.0.1/test HTTP/1.1
Host: 127.0.0.1
`

	// Run CLI with --port flag
	cmd := exec.Command("./http-e2e-test", "--port", port, "--no-color")
	cmd.Stdin = strings.NewReader(request)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("CLI failed: %v\nOutput: %s", err, output)
	}

	// Validate server received the request
	requests := ts.GetRequests()
	if len(requests) != 1 {
		t.Fatalf("Expected 1 request, got %d", len(requests))
	}
}

// TestE2EPipelineFlow tests that output can be piped back to the tool
func TestE2EPipelineFlow(t *testing.T) {
	// Create test server
	ts, err := testserver.New()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer func() { _ = ts.Close() }()

	// Create HTP format request
	host := strings.TrimPrefix(ts.URL, "http://")
	htpRequest := fmt.Sprintf("# Comment\nGET %s/test HTTP/1.1\nHost: %s\n", ts.URL, host)

	// First pass: Convert to RFC format
	cmd1 := exec.Command("./http-e2e-test", "--no-send", "--no-color")
	cmd1.Stdin = strings.NewReader(htpRequest)
	rfcOutput, err := cmd1.CombinedOutput()
	if err != nil {
		t.Fatalf("First pass failed: %v\nOutput: %s", err, rfcOutput)
	}

	// Give server time to be ready
	time.Sleep(50 * time.Millisecond)

	// Second pass: Send RFC format
	cmd2 := exec.Command("./http-e2e-test", "--no-color")
	cmd2.Stdin = strings.NewReader(string(rfcOutput))
	finalOutput, err := cmd2.CombinedOutput()
	if err != nil {
		t.Fatalf("Second pass failed: %v\nOutput: %s", err, finalOutput)
	}

	// Validate server received the request
	requests := ts.GetRequests()
	if len(requests) != 1 {
		t.Fatalf("Expected 1 request, got %d", len(requests))
	}

	if requests[0].Path != "/test" {
		t.Errorf("Expected path /test, got %s", requests[0].Path)
	}
}

// TestE2ERealWorldExample tests a realistic API request scenario
func TestE2ERealWorldExample(t *testing.T) {
	// Create test server that simulates a REST API
	ts, err := testserver.NewWithConfig(&testserver.HandlerConfig{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type":   "application/json",
			"X-Rate-Limit":   "100",
			"X-Request-Time": "0.042s",
		},
		Body: `{"users":[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}],"total":2}`,
	})
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer func() { _ = ts.Close() }()

	// Create realistic API request
	host := strings.TrimPrefix(ts.URL, "http://")
	request := fmt.Sprintf(`GET %s/api/v1/users?limit=10&offset=0 HTTP/1.1
Host: %s
Accept: application/json
Authorization: Bearer api-key-12345
User-Agent: MyApp/1.0
X-Client-Version: 1.0.0
`, ts.URL, host)

	// Run CLI
	cmd := exec.Command("./http-e2e-test", "--no-color")
	cmd.Stdin = strings.NewReader(request)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("CLI failed: %v\nOutput: %s", err, output)
	}

	// Validate response
	outputStr := string(output)
	if !strings.Contains(outputStr, "200") {
		t.Errorf("Expected status 200")
	}

	if !strings.Contains(outputStr, "Alice") || !strings.Contains(outputStr, "Bob") {
		t.Errorf("Expected users in response")
	}

	// Validate request headers
	requests := ts.GetRequests()
	if len(requests) != 1 {
		t.Fatalf("Expected 1 request, got %d", len(requests))
	}

	if !strings.Contains(requests[0].Headers.Get("Authorization"), "api-key-12345") {
		t.Errorf("Expected Authorization header with API key")
	}
}

// TestE2EExampleFiles tests that example files work correctly
func TestE2EExampleFiles(t *testing.T) {
	exampleFiles := []string{
		"examples/simple-get.http",
		"examples/post-json.http",
	}

	for _, file := range exampleFiles {
		t.Run(filepath.Base(file), func(t *testing.T) {
			// Test that example parses correctly with --no-send
			cmd := exec.Command("./http-e2e-test", "-f", file, "--no-send", "--no-color")
			output, err := cmd.CombinedOutput()

			if err != nil {
				t.Errorf("Example file %s failed: %v\nOutput: %s", file, err, output)
			}

			// Validate RFC format output
			outputStr := string(output)
			if !strings.Contains(outputStr, "HTTP/1.1") {
				t.Errorf("Expected RFC format output for %s", file)
			}
		})
	}
}
