# pptx2md-go

A pure-Go CLI that converts PowerPoint (`.pptx`) files into clean, agent-readable Markdown.

Sibling of [word-doc-to-md-skill-go](https://github.com/greenstevester/word-doc-to-md-skill-go). Unlike the docx tool, this needs **no pandoc and no external dependencies** — a `.pptx` is a zip of XML, so conversion happens entirely in-process. The result is a single static binary per platform.

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
