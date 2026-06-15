<p align="center">
  <img src="icon.png" alt="pptx2md — PowerPoint to Markdown" width="391">
</p>

# pptx2md — PowerPoint → Markdown skill for Claude Code

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Claude Code Skill](https://img.shields.io/badge/Claude%20Code-Skill-purple.svg)](https://claude.ai/code)
[![Pure Go](https://img.shields.io/badge/Pure%20Go-no%20deps-00ADD8?logo=go&logoColor=white)](https://go.dev/)

> **Drop a `.pptx`, get clean Markdown.** No pandoc, no Python, no runtime deps.

This is the Claude Code skill that wraps the [`pptx-to-md`](../README.md) engine —
a pure-Go binary that converts PowerPoint files into clean, agent-readable
Markdown. The skill is a thin install + UX wrapper; all conversion logic lives in
the binary built by this repo.

## Why this skill?

| Without it | With it |
|------------|---------|
| Find/build the right binary for your OS | `install.sh` auto-detects OS/arch |
| Remember the CLI flags | Ask Claude in plain English |
| pandoc can't even read `.pptx` | Pure-Go OOXML parser, single static binary |

## What it extracts

- Slide titles → `## Slide N: Title`
- Nested bullets and paragraphs
- Tables → Markdown tables (empty tables dropped)
- Images → `[IMAGE: alt]` text placeholders (no image bytes; output targets agents)
- Speaker notes → `> **Notes:** …`
- Slides in true presentation order, separated by `---`

## Install

```
/plugin marketplace add greenstevester/pptx2md-go
/plugin install pptx2md@pptx2md
```

Restart Claude Code. (The first command only registers the marketplace; the
second installs the plugin.)

**Verify:** ask Claude *"What pptx skills do you have?"*

The conversion binary is fetched on first use by `install.sh` (downloads the
platform build from this repo's GitHub Releases and verifies its sha256). To
install it yourself:

```bash
curl -fsSL https://raw.githubusercontent.com/greenstevester/pptx2md-go/main/skill/install.sh | bash
```

> **Status:** the engine has no published GitHub release yet, so `install.sh`
> reports it can't find a release until the first one is tagged. The script
> itself is complete and correct.

## Update

```
/plugin marketplace update pptx2md
```

## Local development

```bash
claude --plugin-dir /path/to/pptx2md-go
```

## Usage

In a directory with a `.pptx`, ask Claude naturally — or run the binary:

```bash
pptx-to-md deck.pptx                # writes deck.md
pptx-to-md deck.pptx --stdout       # prints to stdout
pptx-to-md postprocess deck.md      # re-run the cleanup pass
```

## License

MIT — see [LICENSE](LICENSE). Engine: [pptx2md-go](https://github.com/greenstevester/pptx2md-go).
