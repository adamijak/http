package models

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
)

// HTTPRequest represents a parsed HTTP request
// This structure is designed to be easily understood and modified by AI agents
type HTTPRequest struct {
	Method  string            // HTTP method (GET, POST, etc.)
	URL     string            // Full URL including scheme
	Version string            // HTTP version (HTTP/1.1, HTTP/2, etc.)
	Headers map[string]string // Request headers
	Body    string            // Request body
}

// NewHTTPRequest creates a new HTTPRequest with default values
func NewHTTPRequest() *HTTPRequest {
	return &HTTPRequest{
		Method:  "GET",
		Version: "HTTP/1.1",
		Headers: make(map[string]string),
	}
}

// Print outputs the request in a readable format
func (r *HTTPRequest) Print(w io.Writer, colored bool) {
	if colored {
		// Colored output
		_, _ = fmt.Fprintf(w, "\033[1;36m%s %s %s\033[0m\n", r.Method, r.URL, r.Version)
		for key, value := range r.Headers {
			_, _ = fmt.Fprintf(w, "\033[0;33m%s:\033[0m %s\n", key, value)
		}
		if r.Body != "" {
			_, _ = fmt.Fprintf(w, "\n\033[0;32m%s\033[0m\n", r.Body)
		}
	} else {
		// Plain output
		_, _ = fmt.Fprintf(w, "%s %s %s\n", r.Method, r.URL, r.Version)
		for key, value := range r.Headers {
			_, _ = fmt.Fprintf(w, "%s: %s\n", key, value)
		}
		if r.Body != "" {
			_, _ = fmt.Fprintf(w, "\n%s\n", r.Body)
		}
	}
}

// ToRawRequest converts the HTTPRequest to raw HTTP format for sending over TCP
// This returns RFC compliant format with only the path (not the full URL) in the request line
func (r *HTTPRequest) ToRawRequest() string {
	var sb strings.Builder

	// Extract path from URL for RFC compliant format
	requestTarget := r.URL
	if parsedURL, err := url.Parse(r.URL); err == nil {
		// Use path and query from URL
		requestTarget = parsedURL.Path
		if requestTarget == "" {
			requestTarget = "/"
		}
		if parsedURL.RawQuery != "" {
			requestTarget += "?" + parsedURL.RawQuery
		}
	}

	// Request line (RFC compliant: METHOD PATH VERSION)
	fmt.Fprintf(&sb, "%s %s %s\r\n", r.Method, requestTarget, r.Version)

	// Headers
	for key, value := range r.Headers {
		fmt.Fprintf(&sb, "%s: %s\r\n", key, value)
	}

	// Empty line between headers and body
	sb.WriteString("\r\n")

	// Body
	if r.Body != "" {
		sb.WriteString(r.Body)
	}

	return sb.String()
}

// SaveToFile saves the RFC compliant HTTP request to a file
func (r *HTTPRequest) SaveToFile(filename string) error {
	content := r.ToRawRequest()
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}
