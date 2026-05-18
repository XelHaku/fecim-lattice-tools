# Research Ledger

This directory stores git-trackable research ingestion artifacts.

Canonical reviewed claims and paper records remain under `citations/`.
This directory stores the retrieval ledger: normalized source metadata,
parser outputs, chunks, manifests, reports, and rebuildable search cache
manifests.

Local-only input:

- `papers/` stores source PDFs before parsing. PDF files are ignored so the
  repository does not accidentally commit paper artifacts.
- `fecim-lattice-tools research rebuild` runs the local ingestion, index,
  claim-audit, and provenance-graph stages as the one-command corpus refresh.
  It writes `reports/rebuild-latest.json` so each refresh is reviewable in git.
- Rebuild also emits the cache status report. Cache warnings are recorded in
  `reports/rebuild-latest.json` without turning a file-only refresh into a
  failed rebuild.
- Use `fecim-lattice-tools research rebuild --skip-index` for file-only CI
  checks or repositories that do not currently have chunk files.
- `fecim-lattice-tools research cache` writes
  `reports/cache-latest.json`, a git-trackable status report for ignored
  rebuildable caches such as the Pyserini index.
- `fecim-lattice-tools research cache --clean` removes only cache directories
  covered by `research/.gitignore`, then rewrites the tracked cache report with
  the cleanup result.
- `fecim-lattice-tools research register-pdfs` writes a report-only inventory
  of local PDFs that still need canonical paper records.
- Add `--write-stubs` to create `needs-review` records under
  `citations/papers/` for unmatched, non-duplicate PDFs.

Paper acquisition:

- `fecim-lattice-tools research acquire` checks citation records that do not
  have a matching local PDF and records any legal OpenAlex open-access PDF
  location it finds.
- `fecim-lattice-tools research acquire --download` also downloads those
  open-access PDFs into `papers/`.
- `fecim-lattice-tools research acquire --doi DOI --download` acquires a new
  open-access paper by DOI using a deterministic provisional `doi_*` key.
- PDF files remain local and ignored. The OpenAlex response, acquisition YAML,
  SHA-256 digest, and latest acquisition report are written under tracked
  ledger paths so changes can be reviewed in git.
- Set `FECIM_OPENALEX_API_KEY` for current OpenAlex API access. Optional
  `FECIM_OPENALEX_MAILTO` is forwarded when present.

Tracked ledger outputs:

- `sources/`
- `parsed/`
- `chunks/`
- `extracted/`
- `graphs/`
- `manifests/`
- `reports/` includes latest acquisition, rebuild, ingestion, cache,
  claim-audit, claim-scan, cite, and graph JSON reports
- `index/` stores rebuildable cache manifests and lightweight placeholders;
  bulky cache contents are ignored.

Claim audit:

- Reviewed claim records live in `citations/claims/*.yaml`.
- Files listed in each record's `used_in` field must contain `[claim: id]`.
- `fecim-lattice-tools research cite CLAIM_ID` writes a deterministic citation
  packet to `reports/cite-latest.json`, resolving sources from
  `citations/papers` first and tracked OpenAlex ledgers second.
- Run `make research-audit` before promoting literature-backed statements into
  facts, config defaults, or trust documentation.
- `fecim-lattice-tools research claim-scan docs/ README.md` writes a
  report-first list of likely uncited scientific claims to
  `reports/claim-scan-latest.json`.
- Add `--fail-on-findings` only for narrow paths that have already been cleaned
  up; broad legacy scans are expected to produce review queues first.

Graph export:

- `fecim-lattice-tools research graph` writes the file-first provenance graph to
  `graphs/provenance-graph.json` and a small summary to
  `reports/graph-latest.json`.
- The graph is an export over committed claim/source files, not an authority.
  Rebuild it after changing `citations/papers`, `citations/claims`, or
  `[claim: id]` references.

Ignored rebuildable caches:

- `index/pyserini/`
- `index/lancedb/`
- `index/models/`
- `.cache/`

The cache report records whether required cache inputs still match their
tracked manifests. Cache directories remain ignored because they are derived
artifacts, but manifest and report changes stay visible in git.
