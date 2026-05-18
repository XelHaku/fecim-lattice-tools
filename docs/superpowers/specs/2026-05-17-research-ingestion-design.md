# Research Ingestion And Claim Ledger Design

Date: 2026-05-17
Status: Approved for implementation planning

## Context

FeCIM Lattice Tools is a simulation and education workspace with an explicit accuracy boundary. Scientific and technical claims must remain traceable to cited literature, validation artifacts, or clearly labeled assumptions.

The repository already has a human-reviewed citation system:

- `citations/papers/*.md` for source records
- `citations/facts.md` for citable facts
- `citations/disputed.md` for weak, conflicting, or contested claims
- `citations/pdfs/` as an optional local PDF drop zone
- `docs/4-research/papers/` as an existing paper library
- validation tests and experimental data tied to literature records

The new research ingestion work must extend this trust model. It must not create a second canonical source of truth.

## Goals

1. Make the local paper corpus searchable and reproducible.
2. Preserve a file-first, git-trackable knowledge ledger for academic review.
3. Keep `citations/` canonical for reviewed sources, facts, disputed claims, and claim records.
4. Treat search indexes and databases as rebuildable caches over committed artifacts.
5. Support legal open-access acquisition for missing papers without bypassing paywalls or access controls.
6. Add Go CLI commands that provide a stable user surface while orchestrating Python research tooling.
7. Enable later CI checks that prevent unsupported or disputed scientific claims from being promoted as facts.

## Non-Goals

- No full research agents in the MVP.
- No generated chat answers as the primary search UX.
- No paywall bypassing, publisher access scraping, private session scraping, or disabled TLS verification.
- No remote AI embedding or summarization APIs by default.
- No database-first source of truth.
- No migration of the existing `citations/` trust model into `research/`.

## Core Architecture

The system is file-first.

`citations/` remains the reviewed trust layer. `research/` becomes the committed research ledger plus local retrieval workspace. Pyserini, LanceDB, and any future database are accelerators over committed files, not authorities.

The Go CLI owns the stable command surface:

```bash
fecim research ingest
fecim research index
fecim research search "HZO coercive field Preisach"
fecim research cite CHUNK_OR_PAPER_ID
fecim research claim-scan docs/ README.md
fecim research audit
fecim research graph
```

The implementation should start with the retrieval MVP:

```bash
fecim research ingest
fecim research index
fecim research search "HZO coercive field Preisach"
```

The Go commands call repo-local Python tooling under `tools/research/`. Python owns GROBID/Marker interaction, metadata lookup, chunking, Pyserini indexing, optional LanceDB indexing, and graph export.

## Directory Layout

Committed knowledge ledger:

```text
research/
  papers/             # New user drop zone for PDFs
  sources/            # Normalized metadata, one YAML file per paper key
  parsed/             # GROBID TEI XML and Marker Markdown
  chunks/             # JSONL chunks, one file per paper key
  extracted/          # Machine-extracted claim candidates, not reviewed facts
  graphs/             # JSONL/GraphML graph exports
  manifests/          # Tool versions, input hashes, output hashes, run manifests
  reports/            # Audit, unmatched, duplicate, and acquisition reports
  index/              # Rebuildable search caches; ignored except manifests

citations/
  claims/             # Reviewed claim YAML records
  papers/             # Existing reviewed source records
  facts.md            # Human-readable fact index
  disputed.md         # Disputed or weak claims
```

Ignored rebuildable caches:

```text
research/index/pyserini/
research/index/lancedb/
research/index/models/
research/.cache/
```

The exact ignore rules should preserve `research/index/manifest.json` or `research/manifests/index-*.json` while excluding bulky binary index data.

## Paper Identity

The canonical paper identity is `paper_key`.

Rules:

1. If a PDF maps to an existing `citations/papers/{key}.md`, use `{key}`.
2. If DOI or arXiv metadata exists but no citation record exists, write `research/sources/{provisional_key}.yaml` and queue creation of `citations/papers/{key}.md`.
3. DOI, arXiv ID, OpenAlex ID, Semantic Scholar ID, local PDF path, and PDF SHA256 are metadata. They do not replace `paper_key`.
4. Chunk IDs use stable, readable IDs:

```text
{paper_key}::sec-{section_number}::chunk-{chunk_number}
```

Example:

```text
park2015_advmat_hzo::sec-03::chunk-002
```

## PDF Discovery

The MVP discovers PDFs from all existing and new local paper locations:

- `docs/4-research/papers/**/*.pdf`
- `research/papers/**/*.pdf`
- `citations/pdfs/**/*.pdf`

Discovery writes a manifest with:

