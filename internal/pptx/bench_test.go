package pptx

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Benchmarks for the convert pipeline.
//
// MB/s is intentionally NOT reported: most of a real .pptx is media/chart bytes
// the converter never parses, so bytes/sec against file size overstates real
// throughput. slides/sec is the honest metric.
//
// The real-world benchmark converts a large local deck in ~1.5s, so the default
// 1s benchtime yields a single noisy sample. For a stable figure run:
//
//	go test -run '^$' -bench BenchmarkConvertRealWorld -benchtime=10x -count=5 -benchmem

const (
	nsP = "http://schemas.openxmlformats.org/presentationml/2006/main"
	nsA = "http://schemas.openxmlformats.org/drawingml/2006/main"
	nsR = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"
	// runFmt/shapeFmt are the formatting subtrees the converter does NOT map;
	// encoding/xml.Skip walks them, which dominates real-deck parse cost.
	runFmt   = `<a:rPr lang="en-US" sz="1800" b="1"><a:solidFill><a:srgbClr val="1F2937"/></a:solidFill><a:latin typeface="Arial"/></a:rPr>`
	shapeFmt = `<p:spPr><a:xfrm><a:off x="100" y="100"/><a:ext cx="500" cy="400"/></a:xfrm><a:prstGeom prst="rect"><a:avLst/></a:prstGeom></p:spPr>`
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

// genSlideXML builds a realistically heavy slide: a formatted title, six
// formatted bullets at varying levels, a picture, run/shape formatting subtrees,
// a timing tree, and (every 5th slide) a table — so the benchmark exercises the
// same encoding/xml.Skip work that dominates real decks, not a best-case stub.
func genSlideXML(i int) string {
	si := strconv.Itoa(i)
	var bullets strings.Builder
	for b := 0; b < 6; b++ {
		bullets.WriteString(`<a:p><a:pPr lvl="` + strconv.Itoa(b%3) + `"><a:buChar char="-"/></a:pPr><a:r>` + runFmt +
			`<a:t>Bullet ` + strconv.Itoa(b) + ` on slide ` + si + ` with some descriptive content.</a:t></a:r></a:p>`)
	}
	pic := `<p:pic><p:nvPicPr><p:cNvPr id="9" name="Picture ` + si + `" descr="Diagram for slide ` + si +
		`"/><p:cNvPicPr/><p:nvPr/></p:nvPicPr><p:blipFill><a:blip r:embed="rId1"/></p:blipFill>` + shapeFmt + `</p:pic>`
	table := ""
	if i%5 == 0 {
		table = `<p:graphicFrame><a:graphic><a:graphicData><a:tbl>` +
			`<a:tr><a:tc><a:txBody><a:p><a:r><a:t>Metric</a:t></a:r></a:p></a:txBody></a:tc><a:tc><a:txBody><a:p><a:r><a:t>Value</a:t></a:r></a:p></a:txBody></a:tc></a:tr>` +
			`<a:tr><a:tc><a:txBody><a:p><a:r><a:t>Revenue</a:t></a:r></a:p></a:txBody></a:tc><a:tc><a:txBody><a:p><a:r><a:t>Up 12%</a:t></a:r></a:p></a:txBody></a:tc></a:tr>` +
			`</a:tbl></a:graphicData></a:graphic></p:graphicFrame>`
	}
	return `<p:sld xmlns:p="` + nsP + `" xmlns:a="` + nsA + `" xmlns:r="` + nsR + `"><p:cSld><p:spTree>` +
		`<p:sp><p:nvSpPr><p:cNvPr id="1" name="Title"/><p:nvPr><p:ph type="title"/></p:nvPr></p:nvSpPr>` + shapeFmt +
		`<p:txBody><a:p><a:r>` + runFmt + `<a:t>Slide ` + si + ` Title</a:t></a:r></a:p></p:txBody></p:sp>` +
		`<p:sp><p:nvSpPr><p:cNvPr id="2" name="Body"/><p:nvPr><p:ph type="body"/></p:nvPr></p:nvSpPr>` + shapeFmt +
		`<p:txBody>` + bullets.String() + `</p:txBody></p:sp>` +
		pic + table +
		`</p:spTree></p:cSld>` +
		`<p:timing><p:tnLst><p:par><p:cTn id="1" dur="indefinite"/></p:par></p:tnLst></p:timing>` +
		`</p:sld>`
}

func genNotesXML(i int) string {
	si := strconv.Itoa(i)
	return `<p:notes xmlns:p="` + nsP + `" xmlns:a="` + nsA + `"><p:cSld><p:spTree>` +
		`<p:sp><p:nvSpPr><p:cNvPr id="2" name="Notes"/><p:nvPr><p:ph type="body"/></p:nvPr></p:nvSpPr>` +
		`<p:txBody><a:p><a:r><a:t>Speaker notes for slide ` + si + `: remember to mention the key point.</a:t></a:r></a:p></p:txBody></p:sp>` +
		`<p:sp><p:nvSpPr><p:cNvPr id="3" name="Num"/><p:nvPr><p:ph type="sldNum"/></p:nvPr></p:nvSpPr>` +
		`<p:txBody><a:p><a:t>` + si + `</a:t></a:p></p:txBody></p:sp>` +
		`</p:spTree></p:cSld></p:notes>`
}

// genDeck builds an in-memory N-slide .pptx (slides + per-slide rels + notes)
// for reproducible large-deck benchmarks with nothing committed.
func genDeck(n int) []byte {
	parts := make(map[string]string, n*3+2)
	var sldIds, rels strings.Builder
	rels.WriteString(`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`)
	for i := 1; i <= n; i++ {
		si := strconv.Itoa(i)
		rid := "rId" + si
		sldIds.WriteString(`<p:sldId r:id="` + rid + `"/>`)
		rels.WriteString(`<Relationship Id="` + rid + `" Type="` + nsR + `/slide" Target="slides/slide` + si + `.xml"/>`)
		parts["ppt/slides/slide"+si+".xml"] = genSlideXML(i)
		parts["ppt/slides/_rels/slide"+si+".xml.rels"] = `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` +
			`<Relationship Id="rIdN" Type="` + nsR + `/notesSlide" Target="../notesSlides/notesSlide` + si + `.xml"/></Relationships>`
		parts["ppt/notesSlides/notesSlide"+si+".xml"] = genNotesXML(i)
	}
	rels.WriteString(`</Relationships>`)
	parts["ppt/presentation.xml"] = `<p:presentation xmlns:p="` + nsP + `" xmlns:r="` + nsR + `"><p:sldIdLst>` + sldIds.String() + `</p:sldIdLst></p:presentation>`
	parts["ppt/_rels/presentation.xml.rels"] = rels.String()
	return zipParts(parts)
}

const benchSlides = 500

func reportSlidesPerSec(b *testing.B, slides int) {
	b.ReportMetric(float64(slides)*float64(b.N)/b.Elapsed().Seconds(), "slides/sec")
}

// BenchmarkConvertSmallFixture: full pipeline on a real (rich) PowerPoint file.
func BenchmarkConvertSmallFixture(b *testing.B) {
	data, err := os.ReadFile(filepath.Join("testdata", "fixtures", "real-rich.pptx"))
	if err != nil {
		b.Fatal(err)
	}
	r := bytes.NewReader(data)
	deck0, err := Extract(r, int64(len(data)))
	if err != nil {
		b.Fatal(err)
	}
	slides := len(deck0.Slides)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		deck, err := Extract(r, int64(len(data)))
		if err != nil {
			b.Fatal(err)
		}
		_ = PostprocessText(ToMarkdown(deck))
	}
	reportSlidesPerSec(b, slides)
}

// BenchmarkExtractLarge: extraction stage only, large generated deck.
func BenchmarkExtractLarge(b *testing.B) {
	data := genDeck(benchSlides)
	r := bytes.NewReader(data)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Extract(r, int64(len(data))); err != nil {
			b.Fatal(err)
		}
	}
	reportSlidesPerSec(b, benchSlides)
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
		_ = PostprocessText(ToMarkdown(deck))
	}
	reportSlidesPerSec(b, benchSlides)
}

// BenchmarkConvertLarge: full pipeline on the large generated deck.
func BenchmarkConvertLarge(b *testing.B) {
	data := genDeck(benchSlides)
	r := bytes.NewReader(data)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		deck, err := Extract(r, int64(len(data)))
		if err != nil {
			b.Fatal(err)
		}
		_ = PostprocessText(ToMarkdown(deck))
	}
	reportSlidesPerSec(b, benchSlides)
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
	deck0, err := Extract(r, int64(len(data)))
	if err != nil {
		b.Fatal(err)
	}
	slides := len(deck0.Slides)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		deck, err := Extract(r, int64(len(data)))
		if err != nil {
			b.Fatal(err)
		}
		_ = PostprocessText(ToMarkdown(deck))
	}
	reportSlidesPerSec(b, slides)
}
