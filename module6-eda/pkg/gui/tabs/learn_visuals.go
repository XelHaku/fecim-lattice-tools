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
	colorStorage     = color.RGBA{100, 150, 255, 255} // Blue for Storage mode
	colorMemory      = color.RGBA{255, 200, 50, 255}  // Gold for Memory mode
	colorCompute     = color.RGBA{0, 255, 150, 255}   // Green for Compute mode
	colorTableHeader = color.RGBA{0, 60, 120, 255}    // Dark blue table headers
	colorTableRow1   = color.RGBA{0, 40, 80, 255}     // Alternating row 1
	colorTableRow2   = color.RGBA{0, 50, 100, 255}    // Alternating row 2
)

// =============================================================================
// OPENLANE FLOW DIAGRAM
// =============================================================================

// OpenLaneFlowDiagram creates a visual pipeline diagram
func OpenLaneFlowDiagram() fyne.CanvasObject {
	// Diagram dimensions - LARGER for better readability
	boxW := float32(160)
	boxH := float32(70)
	spacing := float32(35)
	startX := float32(20)
	startY := float32(55)

	objects := []fyne.CanvasObject{}

	// Title - larger
	title := canvas.NewText("OpenLane RTL-to-GDSII Assembly", colorText)
	title.TextSize = 20
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Move(fyne.NewPos(startX, 5))
	objects = append(objects, title)

	// Subtitle with better contrast
	subtitle := canvas.NewText("CYAN boxes = Our Array Builder contributes here", colorBoxOurs)
	subtitle.TextSize = 14
	subtitle.TextStyle = fyne.TextStyle{Bold: true}
	subtitle.Move(fyne.NewPos(startX, 30))
	objects = append(objects, subtitle)

	// Pipeline stages - row 1
	stages := []struct {
		name   string
		tool   string
		isOurs bool
		x, y   float32
	}{
		{"Verilog", "Input", false, startX, startY},
		{"Synthesis", "Yosys", false, startX + boxW + spacing, startY},
		{"Floorplan", "OpenROAD", false, startX + 2*(boxW+spacing), startY},
		{"Placement", "OpenROAD", false, startX + 3*(boxW+spacing), startY},
	}

	// Row 2 (reversed flow) - more vertical space
	row2Y := startY + boxH + spacing + 30
	stages = append(stages, []struct {
		name   string
		tool   string
		isOurs bool
		x, y   float32
	}{
		{"CTS", "OpenROAD", false, startX + 3*(boxW+spacing), row2Y},
		{"Routing", "TritonRoute", false, startX + 2*(boxW+spacing), row2Y},
		{"Signoff", "Magic/LVS", false, startX + boxW + spacing, row2Y},
		{"GDSII", "Assembly", false, startX, row2Y},
	}...)

	// Mark which stages we contribute to
	stages[0].isOurs = true // Verilog - we provide
	stages[2].isOurs = true // Floorplan - our LEF
	stages[3].isOurs = true // Placement - our DEF (FIXED)

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
		nameText.TextSize = 14
		nameText.TextStyle = fyne.TextStyle{Bold: true}
		nameText.Move(fyne.NewPos(stage.x+10, stage.y+10))
		objects = append(objects, nameText)

		// Tool name
		toolText := canvas.NewText(stage.tool, colorTextMuted)
		toolText.TextSize = 14
		toolText.Move(fyne.NewPos(stage.x+10, stage.y+30))
		objects = append(objects, toolText)

		// Add checkmark indicator for our contributions
		if stage.isOurs {
			checkmark := canvas.NewText("✓", colorHighlight)
			checkmark.TextSize = 16
			checkmark.TextStyle = fyne.TextStyle{Bold: true}
			checkmark.Move(fyne.NewPos(stage.x+boxW-20, stage.y+8))
			objects = append(objects, checkmark)
		}

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

	// Add "Our Files" labels with connecting lines
	// Verilog label
	vLine := canvas.NewLine(colorHighlight)
	vLine.StrokeWidth = 1
	vLine.Position1 = fyne.NewPos(stages[0].x+boxW/2, stages[0].y-5)
	vLine.Position2 = fyne.NewPos(stages[0].x+boxW/2, stages[0].y)
	objects = append(objects, vLine)

	vLabel := canvas.NewText("↑ Our Verilog", colorHighlight)
	vLabel.TextSize = 14
	vLabel.TextStyle = fyne.TextStyle{Bold: true}
	vLabel.Move(fyne.NewPos(stages[0].x+boxW/2-40, stages[0].y-20))
	objects = append(objects, vLabel)

	// LEF label
	lefLine := canvas.NewLine(colorHighlight)
	lefLine.StrokeWidth = 1
	lefLine.Position1 = fyne.NewPos(stages[2].x+boxW/2, stages[2].y-5)
	lefLine.Position2 = fyne.NewPos(stages[2].x+boxW/2, stages[2].y)
	objects = append(objects, lefLine)

	lefLabel := canvas.NewText("↑ Our LEF", colorHighlight)
	lefLabel.TextSize = 14
	lefLabel.TextStyle = fyne.TextStyle{Bold: true}
	lefLabel.Move(fyne.NewPos(stages[2].x+boxW/2-25, stages[2].y-20))
	objects = append(objects, lefLabel)

	// DEF label
	defLine := canvas.NewLine(colorHighlight)
	defLine.StrokeWidth = 1
	defLine.Position1 = fyne.NewPos(stages[3].x+boxW/2, stages[3].y-5)
	defLine.Position2 = fyne.NewPos(stages[3].x+boxW/2, stages[3].y)
	objects = append(objects, defLine)

	defLabel := canvas.NewText("↑ Our DEF (FIXED)", colorHighlight)
	defLabel.TextSize = 14
	defLabel.TextStyle = fyne.TextStyle{Bold: true}
	defLabel.Move(fyne.NewPos(stages[3].x+boxW/2-50, stages[3].y-20))
	objects = append(objects, defLabel)

	// Add legend at bottom
	legendY := row2Y + boxH + float32(25)

	// Standard stage box
	legendBox1 := canvas.NewRectangle(colorBoxStandard)
	legendBox1.Resize(fyne.NewSize(20, 15))
	legendBox1.Move(fyne.NewPos(startX, legendY))
	legendBox1.CornerRadius = 3
	objects = append(objects, legendBox1)

	legendText1 := canvas.NewText("= OpenLane handles", colorTextMuted)
	legendText1.TextSize = 14
	legendText1.Move(fyne.NewPos(startX+25, legendY))
	objects = append(objects, legendText1)

	// Our contribution box
	legendBox2 := canvas.NewRectangle(colorBoxOurs)
	legendBox2.Resize(fyne.NewSize(20, 15))
	legendBox2.Move(fyne.NewPos(startX+180, legendY))
	legendBox2.CornerRadius = 3
	objects = append(objects, legendBox2)

	legendText2 := canvas.NewText("= We provide input files", colorBoxOurs)
	legendText2.TextSize = 14
	legendText2.Move(fyne.NewPos(startX+205, legendY))
	objects = append(objects, legendText2)

	// Calculate proper container bounds with legend
	totalWidth := startX + 4*(boxW+spacing)
	totalHeight := legendY + float32(30)

	cont := container.NewWithoutLayout(objects...)
	cont.Resize(fyne.NewSize(totalWidth, totalHeight))

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
// Uses proper Fyne widgets for responsive layout
func FileFormatCard(title, format, content string) fyne.CanvasObject {
	// Code content with monospace style
	codeLabel := widget.NewLabel(content)
	codeLabel.Wrapping = fyne.TextWrapOff
	codeLabel.TextStyle = fyne.TextStyle{Monospace: true}

	// Use widget.Card for proper layout and MinSize reporting
	card := widget.NewCard(title+" (."+format+")", "", codeLabel)

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

// SPICEPreviewCard shows a SPICE netlist example with FeCap subcircuit.
func SPICEPreviewCard() fyne.CanvasObject {
	content := `* FeCIM Crossbar Array SPICE
.subckt FECAP_HZO wl bl
C_ferro wl bl 28e-15 ; FeCap ~28 fF
.ends
.subckt fefet_cell wl bl
X_cap wl bl FECAP_HZO
R_level wl bl 18182 ; L15
.ends
X0_0 WL[0] BL[0] fefet_cell
; ...run: ngspice design.sp`
	return FileFormatCard("SPICE", "sp", content)
}

// CSVPreviewCard shows a CSV conductance table example.
func CSVPreviewCard() fyne.CanvasObject {
	content := `row,col,level,conductance_uS,resistance_ohm,program_V
0,0,0,10.0000,100000.00,2.0000
0,1,10,41.0345,24369.75,3.0345
0,2,20,72.0690,13876.69,4.0690
0,3,29,100.0000,10000.00,5.0000
; GMin=10µS(OFF) GMax=100µS(ON) 30 levels`
	return FileFormatCard("CSV Table", "csv", content)
}

// =============================================================================
// REFERENCES CARD
// =============================================================================

// ReferencesCard creates an IEEE-formatted references section with category headers
// Uses proper Fyne widgets for responsive layout
func ReferencesCard() fyne.CanvasObject {
	// Category 1: EDA / OpenLane
	refs1 := widget.NewLabel(`[1] M. Shalan et al., "OpenLANE: Digital ASIC Flow," WOSET, 2020.
[2] LEF/DEF 5.8 Specification, Si2 Coalition.`)
	refs1.Wrapping = fyne.TextWrapWord
	card1 := widget.NewCard("EDA / OpenLane", "", refs1)

	// Category 2: FeCIM Device Physics
	refs2 := widget.NewLabel(`[3] S. Shin et al., "Flash In2Se3," Adv. Electron. Mater., 2025.
[4] U. Schroeder et al., "Roadmap Ferroelectric HfZrO," APL Mater., 2023.`)
	refs2.Wrapping = fyne.TextWrapWord
	card2 := widget.NewCard("FeCIM Device Physics", "", refs2)

	// Category 3: CIM Architecture
	refs3 := widget.NewLabel(`[5] Y. Chen, "Sneak Path Solutions," RSC Nanoscale Adv., 2020.
[6] P. Chen, NeuroSim, Georgia Tech, 2021.`)
	refs3.Wrapping = fyne.TextWrapWord
	card3 := widget.NewCard("CIM Architecture", "", refs3)

	return container.NewVBox(card1, card2, card3)
}
