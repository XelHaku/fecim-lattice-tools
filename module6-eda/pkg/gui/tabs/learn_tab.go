// pkg/gui/tabs/learn_tab.go
// Learning Center tab for FeCIM Design Suite
// Explains OpenLane flow and where the FeCIM Array Builder fits in

package tabs

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// MakeLearnTab creates the learning center tab with educational content
func MakeLearnTab(state interface{}, w fyne.Window) fyne.CanvasObject {
	// Topic selector
	topics := []string{
		"1. What is FeCIM EDA?",
		"2. The Crossbar Architecture",
		"3. EDA Files We Generate",
	}

	topicSelector := widget.NewList(
		func() int { return len(topics) },
		func() fyne.CanvasObject {
			label := widget.NewLabel("Template")
			label.Wrapping = fyne.TextWrapWord
			return label
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			label.SetText(topics[id])
			// Add visual hierarchy with bold text for better readability
			if id == 0 || id == 1 || id == 2 {
				label.TextStyle = fyne.TextStyle{Bold: true}
			}
		},
	)
	topicSelector.OnSelected = func(id widget.ListItemID) {
		// Will be connected to content display
	}

	// Content area - increased width for better card layout
	contentScroll := container.NewScroll(makeIntroContent())
	contentScroll.SetMinSize(fyne.NewSize(750, 500))

	// Connect topic selector to content
	topicSelector.OnSelected = func(id widget.ListItemID) {
		var content fyne.CanvasObject
		switch id {
		case 0:
			content = makeIntroContent()
		case 1:
			content = makeCrossbarContent()
		case 2:
			content = makeFilesContent()
		default:
			content = makeIntroContent()
		}
		contentScroll.Content = content
		fyne.Do(func() {
			contentScroll.Refresh()
		})
	}

	// Select first topic by default
	topicSelector.Select(0)

	// Layout with sidebar - increased width and better spacing
	sidebarTitle := widget.NewLabelWithStyle("Topics", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	sidebarSpacer := widget.NewLabel("") // Add spacing after title
	sidebarSpacer.Resize(fyne.NewSize(1, 8))

	sidebar := container.NewBorder(
		container.NewVBox(sidebarTitle, sidebarSpacer),
		nil, nil, nil,
		topicSelector,
	)

	// Main layout - wider sidebar (240px instead of 180px)
	split := container.NewHSplit(sidebar, contentScroll)
	split.SetOffset(0.25) // Wider sidebar: 25% instead of 22%

	// Header
	header := container.NewVBox(
		widget.NewLabelWithStyle("FeCIM Array Builder - Learning Center", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Understanding OpenLane and where our tool fits in"),
		widget.NewSeparator(),
	)

	return container.NewBorder(header, nil, nil, nil, split)
}

// makeBulletList creates a formatted bullet list with a header
func makeBulletList(header string, items ...string) fyne.CanvasObject {
	headerLabel := widget.NewLabelWithStyle(header, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	var listItems []fyne.CanvasObject
	listItems = append(listItems, headerLabel)
	for _, item := range items {
		bullet := widget.NewLabel("  \u2022 " + item)
		bullet.Wrapping = fyne.TextWrapWord
		listItems = append(listItems, bullet)
	}
	return container.NewVBox(listItems...)
}

func makeIntroContent() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("What is FeCIM EDA?", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	intro := widget.NewLabel(`Module 6 is an ARRAY BUILDER that generates EDA files for integrating FeCIM crossbar arrays into the OpenLane flow. This is an educational tool that demonstrates how FeCIM arrays could integrate with open-source EDA. All timing values are placeholders - real values require SPICE characterization with validated models.`)
	intro.Wrapping = fyne.TextWrapWord

	// Larger operation modes visual - wrapped in padded container
	modesVisual := OperationModesVisual()
	modesContainer := container.NewPadded(modesVisual)
	modesContainer.Resize(fyne.NewSize(600, 280))

	// OpenLane flow diagram - wrapped in padded container
	flowDiagram := OpenLaneFlowDiagram()
	flowContainer := container.NewPadded(flowDiagram)
	flowContainer.Resize(fyne.NewSize(750, 320))

	// Stages Explained section
	stagesTitle := widget.NewLabelWithStyle("The Stages Explained", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	stagesText := widget.NewLabel(`1. SYNTHESIS (Yosys) - Converts behavioral Verilog to gate-level netlist
2. FLOORPLAN - Defines die area and I/O pin locations
3. PLACEMENT (RePlAce + OpenDP) - Assigns X,Y coordinates to every cell
4. CTS (Clock Tree Synthesis) - Distributes clock signal evenly (FeCIM arrays often skip this)
5. ROUTING (TritonRoute) - Draws metal wire connections
6. SIGNOFF & GDSII - DRC/LVS verification, final output`)
	stagesText.Wrapping = fyne.TextWrapWord

	// Two-column do/don't layout
	doList := makeBulletList("WHAT WE DO:",
		"Generate LEF files (cell abstracts)",
		"Generate Liberty files (timing - placeholder values)",
		"Generate Verilog netlists (behavioral models)",
		"Generate DEF files (physical placement)",
		"Export OpenLane configuration")

	dontList := makeBulletList("WHAT WE DON'T DO:",
		"We do NOT provide validated FeFET device models",
		"We do NOT generate production-ready layouts",
		"We do NOT characterize real timing values",
		"We do NOT fabricate chips")

	doColumns := container.NewGridWithColumns(2, doList, dontList)

	// Disclaimer banner using widget.NewCard
	disclaimerCard := widget.NewCard("", "DISCLAIMER",
		widget.NewLabel("This project is not affiliated with or endorsed by external research institution, Dr. external research group, or any foundry."))

	// Add spacer elements for better vertical separation
	spacer1 := widget.NewLabel("")
	spacer1.Resize(fyne.NewSize(1, 15))
	spacer2 := widget.NewLabel("")
	spacer2.Resize(fyne.NewSize(1, 15))
	spacer3 := widget.NewLabel("")
	spacer3.Resize(fyne.NewSize(1, 15))

	return container.NewVBox(
		title,
		widget.NewSeparator(),
		intro,
		spacer1,
		widget.NewSeparator(),
		modesContainer,
		spacer2,
		widget.NewSeparator(),
		flowContainer,
		spacer3,
		widget.NewSeparator(),
		stagesTitle,
		stagesText,
		widget.NewSeparator(),
		doColumns,
		widget.NewSeparator(),
		disclaimerCard,
	)
}

func makeCrossbarContent() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("The Crossbar Architecture", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// Stack diagrams vertically (not side-by-side to fit 600px scroll)
	passiveTitle := widget.NewLabelWithStyle("Passive Crossbar", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	passiveDesc := widget.NewLabel(`Ports: WL[], BL[], VDD, VSS | Cell Size: 0.46 x 2.72 um (SKY130 site)
+ Simple, dense packing | + Lower fabrication complexity
- SNEAK PATH CURRENTS | - Limited to small arrays (~32x32)`)
	passiveDesc.Wrapping = fyne.TextWrapWord

	// Wrap passive diagram in padded container
	passiveDiagram := IsometricCrossbar(3, 3, true)
	passiveDiagramContainer := container.NewPadded(passiveDiagram)
	passiveDiagramContainer.Resize(fyne.NewSize(560, 420))

	oneToneRTitle := widget.NewLabelWithStyle("1T1R (1 Transistor + 1 Resistor)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	oneToneRDesc := widget.NewLabel(`Ports: WL[], BL[], SL[], VDD, VSS | Cell Size: 0.92 x 2.72 um (2x width)
+ No sneak paths (transistor isolates) | + Scales to 128x128+ arrays
- Larger cell area (2x) | - More complex routing`)
	oneToneRDesc.Wrapping = fyne.TextWrapWord

	// Wrap 1T1R diagram in padded container
	oneToneRDiagram := Isometric1T1RCrossbar(3, 3)
	oneToneRDiagramContainer := container.NewPadded(oneToneRDiagram)
	oneToneRDiagramContainer.Resize(fyne.NewSize(560, 420))

	// Comparison table - wrapped in padded container
	comparisonTable := CellComparisonTable()
	comparisonContainer := container.NewPadded(comparisonTable)
	comparisonContainer.Resize(fyne.NewSize(440, 220))

	// Sneak path explanation (clean prose, no ASCII)
	sneakPathTitle := widget.NewLabelWithStyle("The Sneak Path Problem", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	sneakPathDesc := widget.NewLabel(`In passive arrays, reading cell (0,0) can cause unintended current paths. The intended path is WL[0] -> Cell(0,0) -> BL[0]. But sneak paths occur through other cells: WL[0] -> Cell(0,1) -> BL[1] -> Cell(1,1) -> Cell(1,0) -> BL[0]. This error grows as N^2 for NxN arrays!`)
	sneakPathDesc.Wrapping = fyne.TextWrapWord

	// Recommendation card
	recommendationCard := widget.NewCard("", "RECOMMENDATION",
		widget.NewLabel("<= 16x16: Passive | 32x32: Either (depends on accuracy needs) | >= 64x64: 1T1R required"))

	// Add spacers for better vertical separation
	spacer1 := widget.NewLabel("")
	spacer1.Resize(fyne.NewSize(1, 20))
	spacer2 := widget.NewLabel("")
	spacer2.Resize(fyne.NewSize(1, 20))
	spacer3 := widget.NewLabel("")
	spacer3.Resize(fyne.NewSize(1, 15))
	spacer4 := widget.NewLabel("")
	spacer4.Resize(fyne.NewSize(1, 15))

	return container.NewVBox(
		title,
		widget.NewSeparator(),
		passiveTitle,
		passiveDesc,
		spacer1,
		passiveDiagramContainer,
		spacer2,
		widget.NewSeparator(),
		oneToneRTitle,
		oneToneRDesc,
		spacer3,
		oneToneRDiagramContainer,
		spacer4,
		widget.NewSeparator(),
		comparisonContainer,
		widget.NewSeparator(),
		sneakPathTitle,
		sneakPathDesc,
		widget.NewSeparator(),
		recommendationCard,
	)
}

func makeFilesContent() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("EDA Files We Generate", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// 2x2 file cards grid with spacing for better readability
	lefCard := LEFPreviewCard()
	defCard := DEFPreviewCard()
	verilogCard := VerilogPreviewCard()
	libertyCard := LibertyPreviewCard()

	spacerH1 := widget.NewLabel("")
	spacerH1.Resize(fyne.NewSize(10, 1))
	spacerH2 := widget.NewLabel("")
	spacerH2.Resize(fyne.NewSize(10, 1))

	cardsRow1 := container.NewHBox(lefCard, spacerH1, defCard)

	// Vertical spacer between card rows
	spacerV := widget.NewLabel("")
	spacerV.Resize(fyne.NewSize(1, 12))

	cardsRow2 := container.NewHBox(verilogCard, spacerH2, libertyCard)

	// File purposes section (clean prose)
	purposesTitle := widget.NewLabelWithStyle("File Purposes", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	purposesText := widget.NewLabel(`LEF (Library Exchange Format) defines cell geometry - size and pin locations. This is an abstract view with no transistors.

DEF (Design Exchange Format) specifies physical placement with X,Y coordinates. The FIXED keyword prevents auto-placement.

Verilog Netlist provides a structural description of the array. Cells are black boxes (behavioral only).

Liberty (.lib) contains timing information for synthesis. WARNING: All values are PLACEHOLDERS!

OpenLane Config (JSON) points OpenLane to our custom files.

Important: LEF is abstract with no real layout. Liberty timing values need SPICE characterization. Verilog doesn't model FeFET physics. Real fabrication requires validated cells.`)
	purposesText.Wrapping = fyne.TextWrapWord

	// References section at bottom
	referencesTitle := widget.NewLabelWithStyle("References", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	refsCard := ReferencesCard()

	return container.NewVBox(
		title,
		widget.NewSeparator(),
		cardsRow1,
		spacerV,
		cardsRow2,
		widget.NewSeparator(),
		purposesTitle,
		purposesText,
		widget.NewSeparator(),
		referencesTitle,
		refsCard,
	)
}
