<p align="center">
  <img src="skill/icon.png" alt="pptx2md — PowerPoint to Markdown" width="391">
</p>

# pptx2md-go

[![CI](https://github.com/greenstevester/pptx2md-go/actions/workflows/ci.yml/badge.svg)](https://github.com/greenstevester/pptx2md-go/actions/workflows/ci.yml)
[![Release](https://github.com/greenstevester/pptx2md-go/actions/workflows/release.yml/badge.svg)](https://github.com/greenstevester/pptx2md-go/actions/workflows/release.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Pure Go](https://img.shields.io/badge/Pure%20Go-no%20deps-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Claude Code Skill](https://img.shields.io/badge/Claude%20Code-Skill-purple.svg)](skill/)

> **Drop a `.pptx`, get clean Markdown.** Pure Go — no pandoc, no Python, no runtime deps.

## Why pptx2md?

Every `.pptx`-to-Markdown converter I could find needed Python. This one doesn't —
it's pure Go, a single static binary, built straight from the OOXML spec
([the gory details below](#background)).

| Other converters | pptx2md |
|------------------|---------|
| Python runtime + a dependency tree | Single static binary, zero runtime deps |
| pandoc can't even read `.pptx` | Pure-Go OOXML parser, in-process |
| Markdown styled for humans | Semantic Markdown aimed at AI agents |
| `pip install` roulette | One checksum-verified download |

## What it extracts

- Slide titles → `## Slide N: Title`
- Bullets (nested by indent level) and paragraphs
- Tables → Markdown tables
- Images → `[IMAGE: alt]` text placeholders (no image bytes; output targets AI agents)
- Speaker notes → `> **Notes:** …`
- Slides in true presentation order, separated by `---`

## Install

### As a Claude Code skill

```
/plugin marketplace add greenstevester/pptx2md-go
/plugin install pptx2md@pptx2md
```
Restart Claude Code, then ask: *"What pptx skills do you have?"* (Adding the
marketplace only registers it — the second command installs the plugin.) The
conversion binary is downloaded (and checksum-verified) for your platform on
first use.

### As a standalone binary

```bash
curl -fsSL https://raw.githubusercontent.com/greenstevester/pptx2md-go/main/skill/install.sh | bash
```
Detects your OS/arch, pulls the matching build from the
[latest release](https://github.com/greenstevester/pptx2md-go/releases/latest),
and verifies its sha256 — or just grab a tarball/zip from the releases page yourself.

## Usage

```
  ┌─────────┐     ┌──────────────┐     ┌─────────────┐
  │  .pptx  │ ──▶ │  pptx-to-md  │ ──▶ │  clean .md  │
  └─────────┘     └──────────────┘     └─────────────┘
                  extract → render → postprocess
```

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

The engine lives in `internal/pptx` (package `pptx`); the thin CLI entrypoint is `cmd/pptx-to-md`.

## Releases

Push to `main` auto-tags (BREAKING→major, feat→minor, fix→patch) and runs GoReleaser to publish platform binaries (linux/macos/windows × amd64/arm64) + checksums, consumed by the skill installer.

## Background

Sibling of [word-doc-to-md-skill-go](https://github.com/greenstevester/word-doc-to-md-skill-go) —
same "output for AI agents, not humans" philosophy, different engine: pandoc has
a `.pptx` *writer* but can't *read* one, so conversion is done in-process in pure Go.

The `.pptx` format is *fully* specified — the catch is that the spec is enormous.
It's part of the **Office Open XML (OOXML)** standard that also defines Word and
Excel. Canonical references:

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

## License

MIT — see [LICENSE](LICENSE).
