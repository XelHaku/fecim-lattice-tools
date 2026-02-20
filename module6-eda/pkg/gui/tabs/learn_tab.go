// pkg/gui/tabs/learn_tab.go
// Learning Center tab for FeCIM Design Suite
// Explains OpenLane flow and where the FeCIM Array Builder fits in

package tabs

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/shared/logging"
	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// sizedContainer wraps a canvas object with a minimum size requirement
// This ensures VBox/VScroll properly allocates space for the child
func sizedContainer(child fyne.CanvasObject, width, height float32) fyne.CanvasObject {
	// Create a transparent spacer rectangle that enforces minimum size
	spacer := canvas.NewRectangle(nil) // nil = transparent
	spacer.SetMinSize(fyne.NewSize(width, height))
	// Position the child at origin within the stack
	child.Move(fyne.NewPos(0, 0))
	// Stack positions the child properly - don't call Resize as the child
	// (container.NewWithoutLayout) already has its own size
	return container.NewStack(spacer, child)
}

// MakeLearnTab creates the learning center tab with educational content
func MakeLearnTab(state interface{}, w fyne.Window) fyne.CanvasObject {
	logging.GlobalDebug("[EDA-Learn] Creating Learn tab")

	// Topic data with titles and descriptions
	type topicInfo struct {
		title string
		desc  string
	}

	topicData := []topicInfo{
		{"Quick Start", "Get started in 5 steps"},
		{"What is FeCIM EDA?", "Overview & OpenLane flow"},
		{"Crossbar Architecture", "Passive / 1T1R / 2T1R"},
		{"EDA Files", "LEF, DEF, Verilog, SDC + more"},
		{"FAQ", "Troubleshooting & tips"},
	}

	topics := make([]string, len(topicData))
	for i, t := range topicData {
		topics[i] = fmt.Sprintf("%d. %s", i+1, t.title)
	}

	topicSelector := widget.NewList(
		func() int { return len(topicData) },
		func() fyne.CanvasObject {
			titleLabel := widget.NewLabelWithStyle("Title", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			descLabel := widget.NewLabel("Description")
			descLabel.TextStyle = fyne.TextStyle{Italic: true}
			return container.NewVBox(titleLabel, descLabel)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			box := obj.(*fyne.Container)
			titleLabel := box.Objects[0].(*widget.Label)
			descLabel := box.Objects[1].(*widget.Label)

			titleLabel.SetText(fmt.Sprintf("%d. %s", id+1, topicData[id].title))
			descLabel.SetText(topicData[id].desc)
		},
	)
	topicSelector.OnSelected = func(id widget.ListItemID) {
		// Will be connected to content display
	}

	// Content area - increased width for better card layout
	contentScroll := container.NewScroll(makeQuickStartContent())
	contentScroll.SetMinSize(fyne.NewSize(360, 260))

	// Connect topic selector to content
	topicSelector.OnSelected = func(id widget.ListItemID) {
		logging.GlobalDebug("[EDA-Learn] Topic selected: %d (%s)", id, topics[id])
		var content fyne.CanvasObject
		switch id {
		case 0:
			content = makeQuickStartContent()
		case 1:
			content = makeIntroContent()
		case 2:
			content = makeCrossbarContent()
		case 3:
			content = makeFilesContent()
		case 4:
			content = makeFAQContent()
		default:
			content = makeQuickStartContent()
		}
		contentScroll.Content = content
		sharedwidgets.SafeDo(func() {
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
	split.SetOffset(0.22) // Slightly narrower sidebar for cleaner look

	// Header
	title := widget.NewLabelWithStyle("FeCIM Array Builder - Learning Center", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	title.Truncation = fyne.TextTruncateEllipsis
	aboutScienceBtn := sharedwidgets.CreateAboutScienceButton(w)
	aboutScienceBtn.Importance = widget.LowImportance
	subTitle := widget.NewLabel("Understanding LibreLane/OpenLane and where our tool fits in")
	subTitle.Wrapping = fyne.TextWrapWord
	header := container.NewVBox(
		container.NewBorder(nil, nil, nil, aboutScienceBtn, title),
		subTitle,
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

	intro := widget.NewLabel(`Module 6 is an ARRAY BUILDER that generates EDA files for assembling FeCIM crossbar arrays within the OpenLane flow. It automates the placement (DEF) and connectivity (Verilog) of pre-existing FeFET cell macros. It does NOT draw transistors from scratch.`)
	intro.Wrapping = fyne.TextWrapWord
	introCard := widget.NewCard("Overview", "", intro)

	// Operation modes visual
	modesContainer := sizedContainer(OperationModesVisual(), 620, 230)
	modesCard := widget.NewCard("FeCIM Operation Modes", "", modesContainer)

	// OpenLane flow diagram
	flowContainer := sizedContainer(OpenLaneFlowDiagram(), 760, 290)
	flowCard := widget.NewCard("OpenLane RTL-to-GDSII Assembly", "", flowContainer)

	// Stages Explained section
	stagesText := widget.NewLabel(`1. SYNTHESIS (Yosys) - Converts behavioral Verilog to gate-level netlist
2. FLOORPLAN - Defines die area and I/O pin locations
3. PLACEMENT (OpenROAD) - Assigns X,Y coordinates to every cell
4. CTS (Clock Tree Synthesis) - Distributes clock signal evenly (FeCIM arrays often skip this)
5. ROUTING (OpenROAD TritonRoute) - Draws metal wire connections
6. SIGNOFF & GDSII - Assembly of pre-existing cell macros into final GDSII`)
	stagesText.Wrapping = fyne.TextWrapWord
	stagesCard := widget.NewCard("The Stages Explained", "", stagesText)

	// Two-column do/don't layout
	doList := makeBulletList("",
		"Generate LEF files (cell abstracts)",
		"Generate Liberty files (timing - placeholder values)",
		"Generate Verilog netlists (behavioral models)",
		"Generate DEF files (physical placement)",
		"Generate SPICE subcircuit netlists (ngspice/HSPICE compatible)",
		"Generate simulation scripts (CrossSim YAML/Python, PySpice, OpenVAF Verilog-A)",
		"Export LibreLane/OpenLane config + flow scripts (run_flow.sh, synthesis.tcl, ...)",
		"Generate design_summary.txt + constraints.sdc + CSV cell assignments")

	dontList := makeBulletList("",
		"We do NOT provide validated FeFET device models",
		"We do NOT generate GDSII geometry from scratch",
		"We do NOT characterize real timing values",
		"We do NOT fabricate chips")

	doCard := widget.NewCard("✓ WHAT WE DO", "", doList)
	dontCard := widget.NewCard("✗ WHAT WE DON'T DO", "", dontList)
	doColumns := container.NewGridWithColumns(2, doCard, dontCard)

	// Disclaimer banner
	disclaimerCard := widget.NewCard("⚠️ DISCLAIMER", "",
		widget.NewLabel("This project is not affiliated with or endorsed by external research institution, Dr. external research group, or any foundry."))

	return container.NewVBox(
		title,
		widget.NewSeparator(),
		introCard,
		modesCard,
		flowCard,
		stagesCard,
		doColumns,
		disclaimerCard,
	)
}

func makeCrossbarContent() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("The Crossbar Architecture", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// Stack diagrams vertically (not side-by-side to fit 600px scroll)
	passiveTitle := widget.NewLabelWithStyle("Passive Crossbar", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	passiveDesc := widget.NewLabel(`Ports: WL[], BL[], VPWR, VGND | Cell Size: 0.46 x 2.72 um (SKY130 site)
+ Simple, dense packing | + Lower fabrication complexity
- SNEAK PATH CURRENTS | - Limited to small arrays (~32x32)`)
	passiveDesc.Wrapping = fyne.TextWrapWord

	// Passive diagram with enforced minimum size
	// For 3x3: rightX ~380, bottomY ~260, legend adds ~60, total ~380x380
	passiveDiagramContainer := sizedContainer(IsometricCrossbar(3, 3, true), 450, 400)

	oneToneRTitle := widget.NewLabelWithStyle("1T1R (1 Transistor + 1 Resistor)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	oneToneRDesc := widget.NewLabel(`Ports: WL[], BL[], SL[], VPWR, VGND | Cell Size: 0.92 x 4.07 µm (2x width, hvl height)
+ No sneak paths (transistor isolates) | + Scales to 128x128+ arrays
- Larger cell area (2x) | - More complex routing`)
	oneToneRDesc.Wrapping = fyne.TextWrapWord

	// 1T1R diagram with enforced minimum size
	// Similar to passive: ~450x400
	oneToneRDiagramContainer := sizedContainer(Isometric1T1RCrossbar(3, 3), 450, 400)

	twoT1RTitle := widget.NewLabelWithStyle("2T1R (2 Transistors + 1 Resistor)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	twoT1RDesc := widget.NewLabel(`Ports: WL[], CSL[], BL[], SL[], VPWR, VGND | Cell Size: ~1.38 x 4.07 µm (3x width)
+ ZERO sneak paths (both row AND column transistors isolate each cell)
+ Scales to 512x512+ arrays | - Highest area and routing complexity`)
	twoT1RDesc.Wrapping = fyne.TextWrapWord

	twoT1RDiagramContainer := sizedContainer(Isometric2T1RCrossbar(3, 3), 500, 460)

	// Comparison table with enforced minimum size
	// Table: colWidths sum to 480, 7 rows * 28 = 196, plus margins ~500x216
	comparisonContainer := sizedContainer(CellComparisonTable(), 500, 220)

	// Sneak path explanation (clean prose, no ASCII)
	sneakPathTitle := widget.NewLabelWithStyle("The Sneak Path Problem", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	sneakPathDesc := widget.NewLabel(`In passive arrays, reading cell (0,0) can cause unintended current paths. The intended path is WL[0] -> Cell(0,0) -> BL[0]. But sneak paths occur through other cells: WL[0] -> Cell(0,1) -> BL[1] -> Cell(1,1) -> Cell(1,0) -> BL[0]. This error grows as N^2 for NxN arrays!`)
	sneakPathDesc.Wrapping = fyne.TextWrapWord

	// Recommendation card
	recommendationCard := widget.NewCard("", "RECOMMENDATION",
		widget.NewLabel("≤16×16: Passive | 32×32: Either | ≥64×64: 1T1R | ≥256×256: 2T1R (row + column isolation)"))

	// Add spacers for better vertical separation
	spacer1 := widget.NewLabel("")
	spacer1.Resize(fyne.NewSize(1, 20))
	spacer2 := widget.NewLabel("")
	spacer2.Resize(fyne.NewSize(1, 20))
	spacer3 := widget.NewLabel("")
	spacer3.Resize(fyne.NewSize(1, 15))
	spacer4 := widget.NewLabel("")
	spacer4.Resize(fyne.NewSize(1, 15))
	spacer5 := widget.NewLabel("")
	spacer5.Resize(fyne.NewSize(1, 15))

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
		twoT1RTitle,
		twoT1RDesc,
		spacer5,
		twoT1RDiagramContainer,
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

	// File cards - using widget.Card which reports MinSize properly
	cardsGrid := container.NewAdaptiveGrid(2,
		LEFPreviewCard(),
		DEFPreviewCard(),
		VerilogPreviewCard(),
		LibertyPreviewCard(),
	)
	filePreviewsCard := widget.NewCard("File Format Examples", "", cardsGrid)

	// === Collapsible sections using Accordion ===

	// Section 1: How We Generate Files
	genText := widget.NewLabel(`VERILOG GENERATION:
• Loop through array dimensions (Rows × Cols)
• Instantiate cell macros: fecim_bitcell (passive), fecim_1t1r_bitcell (1T1R), or fecim_2t1r_bitcell (2T1R)
• Connect WL[row] to each cell's WL pin
• Connect BL[col] to each cell's BL pin
• For 1T1R: also connect SL[col] to each cell's SL pin
• For 2T1R: also connect CSL[col] (column transistor gate) and SL[col] to each cell's pins

DEF GENERATION:
• Calculate die area from cell dimensions + margins
• Generate SITE definition matching the LEF
• Generate ROW definitions (required for placement validation)
• Place each cell at calculated (X, Y) coordinates with FIXED keyword
• Declare all pins: WL[], BL[], SL[] (1T1R/2T1R), CSL[] (2T1R only), VPWR, VGND

LEF GENERATION:
• Define LAYER (met1) for pin geometries
• Define SITE with cell dimensions
• Define MACRO with pin locations and obstruction areas

SPICE GENERATION:
• Write provenance header + mode/array metadata
• Emit LK FeCap subcircuit (FECAP_HZO) — parameters from DefaultHzoFeFETMaterial()
• Emit fefet_cell subcircuit: programmable resistor R_level in series with FECAP_HZO
• For 1T1R: emit SKY130 NMOS model card + fefet_1t1r (row NMOS selector → fefet_cell)
• For 2T1R: emit fefet_2t1r (row NMOS M_ROW + column NMOS M_COL in series → fefet_cell)
• Instantiate active cells: X{row}_{col} with WL[row]/BL[col] connections
• Output: ngspice/HSPICE-compatible .sp file; run via "ngspice design.sp"`)
	genText.Wrapping = fyne.TextWrapWord

	// Section 2: How We Validate
	valText := widget.NewLabel(`YOSYS VALIDATION (Verilog Syntax):
• Runs: yosys -p "read_verilog cell.v array.v"
• Checks syntax errors, undefined modules, port mismatches
• PASS = Verilog is syntactically correct

DEF VALIDATION (Placement Syntax):
• Parses DEF file structure
• Verifies COMPONENTS match expected count
• Checks coordinate format and keywords

CROSS-CHECK VALIDATION:
• Compares cell names across LEF, LIB, and Verilog
• Verifies pin names match between all three files
• Ensures consistency for OpenLane integration

OPENLANE PLACEMENT VALIDATION (via Docker):
• Runs OpenROAD: read_lef, read_def, check_placement
• Verifies cells are placed within die boundaries
• Checks ROW definitions match SITE definitions
• Detects overlapping cells or DRC violations`)
	valText.Wrapping = fyne.TextWrapWord

	// Section 3: Layout Visualization
	imgText := widget.NewLabel(`Layout images are generated using industry-standard EDA tools:

OUR APP USES KLAYOUT (via Docker):
• Automatically invoked when Docker + OpenLane image available
• Reads our generated LEF (cell geometry) and DEF (placement)
• Exports PNG layout image to data/
• Falls back to schematic SVG if KLayout unavailable

MANUAL KLAYOUT:
  klayout -l tech.lyp design.def -o layout.png

USING MAGIC:
  magic -dnull -noconsole << EOF
    def read design.def
    load design
    plot svg layout.svg
  EOF

USING OPENROAD GUI:
  openroad -gui (Interactive placement viewer)`)
	imgText.Wrapping = fyne.TextWrapWord

	// Section 4: File Format Summary
	purposesText := widget.NewLabel(`EDA LAYOUT / PHYSICAL:
  LEF          : Cell geometry abstract (pin locations, obstruction area)
  DEF          : Physical placement with FIXED X,Y coordinates
  SVG          : Schematic layout diagram (browser/Inkscape viewable)

RTL / NETLIST:
  Verilog (.v) : Structural netlist — cell + array-level modules
  SPICE (.sp)  : SPICE subcircuit netlist (ngspice/HSPICE compatible)
  Liberty (.lib): Timing/power library (PLACEHOLDER values — not validated!)
  SDC          : Synthesis timing constraints

SIMULATION / ML:
  CrossSim YAML: CrossSim simulator configuration
  CrossSim .py : CrossSim Python import script
  PySpice .py  : PySpice Python circuit builder script

DESIGN DATA:
  design_summary.txt: Human-readable area, electrical, and timing report
  config.json  : LibreLane/OpenLane v2+ configuration (with MACROS dict)
  config.tcl   : OpenLane v1 TCL configuration (set ::env(...) format)
  macros.cfg   : OpenLane v1 macro placement constraints
  {design}.json: Machine-readable design metadata

VALIDATION SCRIPTS:
  run_drc.sh   : Magic VLSI DRC runner (bash run_drc.sh)
  run_lvs.sh   : Netgen LVS runner — compares SPICE netlist vs layout (bash run_lvs.sh)

FLOW SCRIPTS:
  synthesis.tcl      : Yosys hierarchy check / synthesis script
  openroad_flow.tcl  : OpenROAD placement check + timing TCL script
  opensta_check.tcl  : OpenSTA standalone timing analysis
  gen_gds.py         : KLayout DEF+LEF → GDS II export and PNG rendering
  run_flow.sh        : One-shot Yosys+KLayout+OpenROAD+LibreLane runner
  CSV                : Tab-delimited cell conductance assignment table

⚠️ WARNING: Liberty timing values are placeholders. Real fabrication requires SPICE characterization with validated FeFET models.`)
	purposesText.Wrapping = fyne.TextWrapWord

	// Create accordion with all sections
	accordion := widget.NewAccordion(
		widget.NewAccordionItem("1. How We Generate Files", genText),
		widget.NewAccordionItem("2. How We Validate", valText),
		widget.NewAccordionItem("3. Layout Visualization", imgText),
		widget.NewAccordionItem("4. File Format Summary", purposesText),
	)
	// Open first section by default
	accordion.Open(0)

	// References section
	refsCard := widget.NewCard("References", "", ReferencesCard())

	return container.NewVBox(
		title,
		widget.NewSeparator(),
		filePreviewsCard,
		widget.NewSeparator(),
		accordion,
		widget.NewSeparator(),
		refsCard,
	)
}

func makeQuickStartContent() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Quick Start Guide", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	step1 := widget.NewCard("Step 1: Configure Array", "",
		widget.NewLabel("Set Rows, Cols, Mode (storage/memory/compute), and Architecture (passive/1T1R/2T1R) in the Builder tab."))

	step2 := widget.NewCard("Step 2: Generate Files", "",
		widget.NewLabel("Click 'Generate All' to create Verilog, DEF, LEF, and Liberty files. Files are saved to the data/ directory."))

	step3 := widget.NewCard("Step 3: Validate", "",
		widget.NewLabel("Click 'Validate All' to run Yosys syntax check, DEF validation, and cross-check. Green checkmarks indicate success."))

	step4 := widget.NewCard("Step 4: Export Package", "",
		widget.NewLabel("Click 'Export Package' to bundle all files for LibreLane/OpenLane integration. The package includes LEF/LIB/V, design_summary.txt, config.json + constraints.sdc, flow scripts (synthesis.tcl, openroad_flow.tcl, gen_gds.py, run_flow.sh), {design}.json metadata, and README.\n\nTo run the full flow:\n  cd data/{design}/\n  bash run_flow.sh"))

	step5 := widget.NewCard("Step 5: View Layout", "",
		widget.NewLabel("Use the Layout tab to view generated images from KLayout, OpenROAD, or Yosys. Zoom controls let you inspect details."))

	tipCard := widget.NewCard("💡 Tips", "",
		widget.NewLabel("• Start with a small array (4x4) to verify workflow\n• Use passive architecture for arrays ≤16x16\n• Check validation log for detailed error messages\n• Docker required for KLayout/OpenROAD image generation\n• Tab 5 (Flow Scripts) previews all generated scripts before exporting"))

	return container.NewVBox(
		title,
		widget.NewSeparator(),
		step1, step2, step3, step4, step5,
		widget.NewSeparator(),
		tipCard,
	)
}

func makeFAQContent() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("FAQ & Troubleshooting", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	faq1 := widget.NewCard("Q: Why do validations fail?", "",
		widget.NewLabel("A: Common causes:\n• Missing cell files (run Generate All first)\n• Docker not running (needed for Yosys/OpenROAD)\n• Invalid array dimensions (must be > 0)"))

	faq2 := widget.NewCard("Q: Why are images not generated?", "",
		widget.NewLabel("A: Image generation requires Docker with OpenLane image. Run 'docker pull ghcr.io/the-openroad-project/openlane:latest' or click 'Pull OpenLane Image' button if shown."))

	faq3 := widget.NewCard("Q: What's the difference between passive and 1T1R?", "",
		widget.NewLabel("A: Passive arrays are simpler but suffer from sneak path currents in larger arrays (risk grows as N×M). Literature recommends passive up to ~32×32; beyond that sneak currents typically dominate. 1T1R adds one access transistor per row to suppress row-direction sneak paths, enabling arrays to 128×128+. 2T1R adds a second transistor per column for full isolation, enabling up to ~512×512."))

	faq4 := widget.NewCard("Q: Are Liberty timing values accurate?", "",
		widget.NewLabel("A: NO! Liberty values are placeholders. Real fabrication requires SPICE characterization with validated FeFET device models from a foundry."))

	faq5 := widget.NewCard("Q: How do I use the generated files with LibreLane/OpenLane?", "",
		widget.NewLabel("A: Export Package creates a ready-to-use directory with run_flow.sh.\n\nWith the bundled script (recommended):\n  cd <export_dir>/\n  bash run_flow.sh\n\nWith LibreLane directly:\n  librelane config.json\n\nWith OpenLane (Docker):\n  flow.tcl -design <your_design_name>"))

	faq6 := widget.NewCard("Q: What validations does this tool NOT perform?", "",
		widget.NewLabel("A: The following checks are outside scope — you need an external EDA flow for these:\n"+
			"• Physical DRC (Design Rule Check) — requires foundry PDK + KLayout/Magic\n"+
			"• Parasitic extraction (RC) — requires Magic/OpenRCX; affects timing\n"+
			"• Static timing analysis signoff — requires real Liberty files from SPICE char.\n"+
			"• SPICE-level simulation — requires validated FeFET SPICE model from foundry\n"+
			"• ESD / latch-up checks — requires PDK-specific LVS runsets\n"+
			"• Power integrity (IR drop) — requires post-layout simulation with parasitics\n"+
			"\nGenerated Liberty, LEF, and SDC files are structural templates, not signoff data."))

	faq7 := widget.NewCard("Q: Which PDKs (process design kits) are supported?", "",
		widget.NewLabel("A: Three open PDKs are supported via --tech flag:\n"+
			"  • SKY130   — SkyWater 130 nm (default). Most mature open PDK; sky130_fd_sc_hd site.\n"+
			"  • GF180MCU — GlobalFoundries 180 nm MCU. Larger cells; gf180mcu site.\n"+
			"  • IHP_SG13G2 — IHP 130 nm SiGe BiCMOS. For RF/high-speed applications.\n\n"+
			"Note: This tool generates structural LEF/DEF/Verilog templated for each PDK's site dimensions.\n"+
			"No foundry PDK files are bundled — you need to install the PDK separately for real tape-out.\n"+
			"  • SKY130:    pip install open_pdk  (or volare)\n"+
			"  • GF180MCU:  pip install open_pdk\n"+
			"  • IHP_SG13G2: https://github.com/IHP-GmbH/IHP-Open-PDK"))

	faq8 := widget.NewCard("Q: How do I use the CLI (headless/batch mode)?", "",
		widget.NewLabel("A: The CLI is at cmd/eda-cli/main.go. Build and run:\n"+
			"  go build -o fecim-eda-cli ./module6-eda/cmd/eda-cli\n"+
			"  ./fecim-eda-cli [flags]\n\n"+
			"Key flags:\n"+
			"  --mode     storage | memory | compute (default: compute)\n"+
			"  --arch     passive | 1t1r | 2t1r      (default: passive)\n"+
			"  --tech     SKY130 | GF180MCU | IHP_SG13G2 (default: SKY130)\n"+
			"  --rows N   Array rows  (default: 128)\n"+
			"  --cols N   Array cols  (default: 128)\n"+
			"  --levels N Conductance levels (default: 30)\n"+
			"  --output DIR  Output directory (default: data/)\n"+
			"  --vdd V    Supply voltage (default: 1.8V)\n"+
			"  --gmin μS  Min conductance (default: 10.0 μS)\n"+
			"  --gmax μS  Max conductance (default: 100.0 μS)\n"+
			"  --scripts  Export flow scripts (Yosys, OpenROAD, KLayout, SDC, LibreLane)\n"+
			"  --quiet    Suppress informational output\n\n"+
			"Note: Verilog, DEF, SPICE, JSON, and CSV are exported by default.\n"+
			"Use --scripts=false or --spice=false to suppress specific outputs.\n\n"+
			"Example — 64×64 1T1R compute array with flow scripts:\n"+
			"  ./fecim-eda-cli --arch 1t1r --rows 64 --cols 64 --scripts --output results/"))

	troubleCard := widget.NewCard("🔧 Troubleshooting", "",
		widget.NewLabel("• 'Docker not available': Install Docker Desktop and ensure daemon is running\n• 'Yosys validation failed': Check Verilog syntax in the log output\n• 'DEF validation failed': Ensure cell dimensions match LEF\n• 'Cross-check failed': Regenerate all files to ensure consistency"))

	return container.NewVBox(
		title,
		widget.NewSeparator(),
		faq1, faq2, faq3, faq4, faq5, faq6, faq7, faq8,
		widget.NewSeparator(),
		troubleCard,
	)
}
