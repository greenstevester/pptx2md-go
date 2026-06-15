package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseOutputArgsDefault(t *testing.T) {
	stdout, output := parseOutputArgs(nil, "deck.pptx")
	if stdout || output != "deck.md" {
		t.Fatalf("stdout=%v output=%q", stdout, output)
	}
}

func TestParseOutputArgsStdout(t *testing.T) {
	stdout, output := parseOutputArgs([]string{"--stdout"}, "deck.pptx")
	if !stdout || output != "" {
		t.Fatalf("stdout=%v output=%q", stdout, output)
	}
}

func TestParseOutputArgsExplicit(t *testing.T) {
	stdout, output := parseOutputArgs([]string{"out.md"}, "deck.pptx")
	if stdout || output != "out.md" {
		t.Fatalf("stdout=%v output=%q", stdout, output)
	}
}

func TestConvertWritesFile(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out.md")
	if err := convert("testdata/fixtures/basic.pptx", out, false); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "## Slide 1: Quarterly Review") {
		t.Fatalf("unexpected output:\n%s", data)
	}
}
