package validator

import (
	"strings"
	"testing"

	"github.com/adamijak/http/internal/models"
)

// TestValidateValidRequest tests validation of a valid request
func TestValidateValidRequest(t *testing.T) {
	req := &models.HTTPRequest{
		Method:  "GET",
		URL:     "https://example.com/api",
		Version: "HTTP/1.1",
		Headers: map[string]string{
			"Host": "example.com",
		},
	}

	result := Validate(req, false)

	if result.HasErrors() {
		t.Errorf("Expected no errors, got: %v", result.Errors)
	}
}

// TestValidateStandardMethods tests validation of standard HTTP methods
func TestValidateStandardMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS", "CONNECT", "TRACE"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := &models.HTTPRequest{
				Method:  method,
				URL:     "https://example.com",
				Version: "HTTP/1.1",
				Headers: map[string]string{"Host": "example.com"},
			}

			result := Validate(req, false)

			if result.HasErrors() {
				t.Errorf("Expected no errors for %s method, got: %v", method, result.Errors)
			}
		})
	}
}

// TestValidateNonStandardMethod tests validation warning for non-standard methods
func TestValidateNonStandardMethod(t *testing.T) {
	req := &models.HTTPRequest{
		Method:  "CUSTOM",
		URL:     "https://example.com",
		Version: "HTTP/1.1",
		Headers: map[string]string{"Host": "example.com"},
	}

	result := Validate(req, false)

	if len(result.Warnings) == 0 {
		t.Error("Expected warning for non-standard method")
	}

	hasMethodWarning := false
	for _, warning := range result.Warnings {
		if strings.Contains(strings.ToLower(warning), "method") {
			hasMethodWarning = true
			break
		}
	}

	if !hasMethodWarning {
		t.Error("Expected warning about non-standard method")
	}
}

// TestValidateURLWithoutScheme tests validation error for URL without scheme
func TestValidateURLWithoutScheme(t *testing.T) {
	req := &models.HTTPRequest{
		Method:  "GET",
		URL:     "example.com/api",
		Version: "HTTP/1.1",
		Headers: map[string]string{},
	}

	result := Validate(req, false)

	if !result.HasErrors() {
		t.Error("Expected error for URL without scheme")
	}
}

// TestValidatePathOnlyURL tests validation of path-only URLs
func TestValidatePathOnlyURL(t *testing.T) {
	req := &models.HTTPRequest{
		Method:  "GET",
		URL:     "/api/users",
		Version: "HTTP/1.1",
		Headers: map[string]string{
			"Host": "example.com",
		},
	}

	result := Validate(req, false)

	if result.HasErrors() {
		t.Errorf("Expected no errors for path-only URL with Host header, got: %v", result.Errors)
	}
}

// TestValidatePathOnlyURLWithoutHost tests error for path-only URL without Host header
func TestValidatePathOnlyURLWithoutHost(t *testing.T) {
	req := &models.HTTPRequest{
		Method:  "GET",
		URL:     "/api/users",
		Version: "HTTP/1.1",
		Headers: map[string]string{},
	}

	result := Validate(req, false)

	if !result.HasErrors() {
		t.Error("Expected error for path-only URL without Host header")
	}
}

// TestValidateHTTPVersions tests validation of different HTTP versions
func TestValidateHTTPVersions(t *testing.T) {
	validVersions := []string{"HTTP/1.0", "HTTP/1.1", "HTTP/2", "HTTP/2.0", "HTTP/3"}

	for _, version := range validVersions {
		t.Run(version, func(t *testing.T) {
			req := &models.HTTPRequest{
				Method:  "GET",
				URL:     "https://example.com",
				Version: version,
				Headers: map[string]string{"Host": "example.com"},
			}

			result := Validate(req, false)

			if result.HasErrors() {
				t.Errorf("Expected no errors for version %s, got: %v", version, result.Errors)
			}
		})
	}
}

