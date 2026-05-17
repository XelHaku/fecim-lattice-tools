//go:build legacy_fyne

package widgets

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/shared/presets"
)

// PresetBrowser provides a UI for browsing, loading, and saving presets
type PresetBrowser struct {
	widget.BaseWidget

	manager       *presets.Manager
	currentModule presets.Module
	window        fyne.Window

	// UI components
	categoryFilter *widget.Select
	moduleFilter   *widget.Select
	searchEntry    *widget.Entry
	presetList     *widget.List
	detailsPanel   *fyne.Container

	// Current state
	filteredPresets []*presets.Preset
	selectedIndex   int
	selectedPreset  *presets.Preset

	// Callbacks
	OnPresetApplied func(preset *presets.Preset)
	OnPresetSaved   func(preset *presets.Preset)
}

// NewPresetBrowser creates a new preset browser widget
func NewPresetBrowser(manager *presets.Manager, currentModule presets.Module, window fyne.Window) *PresetBrowser {
	pb := &PresetBrowser{
		manager:       manager,
		currentModule: currentModule,
		window:        window,
		selectedIndex: -1,
	}
	pb.ExtendBaseWidget(pb)
	return pb
}

// CreateRenderer implements fyne.Widget
func (pb *PresetBrowser) CreateRenderer() fyne.WidgetRenderer {
	pb.createUI()
	pb.refreshList()

	content := container.NewBorder(
		pb.createFilterBar(),
		pb.createActionBar(),
		nil,
		nil,
		container.NewHSplit(
			pb.createListPanel(),
			pb.createDetailsPanel(),
		),
	)

	return widget.NewSimpleRenderer(content)
}

// createUI initializes all UI components
func (pb *PresetBrowser) createUI() {
	// Category filter
	categories := []string{"All Categories"}
	for _, c := range pb.manager.GetCategories() {
		categories = append(categories, categoryDisplayName(c))
	}
	pb.categoryFilter = widget.NewSelect(categories, func(s string) {
		pb.refreshList()
	})
	pb.categoryFilter.SetSelected("All Categories")

	// Module filter
	modules := []string{"Current Module", "All Modules"}
	for _, m := range presets.AllModules() {
		if m != presets.ModuleGlobal {
			modules = append(modules, moduleDisplayName(m))
		}
	}
	pb.moduleFilter = widget.NewSelect(modules, func(s string) {
		pb.refreshList()
	})
	pb.moduleFilter.SetSelected("Current Module")

	// Search entry
	pb.searchEntry = widget.NewEntry()
	pb.searchEntry.SetPlaceHolder("Search presets...")
	pb.searchEntry.OnChanged = func(s string) {
		pb.refreshList()
	}

	// Preset list
	pb.presetList = widget.NewList(
		func() int {
			return len(pb.filteredPresets)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.DocumentIcon()),
				widget.NewLabel("Preset Name"),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if i >= len(pb.filteredPresets) {
				return
			}
			preset := pb.filteredPresets[i]
			box := o.(*fyne.Container)
			icon := box.Objects[0].(*widget.Icon)
			label := box.Objects[1].(*widget.Label)

			// Set icon based on category
			switch preset.Metadata.Category {
			case presets.CategoryEducational:
				icon.SetResource(theme.InfoIcon())
			case presets.CategoryResearch:
				icon.SetResource(theme.SearchIcon())
			case presets.CategoryDemo:
				icon.SetResource(theme.MediaPlayIcon())
			case presets.CategoryBenchmark:
				icon.SetResource(theme.ComputerIcon())
			default:
				icon.SetResource(theme.DocumentIcon())
			}

			// Add built-in indicator
			name := preset.Metadata.Name
			if preset.Metadata.BuiltIn {
				name = "★ " + name
			}
			label.SetText(name)
		},
	)
	pb.presetList.OnSelected = func(id widget.ListItemID) {
		if id >= 0 && id < len(pb.filteredPresets) {
			pb.selectedIndex = id
			pb.selectedPreset = pb.filteredPresets[id]
			pb.updateDetails()
		}
	}
}

