<!-- Category: Features | Module: module7-docs | Reading time: ~3 min -->
# Module 7 Features: Documentation Viewer

> Feature inventory for the curriculum documentation viewer.

---

## Navigation

| Feature | Status |
|---------|--------|
| Hierarchical sidebar tree | IMPLEMENTED |
| Deterministic root ordering (modules, research, README, MODULES) | IMPLEMENTED |
| Deterministic module ordering (ELI5, Physics, Features, Tools) | IMPLEMENTED |
| Breadcrumb location indicator | IMPLEMENTED |
| Folder expand/collapse on click | IMPLEMENTED |
| File load on click | IMPLEMENTED |

---

## Search

| Feature | Status |
|---------|--------|
| In-memory full-text search | IMPLEMENTED |
| TF-IDF relevance ranking | IMPLEMENTED |
| Scoped to docs directory | IMPLEMENTED |
| Real-time results as you type | IMPLEMENTED |

---

## Module Shortcuts

| Feature | Status |
|---------|--------|
| ELI5 / Physics / Features / Tools buttons | IMPLEMENTED |
| Context-aware (shows buttons for current module) | IMPLEMENTED |
| Jump between learning layers without tree navigation | IMPLEMENTED |

---

## Content Display

| Feature | Status |
|---------|--------|
| Markdown rendering in content pane | IMPLEMENTED |
| Reading time estimate (ceil(words / 200)) | IMPLEMENTED |
| Category detection from filename | IMPLEMENTED |
| Glossary term highlighting (pill overlays) | IMPLEMENTED |
| Table of contents generation | IMPLEMENTED |

---

## Responsive Layout

| Feature | Status |
|---------|--------|
| Mobile layout (under 600 px) | IMPLEMENTED |
| Tablet layout (600-899 px) | IMPLEMENTED |
| Desktop layout (900-1200 px) | IMPLEMENTED |
| Wide layout (over 1200 px) | IMPLEMENTED |
| Sidebar collapse on narrow windows | IMPLEMENTED |

---

## Interaction Details

| Feature | Status |
|---------|--------|
| Star/favorite button per document | IMPLEMENTED |
| Star click does not trigger document load (suppressSelect) | IMPLEMENTED |
| Folder click toggles branch, file click loads document | IMPLEMENTED |

---

## Known Limitations

- Markdown rendering limited to Fyne widget capabilities (no full
  CommonMark or GFM support).
- Search is in-memory, optimized for repository scale (hundreds of
  documents, not millions).
- No external URLs are fetched or embedded.
- No image rendering in markdown content.

---

## Extension Points

- Add new category rules in `search.go`.
- Extend module shortcuts for new curriculum layers.
- Customize layout breakpoints in `layout.go`.
- Add new document formats beyond markdown.

---

## Where It Lives in Code

| Path | Purpose |
|------|---------|
| `module7-docs/pkg/gui/embedded.go` | Document embedding |
| `module7-docs/pkg/gui/search.go` | Search and categories |
| `module7-docs/pkg/gui/navigation.go` | Tree and breadcrumbs |
| `module7-docs/pkg/gui/glossary_integration.go` | Glossary integration |
| `module7-docs/pkg/gui/layout.go` | Responsive layout |

---

## Status Legend

- **IMPLEMENTED**: Code exists, tests pass, feature is functional

---
*FeCIM Lattice Tools -- Simulation baseline, not validated device data.*