- path
- SHA256
- size
- modification time
- candidate title
- candidate DOI or arXiv ID
- matched `paper_key`
- match confidence
- match method
- quarantine status

Duplicate PDFs are detected by SHA256. Duplicate paths are reported but not treated as fatal unless they map to conflicting paper keys.

## Acquisition Policy

Automatic acquisition is aggressive but legal.

Allowed:

- OpenAlex metadata lookup
- Crossref metadata lookup
- Semantic Scholar metadata lookup
- arXiv PDF downloads
- Unpaywall open-access PDF discovery
- PubMed Central or institutional open-access PDF links
- publisher PDFs that are publicly open
- user-provided PDFs

Forbidden:

- bypassing paywalls
- scraping private sessions
- disabling TLS verification
- ignoring robots or rate limits
- inventing metadata
- treating a failed download as a successful source record

When a legal PDF cannot be found, the tool writes queue/report entries instead of failing silently:

- `research/reports/missing-pdfs.json`
- `citations/queue/to-fetch.md`

## Quarantine Policy

Ambiguous PDFs are quarantined by default.

A PDF is quarantined when it cannot be confidently matched by at least one trusted method:

- existing citation record metadata
- DOI
- arXiv ID
- OpenAlex ID
- Semantic Scholar ID
- exact title match
- manual override file
- known PDF hash

Quarantined outputs go to:

```text
research/sources/_unmatched/*.yaml
research/reports/unmatched-pdfs.json
```

Default search excludes quarantined chunks. A future `--include-unmatched` flag may include them explicitly.

## Parsing

GROBID is the canonical structured parser for scientific PDFs.

Marker is the Markdown fallback and secondary parser.

Committed parser outputs:

```text
research/parsed/{paper_key}/grobid.tei.xml
research/parsed/{paper_key}/marker.md
research/parsed/{paper_key}/manifest.json
```

Ignored parser byproducts:

- rendered page images
- extracted figure crops
- OCR image caches
- service caches
- model caches

If GROBID is unavailable, ingestion may continue with Marker only, but the parser manifest must record the degraded parse status. If both parsers are unavailable, ingestion records the failure and does not produce trusted chunks for that paper.

## Chunking

Chunks are committed as JSONL:

```text
research/chunks/{paper_key}.jsonl
```

Each JSONL record includes at least:

```json
{
  "id": "park2015_advmat_hzo::sec-03::chunk-002",
  "paper_key": "park2015_advmat_hzo",
  "contents": "Chunk text for Pyserini indexing.",
  "section": "Results",
  "section_number": 3,
  "chunk_number": 2,
  "source_parser": "grobid",
  "source_path": "research/parsed/park2015_advmat_hzo/grobid.tei.xml",
  "page_start": 4,
  "page_end": 5,
  "char_start": 1204,
  "char_end": 2879,
  "sha256": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
}
```

`id` and `contents` are required because Pyserini's JSONL collection path expects those fields for Lucene indexing.

Chunking should be deterministic for the same parser output and chunker version. The chunk manifest records chunker version, input file hash, output hash, chunk count, and warnings.

## Metadata

Normalized source metadata is committed as YAML:

```text
research/sources/{paper_key}.yaml
```

Minimum fields:

```yaml
paper_key: park2015_advmat_hzo
title: Ferroelectricity and antiferroelectricity of doped thin HfO2-based films
year: 2015
authors:
  - Park
doi: 10.1002/adma.201404531
arxiv_id: null
openalex_id: null
semantic_scholar_id: null
venue: Advanced Materials
pdf:
  path: research/papers/park2015_advmat_hzo.pdf
  sha256: 0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef
  acquisition: local
  license_status: user-provided
citation_record: citations/papers/park2015_advmat_hzo.md
status: matched
```

Metadata imported from external services must record the service name and retrieval date in the manifest.

## Indexing

`fecim research index` builds rebuildable caches from committed chunks.

Default behavior:

```bash
fecim research index
```

Builds BM25 only.

Semantic search is opt-in and local-only:

```bash
fecim research index --semantic
fecim research index --semantic --embedding-model sentence-transformers/all-MiniLM-L6-v2
```

No remote embedding APIs are allowed by default.

Pyserini cache:

```text
research/index/pyserini/
```

LanceDB cache:

```text
research/index/lancedb/
```

Tracked index manifest:

```text
research/manifests/index-latest.json
```

The manifest records:

- chunks input paths and hashes
- Pyserini version
- Java version when relevant
- BM25 settings
- semantic enabled or disabled
- embedding model name
- embedding model revision or local hash when semantic indexing is enabled
- vector dimension
- cache output paths
- cache creation time

## Search

`fecim research search` returns evidence chunks, not generated answers.

