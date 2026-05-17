//go:build legacy_fyne

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
	Models      string // [P] for Preisach, [LK] for Landau-Khalatnikov, [P,LK] for both
}{
	{"Name", 200, "Material name", ""},
	{"Eng", 65, "Engine support: [P], [LK], [P,LK]", ""},
	{"States", 82, "Number of analog states (bits/cell)", ""},
	{"Pr", 80, "Remanent polarization (µC/cm²)", "[P,LK]"},
	{"Ps", 80, "Saturation polarization (µC/cm²)", "[P,LK]"},
	{"Ec", 80, "Coercive field (MV/cm)", "[P,LK]"},
	{"τ", 70, "Switching time", "[P]"},
	{"εHF", 65, "High-frequency relative permittivity", "[P,LK]"},
	{"β", 75, "Landau β coefficient", "[LK]"},
	{"γ", 75, "Landau γ coefficient", "[LK]"},
	{"ρ", 75, "L-K viscosity coefficient", "[LK]"},
	{"Tc", 70, "Curie temperature", "[P,LK]"},
	{"Endurance", 90, "Write cycle endurance", ""},
	{"Thickness", 78, "Film thickness (nm)", "[P,LK]"},
	{"Reference", 250, "Data source / citation", ""},
}

func isLKCompatible(mat *physics.Material) bool {
	if mat == nil {
		return false
	}
	thermo := mat.Thermodynamics
	return thermo.BetaLandau != 0 && thermo.GammaLandau != 0 && thermo.RhoViscosity != 0
}

func isPreisachCompatible(mat *physics.Material) bool {
	if mat == nil {
		return false
	}
	return mat.PrCM2 > 0 && mat.PsCM2 > 0 && mat.EcVM > 0
}

func engineSupportTag(mat *physics.Material) string {
	hasP := isPreisachCompatible(mat)
	hasLK := isLKCompatible(mat)
	switch {
	case hasP && hasLK:
		return "[P,LK]"
	case hasP:
		return "[P]"
	case hasLK:
		return "[LK]"
	default:
		return "—"
	}
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
	case 1: // Engine support
		return engineSupportTag(mat)
	case 2: // States
		if mat.AnalogStates > 0 {
			bits := math.Log2(float64(mat.AnalogStates))
			return fmt.Sprintf("%d (%.1fb)", mat.AnalogStates, bits)
		}
		return "—"
	case 3: // Pr
		return FormatPolarization(mat.PrCM2)
	case 4: // Ps
		return FormatPolarization(mat.PsCM2)
	case 5: // Ec
		return FormatField(mat.EcVM)
	case 6: // τ (Tau)
		return FormatTime(mat.TauS)
	case 7: // εHF
		if mat.EpsilonHF > 0 {
			return fmt.Sprintf("%.0f", mat.EpsilonHF)
		}
		return "—"
	case 8: // β
		if mat.Thermodynamics.BetaLandau != 0 {
			return fmt.Sprintf("%.2g", mat.Thermodynamics.BetaLandau)
		}
		return "—"
	case 9: // γ
		if mat.Thermodynamics.GammaLandau != 0 {
			return fmt.Sprintf("%.2g", mat.Thermodynamics.GammaLandau)
		}
		return "—"
	case 10: // ρ
		if mat.Thermodynamics.RhoViscosity != 0 {
			return fmt.Sprintf("%.2g", mat.Thermodynamics.RhoViscosity)
		}
		return "—"
	case 11: // Tc (Curie temp)
		if mat.CurieTempK > 0 {
			return fmt.Sprintf("%.0f K", mat.CurieTempK)
		}
		return "—"
	case 12: // Endurance
		return FormatEndurance(mat.EnduranceCycles)
	case 13: // Thickness
		return FormatThickness(mat.ThicknessM)
	case 14: // Reference
		if mat.Reference != "" {
			return TruncateString(mat.Reference, 45)
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
	mp.infoLabel = widget.NewLabel("Click a row to select. Eng tags: [P] Preisach, [LK] Landau-Khalatnikov, [P,LK] supports both.")
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
	// Prefer a wide layout (table + details) on desktop, but keep it usable on
	// smaller windows (e.g., 1024x768 and below).
	return fyne.NewSize(1200, 520)
}

// ShowMaterialPicker displays the material picker in a modal dialog.
func ShowMaterialPicker(parent fyne.Window, currentMaterialID string, onSelected func(string, *physics.Material)) {
	picker := NewMaterialPicker(nil)

	// Pre-select current material.
	if currentMaterialID != "" {
		picker.SetSelected(currentMaterialID)
	}

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

	// Responsive sizing: never exceed the window; keep a sensible minimum.
	// The table has 15 columns totaling ~1360px, so target ~1440px to fit all.
	canvasSize := parent.Canvas().Size()
	w := float32(1440)
	h := float32(640)
	if canvasSize.Width > 0 {
		maxW := canvasSize.Width * 0.95
		if w > maxW {
			w = maxW
		}
		if w < 900 {
			w = 900
		}
	}
	if canvasSize.Height > 0 {
		maxH := canvasSize.Height * 0.90
		if h > maxH {
			h = maxH
		}
		if h < 480 {
			h = 480
		}
	}
	if w <= 0 {
		w = 1200
	}
	if h <= 0 {
		h = 600
	}

	d.Resize(fyne.NewSize(w, h))
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
