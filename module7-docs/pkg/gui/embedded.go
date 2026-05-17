//go:build legacy_fyne

// pkg/gui/embedded.go
// Embeddable documentation viewer for the unified visualizer
package gui

import (
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/shared/canvas"
	sharedWidgets "fecim-lattice-tools/shared/widgets"
)

// EmbeddedDocsApp is the embeddable documentation viewer
type EmbeddedDocsApp struct {
	sharedWidgets.EmbeddedAppBase
	currentFile string
	docsPath    string

	// New components
	window        fyne.Window
	searchIndex   *SearchIndex
	history       *DocsHistory
	layoutManager *LayoutManager

	// UI components
	tree            *widget.Tree
	contentText     *widget.RichText
	contentScroll   *container.Scroll
	breadcrumbs     *BreadcrumbWidget
	toc             *TableOfContentsWidget
	docMetadata     *DocumentMetadataWidget
	glossaryPills   *GlossaryPillsWidget
	searchDialog    *SearchDialog
	moduleShortcuts *ModuleShortcutsPanel

	// State
	pathMap        map[string]*docEntry
	docs           []*docEntry
	suppressSelect map[string]bool
}

func (app *EmbeddedDocsApp) bindHost(_ fyne.App, window fyne.Window) {
	app.window = window
}

// NewEmbeddedDocsApp creates a new embedded docs app instance
func NewEmbeddedDocsApp() *EmbeddedDocsApp {
	return &EmbeddedDocsApp{}
}

// OpenAboutScience navigates the docs viewer to the unified "About the Science" entry point.
// Safe to call after BuildContent; if docsPath is not initialized yet it will no-op.
func (app *EmbeddedDocsApp) OpenAboutScience() {
	if app == nil || app.docsPath == "" {
		return
	}
	path := filepath.Join(app.docsPath, "about", "About.Science.md")
	app.loadDocument(path)
}

// docEntry represents a documentation file or folder
type docEntry struct {
	name     string
	path     string
	isDir    bool
	children []*docEntry
}

// RegisterKeyboard re-registers the docs module's keyboard handler on the
// shared canvas. Called by the unified app when this tab becomes active.
func (app *EmbeddedDocsApp) RegisterKeyboard() {
	if app.window != nil && app.searchDialog != nil {
		SetupSearchShortcut(app.window, app.searchDialog)
	}
}

// BuildContent creates the UI content for embedding in the main app
func (app *EmbeddedDocsApp) BuildContent(fyneApp fyne.App, window fyne.Window) fyne.CanvasObject {
	return app.EmbeddedAppBase.BuildOrReuseContentWithHostSync(fyneApp, window, app.bindHost, func() fyne.CanvasObject {
		app.docsPath = utils.FindDirectory(filepath.Join("docs", "documentation"))
		if app.docsPath == "" {
			app.docsPath = utils.FindDirectory("docs")
		}

		// Initialize search index
		app.searchIndex = NewSearchIndex(app.docsPath)

		// Initialize history persistence
		app.history = NewDocsHistory()

		// Create all UI components
		app.createUIComponents()

		// Setup responsive layout
		app.layoutManager = NewLayoutManager()
		app.layoutManager.SetComponents(
			app.buildSidebar(),     // tree + quick access
			app.buildMainContent(), // breadcrumbs + metadata + content
			app.buildTocSidebar(),  // table of contents
			app.buildTopBar(),      // search button, title
		)

		docsContent := app.layoutManager.BuildLayout()

		// Setup keyboard shortcut for search
		SetupSearchShortcut(window, app.searchDialog)

		// Wrap docs + validation dashboard in tabs
		validationDash := sharedWidgets.NewValidationDashboard()
		tabs := container.NewAppTabs(
			container.NewTabItem("Documentation", docsContent),
			container.NewTabItem("Validation", validationDash.Container),
		)
		return tabs
	})
}

