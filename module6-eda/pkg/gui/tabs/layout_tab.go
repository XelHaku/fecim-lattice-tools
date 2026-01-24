// pkg/gui/tabs/layout_tab.go
package tabs

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/compiler"
	sharedwidgets "multilayer-ferroelectric-cim-visualizer/shared/widgets"
)

// MakeLayoutTab creates the layout visualization tab
func MakeLayoutTab(state interface{}) fyne.CanvasObject {
	appState := state.(*AppState)

	// Cell info display
	cellInfo := widget.NewLabel("Click on a cell to see details")
	cellInfo.Wrapping = fyne.TextWrapWord

	// Grid container that will be populated
	gridContainer := container.NewStack()

	// Refresh button
	refreshButton := widget.NewButton("Refresh Grid", func() {
		if appState.CurrentMapping == nil {
			cellInfo.SetText("No mapping available. Compile weights first (Tab 1).")
			return
		}

		// Build grid visualization
		grid := buildGrid(appState.CurrentMapping, cellInfo)
		gridContainer.Objects = []fyne.CanvasObject{grid}
		fyne.Do(func() {
			gridContainer.Refresh()
		})
		cellInfo.SetText("Grid loaded. Click a cell for details.")
	})

	// Legend
	legend := buildLegend()

	// Layout
	topSection := container.NewVBox(
		widget.NewLabel("CROSSBAR LAYOUT VISUALIZATION"),
		widget.NewLabel("Colors: Blue = low conductance, Red = high conductance"),
		refreshButton,
		widget.NewSeparator(),
	)

	bottomSection := container.NewVBox(
		widget.NewSeparator(),
		legend,
		widget.NewSeparator(),
		widget.NewLabel("CELL DETAILS"),
		cellInfo,
	)

	// Main content with scrollable grid
	scrollableGrid := container.NewScroll(gridContainer)
	scrollableGrid.SetMinSize(fyne.NewSize(400, 300))

	content := container.NewBorder(topSection, bottomSection, nil, nil, scrollableGrid)

	return content
}

func buildGrid(mapping *compiler.CrossbarMapping, cellInfo *widget.Label) fyne.CanvasObject {
	if len(mapping.Cells) == 0 {
		return widget.NewLabel("No cells to display")
	}

	// Find grid dimensions
	maxRow, maxCol := 0, 0
	for _, cell := range mapping.Cells {
		if cell.Row > maxRow {
			maxRow = cell.Row
		}
		if cell.Col > maxCol {
			maxCol = cell.Col
		}
	}

	rows := maxRow + 1
	cols := maxCol + 1

	// Create cell lookup map
	cellMap := make(map[string]*compiler.CellAssignment)
	for i := range mapping.Cells {
		key := fmt.Sprintf("%d_%d", mapping.Cells[i].Row, mapping.Cells[i].Col)
		cellMap[key] = &mapping.Cells[i]
	}

	// Cell size
	cellSize := float32(20)
	if rows > 32 || cols > 32 {
		cellSize = 10
	}
	if rows > 64 || cols > 64 {
		cellSize = 5
	}

	// Build grid of colored rectangles
	var objects []fyne.CanvasObject

	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			key := fmt.Sprintf("%d_%d", r, c)
			cell := cellMap[key]

			var cellColor color.Color
			if cell != nil {
				// Color based on conductance (blue = low, red = high)
				t := (cell.Conductance - mapping.Config.GMin) / (mapping.Config.GMax - mapping.Config.GMin)
				cellColor = conductanceToColor(t)
			} else {
				cellColor = color.RGBA{50, 50, 50, 255} // Gray for empty
			}

			rect := canvas.NewRectangle(cellColor)
			rect.SetMinSize(fyne.NewSize(cellSize, cellSize))

			// Create clickable container
			row, col := r, c
			clickable := newClickableCell(rect, func() {
				if cell := cellMap[fmt.Sprintf("%d_%d", row, col)]; cell != nil {
					cellInfo.SetText(fmt.Sprintf(
						"Cell [%d, %d]\n\n"+
							"Weight: %.6f\n"+
							"Quant Level: %d / %d\n"+
							"Conductance: %.4f μS\n"+
							"Program V: %.4f V\n"+
							"Resistance: %.2f Ω",
						cell.Row, cell.Col,
						cell.WeightValue,
						cell.QuantLevel, mapping.Config.Levels-1,
						cell.Conductance,
						cell.ProgramV,
						1e6/cell.Conductance))
				} else {
					cellInfo.SetText(fmt.Sprintf("Cell [%d, %d]: Empty", row, col))
				}
			})

			objects = append(objects, clickable)
		}
	}

	// Arrange in grid
	grid := container.NewGridWithColumns(cols, objects...)

	return grid
}

