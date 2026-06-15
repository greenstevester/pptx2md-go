# pptx2md skill — TODOs

## Blocking (install.sh can't fetch a binary until these are done)

- [x] Push the engine repo to `github.com/greenstevester/pptx2md-go`
- [ ] Tag the first release so GoReleaser publishes platform binaries +
      `checksums.txt` (the assets `install.sh` downloads)

## Polish

- [x] Add an `icon.png` (PPTX → MD banner logo)
- [x] Root-level `.claude-plugin/marketplace.json` so `/plugin marketplace add
      greenstevester/pptx2md-go` resolves (source → `skill/skills/pptx2md`)
- [ ] Smoke-test `install.sh` end-to-end once a release exists (darwin/linux × amd64/arm64)
- [ ] Windows: verify the manual `.zip` instructions; consider a `.ps1` installer

## Maybe

- [ ] Submit to a Claude Code skills marketplace once released
- [ ] Mirror as a standalone `pptx2md` skill repo if the two-repo split
      (like word-doc-to-md-skill) is preferred later
