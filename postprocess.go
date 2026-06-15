package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	blankLineRe     = regexp.MustCompile(`^\s*$`)
	trailingSpaceRe = regexp.MustCompile(`[ \t]+$`)
)

// postprocessText normalises line endings, trims trailing whitespace, and
// collapses runs of 2+ blank lines to one.
func postprocessText(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	lines := strings.Split(text, "\n")
	out := make([]string, 0, len(lines))
	blank := 0
	for _, line := range lines {
		line = trailingSpaceRe.ReplaceAllString(line, "")
		if blankLineRe.MatchString(line) {
			blank++
			if blank == 1 {
				out = append(out, "")
			}
			continue
		}
		blank = 0
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

// postprocess reads a Markdown file, cleans it, and writes the result.
func postprocess(input, output string, stdout bool) error {
	data, err := os.ReadFile(input)
	if err != nil {
		return fmt.Errorf("read %s: %w", input, err)
	}
	text := postprocessText(string(data))
	if stdout {
		_, err := fmt.Print(text)
		return err
	}
	if output == "" {
		output = input
	}
	return os.WriteFile(output, []byte(text), 0o644)
}
