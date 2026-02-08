// Package widgets provides reusable UI components.
// tooltips.go provides comprehensive educational tooltips for FeCIM GUI elements.
package widgets

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// ════════════════════════════════════════════════════════════════════════════════
// TOOLTIP CONTENT DEFINITIONS
// Educational explanations for all FeCIM parameters and controls
// ════════════════════════════════════════════════════════════════════════════════

// TooltipContent holds structured tooltip information.
type TooltipContent struct {
	Title       string   // Short title (shown as header)
	Description string   // What this parameter does
	Range       string   // Valid range or options
	Physics     string   // Physical meaning / underlying principle
	Tips        []string // Usage tips for new users
}

// Format returns a formatted tooltip string.
func (t TooltipContent) Format() string {
	result := fmt.Sprintf("━━ %s ━━\n\n", t.Title)
	result += t.Description + "\n\n"
	if t.Range != "" {
		result += fmt.Sprintf("Range: %s\n", t.Range)
	}
	if t.Physics != "" {
		result += fmt.Sprintf("\n🔬 Physics: %s\n", t.Physics)
	}
	if len(t.Tips) > 0 {
		result += "\n💡 Tips:\n"
		for _, tip := range t.Tips {
			result += fmt.Sprintf("  • %s\n", tip)
		}
	}
	return result
}

// Short returns a brief tooltip (just description).
func (t TooltipContent) Short() string {
	return t.Description
}

// ════════════════════════════════════════════════════════════════════════════════
// MODULE 1: HYSTERESIS TOOLTIPS
// P-E curve physics and ferroelectric material parameters
// ════════════════════════════════════════════════════════════════════════════════

// HysteresisTooltips contains all tooltips for the Hysteresis module.
var HysteresisTooltips = struct {
	Material      TooltipContent
	Waveform      TooltipContent
	PhysicsEngine TooltipContent
	Levels        TooltipContent
	TargetRange   TooltipContent
	EField        TooltipContent
	Frequency     TooltipContent
	TimeScale     TooltipContent
	Temperature   TooltipContent
	Stress        TooltipContent
	TrailLength   TooltipContent
}{
	Material: TooltipContent{
		Title:       "Ferroelectric Material",
		Description: "The ferroelectric material determines the P-E hysteresis shape and multi-level storage capacity.",
		Range:       "HZO (Si-doped), FeCIM HZO, Literature Superlattice, Cryogenic HZO, AlScN",
		Physics:     "Different materials have different coercive fields (Ec), remnant polarization (Pr), and Curie temperatures. HZO is CMOS-compatible at ~10nm thickness.",
		Tips: []string{
			"Start with 'Literature Superlattice' for typical research parameters",
			"'Cryogenic HZO (4K)' shows enhanced properties at low temperature",
			"Material choice affects the number of stable polarization levels",
		},
	},
	Waveform: TooltipContent{
		Title:       "Drive Waveform",
		Description: "How the electric field is applied over time. Different waveforms reveal different physics.",
		Range:       "Manual, Sine Wave, Triangle Wave, ISPP (Write/Read), Time-Resolved Switching",
		Physics:     "Triangle waves trace the full hysteresis loop. ISPP (Incremental Step Pulse Programming) uses small voltage pulses for precise level targeting—the key to multi-level storage.",
		Tips: []string{
			"Use 'Manual' to explore the P-E curve interactively",
			"'ISPP (Write/Read)' demonstrates multi-level write/verify cycles",
			"'Time-Resolved Switching' shows domain nucleation dynamics",
		},
	},
	PhysicsEngine: TooltipContent{
		Title:       "Physics Engine",
		Description: "The mathematical model used to simulate ferroelectric switching.",
		Range:       "Landau-Khalatnikov (LK), Preisach",
		Physics:     "LK: Time-domain differential equation from Landau-Ginzburg theory—captures switching dynamics and relaxation. Preisach: Statistical hysteresis model treating the material as independent hysterons—better for multi-domain behavior.",
		Tips: []string{
			"LK is better for time-resolved switching simulations",
			"Preisach is better for quasi-static hysteresis loops",
			"Both should give similar steady-state P-E curves",
		},
	},
	Levels: TooltipContent{
		Title:       "Quantization Levels",
		Description: "Number of discrete polarization states for multi-level storage. More levels = more bits per cell.",
		Range:       "2-256 levels (1-8 bits/cell)",
		Physics:     "Each level corresponds to a distinct remnant polarization value. Conference claims suggest 30 levels; peer-reviewed literature shows 32-140 levels depending on conditions.",
		Tips: []string{
			"30 levels ≈ 4.9 bits/cell (conference claim)",
			"Higher levels require better analog precision",
			"Start with 8-16 levels to see clear step separation",
		},
	},
	TargetRange: TooltipContent{
		Title:       "Target Range",
		Description: "Fraction of full polarization swing used for level mapping. Smaller range avoids saturation regions.",
		Range:       "80-100% of Ps (saturation polarization)",
		Physics:     "Operating near saturation (Ps) gives maximum signal but risks distortion. Backing off to 85-95% improves linearity at cost of reduced signal margin.",
		Tips: []string{
			"95% is a good balance of range and linearity",
			"Lower values help with noisy materials",
			"Calibration automatically adjusts pulse amplitudes",
		},
	},
	EField: TooltipContent{
		Title:       "Electric Field",
		Description: "Applied electric field strength. Controls polarization switching in manual mode.",
		Range:       "±1.5 × Ec (coercive field), typically ±1-3 MV/cm",
		Physics:     "E > Ec causes domain switching. The coercive field Ec is the field where polarization reverses. Ferroelectric memories operate by applying fields above/below Ec.",
		Tips: []string{
			"Drag slowly to trace the hysteresis loop",
			"Notice the 'snap' at ±Ec where switching occurs",
			"Field is normalized to material's Ec for comparison",
		},
	},
	Frequency: TooltipContent{
		Title:       "Drive Frequency",
		Description: "Frequency of the periodic waveform (sine, triangle). Affects switching dynamics.",
		Range:       "0.01 Hz - 1 GHz (display limited by time scale)",
		Physics:     "Real switching occurs on ns-µs timescales. Low frequencies let you observe quasi-static behavior; high frequencies reveal dynamic effects like viscous domain wall motion.",
		Tips: []string{
			"Use ~1 Hz with slow time scale to see smooth loops",
			"MHz frequencies are realistic for memory operation",
			"Frequency affects energy dissipation per cycle",
		},
	},
	TimeScale: TooltipContent{
		Title:       "Time Scale",
		Description: "Slow-motion factor for visualization. Does not change physics, only display speed.",
		Range:       "1× (real-time) to 10⁻⁹× (billion times slower)",
		Physics:     "Ferroelectric switching happens in nanoseconds. Time scaling lets you visualize fast dynamics in human-perceptible time without changing the underlying physics.",
		Tips: []string{
			"10⁻⁶× (microsecond view) is good for ISPP demos",
			"Combine with frequency to control perceived speed",
			"Reset clears the trail when changing time scale",
		},
	},
	Temperature: TooltipContent{
		Title:       "Temperature",
		Description: "Device operating temperature. Affects Ec, Pr, and switching behavior.",
		Range:       "200-700 K (-73°C to 427°C)",
		Physics:     "Ferroelectric properties degrade as T approaches Curie temperature (Tc). HZO: Tc ≈ 450°C. Below Tc, Ec and Pr increase at lower temperatures. Above Tc, ferroelectricity vanishes.",
		Tips: []string{
			"Room temperature (300 K) is the default",
			"Try 77 K (liquid nitrogen) for enhanced properties",
			"Watch Ec and Pr markers shift with temperature",
		},
	},
	Stress: TooltipContent{
		Title:       "Mechanical Stress",
		Description: "Applied mechanical stress (electrostriction effect). Affects coercive field and loop shape.",
		Range:       "0-5 GPa",
		Physics:     "Piezoelectric coupling means mechanical stress shifts the energy landscape. Compressive stress typically increases Ec. This is relevant for strained thin films and 3D integration.",
		Tips: []string{
			"1 GPa is a typical thin-film stress value",
			"Higher stress can stabilize certain phases",
			"Observe changes in loop width (Ec) with stress",
		},
	},
	TrailLength: TooltipContent{
		Title:       "Trail Length",
		Description: "Number of history points shown on the P-E plot. Longer trails show more of the hysteresis loop.",
		Range:       "1,000-100,000 points",
		Physics:     "The trail visualizes the path through P-E space over time. A complete hysteresis loop requires enough points to trace both branches.",
		Tips: []string{
			"10,000 points is good for full loop visualization",
			"Reduce for faster rendering on slow systems",
			"Reset clears the trail to start fresh",
		},
	},
}

