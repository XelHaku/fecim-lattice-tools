# CIM Architecture Source Ledgers

Normalized source ledgers for CIM architecture papers are grouped by research responsibility while preserving the shared source-ledger contract used by `tools/research/fecim_research/source_ledgers.py`.

## Topology

- `overviews-surveys/` — survey, landscape, and broad analog-AI overview papers.
- `ferroelectric-devices/` — FeFET, FeCap, FTJ, and multilevel ferroelectric CIM device papers.
- `crossbar-nonidealities/` — crossbar accuracy, sneak-path, and memory-technology non-ideality papers.
- `peripheral-precision/` — ADC precision, pruning/ADC tradeoff, and energy-efficiency papers.
- `system-architectures-applications/` — application-level CIM systems, annealers, CAM, and packing papers.

## Shared contract

Ledger discovery is intentionally path-recursive through `source_ledger_paths(root)`, so moving a `*.yaml` ledger into one of these subfolders does not change its `paper_key`, `citation_path`, or `pdf.path` contract. Keep filenames equal to `paper_key` and keep sidecar metadata next to the relevant ledger when sidecars are added.
