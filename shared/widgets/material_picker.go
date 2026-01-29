// Package widgets provides reusable UI components.
package widgets

import (
	"fmt"
	"image/color"
	"math"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/config/physics"
)

// MaterialPicker provides a dialog for selecting ferroelectric materials.
// Displays all materials in a single table with key parameters as columns.
type MaterialPicker struct {
	widget.BaseWidget

	// Data
	materials   map[string]*physics.Material
	materialIDs []string // Ordered list
	selectedID  string
	selectedRow int

	// Callbacks
	OnSelected func(materialID string, material *physics.Material)

	// UI components
	searchEntry *widget.Entry
	table       *widget.Table
	infoLabel   *widget.Label

	// Filter state
	filterQuery string
	filteredIDs []string
}

// Column definitions for the material table
var materialColumns = []struct {
	Name        string
	Width       float32
	Description string
	Models      string // [P] for Preisach
}{
	{"Name", 160, "Material name", ""},
	{"States", 70, "Number of analog states (bits/cell)", ""},
	{"Pr", 85, "Remanent polarization (µC/cm²)", "[P]"},
	{"Ps", 85, "Saturation polarization (µC/cm²)", "[P]"},
	{"Ec", 85, "Coercive field (MV/cm)", "[P]"},
	{"τ", 70, "Switching time", "[P]"},
	{"Tc", 80, "Curie temperature", "[P]"},
	{"Endurance", 90, "Write cycle endurance", ""},
	{"Thickness", 70, "Film thickness (nm)", "[P]"},
	{"Reference", 200, "Data source / citation", ""},
}

// NewMaterialPicker creates a new material picker widget.
func NewMaterialPicker(onSelected func(string, *physics.Material)) *MaterialPicker {
	mp := &MaterialPicker{
		OnSelected:  onSelected,
		selectedRow: -1,
	}

	// Load materials from config
	cfg, err := physics.Load()
	if err != nil {
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

	for _, id := range preferredOrder {
		if _, exists := mp.materials[id]; exists {
			ordered = append(ordered, id)
			seen[id] = true
		}
	}

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
	// Find row index
	for i, id := range mp.filteredIDs {
		if id == materialID {
			mp.selectedRow = i
			break
		}
	}
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
	} else {
		mp.filteredIDs = []string{}
		for _, id := range mp.materialIDs {
			mat := mp.materials[id]

			if strings.Contains(strings.ToLower(id), query) ||
				strings.Contains(strings.ToLower(mat.Name), query) ||
				strings.Contains(strings.ToLower(mat.Description), query) ||
				strings.Contains(strings.ToLower(mat.Reference), query) {
				mp.filteredIDs = append(mp.filteredIDs, id)
				continue
			}

			if mat.AnalogStates > 0 {
				statesStr := strings.ToLower(FormatDimensionless(float64(mat.AnalogStates)))
				if strings.Contains(statesStr, query) {
					mp.filteredIDs = append(mp.filteredIDs, id)
				}
			}
		}
	}

	// Update selected row index
	mp.selectedRow = -1
	for i, id := range mp.filteredIDs {
		if id == mp.selectedID {
			mp.selectedRow = i
			break
		}
	}

	if mp.table != nil {
		mp.table.Refresh()
	}
}

// getCellValue returns the formatted value for a cell.
func (mp *MaterialPicker) getCellValue(row, col int) string {
	if row < 0 || row >= len(mp.filteredIDs) {
		return ""
	}

	mat := mp.materials[mp.filteredIDs[row]]
	if mat == nil {
		return ""
	}

	switch col {
	case 0: // Name
		return mat.Name
	case 1: // States
		if mat.AnalogStates > 0 {
			bits := math.Log2(float64(mat.AnalogStates))
			return fmt.Sprintf("%d (%.1fb)", mat.AnalogStates, bits)
		}
		return "—"
	case 2: // Pr
		return FormatPolarization(mat.PrCM2)
	case 3: // Ps
		return FormatPolarization(mat.PsCM2)
	case 4: // Ec
		return FormatField(mat.EcVM)
	case 5: // τ (Tau)
		return FormatTime(mat.TauS)
	case 6: // Tc (Curie temp)
		if mat.CurieTempK > 0 {
			return fmt.Sprintf("%.0f K", mat.CurieTempK)
		}
		return "—"
	case 7: // Endurance
		return FormatEndurance(mat.EnduranceCycles)
	case 8: // Thickness
		return FormatThickness(mat.ThicknessM)
	case 9: // Reference
		if mat.Reference != "" {
			return TruncateString(mat.Reference, 35)
		}
		return "—"
	default:
		return ""
	}
}

