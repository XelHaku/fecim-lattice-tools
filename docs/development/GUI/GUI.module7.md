---
Module: module7-docs
Name: FeCIM Documentation Viewer
Entry: N/A (embedded only, accessed via toolbar icon)
Package: fecim-lattice-tools/module7-docs/pkg/gui
Last Updated: 2026-02-02
Description: |
  In-app documentation viewer with full-text search, responsive layout,
  breadcrumb navigation, table of contents, glossary term detection,
  and favorites persistence.
  Scans the docs/ directory for markdown files and renders them with
  navigation features. No physics simulation - utility module only.
---

Conventions:
  - File paths are relative to module7-docs unless noted
  - Widget types refer to Fyne (`widget.*`, `container.*`, `canvas.*`) or shared widgets
  - Bindings list event handlers or UI update calls impacting the component

Bugs:
  (None currently tracked)

Screens:
  - name: EmbeddedDocsApp
    file: embedded.go:17-44
    description: Main documentation viewer interface
    layout:
      - TopBar (Border):
          file: embedded.go:196-211
          components:
            - Title (Label):
                type: widget.Label
                text: "FeCIM Documentation"
                file: embedded.go:207
            - TocToggleButton (Button):
                type: widget.Button
                icon: theme.ListIcon
                purpose: Toggle ToC sidebar visibility
                file: embedded.go:201-203
            - SidebarToggleButton (Button):
                type: widget.Button
                icon: theme.MenuIcon
                purpose: Toggle document tree visibility (mobile overlay / desktop sidebar)
                file: embedded.go
            - SearchButton (Button):
                type: widget.Button
                icon: theme.SearchIcon
                purpose: Open search dialog
                file: embedded.go:197-199
      - LayoutManager (Container):
          file: embedded.go:74-82
          type: LayoutManager
          purpose: Manages responsive 3-panel layout
          components:
            - Sidebar (Border):
                file: embedded.go:130-145
                components:
                  - SidebarTitle (Label):
                      type: widget.Label
                      text: "Documentation"
                      style: Bold, centered
                  - DocTree (Tree):
                      type: widget.Tree
                      purpose: File tree navigation
                      file: embedded.go:200-290
            - MainContent (Border):
                file: embedded.go:167-184
                components:
                  - TopSection (VBox):
                      file: embedded.go:170-173
                      components:
                        - Breadcrumbs (BreadcrumbWidget):
                            type: BreadcrumbWidget
                            purpose: Hierarchical path navigation
                            file: navigation.go
                        - DocumentMetadata (DocumentMetadataWidget):
                            type: DocumentMetadataWidget
                            purpose: Category, reading time, terms
                            file: glossary_integration.go
                  - PillsSection (HBox):
                      file: embedded.go:176
                      components:
                        - GlossaryPills (GlossaryPillsWidget):
                            type: GlossaryPillsWidget
                            purpose: Clickable glossary term buttons
                            file: glossary_integration.go
                  - ContentScroll (VScroll):
                      type: container.VScroll
                      purpose: Scrollable markdown content
                      file: embedded.go:95
                      components:
                        - ContentText (RichText):
                            type: widget.RichText
                            purpose: Rendered markdown
                            file: embedded.go:93
            - TocSidebar (Border):
                file: embedded.go:186-193
                components:
                  - TocTitle (Label):
                      type: widget.Label
                      text: "On This Page"
                      style: Bold, centered
                      file: embedded.go:189
                  - TableOfContents (TableOfContentsWidget):
                      type: TableOfContentsWidget
                      purpose: Auto-generated heading navigation
                      file: navigation.go

  - name: SearchDialog
    file: search.go
    description: Modal full-text search dialog with fuzzy matching
    layout:
      - Entry (Entry):
          type: widget.Entry
          placeholder: "Search documentation..."
          purpose: Query input
      - ResultsList (List):
          type: widget.List
          purpose: Display search results with snippets
          bindings: OnSelected navigates to document

