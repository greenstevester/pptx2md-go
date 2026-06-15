# Design: pptx2md-go

**Date:** 2026-06-15
**Status:** Approved (design phase)

## Summary

A CLI tool that converts PowerPoint files (`.pptx`) into clean, agent-readable
Markdown. It is the sibling of [`word-doc-to-md-skill-go`](https://github.com/greenstevester/word-doc-to-md-skill-go):
same repo shape, same build/release/skill-packaging machinery, same
"output targets AI agents, not humans" philosophy — but a different conversion
engine, because **pandoc cannot read `.pptx`** (it has a pptx *writer* only).

A `.pptx` is an OPC zip of XML, so conversion is done **in-process in pure Go**
(`archive/zip` + `encoding/xml`). This removes the entire pandoc download/extract
layer that exists in the docx repo — there is no `bootstrap.go`, no external
runtime dependency, and the result is a single static `CGO_ENABLED=0` binary.

The parser core is **lifted from an existing working POC** (`pptx2md-go`, which
already solves the hard problems correctly) and conformed to the docx repo's
conventions. See "Provenance" below.

## Goals

- Convert `.pptx` → semantic Markdown suitable for AI-agent consumption.
- Zero runtime dependencies; single cross-compiled binary per platform.
- Mirror `word-doc-to-md-skill-go` so its release pipeline and skill packaging
  transfer with minimal change.

## Non-goals

- Not a renderer. Layout, styling, themes, transitions, animations, gradients,
  positioning, charts-as-images, and master/layout placeholder inheritance are
  out of scope.
- No image bytes extracted — images become text placeholders.

## Decisions (locked)

| Decision | Choice | Rationale |
|---|---|---|
| Engine | Pure Go OOXML parser | Pandoc can't read pptx; pptx is a zip of XML. Hermetic single binary. |
| Layout | Flat single `main` package | Matches docx repo; release config copy-pastes; both engines share one mental model. |
| Images | `[IMAGE: alt]` text placeholders, no media extraction | Output targets agents; consistent with docx repo's explicit policy. |
| CLI | Positional `<in.pptx> [out.md] [--stdout]` + `postprocess` subcommand | Symmetric with the docx tool's CLI. |
| Module / binary | `pptx-to-agent-md` / `pptx-to-md` | Mirrors `docx-to-agent-md` / `docx-to-md`. |
| Hidden slides | Included | Agents usually want all content. |
| Go version | 1.24+ (CI on 1.26.1) | Matches docx repo. |

## Architecture

Three-stage pipeline, single `main` package:

```
.pptx ──▶ [extract] ──▶ Deck model ──▶ [render] ──▶ raw .md ──▶ [postprocess] ──▶ clean .md
         pptx.go +                    render.go                 postprocess.go
         relationships.go
```

| File | Responsibility |
|------|----------------|
| `main.go` | CLI dispatch + arg parsing + usage. Subcommands: *default* (convert), `postprocess` (postprocess-only). No `bootstrap` — nothing to download. Reuses the docx repo's `parseOutputArgs` helper. |
| `pptx.go` | Open the zip, resolve slide order, parse slide/notes XML into the `Deck`/`Slide`/`Block` model. **Lifted** from the POC's `extract.go`. |
| `relationships.go` | Parse `*.rels` parts; resolve relative and absolute (`/`-rooted) targets against a base dir. **Lifted** verbatim from the POC. |
| `model.go` | `Deck` / `Slide` / `Block` types. Pure semantic model — no Markdown in it. |
| `render.go` | `Deck` → Markdown string. Owns *all* presentation concerns (image placeholders, table → md table). **Lifted** from the POC's `markdown.go`, conformed. |
| `postprocess.go` | Light regex cleanup: collapse runs of blank lines, trim trailing whitespace. Powers the `postprocess` subcommand. Much lighter than docx's — we emit clean md ourselves, so there are no grid-table or tracked-change artifacts to scrub. |

### Data model

```go
type Deck  struct { Title string; Slides []Slide }
type Slide struct { Number int; Title string; Blocks []Block; Notes string }
type Block struct {
    Type  string     // "paragraph" | "bullet" | "image" | "table"
    Text  string     // paragraph / bullet text
    Level int        // bullet indent, from the <a:pPr lvl> attribute
    Alt   string     // image: descr, else name
    Rows  [][]string // table cells
}
```

Change from the POC: the model is kept **purely semantic**. The POC stored
images as `Src`/`Alt` (real links) and tables as pre-rendered Markdown strings
inside `extract.go`. Here, extraction stores tables as `Rows` and images as
`Alt` only; all Markdown generation moves into `render.go`. `Src`/`MediaDir`
are dropped entirely (no media extraction).

### Extraction (`pptx.go`)

1. Open zip (`archive/zip`). Read `ppt/presentation.xml` → `<p:sldIdLst>` gives
   slide **order** as `r:id` references. Resolve each against
   `ppt/_rels/presentation.xml.rels` to get the real slide part path.
   **This is the correctness landmine** the POC already handles: slides follow
   presentation order, not a lexical `slide1,slide2,…,slide10` filename sort
   (which mis-sorts and ignores reordered slides).
2. Per slide (`encoding/xml`, matching by local element name so the verbose
   OOXML namespaces don't bite):
   - **Title** ← shape whose placeholder `type` is `title`, `ctrTitle`, or
     `subTitle` (first wins).
   - **Body** ← every other text shape's paragraphs (`<a:p>` → `<a:r>`/`<a:fld>`
     → text), each a `bullet` block indented by its `lvl`, or a `paragraph`
     block when no `<a:pPr>` is present. Empty paragraphs dropped.
   - **Tables** ← `<a:graphicFrame>` → `<a:tbl>` → rows/cells → `Rows`.
   - **Images** ← `<p:pic>` → `Block{Type:"image", Alt: descr|name}`.
   - **Notes** ← follow the slide's rels to the `notesSlide` part; take its body
     text. **Fix vs POC:** skip the slide-number placeholder so a stray page
     number doesn't leak into notes (the POC's `extractAllText` grabs every
     `<a:t>` indiscriminately).
