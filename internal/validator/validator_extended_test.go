package validator

import (
	"testing"

	"github.com/adamijak/http/internal/models"
)

// TestValidateEmptyHeaders tests validation with no headers
func TestValidateEmptyHeaders(t *testing.T) {
	req := &models.HTTPRequest{
		Method:  "GET",
		URL:     "https://example.com",
		Version: "HTTP/1.1",
		Headers: map[string]string{},
	}

	result := Validate(req, false)

	// Should have warning about missing Host header
	if !result.HasWarnings() {
		t.Errorf("Expected warning for missing Host header")
	}
}

// TestValidateLongURL tests validation of very long URLs
func TestValidateLongURL(t *testing.T) {
	longPath := ""
	for i := 0; i < 100; i++ {
		longPath += "/segment"
	}

	req := &models.HTTPRequest{
		Method:  "GET",
		URL:     "https://example.com" + longPath,
		Version: "HTTP/1.1",
		Headers: map[string]string{"Host": "example.com"},
	}

	result := Validate(req, false)

	// Long URL should still be valid
	if result.HasErrors() {
		t.Errorf("Expected long URL to be valid")
	}
}

// TestValidateSpecialMethodNames tests non-standard but valid method names
func TestValidateSpecialMethodNames(t *testing.T) {
	methods := []string{"PROPFIND", "PROPPATCH", "MKCOL", "COPY", "MOVE", "LOCK", "UNLOCK"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := &models.HTTPRequest{
				Method:  method,
				URL:     "https://example.com",
				Version: "HTTP/1.1",
				Headers: map[string]string{"Host": "example.com"},
			}

			result := Validate(req, false)

			// Non-standard methods should generate a warning, not an error
			if result.HasErrors() {
				t.Errorf("Expected method %s to be accepted (possibly with warning)", method)
			}
		})
	}
}

// TestValidateURLWithFragment tests URL with fragment identifier
func TestValidateURLWithFragment(t *testing.T) {
	req := &models.HTTPRequest{
		Method:  "GET",
		URL:     "https://example.com/page#section",
		Version: "HTTP/1.1",
		Headers: map[string]string{"Host": "example.com"},
	}

	result := Validate(req, false)

	// URL with fragment should be valid (though fragments aren't sent to server)
	if result.HasErrors() {
		t.Errorf("Expected URL with fragment to be valid")
	}
}

// TestValidateURLWithSpecialChars tests URLs with encoded characters
func TestValidateURLWithSpecialChars(t *testing.T) {
	req := &models.HTTPRequest{
		Method:  "GET",
		URL:     "https://example.com/path?q=hello%20world&lang=en",
		Version: "HTTP/1.1",
		Headers: map[string]string{"Host": "example.com"},
	}

	result := Validate(req, false)

	if result.HasErrors() {
		t.Errorf("Expected URL with encoded characters to be valid")
	}
}

// TestValidateIPv4Address tests URL with IPv4 address
func TestValidateIPv4Address(t *testing.T) {
	req := &models.HTTPRequest{
		Method:  "GET",
		URL:     "http://192.168.1.1:8080/api",
		Version: "HTTP/1.1",
		Headers: map[string]string{"Host": "192.168.1.1:8080"},
	}

	result := Validate(req, false)

	if result.HasErrors() {
		t.Errorf("Expected IPv4 URL to be valid")
	}
}

// TestValidateIPv6Address tests URL with IPv6 address
func TestValidateIPv6Address(t *testing.T) {
	req := &models.HTTPRequest{
		Method:  "GET",
		URL:     "http://[::1]:8080/api",
		Version: "HTTP/1.1",
		Headers: map[string]string{"Host": "[::1]:8080"},
	}

	result := Validate(req, false)

	if result.HasErrors() {
		t.Errorf("Expected IPv6 URL to be valid")
	}
}

// TestValidateLocalhostVariants tests different localhost formats
func TestValidateLocalhostVariants(t *testing.T) {
	variants := []string{
		"http://localhost/api",
		"http://127.0.0.1/api",
		"http://[::1]/api",
		"http://localhost.localdomain/api",
	}

	for _, url := range variants {
		t.Run(url, func(t *testing.T) {
			req := &models.HTTPRequest{
				Method:  "GET",
				URL:     url,
				Version: "HTTP/1.1",
				Headers: map[string]string{"Host": "localhost"},
			}

			result := Validate(req, false)

			if result.HasErrors() {
				t.Errorf("Expected localhost variant %s to be valid", url)
			}
		})
	}
}

// TestValidateMultipleContentHeaders tests requests with multiple content-related headers
func TestValidateMultipleContentHeaders(t *testing.T) {
	req := &models.HTTPRequest{
		Method:  "POST",
		URL:     "https://example.com/api",
		Version: "HTTP/1.1",
		Headers: map[string]string{
			"Host":              "example.com",
			"Content-Type":      "application/json",
			"Content-Length":    "100",
			"Content-Encoding":  "gzip",
			"Content-Language":  "en-US",
			"Transfer-Encoding": "chunked",
		},
		Body: `{"test":true}`,
	}

	result := Validate(req, false)

	// Should be valid (though some combinations might generate warnings)
	if result.HasErrors() {
		t.Errorf("Expected request with multiple content headers to be valid")
	}
}

