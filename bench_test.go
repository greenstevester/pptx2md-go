package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// zipParts assembles an in-memory zip from name→content parts (benchmark helper;
// panics on error since benchmarks have no *testing.T to fail).
func zipParts(parts map[string]string) []byte {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, content := range parts {
		w, err := zw.Create(name)
		if err != nil {
			panic(err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			panic(err)
		}
	}
	if err := zw.Close(); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func genSlideXML(i int) string {
	return fmt.Sprintf(`<p:sld xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"><p:cSld><p:spTree>`+
		`<p:sp><p:nvSpPr><p:cNvPr id="1" name="Title"/><p:nvPr><p:ph type="title"/></p:nvPr></p:nvSpPr><p:txBody><a:p><a:r><a:t>Slide %d Title</a:t></a:r></a:p></p:txBody></p:sp>`+
		`<p:sp><p:nvSpPr><p:cNvPr id="2" name="Body"/><p:nvPr><p:ph type="body"/></p:nvPr></p:nvSpPr><p:txBody>`+
		`<a:p><a:pPr lvl="0"/><a:r><a:t>First bullet on slide %d</a:t></a:r></a:p>`+
		`<a:p><a:pPr lvl="1"/><a:r><a:t>Nested detail point with some descriptive text</a:t></a:r></a:p>`+
		`<a:p><a:r><a:t>A plain paragraph of body text describing this slide in a sentence.</a:t></a:r></a:p>`+
		`</p:txBody></p:sp>`+
		`</p:spTree></p:cSld></p:sld>`, i, i)
}

// genDeck builds an in-memory N-slide .pptx for reproducible large-deck benchmarks.
func genDeck(n int) []byte {
	parts := make(map[string]string, n+2)
	var sldIds, rels strings.Builder
	rels.WriteString(`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`)
	for i := 1; i <= n; i++ {
		rid := fmt.Sprintf("rId%d", i)
		sldIds.WriteString(fmt.Sprintf(`<p:sldId r:id="%s"/>`, rid))
		rels.WriteString(fmt.Sprintf(`<Relationship Id="%s" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide" Target="slides/slide%d.xml"/>`, rid, i))
		parts[fmt.Sprintf("ppt/slides/slide%d.xml", i)] = genSlideXML(i)
	}
	rels.WriteString(`</Relationships>`)
	parts["ppt/presentation.xml"] = `<p:presentation xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><p:sldIdLst>` + sldIds.String() + `</p:sldIdLst></p:presentation>`
	parts["ppt/_rels/presentation.xml.rels"] = rels.String()
	return zipParts(parts)
}

const benchSlides = 500

// BenchmarkConvertSmallFixture: full pipeline on a real (rich) PowerPoint file.
func BenchmarkConvertSmallFixture(b *testing.B) {
	data, err := os.ReadFile(filepath.Join("testdata", "fixtures", "real-rich.pptx"))
	if err != nil {
		b.Fatal(err)
	}
	r := bytes.NewReader(data)
	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		deck, err := Extract(r, int64(len(data)))
		if err != nil {
			b.Fatal(err)
		}
		_ = postprocessText(ToMarkdown(deck))
	}
}

// BenchmarkExtractLarge: extraction stage only, large generated deck.
func BenchmarkExtractLarge(b *testing.B) {
	data := genDeck(benchSlides)
	r := bytes.NewReader(data)
	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Extract(r, int64(len(data))); err != nil {
			b.Fatal(err)
		}
	}
	b.ReportMetric(float64(benchSlides), "slides/op")
}

// BenchmarkRenderLarge: render+postprocess stage only, large generated deck.
func BenchmarkRenderLarge(b *testing.B) {
	data := genDeck(benchSlides)
	deck, err := Extract(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = postprocessText(ToMarkdown(deck))
	}
	b.ReportMetric(float64(benchSlides), "slides/op")
}

// BenchmarkConvertLarge: full pipeline on the large generated deck.
func BenchmarkConvertLarge(b *testing.B) {
	data := genDeck(benchSlides)
	r := bytes.NewReader(data)
	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		deck, err := Extract(r, int64(len(data)))
		if err != nil {
			b.Fatal(err)
		}
		_ = postprocessText(ToMarkdown(deck))
	}
	b.ReportMetric(float64(benchSlides), "slides/op")
}

// BenchmarkConvertRealWorld: full pipeline on a real local deck if present.
// Set PPTX_BENCH_FILE to override; skips when absent (nothing committed).
func BenchmarkConvertRealWorld(b *testing.B) {
	path := filepath.Join("testdata", "local", "complex-startupx.pptx")
	if p := os.Getenv("PPTX_BENCH_FILE"); p != "" {
		path = p
	}
	data, err := os.ReadFile(path)
	if err != nil {
		b.Skipf("no real-world deck at %s (%v)", path, err)
	}
	r := bytes.NewReader(data)
	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		deck, err := Extract(r, int64(len(data)))
		if err != nil {
			b.Fatal(err)
		}
		_ = postprocessText(ToMarkdown(deck))
	}
}