// createUIComponents initializes all UI widgets
func (app *EmbeddedDocsApp) createUIComponents() {
	// Content viewer
	app.contentText = widget.NewRichTextFromMarkdown("# FeCIM Curriculum\n\nSelect a page from **Curriculum Links** or the navigation tree on the left.\n\nQuick start:\n- Overview\n- Module Index\n- Research Index\n\nKeyboard: use **Ctrl+F** to search documentation.")
	app.contentText.Wrapping = fyne.TextWrapWord
	app.contentScroll = container.NewVScroll(app.contentText)

	// Breadcrumbs
	app.breadcrumbs = NewBreadcrumbWidget(func(path string) {
		app.navigateToFolder(path)
	})

	// ToC
	app.toc = NewTableOfContentsWidget(func(anchor string) {
		app.scrollToSection(anchor)
	})

	// Document metadata
	app.docMetadata = NewDocumentMetadataWidget(app.window)
	app.glossaryPills = NewGlossaryPillsWidget(app.window)

	// Search dialog
	app.searchDialog = NewSearchDialog(app.searchIndex, app.window, func(path string) {
		app.loadDocument(path)
	})

	// Module quick access shortcuts
	app.moduleShortcuts = NewModuleShortcutsPanel(func(path string) {
		app.loadDocument(path)
	})

	if app.suppressSelect == nil {
		app.suppressSelect = make(map[string]bool)
	}
}