// TestValidateAuthorizationHeaders tests various authorization header formats
func TestValidateAuthorizationHeaders(t *testing.T) {
	authTypes := map[string]string{
		"Bearer":    "Bearer token123",
		"Basic":     "Basic dXNlcjpwYXNz",
		"Digest":    "Digest username=\"user\"",
		"OAuth":     "OAuth oauth_token=\"token\"",
		"AWS4-HMAC": "AWS4-HMAC-SHA256 Credential=...",
	}

	for authType, authValue := range authTypes {
		t.Run(authType, func(t *testing.T) {
			req := &models.HTTPRequest{
				Method:  "GET",
				URL:     "https://api.example.com",
				Version: "HTTP/1.1",
				Headers: map[string]string{
					"Host":          "api.example.com",
					"Authorization": authValue,
				},
			}

			result := Validate(req, false)

			if result.HasErrors() {
				t.Errorf("Expected request with %s authorization to be valid", authType)
			}
		})
	}
}

// TestValidateCacheHeaders tests cache-control related headers
func TestValidateCacheHeaders(t *testing.T) {
	req := &models.HTTPRequest{
		Method:  "GET",
		URL:     "https://example.com/resource",
		Version: "HTTP/1.1",
		Headers: map[string]string{
			"Host":              "example.com",
			"Cache-Control":     "no-cache, no-store, must-revalidate",
			"Pragma":            "no-cache",
			"Expires":           "0",
			"If-None-Match":     "\"abc123\"",
			"If-Modified-Since": "Wed, 21 Oct 2015 07:28:00 GMT",
		},
	}

	result := Validate(req, false)

	if result.HasErrors() {
		t.Errorf("Expected request with cache headers to be valid")
	}
}

// TestValidateCORSHeaders tests CORS-related headers
func TestValidateCORSHeaders(t *testing.T) {
	req := &models.HTTPRequest{
		Method:  "OPTIONS",
		URL:     "https://api.example.com/resource",
		Version: "HTTP/1.1",
		Headers: map[string]string{
			"Host":                           "api.example.com",
			"Origin":                         "https://example.com",
			"Access-Control-Request-Method":  "POST",
			"Access-Control-Request-Headers": "Content-Type",
		},
	}

	result := Validate(req, false)

	if result.HasErrors() {
		t.Errorf("Expected CORS preflight request to be valid")
	}
}

// TestValidateCustomHeaders tests requests with many custom headers
func TestValidateCustomHeaders(t *testing.T) {
	headers := map[string]string{
		"Host":              "example.com",
		"X-Request-ID":      "req-12345",
		"X-Correlation-ID":  "corr-67890",
		"X-API-Key":         "key-abc",
		"X-Client-Version":  "1.0.0",
		"X-Platform":        "web",
		"X-User-Agent":      "custom/1.0",
		"X-Forwarded-For":   "192.168.1.1",
		"X-Forwarded-Proto": "https",
		"X-Real-IP":         "192.168.1.1",
	}

	req := &models.HTTPRequest{
		Method:  "GET",
		URL:     "https://example.com/api",
		Version: "HTTP/1.1",
		Headers: headers,
	}

	result := Validate(req, false)

	if result.HasErrors() {
		t.Errorf("Expected request with custom headers to be valid")
	}
}

// TestValidateRangeRequest tests HTTP range requests
func TestValidateRangeRequest(t *testing.T) {
	req := &models.HTTPRequest{
		Method:  "GET",
		URL:     "https://example.com/file.pdf",
		Version: "HTTP/1.1",
		Headers: map[string]string{
			"Host":          "example.com",
			"Range":         "bytes=0-1023",
			"If-Range":      "\"abc123\"",
			"Accept-Ranges": "bytes",
		},
	}

	result := Validate(req, false)

	if result.HasErrors() {
		t.Errorf("Expected range request to be valid")
	}
}

// TestValidateConditionalRequest tests conditional request headers
func TestValidateConditionalRequest(t *testing.T) {
	req := &models.HTTPRequest{
		Method:  "GET",
		URL:     "https://example.com/resource",
		Version: "HTTP/1.1",
		Headers: map[string]string{
			"Host":                "example.com",
			"If-Match":            "\"abc123\"",
			"If-Unmodified-Since": "Wed, 21 Oct 2015 07:28:00 GMT",
		},
	}

	result := Validate(req, false)

	if result.HasErrors() {
		t.Errorf("Expected conditional request to be valid")
	}
}

