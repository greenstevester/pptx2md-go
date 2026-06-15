package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
		err = postprocess(args[1], output, stdout)
	default:
		input := args[0]
		if _, statErr := os.Stat(input); statErr != nil {
			fmt.Fprintf(os.Stderr, "ERROR: file not found: %s\n", input)
			os.Exit(1)
		}
		stdout, output := parseOutputArgs(args[1:], input)
		err = convert(input, output, stdout)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
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
