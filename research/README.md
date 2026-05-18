# Research Ledger

This directory stores git-trackable research ingestion artifacts.

Canonical reviewed claims and paper records remain under `citations/`.
This directory stores the retrieval ledger: normalized source metadata,
parser outputs, chunks, manifests, reports, and rebuildable search cache
manifests.

Local-only input:

- `papers/` stores source PDFs before parsing. PDF files are ignored so the
  repository does not accidentally commit paper artifacts.
- Citation records must not use `research/papers/...` in their canonical
  `**PDF:**` field. Use `not stored` for inbox PDFs, and promote only
  licensing-reviewed PDFs into tracked paper paths.
- `fecim-lattice-tools research rebuild` runs local ingestion, missing-paper
  inventory, index, claim-audit, and provenance-graph stages as the
  one-command corpus refresh.
  It writes `reports/rebuild-latest.json` so each refresh is reviewable in git.
- When a citation record has a stored canonical `**PDF:**` path, ingestion
  prefers that tracked PDF over duplicate local inbox copies.
- Matched PDFs that exist only in ignored `papers/` are reported in
  `reports/local-inbox-pdfs.json` with `needs_promotion`; ingestion does not
  write canonical source, parse, or chunk ledgers for them until promotion.
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
- Stubs generated from ignored inbox PDFs keep `**PDF:**` as `not stored` and
  record the local inbox path separately.
- `fecim-lattice-tools research promote-pdf KEY --to docs/4-research/papers/.../KEY.pdf --license LICENSE --license-url URL --review-note NOTE`
  copies a reviewed inbox PDF into a tracked canonical PDF path, updates the
  citation record's canonical `**PDF:**`, `**SHA256:**`, and `**Size:**`
  fields, and writes `reports/pdf-promotion-latest.json` plus
  `sources/KEY.promotion.yaml` as git-trackable review evidence.
- Registration refreshes `reports/missing-papers-latest.json` so new stubs and
  matched local PDFs are reflected in the acquisition queue immediately.

Paper acquisition:

- `fecim-lattice-tools research missing` writes
  `reports/missing-papers-latest.json`, a local no-network inventory of
  citation records that still lack a matched PDF, including acquisition
  commands for records with DOI metadata.
- Missing-paper reports also write a content-addressed history copy under
  `reports/missing-papers/RUN_ID.json`, so changes to the acquisition backlog
  can be reviewed over time in git.
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
- Promotion is explicit: downloads remain in ignored `papers/` until review,
  then `promote-pdf` copies them into `docs/4-research/papers/` or
  `citations/pdfs/` and refreshes the missing-paper report. Promotion requires
  license metadata and a human review note before any PDF is copied into a
  tracked path.
- PDF files remain local and ignored. The OpenAlex response, acquisition YAML,
  SHA-256 digest, and latest acquisition report are written under tracked
  ledger paths so changes can be reviewed in git.
- Acquisition writes `reports/acquisition-latest.json` plus a content-addressed
  copy under `reports/acquisitions/RUN_ID.json`. The run ID is derived from the
  report payload, so identical acquisition outcomes do not create timestamp
  churn.
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
  plus content-addressed acquisition, missing-paper, search, claim-audit, and
  cite history reports
- `index/` stores rebuildable cache manifests and lightweight placeholders;
  bulky cache contents are ignored.

Claim audit:

- Reviewed claim records live in `citations/claims/*.yaml`.
- Files listed in each record's `used_in` field must contain `[claim: id]`.
- Citation paper records with stored `**PDF:**` paths must point at existing
  repo-relative files, and source ledgers in `sources/*.yaml` must keep their
  citation path, PDF path, and PDF SHA-256 digest in sync with the repository.
- Source ledgers must not point at ignored `research/papers/...` inbox files;
  promote reviewed PDFs before treating them as canonical evidence inputs.
- PDF promotion ledgers in `sources/*.promotion.yaml` are review evidence:
  audit checks their license metadata, citation path, destination PDF path, and
  SHA-256 digest.
- Canonical citation `**PDF:**` paths must have either a promotion ledger or a
  tracked `manifests/pdf-review-backlog.json` entry. The backlog exists only
  for legacy tracked PDFs that still need explicit license review; new tracked
  PDFs should use `promote-pdf` instead of expanding the backlog.
- `fecim-lattice-tools research cite CLAIM_ID` writes a deterministic citation
  packet to `reports/cite-latest.json`, resolving sources from
  `citations/papers` first and tracked OpenAlex ledgers second. It also writes
  a content-addressed copy under `reports/cites/RUN_ID.json` so citation
  packets can be reviewed over time.
- `fecim-lattice-tools research evidence CLAIM_ID` saves the latest
  claim-linked search results to `evidence/CLAIM_ID.json` as candidate evidence
  that still needs review before promotion.
- `make research-audit` also checks `evidence/*.json`: the ledger filename must
  match a known claim id, candidate counts must match the listed candidates, and
  chunk candidates must still point at existing repo-relative JSONL chunk ids
  and digests.
- Run `make research-audit` before promoting literature-backed statements into
  facts, config defaults, or trust documentation.
- Audit writes `reports/claim-audit-latest.json` plus a content-addressed
  history copy under `reports/claim-audits/RUN_ID.json`, so CI provenance
  changes can be reviewed over time.
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
- `fecim-lattice-tools research index --semantic` builds a local vector cache
  under ignored `research/index/lancedb/`. When LanceDB is installed, the cache
  includes a LanceDB table; otherwise it falls back to deterministic local
  JSONL vectors using the `fecim-hashing-bow-v1` embedding model. No remote
  embedding APIs are used.
- `fecim-lattice-tools research search --semantic "HZO coercive field"`
  searches the rebuildable vector cache and writes a reviewable search report.
- `fecim-lattice-tools research search --claim CLAIM_ID --local` uses the
  reviewed claim text as the query and records claim metadata in
  `reports/search-latest.json`.
- Search reports also write content-addressed history copies under
  `reports/searches/RUN_ID.json`, so retrieval changes can be reviewed before
  they become candidate evidence.
- `fecim-lattice-tools research search --inbox "HZO coercive field"` searches
  only unreviewed local inbox sidecar Markdown next to ignored
  `research/papers/*.pdf` files and marks every result `UNREVIEWED`.
- Inbox search is for triage only: it does not create canonical chunks or
  indexes, cannot be combined with `--claim`, and cannot feed
  `fecim-lattice-tools research evidence`.
- `fecim-lattice-tools research search "HZO coercive field"` uses the Pyserini
  BM25 cache built by `fecim-lattice-tools research index`.
- Index commands write backend-specific tracked manifests:
  `research/manifests/index-pyserini.json` validates the BM25 cache, and
  `research/manifests/index-lancedb.json` validates the semantic/vector cache.
  `research/manifests/index-latest.json` is only the most recent index run.
- Search writes `reports/search-latest.json` so query output is reviewable in
  git when a search result becomes part of research work. The latest file is a
  compatibility pointer for `fecim-lattice-tools research evidence`; the
  content-addressed report is the durable review artifact.

Ignored rebuildable caches:

- `index/pyserini/`
- `index/lancedb/`
- `index/models/`
- `.cache/`

The cache report records whether required cache inputs still match their
tracked manifests. Cache directories remain ignored because they are derived
artifacts, but manifest and report changes stay visible in git.
