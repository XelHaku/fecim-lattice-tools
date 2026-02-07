package gui

import (
	"fmt"
	"image/color"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	fecimTheme "fecim-lattice-tools/shared/theme"
	"fecim-lattice-tools/shared/widgets"
)

// GlossaryPillsWidget displays detected glossary terms as clickable pills
type GlossaryPillsWidget struct {
	widget.BaseWidget
	terms       []string
	onTermClick func(term string)
	window      fyne.Window
	container   *fyne.Container
}

// NewGlossaryPillsWidget creates a new glossary pills widget
func NewGlossaryPillsWidget(window fyne.Window) *GlossaryPillsWidget {
	w := &GlossaryPillsWidget{
		window:    window,
		container: container.NewHBox(),
	}
	w.ExtendBaseWidget(w)
	return w
}

// SetTerms updates the displayed terms
func (g *GlossaryPillsWidget) SetTerms(terms []string) {
	g.terms = terms
	g.rebuild()
	g.Refresh()
}

// DetectTerms scans markdown content for glossary terms
func (g *GlossaryPillsWidget) DetectTerms(markdownContent string) []string {
	return DetectGlossaryTerms(markdownContent)
}

// CreateRenderer implements fyne.Widget
func (g *GlossaryPillsWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(g.container)
}

func (g *GlossaryPillsWidget) rebuild() {
	g.container.Objects = nil

	for _, term := range g.terms {
		termCopy := term // Capture for closure
		btn := widget.NewButton(term, func() {
			if g.onTermClick != nil {
				g.onTermClick(termCopy)
			} else {
				// Default behavior: show glossary popup
				widgets.ShowGlossary(termCopy, g.window)
			}
		})
		btn.Importance = widget.LowImportance
		g.container.Add(btn)
	}
}

// DetectGlossaryTerms scans content for glossary terms
func DetectGlossaryTerms(content string) []string {
	// Convert content to lowercase for case-insensitive matching
	lowerContent := strings.ToLower(content)

	// Find unique terms that appear as whole words
	foundTerms := make(map[string]bool)

	for _, entry := range widgets.TermsData {
		termLower := strings.ToLower(entry.Term)
		// Create regex for whole-word match
		pattern := `\b` + regexp.QuoteMeta(termLower) + `\b`
		matched, err := regexp.MatchString(pattern, lowerContent)
		if err == nil && matched {
			// Store with original casing
			foundTerms[entry.Term] = true
		}
	}

	// Convert to sorted slice
	terms := make([]string, 0, len(foundTerms))
	for term := range foundTerms {
		terms = append(terms, term)
	}
	sort.Strings(terms)

	return terms
}

// glossaryTermEntry is used for sorting terms by length
type glossaryTermEntry struct {
	lower    string
	original string
}

// HighlightGlossaryTerms wraps glossary terms in markdown bold formatting
// Returns the modified markdown content with terms highlighted
func HighlightGlossaryTerms(content string) string {
	// Build a map of terms to their original casing
	termMap := make(map[string]string) // lowercase -> original
	for _, entry := range widgets.TermsData {
		termMap[strings.ToLower(entry.Term)] = entry.Term
	}

	// Sort terms by length (longest first) to avoid partial replacements
	// e.g., "Preisach Model" should be matched before "Model"
	var sortedTerms []glossaryTermEntry
	for lower, original := range termMap {
		sortedTerms = append(sortedTerms, glossaryTermEntry{lower, original})
	}
	sort.Slice(sortedTerms, func(i, j int) bool {
		return len(sortedTerms[i].lower) > len(sortedTerms[j].lower)
	})

	// Track which positions have already been modified to avoid double-highlighting
	// We'll use a simple approach: process line by line, skip code blocks and headers
	lines := strings.Split(content, "\n")
	var result []string

	inCodeBlock := false
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Toggle code block state
		if strings.HasPrefix(trimmedLine, "```") {
			inCodeBlock = !inCodeBlock
			result = append(result, line)
			continue
		}

		// Skip code blocks
		if inCodeBlock {
			result = append(result, line)
			continue
		}

		// Skip headers (they're already styled)
		if strings.HasPrefix(trimmedLine, "#") {
			result = append(result, line)
			continue
		}

		// Skip table separator lines (|---|---|)
		if strings.Contains(line, "|") && strings.Contains(line, "---") {
			result = append(result, line)
			continue
		}

		// Skip table header/data rows that start with | (formatted tables)
		if strings.HasPrefix(trimmedLine, "|") {
			result = append(result, line)
			continue
		}

		// Highlight terms in this line (including lines with links)
		highlightedLine := highlightTermsInLine(line, sortedTerms)
		result = append(result, highlightedLine)
	}

	return strings.Join(result, "\n")
}

