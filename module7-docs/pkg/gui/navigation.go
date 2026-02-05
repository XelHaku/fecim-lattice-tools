package gui

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	fecimTheme "fecim-lattice-tools/shared/theme"
)

// BreadcrumbSegment represents a single clickable segment in the breadcrumb trail
type BreadcrumbSegment struct {
	Label string // Display name
	Path  string // Full path to navigate to
}

// BreadcrumbWidget provides hierarchical navigation through document folders
type BreadcrumbWidget struct {
	widget.BaseWidget
	segments   []BreadcrumbSegment
	onNavigate func(path string)
	container  *fyne.Container
}

// NewBreadcrumbWidget creates a new breadcrumb navigation widget
func NewBreadcrumbWidget(onNavigate func(path string)) *BreadcrumbWidget {
	b := &BreadcrumbWidget{
		onNavigate: onNavigate,
		segments:   []BreadcrumbSegment{},
	}
	b.ExtendBaseWidget(b)
	return b
}

// SetPath parses a file path into clickable breadcrumb segments
func (b *BreadcrumbWidget) SetPath(currentPath, docsRoot string) {
	b.segments = []BreadcrumbSegment{}

	// Always start with Home
	b.segments = append(b.segments, BreadcrumbSegment{
		Label: "Home",
		Path:  docsRoot,
	})

	// Parse relative path components
	relPath, err := filepath.Rel(docsRoot, currentPath)
	if err != nil || relPath == "." || relPath == "" {
		b.refresh()
		b.Refresh()
		return
	}

	// Build segments from path components
	parts := strings.Split(filepath.ToSlash(relPath), "/")
	currentBuildPath := docsRoot

	for i, part := range parts {
		if part == "" || part == "." {
			continue
		}

		currentBuildPath = filepath.Join(currentBuildPath, part)

		// Skip file extension for last segment (file name)
		label := part
		if i == len(parts)-1 && filepath.Ext(part) != "" {
			label = strings.TrimSuffix(part, filepath.Ext(part))
		}

		// Humanize label: replace underscores/hyphens with spaces, title case
		label = strings.ReplaceAll(label, "_", " ")
		label = strings.ReplaceAll(label, "-", " ")
		label = strings.Title(label)

		b.segments = append(b.segments, BreadcrumbSegment{
			Label: label,
			Path:  currentBuildPath,
		})
	}

	b.refresh()
	b.Refresh()
}

// CreateRenderer implements the fyne.Widget interface
func (b *BreadcrumbWidget) CreateRenderer() fyne.WidgetRenderer {
	b.refresh()
	return widget.NewSimpleRenderer(b.container)
}

func (b *BreadcrumbWidget) refresh() {
	items := []fyne.CanvasObject{}

	for i, segment := range b.segments {
		// Create clickable label for segment
		segmentPath := segment.Path // Capture for closure
		label := widget.NewButton(segment.Label, func() {
			if b.onNavigate != nil {
				b.onNavigate(segmentPath)
			}
		})
		label.Importance = widget.LowImportance

		items = append(items, label)

		// Add separator (except after last segment)
		if i < len(b.segments)-1 {
			separator := canvas.NewText(" > ", fecimTheme.ColorTextDim)
			separator.TextSize = 14
			items = append(items, separator)
		}
	}

	b.container = container.NewHBox(items...)
	// Note: Don't call b.Refresh() here - it causes infinite recursion
	// CreateRenderer -> refresh -> Refresh -> CreateRenderer -> ...
}

// TOCHeading represents a heading in the table of contents
type TOCHeading struct {
	Text   string
	Level  int    // 1-6 for h1-h6
	Anchor string // lowercase with hyphens for scrolling
}

// TableOfContentsWidget displays an interactive table of contents for the current document
type TableOfContentsWidget struct {
	widget.BaseWidget
	headings       []TOCHeading
	onSelect       func(heading string)
	currentSection string
	container      *fyne.Container
}

// NewTableOfContentsWidget creates a new table of contents widget
func NewTableOfContentsWidget(onSelect func(heading string)) *TableOfContentsWidget {
	toc := &TableOfContentsWidget{
		onSelect:  onSelect,
		headings:  []TOCHeading{},
		container: container.NewVBox(),
	}
	toc.ExtendBaseWidget(toc)
	return toc
}