// ════════════════════════════════════════════════════════════════════════════════
// MODULE 2: CROSSBAR TOOLTIPS
// Compute-in-Memory array physics and non-idealities
// ════════════════════════════════════════════════════════════════════════════════

// CrossbarTooltips contains all tooltips for the Crossbar module.
var CrossbarTooltips = struct {
	ArraySize    TooltipContent
	Noise        TooltipContent
	ADCBits      TooltipContent
	Colormap     TooltipContent
	DemoMode     TooltipContent
	Architecture TooltipContent
	Conductance  TooltipContent
	IRDrop       TooltipContent
	SneakPath    TooltipContent
	MVM          TooltipContent
	WireRes      TooltipContent
}{
	ArraySize: TooltipContent{
		Title:       "Array Size",
		Description: "Dimensions of the crossbar array (N×N). Larger arrays compute bigger matrices but have worse non-idealities.",
		Range:       "8×8 to 128×128",
		Physics:     "Crossbar arrays perform matrix-vector multiplication (MVM) using Ohm's law: I = G × V. Array size limits the matrix dimensions in one operation.",
		Tips: []string{
			"64×64 is a common research demonstration size",
			"Larger arrays have more IR drop and sneak paths",
			"Real chips may tile multiple smaller arrays",
		},
	},
	Noise: TooltipContent{
		Title:       "Read Noise",
		Description: "Random variation added to conductance readings. Simulates thermal and device variability.",
		Range:       "0-20% (standard deviation relative to signal)",
		Physics:     "Sources: Johnson-Nyquist thermal noise in resistors, shot noise in transistors, 1/f flicker noise, and device-to-device variability. Noise limits ADC resolution utility.",
		Tips: []string{
			"2% is typical for well-designed sense amplifiers",
			"5-10% represents challenging device variability",
			"Higher noise degrades inference accuracy",
		},
	},
	ADCBits: TooltipContent{
		Title:       "ADC Resolution",
		Description: "Analog-to-digital converter precision for reading column currents.",
		Range:       "4-10 bits",
		Physics:     "ADC converts the analog sum-of-products current to digital values. More bits = finer resolution but higher power and area. The effective resolution is limited by noise.",
		Tips: []string{
			"6 bits is common for neural network inference",
			"Beyond 8 bits rarely helps due to analog noise",
			"Lower bits save power but reduce accuracy",
		},
	},
	Colormap: TooltipContent{
		Title:       "Colormap",
		Description: "Color scheme for visualizing conductance values in the heatmap.",
		Range:       "fecim, viridis, plasma, coolwarm",
		Physics:     "No physics—purely visual preference. 'fecim' uses cyan-to-yellow for the FeCIM brand colors.",
		Tips: []string{
			"'viridis' is perceptually uniform and colorblind-safe",
			"'coolwarm' shows deviation from middle values",
			"Colorbar always shows the mapping",
		},
	},
	DemoMode: TooltipContent{
		Title:       "Demo Mode",
		Description: "Controls how the demonstration progresses through concepts.",
		Range:       "Manual, Auto Demo, Step-by-Step",
		Physics:     "No physics—controls the presentation flow.",
		Tips: []string{
			"'Manual' lets you explore at your own pace",
			"'Auto Demo' cycles through key concepts",
			"'Step-by-Step' pauses for explanation at each stage",
		},
	},
	Architecture: TooltipContent{
		Title:       "Crossbar Architecture",
		Description: "Cell structure at each crosspoint. Affects sneak path suppression.",
		Range:       "0T1R (passive), 1T1R (one transistor), 1S1R (selector)",
		Physics:     "0T1R: Simplest, but sneak currents flow through unselected cells. 1T1R: Transistor blocks sneak paths (~1000× reduction) but adds area. 1S1R: Non-linear selector (threshold switch) provides blocking without transistors.",
		Tips: []string{
			"Start with 0T1R to see sneak path problems",
			"Switch to 1T1R to see the improvement",
			"Real products typically use 1T1R or 1S1R",
		},
	},
	Conductance: TooltipContent{
		Title:       "Cell Conductance",
		Description: "The programmable conductance value stored in each memristive cell.",
		Range:       "1-100 µS typical (10 kΩ - 1 MΩ resistance)",
		Physics:     "Conductance G = 1/R encodes the synaptic weight. In FeCIM, G is set by the polarization state of the ferroelectric material. Multi-level conductance enables analog computation.",
		Tips: []string{
			"Brighter cells = higher conductance = stronger weight",
			"Click cells to see detailed conductance info",
			"Level 0 = minimum G, Level 29 = maximum G",
		},
	},
	IRDrop: TooltipContent{
		Title:       "IR Drop",
		Description: "Voltage drop along metal interconnects due to wire resistance. Causes computation errors.",
		Range:       "0-30% voltage loss at far corners (size-dependent)",
		Physics:     "V_effective = V_applied - I×R_wire. Cells far from voltage drivers see reduced voltage. This causes conductance×voltage products to be systematically lower than ideal.",
		Tips: []string{
			"IR drop increases with array size (longer wires)",
			"Corner cells are worst affected",
			"Mitigation: wider wires, tiled architectures, calibration",
		},
	},
	SneakPath: TooltipContent{
		Title:       "Sneak Path Currents",
		Description: "Parasitic currents through unselected cells that corrupt the intended signal.",
		Range:       "0-100%+ of signal current (architecture-dependent)",
		Physics:     "In passive (0T1R) arrays, current can flow through paths other than the selected cell. The sneak current depends on the conductance pattern of neighboring cells.",
		Tips: []string{
			"Sneak paths are the main challenge for passive arrays",
			"1T1R architecture eliminates most sneak current",
			"Click cells to see sneak current contribution",
		},
	},
	MVM: TooltipContent{
		Title:       "Matrix-Vector Multiplication",
		Description: "The core compute operation: y = G × x, where G is the conductance matrix and x is the input voltage vector.",
		Range:       "Output depends on array size and weight values",
		Physics:     "Kirchhoff's current law: currents from all cells in a column sum at the bit line. This performs the dot product in analog, with O(1) time complexity regardless of matrix size.",
		Tips: []string{
			"'Run MVM' executes one matrix-vector multiply",
			"Compare ideal vs. actual output to see error",
			"Non-idealities (IR drop, sneak, noise) cause errors",
		},
	},
	WireRes: TooltipContent{
		Title:       "Wire Resistance",
		Description: "Resistance per unit length of the metal interconnect lines.",
		Range:       "0.1-10 Ω per cell (technology-dependent)",
		Physics:     "Metal lines have finite conductivity: R = ρL/A. At nanoscale, surface scattering increases resistivity. Wire resistance causes IR drop proportional to distance from drivers.",
		Tips: []string{
			"Cu interconnects: ~2-5 Ω/µm at 28nm node",
			"Lower resistance = less IR drop = larger arrays",
			"Trade-off: wider wires reduce resistance but use area",
		},
	},
}

