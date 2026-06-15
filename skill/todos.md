# pptx2md skill — TODOs

## Blocking (install.sh can't fetch a binary until these are done)

- [x] Push the engine repo to `github.com/greenstevester/pptx2md-go`
- [x] Tag the first release so GoReleaser publishes platform binaries +
      `checksums.txt` (v0.0.1 / v0.0.2 published)

## Polish

- [x] Add an `icon.png` (PPTX → MD banner logo)
- [x] Root-level `.claude-plugin/marketplace.json` so `/plugin marketplace add
      greenstevester/pptx2md-go` resolves (source → `skill/skills/pptx2md`)
- [x] Smoke-test `install.sh` end-to-end (verified darwin/arm64 pull + checksum + run against v0.0.1)
- [ ] Windows: verify the manual `.zip` instructions; consider a `.ps1` installer

## Skill registries (queued — needs eligibility)

**[awesome-claude-skills](https://github.com/travisvn/awesome-claude-skills)** has
two hard gates we don't meet yet (per its `CONTRIBUTING.md`):

- **≥10 GitHub stars** — otherwise the PR is auto-closed
- **No AI-generated/-submitted PRs** — must be opened by a human

Once pptx2md-go has 10+ stars, a human opens a PR adding this entry under
`## 🌟 Community Skills` → `### Individual Skills` (ready to paste):

    - **[pptx2md](https://github.com/greenstevester/pptx2md-go)** - Pure-Go CLI + Claude Code skill converting PowerPoint (.pptx) to clean, agent-readable Markdown — no pandoc, no Python

(fastlane-skill's PR #36 to this same list was closed without merge — same gates.)

## Maybe

- [ ] Mirror as a standalone `pptx2md` skill repo if the two-repo split
      (like word-doc-to-md-skill) is preferred later