// buildSidebar creates the left sidebar with tree
func (app *EmbeddedDocsApp) buildSidebar() fyne.CanvasObject {
	// Scan docs and build tree
	app.docs = app.scanDocsDirectory()
	app.pathMap = make(map[string]*docEntry)
	app.buildPathMap(app.docs)
	app.tree = app.createDocTree()
	sharedWidgets.SetAccessibleLabel(app.tree, "Document navigation tree")

	sidebarTop := container.NewVBox(
		widget.NewLabelWithStyle("Curriculum", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		app.buildCurriculumLinks(),
		app.moduleShortcuts,
		widget.NewSeparator(),
	)

	return container.NewBorder(
		sidebarTop,
		nil, nil, nil,
		app.tree,
	)
}

// buildCurriculumLinks creates quick links to top-level curriculum pages.
func (app *EmbeddedDocsApp) buildCurriculumLinks() fyne.CanvasObject {
	if app.docsPath == "" {
		return widget.NewLabel("")
	}

	links := []struct {
		label string
		path  string
	}{
		{label: "About the Science", path: filepath.Join(app.docsPath, "about", "About.Science.md")},
		{label: "Overview", path: filepath.Join(app.docsPath, "README.md")},
		{label: "Module Index", path: filepath.Join(app.docsPath, "MODULES.md")},
		{label: "Research Index", path: filepath.Join(app.docsPath, "research-papers", "README.md")},
	}

	items := []fyne.CanvasObject{
		widget.NewLabelWithStyle("Curriculum Links", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	}

	for _, link := range links {
		path := link.path
		btn := widget.NewButtonWithIcon(link.label, theme.DocumentIcon(), func() {
			app.loadDocument(path)
		})
		// Keep quick links clearly readable in dark themes.
		btn.Importance = widget.MediumImportance
		if _, err := os.Stat(path); err != nil {
			btn.Disable()
		}
		items = append(items, btn)
	}

	return container.NewVBox(items...)
}

// buildPathMap recursively builds the path lookup map
func (app *EmbeddedDocsApp) buildPathMap(entries []*docEntry) {
	for _, e := range entries {
		app.pathMap[e.path] = e
		if e.isDir && len(e.children) > 0 {
			app.buildPathMap(e.children)
		}
	}
}

// buildMainContent creates the central content area
func (app *EmbeddedDocsApp) buildMainContent() fyne.CanvasObject {
	// Top metadata section (includes category, reading time, and key terms)
	topSection := container.NewVBox(
		app.breadcrumbs,
		app.docMetadata,
		app.glossaryPills,
	)

	return container.NewBorder(
		topSection,
		nil,
		nil, nil,
		app.contentScroll,
	)
}

// buildTocSidebar creates the right ToC sidebar
func (app *EmbeddedDocsApp) buildTocSidebar() fyne.CanvasObject {
	helper := widget.NewLabel("Select a document to view section headings.")
	helper.Wrapping = fyne.TextWrapWord

	return container.NewBorder(
		widget.NewLabelWithStyle("On This Page", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		helper, nil, nil,
		app.toc,
	)
}

// buildTopBar creates the top action bar with search
func (app *EmbeddedDocsApp) buildTopBar() fyne.CanvasObject {
	searchBtn := widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
		app.showSearch()
	})
	sharedWidgets.SetAccessibleLabel(searchBtn, "Search documentation")

	tocToggleBtn := widget.NewButtonWithIcon("", theme.ListIcon(), func() {
		app.layoutManager.ToggleToc()
	})
	sharedWidgets.SetAccessibleLabel(tocToggleBtn, "Toggle table of contents")
	sidebarToggleBtn := widget.NewButtonWithIcon("", theme.MenuIcon(), func() {
		app.layoutManager.ToggleSidebar()
	})
	sharedWidgets.SetAccessibleLabel(sidebarToggleBtn, "Toggle document sidebar")

	return container.NewBorder(
		nil, nil,
		widget.NewLabel("FeCIM Documentation"),
		container.NewHBox(sidebarToggleBtn, tocToggleBtn, searchBtn),
		nil,
	)
}

// getCategoryIcon returns an appropriate icon for a file based on its name.
func getCategoryIcon(filename string) fyne.Resource {
	filenameLower := strings.ToLower(filename)
	switch filenameLower {
	case "eli5.md":
		return theme.HelpIcon()
	case "physics.md":
		return theme.ComputerIcon()
	case "features.md":
		return theme.ListIcon()
	case "opensource-tools.md":
		return theme.SettingsIcon()
	default:
		return theme.DocumentIcon()
	}
}

// treeCategory returns the badge category for a tree row.
func (app *EmbeddedDocsApp) treeCategory(entry *docEntry) string {
	if entry == nil {
		return ""
	}

	if entry.isDir {
		base := strings.ToLower(filepath.Base(entry.path))
		switch {
		case app.isModuleDir(entry.path):
			return "Module"
		case base == "research-papers":
			return "Research"
		default:
			return ""
		}
	}

	nameLower := strings.ToLower(entry.name)
	switch nameLower {
	case "eli5.md":
		return "ELI5"
	case "physics.md":
		return "Physics"
	case "features.md", "opensource-tools.md", "readme.md", "modules.md":
		return "Guide"
	default:
		if strings.Contains(filepath.ToSlash(entry.path), "/research-papers/") {
			return "Research"
		}
		return ""
	}
}

// createDocTree builds the documentation file tree widget
func (app *EmbeddedDocsApp) createDocTree() *widget.Tree {
	tree := widget.NewTree(
		// ChildUIDs - return child IDs for a node
		func(uid widget.TreeNodeID) []widget.TreeNodeID {
			if uid == "" {
				// Root level
				var ids []widget.TreeNodeID
				for _, d := range app.docs {
					ids = append(ids, d.path)
				}
				return ids
			}
			if entry, ok := app.pathMap[uid]; ok && entry.isDir {
				var ids []widget.TreeNodeID
				for _, c := range entry.children {
					ids = append(ids, c.path)
				}
				return ids
			}
			return nil
		},
		// IsBranch - returns true if node has children
		func(uid widget.TreeNodeID) bool {
			if uid == "" {
				return true
			}
			if entry, ok := app.pathMap[uid]; ok {
				return entry.isDir && len(entry.children) > 0
			}
			return false
		},
		// Create - create a new tree node widget
		func(branch bool) fyne.CanvasObject {
			icon := widget.NewIcon(theme.DocumentIcon())
			badge := NewCategoryBadge("")
			badge.Hide()
			label := widget.NewLabel("Document")
			label.Truncation = fyne.TextTruncateEllipsis
			starBtn := widget.NewButtonWithIcon("", theme.ContentAddIcon(), nil)
			starBtn.Importance = widget.LowImportance
			starBtn.Hidden = true
			sharedWidgets.SetAccessibleLabel(starBtn, "Toggle favorite")

			center := container.NewHBox(badge, label)
			return container.NewBorder(nil, nil, icon, starBtn, center)
		},
		// Update - update tree node with data
		func(uid widget.TreeNodeID, branch bool, node fyne.CanvasObject) {
			box := node.(*fyne.Container)
			// Find widgets by type - don't rely on container object ordering.
			var icon *widget.Icon
			var starBtn *widget.Button
			var label *widget.Label
			var badge *CategoryBadge

			for _, obj := range box.Objects {
				switch v := obj.(type) {
				case *widget.Icon:
					icon = v
				case *widget.Button:
					starBtn = v
				case *fyne.Container:
					// Center container: [badge][label]
					for _, child := range v.Objects {
						switch c := child.(type) {
						case *widget.Label:
							label = c
						case *CategoryBadge:
							badge = c
						}
					}
				}
			}

			if entry, ok := app.pathMap[uid]; ok {
				label.SetText(entry.name)

				category := app.treeCategory(entry)
				if entry.isDir {
					icon.SetResource(theme.FolderIcon())
					starBtn.Hidden = true
				} else {
					icon.SetResource(getCategoryIcon(entry.name))
					starBtn.Hidden = false

					// Update star icon based on favorite status.
					if app.history.IsFavorite(entry.path) {
						starBtn.SetIcon(theme.ContentRemoveIcon())
					} else {
						starBtn.SetIcon(theme.ContentAddIcon())
					}

					// Capture path for closure.
					path := entry.path
					starBtn.OnTapped = func() {
						app.suppressSelect[path] = true
						app.toggleFavorite(path)
						app.tree.Refresh()
					}
				}

				if badge != nil {
					if category == "" {
						badge.Hide()
					} else {
						badge.SetCategory(category)
						badge.Show()
					}
				}
			}
		},
	)

	// Handle selection - load markdown file or toggle folder
	tree.OnSelected = func(uid widget.TreeNodeID) {
		if app.consumeSuppressSelect(uid) {
			return
		}
		if entry, ok := app.pathMap[uid]; ok {
			if entry.isDir {
				app.updateModuleShortcuts(entry.path)
				// Toggle folder open/close when clicking anywhere on the row
				if tree.IsBranchOpen(uid) {
					tree.CloseBranch(uid)
				} else {
					tree.OpenBranch(uid)
				}
			} else {
				app.loadDocument(entry.path)
			}
		}
	}

	return tree
}

// loadDocument loads and displays a markdown document
func (app *EmbeddedDocsApp) loadDocument(path string) {
	content, err := os.ReadFile(path)
	if err != nil {
		fyne.Do(func() {
			app.contentText.ParseMarkdown("# Error\n\nCould not read: " + err.Error())
		})
		return
	}

	markdown := string(content)

	// Highlight glossary terms inline (Wikipedia-style clickable links)
	highlightedMarkdown := HighlightGlossaryTerms(markdown)

	fyne.Do(func() {
		app.contentText.ParseMarkdown(highlightedMarkdown)
		// Add click handlers to glossary:// links
		app.setupGlossaryClickHandlers()
	})
	app.currentFile = path
	app.history.AddRecent(path)
	app.updateModuleShortcuts(path)

	// Update breadcrumbs
	fyne.Do(func() {
		app.breadcrumbs.SetPath(path, app.docsPath)
	})

	// Update ToC (use original markdown to avoid bold markers in headings)
	fyne.Do(func() {
		app.toc.ParseMarkdown(markdown)
		if app.layoutManager != nil {
			app.layoutManager.SetTocVisible(app.toc.HeadingCount() >= 3)
		}
	})

	// Detect glossary terms and update metadata
	terms := DetectGlossaryTerms(markdown)
	meta := app.searchIndex.GetDocMetadata(path)
	if meta != nil {
		fyne.Do(func() {
			app.docMetadata.SetMetadata(meta.Title, meta.Category, meta.ReadingTime, terms)
		})
	}
	fyne.Do(func() {
		app.glossaryPills.SetTerms(terms)
	})
}

// setupGlossaryClickHandlers iterates through RichText segments and adds click handlers for glossary:// links
func (app *EmbeddedDocsApp) setupGlossaryClickHandlers() {
	for _, seg := range app.contentText.Segments {
		if hyperlink, ok := seg.(*widget.HyperlinkSegment); ok {
			if hyperlink.URL != nil && hyperlink.URL.Scheme == "glossary" {
				term := hyperlink.URL.Host
				if term == "" {
					term = strings.TrimPrefix(hyperlink.URL.Path, "/")
				}
				// URL decode the term
				if decoded, err := url.QueryUnescape(term); err == nil {
					term = decoded
				}
				// Capture term for closure
				termCopy := term
				hyperlink.OnTapped = func() {
					sharedWidgets.ShowGlossary(termCopy, app.window)
				}
			}
		}
	}
}

// toggleFavorite adds or removes a document from favorites
func (app *EmbeddedDocsApp) toggleFavorite(path string) {
	app.history.ToggleFavorite(path)
}

// updateModuleShortcuts refreshes the module quick-access panel for the given path.
func (app *EmbeddedDocsApp) updateModuleShortcuts(path string) {
	if app.moduleShortcuts == nil {
		return
	}
	modulePath := app.findModulePath(path)
	fyne.Do(func() {
		app.moduleShortcuts.SetModulePath(modulePath)
	})
}

// findModulePath resolves the module root for a given file or folder path.
func (app *EmbeddedDocsApp) findModulePath(path string) string {
	if app.docsPath == "" || path == "" {
		return ""
	}

	target := path
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		target = filepath.Dir(path)
	}

	rel, err := filepath.Rel(app.docsPath, target)
	if err != nil || rel == "." {
		return ""
	}
	if strings.HasPrefix(rel, "..") {
		return ""
	}

	parts := strings.Split(filepath.ToSlash(rel), "/")
	if len(parts) == 0 {
		return ""
	}
	root := parts[0]
	if !strings.HasPrefix(strings.ToLower(root), "module") {
		return ""
	}

	modulePath := filepath.Join(app.docsPath, root)
	if info, err := os.Stat(modulePath); err == nil && info.IsDir() {
		return modulePath
	}

	return ""
}

