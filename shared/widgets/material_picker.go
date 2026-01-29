// Package widgets provides reusable UI components.
package widgets

import (
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/config/physics"
)

// MaterialPicker provides a dialog for selecting ferroelectric materials.
// It displays all materials with their properties, descriptions, and references.
type MaterialPicker struct {
	widget.BaseWidget

	// Data
	materials   map[string]*physics.Material
	materialIDs []string // Ordered list
	selectedID  string

	// Callbacks
	OnSelected func(materialID string, material *physics.Material)

	// UI components
	searchEntry  *widget.Entry
	cardList     *fyne.Container
	detailPanel  *MaterialDetailPanel
	cards        map[string]*MaterialCard

	// Filter state
	filterQuery string
	filteredIDs []string
}

// NewMaterialPicker creates a new material picker widget.
func NewMaterialPicker(onSelected func(string, *physics.Material)) *MaterialPicker {
	mp := &MaterialPicker{
		OnSelected: onSelected,
		cards:      make(map[string]*MaterialCard),
	}

	// Load materials from config
	cfg, err := physics.Load()
	if err != nil {
		// Fallback: empty picker
		mp.materials = make(map[string]*physics.Material)
		mp.materialIDs = []string{}
	} else {
		mp.materials = cfg.Materials
		mp.materialIDs = mp.getSortedMaterialIDs()
	}
	mp.filteredIDs = mp.materialIDs

	mp.ExtendBaseWidget(mp)
	return mp
}

// getSortedMaterialIDs returns material IDs in a consistent order.
func (mp *MaterialPicker) getSortedMaterialIDs() []string {
	// Preferred order for common materials
	preferredOrder := []string{
		"default_hzo",
		"fecim_hzo",
		"fecim_hzo_target",
		"literature_superlattice",
		"cryogenic_hzo",
		"hzo_standard_32",
		"hzo_ftj_140",
		"alscn",
	}

	ordered := []string{}
	seen := make(map[string]bool)

	// Add materials in preferred order first
	for _, id := range preferredOrder {
		if _, exists := mp.materials[id]; exists {
			ordered = append(ordered, id)
			seen[id] = true
		}
	}

	// Add remaining materials alphabetically
	remaining := []string{}
	for id := range mp.materials {
		if !seen[id] {
			remaining = append(remaining, id)
		}
	}
	sort.Strings(remaining)
	ordered = append(ordered, remaining...)

	return ordered
}

// SetSelected sets the currently selected material.
func (mp *MaterialPicker) SetSelected(materialID string) {
	if _, exists := mp.materials[materialID]; !exists {
		return
	}
	mp.selectedID = materialID
	mp.Refresh()
}

// GetSelected returns the currently selected material ID and Material.
func (mp *MaterialPicker) GetSelected() (string, *physics.Material) {
	if mp.selectedID == "" {
		return "", nil
	}
	return mp.selectedID, mp.materials[mp.selectedID]
}

// updateFilter filters materials based on the search query.
func (mp *MaterialPicker) updateFilter() {
	query := strings.ToLower(strings.TrimSpace(mp.filterQuery))

	if query == "" {
		mp.filteredIDs = mp.materialIDs
		mp.rebuildCardList()
		return
	}

	mp.filteredIDs = []string{}
	for _, id := range mp.materialIDs {
		mat := mp.materials[id]

		// Search in ID, name, description, reference
		if strings.Contains(strings.ToLower(id), query) ||
			strings.Contains(strings.ToLower(mat.Name), query) ||
			strings.Contains(strings.ToLower(mat.Description), query) ||
			strings.Contains(strings.ToLower(mat.Reference), query) {
			mp.filteredIDs = append(mp.filteredIDs, id)
			continue
		}

		// Search for analog states number
		if mat.AnalogStates > 0 {
			statesStr := strings.ToLower(FormatDimensionless(float64(mat.AnalogStates)))
			if strings.Contains(statesStr, query) {
				mp.filteredIDs = append(mp.filteredIDs, id)
			}
		}
	}

	mp.rebuildCardList()
}

// rebuildCardList rebuilds the card list based on filtered IDs.
func (mp *MaterialPicker) rebuildCardList() {
	if mp.cardList == nil {
		return
	}

	mp.cardList.RemoveAll()

	for _, id := range mp.filteredIDs {
		mat := mp.materials[id]
		card := mp.getOrCreateCard(id, mat)
		card.SetSelected(id == mp.selectedID)
		mp.cardList.Add(card)
	}

	mp.cardList.Refresh()
}

// getOrCreateCard returns an existing card or creates a new one.
func (mp *MaterialPicker) getOrCreateCard(id string, mat *physics.Material) *MaterialCard {
	if card, exists := mp.cards[id]; exists {
		return card
	}

	card := NewMaterialCard(id, mat, func(selectedID string) {
		mp.onCardTapped(selectedID)
	})
	mp.cards[id] = card
	return card
}

