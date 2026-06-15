# Design: pptx2md Claude Code skill

**Date:** 2026-06-15
**Status:** Approved (design phase)

## Summary

A Claude Code skill plugin that wraps the `pptx-to-md` engine binary, structured
like [`fastlane-skill`](https://github.com/greenstevester/fastlane-skill) but
nested inside this engine repo under `skill/`. It is a thin install + UX wrapper:
the actual conversion logic lives in the Go binary built by this repo; the skill
installs the right binary for the user's platform and documents how to use it.

Decisions (from brainstorming):
- **Location:** in-repo at `pptx2md-go/skill/` (not a separate repo).
- **One combined skill** named `pptx2md` (install-if-missing → convert).
- **Names:** marketplace `pptx2md`, plugin `pptx2md`, skill `pptx2md`; binary stays `pptx-to-md`.
- **Install:** download the release binary by OS/arch + sha256 verify (no Homebrew —
  unlike fastlane, this tool ships as a static binary via GitHub Releases).

## Layout

```
skill/
  .claude-plugin/
    marketplace.json    # marketplace "pptx2md"; one plugin "pptx2md", source ./skills/pptx2md
    plugin.json         # name/version/description/author/repo/license (MIT)
  skills/
    pptx2md/
      SKILL.md          # the combined skill
  install.sh            # OS+arch detect → download release asset → sha256 verify → install
  README.md             # badges + why + install + usage (fastlane-style)
  CLAUDE.md             # guidance for working on the skill
  LICENSE               # MIT (copied from engine repo)
  todos.md              # roadmap
```

## SKILL.md

Frontmatter: `name: pptx2md`, `description`, `argument-hint: [input.pptx] [output.md|--stdout]`,
`allowed-tools: Bash, Read, Write`.

Body:
- **Pre-flight** (`!` inline checks): is `pptx-to-md` on PATH? list `*.pptx` in cwd.
- **Step 1 — install (first run only):** `curl -fsSL <raw>/skill/install.sh | bash`
  (matching the word-doc-to-md-skill convention), or run the bundled `skill/install.sh`.
- **Step 2 — convert:** the four invocations (`deck.pptx` → `deck.md`; explicit out;
  `--stdout`; `postprocess deck.md`).
- **What it extracts:** `## Slide N: Title`, nested bullets, tables, `[IMAGE: alt]`
  placeholders, `> **Notes:**`, true presentation order, `---` separators.

## install.sh

- `set -euo pipefail`.
- OS: `uname -s` → `darwin`/`linux`; Windows → printed manual-download instructions.
- Arch: `uname -m` → `amd64` (x86_64) / `arm64` (arm64|aarch64).
- Version: `PPTX2MD_VERSION` env override, else latest tag from the GitHub API.
- Asset: `pptx-to-md_<version>_<os>_<arch>.tar.gz` (matches `.goreleaser.yml`).
- Download from `releases/download/<tag>/`, fetch `checksums.txt`, verify sha256
  (skip with a warning if unavailable), extract, `install -m 0755` to
  `${PPTX2MD_BIN_DIR:-$HOME/.local/bin}`, PATH note, run `--help` to verify.

## Caveats (documented in README + todos)

1. **No GitHub releases exist yet** (engine not pushed/tagged). `install.sh` is
   correct but only succeeds once a release is published; until then it errors
   clearly. Same for the `curl | bash` raw URL.
2. Because `.claude-plugin/` is nested under `skill/`, the GitHub shorthand
   `/plugin marketplace add <owner>/<repo>` will not auto-resolve it; install via
   `claude --plugin-dir .../skill` or `/plugin marketplace add <path>/skill`.
3. No `icon.png` is committed (won't fabricate a binary image) — tracked in todos.

## Verification

- `claude --plugin-dir ./skill` loads the plugin and lists `pptx2md`.
- `shellcheck install.sh` clean.
- Dry-run `install.sh` OS/arch detection resolves the correct asset name and
  fails gracefully on the (currently absent) release.

## Scope check

Single focused plan. Static files + one shell script. No decomposition needed.
