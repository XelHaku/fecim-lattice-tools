// pkg/gui/tabs/learn_visuals.go
// Visual components for the Learn tab - OpenLane flow diagram and file format previews

package tabs

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Colors for the diagrams
var (
	colorBgDark      = color.RGBA{0, 30, 60, 255}
	colorBoxStandard = color.RGBA{0, 80, 160, 255}
	colorBoxOurs     = color.RGBA{0, 180, 220, 255} // Cyan for our contribution
	colorArrow       = color.RGBA{100, 150, 200, 255}
	colorText        = color.RGBA{255, 255, 255, 255}
	colorTextMuted   = color.RGBA{180, 200, 220, 255}
	colorWL          = color.RGBA{255, 100, 100, 255} // Red for word lines
	colorBL          = color.RGBA{100, 200, 255, 255} // Blue for bit lines
	colorFeFET       = color.RGBA{255, 200, 50, 255}  // Gold for FeFET devices
	colorHighlight   = color.RGBA{0, 255, 150, 255}   // Green highlight
	colorStorage     = color.RGBA{100, 150, 255, 255}  // Blue for Storage mode
	colorMemory      = color.RGBA{255, 200, 50, 255}   // Gold for Memory mode
	colorCompute     = color.RGBA{0, 255, 150, 255}    // Green for Compute mode
	colorTableHeader = color.RGBA{0, 60, 120, 255}    // Dark blue table headers
	colorTableRow1   = color.RGBA{0, 40, 80, 255}     // Alternating row 1
	colorTableRow2   = color.RGBA{0, 50, 100, 255}    // Alternating row 2
)

// =============================================================================
// OPENLANE FLOW DIAGRAM
// =============================================================================

// OpenLaneFlowDiagram creates a visual pipeline diagram
func OpenLaneFlowDiagram() fyne.CanvasObject {
	// Diagram dimensions - INCREASED 40%
	boxW := float32(150)
	boxH := float32(65)
	spacing := float32(30)
	startX := float32(30)
	startY := float32(50)

	objects := []fyne.CanvasObject{}

	// Title
	title := canvas.NewText("OpenLane RTL-to-GDSII Flow", colorText)
	title.TextSize = 18
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Move(fyne.NewPos(startX, 8))
	objects = append(objects, title)

	// Subtitle
	subtitle := canvas.NewText("CYAN = Our Array Builder files inject here", colorBoxOurs)
	subtitle.TextSize = 13
	subtitle.TextStyle = fyne.TextStyle{Bold: true}
	subtitle.Move(fyne.NewPos(startX, 30))
	objects = append(objects, subtitle)

	// Pipeline stages - row 1
	stages := []struct {
		name    string
		tool    string
		isOurs  bool
		x, y    float32
	}{
		{"Verilog", "Input", false, startX, startY},
		{"Synthesis", "Yosys", false, startX + boxW + spacing, startY},
		{"Floorplan", "OpenROAD", false, startX + 2*(boxW+spacing), startY},
		{"Placement", "RePlAce", false, startX + 3*(boxW+spacing), startY},
	}

	// Row 2 (reversed flow) - more vertical space
	row2Y := startY + boxH + spacing + 30
	stages = append(stages, []struct {
		name    string
		tool    string
		isOurs  bool
		x, y    float32
	}{
		{"CTS", "TritonCTS", false, startX + 3*(boxW+spacing), row2Y},
		{"Routing", "TritonRoute", false, startX + 2*(boxW+spacing), row2Y},
		{"Signoff", "Magic/LVS", false, startX + boxW + spacing, row2Y},
		{"GDSII", "Output", false, startX, row2Y},
	}...)

	// Mark which stages we contribute to
	stages[0].isOurs = true  // Verilog - we provide
	stages[2].isOurs = true  // Floorplan - our LEF
	stages[3].isOurs = true  // Placement - our DEF (FIXED)

	// Draw stages
	for i, stage := range stages {
		boxColor := colorBoxStandard
		if stage.isOurs {
			boxColor = colorBoxOurs
		}

		// Box
		box := canvas.NewRectangle(boxColor)
		box.Resize(fyne.NewSize(boxW, boxH))
		box.Move(fyne.NewPos(stage.x, stage.y))
		box.CornerRadius = 6
		objects = append(objects, box)

		// Stage name
		nameText := canvas.NewText(stage.name, colorText)
		nameText.TextSize = 13
		nameText.TextStyle = fyne.TextStyle{Bold: true}
		nameText.Move(fyne.NewPos(stage.x+10, stage.y+10))
		objects = append(objects, nameText)

		// Tool name
		toolText := canvas.NewText(stage.tool, colorTextMuted)
		toolText.TextSize = 11
		toolText.Move(fyne.NewPos(stage.x+10, stage.y+30))
		objects = append(objects, toolText)

		// Draw arrows between stages
		if i > 0 && i < 4 {
			// Row 1 arrows (left to right)
			arrow := createArrow(
				stages[i-1].x+boxW, stages[i-1].y+boxH/2,
				stage.x, stage.y+boxH/2,
			)
			objects = append(objects, arrow...)
		} else if i == 4 {
			// Down arrow from Placement to CTS
			arrow := createArrowDown(
				stages[3].x+boxW/2, stages[3].y+boxH,
				stage.x+boxW/2, stage.y,
			)
			objects = append(objects, arrow...)
		} else if i > 4 {
			// Row 2 arrows (right to left)
			arrow := createArrow(
				stages[i-1].x, stages[i-1].y+boxH/2,
				stage.x+boxW, stage.y+boxH/2,
			)
			objects = append(objects, arrow...)
		}
	}

	// Add "Our Files" labels
	// LEF label
	lefLabel := canvas.NewText("Our LEF", colorHighlight)
	lefLabel.TextSize = 10
	lefLabel.Move(fyne.NewPos(stages[2].x+10, stages[2].y-15))
	objects = append(objects, lefLabel)

	// DEF label
	defLabel := canvas.NewText("Our DEF (FIXED)", colorHighlight)
	defLabel.TextSize = 10
	defLabel.Move(fyne.NewPos(stages[3].x+5, stages[3].y-15))
	objects = append(objects, defLabel)

	// Verilog label
	vLabel := canvas.NewText("Our Verilog", colorHighlight)
	vLabel.TextSize = 10
	vLabel.Move(fyne.NewPos(stages[0].x+5, stages[0].y-15))
	objects = append(objects, vLabel)

	// Container with fixed size - INCREASED
	cont := container.NewWithoutLayout(objects...)
	cont.Resize(fyne.NewSize(780, 320))

	return cont
}