// ParseMarkdown extracts headings from markdown content
func (toc *TableOfContentsWidget) ParseMarkdown(content string) {
	toc.headings = []TOCHeading{}

	// Regex to match markdown headings: ^(#{1,6})\s+(.+)$
	headingRegex := regexp.MustCompile(`(?m)^(#{1,6})\s+(.+)$`)
	matches := headingRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) != 3 {
			continue
		}

		level := len(match[1]) // Count number of # symbols
		text := strings.TrimSpace(match[2])

		// Generate anchor: lowercase, replace spaces with hyphens, remove special chars
		anchor := strings.ToLower(text)
		anchor = strings.ReplaceAll(anchor, " ", "-")
		anchor = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(anchor, "")

		toc.headings = append(toc.headings, TOCHeading{
			Text:   text,
			Level:  level,
			Anchor: anchor,
		})
	}

	toc.refresh()
	toc.Refresh()
}

// SetCurrentSection highlights the current section in the ToC
func (toc *TableOfContentsWidget) SetCurrentSection(anchor string) {
	toc.currentSection = anchor
	toc.refresh()
	toc.Refresh()
}

// CreateRenderer implements the fyne.Widget interface
func (toc *TableOfContentsWidget) CreateRenderer() fyne.WidgetRenderer {
	toc.refresh()
	return widget.NewSimpleRenderer(toc.container)
}

func (toc *TableOfContentsWidget) refresh() {
	items := []fyne.CanvasObject{}

	// Only show if document has 3+ headings
	if len(toc.headings) < 3 {
		toc.container = container.NewVBox()
		return
	}

	// Render headings with indentation
	for _, heading := range toc.headings {
		anchor := heading.Anchor // Capture for closure

		// Create button for heading
		btn := widget.NewButton(heading.Text, func() {
			if toc.onSelect != nil {
				toc.onSelect(anchor)
			}
		})
		btn.Importance = widget.LowImportance

		// Apply indentation based on level
		var item fyne.CanvasObject
		if heading.Level == 1 || heading.Level == 2 {
			// h1 and h2: normal, no bullet
			item = btn
		} else {
			// h3-h6: indented with bullet
			bullet := canvas.NewText("∙ ", fecimTheme.ColorTextDim)
			bullet.TextSize = 14
			indentation := container.NewHBox()
			for i := 0; i < (heading.Level-2)*2; i++ {
				indentation.Add(canvas.NewText(" ", theme.ForegroundColor()))
			}
			item = container.NewHBox(indentation, bullet, btn)
		}

		// Highlight current section
		if heading.Anchor == toc.currentSection {
			// Create highlighted text instead
			highlighted := canvas.NewText(heading.Text, fecimTheme.ColorPrimary)
			highlighted.TextStyle = fyne.TextStyle{Bold: true}
			highlighted.TextSize = 14
			item = highlighted
		}

		items = append(items, item)
	}

	toc.container = container.NewVBox(items...)
	// Note: Don't call toc.Refresh() here - it causes infinite recursion
	// CreateRenderer -> refresh -> Refresh -> CreateRenderer -> ...
}

// ModuleShortcutsPanel provides quick access to the current module's core pages.
type ModuleShortcutsPanel struct {
	widget.BaseWidget
	modulePath string
	onSelect   func(path string)
	container  *fyne.Container
}

// NewModuleShortcutsPanel creates a new module shortcuts panel.
func NewModuleShortcutsPanel(onSelect func(string)) *ModuleShortcutsPanel {
	panel := &ModuleShortcutsPanel{
		onSelect:  onSelect,
		container: container.NewVBox(),
	}
	panel.ExtendBaseWidget(panel)
	return panel
}

// SetModulePath updates the active module path and refreshes shortcuts.
func (m *ModuleShortcutsPanel) SetModulePath(path string) {
	if m.modulePath == path {
		return
	}
	m.modulePath = path
	m.refresh()
	m.Refresh()
}

// CreateRenderer implements fyne.Widget.
func (m *ModuleShortcutsPanel) CreateRenderer() fyne.WidgetRenderer {
	m.refresh()
	return widget.NewSimpleRenderer(m.container)
}

