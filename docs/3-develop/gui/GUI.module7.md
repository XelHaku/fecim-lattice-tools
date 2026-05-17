---
Module: module7-docs
Name: FeCIM Documentation Viewer
Scope: Legacy Fyne adapter documentation
Default UI Path: `internal/gogpuapp` with `shared/viewmodel` snapshots
Legacy Build Tag: `legacy_fyne`
Entry: N/A (embedded only, accessed via toolbar icon)
Package: fecim-lattice-tools/module7-docs/pkg/gui
Last Updated: 2026-02-02
Description: |
  In-app documentation viewer with curriculum-first navigation, full-text search,
  responsive layout, breadcrumb navigation, table of contents, glossary term detection,
  and favorites persistence.
  Defaults to docs/documentation/ curriculum structure with module shortcuts panel
  (ELI5/Physics/Features/Tools) for guided learning paths.
  Includes quick links to curriculum overview, module index, and research index.
  Category detection provides visual indicators via icons and metadata badges.
  No physics simulation - utility module only.
---

These notes describe tagged legacy Fyne adapters; default UI work belongs in `internal/gogpuapp` and `shared/viewmodel`.

Conventions:
  - File paths are relative to module7-docs unless noted
  - Widget types refer to Fyne (`widget.*`, `container.*`, `canvas.*`) or shared widgets
  - Bindings list event handlers or UI update calls impacting the component

Bugs:
  (None currently tracked)

## Curriculum-First Navigation

Module 7 implements a **curriculum-first navigation model** that guides users through FeCIM concepts
in a structured, progressive learning path. The default view displays `docs/documentation/` which
contains all curriculum materials organized by module and topic.

**Key Principle:** Every module in the curriculum has a consistent structure with four core documents:

| Document | Purpose | Audience |
|----------|---------|----------|
| `ELI5.md` | Explain Like I'm 5 - intuitive overview without math | Beginners, non-technical stakeholders |
| `PHYSICS.md` | Rigorous physics explanations with equations | Researchers, device physicists |
| `FEATURES.md` | Implementation details and system capabilities | Engineers implementing the system |
| `OPENSOURCE-TOOLS.md` | Third-party tools and open-source alternatives | Integration engineers, tool users |

This structure allows learners to start simple (ELI5) and progressively deepen understanding (Physics),
while engineers jump directly to implementation details (Features, Tools).

**Quick Links:** The sidebar includes a compact set of buttons for curriculum overview, module index,
and research index to reduce hunting at the root level.

---

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
                      text: "Curriculum"
                      style: Bold, centered
                  - CurriculumLinks (VBox):
                      type: container.VBox
                      purpose: Quick links to README, MODULES, and research index
                      file: embedded.go (buildCurriculumLinks)
                  - ModuleShortcuts (ModuleShortcutsPanel):
                      type: ModuleShortcutsPanel
                      purpose: Quick links to ELI5/PHYSICS/FEATURES/OPENSOURCE-TOOLS for current module
                      file: navigation.go
                  - DocTree (Tree):
                      type: widget.Tree
                      purpose: Curriculum tree navigation with category-based icons
                      file: embedded.go:215-322
                      icons:
                        - eli5.md: theme.HelpIcon()
                        - physics.md: theme.ComputerIcon()
                        - features.md: theme.ListIcon()
                        - opensource-tools.md: theme.SettingsIcon()
                        - default: theme.DocumentIcon()
                      sorting:
                        - Root level: Module folders (module1, module2...) sorted by number
                        - Within modules: ELI5 → Physics → Features → Tools → Other (alphabetical)
                        - Directories always appear before files
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

---

## Module Shortcuts Panel

The Module Shortcuts Panel provides one-click access to the four core curriculum documents for
the currently selected module. It appears in the left sidebar once a module is selected.

**Panel Behavior:**

- **Inactive State:** Shows message "Select a module page to enable shortcuts."
- **Active State:** Displays four buttons (ELI5, Physics, Features, Tools)
- **Disabled Buttons:** Gray out if the corresponding file doesn't exist for that module
- **Current Module Detection:** Automatically detects module from selected document path

