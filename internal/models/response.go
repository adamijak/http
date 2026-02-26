package models

import (
	"fmt"
	"io"
)

// HTTPResponse represents an HTTP response
type HTTPResponse struct {
	Version    string            // HTTP version
	StatusCode int               // Status code
	Status     string            // Status text
	Headers    map[string]string // Response headers
	Body       string            // Response body
}

// Print outputs the response in a readable format
func (r *HTTPResponse) Print(w io.Writer, colored bool) {
	if colored {
		// Colored output - green for success, red for errors
		statusColor := "\033[0;32m" // green
		if r.StatusCode >= 400 {
			statusColor = "\033[0;31m" // red
		} else if r.StatusCode >= 300 {
			statusColor = "\033[0;33m" // yellow
		}

		fmt.Fprintf(w, "\033[1;36m%s\033[0m %s%d %s\033[0m\n", r.Version, statusColor, r.StatusCode, r.Status)
		for key, value := range r.Headers {
			fmt.Fprintf(w, "\033[0;33m%s:\033[0m %s\n", key, value)
		}
		if r.Body != "" {
			fmt.Fprintf(w, "\n%s\n", r.Body)
		}
	} else {
		// Plain output
		fmt.Fprintf(w, "%s %d %s\n", r.Version, r.StatusCode, r.Status)
		for key, value := range r.Headers {
			fmt.Fprintf(w, "%s: %s\n", key, value)
		}
		if r.Body != "" {
			fmt.Fprintf(w, "\n%s\n", r.Body)
		}
	}
}
