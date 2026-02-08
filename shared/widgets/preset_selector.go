package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/shared/presets"
)

// PresetSelector is a compact preset selector widget for embedding in module controls
type PresetSelector struct {
	widget.BaseWidget

	manager       *presets.Manager
	currentModule presets.Module
	window        fyne.Window

	// UI components
	selectWidget *widget.Select
	saveBtn      *widget.Button
	browseBtn    *widget.Button

	// Callbacks
	OnPresetApplied func(preset *presets.Preset)
	OnPresetSaved   func(preset *presets.Preset)
}

// NewPresetSelector creates a new compact preset selector
func NewPresetSelector(manager *presets.Manager, module presets.Module, window fyne.Window) *PresetSelector {
	ps := &PresetSelector{
		manager:       manager,
		currentModule: module,
		window:        window,
	}
	ps.ExtendBaseWidget(ps)
	return ps
}

// CreateRenderer implements fyne.Widget
func (ps *PresetSelector) CreateRenderer() fyne.WidgetRenderer {
	ps.createUI()

	content := container.NewBorder(
		nil, nil,
		widget.NewLabel("Preset:"),
		container.NewHBox(ps.saveBtn, ps.browseBtn),
		ps.selectWidget,
	)

	return widget.NewSimpleRenderer(content)
}

// createUI initializes all UI components
func (ps *PresetSelector) createUI() {
	// Get presets for this module
	modulePresets := ps.manager.List(presets.FilterByModule(ps.currentModule))

	// Build options list
	options := make([]string, 0, len(modulePresets)+1)
	options = append(options, "-- Select Preset --")
	for _, p := range modulePresets {
		name := p.Metadata.Name
		if p.Metadata.BuiltIn {
			name = "★ " + name
		}
		options = append(options, name)
	}

	// Create select widget
	ps.selectWidget = widget.NewSelect(options, func(selected string) {
		if selected == "-- Select Preset --" {
			return
		}

		// Find and apply the preset
		for _, p := range modulePresets {
			name := p.Metadata.Name
			if p.Metadata.BuiltIn {
				name = "★ " + name
			}
			if name == selected {
				ps.applyPreset(p)
				break
			}
		}
	})
	ps.selectWidget.SetSelected("-- Select Preset --")

	// Save button
	ps.saveBtn = widget.NewButtonWithIcon("", theme.DocumentSaveIcon(), func() {
		ps.showSaveDialog()
	})

	// Browse button
	ps.browseBtn = widget.NewButtonWithIcon("", theme.FolderOpenIcon(), func() {
		ps.showBrowser()
	})
}

// Refresh updates the preset list
func (ps *PresetSelector) Refresh() {
	if ps.selectWidget == nil {
		return
	}

	modulePresets := ps.manager.List(presets.FilterByModule(ps.currentModule))

	options := make([]string, 0, len(modulePresets)+1)
	options = append(options, "-- Select Preset --")
	for _, p := range modulePresets {
		name := p.Metadata.Name
		if p.Metadata.BuiltIn {
			name = "★ " + name
		}
		options = append(options, name)
	}

	ps.selectWidget.Options = options
	ps.selectWidget.Refresh()
}

// applyPreset applies a preset
func (ps *PresetSelector) applyPreset(preset *presets.Preset) {
	if err := ps.manager.Apply(preset.Metadata.ID); err != nil {
		dialog.ShowError(err, ps.window)
		return
	}

	if ps.OnPresetApplied != nil {
		ps.OnPresetApplied(preset)
	}
}

// showSaveDialog shows a dialog to save current config as preset
func (ps *PresetSelector) showSaveDialog() {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Preset name")

	descEntry := widget.NewEntry()
	descEntry.SetPlaceHolder("Description (optional)")

	form := widget.NewForm(
		widget.NewFormItem("Name", nameEntry),
		widget.NewFormItem("Description", descEntry),
	)

	dialog.NewCustomConfirm("Save Preset", "Save", "Cancel", form, func(save bool) {
		if !save || nameEntry.Text == "" {
			return
		}

		preset, err := ps.manager.CreateFromCurrent(
			nameEntry.Text,
			descEntry.Text,
			ps.currentModule,
			presets.CategoryCustom,
		)
		if err != nil {
			dialog.ShowError(err, ps.window)
			return
		}

		ps.Refresh()

		if ps.OnPresetSaved != nil {
			ps.OnPresetSaved(preset)
		}
	}, ps.window).Show()
}

// showBrowser shows the full preset browser dialog
func (ps *PresetSelector) showBrowser() {
	browser := NewPresetBrowser(ps.manager, ps.currentModule, ps.window)
	browser.OnPresetApplied = func(p *presets.Preset) {
		ps.Refresh()
		if ps.OnPresetApplied != nil {
			ps.OnPresetApplied(p)
		}
	}
	browser.OnPresetSaved = func(p *presets.Preset) {
		ps.Refresh()
		if ps.OnPresetSaved != nil {
			ps.OnPresetSaved(p)
		}
	}

	content := container.NewMax(browser)

	dlg := dialog.NewCustom("Preset Browser", "Close", content, ps.window)
	dlg.Resize(fyne.NewSize(800, 600))
	dlg.Show()
}

