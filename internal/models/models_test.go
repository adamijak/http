package models

import (
	"bytes"
	"strings"
	"testing"
)

// TestNewHTTPRequest tests creating a new HTTPRequest
func TestNewHTTPRequest(t *testing.T) {
	req := NewHTTPRequest()

	if req.Method != "GET" {
		t.Errorf("Expected default method GET, got %s", req.Method)
	}

	if req.Version != "HTTP/1.1" {
		t.Errorf("Expected default version HTTP/1.1, got %s", req.Version)
	}

	if req.Headers == nil {
		t.Error("Expected headers map to be initialized")
	}

	if len(req.Headers) != 0 {
		t.Error("Expected empty headers map")
	}
}

// TestHTTPRequestPrint tests printing HTTP request
func TestHTTPRequestPrint(t *testing.T) {
	req := &HTTPRequest{
		Method:  "GET",
		URL:     "https://example.com/api",
		Version: "HTTP/1.1",
		Headers: map[string]string{
			"Host":       "example.com",
			"User-Agent": "test-client",
		},
		Body: "",
	}

	var buf bytes.Buffer
	req.Print(&buf, false)

	output := buf.String()

	if !strings.Contains(output, "GET") {
		t.Error("Expected output to contain method")
	}

	if !strings.Contains(output, "https://example.com/api") {
		t.Error("Expected output to contain URL")
	}

	if !strings.Contains(output, "Host: example.com") {
		t.Error("Expected output to contain Host header")
	}
}

// TestHTTPRequestPrintColored tests printing HTTP request with colors
func TestHTTPRequestPrintColored(t *testing.T) {
	req := &HTTPRequest{
		Method:  "POST",
		URL:     "https://example.com/api",
		Version: "HTTP/1.1",
		Headers: map[string]string{
			"Host":         "example.com",
			"Content-Type": "application/json",
		},
		Body: `{"test":"data"}`,
	}

	var buf bytes.Buffer
	req.Print(&buf, true)

	output := buf.String()

	// Should contain ANSI color codes
	if !strings.Contains(output, "\033[") {
		t.Error("Expected output to contain ANSI color codes")
	}

	if !strings.Contains(output, "POST") {
		t.Error("Expected output to contain method")
	}

	if !strings.Contains(output, `{"test":"data"}`) {
		t.Error("Expected output to contain body")
	}
}

// TestHTTPRequestToRawRequest tests converting to raw HTTP format
func TestHTTPRequestToRawRequest(t *testing.T) {
	req := &HTTPRequest{
		Method:  "GET",
		URL:     "https://example.com/api",
		Version: "HTTP/1.1",
		Headers: map[string]string{
			"Host":       "example.com",
			"User-Agent": "test-client",
		},
		Body: "",
	}

	raw := req.ToRawRequest()

	// Check request line
	if !strings.Contains(raw, "GET https://example.com/api HTTP/1.1") {
		t.Error("Expected raw request to contain request line")
	}

	// Check headers
	if !strings.Contains(raw, "Host: example.com") {
		t.Error("Expected raw request to contain Host header")
	}

	if !strings.Contains(raw, "User-Agent: test-client") {
		t.Error("Expected raw request to contain User-Agent header")
	}

	// Check CRLF line endings
	if !strings.Contains(raw, "\r\n") {
		t.Error("Expected raw request to use CRLF line endings")
	}
}

// TestHTTPRequestToRawRequestWithBody tests converting POST with body to raw format
func TestHTTPRequestToRawRequestWithBody(t *testing.T) {
	req := &HTTPRequest{
		Method:  "POST",
		URL:     "https://example.com/api",
		Version: "HTTP/1.1",
		Headers: map[string]string{
			"Host":           "example.com",
			"Content-Type":   "application/json",
			"Content-Length": "15",
		},
		Body: `{"name":"test"}`,
	}

	raw := req.ToRawRequest()

	// Check body is included
	if !strings.Contains(raw, `{"name":"test"}`) {
		t.Error("Expected raw request to contain body")
	}

	// Check empty line between headers and body
	if !strings.Contains(raw, "\r\n\r\n") {
		t.Error("Expected empty line between headers and body")
	}
}

// TestHTTPResponsePrint tests printing HTTP response
func TestHTTPResponsePrint(t *testing.T) {
	resp := &HTTPResponse{
		Version:    "HTTP/1.1",
		StatusCode: 200,
		Status:     "OK",
		Headers: map[string]string{
			"Content-Type":   "application/json",
			"Content-Length": "15",
		},
		Body: `{"status":"ok"}`,
	}

	var buf bytes.Buffer
	resp.Print(&buf, false)

	output := buf.String()

	if !strings.Contains(output, "HTTP/1.1") {
		t.Error("Expected output to contain HTTP version")
	}

	if !strings.Contains(output, "200") {
		t.Error("Expected output to contain status code")
	}

	if !strings.Contains(output, "OK") {
		t.Error("Expected output to contain status text")
	}

	if !strings.Contains(output, `{"status":"ok"}`) {
		t.Error("Expected output to contain body")
	}
}