// createArrow creates a horizontal arrow
func createArrow(x1, y1, x2, y2 float32) []fyne.CanvasObject {
	// Line
	line := canvas.NewLine(colorArrow)
	line.StrokeWidth = 2
	line.Position1 = fyne.NewPos(x1, y1)
	line.Position2 = fyne.NewPos(x2-8, y2)

	// Arrowhead
	head := canvas.NewLine(colorArrow)
	head.StrokeWidth = 2
	head.Position1 = fyne.NewPos(x2-12, y2-5)
	head.Position2 = fyne.NewPos(x2-4, y2)

	head2 := canvas.NewLine(colorArrow)
	head2.StrokeWidth = 2
	head2.Position1 = fyne.NewPos(x2-12, y2+5)
	head2.Position2 = fyne.NewPos(x2-4, y2)

	return []fyne.CanvasObject{line, head, head2}
}

// createArrowDown creates a vertical down arrow
func createArrowDown(x1, y1, x2, y2 float32) []fyne.CanvasObject {
	line := canvas.NewLine(colorArrow)
	line.StrokeWidth = 2
	line.Position1 = fyne.NewPos(x1, y1)
	line.Position2 = fyne.NewPos(x2, y2-8)

	head := canvas.NewLine(colorArrow)
	head.StrokeWidth = 2
	head.Position1 = fyne.NewPos(x2-5, y2-12)
	head.Position2 = fyne.NewPos(x2, y2-4)

	head2 := canvas.NewLine(colorArrow)
	head2.StrokeWidth = 2
	head2.Position1 = fyne.NewPos(x2+5, y2-12)
	head2.Position2 = fyne.NewPos(x2, y2-4)

	return []fyne.CanvasObject{line, head, head2}
}

// =============================================================================
// FILE FORMAT PREVIEW CARDS
// =============================================================================

