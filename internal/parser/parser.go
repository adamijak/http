package parser

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/adamijak/http/internal/models"
)

// Parse parses an HTTP request in either HTP (HTTP Template Protocol) format
// or RFC compliant format and returns an HTTPRequest
//
// HTP Format (HTTP Template Protocol) - our friendly format with preprocessing:
// - Comments (lines starting with # or //)
// - Environment variables: ${VAR_NAME} or $VAR_NAME
// - Shell command execution: $(command)
// - Uses standard newlines (\n)
//
// RFC Compliant Format - standard HTTP format:
// - No preprocessing directives
// - Uses CRLF line endings (\r\n)
// - Ready to send over the wire
//
// The parser automatically detects the format:
// - If content contains \r\n, treats as RFC compliant (no preprocessing)
// - Otherwise, treats as HTP format (with preprocessing)
//
// Example HTP file:
// # This is a comment
// GET https://api.example.com/users HTTP/1.1
// Host: api.example.com
// Authorization: Bearer ${API_TOKEN}
// X-Request-ID: $(uuidgen)
//
// AI Agent Note: This parser is designed to be simple and extensible.
// Each preprocessing step is clearly separated for easy modification.
func Parse(content string) (*models.HTTPRequest, error) {
	// Auto-detect format based on line endings
	// RFC compliant format uses \r\n, HTP format uses \n
	if strings.Contains(content, "\r\n") {
		// RFC compliant format - no preprocessing
		return parseHTTP(content)
	}

	// HTP format - preprocess the content
	processed := preprocess(content)

	// Parse the preprocessed content
	return parseHTTP(processed)
}

// ParseRFCCompliant parses an RFC compliant HTTP request without preprocessing
// This is used for loading saved requests that are already preprocessed
// Deprecated: Use Parse() which auto-detects the format based on line endings.
// This function is kept for backward compatibility and will be removed in v2.0.
// New code should use Parse() instead, which handles both HTP and RFC compliant formats.
func ParseRFCCompliant(content string) (*models.HTTPRequest, error) {
	// Parse directly without preprocessing
	return parseHTTP(content)
}

// preprocess handles comments, environment variables, and shell commands
func preprocess(content string) string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	var lines []string

	for scanner.Scan() {
		line := scanner.Text()

		// Skip comments (lines starting with # or //)
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//") {
			continue
		}

		// Process environment variables: ${VAR} or $VAR
		line = expandEnvVars(line)

		// Process shell commands: $(command)
		line = executeShellCommands(line)

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// expandEnvVars expands environment variables in the format ${VAR} or $VAR
func expandEnvVars(line string) string {
	// Handle ${VAR_NAME} format
	for {
		start := strings.Index(line, "${")
		if start == -1 {
			break
		}
		end := strings.Index(line[start:], "}")
		if end == -1 {
			break
		}
		end += start

		varName := line[start+2 : end]
		varValue := os.Getenv(varName)
		line = line[:start] + varValue + line[end+1:]
	}

	// Handle $VAR_NAME format (simple case)
	// This is a simplified implementation that works for most cases
	parts := strings.Split(line, "$")
	if len(parts) > 1 {
		var result strings.Builder
		result.WriteString(parts[0])

		for i := 1; i < len(parts); i++ {
			part := parts[i]
			// Find the end of variable name (alphanumeric and underscore)
			varEnd := 0
			for varEnd < len(part) && (isAlphaNum(part[varEnd]) || part[varEnd] == '_') {
				varEnd++
			}

			if varEnd > 0 {
				varName := part[:varEnd]
				varValue := os.Getenv(varName)
				result.WriteString(varValue)
				result.WriteString(part[varEnd:])
			} else {
				result.WriteString("$")
				result.WriteString(part)
			}
		}

		line = result.String()
	}

	return line
}

// executeShellCommands executes shell commands in the format $(command)
func executeShellCommands(line string) string {
	for {
		start := strings.Index(line, "$(")
		if start == -1 {
			break
		}
		end := strings.Index(line[start:], ")")
		if end == -1 {
			break
		}
		end += start

		command := line[start+2 : end]
		output := runCommand(command)
		line = line[:start] + output + line[end+1:]
	}

	return line
}

// runCommand executes a shell command and returns its output
func runCommand(command string) string {
	cmd := exec.Command("sh", "-c", command)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Sprintf("[error: %v]", err)
	}
	return strings.TrimSpace(string(output))
}

// parseHTTP parses the preprocessed HTTP request
func parseHTTP(content string) (*models.HTTPRequest, error) {
	req := models.NewHTTPRequest()

	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty request")
	}

	// Find the request line (first non-empty line)
	requestLineIdx := -1
	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			requestLineIdx = i
			break
		}
	}

	if requestLineIdx == -1 {
		return nil, fmt.Errorf("no request line found")
	}

	// Parse request line: METHOD URL [VERSION]
	requestLine := strings.TrimSpace(lines[requestLineIdx])
	parts := strings.Fields(requestLine)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid request line (must contain at least METHOD and URL): %s", requestLine)
	}

	req.Method = parts[0]
	req.URL = parts[1]
	if len(parts) >= 3 {
		req.Version = parts[2]
	}

	// Parse headers (lines until empty line or body)
	i := requestLineIdx + 1
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			i++
			break
		}

		// Parse header: Key: Value
		colonIdx := strings.Index(line, ":")
		if colonIdx == -1 {
			return nil, fmt.Errorf("invalid header: %s", line)
		}

		key := strings.TrimSpace(line[:colonIdx])
		value := strings.TrimSpace(line[colonIdx+1:])
		req.Headers[key] = value
		i++
	}

	// Parse body (rest of the content)
	if i < len(lines) {
		body := strings.Join(lines[i:], "\n")
		req.Body = strings.TrimSpace(body)
	}

	return req, nil
}

// isAlphaNum checks if a byte is alphanumeric
func isAlphaNum(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}