**Example:** When user clicks `docs/documentation/module1-hysteresis/PHYSICS.md`:
1. `moduleShortcuts.SetModulePath("docs/documentation/module1-hysteresis")`
2. Panel enables all four buttons (all files exist in module1-hysteresis)
3. Clicking "ELI5" navigates to `module1-hysteresis/ELI5.md`
4. Clicking "Tools" navigates to `module1-hysteresis/OPENSOURCE-TOOLS.md`

**Code Reference:** `navigation.go:276-332` (SetModulePath, refresh methods)

---

## Category Detection

Files are automatically categorized based on **filename priority**, then **path patterns**.
Categories are displayed as colored badges next to document names in the tree and search results.

**Category Detection Logic** (`search.go:205-235`):

| Priority | Rule | Category |
|----------|------|----------|
| 1 | Filename = `eli5.md` | **ELI5** (Green) |
| 2 | Filename = `physics.md` | **Physics** (Cyan) |
| 3 | Filename = `features.md` or `opensource-tools.md` | **Guide** (Blue) |
| 4 | Filename contains "research" | **Research** (Orange) |
| 5 | Filename contains "demo" | **Demo** (Purple) |
| 6 | Path contains `/research-papers/` | **Research** (Orange) |
| 7 | Path contains `/cim/` or `/crossbar/` | **Physics** (Cyan) |
| 8 | Default | **Guide** (Blue) |

**Category Colors** (Visual Indicators):

```
ELI5:      Green (#4CAF50)
Physics:   Cyan (#00BCD4)
Research:  Orange (#FF9800)
Demo:      Purple (#9C27B0)
Guide:     Blue (#2196F3) - default
```

**Category Metadata:** Stored in `SearchDocMetadata.Category` during index build.
Used in search results, breadcrumbs, and tree node display.

---

## Keyboard Shortcuts

**Global Shortcuts:**

| Shortcut | Action | File |
|----------|--------|------|
| `Cmd+K` (Mac) or `Ctrl+K` (Linux/Windows) | Open search dialog | search.go, embedded.go:95 |
| `Up/Down Arrow` (in search results) | Navigate results | search.go (SearchDialog) |
| `Enter` (in search results) | Select result | search.go (SearchDialog) |
| `Esc` (in search dialog) | Close search | search.go (SearchDialog) |

**Setup:** Keyboard shortcut initialized in `BuildContent()` -> `SetupSearchShortcut(window, searchDialog)`

---

## Learning Path Support

The curriculum structure enables **progressive learning** through four distinct paths:

### Path 1: Beginner (ELI5)

Start with intuitive explanations without mathematics. Ideal for:
- Non-technical stakeholders
- Product managers
- Briefings
- Students new to ferroelectronics

**Navigation:** Select any module, click "ELI5" button in Module Shortcuts Panel.

### Path 2: Physicist (Physics → Research Papers)

Deep-dive into physics with equations and reported in literature publications. Ideal for:
- Device physicists
- Materials scientists
- Researchers

**Navigation:**
1. Select module "Physics" via shortcuts
2. Use breadcrumbs to navigate to research-papers/ folder
3. Full-text search for specific topics (e.g., "hysteresis", "polarization")

### Path 3: Engineer (Features → Tools → Integration)

Hands-on implementation details with tool references. Ideal for:
- Hardware engineers
- Software architects
- Systems integrators

**Navigation:**
1. Start with "Features" to understand capabilities
2. Click "Tools" to see open-source alternatives
3. Use search to find specific APIs or integration points

### Path 4: Cross-Cutting (Search + Glossary)

Jump directly to specific topics across all modules. Ideal for:
- Quick lookups
- Troubleshooting
- Comparing concepts across modules

**Navigation:**
1. Press `Cmd/Ctrl+K` to open search
2. Type term (e.g., "preisach") - searches all categories
3. Click result to navigate
4. Glossary pills show related terms

**Example Flow:**

User wants to understand Preisach Model:
1. Press `Cmd/Ctrl+K` → Search "preisach"
2. See results from:
   - `module1-hysteresis/ELI5.md` (Category: ELI5)
   - `module1-hysteresis/PHYSICS.md` (Category: Physics)
   - `research-papers/preisach-paper.md` (Category: Research)
3. Click Physics result
4. See "Preisach Model" glossary pill → click to see definition
5. Breadcrumbs allow quick jump to ELI5 version of same module

---