// FileFormatCard creates a styled card showing file format examples
func FileFormatCard(title, format, content string) fyne.CanvasObject {
	// Header - INCREASED HEIGHT for better readability
	headerBg := canvas.NewRectangle(colorBoxOurs)
	headerBg.Resize(fyne.NewSize(360, 40))
	headerBg.CornerRadius = 6

	titleText := canvas.NewText(title+" (."+format+")", colorBgDark)
	titleText.TextSize = 14
	titleText.TextStyle = fyne.TextStyle{Bold: true}
	titleText.Move(fyne.NewPos(12, 10))

	// Content area - INCREASED HEIGHT for better code visibility
	contentBg := canvas.NewRectangle(color.RGBA{0, 25, 50, 255})
	contentBg.Resize(fyne.NewSize(360, 180))
	contentBg.Move(fyne.NewPos(0, 40))
	contentBg.CornerRadius = 4

	// Code content with monospace style and better line spacing
	codeLabel := widget.NewLabel(content)
	codeLabel.Wrapping = fyne.TextWrapOff
	codeLabel.TextStyle = fyne.TextStyle{Monospace: true}

	codeContainer := container.NewPadded(codeLabel)
	codeContainer.Move(fyne.NewPos(8, 44))
	codeContainer.Resize(fyne.NewSize(344, 172))

	card := container.NewWithoutLayout(headerBg, titleText, contentBg, codeContainer)
	card.Resize(fyne.NewSize(360, 224))

	return card
}

// LEFPreviewCard shows a LEF file example
func LEFPreviewCard() fyne.CanvasObject {
	content := `MACRO fecim_bitcell
  SIZE 0.460 BY 2.720 ;
  PIN WL
    DIRECTION INPUT ;
  END WL
END fecim_bitcell`
	return FileFormatCard("LEF", "lef", content)
}

// DEFPreviewCard shows a DEF file example
func DEFPreviewCard() fyne.CanvasObject {
	content := `COMPONENTS 16 ;
  - cell_0_0 fecim_bitcell
    + FIXED ( 10000 10000 ) N ;
  - cell_0_1 fecim_bitcell
    + FIXED ( 10460 10000 ) N ;
END COMPONENTS`
	return FileFormatCard("DEF", "def", content)
}

// VerilogPreviewCard shows a Verilog file example
func VerilogPreviewCard() fyne.CanvasObject {
	content := `module fecim_array_4x4 (
  input  [3:0] WL,
  inout  [3:0] BL
);
  fecim_bitcell c00(.WL(WL[0]),.BL(BL[0]));
  // ... 16 cells
endmodule`
	return FileFormatCard("Verilog", "v", content)
}

// LibertyPreviewCard shows a Liberty timing file example
func LibertyPreviewCard() fyne.CanvasObject {
	content := `library(fecim_cells) {
  cell(fecim_bitcell) {
    area : 1.2512 ;
    pin(WL) { capacitance: 0.002; }
    timing() { cell_rise: 0.1ns; }
  }
}`
	return FileFormatCard("Liberty", "lib", content)
}

// =============================================================================
// REFERENCES CARD
// =============================================================================