// ════════════════════════════════════════════════════════════════════════════════
// MODULE 3: MNIST TOOLTIPS
// Neural network inference and quantization effects
// ════════════════════════════════════════════════════════════════════════════════

// MNISTTooltips contains all tooltips for the MNIST module.
var MNISTTooltips = struct {
	Levels       TooltipContent
	Noise        TooltipContent
	ADCBits      TooltipContent
	DACBits      TooltipContent
	Preprocess   TooltipContent
	DrawCanvas   TooltipContent
	Activations  TooltipContent
	WeightMap    TooltipContent
	Confidence   TooltipContent
	Comparison   TooltipContent
	Energy       TooltipContent
	PresetIdeal  TooltipContent
	PresetHW     TooltipContent
	PresetNoisy  TooltipContent
}{
	Levels: TooltipContent{
		Title:       "Quantization Levels",
		Description: "Number of discrete weight values in the neural network. Affects model accuracy and hardware efficiency.",
		Range:       "2-256 levels (weights from pretrained QAT models)",
		Physics:     "Each level corresponds to a programmable conductance state in the crossbar. Quantization-aware training (QAT) optimizes weights for discrete levels.",
		Tips: []string{
			"30 levels ≈ 4.9 bits (conference claim for FeCIM)",
			"8-16 levels often sufficient for MNIST",
			"Only levels with trained weights are available",
		},
	},
	Noise: TooltipContent{
		Title:       "Read Noise",
		Description: "Gaussian noise added to weight reads, simulating analog hardware imperfections.",
		Range:       "0-20% (standard deviation relative to weight magnitude)",
		Physics:     "Combines thermal noise, device variability, and retention drift. The network must be robust to this noise for reliable inference.",
		Tips: []string{
			"0% = ideal (floating-point equivalent)",
			"3% = typical production hardware target",
			"10%+ = stress test for noise robustness",
		},
	},
	ADCBits: TooltipContent{
		Title:       "ADC Resolution",
		Description: "Bit precision of the analog-to-digital converter reading neuron activations.",
		Range:       "4-10 bits",
		Physics:     "ADC quantizes the analog dot-product result. Lower bits = less power/area but more quantization error. 6 bits is common for edge inference.",
		Tips: []string{
			"6 bits is the typical design point",
			"4 bits works for robust networks",
			"8+ bits rarely improves accuracy",
		},
	},
	DACBits: TooltipContent{
		Title:       "DAC Resolution",
		Description: "Bit precision of the digital-to-analog converter generating input voltages.",
		Range:       "4-10 bits",
		Physics:     "DAC converts digital pixel values to analog voltages applied to word lines. Higher precision = more accurate input representation.",
		Tips: []string{
			"MNIST pixels are 8-bit, so 8-bit DAC is lossless",
			"Lower DAC bits add input quantization noise",
			"Usually less critical than ADC for accuracy",
		},
	},
	Preprocess: TooltipContent{
		Title:       "Input Preprocessing",
		Description: "Normalizes hand-drawn digits to match MNIST training distribution.",
		Range:       "On/Off",
		Physics:     "No physics—data preprocessing. Centers and scales drawn digits to match the 28×28 centered format of MNIST training data.",
		Tips: []string{
			"Enable for better recognition of drawn digits",
			"Random test samples are never preprocessed",
			"Preprocessing: threshold → crop → scale → center",
		},
	},
	DrawCanvas: TooltipContent{
		Title:       "Drawing Canvas",
		Description: "Draw a digit with your mouse to test the neural network.",
		Range:       "28×28 pixels (scaled for display)",
		Physics:     "Your drawing is converted to grayscale pixel values (0-255), then normalized and fed as input voltages to the crossbar.",
		Tips: []string{
			"Draw with a thick stroke for best results",
			"Clear and redraw if recognition fails",
			"Try different digit styles to test robustness",
		},
	},
	Activations: TooltipContent{
		Title:       "Layer Activations",
		Description: "Visualization of neuron activations at each network layer.",
		Range:       "Input (784) → Hidden (varies) → Output (10)",
		Physics:     "Each layer performs: output = ReLU(W × input + bias). Activations show which features the network detects at each stage.",
		Tips: []string{
			"Input layer shows the 28×28 image as 784 pixels",
			"Hidden layers learn feature detectors",
			"Output layer has 10 neurons (one per digit class)",
		},
	},
	WeightMap: TooltipContent{
		Title:       "Weight Heatmap",
		Description: "Visualization of the quantized weight matrix stored in the crossbar.",
		Range:       "Intensity maps to weight value (negative=dark, positive=bright)",
		Physics:     "Weights are stored as conductance values in memristive cells. The heatmap shows the pattern learned during training.",
		Tips: []string{
			"Look for digit-shaped patterns in first layer",
			"Compare FP32 vs quantized weights",
			"Noise adds visible speckle to the pattern",
		},
	},
	Confidence: TooltipContent{
		Title:       "Prediction Confidence",
		Description: "Softmax probability distribution over the 10 digit classes.",
		Range:       "0-100% for each class (sums to 100%)",
		Physics:     "Softmax normalizes raw output scores to probabilities. Higher confidence = more certain prediction.",
		Tips: []string{
			"High confidence on correct class = reliable",
			"Similar scores across classes = uncertain",
			"Noise typically reduces confidence margins",
		},
	},
	Comparison: TooltipContent{
		Title:       "FP32 vs CIM Comparison",
		Description: "Side-by-side comparison of floating-point and compute-in-memory inference.",
		Range:       "Agreement rate typically 95-99%",
		Physics:     "FP32 uses 32-bit floating point (ideal). CIM uses quantized weights + analog noise. Differences show hardware effects.",
		Tips: []string{
			"Green = both agree on prediction",
			"Red = predictions differ (hardware error)",
			"Disagreement increases with noise/fewer levels",
		},
	},
	Energy: TooltipContent{
		Title:       "Energy Consumption",
		Description: "Estimated energy per inference for different compute platforms.",
		Range:       "fJ to mJ per inference (varies by 10⁶×)",
		Physics:     "CIM energy: E = CV² per cell access. No data movement = orders of magnitude less than Von Neumann. GPU moves data through memory hierarchy = high energy.",
		Tips: []string{
			"FeCIM: ~1 fJ/MAC (femtojoules per multiply-accumulate)",
			"GPU: ~1 pJ/MAC (1000× more)",
			"CPU: ~10 pJ/MAC (10000× more)",
		},
	},
	PresetIdeal: TooltipContent{
		Title:       "Ideal Preset",
		Description: "Maximum precision, no noise—equivalent to floating-point inference.",
		Range:       "30 levels, 0% noise",
		Physics:     "Best-case scenario: no analog non-idealities. Used as the accuracy reference.",
		Tips: []string{
			"Use to establish baseline accuracy",
			"Compare against hardware presets",
			"Should match FP32 closely",
		},
	},
	PresetHW: TooltipContent{
		Title:       "Hardware Preset",
		Description: "Realistic hardware parameters based on published FeCIM specifications.",
		Range:       "30 levels, 3% noise, 6-bit ADC",
		Physics:     "Represents achievable performance with current FeCIM technology. Noise includes read variability and retention drift.",
		Tips: []string{
			"This is the target for real hardware",
			"Accuracy should be close to ideal",
			"Small accuracy drop is acceptable",
		},
	},
	PresetNoisy: TooltipContent{
		Title:       "Noisy Preset",
		Description: "Stress test with high noise to evaluate robustness.",
		Range:       "30 levels, 15% noise, 6-bit ADC",
		Physics:     "Represents challenging conditions: high variability, poor retention, or degraded devices. Tests network's noise tolerance.",
		Tips: []string{
			"Expect visible accuracy degradation",
			"Some digits become unreliable",
			"Useful for robustness testing",
		},
	},
}

