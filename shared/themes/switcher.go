package themes

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// ThemeSwitcher is a widget that allows users to switch between themes
type ThemeSwitcher struct {
	widget.BaseWidget

	manager        *Manager
	selector       *widget.Select
	previewCard    *ThemePreviewCard
	removeListener func()
}

// NewThemeSwitcher creates a new theme switcher widget
func NewThemeSwitcher(manager *Manager) *ThemeSwitcher {
	ts := &ThemeSwitcher{
		manager: manager,
	}

	// Create theme selector dropdown
	themeNames := make([]string, len(AllThemes()))
	for i, t := range AllThemes() {
		themeNames[i] = ThemeDisplayName(t)
	}

	ts.selector = widget.NewSelect(themeNames, func(selected string) {
		for _, t := range AllThemes() {
			if ThemeDisplayName(t) == selected {
				manager.SetTheme(t)
				if ts.previewCard != nil {
					ts.previewCard.SetTheme(t)
				}
				break
			}
		}
	})

	// Set current selection
	ts.selector.SetSelected(ThemeDisplayName(manager.CurrentTheme()))

	// Create preview card
	ts.previewCard = NewThemePreviewCard(manager.CurrentTheme())

	// Listen for external theme changes
	ts.removeListener = manager.AddListener(func(t ThemeType) {
		ts.selector.SetSelected(ThemeDisplayName(t))
		if ts.previewCard != nil {
			ts.previewCard.SetTheme(t)
		}
	})

	ts.ExtendBaseWidget(ts)
	return ts
}

// CreateRenderer implements fyne.Widget
func (ts *ThemeSwitcher) CreateRenderer() fyne.WidgetRenderer {
	label := widget.NewLabelWithStyle("Theme", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	content := container.NewVBox(
		container.NewHBox(label, layout.NewSpacer(), ts.selector),
		widget.NewSeparator(),
		ts.previewCard,
	)

	return widget.NewSimpleRenderer(content)
}

// Destroy cleans up resources
func (ts *ThemeSwitcher) Destroy() {
	if ts.removeListener != nil {
		ts.removeListener()
	}
}

// ============================================================================
// Compact Theme Switcher (for toolbars/menus)
// ============================================================================

// CompactThemeSwitcher is a minimal theme switcher for toolbars
type CompactThemeSwitcher struct {
	widget.BaseWidget
	manager        *Manager
	button         *widget.Button
	removeListener func()
}

// NewCompactThemeSwitcher creates a compact theme toggle button
func NewCompactThemeSwitcher(manager *Manager) *CompactThemeSwitcher {
	cts := &CompactThemeSwitcher{
		manager: manager,
	}

	cts.button = widget.NewButton(cts.getIcon(), func() {
		manager.CycleTheme()
	})
	cts.button.Importance = widget.LowImportance

	// Listen for theme changes
	cts.removeListener = manager.AddListener(func(t ThemeType) {
		cts.button.SetText(cts.getIcon())
	})

	cts.ExtendBaseWidget(cts)
	return cts
}

func (cts *CompactThemeSwitcher) getIcon() string {
	switch cts.manager.CurrentTheme() {
	case ThemeLight:
		return "☀"
	case ThemeHighContrast:
		return "◐"
	default:
		return "☾"
	}
}

// CreateRenderer implements fyne.Widget
func (cts *CompactThemeSwitcher) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(cts.button)
}

// Destroy cleans up resources
func (cts *CompactThemeSwitcher) Destroy() {
	if cts.removeListener != nil {
		cts.removeListener()
	}
}

// ============================================================================
// Theme Preview Card
// ============================================================================

// ThemePreviewCard shows a preview of a theme's colors
type ThemePreviewCard struct {
	widget.BaseWidget
	themeType ThemeType
}

// NewThemePreviewCard creates a new theme preview card
func NewThemePreviewCard(t ThemeType) *ThemePreviewCard {
	tpc := &ThemePreviewCard{themeType: t}
	tpc.ExtendBaseWidget(tpc)
	return tpc
}

// SetTheme updates the preview to show a different theme
func (tpc *ThemePreviewCard) SetTheme(t ThemeType) {
	tpc.themeType = t
	tpc.Refresh()
}

// MinSize returns the minimum size for the preview card
func (tpc *ThemePreviewCard) MinSize() fyne.Size {
	return fyne.NewSize(280, 80)
}

// CreateRenderer implements fyne.Widget
func (tpc *ThemePreviewCard) CreateRenderer() fyne.WidgetRenderer {
	return &themePreviewRenderer{card: tpc}
}

