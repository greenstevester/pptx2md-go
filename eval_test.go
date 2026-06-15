package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// goldenFixtures are committed, license-safe decks spanning increasing complexity:
// two synthetic (basic, table) and two real PowerPoint-authored files from the
// python-pptx MIT corpus (see testdata/fixtures/PROVENANCE.md).
var goldenFixtures = []string{"basic", "table", "real-simple", "real-rich"}

// renderFixture runs the full conversion pipeline (extract → render → postprocess)
// to a string — exactly what convert() writes to disk.
func renderFixture(t *testing.T, name string) string {
	t.Helper()
	deck, err := ExtractFile(filepath.Join("testdata", "fixtures", name+".pptx"))
	if err != nil {
		t.Fatalf("extract %s: %v", name, err)
	}
	return postprocessText(ToMarkdown(deck))
}

// TestGoldenEvals proves end-to-end conversion is stable and correct by diffing
// each fixture's Markdown against a committed golden file.
//
// Regenerate goldens after an intentional output change with:
//
//	UPDATE_GOLDEN=1 go test -run TestGoldenEvals
func TestGoldenEvals(t *testing.T) {
	update := os.Getenv("UPDATE_GOLDEN") == "1"
	for _, name := range goldenFixtures {
		t.Run(name, func(t *testing.T) {
			got := renderFixture(t, name)
			golden := filepath.Join("testdata", "golden", name+".md")

			if update {
				if err := os.MkdirAll(filepath.Dir(golden), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
					t.Fatal(err)
				}
				t.Logf("wrote %s (%d bytes)", golden, len(got))
				return
			}

			wantBytes, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("read golden (run UPDATE_GOLDEN=1 to create): %v", err)
			}
			if got != string(wantBytes) {
				t.Fatalf("golden mismatch for %s:\n%s", name, lineDiff(string(wantBytes), got))
			}
		})
	}
}

// lineDiff renders a compact line-by-line diff of two multi-line strings.
func lineDiff(want, got string) string {
	wl := strings.Split(want, "\n")
	gl := strings.Split(got, "\n")
	n := len(wl)
	if len(gl) > n {
		n = len(gl)
	}
	var b strings.Builder
	for i := 0; i < n; i++ {
		var w, g string
		if i < len(wl) {
			w = wl[i]
		}
		if i < len(gl) {
			g = gl[i]
		}
		if w != g {
			fmt.Fprintf(&b, "  line %d:\n    - want: %q\n    + got:  %q\n", i+1, w, g)
		}
	}
	return b.String()
}