// consumeSuppressSelect returns true if selection should be ignored once.
func (app *EmbeddedDocsApp) consumeSuppressSelect(uid string) bool {
	if app.suppressSelect == nil {
		return false
	}
	if app.suppressSelect[uid] {
		delete(app.suppressSelect, uid)
		return true
	}
	return false
}

// showSearch displays the search dialog
func (app *EmbeddedDocsApp) showSearch() {
	app.searchDialog.Show()
}

// scrollToSection scrolls to a section anchor (best effort)
func (app *EmbeddedDocsApp) scrollToSection(anchor string) {
	// Fyne doesn't have built-in anchor scrolling
	// Highlight the section in ToC as visual feedback
	fyne.Do(func() {
		app.toc.SetCurrentSection(anchor)
	})
}

// navigateToFolder expands the tree to show a folder
func (app *EmbeddedDocsApp) navigateToFolder(path string) {
	if app.tree == nil || app.docsPath == "" || path == "" {
		return
	}
	app.updateModuleShortcuts(path)

	targetPath := path
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		if strings.HasSuffix(strings.ToLower(path), ".md") {
			app.loadDocument(path)
		}
		targetPath = filepath.Dir(path)
	}

	relPath, err := filepath.Rel(app.docsPath, targetPath)
	if err != nil {
		return
	}

	parts := strings.Split(filepath.ToSlash(relPath), "/")
	current := app.docsPath

	fyne.Do(func() {
		for _, part := range parts {
			if part == "" || part == "." {
				continue
			}
			current = filepath.Join(current, part)
			if _, ok := app.pathMap[current]; ok {
				app.tree.OpenBranch(current)
			}
		}
	})
}