// ════════════════════════════════════════════════════════════════════════════════
// MODULE 4: CIRCUITS TOOLTIPS
// Peripheral electronics: DAC, TIA, ADC
// ════════════════════════════════════════════════════════════════════════════════

// CircuitsTooltips contains all tooltips for the Circuits module.
var CircuitsTooltips = struct {
	DAC           TooltipContent
	TIA           TooltipContent
	ADC           TooltipContent
	SampleHold    TooltipContent
	PulseGen      TooltipContent
	VoltageLevel  TooltipContent
	CurrentRange  TooltipContent
	TimingDiagram TooltipContent
	OpAmp         TooltipContent
	Reference     TooltipContent
}{
	DAC: TooltipContent{
		Title:       "Digital-to-Analog Converter (DAC)",
		Description: "Converts digital input values to analog voltage pulses applied to crossbar word lines.",
		Range:       "4-12 bits, 0-1V output range typical",
		Physics:     "R-2R ladder or current-steering DAC architectures. Each bit doubles the current contribution: V_out = V_ref × (D / 2^N).",
		Tips: []string{
			"Higher bits = finer voltage resolution",
			"MNIST inputs are 8-bit (256 levels)",
			"DAC settling time limits operation speed",
		},
	},
	TIA: TooltipContent{
		Title:       "Transimpedance Amplifier (TIA)",
		Description: "Converts the small column currents from the crossbar to measurable voltages.",
		Range:       "Gain: 10³-10⁶ V/A (1 kΩ to 1 MΩ transimpedance)",
		Physics:     "Op-amp with feedback resistor: V_out = -I_in × R_f. Virtual ground at input maintains low impedance for current summing. Bandwidth limited by gain-bandwidth product.",
		Tips: []string{
			"TIA is critical for analog sensing accuracy",
			"Higher gain amplifies noise too",
			"Bandwidth must exceed signal frequency",
		},
	},
	ADC: TooltipContent{
		Title:       "Analog-to-Digital Converter (ADC)",
		Description: "Converts analog TIA output voltage to digital values for further processing.",
		Range:       "4-10 bits, 1 MS/s to 1 GS/s",
		Physics:     "Flash, SAR, or delta-sigma architectures. Quantizes continuous voltage into 2^N discrete levels. Resolution limited by thermal noise (kT/C).",
		Tips: []string{
			"6-bit SAR ADC is common for CIM",
			"Flash ADC is faster but uses more power/area",
			"Effective bits often less than nominal due to noise",
		},
	},
	SampleHold: TooltipContent{
		Title:       "Sample-and-Hold Circuit",
		Description: "Captures and holds the analog voltage during ADC conversion.",
		Range:       "Hold time: ns to µs depending on ADC speed",
		Physics:     "Switch + capacitor: closes to sample, opens to hold. Droop rate depends on capacitor leakage. Acquisition time set by RC time constant.",
		Tips: []string{
			"Critical for accurate ADC conversion",
			"Larger capacitor = less droop but slower",
			"Switch charge injection causes offset errors",
		},
	},
	PulseGen: TooltipContent{
		Title:       "Pulse Generator",
		Description: "Generates voltage pulses for programming ferroelectric cells (write operations).",
		Range:       "Amplitude: 0-3V, Width: 10ns-10µs",
		Physics:     "ISPP uses incrementing pulse amplitudes with verify. Pulse width affects switching probability. Shorter pulses need higher voltage.",
		Tips: []string{
			"ISPP: Incremental Step Pulse Programming",
			"Each pulse nudges polarization toward target",
			"Verify read checks if target level is reached",
		},
	},
	VoltageLevel: TooltipContent{
		Title:       "Voltage Levels",
		Description: "The discrete voltage values used for multi-level operation.",
		Range:       "Read: 0.1-0.5V (non-disturb), Write: 1-3V (above Ec)",
		Physics:     "Read voltage must be below coercive field to avoid disturbing stored state. Write voltage must exceed Ec to cause polarization switching.",
		Tips: []string{
			"Read voltage << Ec (typically Ec/5)",
			"Write voltage ≈ 2-3 × Ec for reliable switching",
			"Higher write voltage = faster switching but more power",
		},
	},
	CurrentRange: TooltipContent{
		Title:       "Current Sensing Range",
		Description: "The range of currents the TIA can accurately measure.",
		Range:       "nA to mA depending on array size and conductance",
		Physics:     "Column current = Σ(G_i × V_i) over all cells. Full array with high conductance and high input voltage gives maximum current.",
		Tips: []string{
			"1 µA/cell × 64 cells = 64 µA max per column",
			"TIA gain set for this range without saturation",
			"Lower currents need higher gain (more noise)",
		},
	},
	TimingDiagram: TooltipContent{
		Title:       "Timing Diagram",
		Description: "Visualization of the voltage/current waveforms during read/write operations.",
		Range:       "Time scale: ns to ms",
		Physics:     "Shows sequence: apply word line voltages → wait for settling → sense bit line currents → sample/hold → ADC convert → digital output.",
		Tips: []string{
			"Yellow = word line voltage (input)",
			"Cyan = bit line current (output)",
			"Note delays for settling and conversion",
		},
	},
	OpAmp: TooltipContent{
		Title:       "Operational Amplifier",
		Description: "High-gain differential amplifier used in TIA and other circuits.",
		Range:       "Gain: 10⁵-10⁶, Bandwidth: MHz-GHz",
		Physics:     "Ideal op-amp: infinite gain, infinite input impedance, zero output impedance. Real op-amps have finite gain-bandwidth product (GBW).",
		Tips: []string{
			"TIA uses op-amp in inverting configuration",
			"Bandwidth = GBW / closed-loop gain",
			"Low input offset voltage is important for accuracy",
		},
	},
	Reference: TooltipContent{
		Title:       "Voltage Reference",
		Description: "Stable reference voltage for DAC and ADC operation.",
		Range:       "Typically 1.0-1.8V (bandgap reference)",
		Physics:     "Bandgap reference uses silicon's ~1.2V bandgap to create temperature-stable voltage. All analog conversions reference this voltage.",
		Tips: []string{
			"Reference accuracy limits system accuracy",
			"Temperature drift causes systematic errors",
			"On-chip references simpler but less accurate",
		},
	},
}

