//go:build legacy_fyne

// Package widgets provides reusable UI components.
package widgets

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/config/physics"
)

// MaterialTable displays all properties of a single material in a categorized tabbed view.
type MaterialTable struct {
	widget.BaseWidget

	material   *physics.Material
	properties []FormattedProperty
	tabs       *container.AppTabs
}

// NewMaterialTable creates a new material table widget.
func NewMaterialTable(material *physics.Material) *MaterialTable {
	mt := &MaterialTable{
		material:   material,
		properties: GetMaterialProperties(material),
	}
	mt.ExtendBaseWidget(mt)
	return mt
}

// SetMaterial updates the displayed material.
func (mt *MaterialTable) SetMaterial(material *physics.Material) {
	mt.material = material
	mt.properties = GetMaterialProperties(material)
	mt.Refresh()
}

// CreateRenderer creates the widget renderer.
func (mt *MaterialTable) CreateRenderer() fyne.WidgetRenderer {
	mt.tabs = container.NewAppTabs()

	// Build tabs for each category that has properties
	for _, category := range CategoryOrder {
		catProps := GetPropertiesByCategory(mt.properties, category)
		if len(catProps) == 0 {
			continue
		}
		tabContent := mt.buildCategoryTable(catProps)
		mt.tabs.Append(container.NewTabItem(category, tabContent))
	}

	return widget.NewSimpleRenderer(mt.tabs)
}

// buildCategoryTable creates a table widget for a category's properties.
// Displays property name with model indicator [P] and value. Tooltips show descriptions.
func (mt *MaterialTable) buildCategoryTable(props []FormattedProperty) fyne.CanvasObject {
	if len(props) == 0 {
		return widget.NewLabel("No properties in this category")
	}

	// Create table with property name (with model indicator) and value columns
	table := widget.NewTable(
		// Data size: 3 columns (model indicator, name, value)
		func() (int, int) {
			return len(props), 3
		},
		// Create cell
		func() fyne.CanvasObject {
			label := widget.NewLabel("Template")
			label.Wrapping = fyne.TextWrapOff
			return label
		},
		// Update cell
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)
			prop := props[id.Row]
			switch id.Col {
			case 0:
				// Model indicator column
				modelStr := prop.Models.String()
				label.SetText(modelStr)
				label.TextStyle = fyne.TextStyle{Bold: true}
			case 1:
				// Property name
				label.SetText(prop.Name)
				label.TextStyle = fyne.TextStyle{Bold: false}
			case 2:
				// Value
				label.SetText(prop.Value)
				label.TextStyle = fyne.TextStyle{Monospace: true}
			}
		},
	)

	// Set column widths
	table.SetColumnWidth(0, 30)  // Model indicator [P]
	table.SetColumnWidth(1, 200) // Property name
	table.SetColumnWidth(2, 180) // Value

	// Create info panel for showing tooltips
	infoLabel := widget.NewLabel("Hover over a property to see its description.")
	infoLabel.Wrapping = fyne.TextWrapWord
	infoLabel.TextStyle = fyne.TextStyle{Italic: true}

	// Track selection for tooltip display
	table.OnSelected = func(id widget.TableCellID) {
		if id.Row >= 0 && id.Row < len(props) {
			prop := props[id.Row]
			desc := prop.Description
			if prop.Models.Preisach {
				desc = "[P] = Used in Preisach model\n\n" + desc
			}
			infoLabel.SetText(desc)
		}
		// Unselect to allow re-selection
		table.UnselectAll()
	}

	// Layout: table above, info panel below
	infoBox := container.NewVBox(
		canvas.NewRectangle(color.RGBA{60, 70, 90, 255}),
		container.NewPadded(infoLabel),
	)

	return container.NewBorder(
		nil,
		infoBox,
		nil, nil,
		container.NewVScroll(table),
	)
}

// MinSize returns the minimum size for the table.
func (mt *MaterialTable) MinSize() fyne.Size {
	return fyne.NewSize(400, 300)
}

// MaterialDetailPanel shows detailed information about a material including
// description, references, and all properties in a tabbed interface.
type MaterialDetailPanel struct {
	widget.BaseWidget

	material   *physics.Material
	properties []FormattedProperty

	// UI components
	nameLabel  *widget.Label
	descLabel  *widget.Label
	refLabel   *widget.Label
	propsTable *MaterialTable
}

// NewMaterialDetailPanel creates a new detailed material panel.
func NewMaterialDetailPanel(material *physics.Material) *MaterialDetailPanel {
	mdp := &MaterialDetailPanel{
		material:   material,
		properties: GetMaterialProperties(material),
	}
	mdp.ExtendBaseWidget(mdp)
	return mdp
}

// SetMaterial updates the displayed material.
func (mdp *MaterialDetailPanel) SetMaterial(material *physics.Material) {
	if material == nil {
		return
	}
	mdp.material = material
	mdp.properties = GetMaterialProperties(material)
	mdp.Refresh()
}

// CreateRenderer creates the widget renderer.
func (mdp *MaterialDetailPanel) CreateRenderer() fyne.WidgetRenderer {
	// Header section with name, description, and reference
	mdp.nameLabel = widget.NewLabelWithStyle(
		mdp.material.Name,
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	// Description with wrapping
	mdp.descLabel = widget.NewLabel(mdp.material.Description)
	mdp.descLabel.Wrapping = fyne.TextWrapWord

	// Reference with citation styling
	mdp.refLabel = widget.NewLabel(mdp.material.Reference)
	mdp.refLabel.Wrapping = fyne.TextWrapWord
	mdp.refLabel.TextStyle = fyne.TextStyle{Italic: true}

	// Section dividers
	divider1 := canvas.NewRectangle(color.RGBA{80, 90, 110, 255})
	divider1.SetMinSize(fyne.NewSize(0, 1))
	divider2 := canvas.NewRectangle(color.RGBA{80, 90, 110, 255})
	divider2.SetMinSize(fyne.NewSize(0, 1))

	// Header content
	headerBox := container.NewVBox(
		mdp.nameLabel,
		divider1,
		container.NewPadded(mdp.descLabel),
		divider2,
		container.NewHBox(
			widget.NewLabelWithStyle("Reference: ", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		),
		container.NewPadded(mdp.refLabel),
	)

	// Properties table
	mdp.propsTable = NewMaterialTable(mdp.material)

	// Main layout: header at top, table fills rest
	content := container.NewBorder(
		headerBox,
		nil, nil, nil,
		mdp.propsTable,
	)

	return &materialDetailPanelRenderer{
		panel:   mdp,
		content: content,
	}
}

// MinSize returns the minimum size.
func (mdp *MaterialDetailPanel) MinSize() fyne.Size {
	return fyne.NewSize(450, 400)
}

type materialDetailPanelRenderer struct {
	panel   *MaterialDetailPanel
	content *fyne.Container
}

func (r *materialDetailPanelRenderer) Layout(size fyne.Size) {
	r.content.Resize(size)
}

func (r *materialDetailPanelRenderer) MinSize() fyne.Size {
	return r.panel.MinSize()
}

func (r *materialDetailPanelRenderer) Refresh() {
	if r.panel.material == nil {
		return
	}
	r.panel.nameLabel.SetText(r.panel.material.Name)
	r.panel.descLabel.SetText(r.panel.material.Description)
	r.panel.refLabel.SetText(r.panel.material.Reference)
	r.panel.propsTable.SetMaterial(r.panel.material)
	r.content.Refresh()
}

func (r *materialDetailPanelRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.content}
}

func (r *materialDetailPanelRenderer) Destroy() {}
