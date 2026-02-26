package validator

import (
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/adamijak/http/internal/models"
)

// Supported HTTP versions for validation
// AI Agent Note: Add new versions here as they become standardized
var supportedHTTPVersions = []string{"HTTP/1.0", "HTTP/1.1", "HTTP/2", "HTTP/2.0", "HTTP/3"}

// ValidationResult holds validation errors and warnings
// AI Agent Note: Clear separation between errors (must fix) and warnings (should fix)
type ValidationResult struct {
	Errors   []string
	Warnings []string
}

// HasErrors returns true if there are any errors
func (v *ValidationResult) HasErrors() bool {
	return len(v.Errors) > 0
}

// Print outputs validation results without colors
func (v *ValidationResult) Print(w io.Writer) {
	if len(v.Errors) > 0 {
		_, _ = fmt.Fprintln(w, "Validation Errors:")
		for _, err := range v.Errors {
			_, _ = fmt.Fprintf(w, "  [ERROR] %s\n", err)
		}
	}

	if len(v.Warnings) > 0 {
		_, _ = fmt.Fprintln(w, "Validation Warnings:")
		for _, warn := range v.Warnings {
			_, _ = fmt.Fprintf(w, "  [WARN] %s\n", warn)
		}
	}

	if len(v.Errors) == 0 && len(v.Warnings) == 0 {
		_, _ = fmt.Fprintln(w, "✓ Validation passed")
	}
}

// PrintColored outputs validation results with ANSI colors
func (v *ValidationResult) PrintColored(w io.Writer) {
	if len(v.Errors) > 0 {
		_, _ = fmt.Fprintln(w, "\033[1;31mValidation Errors:\033[0m")
		for _, err := range v.Errors {
			_, _ = fmt.Fprintf(w, "  \033[0;31m[ERROR]\033[0m %s\n", err)
		}
	}

	if len(v.Warnings) > 0 {
		_, _ = fmt.Fprintln(w, "\033[1;33mValidation Warnings:\033[0m")
		for _, warn := range v.Warnings {
			_, _ = fmt.Fprintf(w, "  \033[0;33m[WARN]\033[0m %s\n", warn)
		}
	}

	if len(v.Errors) == 0 && len(v.Warnings) == 0 {
		_, _ = fmt.Fprintln(w, "\033[0;32m✓ Validation passed\033[0m")
	}
}

// Validate validates an HTTP request against RFC standards
//
// AI Agent Note: This function checks common HTTP standards.
// Each check is isolated and can be easily modified or extended.
//
// Checks performed:
// - Valid HTTP method
// - Valid URL format
// - Valid HTTP version
// - Required headers for certain methods
// - Content-Length for requests with body
func Validate(req *models.HTTPRequest, noSecure bool) *ValidationResult {
	result := &ValidationResult{
		Errors:   []string{},
		Warnings: []string{},
	}

	// Validate HTTP method
	validateMethod(req, result)

	// Validate URL
	validateURL(req, result, noSecure)

	// Validate HTTP version
	validateVersion(req, result)

	// Validate headers
	validateHeaders(req, result)

	// Validate body and Content-Length
	validateBody(req, result)

	return result
}

// validateMethod checks if the HTTP method is valid
func validateMethod(req *models.HTTPRequest, result *ValidationResult) {
	validMethods := []string{
		"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS", "CONNECT", "TRACE",
	}

	method := strings.ToUpper(req.Method)
	req.Method = method // Normalize to uppercase

	valid := false
	for _, m := range validMethods {
		if method == m {
			valid = true
			break
		}
	}

	if !valid {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("Non-standard HTTP method: %s", method))
	}
}

// validateURL checks if the URL is valid and has a scheme
// Note: This function may modify req.URL to construct a full URL from a path
func validateURL(req *models.HTTPRequest, result *ValidationResult, noSecure bool) {
	if req.URL == "" {
		result.Errors = append(result.Errors, "URL is required")
		return
	}

	// Check if URL starts with a path (e.g., /path or /path?query)
	// If so, reconstruct full URL from Host header
	if strings.HasPrefix(req.URL, "/") {
		// URL is a path, need Host header to construct full URL
		host, hasHost := req.Headers["Host"]
		if !hasHost {
			result.Errors = append(result.Errors,
				"Host header is required when URL is a path (e.g., /path)")
			return
		}

		// Determine scheme based on --no-secure flag
		scheme := "https"
		if noSecure {
			scheme = "http"
		}

		// Reconstruct full URL
		req.URL = fmt.Sprintf("%s://%s%s", scheme, host, req.URL)
		return
	}

	parsedURL, err := url.Parse(req.URL)
	if err != nil {
		result.Errors = append(result.Errors,
			fmt.Sprintf("Invalid URL: %v", err))
		return
	}

	// Check if scheme is present
	if parsedURL.Scheme == "" {
		result.Errors = append(result.Errors,
			"URL must include scheme (http:// or https://) or be a path starting with /")
		return
	}

	// Apply --no-secure flag to force HTTP
	if noSecure && parsedURL.Scheme == "https" {
		parsedURL.Scheme = "http"
		req.URL = parsedURL.String()
	}

	// Check if scheme is http or https
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("Non-standard URL scheme: %s", parsedURL.Scheme))
	}

	// Check if host is present
	if parsedURL.Host == "" {
		result.Errors = append(result.Errors, "URL must include host")
	}
}

// validateVersion checks if the HTTP version is valid
func validateVersion(req *models.HTTPRequest, result *ValidationResult) {
	valid := false
	for _, v := range supportedHTTPVersions {
		if req.Version == v {
			valid = true
			break
		}
	}

	if !valid {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("Non-standard HTTP version: %s (using anyway)", req.Version))
	}
}

// validateHeaders checks for required headers and common issues
func validateHeaders(req *models.HTTPRequest, result *ValidationResult) {
	// Check for Host header (required in HTTP/1.1)
	if req.Version == "HTTP/1.1" {
		if _, ok := req.Headers["Host"]; !ok {
			// Try to add Host from URL
			parsedURL, err := url.Parse(req.URL)
			if err == nil && parsedURL.Host != "" {
				req.Headers["Host"] = parsedURL.Host
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("Added missing Host header: %s", parsedURL.Host))
			} else {
				result.Errors = append(result.Errors,
					"Host header is required for HTTP/1.1")
			}
		}
	}

	// Check for duplicate headers (case-insensitive)
	headerKeys := make(map[string]bool)
	for key := range req.Headers {
		lowerKey := strings.ToLower(key)
		if headerKeys[lowerKey] {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Duplicate header (case-insensitive): %s", key))
		}
		headerKeys[lowerKey] = true
	}
}

// validateBody checks body-related requirements
func validateBody(req *models.HTTPRequest, result *ValidationResult) {
	hasBody := req.Body != ""

	// Methods that typically shouldn't have a body
	noBodyMethods := []string{"GET", "HEAD", "DELETE", "CONNECT", "TRACE"}
	for _, method := range noBodyMethods {
		if req.Method == method && hasBody {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("%s requests typically should not have a body", req.Method))
			break
		}
	}

	// Check Content-Length header when body is present
	if hasBody {
		if _, ok := req.Headers["Content-Length"]; !ok {
			// Auto-add Content-Length
			req.Headers["Content-Length"] = fmt.Sprintf("%d", len(req.Body))
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Added missing Content-Length header: %d", len(req.Body)))
		}
	}

	// Check for Content-Type when body is present
	if hasBody {
		if _, ok := req.Headers["Content-Type"]; !ok {
			result.Warnings = append(result.Warnings,
				"Content-Type header is recommended when sending a body")
		}
	}
}