3. **Deck title** ← `docProps/core.xml` `<dc:title>`, else first slide's title,
   else the input filename stem.

Canonical block order is title → body → tables → images. PowerPoint shape order
is z-order, not reading order, so imposing a predictable order is more useful to
an agent than trusting raw XML order.

### Rendering (`render.go`)

```
# <Deck title>

---

## Slide N: <slide title or "Slide N">

- bullet (nested by Level via two-space indent)
paragraph text

[IMAGE: alt]

| col | col |
| --- | --- |
| cell | cell |

> **Notes:** <speaker notes text>

---
```

- Pipe characters in table cells are escaped.
- Notes rendered as a single `> **Notes:** …` blockquote (the approved format).

### CLI (`main.go`)

```
pptx-to-md <input.pptx> [output.md] [--stdout]
pptx-to-md postprocess <input.md> [output.md] [--stdout]
```

Default output path is the input with its extension swapped to `.md`. Errors go
to stderr + exit 1 (not a JSON envelope — it's a CLI matching the sibling tool).
Bad/non-zip input, missing `ppt/presentation.xml`, or a deck with no slides →
clear error. A slide with no title still emits `## Slide N`.

## Error handling

- Non-zip / corrupt file → error from `zip.NewReader`, surfaced with context.
- Missing required part (`presentation.xml`, rels) → `missing <part>` error.
- Empty `sldIdLst` → `presentation has no slides`.
- A slide whose rel can't be resolved is skipped, not fatal.

## Testing

- **Lift the POC's fixtures**: in-tree minimal `.pptx` files under
  `testdata/fixtures/` (`basic.pptx`, `table.pptx`) — real OOXML zips built for
  testing, not binary blobs of real decks.
- **Extraction tests** (`pptx_test.go`): slide ordering via `sldIdLst` (incl. the
  10-slide lexical-sort trap), title vs body, nested bullets by `lvl`, a table,
  an image alt placeholder, speaker notes (and that the slide-number placeholder
  is *not* leaked), and the `dc:title`-absent filename fallback.
- **Golden Markdown test** (`render_test.go`): extract `basic.pptx` → compare to
  a committed golden `.md` under `testdata/`.
- **Postprocess tests** (`postprocess_test.go`): blank-line collapse, trailing
  whitespace trim.
- `make test` runs with `-race`.

## Carried over from the docx repo (adapted, pandoc-stripped)

`Makefile`, `.goreleaser.yml` (`CGO_ENABLED=0`, `-trimpath`, 5 targets,
`pptx-to-md_{version}_{os}_{arch}` archives, checksums), `.github/workflows/ci.yml`
+ `release.yml` (matrix test, lint, security scan, auto-semver on push to main +
GoReleaser), `.gitignore`, `LICENSE`, `README.md`, `CLAUDE.md` — all with pandoc
and bootstrap references removed.

## Provenance

The parser core is lifted from a working POC at `~/Downloads/pptx2md-go`
(module `github.com/steve/pptx2md`), which already:
- chose the pure-Go OOXML engine,
- resolves slide order correctly via `presentation.xml` + rels,
- handles relative/absolute rel targets robustly,
- extracts titles/bullets-with-level/tables/image-alt/notes,
- ships golden-file evals and in-tree fixtures.

It is **conformed** to this repo: flat `main` package (not `cmd/`+`internal/`),
image placeholders (not media links), positional CLI + `postprocess`
subcommand (not `-out`/`-media-dir` flags), bare module name, and the full
docx-style release pipeline added. The notes slide-number leak is fixed.

## Scope check

Single, focused implementation plan. No decomposition needed.