// scanDocsDirectory recursively scans the docs directory
func (app *EmbeddedDocsApp) scanDocsDirectory() []*docEntry {
	if app.docsPath == "" {
		return []*docEntry{{
			name:  "Docs not found",
			path:  "notfound",
			isDir: false,
		}}
	}

	var entries []*docEntry
	files, err := os.ReadDir(app.docsPath)
	if err != nil {
		return entries
	}

	for _, f := range files {
		entry := app.scanEntry(filepath.Join(app.docsPath, f.Name()), f)
		if entry != nil {
			entries = append(entries, entry)
		}
	}

	app.sortEntries(entries, app.docsPath)
	return entries
}

// scanEntry recursively scans a directory entry
func (app *EmbeddedDocsApp) scanEntry(path string, info os.DirEntry) *docEntry {
	name := info.Name()

	// Skip hidden files and non-markdown files
	if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
		return nil
	}

	if info.IsDir() {
		// Recursively scan directory
		files, err := os.ReadDir(path)
		if err != nil {
			return nil
		}

		var children []*docEntry
		for _, f := range files {
			child := app.scanEntry(filepath.Join(path, f.Name()), f)
			if child != nil {
				children = append(children, child)
			}
		}

		// Only include directories that have markdown content
		if len(children) > 0 {
			return &docEntry{
				name:     name,
				path:     path,
				isDir:    true,
				children: children,
			}
		}
		return nil
	}

	// Only include markdown files
	if strings.HasSuffix(strings.ToLower(name), ".md") {
		return &docEntry{
			name:  name,
			path:  path,
			isDir: false,
		}
	}

	return nil
}

