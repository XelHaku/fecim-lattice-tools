// Package help provides embedded help topics for the FeCIM application.
package help

// registerEmbeddedTopics registers all built-in help content.
func registerEmbeddedTopics(hs *HelpSystem) {
	// Overview / Getting Started
	hs.RegisterTopic(&HelpTopic{
		ID:       "overview",
		Title:    "Welcome to FeCIM Lattice Tools",
		Summary:  "Getting started with the FeCIM visualization and simulation suite",
		Category: "Getting Started",
		Keywords: []string{"start", "begin", "introduction", "welcome", "overview"},
		ContextKeys: []string{"home", "launcher"},
		Content: `# Welcome to FeCIM Lattice Tools

FeCIM Lattice Tools is an interactive educational platform for exploring Ferroelectric Compute-in-Memory technology.

## Quick Start

1. **Navigate between modules** using the tabs at the top or the Home screen cards
2. **Press F1** at any time for context-sensitive help
3. **Press Shift+F1** to open the full help browser
4. **Use Ctrl+K** in the Documentation tab to search docs

## The Six Modules

| Module | What You'll Learn |
|--------|-------------------|
| **Hysteresis** | How ferroelectric memory cells store data |
| **Crossbar** | How arrays compute matrix operations |
| **MNIST** | Real neural network inference on hardware |
| **Circuits** | Peripheral circuits (DAC, ADC, drivers) |
| **Comparison** | Why FeCIM beats traditional computing |
| **EDA** | Chip design workflow integration |

## Keyboard Shortcuts

- **F1** - Contextual help for current screen
- **Shift+F1** - Open help browser
- **Ctrl+K** - Search documentation (in Docs tab)
- **Arrow keys** - Navigate controls
- **Space/Enter** - Activate buttons

## Tips

- Hover over sliders and controls to see tooltips
- Look for the ℹ️ icons for inline explanations
- The glossary (right sidebar) explains technical terms
`,
	})

	// Hysteresis Module
	hs.RegisterTopic(&HelpTopic{
		ID:       "module.hysteresis",
		Title:    "Hysteresis Module",
		Summary:  "Understanding ferroelectric memory cells and the P-E hysteresis loop",
		Category: "Modules",
		Module:   "hysteresis",
		Keywords: []string{"hysteresis", "polarization", "ferroelectric", "memory", "loop", "P-E curve", "Preisach"},
		ContextKeys: []string{"hysteresis", "FeCIM Hysteresis Simulation"},
		Content: `# Hysteresis Module

This module visualizes how ferroelectric materials store information through polarization states.

## What You'll See

### The P-E Loop
The main visualization shows the Polarization (P) vs. Electric Field (E) hysteresis loop:
- **Positive saturation**: Maximum positive polarization (Ps+)
- **Negative saturation**: Maximum negative polarization (Ps-)
- **Coercive field (Ec)**: Field needed to switch polarization
- **Remnant polarization (Pr)**: Polarization remaining after field removal

### Controls

**Material Selection**
Choose from different ferroelectric materials (HZO, PZT, BTO) to see how material properties affect the loop shape.

**Temperature**
Adjust temperature to observe how thermal effects influence polarization behavior. Higher temperatures reduce Pr.

**Voltage Sweep**
Control the applied voltage to trace the hysteresis loop manually or use auto-play mode.

**Animation Speed**
Adjust how fast the simulation runs during auto-play.

## Key Concepts

### Why Hysteresis Matters for Memory
The hysteresis loop enables non-volatile memory:
1. Apply positive field → positive polarization state ("1")
2. Remove field → polarization remains (memory retention)
3. Apply negative field → negative polarization state ("0")
4. Remove field → polarization remains

### Analog States
Unlike binary memory, ferroelectric cells can store multiple levels:
- Partial polarization = intermediate states
- Enables higher density storage
- Used for analog compute-in-memory
`,
	})

	// Crossbar Module
	hs.RegisterTopic(&HelpTopic{
		ID:       "module.crossbar",
		Title:    "Crossbar Module",
		Summary:  "Visualizing matrix-vector multiplication in crossbar arrays",
		Category: "Modules",
		Module:   "crossbar",
		Keywords: []string{"crossbar", "array", "MVM", "matrix", "vector", "multiply", "conductance"},
		ContextKeys: []string{"crossbar", "FeCIM Crossbar Array Visualization"},
		Content: `# Crossbar Module

This module demonstrates how crossbar arrays perform matrix-vector multiplication (MVM) in a single step using analog physics.

## How It Works

### The Crossbar Structure
- **Rows (wordlines)**: Carry input voltages
- **Columns (bitlines)**: Collect output currents
- **Cells**: Ferroelectric devices with programmable conductance

### The Math
At each crossing point:
- Current = Voltage × Conductance (Ohm's law)
- Column sums all currents (Kirchhoff's law)
- Result: I = G × V (matrix-vector multiply!)

## Controls

**Array Size**
Set the dimensions of the crossbar (e.g., 8×8, 16×16, 32×32).

**Input Vector**
Configure the voltage values applied to wordlines.

**Weight Matrix**
Set the conductance values representing neural network weights.

**Non-Idealities**
Enable realistic effects:
- **IR Drop**: Voltage loss in long wires
- **Sneak Paths**: Unintended current paths
- **Noise**: Device-to-device variation
- **Quantization**: Limited precision levels

## Visualization

**Heatmap View**
Shows conductance values across the array. Brighter = higher conductance.

**Current Flow**
Arrows indicate current direction and magnitude through cells.

**Output Vector**
Bar chart showing computed output currents.

## Why This Matters

Traditional digital computing requires:
- Fetch weights from memory
- Multiply each weight × input
- Accumulate results
- Many clock cycles

Crossbar computing:
- Weights stored in-place
- All multiplications happen simultaneously
- Accumulation is automatic (physics!)
- Single step = massive parallelism
`,
	})

	// MNIST Module
	hs.RegisterTopic(&HelpTopic{
		ID:       "module.mnist",
		Title:    "MNIST Neural Network",
		Summary:  "Real neural network inference comparing floating-point vs CIM hardware",
		Category: "Modules",
		Module:   "mnist",
		Keywords: []string{"MNIST", "neural", "network", "inference", "AI", "digits", "recognition", "accuracy"},
		ContextKeys: []string{"mnist", "FeCIM MNIST Neural Network"},
		Content: `# MNIST Neural Network Module

This module demonstrates real neural network inference on handwritten digit recognition, comparing ideal floating-point computation with realistic FeCIM hardware simulation.

## What's MNIST?

MNIST is the "Hello World" of machine learning:
- 70,000 images of handwritten digits (0-9)
- 28×28 pixel grayscale images
- Standard benchmark for neural network performance

## The Dual-Mode Display

### Floating-Point (Left)
- Ideal mathematical computation
- Perfect precision
- Reference for comparison

### CIM Hardware (Right)
- Realistic hardware simulation
- Includes all non-idealities
- Shows actual deployment behavior

## Controls

**Draw Your Own Digit**
Use the canvas to draw a digit and see real-time classification.

**Sample from Dataset**
Load random test images from the MNIST dataset.

**Network Architecture**
View the neural network layer structure.

**Quantization**
Adjust weight precision (4-bit, 8-bit, etc.) to see accuracy impact.

**Non-Ideality Sliders**
- **Device Variation**: Random conductance differences
- **Read Noise**: Measurement uncertainty
- **IR Drop**: Wire resistance effects

## Understanding Results

**Confidence Bars**
Show the network's certainty for each digit class (0-9).

**Accuracy Metrics**
- Top-1 accuracy: Correct on first guess
- Comparison: FP vs CIM accuracy gap

**Energy Comparison**
Shows the energy efficiency advantage of CIM over digital.

## Key Insight

Even with hardware non-idealities, FeCIM maintains high accuracy while using orders of magnitude less energy than digital computation.
`,
	})

	// Circuits Module
	hs.RegisterTopic(&HelpTopic{
		ID:       "module.circuits",
		Title:    "Peripheral Circuits",
		Summary:  "Understanding DACs, ADCs, and driver circuits for FeCIM systems",
		Category: "Modules",
		Module:   "circuits",
		Keywords: []string{"circuits", "DAC", "ADC", "driver", "peripheral", "analog", "digital", "converter"},
		ContextKeys: []string{"circuits", "FeCIM Peripheral Circuits Visualizer"},
		Content: `# Peripheral Circuits Module

This module explores the supporting circuits that interface digital systems with analog FeCIM arrays.

## Circuit Types

### Digital-to-Analog Converters (DAC)
Convert digital input values to analog voltages for crossbar wordlines.

**Key Parameters:**
- Resolution (bits): Precision of voltage levels
- Speed: Conversion rate
- Power: Energy consumption
- INL/DNL: Linearity errors

### Analog-to-Digital Converters (ADC)
Convert analog output currents to digital values for further processing.

**Types Shown:**
- Flash ADC: Fast but power-hungry
- SAR ADC: Balanced speed/power
- Sigma-Delta: High precision, slower

### Wordline Drivers
Buffer and amplify signals to drive crossbar rows.

**Considerations:**
- Drive strength vs. power
- Slew rate for fast switching
- Voltage levels for programming

### Bitline Sense Amplifiers
Detect small currents from crossbar columns.

**Types:**
- Current-mode sense amps
- Transimpedance amplifiers
- Integrating ADCs

## Interactive Features

**Timing Diagrams**
Visualize signal propagation through the circuit chain.

**Parameter Sweeps**
See how changing resolution or speed affects system performance.

**Power Breakdown**
Pie chart showing where power is consumed.

## Design Trade-offs

The module illustrates key trade-offs:
- Higher precision → more power
- Faster conversion → more power
- Lower power → reduced accuracy
- Finding the sweet spot for your application
`,
	})

	// Comparison Module
	hs.RegisterTopic(&HelpTopic{
		ID:       "module.comparison",
		Title:    "Technology Comparison",
		Summary:  "Why FeCIM outperforms traditional computing for AI workloads",
		Category: "Modules",
		Module:   "comparison",
		Keywords: []string{"comparison", "energy", "efficiency", "GPU", "CPU", "TOPS", "performance"},
		ContextKeys: []string{"comparison", "FeCIM: The Energy Revolution"},
		Content: `# Technology Comparison Module

This module presents the compelling case for FeCIM technology through direct comparison with conventional computing.

## The Energy Problem

### Von Neumann Bottleneck
Traditional computers separate memory and processing:
- Data must travel between memory and CPU/GPU
- This "data movement" consumes most of the energy
- Bandwidth limits performance

### The AI Explosion
- AI models are doubling in size every 3.4 months
- Energy consumption growing unsustainably
- Data centers consume 1-2% of global electricity

## FeCIM Advantages

### In-Memory Computing
- Computation happens where data lives
- No memory-processor data shuttling
- Massive energy savings

### Analog Multiply-Accumulate
- Physics does the math (Ohm's law)
- Thousands of operations in one step
- No clock cycles wasted

### Energy Efficiency
- 10-1000× better TOPS/W than GPUs
- Enables edge AI without batteries
- Sustainable AI infrastructure

## Interactive Comparisons

**Workload Selector**
Choose AI tasks to compare:
- Image classification
- Language models
- Recommendation systems

**Metrics Dashboard**
- TOPS (Tera Operations Per Second)
- TOPS/W (Energy efficiency)
- Latency comparison
- Total cost of ownership

**Scaling Projections**
See how efficiency gap grows with model size.

## The Bottom Line

FeCIM isn't just incrementally better—it's a paradigm shift that makes previously impossible AI applications practical.
`,
	})

	// EDA Module
	hs.RegisterTopic(&HelpTopic{
		ID:       "module.eda",
		Title:    "EDA Design Suite",
		Summary:  "Integration with open-source chip design tools",
		Category: "Modules",
		Module:   "eda",
		Keywords: []string{"EDA", "OpenLane", "PDK", "chip", "design", "GDSII", "layout", "synthesis"},
		ContextKeys: []string{"eda", "FeCIM EDA Design Suite (Work In Progress)"},
		Content: `# EDA Design Suite

This module bridges educational concepts with real chip design workflows using open-source EDA (Electronic Design Automation) tools.

## Overview

The EDA suite integrates with:
- **OpenLane**: Automated RTL-to-GDSII flow
- **SkyWater PDK**: Open-source 130nm process
- **Magic**: Layout viewer and DRC
- **ngspice**: Circuit simulation

## Workflow Stages

### 1. Design Entry
- Verilog RTL for digital logic
- SPICE netlists for analog blocks
- FeCIM macro definitions

### 2. Synthesis
- Convert RTL to gate-level netlist
- Technology mapping to PDK cells
- Timing analysis

### 3. Placement & Routing
- Physical placement of cells
- Metal interconnect routing
- Clock tree synthesis

### 4. Verification
- Design Rule Check (DRC)
- Layout vs. Schematic (LVS)
- Timing signoff

### 5. Export
- Generate GDSII for fabrication
- Create documentation
- Generate reports

## Interactive Features

**Flow Visualization**
See the chip design flow step-by-step.

**Example Designs**
Pre-built FeCIM macro examples to explore.

**Tool Integration**
Launch external tools from within the app.

## Note

This module is marked "Work In Progress" as full tool integration depends on external tool installation and configuration.
`,
	})

	// Documentation Module
	hs.RegisterTopic(&HelpTopic{
		ID:       "module.docs",
		Title:    "Documentation Browser",
		Summary:  "Navigating the built-in documentation and curriculum",
		Category: "Modules",
		Module:   "docs",
		Keywords: []string{"documentation", "docs", "curriculum", "search", "glossary"},
		ContextKeys: []string{"docs", "Documentation"},
		Content: `# Documentation Browser

The Documentation module provides access to the complete FeCIM curriculum and reference materials.

## Navigation

### File Tree (Left Panel)
Browse documentation organized by module and topic.

### Content Area (Center)
Read markdown-formatted documentation with:
- Syntax highlighting
- Embedded images
- Clickable glossary terms
- Table of contents

### Quick Access (Right Panel)
Jump to sections within the current document.

## Search

**Keyboard Shortcut: Ctrl+K**

The search feature provides:
- Full-text search across all docs
- Fuzzy matching for typos
- Category filtering
- Relevance ranking

## Glossary Integration

Technical terms are automatically highlighted and clickable:
- Click any highlighted term for its definition
- Definitions include context and related terms
- Learn vocabulary as you read

## Curriculum Structure

### Per-Module Documentation
Each module has:
- **ELI5.md**: Explain Like I'm 5 - simple explanations
- **Physics.md**: Deep technical details
- **Features.md**: User guide for the module
- **OpenSource-Tools.md**: Related external tools

### Research Papers
Curated collection of key publications with summaries.

## Tips

- Use breadcrumbs to navigate up the hierarchy
- Star frequently-used documents for quick access
- The reading time estimate helps plan your study
`,
	})

	// Keyboard shortcuts reference
	hs.RegisterTopic(&HelpTopic{
		ID:       "shortcuts",
		Title:    "Keyboard Shortcuts",
		Summary:  "Complete list of keyboard shortcuts",
		Category: "Reference",
		Keywords: []string{"keyboard", "shortcuts", "keys", "hotkeys", "accelerators"},
		Content: `# Keyboard Shortcuts

## Global Shortcuts

| Key | Action |
|-----|--------|
| F1 | Context-sensitive help |
| Shift+F1 | Open help browser |
| Ctrl+K | Search (in Docs tab) |
| Escape | Close dialogs/cancel |

## Navigation

| Key | Action |
|-----|--------|
| Tab | Next control |
| Shift+Tab | Previous control |
| Arrow Keys | Navigate lists/sliders |
| Enter/Space | Activate button |

## Module-Specific

### Hysteresis Module
| Key | Action |
|-----|--------|
| Space | Play/pause animation |
| R | Reset to initial state |

### MNIST Module
| Key | Action |
|-----|--------|
| C | Clear drawing canvas |
| N | Next random sample |

### Documentation
| Key | Action |
|-----|--------|
| Ctrl+K | Open search |
| Ctrl+F | Find in document |
`,
	})

	// Troubleshooting
	hs.RegisterTopic(&HelpTopic{
		ID:       "troubleshooting",
		Title:    "Troubleshooting",
		Summary:  "Solutions to common problems",
		Category: "Reference",
		Keywords: []string{"troubleshooting", "problems", "issues", "errors", "help", "fix"},
		Content: `# Troubleshooting

## Display Issues

### Window appears blank or corrupted
- Try resizing the window
- Restart the application
- Check graphics drivers are up to date

### Text appears too small/large
- The app scales with system DPI settings
- Adjust your system's display scaling

### Colors look wrong
- Ensure your monitor is calibrated
- Try toggling between light/dark themes

## Performance Issues

### Application runs slowly
- Close other resource-intensive applications
- Reduce animation speeds in settings
- Use smaller array sizes in crossbar demo

### High CPU usage
- Pause running simulations when not in use
- Reduce update frequency in settings

## Recording Issues

### No audio in recordings
- Ensure microphone permissions are granted
- Check PulseAudio/PipeWire is running
- Verify audio input device is selected

### FFmpeg not found
- Install FFmpeg: sudo apt install ffmpeg
- Ensure FFmpeg is in your PATH

## Module-Specific Issues

### Hysteresis: Calibration errors
- Run with --calibrate --force to recalibrate
- Check material data files exist

### MNIST: Model loading fails
- Ensure model files are present in data/
- Check file permissions

### EDA: External tools not launching
- Verify tools are installed and in PATH
- Check tool configuration in settings

## Getting More Help

If problems persist:
1. Check the logs: Enable with --logger flag
2. Review docs in the Documentation module
3. Report issues on GitHub
`,
	})

	// About FeCIM technology
	hs.RegisterTopic(&HelpTopic{
		ID:       "about.fecim",
		Title:    "What is FeCIM?",
		Summary:  "Introduction to Ferroelectric Compute-in-Memory technology",
		Category: "About",
		Keywords: []string{"FeCIM", "ferroelectric", "CIM", "compute-in-memory", "technology"},
		Content: `# What is FeCIM?

**FeCIM** stands for **Ferroelectric Compute-in-Memory**.

## The Core Idea

Traditional computers have a fundamental limitation: processing (CPU/GPU) and memory (RAM/storage) are separate. Every computation requires:

1. Fetch data from memory
2. Send it to the processor
3. Compute results
4. Write results back to memory

This "Von Neumann bottleneck" wastes enormous energy moving data around.

**FeCIM eliminates this problem** by performing computations directly inside memory.

## How It Works

### Ferroelectric Materials
- Materials like HZO (Hafnium Zirconium Oxide) exhibit ferroelectricity
- Electric polarization can be set and retained without power
- Multiple polarization levels enable analog storage

### Crossbar Arrays
- Arrange ferroelectric devices in a grid
- Rows carry input voltages
- Columns collect output currents
- Physics (Ohm's law) performs multiplication!

### Matrix Operations
- Each cell stores a weight (conductance)
- Input voltages are multiplied by weights
- Column currents sum the products
- Complete matrix-vector multiply in one step!

## Why It Matters

### For AI/Machine Learning
- Neural networks are dominated by matrix operations
- FeCIM performs these 10-1000× more efficiently than GPUs
- Enables AI on battery-powered devices

### For Sustainability
- Data centers consume 1-2% of global electricity
- AI energy consumption growing exponentially
- FeCIM could dramatically reduce this footprint

### For Edge Computing
- Sensors and IoT devices have limited power
- FeCIM enables intelligent processing without the cloud
- Faster response, better privacy

## Learn More

Explore each module to dive deeper into specific aspects of the technology.
`,
	})
}
