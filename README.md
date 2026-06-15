# pptx2md-go

A pure-Go CLI that converts PowerPoint (`.pptx`) files into clean, agent-readable Markdown.

Sibling of [word-doc-to-md-skill-go](https://github.com/greenstevester/word-doc-to-md-skill-go). Unlike the docx tool, this needs **no pandoc and no external dependencies** — a `.pptx` is a zip of XML, so conversion happens entirely in-process. The result is a single static binary per platform.

## Background

Every `.pptx`-to-Markdown converter I could find was Python-based. Rather than drag a Python runtime and its dependency tree into an agent toolchain, I wrote one in Go, straight from the spec.

The `.pptx` format is *fully* specified — the catch is that the spec is enormous. It's part of the **Office Open XML (OOXML)** standard that also defines Word and Excel. Canonical references:

- **ECMA-376** — the Office Open XML standard, with the `.xsd` schemas bundled as electronic inserts ([ecma-international.org](https://ecma-international.org/publications-and-standards/standards/ecma-376/))
- **ISO/IEC 29500** — the ISO/IEC ratification of the same standard ([iso.org catalogue](https://www.iso.org/standard/71691.html); the current edition is free via [ISO ITTF](https://standards.iso.org/ittf/PubliclyAvailableStandards/))
- **Open XML SDK** documentation ([learn.microsoft.com](https://learn.microsoft.com/en-us/office/open-xml/open-xml-sdk))
- The complete **XML schemas (`.xsd`)** for OOXML ship as normative inserts to the standard (and with the Open XML SDK); a browsable copy lives at [schemas.liquid-technologies.com](https://schemas.liquid-technologies.com/officeopenxml/2006/)

ECMA-376 alone runs to several thousand pages across four parts: (1) fundamentals & packaging, (2) Open Packaging Conventions — the ZIP container, (3) markup language reference, and (4) transitional migration features.

Under the hood a `.pptx` is just a ZIP of XML documents linked through relationship (`.rels`) files:

```
presentation.pptx
├── [Content_Types].xml
├── _rels/
├── ppt/
│   ├── presentation.xml
│   ├── slides/        slide1.xml, slide2.xml, …
│   ├── slideLayouts/  slideMasters/  theme/
│   ├── media/         charts/        notesSlides/
└── docProps/
```

Coverage is a long tail:

- **Phase 1 (~70–80% of decks):** open the ZIP, parse `[Content_Types].xml` and `presentation.xml`, follow relationships, read slide masters / layouts / slides, and render text boxes, images, shapes, fills, and fonts.
- **Phase 2 (~90%):** tables, grouped shapes, theme inheritance, bullets & numbering, hyperlinks, notes, and (optionally) animations.

This tool isn't a renderer — it targets agent-readable Markdown, so it pulls the **semantic** content (titles, bullets, tables, notes, image alt text) and deliberately drops pure presentation (themes, layout, animation, image bytes).

## What it extracts

- Slide titles → `## Slide N: Title`
- Bullets (nested by indent level) and paragraphs
- Tables → Markdown tables
- Images → `[IMAGE: alt]` text placeholders (no image bytes; output targets AI agents)
- Speaker notes → `> **Notes:** …`
- Slides in true presentation order, separated by `---`

## Usage

```bash
pptx-to-md deck.pptx                # writes deck.md
pptx-to-md deck.pptx out.md         # writes out.md
pptx-to-md deck.pptx --stdout       # writes to stdout
pptx-to-md postprocess deck.md      # re-run the cleanup pass on existing Markdown
```

## Build & test

```bash
make build    # -> build/pptx-to-md
make test     # go test -race ./...
make ci       # full local CI pipeline
```

## Releases

Push to `main` auto-tags (BREAKING→major, feat→minor, fix→patch) and runs GoReleaser to publish platform binaries (linux/macos/windows × amd64/arm64), consumed by the skill package's installer.
