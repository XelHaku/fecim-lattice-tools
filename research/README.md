# Research Ledger

This directory stores git-trackable research ingestion artifacts.

Canonical reviewed claims and paper records remain under `citations/`.
This directory stores the retrieval ledger: normalized source metadata,
parser outputs, chunks, manifests, reports, and rebuildable search cache
manifests.

Tracked:

- `papers/` stores source PDFs and downloaded paper files before parsing.
- `sources/`
- `parsed/`
- `chunks/`
- `extracted/`
- `graphs/`
- `manifests/`
- `reports/`
- `index/` stores rebuildable cache manifests and lightweight placeholders;
  bulky cache contents are ignored.

Ignored rebuildable caches:

- `index/pyserini/`
- `index/lancedb/`
- `index/models/`
- `.cache/`
