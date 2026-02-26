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
	noSend := flag.Bool("no-send", false, "Output the RFC compliant request to stdout without sending")
	inputFile := flag.String("f", "", "Read request from file (supports both HTP and RFC compliant formats)")
	strict := flag.Bool("strict", false, "Strict mode: fail on any validation warnings (RFC compliance enforcement)")

	flag.Parse()

	// Show version
	if *version {
		fmt.Printf("%s version %s\n", AppName, Version)
		return
	}

	var req *models.HTTPRequest
	var err error
	var input []byte

	// Read from file or stdin
	if *inputFile != "" {
		input, err = os.ReadFile(*inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}
	} else {
		input, err = io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			os.Exit(1)
		}
	}

	// Parse the request (auto-detects HTP or RFC compliant format)
	req, err = parser.Parse(string(input))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		os.Exit(1)
	}

	// Validate the request
	validationResult := validator.Validate(req, *noSecure)

	// Output RFC compliant request to stdout if requested (skip validation output)
	if *noSend {
		// In strict mode, still fail on warnings
		if *strict && validationResult.HasWarnings() {
			// Show validation in this case
			if !*noColor {
				validationResult.PrintColored(os.Stderr)
			} else {
				validationResult.Print(os.Stderr)
			}
			fmt.Fprintf(os.Stderr, "\nStrict mode: Request has validation warnings and cannot be output\n")
			os.Exit(1)
		}
		// Exit if there are errors
		if validationResult.HasErrors() {
			// Show validation errors
			if !*noColor {
				validationResult.PrintColored(os.Stderr)
			} else {
				validationResult.Print(os.Stderr)
			}
			os.Exit(1)
		}
		// Output raw request to stdout (no validation messages)
		fmt.Print(req.ToRawRequest())
		return
	}

	// For normal mode, show validation results
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
