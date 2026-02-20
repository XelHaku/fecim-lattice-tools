<!-- Category: Open-Source Tools | Module: module7-docs | Reading time: ~3 min -->
# Module 7 Open-Source Tools: Documentation

> Tools for building and viewing technical documentation.

---

## Tools Used in This Module

| Tool | Role |
|------|------|
| Go toolchain | Document indexer, search engine, and renderer |
| Fyne | GUI framework for the viewer interface |

The documentation viewer is implemented entirely in Go using the Fyne
GUI toolkit. It reads markdown files from the docs directory and
renders them in a structured curriculum layout.

---

## External Documentation Tools

These open-source tools are relevant for creating, processing, or
hosting technical documentation.

### Markdown Processing

| Tool | Description |
|------|-------------|
| **Pandoc** | Universal document converter. Converts markdown to PDF, HTML, DOCX, LaTeX, and many other formats. |
| **mdBook** | Rust-based book builder from markdown. Used by the Rust project. Creates HTML books from a directory of .md files. |
| **MkDocs** | Python-based static site generator for project documentation. Supports themes, search, and navigation. |
| **Hugo** | Fast static site generator written in Go. Supports markdown content with flexible theming. |

### Diagram Tools

| Tool | Description |
|------|-------------|
| **Mermaid** | Text-based diagramming. Write flowcharts, sequence diagrams, and state diagrams in markdown-compatible syntax. |
| **PlantUML** | Text-to-UML diagram generator. Supports class, sequence, activity, and component diagrams. |
| **draw.io / diagrams.net** | Browser-based diagram editor. Exports SVG, PNG, PDF. |
| **Graphviz** | Graph visualization from DOT language descriptions. Good for dependency and architecture diagrams. |

### Search and Indexing

| Tool | Description |
|------|-------------|
| **Lunr.js** | Client-side full-text search for static sites. Similar TF-IDF approach to Module 7's built-in search. |
| **Typesense** | Open-source search engine. Typo-tolerant, fast. Good for documentation search at larger scale. |
| **Algolia DocSearch** | Free documentation search for open-source projects (hosted service, not self-hosted). |

### Hosting

| Tool | Description |
|------|-------------|
| **GitHub Pages** | Free static hosting from a Git repository. Works with MkDocs, Hugo, mdBook. |
| **Netlify** | Static site hosting with CI/CD. Automatic builds from Git pushes. |
| **Read the Docs** | Documentation hosting with versioning. Commonly used for Python projects. |

---

## Integration Notes

- Module 7 renders markdown locally using Fyne widgets. It does not
  use any of the external tools listed above.
- The curriculum structure is defined in `docs/documentation/`.
- Search category rules live in `module7-docs/pkg/gui/search.go`.
- Viewer behavior documentation: `docs/development/GUI/GUI.module7.md`.

---

## Code Locations

| Path | Purpose |
|------|---------|
| `module7-docs/pkg/gui/embedded.go` | Document embedding and rendering |
| `module7-docs/pkg/gui/search.go` | Search indexing and ranking |
| `module7-docs/pkg/gui/navigation.go` | Tree and breadcrumb navigation |
| `module7-docs/pkg/gui/glossary_integration.go` | Glossary integration |

---
*FeCIM Lattice Tools -- Simulation baseline, not validated device data.*
