# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

CLI tool that converts PowerPoint presentations (.pptx) to clean, agent-readable Markdown. Pure Go, no external dependencies — a .pptx is an OPC zip of XML, parsed in-process with `archive/zip` + `encoding/xml`. (Note: pandoc cannot read .pptx, so unlike the docx sibling there is no pandoc/bootstrap layer.)

This repo builds the platform-specific binaries (Windows, macOS, Linux × amd64/arm64) consumed by the PowerPoint-to-Markdown skill package. The installer detects OS/arch and selects the correct binary from this repo's releases.

Module name: `pptx-to-agent-md` | Go 1.24+ (CI runs 1.26.1) | Single `main` package.

## Build & Test Commands

```bash
make build                       # Build for current platform -> build/pptx-to-md
make build-all                   # Cross-compile all 5 targets -> build/
make test                        # Run all tests with -race
make lint                        # golangci-lint
make ci                          # Full CI pipeline locally
go test -run TestRenderTable     # Run a single test by name
```

## Architecture

Three-stage in-process pipeline, all in package `main`:

1. **pptx.go** — opens the zip, resolves slide order from `ppt/presentation.xml` `<p:sldIdLst>` via `ppt/_rels/presentation.xml.rels` (NOT lexical filename order), and extracts title/body/tables/images/notes into a `Deck` model. `relationships.go` resolves `.rels` targets.
2. **render.go** — renders the `Deck` to Markdown: `## Slide N: Title`, bullets nested by level, Markdown tables, `[IMAGE: alt]` placeholders, speaker notes as `> **Notes:**`, `---` between slides.
3. **postprocess.go** — light cleanup (normalize CRLF, trim trailing whitespace, collapse blank lines).

`main.go` dispatches subcommands: default (full convert via `convert.go`), `postprocess` (postprocess-only). Output goes to file by default or stdout with `--stdout`.

## Key Design Decisions

- Pure Go OOXML parsing — no pandoc, no system dependency, no bootstrap.
- Images are intentionally replaced with `[IMAGE: alt]` text placeholders (output targets AI agents, not humans).
- Slides are emitted in true presentation order; block order within a slide is canonical (title → body → tables → images), since PowerPoint shape order is z-order, not reading order.
- Speaker notes exclude the slide-number placeholder.

## Provenance

Parser core lifted from a POC (`pptx2md-go`) and conformed to this repo's conventions. See `docs/superpowers/specs/2026-06-15-pptx-to-md-design.md`.
