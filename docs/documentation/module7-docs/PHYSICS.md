<!-- Category: Physics | Module: module7-docs | Reading time: ~3 min -->
# Module 7 Physics: Search and Layout Architecture

> Module 7 is a documentation viewer, not a physics simulator. The
> "physics" here describes the information retrieval and layout models.

---

## Search Ranking

Documents are indexed and ranked using a term-frequency /
inverse-document-frequency (TF-IDF) model:

```
score(term, doc) = tf(term, doc) * idf(term)
```

| Symbol | Meaning | Units |
|--------|---------|-------|
| tf | Term frequency -- how often the word appears in this document | count |
| idf | Inverse document frequency -- rarity across all documents | unitless |
| score | Relevance score for ranking | unitless |

Documents with higher scores for a given search term appear first
in the results list.

---

## Reading Time Estimation

```
reading_time_minutes = ceil(word_count / 200)
```

| Symbol | Meaning | Units |
|--------|---------|-------|
| word_count | Tokenized word count in the document | words |
| 200 | Reading speed assumption | words/minute |
| reading_time | Estimated reading time | minutes |

The 200 wpm figure is a fixed heuristic, not personalized to the
reader.

---

## Category Detection

Document categories (ELI5, Physics, Features, Tools) are detected
from the filename, not the file path or content:

```
  eli5.md          --> ELI5
  PHYSICS.md       --> Physics
  features.md      --> Features
  tools.md         --> Open-Source Tools
  OPENSOURCE-TOOLS  --> Open-Source Tools
```

Filename-driven detection ensures deterministic behavior regardless
of where the file sits in the directory tree.

---

## Navigation Model

The viewer provides three parallel navigation paths:

1. **Tree navigation** -- hierarchical folder/file sidebar
2. **Breadcrumb navigation** -- location indicator with clickable
   path segments
3. **Search navigation** -- keyword-based document lookup

Root sidebar ordering is deterministic:
1. Module folders (sorted by module number)
2. research-papers/
3. README.md
4. MODULES.md

Within each module folder, documents are ordered:
1. ELI5
2. PHYSICS
3. FEATURES
4. OPENSOURCE-TOOLS

---

## Responsive Layout Model

Layout breakpoints adapt the UI to window size:

| Breakpoint | Width Range | Behavior |
|------------|------------|----------|
| Mobile | Under 600 px | Sidebar collapses, content fills width |
| Tablet | 600-899 px | Narrow sidebar alongside content |
| Desktop | 900-1200 px | Full sidebar + content pane |
| Wide | Over 1200 px | Extra margins for readability |

---

## Assumptions and Limits

- Search is scoped to the docs root directory.
- Ranking is heuristic, not a research-grade IR system.
- Category rules prioritize filename over path heuristics.
- Reading time uses a fixed 200 wpm assumption.
- Markdown rendering is limited to Fyne widget capabilities.

---

## Where It Lives in Code

| Path | Purpose |
|------|---------|
| `module7-docs/pkg/gui/embedded.go` | Document embedding and rendering |
| `module7-docs/pkg/gui/search.go` | Search indexing and ranking |
| `module7-docs/pkg/gui/navigation.go` | Tree and breadcrumb navigation |
| `module7-docs/pkg/gui/glossary_integration.go` | Glossary pill overlays |
| `module7-docs/pkg/gui/layout.go` | Responsive breakpoints |

---
*FeCIM Lattice Tools -- Simulation baseline, not validated device data.*