func (m *ModuleShortcutsPanel) refresh() {
	items := []fyne.CanvasObject{}

	title := widget.NewLabelWithStyle("Current Module", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	items = append(items, title)

	if m.modulePath == "" {
		note := widget.NewLabel("Select a module page to enable shortcuts.")
		note.Wrapping = fyne.TextWrapWord
		items = append(items, note)
		m.container = container.NewVBox(items...)
		return
	}

	shortcuts := []struct {
		label    string
		filename string
	}{
		{label: "ELI5", filename: "ELI5.md"},
		{label: "Physics", filename: "PHYSICS.md"},
		{label: "Features", filename: "FEATURES.md"},
		{label: "Tools", filename: "OPENSOURCE-TOOLS.md"},
	}

	for _, shortcut := range shortcuts {
		path := filepath.Join(m.modulePath, shortcut.filename)
		btn := widget.NewButton(shortcut.label, func() {
			if m.onSelect != nil {
				m.onSelect(path)
			}
		})
		btn.Importance = widget.LowImportance
		if _, err := os.Stat(path); err != nil {
			btn.Disable()
		}
		items = append(items, btn)
	}

	m.container = container.NewVBox(items...)
	// Note: Don't call m.Refresh() here - it causes infinite recursion
}

// QuickAccessPanel provides quick access to favorite documents
type QuickAccessPanel struct {
	widget.BaseWidget
	favorites        []string // Starred docs
	onSelect         func(path string)
	onToggleFavorite func(path string)
	container        *fyne.Container
}

// NewQuickAccessPanel creates a new quick access panel
func NewQuickAccessPanel(onSelect func(string), onToggleFavorite func(string)) *QuickAccessPanel {
	qap := &QuickAccessPanel{
		favorites:        []string{},
		onSelect:         onSelect,
		onToggleFavorite: onToggleFavorite,
		container:        container.NewVBox(),
	}
	qap.ExtendBaseWidget(qap)
	return qap
}

// ToggleFavorite adds or removes a document from favorites
func (qap *QuickAccessPanel) ToggleFavorite(path string) {
	for i, p := range qap.favorites {
		if p == path {
			// Remove from favorites
			qap.favorites = append(qap.favorites[:i], qap.favorites[i+1:]...)
			if qap.onToggleFavorite != nil {
				qap.onToggleFavorite(path)
			}
			qap.refresh()
			qap.Refresh()
			return
		}
	}

	// Add to favorites
	qap.favorites = append(qap.favorites, path)
	if qap.onToggleFavorite != nil {
		qap.onToggleFavorite(path)
	}
	qap.refresh()
	qap.Refresh()
}

// IsFavorite checks if a document is in favorites
func (qap *QuickAccessPanel) IsFavorite(path string) bool {
	for _, p := range qap.favorites {
		if p == path {
			return true
		}
	}
	return false
}

// CreateRenderer implements the fyne.Widget interface
func (qap *QuickAccessPanel) CreateRenderer() fyne.WidgetRenderer {
	qap.refresh()
	return widget.NewSimpleRenderer(qap.container)
}

func (qap *QuickAccessPanel) refresh() {
	items := []fyne.CanvasObject{}

	// Favorites section
	if len(qap.favorites) > 0 {
		favTitle := container.NewHBox(
			canvas.NewText("★", fecimTheme.ColorWarning),
			widget.NewLabel("Favorites"),
		)
		items = append(items, favTitle)

		for _, path := range qap.favorites {
			p := path // Capture for closure
			label := filepath.Base(path)
			label = strings.TrimSuffix(label, filepath.Ext(label))
			label = strings.ReplaceAll(label, "_", " ")
			label = strings.ReplaceAll(label, "-", " ")

			btn := widget.NewButton(label, func() {
				if qap.onSelect != nil {
					qap.onSelect(p)
				}
			})
			btn.Importance = widget.LowImportance
			items = append(items, btn)
		}

	}

	qap.container = container.NewVBox(items...)
	// Note: Don't call qap.Refresh() here - it causes infinite recursion
	// CreateRenderer -> refresh -> Refresh -> CreateRenderer -> ...
}
