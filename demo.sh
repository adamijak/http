#!/bin/bash
# Demo script to showcase HTTP client features
# AI Agent Note: This demonstrates all major features visually

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║         HTTP Client Tool - Feature Demonstration            ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

# Build first
echo "Building http client..."
go build -o http
echo ""

# Feature 1: Simple Request
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Feature 1: Simple GET Request with Validation"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
cat examples/simple-get.http | ./http -dry-run
echo ""

# Feature 2: Environment Variables
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Feature 2: Environment Variable Substitution"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
export API_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
export API_KEY="sk-1234567890abcdef"
echo "Environment variables set:"
echo "  API_TOKEN=$API_TOKEN"
echo "  API_KEY=$API_KEY"
echo ""
cat examples/with-env-vars.http | ./http -dry-run
echo ""

# Feature 3: Shell Commands
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Feature 3: Dynamic Shell Command Execution"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
cat examples/with-shell-commands.http | ./http -dry-run
echo ""

# Feature 4: POST with Body
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Feature 4: POST Request with JSON Body (Auto Content-Length)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
cat examples/post-json.http | ./http -dry-run
echo ""

# Feature 5: Validation Errors
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Feature 5: Request Validation with Error Detection"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
./http -dry-run <<'EOF' || true
GET example.com HTTP/1.1
Content-Type: application/json
EOF
echo ""

# Feature 6: No Color Mode
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Feature 6: Plain Output (No Colors)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
cat examples/simple-get.http | ./http -dry-run -no-color
echo ""

# Feature 7: Comments
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Feature 7: Comment Preprocessing (# and //)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
./http -dry-run <<'EOF'
# This is a comment - ignored
// This is also a comment - ignored
GET https://api.github.com HTTP/1.1
Host: api.github.com
# User-Agent is added below
User-Agent: demo-client/1.0
EOF
echo ""

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║                    Demo Complete!                           ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""
echo "Try it yourself:"
echo "  echo 'GET https://httpbin.org/get HTTP/1.1' | ./http -dry-run"
echo ""
