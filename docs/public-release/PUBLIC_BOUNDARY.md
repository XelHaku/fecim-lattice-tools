# Public Boundary

## Allowed Material
- Source code, tests, validation harnesses, and safe example data or fixtures.
- Build, run, contributor, and maintenance documentation needed to build or review the source tree.
- Research summaries, methodology notes, and educational writeups written in project language.
- DOI lists, arXiv links, metadata records, and paper download scripts.
- Release-policy documents, audit records, and enforcement scripts used to keep the public tree inside the approved boundary.
- Third-party PDFs only when `docs/public-release/THIRD_PARTY_PDF_AUDIT.csv` marks the tracked file as `keep` or `keep-with-conditions`.

## Source-Only Rule
The public repository is source-only.

Hand-authored code, documentation, scripts, audit records, and safe fixture data may be tracked when they are intentionally part of the source release.

Generated outputs may not be tracked. Generated outputs include:
- compiled binaries and packaged builds
- logs, including nested module GUI log directories, screenshots, recordings, and other run artifacts
- validation output, benchmark output, and exported reports
- generated EDA and export files unless they are intentionally curated sample inputs or fixtures

Deterministic validation fixtures that support tests may be tracked only under `testdata/` paths. They must not remain in runtime output directories such as `output/`, `validation/output/`, or module-local `output/` trees.

## Disallowed Material
- `docs/archive/**`
- `docs/4-research/internal-analysis/**`
- `docs/4-research/transcripts/COSM_2025_AI_Hardware_Breakthrough/**`
- `docs/4-research/transcripts/ironlattice-youtube-script.md`
- `docs/4-research/tour-group-ironlattice-research.md`
- `docs/4-research/superlattice-material-analysis.md`
- `docs/superpowers/**`
- `docs/research-papers/tourlab/**`
- `docs/4-research/papers/tourlab/**`
- `demo2-crossbar/pkg/_layers_experimental/kvcache_fjh.go`
- `demo2-crossbar/pkg/layers/kvcache_fjh.go`
- `module2-crossbar/pkg/_layers_experimental/kvcache_fjh.go`
- `docs/4-research/papers/DOWNLOAD_PLAN.md`
- `docs/4-research/opensource-tools/research_notes_final.md`
- `validation/literature/_incoming/**`
- Repo-local test logs such as `module1-hysteresis/test_output*.txt`
- Generated build or run artifacts such as binaries, logs, screenshots, recordings, and export output.
- Runtime validation output directories outside `testdata/`, including root `output/`, `validation/output/`, and module-local `output/` trees.
- Personal roleplay material, outreach planning, or NDA/confidential material.
- Third-party PDFs without explicit redistribution evidence recorded in the audit.