DataFlow:
  - event: User clicks document in tree
    trigger: tree.OnSelected (embedded.go:300-304)
    flow:
      1. loadDocument(path) called
      2. Read markdown file from disk
      3. Update contentText via fyne.Do()
      4. Update breadcrumbs with path
      5. Detect module path and update shortcuts panel (SetModulePath called)
      6. Parse ToC headings from markdown
      7. Detect glossary terms via DetectGlossaryTerms()
      8. Get document metadata from SearchIndex (category, reading time, etc.)
      9. Add document to recent history
      10. Highlight glossary terms in content
    updates:
      - contentText.ParseMarkdown()
      - breadcrumbs.SetPath()
      - moduleShortcuts.SetModulePath()
      - toc.ParseMarkdown()
      - glossaryPills.SetTerms()
      - docMetadata.SetMetadata()

  - event: User selects module shortcut
    trigger: ModuleShortcutsPanel button.OnTapped
    flow:
      1. Get module path from button action
      2. Call loadDocument(modulePath/shortcutFile)
      3. Update all UI components (breadcrumbs, ToC, metadata, content)
      4. Module shortcuts stays with same modulePath (no change)
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
      2. TF-IDF scoring applied (per category)
      3. Boosts applied:
         - Title match: 3.0x
         - Heading match: 2.0x
         - Glossary term match: 1.5x
         - Exact match: 1.5x
      4. Return top 10 results with snippets and category badges
      5. Sort by relevance and display match type (title/heading/content/glossary)
    updates:
      - resultsList.Refresh() with SearchResult entries

  - event: User selects search result
    trigger: resultsList.OnSelected
    flow:
      1. Extract DocPath from SearchResult
      2. Call loadDocument(docPath)
      3. Update all UI as in "tree selection" flow
    updates:
      - All contentText, breadcrumbs, module shortcuts, ToC, metadata

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
      1. suppressSelect[uid] set to true
      2. history.ToggleFavorite(path)
      3. Persist to JSON
      4. Refresh tree to update star icon
    updates:
      - tree.Refresh()

SharedState:
  - SearchIndex:
      type: *SearchIndex
      purpose: Full-text search index with TF-IDF, category detection, and glossary awareness
      file: search.go:43-49
      fields:
        - index: map[string][]IndexEntry (inverted index: term -> documents with frequency, position, context)
        - docs: map[string]*SearchDocMetadata (cached document metadata including category)
        - docsPath: string
      thread_safety: sync.RWMutex
      initialization: Lazy build on first Query() call for fast startup
      category_field: SearchDocMetadata.Category (ELI5, Physics, Research, Demo, Guide)

  - DocsHistory:
      type: *DocsHistory
      purpose: Persist favorites and recent documents to JSON
      file: persistence.go
      fields:
        - recent: []string (LRU, last 10 viewed documents)
        - favorites: map[string]bool (in-memory set of starred documents)
        - favoritesList: []string (persisted to disk)
        - filePath: string (.omc/docs-history.json)
      thread_safety: sync.RWMutex
      persistence: Asynchronous write to .omc/docs-history.json

  - LayoutManager:
      type: *LayoutManager
      purpose: Responsive layout management with curriculum-first default
      file: layout.go
      breakpoints:
        - Mobile: < 600px (sidebar overlay, no ToC)
        - Tablet: 600-900px (sidebar sidebar, minimal ToC)
        - Desktop: 900-1200px (three-panel layout)
        - Wide: > 1200px (expansive three-panel layout)

Notes:
  - Thread safety: All UI updates from goroutines use fyne.Do()
  - Search index builds lazily on first query; call Build() for synchronous build
  - DocsHistory persists asynchronously to .omc/docs-history.json
  - Glossary terms detected via shared/widgets.TermsData (global glossary)
  - Responsive breakpoints: Mobile (<600), Tablet (600-900), Desktop (900-1200), Wide (>1200)
  - No physics simulation - documentation utility module only
  - No standalone entry point - embedded via toolbar icon in unified launcher
  - Keyboard shortcut: Cmd/Ctrl+K opens search dialog from anywhere in the app
  - Default navigation root: docs/documentation/ (curriculum-first structure)
  - Category detection: File-based (ELI5.md, PHYSICS.md, etc.) with path fallback
  - Progressive learning: ELI5 → Physics → Features/Tools → Research Papers
  - Curriculum structure: Every module has four core documents (ELI5/Physics/Features/Tools)
  - Module shortcuts: Auto-enable when module selected, auto-disable otherwise
  - Tree organization: Modules sorted by number, within modules by curriculum order
  - Category icons: ELI5 (Help), Physics (Computer), Features/Tools (List/Settings), Default (Document)
  - Search categories: Boost scoring per category (title/heading/glossary/exact matches)
  - Breadcrumb humanization: Underscores/hyphens → spaces, title case formatting

