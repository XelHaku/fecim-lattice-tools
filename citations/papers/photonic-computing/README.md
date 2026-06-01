# Photonic Computing Citation Records

This folder groups photonic-computing paper records. The records currently under
`needs-review/` are intake stubs whose bibliographic metadata has not been
verified.

## Needs-Review Stub Contract

Every `needs-review/` record keeps the citation-key metadata, PDF pointer,
SHA256, and file size from ingestion. Before using one of these papers as
evidence, complete this shared checklist:

- [ ] Confirm bibliographic metadata.
- [ ] Add DOI or arXiv ID when available.
- [ ] Add source notes only after reading the paper.
- [ ] Move evidence-bearing claims into `citations/claims/*.yaml`.

Repository tooling should resolve records by citation key through the recursive
paper-record locator rather than assuming a fixed one-level path.