Default output includes:

- rank
- score
- `paper_key`
- title/year when known
- section and page when available
- chunk ID
- short snippet
- source chunk file path
- parser provenance
- linked reviewed claim IDs when present

Example:

```text
1. park2015_advmat_hzo  score=12.84  sec=Results  page=4
   chunk: park2015_advmat_hzo::sec-03::chunk-002
   claims: hzo-remanent-polarization-range
   Snippet: HZO remanent polarization appears in the supporting text.
```

Machine-readable output is available with:

```bash
fecim research search --json "HZO coercive field Preisach"
```

Generated summaries or RAG answers are deferred. If added later, they must be local-only and must cite returned chunks.

## Claim Registry

Reviewed claims live in:

```text
citations/claims/*.yaml
```

Machine-extracted candidates live in:

```text
research/extracted/{paper_key}.claims.jsonl
```

Generated candidates are not facts. A claim becomes citable only after a reviewed YAML record exists under `citations/claims/`.

Reviewed claim schema:

```yaml
id: hzo-remanent-polarization-range
claim: HZO devices commonly report remanent polarization in a specific range.
status: literature-backed
confidence: medium
sources:
  - paper_key: park2015_advmat_hzo
    chunks:
      - park2015_advmat_hzo::sec-03::chunk-002
  - paper_key: materlik2015_jap_hfo2_origin
used_in:
  - config/materials.yaml
  - docs/TRUST.md
review:
  reviewer: human
  date: 2026-05-17
```

Allowed claim statuses:

- `literature-backed`
- `validated`
- `assumption`
- `disputed`
- `deprecated`
- `needs-review`

Only `literature-backed` and `validated` claims may be promoted into public facts without explicit caveats.

## Audit And CI

`make research-audit` should eventually run:

```bash
fecim research audit
fecim research claim-scan docs/ README.md config/ module*/ shared/
```

Required checks:

1. Every reviewed claim has a stable claim ID.
2. Every reviewed claim has at least one source or validation artifact.
3. Every cited source exists in `citations/papers/` or `research/sources/`.
4. Every cited chunk exists in `research/chunks/`.
5. Every cited PDF path or metadata record exists.
6. Docs and config do not contain uncited scientific numeric claims above the configured strictness threshold.
7. Disputed claims are not promoted into `citations/facts.md` as facts.
8. Quarantined papers are excluded from trusted search and claim promotion.
9. Parser/index manifests are internally consistent.

CI should not require large model downloads. BM25 and schema validation are mandatory. Semantic index validation is opt-in.

## Error Handling

The pipeline fails closed for trust-sensitive operations:

- ambiguous PDF match creates quarantine output
- missing legal PDF creates missing-PDF report and queue entry
- parser failure records a parse report and skips trusted chunks for that paper
- metadata disagreement records a conflict report
- disputed claims cannot be promoted as facts
- semantic index requests fail if the local embedding model is unavailable

Errors should be actionable and point to the relevant report path.

## Implementation Sequence

1. Add `research/` layout, ignore rules for caches, and skeleton README files.
2. Add claim YAML directory and schema examples under `citations/claims/`.
3. Add Python package under `tools/research/` for discovery, metadata normalization, parsing, chunking, indexing, and search.
4. Add Go `research` subcommand dispatcher that shells out to the Python tool.
5. Implement `ingest` for local PDF discovery, hashing, citation-key matching, source YAML, parser manifests, and chunk JSONL.
6. Implement `index` for BM25 from `research/chunks/*.jsonl`.
7. Implement `search` over BM25 with evidence-only output and `--json`.
8. Add legal OA acquisition for queued or missing papers.
9. Add semantic indexing as local-only opt-in.
10. Add claim registry validation and `make research-audit`.

## MVP Acceptance Criteria

The first milestone is complete when:

1. A user can drop PDFs into `research/papers/`.
2. Existing PDFs under `docs/4-research/papers/` are discovered.
3. `fecim research ingest` creates git-trackable metadata, parse outputs, chunks, manifests, and reports.
4. Ambiguous PDFs are quarantined and excluded from default search.
5. `fecim research index` builds a BM25 cache from committed chunks.
6. `fecim research search "HZO coercive field Preisach"` returns evidence chunks with paper keys and provenance.
7. Search results can link to reviewed claim YAML records when present.
8. Generated caches can be deleted and rebuilt from committed files.

## Deferred Work

These are intentionally outside the retrieval MVP:

- full local research agents
- Neo4j-backed graph serving
- local LLM answer generation
- figure/table image extraction as committed evidence
- semantic indexing in mandatory CI
- automatic promotion of extracted claim candidates into reviewed claims
