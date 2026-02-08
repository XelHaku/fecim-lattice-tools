// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// tooltips.go provides comprehensive educational tooltips for all Circuits module parameters.
package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ════════════════════════════════════════════════════════════════════════════════
// CIRCUITS MODULE TOOLTIP CONTENT
// Educational explanations for peripheral electronics parameters
// ════════════════════════════════════════════════════════════════════════════════

// circuitsTooltip holds structured tooltip information.
type circuitsTooltip struct {
	Title       string
	Description string
	Range       string
	Physics     string
	Tips        []string
}

// Format returns a formatted tooltip string for display.
func (t circuitsTooltip) Format() string {
	result := t.Description + "\n\n"
	if t.Range != "" {
		result += "📊 Range: " + t.Range + "\n\n"
	}
	if t.Physics != "" {
		result += "🔬 Physics: " + t.Physics + "\n"
	}
	if len(t.Tips) > 0 {
		result += "\n💡 Tips:\n"
		for _, tip := range t.Tips {
			result += "• " + tip + "\n"
		}
	}
	return result
}

// CircuitsTooltips contains all tooltips for the Circuits module.
var CircuitsTooltips = struct {
	// Mode tooltips
	ReadMode    circuitsTooltip
	WriteMode   circuitsTooltip
	ComputeMode circuitsTooltip

	// Architecture tooltips
	Architecture0T1R circuitsTooltip
	Architecture1T1R circuitsTooltip

	// Sensing tooltips
	TIAGain      circuitsTooltip
	ADCVref      circuitsTooltip
	ADCBits      circuitsTooltip
	ReadVoltage  circuitsTooltip
	SamplePeriod circuitsTooltip

	// Write tooltips
	WriteVoltage circuitsTooltip
	PulseWidth   circuitsTooltip
	TargetLevel  circuitsTooltip
	ISPPMode     circuitsTooltip

	// Compute tooltips
	InputVector   circuitsTooltip
	MVMOperation  circuitsTooltip
	OutputCurrent circuitsTooltip

	// Array tooltips
	ArraySize     circuitsTooltip
	CellState     circuitsTooltip
	Conductance   circuitsTooltip
	IRDrop        circuitsTooltip
	SneakPath     circuitsTooltip

	// Visualization tooltips
	TimingDiagram   circuitsTooltip
	VoltageWaveform circuitsTooltip
	CurrentWaveform circuitsTooltip
	Colormap        circuitsTooltip
	Zoom            circuitsTooltip
}{
	// Mode tooltips
	ReadMode: circuitsTooltip{
		Title:       "Read Mode",
		Description: "Non-destructive sensing of stored conductance states. Applies low read voltage and measures current.",
		Range:       "N/A (mode selection)",
		Physics:     "Read voltage must be below coercive field (Ec) to avoid disturbing polarization. I = G × V_read. TIA converts current to voltage, ADC digitizes.",
		Tips: []string{
			"Use low voltage (0.1-0.3V) to avoid disturbing stored states",
			"Click individual cells to read their conductance",
			"Multiple reads give same result (non-destructive)",
		},
	},
	WriteMode: circuitsTooltip{
		Title:       "Write Mode",
		Description: "Program cells to desired conductance levels using voltage pulses above the coercive field.",
		Range:       "N/A (mode selection)",
		Physics:     "Write voltage exceeds Ec to switch polarization domains. ISPP (Incremental Step Pulse Programming) uses small steps with verify reads for precise targeting.",
		Tips: []string{
			"Select target level before clicking cells to write",
			"ISPP mode automatically adjusts pulse amplitude",
			"Verify read confirms successful programming",
		},
	},
	ComputeMode: circuitsTooltip{
		Title:       "Compute Mode",
		Description: "Perform matrix-vector multiplication using the crossbar array. Apply input vector, measure output currents.",
		Range:       "N/A (mode selection)",
		Physics:     "Input voltages applied to word lines. Each column sums I = Σ(G_ij × V_i) via Kirchhoff's current law. TIAs and ADCs read all columns in parallel.",
		Tips: []string{
			"Enter input vector values (0.0 to 1.0)",
			"Output is weighted sum of inputs × conductances",
			"Watch timing diagram to see parallel operation",
		},
	},

	// Architecture tooltips
	Architecture0T1R: circuitsTooltip{
		Title:       "0T1R (Passive) Architecture",
		Description: "One resistor per crosspoint, no selector transistor. Simplest structure but susceptible to sneak paths.",
		Range:       "N/A (architecture selection)",
		Physics:     "All word lines are always connected to cells. Sneak currents flow through unselected cells, corrupting readings. Larger arrays have worse sneak path problems.",
		Tips: []string{
			"Good for small arrays (<32×32)",
			"Observe sneak current in analyze mode",
			"Real products typically avoid pure passive",
		},
	},
	Architecture1T1R: circuitsTooltip{
		Title:       "1T1R Architecture",
		Description: "One transistor + one resistor per crosspoint. Transistor acts as selector to block sneak paths.",
		Range:       "N/A (architecture selection)",
		Physics:     "Transistor gate controls cell selection. OFF transistors have ~10⁶× higher resistance, blocking sneak current. Enables larger arrays.",
		Tips: []string{
			"Industry standard for production CIM",
			"~1000× reduction in sneak current",
			"Adds area but essential for accuracy",
		},
	},

	// Sensing tooltips
	TIAGain: circuitsTooltip{
		Title:       "TIA Transimpedance (Rf)",
		Description: "Feedback resistance of the transimpedance amplifier. Converts current to voltage: V_out = -I_in × Rf.",
		Range:       "1 kΩ to 100 kΩ typical",
		Physics:     "Higher Rf = more gain = larger output voltage for small currents. But also amplifies noise. Bandwidth = GBW / (1 + Rf×C_parasitic).",
		Tips: []string{
			"10-50 kΩ is typical for FeCIM sensing",
			"Match Rf to expected current range",
			"Higher Rf needed for high-resistance cells",
		},
	},
	ADCVref: circuitsTooltip{
		Title:       "ADC Reference Voltage",
		Description: "Full-scale reference for the analog-to-digital converter. Sets the maximum measurable voltage.",
		Range:       "0.5V to 1.5V typical",
		Physics:     "ADC output = (V_in / V_ref) × 2^N. If V_in > V_ref, output saturates. V_ref should match TIA output range for full resolution.",
		Tips: []string{
			"Match Vref to expected TIA output swing",
			"Lower Vref = finer resolution but risk of clipping",
			"Calibrate with known conductance values",
		},
	},
	ADCBits: circuitsTooltip{
		Title:       "ADC Resolution",
		Description: "Number of bits in the ADC output. Determines quantization precision.",
		Range:       "4-10 bits (6 bits typical for CIM)",
		Physics:     "LSB = V_ref / 2^N. More bits = finer resolution but noise may dominate. Effective bits limited by thermal noise (kT/C).",
		Tips: []string{
			"6 bits sufficient for most neural networks",
			"Beyond 8 bits rarely helps due to noise",
			"Power increases ~2× per bit",
		},
	},
	ReadVoltage: circuitsTooltip{
		Title:       "Read Voltage",
		Description: "Voltage applied to word lines during read operations. Must be low enough to avoid disturbing stored state.",
		Range:       "0.05V to 0.5V (typically ~Ec/5)",
		Physics:     "V_read << V_coercive to prevent domain switching. Higher voltage = more current = better SNR, but risk of disturb increases.",
		Tips: []string{
			"Start with 0.1V for safe non-disturb reads",
			"Increase for higher SNR if disturb is acceptable",
			"Monitor cell state after repeated reads",
		},
	},
	SamplePeriod: circuitsTooltip{
		Title:       "Sample Period",
		Description: "Duration of the sample-and-hold phase before ADC conversion.",
		Range:       "10 ns to 1 µs",
		Physics:     "Must be long enough for settling: T_sample > 5×RC. Shorter = faster but less accurate. S/H capacitor needs time to acquire signal.",
		Tips: []string{
			"100 ns is typical for moderate speed",
			"Watch for incomplete settling at high speed",
			"Longer periods improve accuracy at cost of throughput",
		},
	},

	// Write tooltips
	WriteVoltage: circuitsTooltip{
		Title:       "Write Voltage",
		Description: "Amplitude of voltage pulse for programming cells. Must exceed coercive field.",
		Range:       "1.0V to 3.5V (2-3× Ec typical)",
		Physics:     "V_write > V_coercive causes domain nucleation and growth. Higher voltage = faster switching but more power. Pulse shape affects switching dynamics.",
		Tips: []string{
			"Start at 2× Ec for reliable switching",
			"ISPP automatically adjusts amplitude",
			"Too high voltage may cause breakdown",
		},
	},
	PulseWidth: circuitsTooltip{
		Title:       "Write Pulse Width",
		Description: "Duration of the write voltage pulse. Affects switching probability and energy.",
		Range:       "10 ns to 10 µs",
		Physics:     "Domain switching follows Kolmogorov-Avrami kinetics. Shorter pulse + higher voltage = same switching. Energy ~ C×V²×t.",
		Tips: []string{
			"100 ns is typical for ISPP",
			"Shorter pulses need higher voltage",
			"Multiple short pulses for gradual approach",
		},
	},
	TargetLevel: circuitsTooltip{
		Title:       "Target Conductance Level",
		Description: "Desired discrete level for write operation (0 to N-1 for N-level device).",
		Range:       "0 to 29 for 30-level device",
		Physics:     "Each level corresponds to distinct remnant polarization. Level 0 = minimum conductance, Level N-1 = maximum conductance.",
		Tips: []string{
			"Levels are uniformly spaced in conductance",
			"ISPP iterates until target is reached",
			"Neighboring levels may be hard to distinguish",
		},
	},
	ISPPMode: circuitsTooltip{
		Title:       "ISPP (Incremental Step Pulse Programming)",
		Description: "Programming mode that applies incrementing voltage pulses with verify reads between each pulse.",
		Range:       "Step size: 10-50 mV typical",
		Physics:     "Each pulse nudges polarization toward target. Verify read checks progress. Adaptive step size converges to target level with minimal overshoot.",
		Tips: []string{
			"More reliable than single-shot programming",
			"Slower but more precise",
			"Essential for multi-level storage",
		},
	},

	// Compute tooltips
	InputVector: circuitsTooltip{
		Title:       "Input Vector",
		Description: "Voltage values applied to word lines for MVM computation. Represents the 'x' in y = Wx.",
		Range:       "0.0 to 1.0 (normalized), mapped to read voltage range",
		Physics:     "Each word line receives V_i × V_read. Parallel application enables O(1) MVM regardless of matrix size.",
		Tips: []string{
			"Enter values for each word line",
			"Values are scaled to safe read range",
			"Zero input = no current contribution",
		},
	},
	MVMOperation: circuitsTooltip{
		Title:       "Matrix-Vector Multiply (MVM)",
		Description: "Core CIM operation: output = conductance_matrix × input_vector, computed in analog.",
		Range:       "Output current depends on G and V values",
		Physics:     "Ohm's law (I=GV) at each cell. Kirchhoff's current law (currents sum) at each column. Single cycle computes all outputs in parallel.",
		Tips: []string{
			"Watch all columns compute simultaneously",
			"Compare ideal vs actual (with non-idealities)",
			"This is the source of energy efficiency",
		},
	},
	OutputCurrent: circuitsTooltip{
		Title:       "Output Current",
		Description: "Summed current at each bit line (column). Represents one element of the output vector.",
		Range:       "nA to mA depending on array size and inputs",
		Physics:     "I_out,j = Σ_i (G_ij × V_i). TIA converts to voltage: V_TIA = -I_out × Rf. ADC digitizes for downstream processing.",
		Tips: []string{
			"Higher G cells contribute more current",
			"Clipping occurs if TIA saturates",
			"Watch timing diagram for settling",
		},
	},

	// Array tooltips
	ArraySize: circuitsTooltip{
		Title:       "Array Dimensions",
		Description: "Number of rows (word lines) and columns (bit lines) in the crossbar array.",
		Range:       "4×4 to 128×128 (64×64 typical)",
		Physics:     "Rows = input vector size, Columns = output vector size. Larger arrays enable bigger matrices but have worse IR drop and sneak paths.",
		Tips: []string{
			"8×8 for learning, 64×64 for realistic demos",
			"Real chips tile multiple smaller arrays",
			"Size affects non-ideality severity",
		},
	},
	CellState: circuitsTooltip{
		Title:       "Cell State",
		Description: "Current polarization/conductance level of a single memory cell.",
		Range:       "Level 0 to Level N-1",
		Physics:     "Ferroelectric domains aligned in different ratios. Each level is a metastable state on the P-E loop.",
		Tips: []string{
			"Click cells to see detailed state info",
			"Color indicates conductance magnitude",
			"States persist until next write",
		},
	},
	Conductance: circuitsTooltip{
		Title:       "Cell Conductance",
		Description: "Electrical conductance (1/resistance) of the memory cell in its current state.",
		Range:       "1-100 µS (10 kΩ to 1 MΩ)",
		Physics:     "Conductance is proportional to polarization state. G = G_min + (G_max - G_min) × (level / (N-1)).",
		Tips: []string{
			"Higher G = brighter in heatmap = larger weight",
			"G_min/G_max ratio affects bit precision",
			"Conductance drift affects retention",
		},
	},
	IRDrop: circuitsTooltip{
		Title:       "IR Drop",
		Description: "Voltage loss along metal interconnects due to wire resistance. Causes cells far from drivers to see reduced voltage.",
		Range:       "0-30% at array corners (size-dependent)",
		Physics:     "V_eff = V_applied - I × R_wire. Metal lines have finite conductivity. Cells at far corners have longest wire paths.",
		Tips: []string{
			"Worse for larger arrays",
			"Corner cells most affected",
			"Mitigation: wider wires, tiled architectures",
		},
	},
	SneakPath: circuitsTooltip{
		Title:       "Sneak Path Current",
		Description: "Parasitic current flowing through unselected cells that corrupts the intended measurement.",
		Range:       "0-100%+ of signal (depends on architecture)",
		Physics:     "In passive arrays, current can take alternative paths through neighboring cells. 1T1R selector transistors block these paths.",
		Tips: []string{
			"Main problem for passive (0T1R) arrays",
			"Worse when target cell has low G",
			"1T1R provides ~1000× suppression",
		},
	},

	// Visualization tooltips
	TimingDiagram: circuitsTooltip{
		Title:       "Timing Diagram",
		Description: "Visualization of voltage and current waveforms over time during operations.",
		Range:       "Time scale: ns to ms",
		Physics:     "Shows sequence: word line voltages, bit line currents, ADC sampling windows. Helps understand operation timing.",
		Tips: []string{
			"Yellow traces = voltages (word lines)",
			"Cyan traces = currents (bit lines)",
			"Vertical lines = sampling points",
		},
	},
	VoltageWaveform: circuitsTooltip{
		Title:       "Voltage Waveform",
		Description: "Time-domain view of word line voltage during operations.",
		Range:       "0 to V_max (typically 0-2V)",
		Physics:     "DAC output rises/falls with finite slew rate. Pulse shapes affect switching dynamics and energy.",
		Tips: []string{
			"Watch for ringing or overshoot",
			"Rise/fall time affects settling",
			"Ideal pulses have sharp edges",
		},
	},
	CurrentWaveform: circuitsTooltip{
		Title:       "Current Waveform",
		Description: "Time-domain view of bit line current during read/compute operations.",
		Range:       "nA to mA depending on operation",
		Physics:     "Current rises as cells conduct, settles when parasitic capacitances charge. Sample after settling for accurate read.",
		Tips: []string{
			"Wait for settling before sampling",
			"Overshoot indicates parasitic inductance",
			"Noise visible on small currents",
		},
	},
	Colormap: circuitsTooltip{
		Title:       "Colormap",
		Description: "Color scheme for visualizing conductance values in the array heatmap.",
		Range:       "fecim, viridis, plasma, coolwarm",
		Physics:     "No physics—purely visualization preference. Choose for clarity and accessibility.",
		Tips: []string{
			"'viridis' is perceptually uniform",
			"'coolwarm' shows +/- deviation",
			"'fecim' uses project brand colors",
		},
	},
	Zoom: circuitsTooltip{
		Title:       "Zoom Level",
		Description: "Magnification factor for the array visualization.",
		Range:       "0.5× to 3.0×",
		Physics:     "No physics—pure UI control.",
		Tips: []string{
			"Zoom in to see individual cell details",
			"Zoom out for array-wide patterns",
			"Click cells works at any zoom",
		},
	},
}

