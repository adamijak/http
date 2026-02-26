package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/adamijak/http/internal/client"
	"github.com/adamijak/http/internal/models"
	"github.com/adamijak/http/internal/parser"
	"github.com/adamijak/http/internal/validator"
)

// Main entry point for the HTTP client tool
// This tool processes .http files, validates them against HTTP standards,
// and sends requests over TCP with colored output
func main() {
	// Command-line flags
	dryRun := flag.Bool("dry-run", false, "Show preprocessed and validated request without sending")
	noColor := flag.Bool("no-color", false, "Disable colored output")
	verbose := flag.Bool("v", false, "Verbose output")
	version := flag.Bool("version", false, "Show version information")
	noSecure := flag.Bool("no-secure", false, "Send request in plain HTTP instead of HTTPS")
	saveRequest := flag.String("save-request", "", "Save the preprocessed RFC compliant request to a file instead of sending")
	loadRequest := flag.String("load-request", "", "Load an RFC compliant request from a file (bypasses preprocessing)")
	strict := flag.Bool("strict", false, "Strict mode: fail on any validation warnings (RFC compliance enforcement)")

	flag.Parse()

	// Show version
	if *version {
		fmt.Printf("%s version %s\n", AppName, Version)
		return
	}

	var req *models.HTTPRequest
	var err error

	// Load request from file or stdin
	if *loadRequest != "" {
		// Load from file (RFC compliant format, no preprocessing)
		content, err := os.ReadFile(*loadRequest)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}
		req, err = parser.ParseRFCCompliant(string(content))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Read from stdin
		input, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			os.Exit(1)
		}

		// Parse the .http file
		req, err = parser.Parse(string(input))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
			os.Exit(1)
		}
	}

	// Validate the request
	validationResult := validator.Validate(req, *noSecure)
	if !*noColor {
		validationResult.PrintColored(os.Stdout)
	} else {
		validationResult.Print(os.Stdout)
	}

	// Exit if there are errors
	if validationResult.HasErrors() {
		os.Exit(1)
	}

	// In strict mode, fail on warnings too
	if *strict && validationResult.HasWarnings() {
		fmt.Fprintf(os.Stderr, "\nStrict mode: Request has validation warnings and cannot be sent\n")
		os.Exit(1)
	}

	// Save request to file if requested
	if *saveRequest != "" {
		err := req.SaveToFile(*saveRequest)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error saving request: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("\n✓ RFC compliant request saved to: %s\n", *saveRequest)
		return
	}

	// If dry-run, just show the preprocessed request
	if *dryRun {
		fmt.Println("\n--- Preprocessed Request ---")
		req.Print(os.Stdout, !*noColor)
		return
	}

	// Send the request
	if *verbose {
		fmt.Println("\n--- Sending Request ---")
	}

	resp, err := client.Send(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error sending request: %v\n", err)
		os.Exit(1)
	}

	// Print response
	resp.Print(os.Stdout, !*noColor)
}
