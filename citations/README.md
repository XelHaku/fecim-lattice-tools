# Citation System

This directory tracks project citations as plain text. Citations are treated like code: versioned, reviewable, searchable with standard command-line tools, and public by default.

The goal is not to collect papers for its own sake. The goal is to make every scientific or technical claim traceable to one of three statuses:

- **Cited:** supported by an external source recorded in `citations/papers/`.
- **Demonstrated:** supported by a reproducible project artifact in `validation/`.
- **Hypothesized:** explicitly labeled as a planned or educational assumption.

Claims that cannot fit one of those categories should not be published.

## Directory Layout

| Path | Purpose |
|---|---|
| `TEMPLATE.md` | Template for one paper record |
| `papers/` | One Markdown file per paper or source |
| `facts.md` | Cross-paper index of citable facts |
| `disputed.md` | Conflicting, weak, or contested claims |
| `refs.bib` | BibTeX bibliography compiled from paper records |
| `pdfs/` | Optional local PDFs, ignored by default |
| `queue/` | Lightweight paper intake queues |
| `reports/` | Citation audit and coverage reports |

## Paper Record Rule

Every paper in `citations/papers/` must have:

- a stable citation key
- title, authors, year, and venue
- DOI, arXiv ID, URL, or a clear note that no stable identifier was found
- reading status
- relevance to FeCIM Lattice Tools
- exact facts only after they have been read and verified
- a fenced `bibtex` block when ready for `refs.bib`

Do not add example numbers, inferred values, or remembered claims. If a fact has not been checked against the source, leave it out or mark it `[VERIFY]`.

## Workflow

1. Add candidate papers to `queue/to-fetch.md`.
2. Create a paper record from `TEMPLATE.md` in `papers/`.
3. Read or skim the source and update its status.
4. Extract only verified facts into the paper record.
5. Promote broadly useful facts to `facts.md`.
6. Add conflicts or weak evidence to `disputed.md`.
7. Run `scripts/citations/compile_bib.sh`.
8. Commit citation updates with a focused message.

## Agent Roles

The agent prompts in `prompts/citations/` define five bounded roles:

- **Fetcher:** find metadata and create paper records.
- **Summarizer:** summarize a paper's thesis and relevance.
- **Extractor:** extract quantitative facts with source location and conditions.
- **Auditor:** scan repository claims and flag missing or broken citations.
- **Writer:** insert the right citation when writing code, docs, or the paper.

The prompts are intentionally strict: agents must not invent metadata, fabricate DOIs, or treat weak sources as peer-reviewed evidence.

## Claim Standards

- Every number needs a citation or validation artifact.
- Code constants that encode physical parameters should include a nearby source comment.
- Conference or preprint evidence must be labeled as such.
- Secondary sources are acceptable for broad survey statements, but primary sources are preferred for numeric claims.
- Disagreement between sources belongs in `disputed.md`, not hidden in prose.

## Scripts

Scripts live in `scripts/citations/`.

```bash
bash scripts/citations/search.sh "hysteresis"
bash scripts/citations/compile_bib.sh
```

Agent-backed commands such as `add.sh` and `extract.sh` require `CITATION_AGENT_CMD` to be configured. Until that is set, they fail closed instead of producing unreviewed records.