// TestValidateNonStandardVersion tests warning for non-standard HTTP version
func TestValidateNonStandardVersion(t *testing.T) {
	req := &models.HTTPRequest{
		Method:  "GET",
		URL:     "https://example.com",
		Version: "HTTP/4.0",
		Headers: map[string]string{"Host": "example.com"},
	}

	result := Validate(req, false)

	if len(result.Warnings) == 0 {
		t.Error("Expected warning for non-standard HTTP version")
	}
}

// TestValidateMissingHostHeader tests validation for missing Host header
func TestValidateMissingHostHeader(t *testing.T) {
	req := &models.HTTPRequest{
		Method:  "GET",
		URL:     "https://example.com",
		Version: "HTTP/1.1",
		Headers: map[string]string{},
	}

	result := Validate(req, false)

	// Should auto-add Host header, so might be a warning or no error
	// depending on implementation
	if result.HasErrors() {
		hasHostError := false
		for _, err := range result.Errors {
			if strings.Contains(strings.ToLower(err), "host") {
				hasHostError = true
				break
			}
		}
		if !hasHostError {
			t.Errorf("Expected Host-related error or warning, got: %v", result.Errors)
		}
	}
}

// TestValidateBodyWithContentLength tests validation of body with Content-Length
func TestValidateBodyWithContentLength(t *testing.T) {
	body := `{"name":"test"}`
	req := &models.HTTPRequest{
		Method:  "POST",
		URL:     "https://example.com",
		Version: "HTTP/1.1",
		Headers: map[string]string{
			"Host":           "example.com",
			"Content-Length": "15",
			"Content-Type":   "application/json",
		},
		Body: body,
	}

	result := Validate(req, false)

	if result.HasErrors() {
		t.Errorf("Expected no errors for POST with body and Content-Length, got: %v", result.Errors)
	}
}

// TestValidateBodyWithoutContentType tests warning for body without Content-Type
func TestValidateBodyWithoutContentType(t *testing.T) {
	req := &models.HTTPRequest{
		Method:  "POST",
		URL:     "https://example.com",
		Version: "HTTP/1.1",
		Headers: map[string]string{
			"Host": "example.com",
		},
		Body: `{"name":"test"}`,
	}

	result := Validate(req, false)

	// Should have warning about missing Content-Type
	if len(result.Warnings) == 0 {
		t.Error("Expected warning for body without Content-Type")
	}
}

// TestValidateGETWithBody tests warning for GET request with body
func TestValidateGETWithBody(t *testing.T) {
	req := &models.HTTPRequest{
		Method:  "GET",
		URL:     "https://example.com",
		Version: "HTTP/1.1",
		Headers: map[string]string{
			"Host": "example.com",
		},
		Body: "unexpected body",
	}

	result := Validate(req, false)

	// Should have warning about GET with body
	if len(result.Warnings) == 0 {
		t.Error("Expected warning for GET request with body")
	}
}

// TestValidateNoSecureFlag tests validation with no-secure flag
func TestValidateNoSecureFlag(t *testing.T) {
	req := &models.HTTPRequest{
		Method:  "GET",
		URL:     "/api/users",
		Version: "HTTP/1.1",
		Headers: map[string]string{
			"Host": "example.com",
		},
	}

	// With noSecure=true, should use HTTP
	result := Validate(req, true)

	if result.HasErrors() {
		t.Errorf("Expected no errors with no-secure flag, got: %v", result.Errors)
	}
}

// TestValidateInvalidURL tests validation of invalid URLs
func TestValidateInvalidURL(t *testing.T) {
	invalidURLs := []string{
		"ht!tp://invalid",
		"://noscheme.com",
		"",
	}

	for _, url := range invalidURLs {
		t.Run("Invalid_"+url, func(t *testing.T) {
			req := &models.HTTPRequest{
				Method:  "GET",
				URL:     url,
				Version: "HTTP/1.1",
				Headers: map[string]string{},
			}

			result := Validate(req, false)

			if !result.HasErrors() {
				t.Errorf("Expected error for invalid URL: %s", url)
			}
		})
	}
}