// TestHTTPResponsePrintColored tests printing response with colors
func TestHTTPResponsePrintColored(t *testing.T) {
	// Test success status (200)
	resp200 := &HTTPResponse{
		Version:    "HTTP/1.1",
		StatusCode: 200,
		Status:     "OK",
		Headers:    map[string]string{},
		Body:       "success",
	}

	var buf200 bytes.Buffer
	resp200.Print(&buf200, true)
	output200 := buf200.String()

	if !strings.Contains(output200, "\033[") {
		t.Error("Expected colored output for 200 status")
	}

	// Test redirect status (301)
	resp301 := &HTTPResponse{
		Version:    "HTTP/1.1",
		StatusCode: 301,
		Status:     "Moved Permanently",
		Headers:    map[string]string{},
		Body:       "",
	}

	var buf301 bytes.Buffer
	resp301.Print(&buf301, true)
	output301 := buf301.String()

	if !strings.Contains(output301, "\033[") {
		t.Error("Expected colored output for 301 status")
	}

	// Test error status (404)
	resp404 := &HTTPResponse{
		Version:    "HTTP/1.1",
		StatusCode: 404,
		Status:     "Not Found",
		Headers:    map[string]string{},
		Body:       "",
	}

	var buf404 bytes.Buffer
	resp404.Print(&buf404, true)
	output404 := buf404.String()

	if !strings.Contains(output404, "\033[") {
		t.Error("Expected colored output for 404 status")
	}
}

// TestHTTPResponseStatusCodes tests different status codes
func TestHTTPResponseStatusCodes(t *testing.T) {
	statusCodes := map[int]string{
		200: "OK",
		201: "Created",
		204: "No Content",
		301: "Moved Permanently",
		302: "Found",
		400: "Bad Request",
		401: "Unauthorized",
		403: "Forbidden",
		404: "Not Found",
		500: "Internal Server Error",
	}

	for code, status := range statusCodes {
		t.Run("Status_"+status, func(t *testing.T) {
			resp := &HTTPResponse{
				Version:    "HTTP/1.1",
				StatusCode: code,
				Status:     status,
				Headers:    map[string]string{},
			}

			var buf bytes.Buffer
			resp.Print(&buf, false)
			output := buf.String()

			if !strings.Contains(output, status) {
				t.Errorf("Expected output to contain status text: %s", status)
			}
		})
	}
}

// TestHTTPRequestEmptyBody tests request without body
func TestHTTPRequestEmptyBody(t *testing.T) {
	req := &HTTPRequest{
		Method:  "GET",
		URL:     "https://example.com",
		Version: "HTTP/1.1",
		Headers: map[string]string{"Host": "example.com"},
		Body:    "",
	}

	var buf bytes.Buffer
	req.Print(&buf, false)
	output := buf.String()

	// Should not have extra blank lines for empty body
	lines := strings.Split(output, "\n")
	if len(lines) > 5 {
		t.Error("Expected compact output for request without body")
	}
}

// TestHTTPResponseEmptyBody tests response without body
func TestHTTPResponseEmptyBody(t *testing.T) {
	resp := &HTTPResponse{
		Version:    "HTTP/1.1",
		StatusCode: 204,
		Status:     "No Content",
		Headers:    map[string]string{},
		Body:       "",
	}

	var buf bytes.Buffer
	resp.Print(&buf, false)
	output := buf.String()

	if !strings.Contains(output, "204") {
		t.Error("Expected output to contain status code")
	}
}

// TestHTTPRequestMultipleHeaders tests request with many headers
func TestHTTPRequestMultipleHeaders(t *testing.T) {
	req := &HTTPRequest{
		Method:  "GET",
		URL:     "https://example.com",
		Version: "HTTP/1.1",
		Headers: map[string]string{
			"Host":          "example.com",
			"User-Agent":    "test-client",
			"Accept":        "application/json",
			"Authorization": "Bearer token",
			"X-Custom":      "value",
		},
	}

	var buf bytes.Buffer
	req.Print(&buf, false)
	output := buf.String()

	// Check all headers are present
	expectedHeaders := []string{"Host:", "User-Agent:", "Accept:", "Authorization:", "X-Custom:"}
	for _, header := range expectedHeaders {
		if !strings.Contains(output, header) {
			t.Errorf("Expected output to contain header: %s", header)
		}
	}
}