// onCardTapped handles when a material card is tapped.
func (mp *MaterialPicker) onCardTapped(materialID string) {
	// Update selection
	oldSelected := mp.selectedID
	mp.selectedID = materialID

	// Update card visual states
	if oldCard, exists := mp.cards[oldSelected]; exists {
		oldCard.SetSelected(false)
	}
	if newCard, exists := mp.cards[materialID]; exists {
		newCard.SetSelected(true)
	}

	// Update detail panel
	if mp.detailPanel != nil && mp.materials[materialID] != nil {
		mp.detailPanel.SetMaterial(mp.materials[materialID])
	}
}

// CreateRenderer creates the widget renderer.
func (mp *MaterialPicker) CreateRenderer() fyne.WidgetRenderer {
	// Search entry
	mp.searchEntry = widget.NewEntry()
	mp.searchEntry.SetPlaceHolder("Search materials...")
	mp.searchEntry.OnChanged = func(s string) {
		mp.filterQuery = s
		mp.updateFilter()
	}

	// Card list container
	mp.cardList = container.NewVBox()
	for _, id := range mp.filteredIDs {
		mat := mp.materials[id]
		card := mp.getOrCreateCard(id, mat)
		card.SetSelected(id == mp.selectedID)
		mp.cardList.Add(card)
	}

	// Detail panel - show first material by default
	var firstMat *physics.Material
	if len(mp.materialIDs) > 0 {
		firstID := mp.materialIDs[0]
		firstMat = mp.materials[firstID]
		if mp.selectedID == "" {
			mp.selectedID = firstID
			if card, exists := mp.cards[firstID]; exists {
				card.SetSelected(true)
			}
		}
	}
	if mp.selectedID != "" {
		firstMat = mp.materials[mp.selectedID]
	}
	if firstMat == nil {
		// Fallback empty material
		firstMat = &physics.Material{Name: "No materials loaded", Description: "Check config/materials.yaml"}
	}
	mp.detailPanel = NewMaterialDetailPanel(firstMat)

	// Select button
	selectBtn := widget.NewButtonWithIcon("Select Material", theme.ConfirmIcon(), func() {
		if mp.OnSelected != nil && mp.selectedID != "" {
			mp.OnSelected(mp.selectedID, mp.materials[mp.selectedID])
		}
	})
	selectBtn.Importance = widget.HighImportance

	// Left pane: search + scrollable card list
	leftPane := container.NewBorder(
		mp.searchEntry,
		nil, nil, nil,
		container.NewVScroll(mp.cardList),
	)

	// Right pane: detail panel + select button
	rightPane := container.NewBorder(
		nil,
		container.NewPadded(selectBtn),
		nil, nil,
		mp.detailPanel,
	)

	// Main split layout
	split := container.NewHSplit(leftPane, rightPane)
	split.SetOffset(0.35)

	return widget.NewSimpleRenderer(split)
}

// MinSize returns the minimum size for the picker.
func (mp *MaterialPicker) MinSize() fyne.Size {
	return fyne.NewSize(850, 550)
}

// ShowMaterialPicker displays the material picker in a modal dialog.
func ShowMaterialPicker(parent fyne.Window, currentMaterialID string, onSelected func(string, *physics.Material)) {
	picker := NewMaterialPicker(nil)

	// Track if selection was made
	var selectedID string
	var selectedMat *physics.Material

	picker.OnSelected = func(id string, mat *physics.Material) {
		selectedID = id
		selectedMat = mat
	}

	// Pre-select current material
	if currentMaterialID != "" {
		picker.SetSelected(currentMaterialID)
	}

	// Create dialog
	d := dialog.NewCustomConfirm(
		"Select Ferroelectric Material",
		"Select",
		"Cancel",
		picker,
		func(confirmed bool) {
			if confirmed && onSelected != nil && selectedID != "" {
				onSelected(selectedID, selectedMat)
			} else if confirmed && onSelected != nil && picker.selectedID != "" {
				// Use picker's current selection if callback wasn't triggered
				onSelected(picker.selectedID, picker.materials[picker.selectedID])
			}
		},
		parent,
	)

	d.Resize(fyne.NewSize(900, 650))
	d.Show()
}

// CreateMaterialPickerButton creates a button that opens the material picker dialog.
func CreateMaterialPickerButton(parent fyne.Window, currentID string, onSelected func(string, *physics.Material)) *widget.Button {
	btn := widget.NewButtonWithIcon("Browse Materials...", theme.FolderOpenIcon(), func() {
		ShowMaterialPicker(parent, currentID, onSelected)
	})
	return btn
}

// GetMaterialByID returns a material by its ID from the global config.
// Returns nil if not found.
func GetMaterialByID(materialID string) *physics.Material {
	cfg, err := physics.Load()
	if err != nil {
		return nil
	}
	return cfg.GetMaterial(materialID)
}

// GetAllMaterialIDs returns all available material IDs.
func GetAllMaterialIDs() []string {
	cfg, err := physics.Load()
	if err != nil {
		return nil
	}
	return cfg.MaterialNames()
}