// ReferencesCard creates an IEEE-formatted references section with category headers
func ReferencesCard() fyne.CanvasObject {
	objects := []fyne.CanvasObject{}

	// Card dimensions
	cardWidth := float32(420)
	cardHeight := float32(310)

	// Category header dimensions
	headerHeight := float32(26)
	categorySpacing := float32(8)

	// Starting positions
	startX := float32(0)
	currentY := float32(0)

	// Category 1: EDA / OpenLane
	header1 := canvas.NewRectangle(colorBoxOurs)
	header1.Resize(fyne.NewSize(cardWidth, headerHeight))
	header1.Move(fyne.NewPos(startX, currentY))
	header1.CornerRadius = 4
	objects = append(objects, header1)

	headerText1 := canvas.NewText("EDA / OpenLane", colorText)
	headerText1.TextSize = 12
	headerText1.TextStyle = fyne.TextStyle{Bold: true}
	headerText1.Move(fyne.NewPos(startX+10, currentY+6))
	objects = append(objects, headerText1)

	currentY += headerHeight

	// References for EDA / OpenLane
	refs1 := `[1] M. Shalan et al., "OpenLANE: Digital ASIC Flow," WOSET, 2020.
[2] LEF/DEF 5.8 Specification, Si2 Coalition.`

	refBg1 := canvas.NewRectangle(colorBgDark)
	refBg1.Resize(fyne.NewSize(cardWidth, 50))
	refBg1.Move(fyne.NewPos(startX, currentY))
	objects = append(objects, refBg1)

	refLabel1 := widget.NewLabel(refs1)
	refLabel1.Wrapping = fyne.TextWrapWord
	refLabel1.TextStyle = fyne.TextStyle{}
	refContainer1 := container.NewPadded(refLabel1)
	refContainer1.Move(fyne.NewPos(startX+5, currentY+2))
	refContainer1.Resize(fyne.NewSize(cardWidth-10, 46))
	objects = append(objects, refContainer1)

	currentY += 50 + categorySpacing

	// Category 2: FeCIM Device Physics
	header2 := canvas.NewRectangle(colorBoxOurs)
	header2.Resize(fyne.NewSize(cardWidth, headerHeight))
	header2.Move(fyne.NewPos(startX, currentY))
	header2.CornerRadius = 4
	objects = append(objects, header2)

	headerText2 := canvas.NewText("FeCIM Device Physics", colorText)
	headerText2.TextSize = 12
	headerText2.TextStyle = fyne.TextStyle{Bold: true}
	headerText2.Move(fyne.NewPos(startX+10, currentY+6))
	objects = append(objects, headerText2)

	currentY += headerHeight

	// References for FeCIM Device Physics
	refs2 := `[3] S. Shin et al., "Flash In2Se3," Adv. Electron. Mater., 2025.
[4] U. Schroeder et al., "Roadmap Ferroelectric HfZrO," APL Mater., 2023.`

	refBg2 := canvas.NewRectangle(colorBgDark)
	refBg2.Resize(fyne.NewSize(cardWidth, 50))
	refBg2.Move(fyne.NewPos(startX, currentY))
	objects = append(objects, refBg2)

	refLabel2 := widget.NewLabel(refs2)
	refLabel2.Wrapping = fyne.TextWrapWord
	refLabel2.TextStyle = fyne.TextStyle{}
	refContainer2 := container.NewPadded(refLabel2)
	refContainer2.Move(fyne.NewPos(startX+5, currentY+2))
	refContainer2.Resize(fyne.NewSize(cardWidth-10, 46))
	objects = append(objects, refContainer2)

	currentY += 50 + categorySpacing

	// Category 3: CIM Architecture
	header3 := canvas.NewRectangle(colorBoxOurs)
	header3.Resize(fyne.NewSize(cardWidth, headerHeight))
	header3.Move(fyne.NewPos(startX, currentY))
	header3.CornerRadius = 4
	objects = append(objects, header3)

	headerText3 := canvas.NewText("CIM Architecture", colorText)
	headerText3.TextSize = 12
	headerText3.TextStyle = fyne.TextStyle{Bold: true}
	headerText3.Move(fyne.NewPos(startX+10, currentY+6))
	objects = append(objects, headerText3)

	currentY += headerHeight

	// References for CIM Architecture
	refs3 := `[5] Y. Chen, "Sneak Path Solutions," RSC Nanoscale Adv., 2020.
[6] P. Chen, NeuroSim, Georgia Tech, 2021.`

	refBg3 := canvas.NewRectangle(colorBgDark)
	refBg3.Resize(fyne.NewSize(cardWidth, 50))
	refBg3.Move(fyne.NewPos(startX, currentY))
	objects = append(objects, refBg3)

	refLabel3 := widget.NewLabel(refs3)
	refLabel3.Wrapping = fyne.TextWrapWord
	refLabel3.TextStyle = fyne.TextStyle{}
	refContainer3 := container.NewPadded(refLabel3)
	refContainer3.Move(fyne.NewPos(startX+5, currentY+2))
	refContainer3.Resize(fyne.NewSize(cardWidth-10, 46))
	objects = append(objects, refContainer3)

	// Container with fixed size
	cont := container.NewWithoutLayout(objects...)
	cont.Resize(fyne.NewSize(cardWidth, cardHeight))

	return cont
}
