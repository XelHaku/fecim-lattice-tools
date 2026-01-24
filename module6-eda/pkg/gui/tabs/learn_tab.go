// pkg/gui/tabs/learn_tab.go
// Learning Center tab for FeCIM Design Suite
// Explains OpenLane flow and where the FeCIM Array Builder fits in

package tabs

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// MakeLearnTab creates the learning center tab with educational content
func MakeLearnTab(state interface{}, w fyne.Window) fyne.CanvasObject {
	// Topic selector
	topics := []string{
		"1. Overview",
		"2. OpenLane Flow",
		"3. Where We Fit In",
		"4. What We Generate",
		"5. Cell Types",
		"6. References",
	}

	topicSelector := widget.NewList(
		func() int { return len(topics) },
		func() fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(topics[id])
		},
	)
	topicSelector.OnSelected = func(id widget.ListItemID) {
		// Will be connected to content display
	}

	// Content area
	contentScroll := container.NewScroll(makeOverviewContent())
	contentScroll.SetMinSize(fyne.NewSize(600, 500))

	// Connect topic selector to content
	topicSelector.OnSelected = func(id widget.ListItemID) {
		var content fyne.CanvasObject
		switch id {
		case 0:
			content = makeOverviewContent()
		case 1:
			content = makeOpenLaneFlowContent()
		case 2:
			content = makeWhereWeFitContent()
		case 3:
			content = makeWhatWeGenerateContent()
		case 4:
			content = makeCellTypesContent()
		case 5:
			content = makeReferencesContent()
		default:
			content = makeOverviewContent()
		}
		contentScroll.Content = content
		fyne.Do(func() {
			contentScroll.Refresh()
		})
	}

	// Select first topic by default
	topicSelector.Select(0)

	// Layout with sidebar
	sidebar := container.NewBorder(
		widget.NewLabelWithStyle("Topics", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		nil, nil, nil,
		topicSelector,
	)
	sidebar.Resize(fyne.NewSize(180, 500))

	// Main layout
	split := container.NewHSplit(sidebar, contentScroll)
	split.SetOffset(0.22)

	// Header
	header := container.NewVBox(
		widget.NewLabelWithStyle("FeCIM Array Builder - Learning Center", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Understanding OpenLane and where our tool fits in"),
		widget.NewSeparator(),
	)

	return container.NewBorder(header, nil, nil, nil, split)
}

func makeOverviewContent() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("What This Tool Does", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	intro := widget.NewLabel(`Module 6 is an ARRAY BUILDER that generates EDA files
for integrating FeCIM crossbar arrays into the OpenLane flow.

WHAT WE DO:
-----------
  * Generate LEF files (cell abstracts)
  * Generate Liberty files (timing - placeholder values)
  * Generate Verilog netlists (behavioral models)
  * Generate DEF files (physical placement)
  * Export OpenLane configuration

WHAT WE DON'T DO:
-----------------
  * We do NOT provide validated FeFET device models
  * We do NOT generate production-ready layouts
  * We do NOT characterize real timing values
  * We do NOT fabricate chips

PURPOSE:
--------
This is an EDUCATIONAL tool that demonstrates how
FeCIM arrays could integrate with open-source EDA.
All timing values are placeholders - real values
require SPICE characterization with validated models.

DISCLAIMER:
-----------
This project is not affiliated with or endorsed by
external research institution, Dr. external research group, or any foundry.`)
	intro.Wrapping = fyne.TextWrapWord

	return container.NewVBox(title, widget.NewSeparator(), intro)
}

func makeOpenLaneFlowContent() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("The OpenLane RTL-to-GDSII Flow", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	content := widget.NewLabel(`OpenLane is an open-source flow that takes hardware
descriptions and produces manufacturable chip layouts.

THE COMPLETE FLOW:
==================

  STAGE 1: SYNTHESIS (Yosys)
  --------------------------
  Input:  Verilog code (behavioral description)
  Output: Gate-level netlist (standard cells)
  Tool:   Yosys

  Example: "assign out = a & b;"
        -> sky130_fd_sc_hd__and2_1

            |
            v

  STAGE 2: FLOORPLANNING
  ----------------------
  Input:  Netlist + constraints
  Output: Die area, I/O pin locations
  Tool:   OpenROAD init_floorplan

  Defines the chip boundary and where
  external connections will be placed.

            |
            v

  STAGE 3: PLACEMENT (RePlAce + OpenDP)
  -------------------------------------
  Input:  Floorplan + netlist
  Output: X,Y coordinates for every cell
  Tool:   OpenROAD place_cells

  Positions cells to minimize wire length
  while meeting timing constraints.

            |
            v

  STAGE 4: CLOCK TREE SYNTHESIS (TritonCTS)
  -----------------------------------------
  Input:  Placed design + clock constraints
  Output: Balanced clock distribution
  Tool:   OpenROAD cts

  Note: FeCIM arrays often don't need this
  (no synchronous clock in the memory array).

            |
            v

  STAGE 5: ROUTING (TritonRoute)
  ------------------------------
  Input:  Placed cells + netlist
  Output: Metal wire connections
  Tool:   OpenROAD route

  Draws actual metal paths between cells
  on multiple metal layers.

            |
            v

  STAGE 6: SIGNOFF & GDSII
  ------------------------
  Input:  Routed design
  Output: GDSII file (factory-ready)
  Tools:  Magic (DRC), Netgen (LVS)

  Verifies design rules and extracts
  the final manufacturing format.

REFERENCES:
-----------
  * OpenLane Docs: openlane.readthedocs.io
  * OpenLane Paper: WOSET 2020
  * OpenROAD: openroad.readthedocs.io`)
	content.Wrapping = fyne.TextWrapWord

	return container.NewVBox(title, widget.NewSeparator(), content)
}

func makeWhereWeFitContent() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Where the FeCIM Array Builder Fits In", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	content := widget.NewLabel(`Our tool generates files that plug into OpenLane
at specific points in the flow.

THE STANDARD OPENLANE FLOW:
===========================

  [Verilog]
      |
      v
  +-------------+
  | SYNTHESIS   |  <-- Yosys converts RTL to gates
  +-------------+
      |
      v
  +-------------+
  | FLOORPLAN   |  <-- Define die area
  +-------------+
      |
      v
  +-------------+
  | PLACEMENT   |  <-- Position cells
  +-------------+
      |
      v
  +-------------+
  | ROUTING     |  <-- Connect with wires
  +-------------+
      |
      v
  [GDSII]


WHERE WE INJECT OUR FILES:
==========================

  [Our LEF]  -----> Defines cell geometry
                    (size, pin locations)

  [Our LIB]  -----> Defines timing
                    (PLACEHOLDER values!)

  [Our Verilog] --> Used by synthesis
                    (behavioral model only)

  [Our DEF]  -----> PRE-PLACED cells
                    (skips auto-placement)

  [Our Config] ---> OpenLane settings


THE KEY INSIGHT:
================
Standard OpenLane doesn't understand crossbar arrays.
Auto-placement would scatter our cells randomly.

We provide a DEF file with FIXED positions so cells
stay in the regular grid structure that FeCIM needs:

   BL[0] BL[1] BL[2] BL[3]
     |     |     |     |
WL[0]-*-----*-----*-----*-
     |     |     |     |
WL[1]-*-----*-----*-----*-
     |     |     |     |
WL[2]-*-----*-----*-----*-
     |     |     |     |

This regular structure enables:
  * Matrix-vector multiply by Kirchhoff's law
  * Predictable IR-drop modeling
  * Uniform sneak path analysis


WHAT OPENLANE STILL DOES:
=========================
  * Routing (connecting our pre-placed cells)
  * DRC checking
  * Final GDSII generation

We provide the WHAT (cells + positions).
OpenLane figures out the HOW (wires + verification).`)
	content.Wrapping = fyne.TextWrapWord

	return container.NewVBox(title, widget.NewSeparator(), content)
}

func makeWhatWeGenerateContent() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("What Files We Generate", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	content := widget.NewLabel(`The Array Builder generates several EDA file formats.
Each serves a specific purpose in the OpenLane flow.

FILE 1: LEF (Library Exchange Format)
=====================================
Purpose: Defines cell GEOMETRY (abstract view)

  VERSION 5.8 ;
  MACRO fecim_bitcell
    CLASS CORE ;
    SIZE 0.460 BY 2.720 ;        <- Cell dimensions
    PIN WL
      DIRECTION INPUT ;
      PORT LAYER met1 ;          <- Pin on metal 1
        RECT 0 1.2 0.14 1.48 ;
      END
    END WL
    PIN BL
      DIRECTION INOUT ;
      PORT LAYER met2 ;          <- Pin on metal 2
    END BL
  END fecim_bitcell

Note: This is an ABSTRACT - no transistors drawn.
Real layout requires Magic VLSI or KLayout design.


FILE 2: Liberty (.lib) Timing Library
=====================================
Purpose: Defines cell TIMING for synthesis

  library(fecim_lib) {
    cell(fecim_bitcell) {
      pin(WL) {
        direction : input;
        capacitance : 0.002;     <- PLACEHOLDER!
      }
      timing() {
        rise_transition : 0.1;   <- PLACEHOLDER!
      }
    }
  }

WARNING: All timing values are PLACEHOLDERS.
Real characterization needs SPICE simulation
with validated FeFET device models.


FILE 3: Verilog Netlist
=======================
Purpose: Structural description for synthesis

  module fecim_array_4x4 (
    input  [3:0] WL,
    inout  [3:0] BL
  );
    fecim_bitcell cell_0_0 (.WL(WL[0]), .BL(BL[0]));
    fecim_bitcell cell_0_1 (.WL(WL[0]), .BL(BL[1]));
    // ... 16 cells total
  endmodule

Note: This is BEHAVIORAL only - cells are black boxes.
The actual FeFET physics are NOT modeled here.


FILE 4: DEF (Design Exchange Format)
====================================
Purpose: Physical PLACEMENT of cells

  VERSION 5.8 ;
  DESIGN fecim_array_4x4 ;
  UNITS DISTANCE MICRONS 1000 ;
  DIEAREA ( 0 0 ) ( 31840 40880 ) ;

  COMPONENTS 16 ;
    - cell_0_0 fecim_bitcell + FIXED ( 10000 10000 ) N ;
    - cell_0_1 fecim_bitcell + FIXED ( 10460 10000 ) N ;
    // ... with X,Y in database units (nm)
  END COMPONENTS

The FIXED keyword tells OpenLane:
"Don't move these cells - I placed them deliberately."


FILE 5: OpenLane Config
=======================
Purpose: Configure the OpenLane flow

  {
    "DESIGN_NAME": "fecim_array_4x4",
    "EXTRA_LEFS": ["fecim_bitcell.lef"],
    "EXTRA_LIBS": ["fecim_bitcell.lib"],
    "FP_DEF_TEMPLATE": "fecim_array_4x4.def"
  }

Tells OpenLane to use our custom cells and
start with our pre-placed DEF file.`)
	content.Wrapping = fyne.TextWrapWord

	return container.NewVBox(title, widget.NewSeparator(), content)
}

func makeCellTypesContent() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Cell Types: Passive vs 1T1R", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	content := widget.NewLabel(`The Array Builder supports two crossbar architectures.

PASSIVE CROSSBAR
================
Structure:
       BL[0]  BL[1]  BL[2]
         |      |      |
  WL[0]--*------*------*--
         |      |      |
  WL[1]--*------*------*--
         |      |      |

  * = FeFET device (no select transistor)

Ports: WL[], BL[], VDD, VSS
Cell Size: 0.46 x 2.72 um (SKY130 site)

Advantages:
  + Simple structure
  + Dense packing (smaller cells)
  + Lower fabrication complexity

Disadvantages:
  - SNEAK PATH CURRENTS
  - Read accuracy degrades with array size
  - Limited to small arrays (~32x32)


1T1R (1 Transistor + 1 Resistor)
================================
Structure:
       BL[0]  BL[1]  SL[0]  SL[1]
         |      |      |      |
  WL[0]--T------T------+------+--
         |      |      |      |
         R      R      |      |
         |      |      |      |
  WL[1]--T------T------+------+--
         |      |      |      |
         R      R      |      |

  T = Select transistor (gate on WL)
  R = FeFET memory element

Ports: WL[], BL[], SL[], VDD, VSS
Cell Size: 0.92 x 2.72 um (2x passive width)

Advantages:
  + No sneak paths (transistor isolates cells)
  + Scales to large arrays (128x128+)
  + Better read accuracy

Disadvantages:
  - Larger cell area (2x)
  - More complex routing (SL lines)
  - Higher fabrication complexity


THE SNEAK PATH PROBLEM
======================
In passive arrays, reading cell (0,0):

  Apply V to WL[0], ground BL[0]

  INTENDED path:  WL[0] -> Cell(0,0) -> BL[0]

  SNEAK path:     WL[0] -> Cell(0,1) -> BL[1]
                        -> Cell(1,1) -> WL[1]
                        -> Cell(1,0) -> BL[0]

This parasitic current corrupts the read signal.
Error grows as N^2 for NxN arrays.


RECOMMENDATION
==============
  Array Size    |  Recommended
  --------------|---------------
  <= 16x16      |  Passive
  32x32         |  Either (depends on accuracy needs)
  >= 64x64      |  1T1R


REFERENCES
==========
  * Sneak paths: RSC Nanoscale Advances 2020
  * 1T1R for CIM: IEEE JSSC multiple papers
  * IR-drop: Science China 2025`)
	content.Wrapping = fyne.TextWrapWord

	return container.NewVBox(title, widget.NewSeparator(), content)
}

func makeReferencesContent() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("References", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	content := widget.NewLabel(`All claims in this tool are backed by published research.

OPENLANE / EDA
==============
[1] "OpenLANE: The Open-Source Digital ASIC Flow"
    WOSET 2020
    https://woset-workshop.github.io/PDFs/2020/a21.pdf
    -> Our DEF/Verilog formats follow this spec

[2] OpenLane Documentation
    https://openlane.readthedocs.io/
    -> Configuration variables we generate

[3] LEF/DEF 5.8 Specification
    Si2/OpenAccess Coalition
    -> File format standards we follow


SKYWATER PDK
============
[4] SkyWater SKY130 Process Design Kit
    https://skywater-pdk.readthedocs.io/
    -> Cell dimensions: 0.46um site, 2.72um height
    -> Metal layer specs for pin placement

[5] SKY130 Standard Cells
    https://github.com/google/skywater-pdk
    -> LEF/Liberty format examples


FECIM DEVICE PHYSICS
====================
[6] Shin et al. "Flash In2Se3 for Neuromorphic Computing"
    Advanced Electronic Materials, 2025
    -> 30 discrete analog states

[7] Tour Group HfO2-ZrO2 Superlattice Research
    external research institution
    -> Ferroelectric switching dynamics

[8] "Roadmap on Ferroelectric Hafnia/Zirconia"
    APL Materials, 2023
    -> HfO2/ZrO2 material properties


CIM ARCHITECTURE
================
[9] "Sneak Path Solutions for Crossbar Arrays"
    RSC Nanoscale Advances, 2020
    -> Passive vs 1T1R trade-offs

[10] "Optimizing Hardware-Software Co-Design"
     Science China, 2025
     -> IR-drop and non-ideality analysis

[11] NeuroSim (Georgia Tech)
     -> Circuit-level CIM modeling methodology


DISCLAIMER
==========
This project is NOT affiliated with or endorsed by:
  * external research institution
  * Dr. external research group or Tour Lab
  * SkyWater Technology
  * Google
  * Any foundry or research institution

All references are to publicly available
published research. We implement SIMULATIONS
based on this published data, not validated
against actual hardware.

For the full reference list with DOIs and
paper links, see: docs/eda/REFERENCES.md`)
	content.Wrapping = fyne.TextWrapWord

	// Add a visual separator
	line := canvas.NewLine(theme.ForegroundColor())
	line.StrokeWidth = 1

	return container.NewVBox(title, widget.NewSeparator(), content)
}
