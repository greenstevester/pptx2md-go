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

Because the plugin manifest lives in this `skill/` subdirectory (not the repo
root), add it by path:

```bash
# Quick local test
claude --plugin-dir /path/to/pptx2md-go/skill

# Or register the marketplace and install
/plugin marketplace add /path/to/pptx2md-go/skill
/plugin install pptx2md@pptx2md
```

Restart Claude Code, then ask: *"What pptx skills do you have?"*

The binary itself is fetched on first use by `install.sh` (downloads the
platform build from this repo's GitHub Releases and verifies its sha256). You can
also run it directly:

```bash
curl -fsSL https://raw.githubusercontent.com/greenstevester/pptx2md-go/main/skill/install.sh | bash
```

> **Status:** the engine has no published GitHub release yet, so `install.sh`
> will report that it can't find a release until the first one is tagged. The
> script itself is complete and correct.

## Usage

In a directory with a `.pptx`, ask Claude naturally — or run the binary:

```bash
pptx-to-md deck.pptx                # writes deck.md
pptx-to-md deck.pptx --stdout       # prints to stdout
pptx-to-md postprocess deck.md      # re-run the cleanup pass
```

## License

MIT — see [LICENSE](LICENSE). Engine: [pptx2md-go](https://github.com/greenstevester/pptx2md-go).