// ════════════════════════════════════════════════════════════════════════════════
// MODULE 5: COMPARISON TOOLTIPS
// Technology benchmarks and energy calculations
// ════════════════════════════════════════════════════════════════════════════════

// ComparisonTooltips contains all tooltips for the Comparison module.
var ComparisonTooltips = struct {
	Workload      TooltipContent
	EnergyPerMAC  TooltipContent
	Throughput    TooltipContent
	PowerDraw     TooltipContent
	CostPerOp     TooltipContent
	TRL           TooltipContent
	DRAM          TooltipContent
	NAND          TooltipContent
	GPU           TooltipContent
	CPU           TooltipContent
	FeCIM         TooltipContent
	DataCenter    TooltipContent
}{
	Workload: TooltipContent{
		Title:       "Workload Selection",
		Description: "The neural network model used for comparison calculations.",
		Range:       "MNIST (small), ResNet-50 (medium), GPT-2 (large)",
		Physics:     "Each workload has different MAC (multiply-accumulate) counts and memory requirements. Larger models benefit more from CIM's memory bandwidth advantage.",
		Tips: []string{
			"MNIST: 0.4M MACs (tiny, for demonstration)",
			"ResNet-50: 4B MACs (typical vision model)",
			"GPT-2: 1.5B+ parameters (language model)",
		},
	},
	EnergyPerMAC: TooltipContent{
		Title:       "Energy per MAC Operation",
		Description: "Energy consumed for one multiply-accumulate operation.",
		Range:       "fJ (FeCIM) to pJ (GPU) to nJ (CPU data movement)",
		Physics:     "CIM: E = CV² per cell read. GPU/CPU: Dominated by SRAM/DRAM access energy, not compute. Memory wall makes data movement the bottleneck.",
		Tips: []string{
			"FeCIM: ~1 fJ/MAC (analog compute in place)",
			"GPU: ~1 pJ/MAC (limited by memory bandwidth)",
			"1000× improvement is the CIM value proposition",
		},
	},
	Throughput: TooltipContent{
		Title:       "Throughput",
		Description: "Inferences per second for the selected workload.",
		Range:       "1-10⁶ inferences/second depending on hardware",
		Physics:     "Limited by: compute (FLOPs), memory bandwidth (GB/s), or latency (data dependencies). CIM removes memory bottleneck for weight-bound workloads.",
		Tips: []string{
			"MNIST is latency-bound (tiny model)",
			"Large models are memory-bandwidth-bound",
			"CIM helps most with large, weight-bound models",
		},
	},
	PowerDraw: TooltipContent{
		Title:       "Power Consumption",
		Description: "Total power draw of the compute system.",
		Range:       "mW (edge accelerator) to 100s of W (GPU/CPU)",
		Physics:     "P = E × ops/sec. Dynamic power: CV²f. Static power: leakage currents. CIM reduces dynamic power but may have leakage from always-on arrays.",
		Tips: []string{
			"GPU: 200-400W (high throughput, high power)",
			"FeCIM target: <10W for edge deployment",
			"Power efficiency = throughput / power",
		},
	},
	CostPerOp: TooltipContent{
		Title:       "Cost per Operation",
		Description: "Operating cost per inference, including energy costs.",
		Range:       "Micro-cents to cents per inference",
		Physics:     "Cost = Energy × electricity_rate. Data center rate ~$0.10/kWh. At scale (10⁹ inferences/day), small per-op savings are significant.",
		Tips: []string{
			"Energy cost dominates at data center scale",
			"10× power reduction = 10× cost reduction",
			"Capital cost (chip price) also matters",
		},
	},
	TRL: TooltipContent{
		Title:       "Technology Readiness Level",
		Description: "NASA's scale for technology maturity, from concept (TRL 1) to production (TRL 9).",
		Range:       "1-9 (FeCIM currently at TRL 4)",
		Physics:     "No physics—project management metric. TRL 4 = Component validated in lab. TRL 6 = System demo. TRL 9 = Operational in production.",
		Tips: []string{
			"GPU/CPU: TRL 9 (production)",
			"FeCIM: TRL 4 (lab demonstrations)",
			"Comparison numbers are projections, not measurements",
		},
	},
	DRAM: TooltipContent{
		Title:       "DRAM (Dynamic RAM)",
		Description: "Main memory technology using capacitor charge storage.",
		Range:       "~10 pJ/bit access, ~10 ns latency, 64 GB capacity typical",
		Physics:     "1T1C: one transistor + one capacitor per bit. Capacitor leaks → requires refresh. High density but volatile.",
		Tips: []string{
			"Used for: main memory, weight storage",
			"Limitation: memory bandwidth is the bottleneck",
			"CIM eliminates DRAM access for weights",
		},
	},
	NAND: TooltipContent{
		Title:       "NAND Flash",
		Description: "Non-volatile storage using floating-gate or charge-trap transistors.",
		Range:       "~100 pJ/bit access, ~100 µs latency, TB capacity",
		Physics:     "Stores charge in insulated gate. MLC/TLC/QLC pack 2-4 bits/cell. Slow writes, limited endurance (~10³-10⁵ cycles).",
		Tips: []string{
			"Used for: SSDs, model storage",
			"Limitation: slow for training, limited cycles",
			"FeCIM has better endurance (~10⁹ cycles)",
		},
	},
	GPU: TooltipContent{
		Title:       "GPU (Graphics Processing Unit)",
		Description: "Massively parallel processor optimized for matrix operations.",
		Range:       "100s of TFLOPS, 200-400W, $10K-$40K",
		Physics:     "SIMD architecture with 1000s of cores. High compute density but memory bandwidth limited (1-2 TB/s). HBM helps but adds cost.",
		Tips: []string{
			"Current standard for AI training/inference",
			"Memory wall limits efficiency for large models",
			"CIM competes on inference, not training",
		},
	},
	CPU: TooltipContent{
		Title:       "CPU (Central Processing Unit)",
		Description: "General-purpose processor optimized for sequential and branching code.",
		Range:       "~1 TFLOPS, 100-300W, $1K-$10K",
		Physics:     "Complex cores with speculation, out-of-order execution. Low utilization for matrix ops. Cache hierarchy dominates energy.",
		Tips: []string{
			"Not optimized for neural networks",
			"Used as baseline (10-100× worse than GPU)",
			"Still used for control flow and I/O",
		},
	},
	FeCIM: TooltipContent{
		Title:       "FeCIM (Ferroelectric Compute-in-Memory)",
		Description: "Emerging technology performing matrix ops directly in memory using ferroelectric devices.",
		Range:       "~1 fJ/MAC (projected), ~10 ns latency, 30+ levels",
		Physics:     "Ferroelectric HfO₂ stores analog weights as polarization. Ohm's law computes dot products. No data movement = massive energy savings.",
		Tips: []string{
			"Key advantage: 1000× energy efficiency vs GPU",
			"Challenge: analog non-idealities limit precision",
			"Status: TRL 4 (lab demonstrations only)",
		},
	},
	DataCenter: TooltipContent{
		Title:       "Data Center Impact",
		Description: "Projected energy and cost savings at data center scale.",
		Range:       "Millions of dollars annual savings (model projection)",
		Physics:     "Scale factor: 10,000 servers × energy reduction × operating hours × electricity cost. Assumes FeCIM projections are achieved.",
		Tips: []string{
			"These are MODEL PROJECTIONS, not measurements",
			"Actual savings depend on achieving TRL 9",
			"Cooling costs also reduced with lower power",
		},
	},
}