// highlightTermsInLine highlights glossary terms in a single line
func highlightTermsInLine(line string, terms []glossaryTermEntry) string {
	// Track positions that should be skipped (inside links, bold, code)
	skipPositions := make([]bool, len(line))

	// Mark positions inside markdown links [text](url) as skip
	linkRe := regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)
	for _, match := range linkRe.FindAllStringIndex(line, -1) {
		for i := match[0]; i < match[1] && i < len(skipPositions); i++ {
			skipPositions[i] = true
		}
	}

	// Mark positions inside inline code `code` as skip
	codeRe := regexp.MustCompile("`[^`]+`")
	for _, match := range codeRe.FindAllStringIndex(line, -1) {
		for i := match[0]; i < match[1] && i < len(skipPositions); i++ {
			skipPositions[i] = true
		}
	}

	// Mark positions inside existing bold **text** as skip
	boldRe := regexp.MustCompile(`\*\*[^*]+\*\*`)
	for _, match := range boldRe.FindAllStringIndex(line, -1) {
		for i := match[0]; i < match[1] && i < len(skipPositions); i++ {
			skipPositions[i] = true
		}
	}

	type replacement struct {
		start, end int
		newText    string
	}
	var replacements []replacement

	for _, term := range terms {
		// Create case-insensitive regex for whole-word match
		pattern := `(?i)\b` + regexp.QuoteMeta(term.lower) + `\b`
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}

		matches := re.FindAllStringIndex(line, -1)
		for _, match := range matches {
			start, end := match[0], match[1]

			// Check if any position in this match should be skipped
			shouldSkip := false
			for i := start; i < end && i < len(skipPositions); i++ {
				if skipPositions[i] {
					shouldSkip = true
					break
				}
			}
			if shouldSkip {
				continue
			}

			// Mark positions as used
			for i := start; i < end && i < len(skipPositions); i++ {
				skipPositions[i] = true
			}

			// Get the actual text (preserve original casing from content)
			originalText := line[start:end]
			escapedTerm := url.PathEscape(term.original)
			// Use markdown link with glossary:// scheme for clickable terms
			replacements = append(replacements, replacement{
				start:   start,
				end:     end,
				newText: "[" + originalText + "](glossary://" + escapedTerm + ")",
			})
		}
	}

	// Apply replacements from end to start to preserve positions
	sort.Slice(replacements, func(i, j int) bool {
		return replacements[i].start > replacements[j].start
	})

	for _, r := range replacements {
		line = line[:r.start] + r.newText + line[r.end:]
	}

	return line
}

// CategoryBadge displays a colored category indicator
type CategoryBadge struct {
	widget.BaseWidget
	category string
	color    color.Color
}

// CategoryColors maps category names to colors
var CategoryColors = map[string]color.Color{
	"ELI5":     color.RGBA{76, 175, 80, 255},  // Green #4CAF50
	"Physics":  color.RGBA{0, 188, 212, 255},  // Cyan #00BCD4
	"Research": color.RGBA{255, 152, 0, 255},  // Orange #FF9800
	"Demo":     color.RGBA{156, 39, 176, 255}, // Purple #9C27B0
	"Guide":    color.RGBA{33, 150, 243, 255}, // Blue #2196F3
	"Module":   color.RGBA{96, 125, 139, 255}, // Blue Grey #607D8B
}

// NewCategoryBadge creates a new category badge
func NewCategoryBadge(category string) *CategoryBadge {
	badgeColor, ok := CategoryColors[category]
	if !ok {
		badgeColor = fecimTheme.ColorPrimary // Default to cyan
	}

	b := &CategoryBadge{
		category: category,
		color:    badgeColor,
	}
	b.ExtendBaseWidget(b)
	return b
}

