# Photonic Computing Needs-Review Records

This folder holds photonic-computing intake stubs whose bibliographic metadata has
not been verified. Subfolders group records by the first review responsibility so
reviewers can triage similar papers together without changing citation keys.

## Shared Stub Contract

Every record in this tree keeps the citation-key metadata, PDF pointer, SHA256,
and file size from ingestion. Before using one of these papers as evidence:

- [ ] Confirm bibliographic metadata.
- [ ] Add DOI or arXiv ID when available.
- [ ] Add source notes only after reading the paper.
- [ ] Move evidence-bearing claims into `citations/claims/*.yaml`.

Repository tooling should resolve records by citation key through the recursive
paper-record locator rather than assuming a fixed one-level path.