// SetModule changes the current module
func (ps *PresetSelector) SetModule(module presets.Module) {
	ps.currentModule = module
	ps.Refresh()
}

// PresetButton is a simple button that opens the preset browser
type PresetButton struct {
	widget.BaseWidget

	manager *presets.Manager
	module  presets.Module
	window  fyne.Window

	btn *widget.Button

	OnPresetApplied func(preset *presets.Preset)
}

// NewPresetButton creates a preset browser button
func NewPresetButton(manager *presets.Manager, module presets.Module, window fyne.Window) *PresetButton {
	pb := &PresetButton{
		manager: manager,
		module:  module,
		window:  window,
	}
	pb.ExtendBaseWidget(pb)
	return pb
}

// CreateRenderer implements fyne.Widget
func (pb *PresetButton) CreateRenderer() fyne.WidgetRenderer {
	pb.btn = widget.NewButtonWithIcon("Presets", theme.FolderIcon(), func() {
		pb.showBrowser()
	})
	return widget.NewSimpleRenderer(pb.btn)
}

// showBrowser shows the full preset browser dialog
func (pb *PresetButton) showBrowser() {
	browser := NewPresetBrowser(pb.manager, pb.module, pb.window)
	browser.OnPresetApplied = func(p *presets.Preset) {
		if pb.OnPresetApplied != nil {
			pb.OnPresetApplied(p)
		}
	}

	content := container.NewMax(browser)

	dlg := dialog.NewCustom("Preset Browser", "Close", content, pb.window)
	dlg.Resize(fyne.NewSize(800, 600))
	dlg.Show()
}

// PresetQuickSelect is a dropdown that shows only the most relevant presets
type PresetQuickSelect struct {
	widget.BaseWidget

	manager  *presets.Manager
	module   presets.Module
	category presets.Category

	selectWidget *widget.Select
	presetMap    map[string]*presets.Preset

	OnSelected func(preset *presets.Preset)
}

// NewPresetQuickSelect creates a quick select dropdown for a specific category
func NewPresetQuickSelect(manager *presets.Manager, module presets.Module, category presets.Category) *PresetQuickSelect {
	pqs := &PresetQuickSelect{
		manager:   manager,
		module:    module,
		category:  category,
		presetMap: make(map[string]*presets.Preset),
	}
	pqs.ExtendBaseWidget(pqs)
	return pqs
}

// CreateRenderer implements fyne.Widget
func (pqs *PresetQuickSelect) CreateRenderer() fyne.WidgetRenderer {
	// Get relevant presets
	filters := []presets.ListFilter{
		presets.FilterByModule(pqs.module),
	}
	if pqs.category != "" {
		filters = append(filters, presets.FilterByCategory(pqs.category))
	}

	modulePresets := pqs.manager.List(filters...)

	options := make([]string, 0, len(modulePresets)+1)
	options = append(options, "Quick Presets...")
	for _, p := range modulePresets {
		displayName := p.Metadata.Name
		if p.Metadata.BuiltIn {
			displayName = "★ " + displayName
		}
		options = append(options, displayName)
		pqs.presetMap[displayName] = p
	}

	pqs.selectWidget = widget.NewSelect(options, func(selected string) {
		if selected == "Quick Presets..." {
			return
		}

		if preset, ok := pqs.presetMap[selected]; ok {
			if pqs.OnSelected != nil {
				pqs.OnSelected(preset)
			}
		}

		// Reset to placeholder
		pqs.selectWidget.SetSelected("Quick Presets...")
	})
	pqs.selectWidget.SetSelected("Quick Presets...")

	return widget.NewSimpleRenderer(pqs.selectWidget)
}

// Refresh updates the preset list
func (pqs *PresetQuickSelect) Refresh() {
	if pqs.selectWidget == nil {
		return
	}

	filters := []presets.ListFilter{
		presets.FilterByModule(pqs.module),
	}
	if pqs.category != "" {
		filters = append(filters, presets.FilterByCategory(pqs.category))
	}

	modulePresets := pqs.manager.List(filters...)

	options := make([]string, 0, len(modulePresets)+1)
	options = append(options, "Quick Presets...")
	pqs.presetMap = make(map[string]*presets.Preset)

	for _, p := range modulePresets {
		displayName := p.Metadata.Name
		if p.Metadata.BuiltIn {
			displayName = "★ " + displayName
		}
		options = append(options, displayName)
		pqs.presetMap[displayName] = p
	}

	pqs.selectWidget.Options = options
	pqs.selectWidget.Refresh()
}
