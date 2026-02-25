package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/adamijak/http/internal/client"
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
	
	flag.Parse()

	// Show version
	if *version {
		fmt.Printf("%s version %s\n", AppName, Version)
		return
	}

	// Read from stdin
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	// Parse the .http file
	req, err := parser.Parse(string(input))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		os.Exit(1)
	}

	// Validate the request
	validationResult := validator.Validate(req)
	if !*noColor {
		validationResult.PrintColored(os.Stdout)
	} else {
		validationResult.Print(os.Stdout)
	}

	// Exit if there are errors
	if validationResult.HasErrors() {
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