DataFlow:
  - event: User clicks document in tree
    trigger: tree.OnSelected (embedded.go:300-304)
    flow:
      1. loadDocument(path) called
      2. Read markdown file from disk
      3. Update contentText via fyne.Do()
      4. Update breadcrumbs with path
      5. Parse ToC headings from markdown
      6. Detect glossary terms
      7. Get document metadata from SearchIndex
      8. Add document to recent history
    updates:
      - contentText.ParseMarkdown()
      - breadcrumbs.SetPath()
      - toc.ParseMarkdown()
      - glossaryPills.SetTerms()
      - docMetadata.SetMetadata()

  - event: User types in search dialog
    trigger: SearchDialog entry.OnChanged
    flow:
      1. SearchIndex.Query() with fuzzy matching
      2. TF-IDF scoring applied
      3. Boost for title/heading matches
      4. Return top 10 results with snippets
    updates:
      - resultsList.Refresh()

  - event: Window resize
    trigger: LayoutManager.OnResize()
    flow:
      1. Determine breakpoint (Mobile/Tablet/Desktop/Wide)
      2. Rebuild layout based on mode
      3. Show/hide ToC based on width
    updates:
      - container.Objects reassigned
      - container.Refresh()

  - event: Toggle favorite
    trigger: starBtn.OnTapped in tree
    flow:
      1. history.ToggleFavorite(path)
      2. Persist to JSON
      3. Refresh tree to update star icon
    updates:
      - tree.Refresh()

SharedState:
  - SearchIndex:
      type: *SearchIndex
      purpose: Full-text search index with TF-IDF
      file: search.go:45-50
      fields:
        - index: map[string][]IndexEntry (inverted index)
        - docs: map[string]*SearchDocMetadata
        - docsPath: string
      thread_safety: sync.RWMutex

  - DocsHistory:
      type: *DocsHistory
      purpose: Persist favorites to JSON
      file: persistence.go
      fields:
        - recent: []string (LRU, last 10)
        - favorites: map[string]bool (in-memory)
        - favoritesList: []string (persisted)
        - filePath: string (.omc/docs-history.json)
      thread_safety: sync.RWMutex

  - LayoutManager:
      type: *LayoutManager
      purpose: Responsive layout management
      file: layout.go
      breakpoints:
        - Mobile: < 600px
        - Tablet: 600-900px
        - Desktop: 900-1200px
        - Wide: > 1200px

Notes:
  - Thread safety: All UI updates from goroutines use fyne.Do()
  - Search index builds lazily on first query; call Build() for a synchronous build
  - DocsHistory persists to .omc/docs-history.json asynchronously
  - Glossary terms detected via shared/widgets.TermsData
  - Responsive breakpoints: Mobile (<600), Tablet (600-900), Desktop (900-1200), Wide (>1200)
  - No physics simulation - documentation utility module only
  - No standalone entry point - embedded via toolbar icon in unified launcher
  - Keyboard shortcut: Cmd/Ctrl+K opens search dialog

CustomWidgets:
  - BreadcrumbWidget:
      file: navigation.go
      purpose: Shows hierarchical path with clickable segments
      example: "docs > development > GUI > GUI.module7.md"

  - TableOfContentsWidget:
      file: navigation.go
      purpose: Auto-generated from markdown headings (#, ##, ###)
      features:
        - Clickable heading links
        - Visual current section indicator
        - Supports up to 3 heading levels

  - GlossaryPillsWidget:
      file: glossary_integration.go
      purpose: Display clickable buttons for detected glossary terms
      features:
        - Detects terms from shared/widgets.TermsData
        - Click shows term definition popup

  - DocumentMetadataWidget:
      file: glossary_integration.go
      purpose: Shows document category, reading time, term count
      displays:
        - Category (ELI5, Physics, Research, Demo, Guide)
        - Reading time (minutes, based on word count / 200)
        - Glossary term count

  - SearchDialog:
      file: search.go
      purpose: Modal search with fuzzy matching
      features:
        - Full-text search via inverted index
        - TF-IDF relevance scoring
        - Title/heading match boosting
        - Snippet preview with context
        - Keyboard navigation (up/down, enter)

SearchAlgorithm:
  - index_type: Inverted index (term -> documents)
  - scoring: TF-IDF with boosts
  - boosts:
      - title_match: 3.0x
      - heading_match: 2.0x
      - glossary_match: 1.5x
      - exact_match: 1.5x
  - fuzzy_matching: Levenshtein-based for typo tolerance
  - snippet_length: ~100 characters with match context

LifecycleNotes:
  - Start(): No-op (no background processes)
  - Stop(): No-op (no background processes)
  - BuildContent(): Called once at app startup
    - Initializes SearchIndex (lazy build; Build() optional for sync)
    - Loads DocsHistory from disk
    - Creates all UI components
    - Sets up keyboard shortcut (Cmd/Ctrl+K)

Performance:
  - Search index build: ~100-500ms for typical docs/ size
  - Search query: < 50ms (in-memory index lookup)
  - Document load: < 100ms (file read + markdown parse)
  - Memory: ~5-10MB for index (depends on docs size)
