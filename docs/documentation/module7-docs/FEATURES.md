# Module 7: Docs - Features

## What This Module Does

- Provides a curriculum-first documentation viewer.
- Offers search, breadcrumbs, ToC, and glossary integration.
- Adds module shortcuts for ELI5, PHYSICS, FEATURES, and TOOLS.

## Primary Components

- `module7-docs/pkg/gui/embedded.go`
- `module7-docs/pkg/gui/search.go`
- `module7-docs/pkg/gui/navigation.go`
- `module7-docs/pkg/gui/glossary_integration.go`

## Key Workflows

- Select a document from the curriculum tree (module-number sort at root, curriculum-file order inside modules).
- Use module shortcuts to jump between learning layers (ELI5 → PHYSICS → FEATURES → OPENSOURCE-TOOLS).
- Use search and glossary pills for cross-topic navigation.

## Extension Points

- Add new category rules in `search.go`.
- Extend the module shortcuts panel for new curriculum layers.
- Customize layout breakpoints in `layout.go`.

## Layout + Interaction Notes (Validated)

- Breakpoints: Mobile `<600`, Tablet `600-899`, Desktop `900-1200`, Wide `>1200` (see `layout.go`).
- Tree row click target behavior:
  - Clicking a folder row toggles branch open/close.
  - Clicking a file row loads markdown into the content pane.
- Star/favorite button consumes the row-selection event once (`suppressSelect`) to prevent accidental document loads.

## Known Limitations

- Markdown rendering is limited to supported Fyne widgets.
- Search is in-memory and optimized for repo scale.
- No external URLs are fetched or embedded.