func conductanceToColor(t float64) color.Color {
	// Blue (low) to Red (high) gradient
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}

	// Blue -> Cyan -> Green -> Yellow -> Red
	var r, g, b uint8
	if t < 0.25 {
		// Blue to Cyan
		r = 0
		g = uint8(255 * t * 4)
		b = 255
	} else if t < 0.5 {
		// Cyan to Green
		r = 0
		g = 255
		b = uint8(255 * (1 - (t-0.25)*4))
	} else if t < 0.75 {
		// Green to Yellow
		r = uint8(255 * (t - 0.5) * 4)
		g = 255
		b = 0
	} else {
		// Yellow to Red
		r = 255
		g = uint8(255 * (1 - (t-0.75)*4))
		b = 0
	}

	return color.RGBA{r, g, b, 255}
}

func buildLegend() fyne.CanvasObject {
	legendItems := []fyne.CanvasObject{}

	// Create gradient legend
	for i := 0; i < 10; i++ {
		t := float64(i) / 9.0
		rect := canvas.NewRectangle(conductanceToColor(t))
		rect.SetMinSize(fyne.NewSize(20, 15))
		legendItems = append(legendItems, rect)
	}

	gradient := container.NewHBox(legendItems...)

	labels := container.NewHBox(
		widget.NewLabel("Low G"),
		widget.NewLabel("                    "),
		widget.NewLabel("High G"),
	)

	return container.NewVBox(
		widget.NewLabel("Conductance Legend:"),
		gradient,
		labels,
	)
}

// clickableCell wraps a canvas object with click handling
type clickableCell struct {
	widget.BaseWidget
	content fyne.CanvasObject
	onClick func()
}

func newClickableCell(content fyne.CanvasObject, onClick func()) *clickableCell {
	c := &clickableCell{content: content, onClick: onClick}
	c.ExtendBaseWidget(c)
	return c
}

func (c *clickableCell) CreateRenderer() fyne.WidgetRenderer {
	return &clickableCellRenderer{cell: c}
}

func (c *clickableCell) Tapped(*fyne.PointEvent) {
	if c.onClick != nil {
		c.onClick()
	}
}

type clickableCellRenderer struct {
	cell  *clickableCell
	cache sharedwidgets.LayoutCache // Shared utility for safe layout
}

func (r *clickableCellRenderer) Layout(size fyne.Size) {
	sharedwidgets.DebugLayoutCall("clickableCellRenderer", size)
	if !r.cache.ShouldLayout(size) {
		return
	}
	r.cell.content.Resize(size)
	r.cache.MarkLayout(size)
}

func (r *clickableCellRenderer) MinSize() fyne.Size {
	return r.cell.content.MinSize()
}

func (r *clickableCellRenderer) Refresh() {
	sharedwidgets.DebugRefreshCall("clickableCellRenderer", r.cell.content.Size())
	r.cell.content.Refresh()
	// Mark current size only if valid
	size := r.cell.content.Size()
	if size.Width > 0 && size.Height > 0 {
		r.cache.MarkLayout(size)
	}
}

func (r *clickableCellRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.cell.content}
}

func (r *clickableCellRenderer) Destroy() {}
