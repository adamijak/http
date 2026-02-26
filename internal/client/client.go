package client

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/adamijak/http/internal/models"
)

// Send sends an HTTP request over TCP and returns the response
//
// AI Agent Note: This function handles both HTTP and HTTPS connections.
// The connection logic is straightforward and uses standard library only.
//
// Steps:
// 1. Parse URL to get host and port
// 2. Establish TCP connection (with TLS for HTTPS)
// 3. Send raw HTTP request
// 4. Read and parse response
func Send(req *models.HTTPRequest) (*models.HTTPResponse, error) {
	// Parse URL
	parsedURL, err := url.Parse(req.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Determine host and port
	host := parsedURL.Host
	port := parsedURL.Port()

	if port == "" {
		if parsedURL.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
		host = fmt.Sprintf("%s:%s", parsedURL.Hostname(), port)
	}

	// Establish connection
	var conn net.Conn
	if parsedURL.Scheme == "https" {
		// HTTPS connection with TLS
		tlsConfig := &tls.Config{
			ServerName: parsedURL.Hostname(),
		}
		conn, err = tls.Dial("tcp", host, tlsConfig)
	} else {
		// Plain HTTP connection
		conn, err = net.Dial("tcp", host)
	}

	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}
	defer conn.Close()

	// Set timeout
	err = conn.SetDeadline(time.Now().Add(30 * time.Second))
	if err != nil {
		return nil, fmt.Errorf("failed to set deadline: %w", err)
	}

	// Prepare the request path
	path := parsedURL.Path
	if path == "" {
		path = "/"
	}
	if parsedURL.RawQuery != "" {
		path += "?" + parsedURL.RawQuery
	}

	// Build raw request
	rawRequest := buildRawRequest(req, path)

	// Send request
	_, err = conn.Write([]byte(rawRequest))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Read response
	return readResponse(conn)
}

// buildRawRequest builds the raw HTTP request string
func buildRawRequest(req *models.HTTPRequest, path string) string {
	var sb strings.Builder

	// Request line
	sb.WriteString(fmt.Sprintf("%s %s %s\r\n", req.Method, path, req.Version))

	// Headers
	for key, value := range req.Headers {
		sb.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}

	// Empty line between headers and body
	sb.WriteString("\r\n")

	// Body
	if req.Body != "" {
		sb.WriteString(req.Body)
	}

	return sb.String()
}

// readResponse reads and parses the HTTP response
func readResponse(conn net.Conn) (*models.HTTPResponse, error) {
	resp := &models.HTTPResponse{
		Headers: make(map[string]string),
	}

	reader := bufio.NewReader(conn)

	// Read status line
	statusLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read status line: %w", err)
	}

	// Parse status line: VERSION STATUS_CODE STATUS_TEXT
	statusLine = strings.TrimSpace(statusLine)
	parts := strings.SplitN(statusLine, " ", 3)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid status line: %s", statusLine)
	}

	resp.Version = parts[0]
	resp.StatusCode, err = strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid status code: %s", parts[1])
	}
	if len(parts) >= 3 {
		resp.Status = parts[2]
	}

	// Read headers
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read headers: %w", err)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			break
		}

		// Parse header
		colonIdx := strings.Index(line, ":")
		if colonIdx == -1 {
			continue
		}

		key := strings.TrimSpace(line[:colonIdx])
		value := strings.TrimSpace(line[colonIdx+1:])
		resp.Headers[key] = value
	}

	// Read body
	// Check for Content-Length
	contentLengthStr, hasContentLength := resp.Headers["Content-Length"]
	if hasContentLength {
		contentLength, err := strconv.Atoi(contentLengthStr)
		if err == nil && contentLength > 0 {
			body := make([]byte, contentLength)
			_, err = io.ReadFull(reader, body)
			if err != nil {
				return nil, fmt.Errorf("failed to read body: %w", err)
			}
			resp.Body = string(body)
			return resp, nil
		}
	}

	// Check for chunked encoding
	transferEncoding, hasTE := resp.Headers["Transfer-Encoding"]
	if hasTE && strings.Contains(strings.ToLower(transferEncoding), "chunked") {
		body, err := readChunkedBody(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read chunked body: %w", err)
		}
		resp.Body = body
		return resp, nil
	}

	// Read until connection closes (for HTTP/1.0 or no Content-Length)
	var sb strings.Builder
	buf := make([]byte, 4096)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			sb.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}
	resp.Body = sb.String()

	return resp, nil
}

// readChunkedBody reads a chunked transfer encoded body
func readChunkedBody(reader *bufio.Reader) (string, error) {
	var sb strings.Builder

	for {
		// Read chunk size
		sizeLine, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		sizeLine = strings.TrimSpace(sizeLine)
		// Parse hex chunk size
		chunkSize, err := strconv.ParseInt(sizeLine, 16, 64)
		if err != nil {
			return "", fmt.Errorf("invalid chunk size: %s", sizeLine)
		}

		// Last chunk
		if chunkSize == 0 {
			// Read trailing CRLF
			_, err = reader.ReadString('\n')
			if err != nil {
				return "", err
			}
			break
		}

		// Read chunk data
		chunk := make([]byte, chunkSize)
		_, err = io.ReadFull(reader, chunk)
		if err != nil {
			return "", err
		}
		sb.Write(chunk)

		// Read trailing CRLF after chunk
		_, err = reader.ReadString('\n')
		if err != nil {
			return "", err
		}
	}

	return sb.String(), nil
}
