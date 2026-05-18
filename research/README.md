# Research Ledger

This directory stores git-trackable research ingestion artifacts.

Canonical reviewed claims and paper records remain under `citations/`.
This directory stores the retrieval ledger: normalized source metadata,
parser outputs, chunks, manifests, reports, and rebuildable search cache
manifests.

Local-only input:

- `papers/` stores source PDFs before parsing. PDF files are ignored so the
  repository does not accidentally commit paper artifacts.

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
- `reports/` includes latest acquisition and claim-audit JSON reports
- `index/` stores rebuildable cache manifests and lightweight placeholders;
  bulky cache contents are ignored.

Claim audit:

- Reviewed claim records live in `citations/claims/*.yaml`.
- Files listed in each record's `used_in` field must contain `[claim: id]`.
- Run `make research-audit` before promoting literature-backed statements into
  facts, config defaults, or trust documentation.
- `fecim-lattice-tools research claim-scan docs/ README.md` writes a
  report-first list of likely uncited scientific claims to
  `reports/claim-scan-latest.json`.
- Add `--fail-on-findings` only for narrow paths that have already been cleaned
  up; broad legacy scans are expected to produce review queues first.

Ignored rebuildable caches:

- `index/pyserini/`
- `index/lancedb/`
- `index/models/`
- `.cache/`
