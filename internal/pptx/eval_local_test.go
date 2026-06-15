package pptx

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
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
			md := PostprocessText(ToMarkdown(deck))

			if strings.TrimSpace(md) == "" {
				t.Fatal("empty output")
			}
			if n := len(h1Re.FindAllString(md, -1)); n != 1 {
				t.Fatalf("want exactly one H1 deck title, got %d", n)
			}
			if len(deck.Slides) == 0 {
				t.Fatal("no slides extracted")
			}

			// Independent oracle: re-derive expected slide numbers straight from
			// the zip's sldIdLst + rels, NOT via Extract, then check extraction
			// matched it. This anchors count/order to the file, not to itself.
			wantNums := independentSlideNumbers(t, path)
			gotNums := make([]int, len(deck.Slides))
			for i, s := range deck.Slides {
				gotNums[i] = s.Number
			}
			if !slices.Equal(wantNums, gotNums) {
				t.Fatalf("slide numbers/order: file=%v extracted=%v", wantNums, gotNums)
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

			withBody := 0
			for _, s := range deck.Slides {
				if len(s.Blocks) > 0 {
					withBody++
				}
			}
			t.Logf("OK: %d slides (%d with body content, %.0f%%), %d bytes of Markdown",
				len(deck.Slides), withBody, 100*float64(withBody)/float64(len(deck.Slides)), len(md))
		})
	}
}

// independentSlideNumbers re-derives the expected slide Numbers (1-based
// sldIdLst positions of resolvable slide relationships whose part exists)
// directly from the zip, bypassing Extract — an independent oracle for the
// slide count and ordering that the extractor's own loop must reproduce.
func independentSlideNumbers(t *testing.T, path string) []int {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	present := make(map[string]bool, len(zr.File))
	read := func(name string) []byte {
		for _, f := range zr.File {
			if f.Name == name {
				rc, err := f.Open()
				if err != nil {
					t.Fatal(err)
				}
				defer func() { _ = rc.Close() }()
				b, err := io.ReadAll(rc)
				if err != nil {
					t.Fatal(err)
				}
				return b
			}
		}
		return nil
	}
	for _, f := range zr.File {
		present[f.Name] = true
	}

	var pres presentationXML
	if err := xml.Unmarshal(read("ppt/presentation.xml"), &pres); err != nil {
		t.Fatal(err)
	}
	rels, err := parseRelationships(read("ppt/_rels/presentation.xml.rels"), "ppt")
	if err != nil {
		t.Fatal(err)
	}

	var nums []int
	for i, sid := range pres.SldIDLst {
		rel, ok := rels[sid.RelID]
		if ok && strings.Contains(rel.Type, "/slide") && present[rel.Target] {
			nums = append(nums, i+1)
		}
	}
	return nums
}
