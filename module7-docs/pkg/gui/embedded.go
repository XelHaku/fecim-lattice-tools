// pkg/gui/embedded.go
// Embeddable documentation viewer for the unified visualizer
package gui

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	sharedWidgets "fecim-lattice-tools/shared/widgets"
)

// EmbeddedDocsApp is the embeddable documentation viewer
type EmbeddedDocsApp struct {
	// Existing
	content     fyne.CanvasObject
	currentFile string
	docsPath    string

	// New components
	window        fyne.Window
	searchIndex   *SearchIndex
	history       *DocsHistory
	layoutManager *LayoutManager

	// UI components
	tree          *widget.Tree
	contentText   *widget.RichText
	contentScroll *container.Scroll
	breadcrumbs   *BreadcrumbWidget
	toc           *TableOfContentsWidget
	docMetadata   *DocumentMetadataWidget
	searchDialog  *SearchDialog

	// State
	pathMap map[string]*docEntry
	docs    []*docEntry
}

// NewEmbeddedDocsApp creates a new embedded docs app instance
func NewEmbeddedDocsApp() *EmbeddedDocsApp {
	return &EmbeddedDocsApp{}
}

// docEntry represents a documentation file or folder
type docEntry struct {
	name     string
	path     string
	isDir    bool
	children []*docEntry
}

// BuildContent creates the UI content for embedding in the main app
func (app *EmbeddedDocsApp) BuildContent(fyneApp fyne.App, window fyne.Window) fyne.CanvasObject {
	app.window = window
	app.docsPath = findDocsPath()

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

	app.content = app.layoutManager.BuildLayout()

	// Setup keyboard shortcut for search
	SetupSearchShortcut(window, app.searchDialog)

	return app.content
}

// createUIComponents initializes all UI widgets
func (app *EmbeddedDocsApp) createUIComponents() {
	// Content viewer
	app.contentText = widget.NewRichTextFromMarkdown("# FeCIM Documentation\n\nSelect a document from the tree.")
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

	// Search dialog
	app.searchDialog = NewSearchDialog(app.searchIndex, app.window, func(path string) {
		app.loadDocument(path)
	})
}

// buildSidebar creates the left sidebar with tree
func (app *EmbeddedDocsApp) buildSidebar() fyne.CanvasObject {
	// Scan docs and build tree
	app.docs = app.scanDocsDirectory()
	app.pathMap = make(map[string]*docEntry)
	app.buildPathMap(app.docs)
	app.tree = app.createDocTree()

	return container.NewBorder(
		widget.NewLabelWithStyle("Documentation", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		nil, nil, nil,
		app.tree,
	)
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
	return container.NewBorder(
		widget.NewLabelWithStyle("On This Page", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		nil, nil, nil,
		app.toc,
	)
}

// buildTopBar creates the top action bar with search
func (app *EmbeddedDocsApp) buildTopBar() fyne.CanvasObject {
	searchBtn := widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
		app.showSearch()
	})

	tocToggleBtn := widget.NewButtonWithIcon("", theme.ListIcon(), func() {
		app.layoutManager.ToggleToc()
	})

	return container.NewBorder(
		nil, nil,
		widget.NewLabel("FeCIM Documentation"),
		container.NewHBox(tocToggleBtn, searchBtn),
		nil,
	)
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
			label := widget.NewLabel("Document")
			starBtn := widget.NewButtonWithIcon("", theme.ContentAddIcon(), nil)
			starBtn.Importance = widget.LowImportance
			starBtn.Hidden = true
			return container.NewBorder(nil, nil, icon, starBtn, label)
		},
		// Update - update tree node with data
		func(uid widget.TreeNodeID, branch bool, node fyne.CanvasObject) {
			box := node.(*fyne.Container)
			// Find widgets by type - don't rely on container object ordering
			var icon *widget.Icon
			var label *widget.Label
			var starBtn *widget.Button
			for _, obj := range box.Objects {
				switch v := obj.(type) {
				case *widget.Icon:
					icon = v
				case *widget.Label:
					label = v
				case *widget.Button:
					starBtn = v
				}
			}

			if entry, ok := app.pathMap[uid]; ok {
				label.SetText(entry.name)
				if entry.isDir {
					icon.SetResource(theme.FolderIcon())
					starBtn.Hidden = true
				} else {
					icon.SetResource(theme.DocumentIcon())
					starBtn.Hidden = false

					// Update star icon based on favorite status
					if app.history.IsFavorite(entry.path) {
						starBtn.SetIcon(theme.ContentRemoveIcon())
					} else {
						starBtn.SetIcon(theme.ContentAddIcon())
					}

					// Capture path for closure
					path := entry.path
					starBtn.OnTapped = func() {
						app.toggleFavorite(path)
						app.tree.Refresh()
					}
				}
			}
		},
	)

	// Handle selection - load markdown file or toggle folder
	tree.OnSelected = func(uid widget.TreeNodeID) {
		if entry, ok := app.pathMap[uid]; ok {
			if entry.isDir {
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

	// Update breadcrumbs
	fyne.Do(func() {
		app.breadcrumbs.SetPath(path, app.docsPath)
	})

	// Update ToC (use original markdown to avoid bold markers in headings)
	fyne.Do(func() {
		app.toc.ParseMarkdown(markdown)
	})

	// Detect glossary terms and update metadata
	terms := DetectGlossaryTerms(markdown)
	meta := app.searchIndex.GetDocMetadata(path)
	if meta != nil {
		fyne.Do(func() {
			app.docMetadata.SetMetadata(meta.Title, meta.Category, meta.ReadingTime, terms)
		})
	}
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
	fyne.Do(func() {
		app.tree.OpenBranch(path)
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

// findDocsPath locates the docs directory
func findDocsPath() string {
	// Try relative to working directory
	candidates := []string{
		"docs",
		"../docs",
		"../../docs",
	}

	// Also try relative to executable
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, "docs"),
			filepath.Join(exeDir, "..", "docs"),
		)
	}

	for _, candidate := range candidates {
		if abs, err := filepath.Abs(candidate); err == nil {
			if info, err := os.Stat(abs); err == nil && info.IsDir() {
				return abs
			}
		}
	}

	return ""
}

// Start is called when this demo tab is selected
func (app *EmbeddedDocsApp) Start() {
	// No background processes to start
}

// Stop is called when this demo tab is deselected
func (app *EmbeddedDocsApp) Stop() {
	// No background processes to stop
}
