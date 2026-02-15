// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// This file contains the VOLTAGE RULES section of the REFERENCE tab.
package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// ============================================================================
// REFERENCE VOLTAGE RULES SECTION
// ============================================================================

// createReferenceVoltageSection creates the VOLTAGE RULES educational section
func (ca *CircuitsApp) createReferenceVoltageSection() fyne.CanvasObject {
	header := container.NewVBox(
		widget.NewLabelWithStyle("Voltage Rules by Architecture & Mode", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Learn the voltage constraints for safe operation of FeCIM crossbar arrays."),
		widget.NewSeparator(),
	)

	content := ca.createVoltageRulesContent()

	return container.NewBorder(header, nil, nil, nil, content)
}

// createVoltageRulesContent creates the accordion-based content for voltage rules
func (ca *CircuitsApp) createVoltageRulesContent() fyne.CanvasObject {
	// Create accordion with sections for each architecture
	accordion := widget.NewAccordion(
		widget.NewAccordionItem("Passive (0T1R) Architecture", ca.createPassiveVoltageRules()),
		widget.NewAccordionItem("1T1R Architecture", ca.create1T1RVoltageRules()),
		widget.NewAccordionItem("2T1R Architecture", ca.create2T1RVoltageRules()),
		widget.NewAccordionItem("Voltage Safety Rules", ca.createVoltageSafetyRules()),
	)
	accordion.MultiOpen = true
	return container.NewVScroll(accordion)
}

// createPassiveVoltageRules creates the voltage rules content for Passive (0T1R) architecture
func (ca *CircuitsApp) createPassiveVoltageRules() fyne.CanvasObject {
	// Get key values from physics
	readMax := ca.deviceState.GetReadRange().Max

	content := widget.NewRichTextFromMarkdown(fmt.Sprintf(`
## Passive (0T1R) - No Transistor Isolation

**Key Characteristic:** All word lines (WLs) are always active. No transistor gating.

### READ Mode
| Parameter | Value | Constraint |
|-----------|-------|------------|
| WL Voltage | 0.1-0.5V | Empirical simulation default (read-disturb not yet device-calibrated) |
| BL Voltage | Sense | Current flows to TIA |
| Unselected WLs | 0V | Grounded |
| **Read Max** | %.2fV | Empirical guardrail: < 0.7 x Vc |

**Why?** The 70%% of Vc read limit is an empirical simulator guard band (assumed for now) to reduce disturb risk; treat as architecture guidance, not a universally validated device limit.

**Thickness dependence:** Vc ≈ Ec×t. Both Ec and thickness (t) depend on the film stack. Example calibrated references used elsewhere in this repo:
- Park et al. 2015 (10 nm HZO), doi:10.1002/adma.201404531
- Cheema et al. 2020 (5 nm HZO superlattice), doi:10.1038/s41586-020-2208-x

### WRITE Mode (V/2 Half-Select Scheme)
| Parameter | Value | Purpose |
|-----------|-------|---------|
| Selected WL | +V/2 | Half of write voltage |
| Selected BL | -V/2 | Opposite polarity |
| Target Cell dV | +/-1.5V | Full switching (> Vc) |
| Half-Selected dV | +/-0.75V | Below Vc (minimal disturb) |

**V/2 Scheme:** Splits write voltage between WL and BL to minimize disturb on half-selected cells.

### COMPUTE (MVM) Mode
| Parameter | Value | Purpose |
|-----------|-------|---------|
| All WLs | 1.0V | All rows active (no gating) |
| BL Voltage (DAC) | 0-1.0V | Input vector encoding |
| Output Current | I = Sum(G x V) | Column sum to ADC |

**Limitation:** Cannot selectively activate rows - all cells participate in MVM.

### Sneak Path Warning
Passive arrays suffer from sneak currents through unselected cells.
- Sneak/Signal ratio: ~2:1 (200%%) in worst case
- Mitigation: V/2 biasing reduces but doesn't eliminate sneak paths
- Best for: Small arrays (<=32x32)
`, readMax))

	return container.NewPadded(content)
}

// create1T1RVoltageRules creates the voltage rules content for 1T1R architecture
func (ca *CircuitsApp) create1T1RVoltageRules() fyne.CanvasObject {
	content := widget.NewRichTextFromMarkdown(`
## 1T1R - One Transistor Per Cell

**Key Characteristic:** Each cell has a select transistor. WL controls access.

### READ Mode
| Parameter | Value | Constraint |
|-----------|-------|------------|
| WL (selected) | 1.0V (HIGH) | Turn ON transistor |
| WL (unselected) | 0.0V (LOW) | Turn OFF transistor |
| BL Voltage | 0.2V | Read voltage through transistor |
| SL Voltage | 0.0V | Source line grounded |

**Transistor Isolation:**
- ON state: R_on ~ 1 kOhm (cell accessible)
- OFF state: R_off ~ 1 GOhm (cell isolated, 1000x sneak reduction)

### WRITE Mode
| Parameter | Value | Purpose |
|-----------|-------|---------|
| Selected WL | 1.0V | Transistor ON |
| Unselected WLs | 0.0V | Transistors OFF (isolated) |
| Selected BL | +/-1.5V | Write voltage |
| SL | 0.0V | Ground reference |

**No V/2 Needed:** Transistor OFF-state provides complete isolation. Only target cell sees write voltage.

### COMPUTE (MVM) Mode
| Parameter | Value | Purpose |
|-----------|-------|---------|
| Active WLs | 1.0V | User-selected rows ON |
| Inactive WLs | 0.0V | Rows OFF (excluded from MVM) |
| BL Voltage (DAC) | 0-1.0V | Input vector |
| SL | 0.0V | Ground |

**Advantage:** Row-selective MVM! Can choose which rows participate in computation.

### Sneak Path Suppression
- Sneak/Signal ratio: ~0.002:1 (0.2%)
- 1000x improvement over passive
- Best for: Medium arrays (64-256x256)
`)

	return container.NewPadded(content)
}

// create2T1RVoltageRules creates the voltage rules content for 2T1R architecture
func (ca *CircuitsApp) create2T1RVoltageRules() fyne.CanvasObject {
	content := widget.NewRichTextFromMarkdown(`
## 2T1R - Separate Read/Write Paths

**Key Characteristic:** Dual transistors provide independent read and write paths.

### READ Mode
| Parameter | Value | Purpose |
|-----------|-------|---------|
| WL_read (selected) | 1.0V | Read transistor ON |
| WL_write (all) | 0.0V | Write transistor OFF |
| BL Voltage | 0.2V | Read voltage |
| SL | 0.0V | Ground |

**Path Isolation:** Write path completely isolated during read operations.

### WRITE Mode
| Parameter | Value | Purpose |
|-----------|-------|---------|
| WL_write (selected) | 1.0V | Write transistor ON |
| WL_read (all) | 0.0V | Read transistor OFF |
| BL | +/-1.5V | Write voltage |
| SL | 0.0V | Ground |

**Path Isolation:** Read path completely isolated during write operations.

### COMPUTE (MVM) Mode
| Parameter | Value | Purpose |
|-----------|-------|---------|
| WL_read (active rows) | 1.0V | Read path enabled |
| WL_write (all) | 0.0V | Write path isolated |
| BL Voltage (DAC) | 0-1.0V | Input vector |
| SL | 0.0V | Ground |

**Ultra-Low Disturb:** Write circuitry completely isolated during compute.

### Advantages
- Zero write stress during MVM operations
- Independent voltage optimization for read/write
- Complete path isolation
- Best for: Large arrays (>256x256), high-precision computing
`)

	return container.NewPadded(content)
}

// createVoltageSafetyRules creates the voltage safety rules content
func (ca *CircuitsApp) createVoltageSafetyRules() fyne.CanvasObject {
	mat := ca.deviceState.GetMaterial()
	Vc := mat.CoerciveVoltage()
	readMax := ca.deviceState.GetReadRange().Max
	writeMin := ca.deviceState.GetWriteRange().Min
	writeMax := ca.deviceState.GetWriteRange().Max

	content := widget.NewRichTextFromMarkdown(fmt.Sprintf(`
## Voltage Safety Rules

### Material Parameters (Current: %s)
| Parameter | Value | Source |
|-----------|-------|--------|
| Coercive Voltage (Vc) | %.2fV | Derived: Ec x thickness (sim model); Ec is thickness/process dependent |
| Read Max | %.2fV | Empirical simulator guard band: 0.7 x Vc (assumed; needs per-device validation) |
| Write Min | %.2fV | Derived from Vc threshold in current model |
| Write Max | %.2fV | Simulator ceiling: 2.5 x Vc (engineering safety margin) |

**Thickness-dependent note:** Reported Ec varies with stack engineering/thickness (often cited ~0.6-1.5 MV/cm for HZO families), so required write voltage is not universal.
**Literature context:** Sub-1V switching has been reported in aggressively scaled ferroelectric stacks (~3.6 nm) (ACS Applied Materials & Interfaces, 2024). This simulator keeps conservative defaults for educational stability.

### Critical Safety Rules

**Rule 1: V_read < 0.7 x Vc (empirical simulator rule)**
- Read voltage stays 30%% below coercive voltage in this model
- Used as a conservative non-disturb assumption (not yet tied to a single measured source)
- Current setting: %.2fV < %.2fV (OK)

**Rule 2: V_write >= Vc**
- Write voltage must exceed coercive voltage
- Ensures polarization switching
- Current range: %.2fV - %.2fV

**Rule 3: V_half_select < Vc (Passive only)**
- Half-select voltage must be below Vc
- Minimizes disturb on unselected cells
- V/2 = 0.75V < Vc = %.2fV (OK)

**Rule 4: MVM range = ADC range**
- Input DAC: 0-1.0V
- Output ADC: 0-1.0V (after TIA)
- Prevents clipping and saturation
- ADC quantization policy: round-to-nearest code (ties half-up) after clamping to the ADC reference window

### Architecture Selection Guide

| Array Size | Recommended | Reason |
|------------|-------------|--------|
| <=32x32 | Passive (0T1R) | Maximum density, V/2 manageable |
| 64-256x256 | 1T1R | 1000x sneak reduction, row-selective |
| >256x256 | 2T1R | Ultra-precision, dual-path isolation |
`, mat.Name, Vc, readMax, writeMin, writeMax, readMax, 0.7*Vc, writeMin, writeMax, Vc))

	readMarginBadge := container.NewHBox(
		widget.NewLabel("Read margin confidence:"),
		sharedwidgets.NewConfidenceBadge(sharedwidgets.Estimated),
	)
	return container.NewPadded(container.NewVBox(readMarginBadge, content))
}
