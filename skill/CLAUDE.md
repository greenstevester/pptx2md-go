# CLAUDE.md

Guidance for Claude Code when working on the `pptx2md` skill (this `skill/`
subdirectory). For the engine binary, see `../CLAUDE.md`.

## Project Overview

A Claude Code skill plugin that wraps the `pptx-to-md` engine binary. It is a
thin install + UX wrapper, not a codebase with build/test steps — the conversion
logic lives in the Go binary built by the parent repo. Structured like
[`fastlane-skill`](https://github.com/greenstevester/fastlane-skill), nested
inside the engine repo.

## Structure

```
.claude-plugin/
  marketplace.json   # marketplace "pptx2md"; plugin "pptx2md", source ./skills/pptx2md
  plugin.json        # plugin metadata
skills/
  pptx2md/
    SKILL.md         # the combined skill (install-if-missing → convert)
install.sh           # OS/arch detect → download release asset → sha256 verify → install
todos.md             # roadmap
```

## SKILL.md format

- **Frontmatter** (YAML): `name`, `description`, `argument-hint`, `allowed-tools`.
- **Inline commands**: `` !`cmd` `` runs a shell command when the skill loads
  (used for pre-flight checks).
- **Placeholders**: `${ARGUMENTS}` is replaced with the user's arguments.

## Testing

No build step.

```bash
claude --plugin-dir /path/to/pptx2md-go/skill   # load locally
shellcheck install.sh                                          # lint the installer
```

Then run the `pptx2md` skill in a directory containing a `.pptx`.

## Key details

- Unlike fastlane-skill (Homebrew), this installs a static binary from GitHub
  Releases. `install.sh` asset names must stay in sync with `../.goreleaser.yml`
  (`pptx-to-md_<version>_<os>_<arch>.tar.gz`).
- The plugin manifest is in this subdirectory, so the GitHub-shorthand
  `/plugin marketplace add owner/repo` will not auto-resolve it — install by path.
- Keep the skill's documented behaviour in sync with the engine CLI
  (`../main.go`): positional `<in.pptx> [out.md] [--stdout]` + `postprocess`.