// createFilterBar creates the filter controls bar
func (pb *PresetBrowser) createFilterBar() fyne.CanvasObject {
	return container.NewVBox(
		container.NewGridWithColumns(3,
			container.NewBorder(nil, nil, widget.NewLabel("Category:"), nil, pb.categoryFilter),
			container.NewBorder(nil, nil, widget.NewLabel("Module:"), nil, pb.moduleFilter),
			container.NewBorder(nil, nil, widget.NewIcon(theme.SearchIcon()), nil, pb.searchEntry),
		),
		widget.NewSeparator(),
	)
}

// createListPanel creates the preset list panel
func (pb *PresetBrowser) createListPanel() fyne.CanvasObject {
	return container.NewBorder(
		widget.NewLabel("Presets"),
		nil, nil, nil,
		pb.presetList,
	)
}

// createDetailsPanel creates the preset details panel
func (pb *PresetBrowser) createDetailsPanel() fyne.CanvasObject {
	pb.detailsPanel = container.NewVBox(
		widget.NewLabel("Select a preset to view details"),
	)
	return container.NewBorder(
		widget.NewLabel("Details"),
		nil, nil, nil,
		container.NewVScroll(pb.detailsPanel),
	)
}

// createActionBar creates the action buttons bar
func (pb *PresetBrowser) createActionBar() fyne.CanvasObject {
	applyBtn := widget.NewButtonWithIcon("Apply", theme.ConfirmIcon(), func() {
		if pb.selectedPreset != nil {
			pb.applyPreset(pb.selectedPreset)
		}
	})
	applyBtn.Importance = widget.HighImportance

	saveBtn := widget.NewButtonWithIcon("Save Current", theme.DocumentSaveIcon(), func() {
		pb.showSaveDialog()
	})

	importBtn := widget.NewButtonWithIcon("Import", theme.FolderOpenIcon(), func() {
		pb.showImportDialog()
	})

	exportBtn := widget.NewButtonWithIcon("Export", theme.DownloadIcon(), func() {
		if pb.selectedPreset != nil {
			pb.showExportDialog(pb.selectedPreset)
		}
	})

	deleteBtn := widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), func() {
		if pb.selectedPreset != nil && !pb.selectedPreset.Metadata.BuiltIn {
			pb.showDeleteDialog(pb.selectedPreset)
		}
	})

	return container.NewHBox(
		applyBtn,
		widget.NewSeparator(),
		saveBtn,
		widget.NewSeparator(),
		importBtn,
		exportBtn,
		widget.NewSeparator(),
		deleteBtn,
	)
}

// refreshList updates the preset list based on current filters
func (pb *PresetBrowser) refreshList() {
	var filters []presets.ListFilter

	// Module filter
	selectedModule := pb.moduleFilter.Selected
	if selectedModule == "Current Module" {
		filters = append(filters, presets.FilterByModule(pb.currentModule))
	} else if selectedModule != "All Modules" {
		for _, m := range presets.AllModules() {
			if moduleDisplayName(m) == selectedModule {
				filters = append(filters, presets.FilterByModule(m))
				break
			}
		}
	}

	// Category filter
	selectedCategory := pb.categoryFilter.Selected
	if selectedCategory != "All Categories" {
		for _, c := range []presets.Category{
			presets.CategoryEducational,
			presets.CategoryResearch,
			presets.CategoryDemo,
			presets.CategoryCustom,
			presets.CategoryBenchmark,
		} {
			if categoryDisplayName(c) == selectedCategory {
				filters = append(filters, presets.FilterByCategory(c))
				break
			}
		}
	}

	// Search filter
	if query := strings.TrimSpace(pb.searchEntry.Text); query != "" {
		filters = append(filters, presets.SearchByName(query))
	}

	pb.filteredPresets = pb.manager.List(filters...)
	if pb.presetList != nil {
		pb.presetList.Refresh()
	}

	// Clear selection if no longer valid
	if pb.selectedIndex >= len(pb.filteredPresets) {
		pb.selectedIndex = -1
		pb.selectedPreset = nil
		pb.updateDetails()
	}
}