CustomWidgets:
  - BreadcrumbWidget:
      file: navigation.go:24-124
      purpose: Shows hierarchical path with clickable segments for navigation
      example: "Home > module1-hysteresis > PHYSICS"
      behavior:
        - Clickable segments navigate to folder/file
        - Auto-humanizes path (removes underscores, title case)
        - Always starts with "Home" link to docs root

  - TableOfContentsWidget:
      file: navigation.go:133-256
      purpose: Auto-generated from markdown headings (#, ##, ###) for on-page navigation
      features:
        - Clickable heading links (h1-h3)
        - Visual current section indicator (highlighted in primary color)
        - Indentation for h3+ (bullet points)
        - Only displays if document has 3+ headings
      behavior:
        - Parses markdown regex: `^(#{1,6})\s+(.+)$`
        - Generates anchors: lowercase, hyphens, alphanumeric only

  - GlossaryPillsWidget:
      file: glossary_integration.go:21-102
      purpose: Display clickable buttons for detected glossary terms with definitions
      features:
        - Detects terms from shared/widgets.TermsData
        - Whole-word matching (regex: `\b{term}\b`)
        - Click shows term definition popup
        - Case-insensitive matching with original casing preserved
        - Sorted alphabetically for consistency

  - DocumentMetadataWidget:
      file: glossary_integration.go:324-425
      purpose: Shows document category, reading time, and term count with visual badges
      displays:
        - Category badge (colored pill: ELI5/Physics/Research/Demo/Guide)
        - Reading time (estimated minutes, calculated as words / 200 wpm)
        - Glossary term count ("5 terms in this document")
      behavior:
        - Category badge click-through disabled (info-only)
        - Colors match CategoryColors map
        - Reading time >= 1 minute minimum
      category_colors:
        - ELI5: green (#4CAF50)
        - Physics: cyan (#00BCD4)
        - Research: orange (#FF9800)
        - Demo: purple (#9C27B0)
        - Guide: blue (#2196F3)

  - ModuleShortcutsPanel:
      file: navigation.go:258-332
      purpose: Quick-access buttons to current module's four core documents
      structure:
        - Title: "Current Module"
        - Buttons: ELI5, Physics, Features, Tools
        - Maps to: ELI5.md, PHYSICS.md, FEATURES.md, OPENSOURCE-TOOLS.md
      behavior:
        - Disabled until a module page/folder selected
        - Shows status: "Select a module page to enable shortcuts."
        - Disabled buttons if file doesn't exist
        - Enables on any document in a module folder
        - SetModulePath() called on every document load
      usage_example:
        - User clicks `module1-hysteresis/PHYSICS.md`
        - ModuleShortcutsPanel.SetModulePath(`module1-hysteresis`)
        - All four buttons enabled (module1-hysteresis/ELI5.md exists, etc.)
        - Click "Tools" → navigate to `module1-hysteresis/OPENSOURCE-TOOLS.md`

  - CategoryBadge:
      file: glossary_integration.go:265-298
      purpose: Visual indicator of document category with color coding
      displays:
        - Category name in colored pill
        - Color from CategoryColors map
      integration:
        - Used in DocumentMetadataWidget
        - Used in tree node rendering
        - Used in search result display

  - SearchDialog:
      file: search.go (SearchDialog type and methods)
      purpose: Modal search with fuzzy matching, category awareness, and match type display
      features:
        - Full-text search via inverted index
        - TF-IDF relevance scoring
        - Title/heading match boosting
        - Category-aware result grouping
        - Snippet preview with context (~100 chars)
        - Match type display ("title", "heading", "content", "glossary")
        - Keyboard navigation (up/down, enter, esc)
        - Category icon per result type
      behavior:
        - Fuzzy matching with Levenshtein distance
        - Top 10 results returned
        - Results sorted by relevance score
        - Snippet includes context around match term

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
