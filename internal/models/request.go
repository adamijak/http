package models

import (
	"fmt"
	"io"
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
		fmt.Fprintf(w, "\033[1;36m%s %s %s\033[0m\n", r.Method, r.URL, r.Version)
		for key, value := range r.Headers {
			fmt.Fprintf(w, "\033[0;33m%s:\033[0m %s\n", key, value)
		}
		if r.Body != "" {
			fmt.Fprintf(w, "\n\033[0;32m%s\033[0m\n", r.Body)
		}
	} else {
		// Plain output
		fmt.Fprintf(w, "%s %s %s\n", r.Method, r.URL, r.Version)
		for key, value := range r.Headers {
			fmt.Fprintf(w, "%s: %s\n", key, value)
		}
		if r.Body != "" {
			fmt.Fprintf(w, "\n%s\n", r.Body)
		}
	}
}

// ToRawRequest converts the HTTPRequest to raw HTTP format for sending over TCP
func (r *HTTPRequest) ToRawRequest() string {
	var sb strings.Builder
	
	// Request line
	sb.WriteString(fmt.Sprintf("%s %s %s\r\n", r.Method, r.URL, r.Version))
	
	// Headers
	for key, value := range r.Headers {
		sb.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}
	
	// Empty line between headers and body
	sb.WriteString("\r\n")
	
	// Body
	if r.Body != "" {
		sb.WriteString(r.Body)
	}
	
	return sb.String()
}