// ════════════════════════════════════════════════════════════════════════════════
// MODULE 6: EDA TOOLTIPS
// Chip layout and electronic design automation
// ════════════════════════════════════════════════════════════════════════════════

// EDATooltips contains all tooltips for the EDA module.
var EDATooltips = struct {
	CellLibrary  TooltipContent
	Placement    TooltipContent
	Routing      TooltipContent
	GDS          TooltipContent
	DEF          TooltipContent
	LEF          TooltipContent
	DRC          TooltipContent
	LVS          TooltipContent
	Floorplan    TooltipContent
	PowerGrid    TooltipContent
}{
	CellLibrary: TooltipContent{
		Title:       "Standard Cell Library",
		Description: "Pre-designed logic gates and memory cells for chip design.",
		Range:       "Inverters, NANDs, flip-flops, SRAM cells, etc.",
		Physics:     "Each cell has characterized timing (delay), power, and area. Layout is DRC-clean and LVS-verified. Cells snap to grid for routing.",
		Tips: []string{
			"Foundries provide cell libraries for each process node",
			"Different drive strengths trade speed vs. power",
			"Custom cells (like FeCIM arrays) need manual design",
		},
	},
	Placement: TooltipContent{
		Title:       "Cell Placement",
		Description: "Determining the physical location of each cell on the die.",
		Range:       "Millions of cells on a modern chip",
		Physics:     "Minimize wire length → lower delay, power, congestion. Cells in rows, aligned to grid. Timing-critical paths placed close together.",
		Tips: []string{
			"Placement strongly affects routability",
			"Algorithms: simulated annealing, force-directed",
			"Power/ground rails run horizontally through rows",
		},
	},
	Routing: TooltipContent{
		Title:       "Wire Routing",
		Description: "Connecting placed cells with metal interconnect wires.",
		Range:       "10+ metal layers on advanced nodes",
		Physics:     "Metal resistance causes IR drop and delay. Capacitance causes crosstalk and power. Via resistance adds to path resistance.",
		Tips: []string{
			"Lower metals: local connections (short)",
			"Upper metals: global signals, power grid (thick)",
			"Congestion = not enough routing tracks",
		},
	},
	GDS: TooltipContent{
		Title:       "GDS (GDSII Format)",
		Description: "Industry-standard file format for chip layout data sent to foundry.",
		Range:       "Binary format, can be gigabytes for large chips",
		Physics:     "Polygons on layers representing mask patterns. Each layer becomes a photolithography mask for fabrication.",
		Tips: []string{
			"GDS = final output for tapeout",
			"View with KLayout or Calibre",
			"Includes all layers: diffusion, poly, metals, vias",
		},
	},
	DEF: TooltipContent{
		Title:       "DEF (Design Exchange Format)",
		Description: "Text format describing placed and routed design.",
		Range:       "Human-readable ASCII, megabytes typical",
		Physics:     "Contains: die size, cell instances, positions, routing. Used for interchange between EDA tools.",
		Tips: []string{
			"Pairs with LEF (library description)",
			"Used in OpenLane/OpenROAD flow",
			"Can be visualized in magic or KLayout",
		},
	},
	LEF: TooltipContent{
		Title:       "LEF (Library Exchange Format)",
		Description: "Describes abstract cell views for placement and routing.",
		Range:       "One file per library, kilobytes",
		Physics:     "Cell size, pin locations, routing blockages. Abstract view hides internal transistors—just shows interface for routing.",
		Tips: []string{
			"Tech LEF: layer definitions, via rules",
			"Cell LEF: pin shapes, obstructions",
			"Routers use LEF, not full GDS",
		},
	},
	DRC: TooltipContent{
		Title:       "Design Rule Check",
		Description: "Verifies layout meets manufacturing constraints.",
		Range:       "100s of rules for advanced nodes",
		Physics:     "Rules prevent: shorts (spacing), opens (width), lithography failures (density). Derived from process capability.",
		Tips: []string{
			"Must be 0 errors before tapeout",
			"Rules: min width, min spacing, enclosure",
			"Tools: Calibre, Assura, KLayout",
		},
	},
	LVS: TooltipContent{
		Title:       "Layout vs. Schematic",
		Description: "Verifies layout matches intended circuit connectivity.",
		Range:       "Must match exactly (0 errors, 0 warnings)",
		Physics:     "Extracts transistors and connections from layout polygons. Compares against netlist from synthesis. Catches shorts, opens, missing devices.",
		Tips: []string{
			"LVS errors = serious bugs",
			"Run after DRC clean",
			"Labels help match ports between views",
		},
	},
	Floorplan: TooltipContent{
		Title:       "Floorplanning",
		Description: "High-level arrangement of major blocks on the die.",
		Range:       "Die size: mm², block count: 10s-100s",
		Physics:     "Determines: aspect ratio, I/O ring, power grid topology, block shapes. Affects routability and timing closure.",
		Tips: []string{
			"Memory arrays often placed as hard macros",
			"I/O pads typically on die periphery",
			"Leave channels for routing between blocks",
		},
	},
	PowerGrid: TooltipContent{
		Title:       "Power Grid",
		Description: "Network of metal stripes delivering VDD and GND to all cells.",
		Range:       "Wide stripes on upper metals, straps to lower metals",
		Physics:     "IR drop = I × R_grid. Must keep voltage within spec everywhere. Hot spots need extra straps. Electromigration limits current density.",
		Tips: []string{
			"Power grid designed during floorplanning",
			"Analyze IR drop with static/dynamic tools",
			"FeCIM arrays need careful power planning",
		},
	},
}

