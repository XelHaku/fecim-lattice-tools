// Package widgets provides shared widget utilities for Fyne GUI development.
// This file implements a reusable EducationalPanel widget for displaying
// context-sensitive explanations in demo applications.
package widgets

import (
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// EducationalPanel is a reusable widget for displaying educational content
// with a title and scrollable body. It's designed for "Live Slide" style demos
// where context-sensitive explanations are shown alongside visualizations.
type EducationalPanel struct {
	widget.BaseWidget

	mu       sync.RWMutex
	title    string
	content  string
	minSize  fyne.Size
	renderer *educationalPanelRenderer
}

// EducationalPanelConfig holds configuration for creating an EducationalPanel.
type EducationalPanelConfig struct {
	Title      string
	Content    string
	MinSize    fyne.Size
	TitleStyle fyne.TextStyle
}

// NewEducationalPanel creates a new educational panel widget.
func NewEducationalPanel(config EducationalPanelConfig) *EducationalPanel {
	if config.MinSize.Width <= 0 {
		config.MinSize.Width = 200
	}
	if config.MinSize.Height <= 0 {
		config.MinSize.Height = 200
	}
	if config.Title == "" {
		config.Title = "About"
	}
	if config.Content == "" {
		config.Content = "Information will appear here."
	}

	e := &EducationalPanel{
		title:   config.Title,
		content: config.Content,
		minSize: config.MinSize,
	}
	e.ExtendBaseWidget(e)
	return e
}

// SetContent updates the panel content.
func (e *EducationalPanel) SetContent(title, content string) {
	e.mu.Lock()
	e.title = title
	e.content = content
	e.mu.Unlock()
	fyne.Do(func() {
		e.Refresh()
	})
}

// GetContent returns the current title and content.
func (e *EducationalPanel) GetContent() (string, string) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.title, e.content
}

// SetTitle updates only the title.
func (e *EducationalPanel) SetTitle(title string) {
	e.mu.Lock()
	e.title = title
	e.mu.Unlock()
	fyne.Do(func() {
		e.Refresh()
	})
}

// AppendContent appends text to the current content.
func (e *EducationalPanel) AppendContent(text string) {
	e.mu.Lock()
	e.content += "\n" + text
	e.mu.Unlock()
	fyne.Do(func() {
		e.Refresh()
	})
}

// MinSize returns the minimum size for the widget.
func (e *EducationalPanel) MinSize() fyne.Size {
	return e.minSize
}

// CreateRenderer implements fyne.Widget.
func (e *EducationalPanel) CreateRenderer() fyne.WidgetRenderer {
	e.mu.RLock()
	title := e.title
	content := e.content
	e.mu.RUnlock()

	// Title label with emphasis
	titleLabel := widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	titleLabel.Importance = widget.HighImportance

	// Content label with wrapping
	contentLabel := widget.NewLabel(content)
	contentLabel.Wrapping = fyne.TextWrapWord

	// Scrollable content container
	contentScroll := container.NewScroll(contentLabel)
	contentScroll.SetMinSize(fyne.NewSize(e.minSize.Width-20, e.minSize.Height-60))

	// Layout with separator
	box := container.NewBorder(
		container.NewVBox(titleLabel, widget.NewSeparator()),
		nil, nil, nil,
		container.NewPadded(contentScroll),
	)

	r := &educationalPanelRenderer{
		panel:        e,
		titleLabel:   titleLabel,
		contentLabel: contentLabel,
		container:    box,
	}
	e.renderer = r
	return widget.NewSimpleRenderer(box)
}

type educationalPanelRenderer struct {
	panel        *EducationalPanel
	titleLabel   *widget.Label
	contentLabel *widget.Label
	container    *fyne.Container
}

// UpdateContent updates the renderer's labels with current panel content.
// Call this from Refresh() to sync displayed content with panel state.
func (e *EducationalPanel) updateRenderer() {
	if e.renderer == nil {
		return
	}
	e.mu.RLock()
	title := e.title
	content := e.content
	e.mu.RUnlock()

	e.renderer.titleLabel.SetText(title)
	e.renderer.contentLabel.SetText(content)
}

// Refresh updates the widget display.
func (e *EducationalPanel) Refresh() {
	e.updateRenderer()
	e.BaseWidget.Refresh()
}

// EducationalSection represents a collapsible section for educational content.
type EducationalSection struct {
	Title     string
	IconName  fyne.ThemeIconName
	Summary   string
	Details   string
	LearnMore string // URL for external reference
}

// CreateEducationalAccordion creates an accordion widget from educational sections.
func CreateEducationalAccordion(sections []EducationalSection) *widget.Accordion {
	items := make([]*widget.AccordionItem, len(sections))

	for i, section := range sections {
		// Create content for this section
		detailsLabel := widget.NewLabel(section.Details)
		detailsLabel.Wrapping = fyne.TextWrapWord

		var content fyne.CanvasObject
		if section.LearnMore != "" {
			learnBtn := widget.NewButtonWithIcon("Learn More", theme.DocumentIcon(), nil)
			learnBtn.Importance = widget.LowImportance
			// Note: URL handling should be done by the caller
			content = container.NewVBox(detailsLabel, container.NewHBox(learnBtn))
		} else {
			content = detailsLabel
		}

		items[i] = widget.NewAccordionItem(
			section.Title+" - "+section.Summary,
			content,
		)
	}

	accordion := widget.NewAccordion(items...)
	accordion.MultiOpen = true
	return accordion
}