// TestValidateWebSocketUpgrade tests WebSocket upgrade request
func TestValidateWebSocketUpgrade(t *testing.T) {
	req := &models.HTTPRequest{
		Method:  "GET",
		URL:     "https://example.com/ws",
		Version: "HTTP/1.1",
		Headers: map[string]string{
			"Host":                   "example.com",
			"Upgrade":                "websocket",
			"Connection":             "Upgrade",
			"Sec-WebSocket-Key":      "dGhlIHNhbXBsZSBub25jZQ==",
			"Sec-WebSocket-Version":  "13",
			"Sec-WebSocket-Protocol": "chat",
		},
	}

	result := Validate(req, false)

	// WebSocket upgrade should be valid
	if result.HasErrors() {
		t.Errorf("Expected WebSocket upgrade request to be valid")
	}
}

// TestValidateHTTP2Request tests HTTP/2 specific validation
func TestValidateHTTP2Request(t *testing.T) {
	req := &models.HTTPRequest{
		Method:  "GET",
		URL:     "https://example.com/api",
		Version: "HTTP/2",
		Headers: map[string]string{
			"Host": "example.com",
		},
	}

	result := Validate(req, false)

	// HTTP/2 should be recognized
	if result.HasErrors() {
		t.Errorf("Expected HTTP/2 request to be valid")
	}
}

// TestValidateContentNegotiation tests content negotiation headers
func TestValidateContentNegotiation(t *testing.T) {
	req := &models.HTTPRequest{
		Method:  "GET",
		URL:     "https://api.example.com/data",
		Version: "HTTP/1.1",
		Headers: map[string]string{
			"Host":            "api.example.com",
			"Accept":          "application/json, application/xml;q=0.9, */*;q=0.8",
			"Accept-Encoding": "gzip, deflate, br",
			"Accept-Language": "en-US,en;q=0.9,es;q=0.8",
			"Accept-Charset":  "utf-8, iso-8859-1;q=0.5",
		},
	}

	result := Validate(req, false)

	if result.HasErrors() {
		t.Errorf("Expected request with content negotiation to be valid")
	}
}

// TestValidateCookieHeaders tests cookie handling
func TestValidateCookieHeaders(t *testing.T) {
	req := &models.HTTPRequest{
		Method:  "GET",
		URL:     "https://example.com/page",
		Version: "HTTP/1.1",
		Headers: map[string]string{
			"Host":   "example.com",
			"Cookie": "session=abc123; user=john; preferences=dark_mode",
		},
	}

	result := Validate(req, false)

	if result.HasErrors() {
		t.Errorf("Expected request with cookies to be valid")
	}
}

// TestValidateMultipartFormData tests multipart form data
func TestValidateMultipartFormData(t *testing.T) {
	req := &models.HTTPRequest{
		Method:  "POST",
		URL:     "https://example.com/upload",
		Version: "HTTP/1.1",
		Headers: map[string]string{
			"Host":         "example.com",
			"Content-Type": "multipart/form-data; boundary=----WebKitFormBoundary",
		},
		Body: "------WebKitFormBoundary\r\nContent-Disposition: form-data; name=\"field\"\r\n\r\nvalue\r\n------WebKitFormBoundary--",
	}

	result := Validate(req, false)

	if result.HasErrors() {
		t.Errorf("Expected multipart form data request to be valid")
	}
}

// TestValidateRedirectHeaders tests redirect-related headers
func TestValidateRedirectHeaders(t *testing.T) {
	req := &models.HTTPRequest{
		Method:  "GET",
		URL:     "https://example.com/old-page",
		Version: "HTTP/1.1",
		Headers: map[string]string{
			"Host":     "example.com",
			"Referer":  "https://example.com/previous-page",
			"Location": "https://example.com/new-page",
		},
	}

	result := Validate(req, false)

	if result.HasErrors() {
		t.Errorf("Expected request with redirect headers to be valid")
	}
}

// TestValidateProxyHeaders tests proxy-related headers
func TestValidateProxyHeaders(t *testing.T) {
	req := &models.HTTPRequest{
		Method:  "GET",
		URL:     "https://example.com/api",
		Version: "HTTP/1.1",
		Headers: map[string]string{
			"Host":              "example.com",
			"X-Forwarded-For":   "192.168.1.1, 10.0.0.1",
			"X-Forwarded-Proto": "https",
			"X-Forwarded-Host":  "original.example.com",
			"X-Forwarded-Port":  "443",
			"Forwarded":         "for=192.168.1.1;proto=https;host=example.com",
			"Via":               "1.1 proxy.example.com",
		},
	}

	result := Validate(req, false)

	if result.HasErrors() {
		t.Errorf("Expected request with proxy headers to be valid")
	}
}

// TestValidateLargeBodyWithContentLength tests validation of body size
func TestValidateLargeBodyWithContentLength(t *testing.T) {
	largeBody := string(make([]byte, 10000))

	req := &models.HTTPRequest{
		Method:  "POST",
		URL:     "https://example.com/upload",
		Version: "HTTP/1.1",
		Headers: map[string]string{
			"Host":           "example.com",
			"Content-Type":   "application/octet-stream",
			"Content-Length": "10000",
		},
		Body: largeBody,
	}

	result := Validate(req, false)

	// Should validate without errors (size warnings are implementation-specific)
	if result.HasErrors() {
		t.Errorf("Expected large body request to be valid")
	}
}