type themePreviewRenderer struct {
	card    *ThemePreviewCard
	objects []fyne.CanvasObject
}

func (r *themePreviewRenderer) MinSize() fyne.Size {
	return r.card.MinSize()
}

func (r *themePreviewRenderer) Layout(size fyne.Size) {
	r.buildPreview(size)
}

func (r *themePreviewRenderer) buildPreview(size fyne.Size) {
	r.objects = nil

	theme := GetTheme(r.card.themeType)

	// Background
	bg := canvas.NewRectangle(theme.Color("background", 0))
	bg.Resize(size)
	bg.CornerRadius = 4
	r.objects = append(r.objects, bg)

	// Border
	border := canvas.NewRectangle(color.Transparent)
	border.Resize(size)
	border.CornerRadius = 4
	border.StrokeWidth = 1
	border.StrokeColor = theme.Color("separator", 0)
	r.objects = append(r.objects, border)

	// Color swatches
	swatchSize := float32(20)
	swatchY := size.Height/2 - swatchSize/2
	startX := float32(10)
	spacing := float32(6)

	colors := []fyne.ThemeColorName{
		"primary",
		"success",
		"warning",
		"error",
	}

	for i, colorName := range colors {
		swatch := canvas.NewRectangle(theme.Color(colorName, 0))
		swatch.Resize(fyne.NewSize(swatchSize, swatchSize))
		swatch.Move(fyne.NewPos(startX+float32(i)*(swatchSize+spacing), swatchY))
		swatch.CornerRadius = 3
		r.objects = append(r.objects, swatch)
	}

	// Sample text
	textX := startX + 4*(swatchSize+spacing) + 10
	sampleText := canvas.NewText("Sample Text", theme.Color("foreground", 0))
	sampleText.TextSize = 14
	sampleText.Move(fyne.NewPos(textX, size.Height/2-10))
	r.objects = append(r.objects, sampleText)

	// Surface preview
	surfaceWidth := float32(60)
	surfaceHeight := float32(30)
	surfaceX := size.Width - surfaceWidth - 10
	surfaceY := size.Height/2 - surfaceHeight/2
	surface := canvas.NewRectangle(theme.Color("menuBackground", 0))
	surface.Resize(fyne.NewSize(surfaceWidth, surfaceHeight))
	surface.Move(fyne.NewPos(surfaceX, surfaceY))
	surface.CornerRadius = 4
	r.objects = append(r.objects, surface)

	// Surface label
	surfaceLabel := canvas.NewText("Card", theme.Color("foreground", 0))
	surfaceLabel.TextSize = 14
	surfaceLabel.Move(fyne.NewPos(surfaceX+surfaceWidth/2-15, surfaceY+surfaceHeight/2-6))
	r.objects = append(r.objects, surfaceLabel)
}

func (r *themePreviewRenderer) Refresh() {
	r.buildPreview(r.card.Size())
	canvas.Refresh(r.card)
}

func (r *themePreviewRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *themePreviewRenderer) Destroy() {}

// ============================================================================
// Settings Panel Integration
// ============================================================================

// CreateSettingsSection creates a complete settings section for theme selection.
// This can be embedded in a settings dialog or preferences panel.
func CreateSettingsSection(manager *Manager) fyne.CanvasObject {
	switcher := NewThemeSwitcher(manager)

	description := widget.NewLabel("Choose how FeCIM Lattice Tools appears. Dark is optimized for low-light environments, Light for bright conditions, and High Contrast for maximum readability.")
	description.Wrapping = fyne.TextWrapWord

	return container.NewVBox(
		widget.NewLabelWithStyle("Appearance", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		description,
		container.NewPadded(switcher),
	)
}

// CreateQuickToggle creates a button that cycles through themes.
// Useful for adding to toolbars or menus.
func CreateQuickToggle(manager *Manager) *widget.Button {
	getLabel := func() string {
		switch manager.CurrentTheme() {
		case ThemeLight:
			return "☀ Light"
		case ThemeHighContrast:
			return "◐ High Contrast"
		default:
			return "☾ Dark"
		}
	}

	var btn *widget.Button
	btn = widget.NewButton(getLabel(), func() {
		manager.CycleTheme()
		btn.SetText(getLabel())
	})
	btn.Importance = widget.LowImportance

	// Update on external changes
	manager.AddListener(func(t ThemeType) {
		btn.SetText(getLabel())
	})

	return btn
}