// updateDetails updates the details panel for the selected preset
func (pb *PresetBrowser) updateDetails() {
	pb.detailsPanel.Objects = nil

	if pb.selectedPreset == nil {
		pb.detailsPanel.Add(widget.NewLabel("Select a preset to view details"))
		pb.detailsPanel.Refresh()
		return
	}

	p := pb.selectedPreset

	// Header
	nameLabel := widget.NewLabel(p.Metadata.Name)
	nameLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Badges
	var badges []fyne.CanvasObject
	if p.Metadata.BuiltIn {
		badges = append(badges, widget.NewLabel("★ Built-in"))
	}
	badges = append(badges, widget.NewLabel(categoryDisplayName(p.Metadata.Category)))
	badges = append(badges, widget.NewLabel(moduleDisplayName(p.Metadata.Module)))

	// Description
	descLabel := widget.NewLabel(p.Metadata.Description)
	descLabel.Wrapping = fyne.TextWrapWord

	// Tags
	var tagsText string
	if len(p.Metadata.Tags) > 0 {
		tagsText = "Tags: " + strings.Join(p.Metadata.Tags, ", ")
	}

	// Configuration
	configLabel := widget.NewLabel("Configuration:")
	configLabel.TextStyle = fyne.TextStyle{Bold: true}

	configItems := container.NewVBox()
	for key, value := range p.Config {
		configItems.Add(widget.NewLabel(fmt.Sprintf("  %s: %v", key, value)))
	}

	// Build details panel
	pb.detailsPanel.Add(nameLabel)
	pb.detailsPanel.Add(container.NewHBox(badges...))
	pb.detailsPanel.Add(widget.NewSeparator())
	pb.detailsPanel.Add(descLabel)
	if tagsText != "" {
		pb.detailsPanel.Add(widget.NewLabel(tagsText))
	}
	pb.detailsPanel.Add(widget.NewSeparator())
	pb.detailsPanel.Add(configLabel)
	pb.detailsPanel.Add(configItems)

	// Metadata
	pb.detailsPanel.Add(widget.NewSeparator())
	pb.detailsPanel.Add(widget.NewLabel(fmt.Sprintf("Version: %s", p.Metadata.Version)))
	pb.detailsPanel.Add(widget.NewLabel(fmt.Sprintf("Created: %s", p.Metadata.CreatedAt.Format("2006-01-02"))))
	if p.Metadata.Author != "" {
		pb.detailsPanel.Add(widget.NewLabel(fmt.Sprintf("Author: %s", p.Metadata.Author)))
	}

	pb.detailsPanel.Refresh()
}

// applyPreset applies the selected preset
func (pb *PresetBrowser) applyPreset(preset *presets.Preset) {
	if err := pb.manager.Apply(preset.Metadata.ID); err != nil {
		dialog.ShowError(err, pb.window)
		return
	}

	dialog.ShowInformation("Preset Applied",
		fmt.Sprintf("Successfully applied preset: %s", preset.Metadata.Name),
		pb.window)

	if pb.OnPresetApplied != nil {
		pb.OnPresetApplied(preset)
	}
}

// showSaveDialog shows a dialog to save the current configuration as a preset
func (pb *PresetBrowser) showSaveDialog() {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Preset name")

	descEntry := widget.NewMultiLineEntry()
	descEntry.SetPlaceHolder("Description")
	descEntry.SetMinRowsVisible(3)

	categorySelect := widget.NewSelect([]string{
		categoryDisplayName(presets.CategoryCustom),
		categoryDisplayName(presets.CategoryEducational),
		categoryDisplayName(presets.CategoryResearch),
		categoryDisplayName(presets.CategoryDemo),
		categoryDisplayName(presets.CategoryBenchmark),
	}, nil)
	categorySelect.SetSelected(categoryDisplayName(presets.CategoryCustom))

	tagsEntry := widget.NewEntry()
	tagsEntry.SetPlaceHolder("Tags (comma-separated)")

	form := widget.NewForm(
		widget.NewFormItem("Name", nameEntry),
		widget.NewFormItem("Description", descEntry),
		widget.NewFormItem("Category", categorySelect),
		widget.NewFormItem("Tags", tagsEntry),
	)

	dlg := dialog.NewCustomConfirm("Save Preset", "Save", "Cancel", form, func(save bool) {
		if !save {
			return
		}

		name := strings.TrimSpace(nameEntry.Text)
		if name == "" {
			dialog.ShowError(fmt.Errorf("preset name is required"), pb.window)
			return
		}

		// Get category
		var category presets.Category
		switch categorySelect.Selected {
		case categoryDisplayName(presets.CategoryEducational):
			category = presets.CategoryEducational
		case categoryDisplayName(presets.CategoryResearch):
			category = presets.CategoryResearch
		case categoryDisplayName(presets.CategoryDemo):
			category = presets.CategoryDemo
		case categoryDisplayName(presets.CategoryBenchmark):
			category = presets.CategoryBenchmark
		default:
			category = presets.CategoryCustom
		}

		preset, err := pb.manager.CreateFromCurrent(
			name,
			strings.TrimSpace(descEntry.Text),
			pb.currentModule,
			category,
		)
		if err != nil {
			dialog.ShowError(err, pb.window)
			return
		}

		// Add tags
		if tags := strings.TrimSpace(tagsEntry.Text); tags != "" {
			for _, t := range strings.Split(tags, ",") {
				t = strings.TrimSpace(t)
				if t != "" {
					preset.Metadata.Tags = append(preset.Metadata.Tags, t)
				}
			}
			pb.manager.Save(preset)
		}

		pb.refreshList()
		dialog.ShowInformation("Preset Saved",
			fmt.Sprintf("Successfully saved preset: %s", name),
			pb.window)

		if pb.OnPresetSaved != nil {
			pb.OnPresetSaved(preset)
		}
	}, pb.window)

	dlg.Resize(fyne.NewSize(400, 300))
	dlg.Show()
}

