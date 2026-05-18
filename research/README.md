# Research Ledger

This directory stores git-trackable research ingestion artifacts.

Canonical reviewed claims and paper records remain under `citations/`.
This directory stores the retrieval ledger: normalized source metadata,
parser outputs, chunks, manifests, reports, and rebuildable search cache
manifests.

Local-only input:

- `papers/` stores source PDFs before parsing. PDF files are ignored so the
  repository does not accidentally commit paper artifacts.
- `fecim-lattice-tools research rebuild` runs local ingestion, missing-paper
  inventory, index, claim-audit, and provenance-graph stages as the
  one-command corpus refresh.
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

- `fecim-lattice-tools research missing` writes
  `reports/missing-papers-latest.json`, a local no-network inventory of
  citation records that still lack a matched PDF, including acquisition
  commands for records with DOI metadata.
- `fecim-lattice-tools research rebuild` refreshes this missing-paper report
  after ingestion so the paper acquisition queue stays reviewable in git.
- `fecim-lattice-tools research acquire` checks citation records that do not
  have a matching local PDF and records any legal OpenAlex open-access PDF
  location it finds.
- Acquisition also refreshes `reports/missing-papers-latest.json`, so planned,
  failed, and downloaded papers are reflected in the git-reviewable queue.
- `fecim-lattice-tools research acquire --download` also downloads those
  open-access PDFs into `papers/`.
- `fecim-lattice-tools research acquire --doi DOI --download` acquires a new
  open-access paper by DOI using a deterministic provisional `doi_*` key.
- New DOI acquisition writes a `needs-review` citation stub under
  `citations/papers/doi_*.md` and records the ignored local PDF path in the
  acquisition notes. The canonical `**PDF:**` field remains `not stored` until
  the paper is moved into a tracked PDF collection.
- PDF files remain local and ignored. The OpenAlex response, acquisition YAML,
  SHA-256 digest, and latest acquisition report are written under tracked
  ledger paths so changes can be reviewed in git.
- Set `FECIM_OPENALEX_API_KEY` for current OpenAlex API access. Optional
  `FECIM_OPENALEX_MAILTO` is forwarded when present.

Tracked ledger outputs:

- `sources/`
- `parsed/`
- `chunks/`
- `evidence/`
- `extracted/`
- `graphs/`
- `manifests/`
- `reports/` includes latest acquisition, rebuild, ingestion, cache,
  claim-audit, claim-scan, cite, graph, missing-paper, and search JSON reports
- `index/` stores rebuildable cache manifests and lightweight placeholders;
  bulky cache contents are ignored.

Claim audit:

- Reviewed claim records live in `citations/claims/*.yaml`.
- Files listed in each record's `used_in` field must contain `[claim: id]`.
- Citation paper records with stored `**PDF:**` paths must point at existing
  repo-relative files, and source ledgers in `sources/*.yaml` must keep their
  citation path, PDF path, and PDF SHA-256 digest in sync with the repository.
- `fecim-lattice-tools research cite CLAIM_ID` writes a deterministic citation
  packet to `reports/cite-latest.json`, resolving sources from
  `citations/papers` first and tracked OpenAlex ledgers second.
- `fecim-lattice-tools research evidence CLAIM_ID` saves the latest
  claim-linked search results to `evidence/CLAIM_ID.json` as candidate evidence
  that still needs review before promotion.
- `make research-audit` also checks `evidence/*.json`: the ledger filename must
  match a known claim id, candidate counts must match the listed candidates, and
  chunk candidates must still point at existing repo-relative JSONL chunk ids
  and digests.
- Run `make research-audit` before promoting literature-backed statements into
  facts, config defaults, or trust documentation.
- `make ci` runs `test-research` and `research-audit`, so PR CI verifies both
  the research command behavior and the tracked provenance ledger.
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

Search:

- `fecim-lattice-tools research search --local "HZO coercive field"` searches
  tracked JSONL chunks without Pyserini, Java, or rebuildable index caches.
- `fecim-lattice-tools research search --claim CLAIM_ID --local` uses the
  reviewed claim text as the query and records claim metadata in
  `reports/search-latest.json`.
- `fecim-lattice-tools research search "HZO coercive field"` uses the Pyserini
  BM25 cache built by `fecim-lattice-tools research index`.
- Search writes `reports/search-latest.json` so query output is reviewable in
  git when a search result becomes part of research work.

Ignored rebuildable caches:

- `index/pyserini/`
- `index/lancedb/`
- `index/models/`
- `.cache/`

The cache report records whether required cache inputs still match their
tracked manifests. Cache directories remain ignored because they are derived
artifacts, but manifest and report changes stay visible in git.
