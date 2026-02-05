# Module 7: Docs - Physics

## Prerequisites

- Basic understanding of markdown
- Familiarity with hierarchical navigation

## Core Model

- Documents are indexed and ranked for search.
- The UI provides three parallel navigation paths: tree, breadcrumbs, and search.
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

- `docs/development/GUI/GUI.module7.md`
- `docs/development/scriptReference.md#module-7-documentation-module7-docs`