// sortEntries orders documentation entries based on curriculum-first rules.
func (app *EmbeddedDocsApp) sortEntries(entries []*docEntry, parentPath string) {
	sort.SliceStable(entries, func(i, j int) bool {
		return app.entryLess(entries[i], entries[j], parentPath)
	})

	for _, entry := range entries {
		if entry.isDir && len(entry.children) > 0 {
			app.sortEntries(entry.children, entry.path)
		}
	}
}

func (app *EmbeddedDocsApp) entryLess(a, b *docEntry, parentPath string) bool {
	if parentPath == app.docsPath {
		rankA := app.rootRank(a)
		rankB := app.rootRank(b)
		if rankA != rankB {
			return rankA < rankB
		}
		return strings.ToLower(a.name) < strings.ToLower(b.name)
	}

	if app.isModuleDir(parentPath) {
		if a.isDir != b.isDir {
			return a.isDir
		}
		if !a.isDir {
			rankA := moduleFileRank(a)
			rankB := moduleFileRank(b)
			if rankA != rankB {
				return rankA < rankB
			}
		}
	}

	// Default: directories first, then alphabetical
	if a.isDir != b.isDir {
		return a.isDir
	}
	return strings.ToLower(a.name) < strings.ToLower(b.name)
}

func (app *EmbeddedDocsApp) rootRank(entry *docEntry) int {
	nameLower := strings.ToLower(entry.name)

	if entry.isDir {
		if idx, ok := moduleIndex(nameLower); ok {
			return idx
		}
		if nameLower == "research-papers" {
			return 100
		}
	}

	switch nameLower {
	case "readme.md":
		return 110
	case "modules.md":
		return 111
	}

	return 200
}

func (app *EmbeddedDocsApp) isModuleDir(path string) bool {
	if path == "" {
		return false
	}
	base := strings.ToLower(filepath.Base(path))
	if _, ok := moduleIndex(base); !ok {
		return false
	}
	if app.docsPath == "" {
		return false
	}
	rel, err := filepath.Rel(app.docsPath, path)
	if err != nil || rel == "." {
		return false
	}
	if strings.HasPrefix(rel, "..") {
		return false
	}
	return true
}

func moduleIndex(name string) (int, bool) {
	if !strings.HasPrefix(name, "module") {
		return 0, false
	}
	rest := name[len("module"):]
	digits := ""
	for i := 0; i < len(rest); i++ {
		if rest[i] < '0' || rest[i] > '9' {
			break
		}
		digits += string(rest[i])
	}
	if digits == "" {
		return 0, false
	}
	idx, err := strconv.Atoi(digits)
	if err != nil {
		return 0, false
	}
	return idx, true
}

func moduleFileRank(entry *docEntry) int {
	if entry.isDir {
		return 100
	}
	switch strings.ToLower(entry.name) {
	case "eli5.md":
		return 0
	case "physics.md":
		return 1
	case "features.md":
		return 2
	case "opensource-tools.md":
		return 3
	default:
		return 10
	}
}

// Start is called when this demo tab is selected
func (app *EmbeddedDocsApp) Start() {
	app.EmbeddedAppBase.Start()
	// No background processes to start
}

// Stop is called when this demo tab is deselected
func (app *EmbeddedDocsApp) Stop() {
	// No background processes to stop
	app.EmbeddedAppBase.Stop()
}
