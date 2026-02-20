<!-- Category: ELI5 | Module: module7-docs | Reading time: ~3 min -->
# Module 7 ELI5: The Documentation Viewer

> Module 7 is the built-in library for the FeCIM curriculum. It
> organizes all the learning material and lets you navigate it
> without leaving the application.

---

## What It Does

The documentation viewer is a guided library built into the FeCIM
application. Instead of opening a web browser or a folder of markdown
files, you get a structured curriculum viewer with:

- A **sidebar tree** showing all modules and their documents
- **Breadcrumbs** at the top so you always know where you are
- A **search bar** to find topics quickly
- **Module shortcuts** to jump between ELI5, Physics, Features, and
  Tools for the same module

```
  +------------------+----------------------------------+
  |  Sidebar Tree    |  Content Pane                    |
  |                  |                                  |
  |  module1/        |  # Module 1 ELI5                 |
  |    ELI5.md    <--|                                  |
  |    PHYSICS.md    |  The ferroelectric effect is...   |
  |    FEATURES.md   |                                  |
  |  module2/        |  [diagram here]                  |
  |    ELI5.md       |                                  |
  |    ...           |  ## Next Steps                   |
  |                  |  - PHYSICS.md                    |
  |  [Search...]     |  - FEATURES.md                   |
  +------------------+----------------------------------+
  |  Breadcrumbs: docs > module1 > ELI5.md             |
  +----------------------------------------------------+
```

---

## Navigation: Three Ways to Get Around

### 1. The Tree

The sidebar shows the curriculum in a fixed order:

```
  Root
  +-- module1-hysteresis/
  |     ELI5 --> PHYSICS --> FEATURES --> OPENSOURCE-TOOLS
  +-- module2-crossbar/
  |     ELI5 --> PHYSICS --> FEATURES --> OPENSOURCE-TOOLS
  +-- module3-mnist/
  +-- module4-circuits/
  +-- module5-comparison/
  +-- module6-eda/
  +-- module7-docs/
  +-- research-papers/
  +-- README.md
  +-- MODULES.md
```

Click a folder to expand it. Click a file to load it in the
content pane.

### 2. Search

Type a term in the search bar and results are ranked by relevance.
The search index covers all documents in the curriculum.

The ranking formula is simple:

```
  score = term_frequency(word, document) * inverse_document_frequency(word)
```

Documents that contain your search term more often (and where the
term is relatively rare across all documents) rank higher.

### 3. Module Shortcuts

When viewing a document inside a module folder, shortcut buttons
appear:

```
  [ELI5] [PHYSICS] [FEATURES] [TOOLS]
```

Click any button to jump to that layer for the current module.
This makes it easy to go from the beginner explanation to the
equations to the feature list without scrolling through the tree.

---

## Reading Time

Each document shows an estimated reading time:

```
  reading_time = ceil(word_count / 200)   minutes
```

This assumes 200 words per minute, which is a rough average for
technical material.

---

## Responsive Layout

The viewer adapts to different window sizes:

| Window Width | Layout |
|-------------|--------|
| Under 600 px | Mobile -- sidebar collapses |
| 600-899 px | Tablet -- narrow sidebar |
| 900-1200 px | Desktop -- full sidebar + content |
| Over 1200 px | Wide -- extra margin for readability |

---

## What the Viewer Simplifies

- Markdown rendering is limited to what Fyne widgets support.
- Search is in-memory and scoped to the docs directory.
- No external URLs are fetched or embedded.
- Rendering focuses on clarity, not pixel-perfect typography.

---

## Next Steps

- [PHYSICS.md](PHYSICS.md) -- search ranking and layout architecture.
- [FEATURES.md](FEATURES.md) -- what the viewer implements.
- [OPENSOURCE-TOOLS.md](OPENSOURCE-TOOLS.md) -- documentation tools.

---
*FeCIM Lattice Tools -- Simulation baseline, not validated device data.*
