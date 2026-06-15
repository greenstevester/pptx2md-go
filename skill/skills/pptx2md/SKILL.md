---
name: pptx2md
description: Convert PowerPoint (.pptx) files to clean, agent-readable Markdown
argument-hint: [input.pptx] [output.md|--stdout]
allowed-tools: Bash, Read, Write
---

## PowerPoint → Markdown

```
┌─────────────────────────────────────────────────────────────────┐
│  .pptx  ──▶  pptx-to-md  ──▶  clean, agent-readable Markdown      │
│                                                                   │
│  • titles → ## Slide N: Title      • tables → Markdown tables      │
│  • nested bullets & paragraphs     • images → [IMAGE: alt]         │
│  • speaker notes → > **Notes:**    • true slide order, --- between │
│                                                                   │
│  Pure Go binary. No pandoc, no Python, no runtime dependencies.   │
└─────────────────────────────────────────────────────────────────┘
```

### Pre-flight
- Binary: !`command -v pptx-to-md >/dev/null 2>&1 && echo "✓ installed ($(command -v pptx-to-md))" || echo "✗ not installed — do Step 1"`
- Decks here: !`ls *.pptx 2>/dev/null | head -5 || echo "none in current directory"`

### Target: ${ARGUMENTS:-(pass a .pptx path, e.g. deck.pptx)}

---

## Step 1: Install the binary (first run only)

If the pre-flight shows the binary is missing, install the right build for your
platform (downloads from GitHub Releases and verifies its sha256 checksum):

```bash
curl -fsSL https://raw.githubusercontent.com/greenstevester/powerpoint-to-md-skill-go/main/skill/install.sh | bash
```

Installs to `~/.local/bin` by default (override with `PPTX2MD_BIN_DIR`). If that
directory is not on your `PATH`, add it: `export PATH="$HOME/.local/bin:$PATH"`.

> **Windows:** download the `_windows_amd64.zip` from the
> [releases page](https://github.com/greenstevester/powerpoint-to-md-skill-go/releases/latest)
> and put `pptx-to-md.exe` on your `PATH`.

---

## Step 2: Convert

```bash
pptx-to-md deck.pptx                # writes deck.md
pptx-to-md deck.pptx out.md         # writes out.md
pptx-to-md deck.pptx --stdout       # prints Markdown to stdout
pptx-to-md postprocess deck.md      # re-run the cleanup pass on existing Markdown
```

---

## What you get

| PowerPoint element | Markdown output |
|--------------------|-----------------|
| Slide title | `## Slide N: Title` (or `## Slide N` if untitled) |
| Bullets | `- item`, nested by indent level |
| Body text | paragraphs |
| Table | Markdown table (empty tables are dropped) |
| Image | `[IMAGE: alt]` placeholder — no image bytes (output targets AI agents) |
| Speaker notes | `> **Notes:** …` (slide-number placeholder excluded) |
| Slide order | true presentation order, slides separated by `---` |

The deck title comes from `docProps/core.xml` `<dc:title>`, else the first slide's
title, else the filename stem.