// getHeaderText returns the header text for a column.
func (mp *MaterialPicker) getHeaderText(col int) string {
	if col < 0 || col >= len(materialColumns) {
		return ""
	}
	c := materialColumns[col]
	if c.Models != "" {
		return fmt.Sprintf("%s %s", c.Models, c.Name)
	}
	return c.Name
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

	// Info label for tooltips
	mp.infoLabel = widget.NewLabel("Click a row to select. [P] = Used in Preisach model.")
	mp.infoLabel.Wrapping = fyne.TextWrapWord
	mp.infoLabel.TextStyle = fyne.TextStyle{Italic: true}

	// Colors for row highlighting
	selectedBgColor := color.RGBA{40, 80, 120, 255}
	normalBgColor := color.RGBA{0, 0, 0, 0} // transparent
	headerBgColor := color.RGBA{50, 55, 65, 255}

	// Create table
	mp.table = widget.NewTable(
		// Size: rows = materials + 1 header, cols = parameters
		func() (int, int) {
			return len(mp.filteredIDs) + 1, len(materialColumns)
		},
		// Create cell with background rectangle for row highlighting
		func() fyne.CanvasObject {
			bg := canvas.NewRectangle(normalBgColor)
			label := widget.NewLabel("Template Text")
			label.Wrapping = fyne.TextWrapOff
			return container.NewStack(bg, label)
		},
		// Update cell
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			stack := cell.(*fyne.Container)
			bg := stack.Objects[0].(*canvas.Rectangle)
			label := stack.Objects[1].(*widget.Label)

			if id.Row == 0 {
				// Header row
				label.SetText(mp.getHeaderText(id.Col))
				label.TextStyle = fyne.TextStyle{Bold: true}
				bg.FillColor = headerBgColor
			} else {
				// Data row (adjust for header)
				dataRow := id.Row - 1
				label.SetText(mp.getCellValue(dataRow, id.Col))

				// Highlight entire selected row
				if dataRow == mp.selectedRow {
					label.TextStyle = fyne.TextStyle{Bold: true}
					bg.FillColor = selectedBgColor
				} else {
					label.TextStyle = fyne.TextStyle{}
					bg.FillColor = normalBgColor
				}
			}
			bg.Refresh()
		},
	)

	// Set column widths
	for i, col := range materialColumns {
		mp.table.SetColumnWidth(i, col.Width)
	}

	// Handle row selection
	mp.table.OnSelected = func(id widget.TableCellID) {
		if id.Row == 0 {
			// Header clicked - show column description
			if id.Col >= 0 && id.Col < len(materialColumns) {
				col := materialColumns[id.Col]
				desc := col.Description
				if col.Models != "" {
					desc = col.Models + " " + desc
				}
				mp.infoLabel.SetText(desc)
			}
			mp.table.UnselectAll()
			return
		}

		dataRow := id.Row - 1
		if dataRow >= 0 && dataRow < len(mp.filteredIDs) {
			mp.selectedRow = dataRow
			mp.selectedID = mp.filteredIDs[dataRow]
			mat := mp.materials[mp.selectedID]

			// Update info with material description
			info := mat.Name
			if mat.Description != "" {
				info += ": " + mat.Description
			}
			if mat.Reference != "" {
				info += " [" + TruncateString(mat.Reference, 60) + "]"
			}
			mp.infoLabel.SetText(info)

			mp.table.Refresh()
		}
		mp.table.UnselectAll()
	}

	// Select first material by default
	if len(mp.filteredIDs) > 0 && mp.selectedID == "" {
		mp.selectedID = mp.filteredIDs[0]
		mp.selectedRow = 0
	}

	// Layout: search at top, table in middle, info at bottom
	content := container.NewBorder(
		mp.searchEntry,
		container.NewPadded(mp.infoLabel),
		nil, nil,
		mp.table,
	)

	return widget.NewSimpleRenderer(content)
}

// MinSize returns the minimum size for the picker.
func (mp *MaterialPicker) MinSize() fyne.Size {
	return fyne.NewSize(1100, 450)
}

// ShowMaterialPicker displays the material picker in a modal dialog.
func ShowMaterialPicker(parent fyne.Window, currentMaterialID string, onSelected func(string, *physics.Material)) {
	picker := NewMaterialPicker(nil)

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
			if confirmed && onSelected != nil && picker.selectedID != "" {
				onSelected(picker.selectedID, picker.materials[picker.selectedID])
			}
		},
		parent,
	)

	d.Resize(fyne.NewSize(1150, 500))
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