// SetCategory updates the badge label + color.
func (b *CategoryBadge) SetCategory(category string) {
	b.category = category
	badgeColor, ok := CategoryColors[category]
	if !ok {
		badgeColor = fecimTheme.ColorPrimary
	}
	b.color = badgeColor
	b.Refresh()
}

type categoryBadgeRenderer struct {
	badge  *CategoryBadge
	border *canvas.Rectangle
	label  *canvas.Text
	root   *fyne.Container
}

func (r *categoryBadgeRenderer) Layout(size fyne.Size) {
	r.root.Resize(size)
}

func (r *categoryBadgeRenderer) MinSize() fyne.Size {
	return r.root.MinSize()
}

func (r *categoryBadgeRenderer) Refresh() {
	r.border.FillColor = r.badge.color
	r.border.Refresh()
	r.label.Text = r.badge.category
	r.label.Refresh()
}

func (r *categoryBadgeRenderer) BackgroundColor() color.Color { return color.Transparent }
func (r *categoryBadgeRenderer) Objects() []fyne.CanvasObject { return []fyne.CanvasObject{r.root} }
func (r *categoryBadgeRenderer) Destroy()                     {}

// CreateRenderer implements fyne.Widget.
func (b *CategoryBadge) CreateRenderer() fyne.WidgetRenderer {
	border := canvas.NewRectangle(b.color)
	border.SetMinSize(fyne.NewSize(4, 0))

	label := canvas.NewText(b.category, theme.ForegroundColor())
	label.TextStyle = fyne.TextStyle{Bold: true}

	content := container.NewBorder(
		nil, nil,
		border, nil,
		container.NewPadded(label),
	)

	return &categoryBadgeRenderer{badge: b, border: border, label: label, root: content}
}

// DocumentMetadataWidget displays document metadata with category and reading time
type DocumentMetadataWidget struct {
	widget.BaseWidget
	title       string
	category    string
	readingTime int // minutes
	termCount   int
	window      fyne.Window
	container   *fyne.Container
}

// NewDocumentMetadataWidget creates a new document metadata widget
func NewDocumentMetadataWidget(window fyne.Window) *DocumentMetadataWidget {
	w := &DocumentMetadataWidget{
		window:    window,
		container: container.NewHBox(),
	}
	w.ExtendBaseWidget(w)
	return w
}

// SetMetadata updates the metadata display
func (d *DocumentMetadataWidget) SetMetadata(title, category string, readingTime int, terms []string) {
	d.title = title
	d.category = category
	d.readingTime = readingTime
	d.termCount = 0
	if len(terms) > 0 {
		d.termCount = len(terms)
	}
	d.rebuild()
	d.Refresh()
}

// CreateRenderer implements fyne.Widget
func (d *DocumentMetadataWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(d.container)
}

func (d *DocumentMetadataWidget) rebuild() {
	d.container.Objects = nil

	if d.category != "" {
		badge := NewCategoryBadge(d.category)
		d.container.Add(badge)
	}

	if d.readingTime > 0 {
		if len(d.container.Objects) > 0 {
			d.container.Add(widget.NewLabel("|"))
		}
		readingLabel := widget.NewLabel(formatReadingTime(d.readingTime))
		readingLabel.TextStyle = fyne.TextStyle{Italic: true}
		d.container.Add(readingLabel)
	}

	if d.termCount > 0 {
		if len(d.container.Objects) > 0 {
			d.container.Add(widget.NewLabel("|"))
		}
		termLabel := widget.NewLabel(formatTermCount(d.termCount))
		termLabel.TextStyle = fyne.TextStyle{Italic: true}
		d.container.Add(termLabel)
	}
}

func formatReadingTime(minutes int) string {
	if minutes == 1 {
		return "1 min read"
	}
	return fmt.Sprintf("%d min read", minutes)
}

func formatTermCount(count int) string {
	if count == 1 {
		return "1 term"
	}
	return fmt.Sprintf("%d terms", count)
}
