package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

var (
	h1Re        = regexp.MustCompile(`(?m)^# `)
	slideHeadRe = regexp.MustCompile(`(?m)^## Slide (\d+): `)
	ooxmlTagRe  = regexp.MustCompile(`<[ap]:[A-Za-z]`)
	notesRe     = regexp.MustCompile(`(?m)^> \*\*Notes:\*\* (.*)$`)
	onlyDigits  = regexp.MustCompile(`^\d+$`)
)

// localEvalDir is the gitignored directory holding real decks, overridable via
// PPTX_EVAL_DIR. Its contents are never committed (may be licensed/personal).
func localEvalDir() string {
	if d := os.Getenv("PPTX_EVAL_DIR"); d != "" {
		return d
	}
	return filepath.Join("testdata", "local")
}

// TestRealWorldEvals converts every real deck found locally and asserts
// structural invariants (not exact text — real content is large and unstable).
// Skips cleanly when no decks are present, so CI stays green with nothing committed.
func TestRealWorldEvals(t *testing.T) {
	matches, _ := filepath.Glob(filepath.Join(localEvalDir(), "*.pptx"))
	if len(matches) == 0 {
		t.Skipf("no decks in %s (set PPTX_EVAL_DIR or drop .pptx files there)", localEvalDir())
	}

	for _, path := range matches {
		t.Run(filepath.Base(path), func(t *testing.T) {
			deck, err := ExtractFile(path)
			if err != nil {
				t.Fatalf("extract: %v", err)
			}
			md := postprocessText(ToMarkdown(deck))

			if strings.TrimSpace(md) == "" {
				t.Fatal("empty output")
			}
			if n := len(h1Re.FindAllString(md, -1)); n != 1 {
				t.Fatalf("want exactly one H1 deck title, got %d", n)
			}
			if len(deck.Slides) == 0 {
				t.Fatal("no slides extracted")
			}

			// Rendered headings must match extracted slides one-for-one, carry each
			// slide's number, and stay in strictly increasing presentation order.
			heads := slideHeadRe.FindAllStringSubmatch(md, -1)
			if len(heads) != len(deck.Slides) {
				t.Fatalf("slide headings = %d, extracted slides = %d", len(heads), len(deck.Slides))
			}
			prev := 0
			for i, h := range heads {
				if h[1] != strconv.Itoa(deck.Slides[i].Number) {
					t.Fatalf("heading %d numbered %q, slide.Number = %d", i, h[1], deck.Slides[i].Number)
				}
				num, _ := strconv.Atoi(h[1])
				if num <= prev {
					t.Fatalf("slide order not strictly increasing: %d after %d", num, prev)
				}
				prev = num
			}

			// Image policy: placeholders only, never links or media paths.
			if strings.Contains(md, "![") || strings.Contains(md, "media/") {
				t.Fatal("image link or media/ path leaked into output")
			}
			// No raw OOXML tags survived into the Markdown.
			if tag := ooxmlTagRe.FindString(md); tag != "" {
				t.Fatalf("raw OOXML tag leaked: %q", tag)
			}
			// Speaker notes must not be a bare slide number (slide-number-leak guard).
			for _, m := range notesRe.FindAllStringSubmatch(md, -1) {
				if onlyDigits.MatchString(strings.TrimSpace(m[1])) {
					t.Fatalf("note looks like a leaked slide number: %q", m[1])
				}
			}

			t.Logf("OK: %d slides, %d bytes of Markdown", len(deck.Slides), len(md))
		})
	}
}
