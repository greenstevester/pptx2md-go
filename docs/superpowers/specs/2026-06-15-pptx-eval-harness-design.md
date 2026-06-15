# Design: pptx-to-md eval harness & performance benchmarks

**Date:** 2026-06-15
**Status:** Approved (design phase)

## Summary

Add a test/eval harness that proves `pptx-to-md` works across decks of varying
complexity, and benchmarks that characterise its performance. This is purely
test infrastructure — the converter (`pptx.go`, `render.go`, `postprocess.go`,
`convert.go`, `main.go`) is **not** modified.

Evidence is delivered in three tiers, scaled by what is safe to commit. Anything
committed must be license-safe and free of personal/confidential content; real
large/third-party decks are exercised locally but never committed.

## Goals

- Demonstrate correct conversion on real-world decks of low/medium/high complexity.
- Provide committed, reproducible, diffable evidence (golden files) that runs in CI.
- Characterise performance (throughput, allocations) on a large deck.

## Non-goals

- No changes to conversion behaviour. If an eval reveals a real bug, that is a
  separate fix with its own commit; this work only adds tests/benchmarks.
- No LLM-judge evals — golden files + structural invariants only.

## Tiers

### Tier 1 — Committed golden-file evals (CI-reproducible)

- **Fixtures** (`testdata/fixtures/`): keep synthetic `basic.pptx`, `table.pptx`;
  add 2–3 small `python-pptx` test decks of escalating complexity, downloaded
  from the upstream MIT-licensed repo.
- **Provenance** (`testdata/fixtures/PROVENANCE.md`): for every non-synthetic
  fixture, record source URL, upstream license, and sha256. The synthetic
  fixtures are recorded as repo-authored.
- **Golden output** (`testdata/golden/<name>.md`): committed expected Markdown
  for each fixture.
- **Eval test** (`eval_test.go`): table-driven. For each fixture, run the full
  `convert` pipeline (extract → render → postprocess) to a string and compare to
  its golden file. On mismatch, print a unified line diff. Setting
  `UPDATE_GOLDEN=1` rewrites the golden files instead of asserting.

### Tier 2 — Local-only real-world eval (evidence now, nothing committed)

- **Source**: a gitignored `testdata/local/` directory (added to `.gitignore`),
  overridable with the `PPTX_EVAL_DIR` env var. Operators drop real decks here.
- **Test** (`eval_local_test.go`): for every `*.pptx` found, convert and assert
  **structural invariants** rather than exact text (real content is large and
  unstable):
  1. conversion returns no error and non-empty output;
  2. output starts with a single `# ` deck title;
  3. the number of `## Slide N:` headings equals the deck's `sldIdLst` slide count;
  4. no `![` image links and no `media/` paths leak (image policy holds);
  5. no raw OOXML tags (`<a:`, `<p:`) survive into the Markdown;
  6. no rendered note is a bare slide number (slide-number-leak guard).
- If the directory is empty/absent, the test calls `t.Skip` so CI stays green and
  no sensitive content is ever required.
- For *this* delivery, the three real local decks (`Venkat` simple, `focus-math`
  medium, `startup_x` 484-slide complex) are copied into `testdata/local/` to
  produce real evidence, then left uncommitted.

### Tier 3 — Performance benchmarks

- **Bench test** (`bench_test.go`):
  - `BenchmarkConvertBasic` — small committed fixture, end-to-end.
  - `BenchmarkExtractLarge` / `BenchmarkConvertLarge` — an in-memory generated
    deck of N slides (default ~500), built with the existing `buildZip` helper,
    so the large-deck benchmark is reproducible everywhere with nothing committed.
  - `BenchmarkConvertRealWorld` — converts local `startup_x` if present, else
    `b.Skip`; reports real-world numbers when available.
- Run via the existing `make bench` (`go test -bench=. -benchmem`). Report
  ns/op, B/op, allocs/op, and derived throughput (slides/sec, MB/sec).

## Execution & evidence

1. **Functionality**: `make test` (runs unit tests + Tier 1 golden evals + Tier 2
   local evals with real decks present). Capture pass/fail output.
2. **Performance**: `make bench`. Capture the benchmark table.

Both outputs are reported back as the evidence deliverable.

## Files

| File | Purpose | Committed |
|------|---------|-----------|
| `testdata/fixtures/*.pptx` | synthetic + downloaded MIT decks | yes (small) |
| `testdata/fixtures/PROVENANCE.md` | source/license/sha256 record | yes |
| `testdata/golden/*.md` | golden expected Markdown | yes |
| `eval_test.go` | Tier 1 golden-file eval | yes |
| `eval_local_test.go` | Tier 2 structural-invariant eval | yes |
| `bench_test.go` | Tier 3 benchmarks | yes |
| `testdata/local/**` | real decks for local eval | no (gitignored) |

## Risks

- A downloaded fixture might exercise an OOXML construct the converter mishandles
  (e.g. grouped shapes, charts-as-images). If so, the eval surfaces it as a
  failing golden; we capture the actual output, decide whether it is acceptable
  (adjust golden) or a real bug (separate fix), and never weaken an invariant.
- `testdata/local/` must be gitignored *before* any real deck is copied in, to
  avoid accidentally staging sensitive content.

## Scope check

Single focused plan. No decomposition needed.
