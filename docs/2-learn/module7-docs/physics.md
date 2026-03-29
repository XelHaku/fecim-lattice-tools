# Module 7: Documentation - Physics

## Evidence Status (Demonstrated vs Modeled vs Aspirational)

- **Demonstrated:** Repository structure, navigation behavior, and code paths referenced in this page are implemented in this repo and verifiable from source/tests.
- **Modeled:** Equations, defaults, and performance/quality estimates are simulator or documentation models unless explicitly tied to cited measured data.
- **Aspirational:** Any production-scale, silicon-parity, or ecosystem-wide claims are roadmap intent and must not be reported as demonstrated results.

## Prerequisites

- Basic understanding of markdown
- Familiarity with hierarchical navigation

## Core Model

- Documents are indexed and ranked for search.
- The UI provides three parallel navigation paths: tree, breadcrumbs, and search.
- Root sidebar ordering is deterministic: module folders first, then `research-papers`, then `README.md`, then `MODULES.md`.
- Category detection is filename-driven for deterministic behavior.

## Key Equations (Simplified)

```
score(term, doc) = tf(term, doc) * idf(term)
reading_time_minutes = ceil(word_count / 200)
```

## Parameters And Units

| Symbol | Meaning | Units |
|---|---|---|
| tf | Term frequency | count |
| idf | Inverse document frequency | unitless |
| score | Relevance score | unitless |
| t_read | Reading time | minutes |
| word_count | Tokenized word count | words |
| 200 | Reading speed assumption | words/minute |

## Assumptions And Limits

- Search is scoped to the docs root.
- Ranking is heuristic, not a research-grade IR system.
- Category rules prioritize filename over path heuristics.
- Reading time uses a fixed 200 wpm heuristic (not user-specific).

## Where It Lives In Code

- `module7-docs/pkg/gui/embedded.go`
- `module7-docs/pkg/gui/search.go`
- `module7-docs/pkg/gui/navigation.go`

## Sources

- `docs/3-develop/gui/GUI.module7.md`
- `docs/3-develop/api-reference.md`
