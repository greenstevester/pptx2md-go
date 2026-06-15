package main

import (
	"fmt"
	"os"
)

// convert extracts a .pptx, renders Markdown, cleans it, and writes the result.
func convert(input, output string, stdout bool) error {
	deck, err := ExtractFile(input)
	if err != nil {
		return fmt.Errorf("extract: %w", err)
	}
	md := postprocessText(ToMarkdown(deck))
	if stdout {
		_, err := fmt.Print(md)
		return err
	}
	if err := os.WriteFile(output, []byte(md), 0o644); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Written to %s\n", output)
	return nil
}