// ════════════════════════════════════════════════════════════════════════════════
// TOOLTIP WIDGETS
// Helper widgets for displaying tooltips in Fyne
// ════════════════════════════════════════════════════════════════════════════════

// HoverTooltip wraps any widget to show a tooltip on hover.
type HoverTooltip struct {
	widget.BaseWidget
	content     fyne.CanvasObject
	tooltipText string
	tooltip     *widget.PopUp
	window      fyne.Window
	maxWidth    float32
}

// NewHoverTooltip creates a widget that shows tooltip on hover.
// If window is nil, tooltips are disabled.
func NewHoverTooltip(content fyne.CanvasObject, tooltip string, window fyne.Window) *HoverTooltip {
	h := &HoverTooltip{
		content:     content,
		tooltipText: tooltip,
		window:      window,
		maxWidth:    400,
	}
	h.ExtendBaseWidget(h)
	return h
}

// NewHoverTooltipFromContent creates a tooltip from structured content.
func NewHoverTooltipFromContent(content fyne.CanvasObject, tc TooltipContent, window fyne.Window) *HoverTooltip {
	return NewHoverTooltip(content, tc.Format(), window)
}

// SetMaxWidth sets the maximum tooltip width.
func (h *HoverTooltip) SetMaxWidth(width float32) {
	h.maxWidth = width
}