// TestValidateHTTPSURL tests validation of HTTPS URL
func TestValidateHTTPSURL(t *testing.T) {
	req := &models.HTTPRequest{
		Method:  "GET",
		URL:     "https://secure.example.com/api",
		Version: "HTTP/1.1",
		Headers: map[string]string{
			"Host": "secure.example.com",
		},
	}

	result := Validate(req, false)

	if result.HasErrors() {
		t.Errorf("Expected no errors for HTTPS URL, got: %v", result.Errors)
	}
}

// TestValidateHTTPURL tests validation of HTTP URL
func TestValidateHTTPURL(t *testing.T) {
	req := &models.HTTPRequest{
		Method:  "GET",
		URL:     "http://example.com/api",
		Version: "HTTP/1.1",
		Headers: map[string]string{
			"Host": "example.com",
		},
	}

	result := Validate(req, false)

	if result.HasErrors() {
		t.Errorf("Expected no errors for HTTP URL, got: %v", result.Errors)
	}
}

// TestValidateComplexRequest tests validation of a complex request
func TestValidateComplexRequest(t *testing.T) {
	req := &models.HTTPRequest{
		Method:  "POST",
		URL:     "https://api.example.com/users?page=1&limit=10",
		Version: "HTTP/1.1",
		Headers: map[string]string{
			"Host":          "api.example.com",
			"Content-Type":  "application/json",
			"Authorization": "Bearer token123",
			"User-Agent":    "test-client/1.0",
		},
		Body: `{"name":"John Doe","email":"john@example.com"}`,
	}

	result := Validate(req, false)

	if result.HasErrors() {
		t.Errorf("Expected no errors for complex valid request, got: %v", result.Errors)
	}
}

// TestValidateRequestLineFormatWithFullURL tests validation of request line with full URL
func TestValidateRequestLineFormatWithFullURL(t *testing.T) {
	req := &models.HTTPRequest{
		Method:  "GET",
		URL:     "https://graph.microsoft.com/v1.0/me",
		Version: "HTTP/1.1",
		Headers: map[string]string{
			"Host": "graph.microsoft.com",
		},
	}

	result := Validate(req, false)

	// Should have a warning about full URL in request line
	if !result.HasWarnings() {
		t.Error("Expected warning for full URL in request line")
	}

	// Check for specific warning message
	found := false
	for _, warning := range result.Warnings {
		if strings.Contains(warning, "Request line contains full URL") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected warning about request line format, got warnings: %v", result.Warnings)
	}
}

// TestValidateRequestLineFormatWithPath tests validation of request line with path only
func TestValidateRequestLineFormatWithPath(t *testing.T) {
	req := &models.HTTPRequest{
		Method:  "GET",
		URL:     "/v1.0/me",
		Version: "HTTP/1.1",
		Headers: map[string]string{
			"Host": "graph.microsoft.com",
		},
	}

	result := Validate(req, false)

	// Should not have a warning about request line format
	for _, warning := range result.Warnings {
		if strings.Contains(warning, "Request line contains full URL") {
			t.Errorf("Unexpected warning for path-only URL: %s", warning)
		}
	}
}

// TestValidateRequestLineFormatWithHTTPURL tests validation with http:// URL
func TestValidateRequestLineFormatWithHTTPURL(t *testing.T) {
	req := &models.HTTPRequest{
		Method:  "GET",
		URL:     "http://example.com/api/data",
		Version: "HTTP/1.1",
		Headers: map[string]string{
			"Host": "example.com",
		},
	}

	result := Validate(req, false)

	// Should have a warning about full URL in request line
	if !result.HasWarnings() {
		t.Error("Expected warning for full URL in request line")
	}

	// Check for specific warning message
	found := false
	for _, warning := range result.Warnings {
		if strings.Contains(warning, "Request line contains full URL") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected warning about request line format, got warnings: %v", result.Warnings)
	}
}
