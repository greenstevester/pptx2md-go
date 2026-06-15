package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"pptx-to-agent-md/internal/pptx"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		printUsage()
		os.Exit(0)
	}

	var err error
	switch args[0] {
	case "postprocess":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "ERROR: postprocess requires an input file")
			os.Exit(1)
		}
		stdout, output := parseOutputArgs(args[2:], args[1])
		err = runPostprocess(args[1], output, stdout)
	default:
		input := args[0]
		if _, statErr := os.Stat(input); statErr != nil {
			fmt.Fprintf(os.Stderr, "ERROR: file not found: %s\n", input)
			os.Exit(1)
		}
		stdout, output := parseOutputArgs(args[1:], input)
		err = runConvert(input, output, stdout)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}

// runConvert converts a .pptx and writes the Markdown to a file or stdout.
func runConvert(input, output string, stdout bool) error {
	md, err := pptx.Convert(input)
	if err != nil {
		return err
	}
	return write(md, output, stdout)
}

// runPostprocess re-runs the cleanup pass over an existing Markdown file.
func runPostprocess(input, output string, stdout bool) error {
	data, err := os.ReadFile(input)
	if err != nil {
		return fmt.Errorf("read %s: %w", input, err)
	}
	return write(pptx.PostprocessText(string(data)), output, stdout)
}

func write(content, output string, stdout bool) error {
	if stdout {
		_, err := fmt.Print(content)
		return err
	}
	if err := os.WriteFile(output, []byte(content), 0o644); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Written to %s\n", output)
	return nil
}

func parseOutputArgs(args []string, input string) (bool, string) {
	var stdout bool
	var output string
	for _, a := range args {
		if a == "--stdout" {
			stdout = true
		} else {
			output = a
		}
	}
	if output == "" && !stdout {
		ext := filepath.Ext(input)
		output = strings.TrimSuffix(input, ext) + ".md"
	}
	return stdout, output
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `Usage:
  pptx-to-md <input.pptx> [output.md] [--stdout]
  pptx-to-md postprocess <input.md> [output.md] [--stdout]

Convert PowerPoint presentations to clean, agent-readable Markdown.
Pure Go — no external dependencies.`)
}