// MouseIn shows the tooltip.
func (h *HoverTooltip) MouseIn(ev *desktop.MouseEvent) {
	h.showTooltip()
}

// MouseMoved satisfies desktop.Hoverable.
func (h *HoverTooltip) MouseMoved(ev *desktop.MouseEvent) {}

// MouseOut hides the tooltip.
func (h *HoverTooltip) MouseOut() {
	h.hideTooltip()
}

func (h *HoverTooltip) showTooltip() {
	if h.tooltipText == "" || h.window == nil {
		return
	}
	if h.tooltip != nil {
		h.tooltip.Hide()
	}

	label := widget.NewLabel(h.tooltipText)
	label.Wrapping = fyne.TextWrapWord

	// Constrain width
	scroll := container.NewScroll(label)
	scroll.SetMinSize(fyne.NewSize(h.maxWidth, 0))

	h.tooltip = widget.NewPopUp(
		container.NewPadded(scroll),
		h.window.Canvas(),
	)

	pos := fyne.CurrentApp().Driver().AbsolutePositionForObject(h)
	h.tooltip.ShowAtPosition(fyne.NewPos(pos.X, pos.Y+h.Size().Height+4))
}

func (h *HoverTooltip) hideTooltip() {
	if h.tooltip == nil {
		return
	}
	h.tooltip.Hide()
	h.tooltip = nil
}

// CreateRenderer implements fyne.Widget.
func (h *HoverTooltip) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(h.content)
}

// InfoButton creates a small info icon button that shows a tooltip on click.
func InfoButton(tc TooltipContent, window fyne.Window) *widget.Button {
	return widget.NewButtonWithIcon("", fyne.NewStaticResource("info", nil), func() {
		ShowTooltipDialog(tc, window)
	})
}

// ShowTooltipDialog shows a modal dialog with the full tooltip content.
func ShowTooltipDialog(tc TooltipContent, window fyne.Window) {
	if window == nil {
		return
	}

	titleLabel := widget.NewLabelWithStyle(tc.Title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	descLabel := widget.NewLabel(tc.Description)
	descLabel.Wrapping = fyne.TextWrapWord

	content := container.NewVBox(titleLabel, widget.NewSeparator(), descLabel)

	if tc.Range != "" {
		rangeLabel := widget.NewLabel("Range: " + tc.Range)
		rangeLabel.Wrapping = fyne.TextWrapWord
		content.Add(rangeLabel)
	}

	if tc.Physics != "" {
		physicsTitle := widget.NewLabelWithStyle("🔬 Physics", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		physicsLabel := widget.NewLabel(tc.Physics)
		physicsLabel.Wrapping = fyne.TextWrapWord
		content.Add(widget.NewSeparator())
		content.Add(physicsTitle)
		content.Add(physicsLabel)
	}

	if len(tc.Tips) > 0 {
		tipsTitle := widget.NewLabelWithStyle("💡 Tips", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		content.Add(widget.NewSeparator())
		content.Add(tipsTitle)
		for _, tip := range tc.Tips {
			tipLabel := widget.NewLabel("• " + tip)
			tipLabel.Wrapping = fyne.TextWrapWord
			content.Add(tipLabel)
		}
	}

	scroll := container.NewScroll(content)
	scroll.SetMinSize(fyne.NewSize(450, 300))

	popup := widget.NewModalPopUp(
		container.NewBorder(nil,
			widget.NewButton("Close", nil),
			nil, nil,
			container.NewPadded(scroll),
		),
		window.Canvas(),
	)

	// Wire up close button
	closeBtn := popup.Content.(*fyne.Container).Objects[1].(*widget.Button)
	closeBtn.OnTapped = func() { popup.Hide() }

	popup.Show()
}

// AddLabelTooltip wraps a label with hover tooltip functionality.
func AddLabelTooltip(label *widget.Label, tc TooltipContent, window fyne.Window) fyne.CanvasObject {
	return NewHoverTooltipFromContent(label, tc, window)
}

// AddSliderTooltip creates a labeled slider with tooltip info button.
func AddSliderTooltip(slider *widget.Slider, label *widget.Label, tc TooltipContent, window fyne.Window) fyne.CanvasObject {
	infoBtn := widget.NewButtonWithIcon("", nil, func() {
		ShowTooltipDialog(tc, window)
	})
	infoBtn.Importance = widget.LowImportance

	// Minimal info button text
	infoBtn.SetText("ⓘ")

	return container.NewBorder(nil, nil, container.NewHBox(label, infoBtn), nil, slider)
}