// showImportDialog shows a file dialog to import a preset
func (pb *PresetBrowser) showImportDialog() {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, pb.window)
			return
		}
		if reader == nil {
			return // Cancelled
		}
		defer reader.Close()

		preset, err := pb.manager.Import(reader.URI().Path())
		if err != nil {
			dialog.ShowError(err, pb.window)
			return
		}

		pb.refreshList()
		dialog.ShowInformation("Preset Imported",
			fmt.Sprintf("Successfully imported preset: %s", preset.Metadata.Name),
			pb.window)
	}, pb.window)
}

// showExportDialog shows a file dialog to export a preset
func (pb *PresetBrowser) showExportDialog(preset *presets.Preset) {
	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, pb.window)
			return
		}
		if writer == nil {
			return // Cancelled
		}
		defer writer.Close()

		if err := pb.manager.Export(preset.Metadata.ID, writer.URI().Path()); err != nil {
			dialog.ShowError(err, pb.window)
			return
		}

		dialog.ShowInformation("Preset Exported",
			fmt.Sprintf("Successfully exported preset to: %s", writer.URI().Path()),
			pb.window)
	}, pb.window)
}

// showDeleteDialog shows a confirmation dialog to delete a preset
func (pb *PresetBrowser) showDeleteDialog(preset *presets.Preset) {
	dialog.ShowConfirm("Delete Preset",
		fmt.Sprintf("Are you sure you want to delete the preset '%s'?\n\nThis action cannot be undone.", preset.Metadata.Name),
		func(delete bool) {
			if !delete {
				return
			}

			if err := pb.manager.Delete(preset.Metadata.ID); err != nil {
				dialog.ShowError(err, pb.window)
				return
			}

			pb.selectedIndex = -1
			pb.selectedPreset = nil
			pb.refreshList()
			pb.updateDetails()

			dialog.ShowInformation("Preset Deleted",
				fmt.Sprintf("Successfully deleted preset: %s", preset.Metadata.Name),
				pb.window)
		}, pb.window)
}

// SetModule changes the current module filter
func (pb *PresetBrowser) SetModule(module presets.Module) {
	pb.currentModule = module
	pb.refreshList()
}

// Helper functions

func categoryDisplayName(c presets.Category) string {
	switch c {
	case presets.CategoryEducational:
		return "📚 Educational"
	case presets.CategoryResearch:
		return "🔬 Research"
	case presets.CategoryDemo:
		return "🎬 Demo"
	case presets.CategoryCustom:
		return "📁 Custom"
	case presets.CategoryBenchmark:
		return "📊 Benchmark"
	default:
		return string(c)
	}
}

func moduleDisplayName(m presets.Module) string {
	switch m {
	case presets.ModuleGlobal:
		return "🌐 Global"
	case presets.ModuleHysteresis:
		return "📈 Hysteresis"
	case presets.ModuleCrossbar:
		return "⊞ Crossbar"
	case presets.ModuleMNIST:
		return "🔢 MNIST"
	case presets.ModuleCircuits:
		return "⚡ Circuits"
	case presets.ModuleComparison:
		return "⚖️ Comparison"
	case presets.ModuleEDA:
		return "🔧 EDA"
	default:
		return string(m)
	}
}