// ShowCircuitsTooltip displays a tooltip dialog for the given content.
func ShowCircuitsTooltip(t circuitsTooltip, window fyne.Window) {
	if window == nil {
		return
	}

	titleLabel := widget.NewLabelWithStyle(t.Title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	content := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		widget.NewLabel(t.Description),
	)

	if t.Range != "" {
		rangeTitle := widget.NewLabelWithStyle("📊 Valid Range", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		rangeLabel := widget.NewLabel(t.Range)
		rangeLabel.Wrapping = fyne.TextWrapWord
		content.Add(widget.NewSeparator())
		content.Add(rangeTitle)
		content.Add(rangeLabel)
	}

	if t.Physics != "" {
		physicsTitle := widget.NewLabelWithStyle("🔬 Physical Meaning", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		physicsLabel := widget.NewLabel(t.Physics)
		physicsLabel.Wrapping = fyne.TextWrapWord
		content.Add(widget.NewSeparator())
		content.Add(physicsTitle)
		content.Add(physicsLabel)
	}

	if len(t.Tips) > 0 {
		tipsTitle := widget.NewLabelWithStyle("💡 Tips for New Users", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		content.Add(widget.NewSeparator())
		content.Add(tipsTitle)
		for _, tip := range t.Tips {
			tipLabel := widget.NewLabel("• " + tip)
			tipLabel.Wrapping = fyne.TextWrapWord
			content.Add(tipLabel)
		}
	}

	scroll := container.NewScroll(content)
	scroll.SetMinSize(fyne.NewSize(450, 300))

	dialog.ShowCustom(t.Title, "Got it!", container.NewPadded(scroll), window)
}

// CircuitsInfoButton creates an info button that shows the given tooltip.
func CircuitsInfoButton(t circuitsTooltip, window fyne.Window) *widget.Button {
	btn := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		ShowCircuitsTooltip(t, window)
	})
	btn.Importance = widget.LowImportance
	return btn
}

// WithCircuitsTooltip wraps content with an info button.
func WithCircuitsTooltip(content fyne.CanvasObject, t circuitsTooltip, window fyne.Window) fyne.CanvasObject {
	infoBtn := CircuitsInfoButton(t, window)
	return container.NewBorder(nil, nil, nil, infoBtn, content)
}
