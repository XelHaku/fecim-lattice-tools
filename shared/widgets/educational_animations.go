// Package widgets provides shared UI components for FeCIM visualizers.
// educational_animations.go provides pre-built educational animation sequences
// for demonstrating core FeCIM concepts.
package widgets

import (
	"fmt"
	"time"
)

// AnimationPreset represents a named collection of animation frames.
type AnimationPreset struct {
	Name        string
	Description string
	Module      string // Which module this is for
	Difficulty  TutorialLevel
	Frames      []AnimationFrame
	Tags        []string
}

// GetHysteresisAnimations returns educational animations for the hysteresis module.
func GetHysteresisAnimations() []AnimationPreset {
	return []AnimationPreset{
		{
			Name:        "hysteresis-loop-basics",
			Description: "Understanding the P-E Hysteresis Loop",
			Module:      "module1-hysteresis",
			Difficulty:  LevelBeginner,
			Tags:        []string{"physics", "fundamentals", "ferroelectric"},
			Frames: []AnimationFrame{
				{
					Title:    "1. What is Ferroelectric?",
					Content:  "Ferroelectric materials have electric dipoles that can be switched by an electric field, similar to how ferromagnetic materials respond to magnetic fields.",
					Duration: 4 * time.Second,
				},
				{
					Title:    "2. The Hysteresis Loop",
					Content:  "When we apply an electric field (E), the polarization (P) follows a characteristic S-shaped curve. This path-dependent behavior is called hysteresis.",
					Duration: 4 * time.Second,
				},
				{
					Title:    "3. Saturation Polarization (Ps)",
					Content:  "At high fields, all dipoles align → maximum polarization Ps ≈ 45 µC/cm² for HZO.",
					Highlight: "ps-marker",
					Duration:  3 * time.Second,
				},
				{
					Title:    "4. Remanent Polarization (Pr)",
					Content:  "When E returns to zero, polarization REMAINS at ±Pr. This is the 'memory' effect! Pr ≈ 35 µC/cm²",
					Highlight: "pr-marker",
					Duration:  4 * time.Second,
				},
				{
					Title:    "5. Coercive Field (Ec)",
					Content:  "To switch polarity, we must overcome Ec ≈ 1-2 MV/cm. This is the 'switching threshold'.",
					Highlight: "ec-marker",
					Duration:  3 * time.Second,
				},
				{
					Title:    "6. Why This Matters",
					Content:  "The bistable states (+Pr and -Pr) can represent binary data. The intermediate states enable analog computing!",
					Duration: 4 * time.Second,
				},
			},
		},
		{
			Name:        "analog-levels-demo",
			Description: "How 30 Analog Levels Enable Compute-in-Memory",
			Module:      "module1-hysteresis",
			Difficulty:  LevelIntermediate,
			Tags:        []string{"analog", "CIM", "levels"},
			Frames: []AnimationFrame{
				{
					Title:    "1. Beyond Binary",
					Content:  "Traditional memory stores 1 or 0. FeCIM stores 30+ discrete levels using partial polarization switching.",
					Duration: 4 * time.Second,
				},
				{
					Title:    "2. Information Density",
					Content:  "30 levels = log₂(30) ≈ 4.9 bits per cell. This is 5× more efficient than binary!",
					Duration: 4 * time.Second,
				},
				{
					Title:    "3. How Levels Are Set",
					Content:  "Incremental Step Pulse Programming (ISPP): Apply calibrated voltage pulses to reach target polarization.",
					Duration: 4 * time.Second,
				},
				{
					Title:    "4. Level Separation",
					Content:  "Each level is separated by ~2-3 µC/cm². Signal-to-noise ratio must exceed this for reliable readout.",
					Duration: 4 * time.Second,
				},
				{
					Title:    "5. Compute-in-Memory",
					Content:  "These analog levels act as WEIGHTS in neural network computation. Multiplication happens during memory read!",
					Duration: 5 * time.Second,
				},
				{
					Title:    "6. Energy Efficiency",
					Content:  "By computing where data lives, we avoid the 'memory wall' that limits traditional processors.",
					Duration: 4 * time.Second,
				},
			},
		},
		{
			Name:        "preisach-model-explained",
			Description: "The Preisach Model: Capturing Complex Hysteresis",
			Module:      "module1-hysteresis",
			Difficulty:  LevelAdvanced,
			Tags:        []string{"physics", "model", "preisach"},
			Frames: []AnimationFrame{
				{
					Title:    "1. Why Model Hysteresis?",
					Content:  "Real ferroelectrics have complex switching behavior. The Preisach model captures this accurately.",
					Duration: 4 * time.Second,
				},
				{
					Title:    "2. Hysterons",
					Content:  "The model treats the material as a collection of microscopic dipole units called 'hysterons'.",
					Duration: 4 * time.Second,
				},
				{
					Title:    "3. Switching Thresholds",
					Content:  "Each hysteron has two thresholds: α (switch up) and β (switch down). These define the Preisach plane.",
					Duration: 5 * time.Second,
				},
				{
					Title:    "4. Distribution Function",
					Content:  "The Preisach distribution µ(α,β) describes how many hysterons have each threshold pair.",
					Duration: 4 * time.Second,
				},
				{
					Title:    "5. Memory Effect",
					Content:  "The wiping-out property: only extreme field values are 'remembered'. This explains minor loop behavior.",
					Duration: 5 * time.Second,
				},
				{
					Title:    "6. First-Order Reversal Curves",
					Content:  "FORCs allow extracting µ(α,β) from experiments. This is the bridge between theory and measurement.",
					Duration: 5 * time.Second,
				},
			},
		},
	}
}

// GetCrossbarAnimations returns educational animations for the crossbar module.
func GetCrossbarAnimations() []AnimationPreset {
	return []AnimationPreset{
		{
			Name:        "crossbar-array-basics",
			Description: "Introduction to Crossbar Array Architecture",
			Module:      "module2-crossbar",
			Difficulty:  LevelBeginner,
			Tags:        []string{"architecture", "crossbar", "fundamentals"},
			Frames: []AnimationFrame{
				{
					Title:    "1. What is a Crossbar?",
					Content:  "A crossbar array is a grid of horizontal word lines (rows) and vertical bit lines (columns), with memory cells at each intersection.",
					Duration: 4 * time.Second,
				},
				{
					Title:    "2. Storing Weights",
					Content:  "Each intersection stores a conductance value G = 1/R. In FeCIM, this is the programmed polarization state.",
					Duration: 4 * time.Second,
				},
				{
					Title:    "3. Matrix-Vector Multiplication",
					Content:  "Apply voltages V to rows. By Ohm's law, current I = G×V flows through each cell. Columns sum currents!",
					Duration: 5 * time.Second,
				},
				{
					Title:    "4. Parallel Computation",
					Content:  "All multiplications happen simultaneously in a single cycle. This is O(1) instead of O(n²)!",
					Duration: 4 * time.Second,
				},
				{
					Title:    "5. Reading Results",
					Content:  "Sense amplifiers at column ends measure total current. Each column gives one output value.",
					Duration: 4 * time.Second,
				},
				{
					Title:    "6. Neural Network Connection",
					Content:  "This MVM operation is the core of neural network inference. Crossbar = hardware matrix multiplication!",
					Duration: 4 * time.Second,
				},
			},
		},
		{
			Name:        "ir-drop-explained",
			Description: "Understanding IR Drop Non-Ideality",
			Module:      "module2-crossbar",
			Difficulty:  LevelIntermediate,
			Tags:        []string{"non-ideality", "IR-drop", "physics"},
			Frames: []AnimationFrame{
				{
					Title:    "1. The Problem",
					Content:  "Real metal wires have resistance. As current flows, voltage drops along the lines (V = IR).",
					Duration: 4 * time.Second,
				},
				{
					Title:    "2. Position Dependence",
					Content:  "Cells far from voltage sources see lower effective voltage. This creates systematic errors.",
					Duration: 4 * time.Second,
				},
				{
					Title:    "3. Worst Case",
					Content:  "When many cells are active (all high conductance), IR drop is maximized. Corner cells suffer most.",
					Duration: 4 * time.Second,
				},
				{
					Title:    "4. Impact on Accuracy",
					Content:  "Without compensation, large arrays (>128×128) can see 5-10% accuracy degradation.",
					Duration: 4 * time.Second,
				},
				{
					Title:    "5. Mitigation: Line Resistance",
					Content:  "Solution 1: Use lower-resistance metals (Cu, Al with wider traces) or multi-layer routing.",
					Duration: 4 * time.Second,
				},
				{
					Title:    "6. Mitigation: Compensation",
					Content:  "Solution 2: Software compensation maps that pre-adjust weights based on position.",
					Duration: 4 * time.Second,
				},
			},
		},
		{
			Name:        "sneak-path-demo",
			Description: "Sneak Path Currents in Passive Arrays",
			Module:      "module2-crossbar",
			Difficulty:  LevelAdvanced,
			Tags:        []string{"non-ideality", "sneak-path", "physics"},
			Frames: []AnimationFrame{
				{
					Title:    "1. The Sneak Path Problem",
					Content:  "In passive arrays (no transistor per cell), current can flow through unintended paths via neighboring cells.",
					Duration: 4 * time.Second,
				},
				{
					Title:    "2. Visualizing Sneak Paths",
					Content:  "When reading cell (i,j): current can go i→k→j through any cell (i,k) and (k,j) in the same row/column.",
					Duration: 5 * time.Second,
				},
				{
					Title:    "3. Impact",
					Content:  "Sneak currents add to sensed current, corrupting the read value. Worse with more high-conductance neighbors.",
					Duration: 4 * time.Second,
				},
				{
					Title:    "4. 1T1R Solution",
					Content:  "Add one transistor per cell (1T1R). The transistor acts as a switch, blocking sneak paths when off.",
					Duration: 4 * time.Second,
				},
				{
					Title:    "5. Selector Devices",
					Content:  "Alternative: Use self-rectifying cells or add selector devices (diodes, OTS) to block reverse current.",
					Duration: 4 * time.Second,
				},
				{
					Title:    "6. Design Tradeoffs",
					Content:  "1T1R: More area, but cleaner signals. Selector: Smaller, but adds complexity. Choose based on application.",
					Duration: 4 * time.Second,
				},
			},
		},
	}
}

// GetMNISTAnimations returns educational animations for the MNIST/neural network module.
func GetMNISTAnimations() []AnimationPreset {
	return []AnimationPreset{
		{
			Name:        "mnist-inference-flow",
			Description: "How a Digit is Recognized",
			Module:      "module3-mnist",
			Difficulty:  LevelBeginner,
			Tags:        []string{"inference", "neural-network", "MNIST"},
			Frames: []AnimationFrame{
				{
					Title:    "1. Input: A Handwritten Digit",
					Content:  "MNIST images are 28×28 pixels in grayscale. We'll classify this as 0-9.",
					Duration: 3 * time.Second,
				},
				{
					Title:    "2. Flatten to Vector",
					Content:  "The 2D image becomes a 784-element vector. Each pixel value is 0.0-1.0.",
					Duration: 3 * time.Second,
				},
				{
					Title:    "3. Layer 1: Hidden Layer",
					Content:  "Matrix multiplication: 784 inputs × weights → 128 hidden neurons. Add bias, apply ReLU.",
					Duration: 4 * time.Second,
				},
				{
					Title:    "4. In FeCIM Hardware",
					Content:  "This MVM happens in the crossbar! 784 row inputs, 128 column outputs. One cycle!",
					Duration: 4 * time.Second,
				},
				{
					Title:    "5. Layer 2: Output Layer",
					Content:  "128 hidden → 10 output classes. Each output is a 'score' for digits 0-9.",
					Duration: 4 * time.Second,
				},
				{
					Title:    "6. Prediction",
					Content:  "Softmax converts scores to probabilities. Highest probability = predicted digit!",
					Duration: 4 * time.Second,
				},
			},
		},
		{
			Name:        "quantization-effects",
			Description: "How Limited Levels Affect Accuracy",
			Module:      "module3-mnist",
			Difficulty:  LevelIntermediate,
			Tags:        []string{"quantization", "accuracy", "analog"},
			Frames: []AnimationFrame{
				{
					Title:    "1. The Quantization Challenge",
					Content:  "Digital training uses 32-bit floats. FeCIM has only 30 discrete levels. How do we bridge this gap?",
					Duration: 4 * time.Second,
				},
				{
					Title:    "2. Weight Distribution",
					Content:  "Trained weights follow a bell curve. We map this continuous range to 30 discrete levels.",
					Duration: 4 * time.Second,
				},
				{
					Title:    "3. Quantization Error",
					Content:  "Each weight gets rounded to nearest level. Error = weight - quantized_weight.",
					Duration: 4 * time.Second,
				},
				{
					Title:    "4. Accumulated Error",
					Content:  "Errors accumulate through layers. With many weights, small errors become significant.",
					Duration: 4 * time.Second,
				},
				{
					Title:    "5. Binary vs 30 Levels",
					Content:  "Binary (2 levels): ~85% MNIST accuracy. 30 levels: ~97% accuracy. The difference is huge!",
					Duration: 5 * time.Second,
				},
				{
					Title:    "6. Quantization-Aware Training",
					Content:  "Solution: Train with quantization in the loop. Network learns to be robust to level constraints.",
					Duration: 4 * time.Second,
				},
			},
		},
		{
			Name:        "cim-vs-traditional",
			Description: "Compute-in-Memory vs Traditional Architecture",
			Module:      "module3-mnist",
			Difficulty:  LevelIntermediate,
			Tags:        []string{"CIM", "comparison", "efficiency"},
			Frames: []AnimationFrame{
				{
					Title:    "1. The Memory Wall",
					Content:  "Traditional: CPU fetches weights from memory. Memory bandwidth limits performance.",
					Duration: 4 * time.Second,
				},
				{
					Title:    "2. Energy Breakdown",
					Content:  "Data movement costs ~100× more energy than computation. This dominates AI workloads!",
					Duration: 4 * time.Second,
				},
				{
					Title:    "3. Compute-in-Memory Solution",
					Content:  "FeCIM: Weights stay in memory. Input voltages trigger immediate analog computation.",
					Duration: 4 * time.Second,
				},
				{
					Title:    "4. Energy Comparison",
					Content:  "GPU: ~100 pJ/MAC. FeCIM: ~1 pJ/MAC. That's 100× energy efficiency!",
					Duration: 4 * time.Second,
				},
				{
					Title:    "5. Throughput",
					Content:  "N×N matrix multiply: Traditional O(N²) cycles. Crossbar: O(1) cycles!",
					Duration: 4 * time.Second,
				},
				{
					Title:    "6. The Tradeoff",
					Content:  "CIM trades precision for efficiency. With 30 levels, we keep high accuracy while gaining 100× efficiency.",
					Duration: 5 * time.Second,
				},
			},
		},
	}
}

// GetCircuitsAnimations returns educational animations for the circuits module.
func GetCircuitsAnimations() []AnimationPreset {
	return []AnimationPreset{
		{
			Name:        "adc-dac-basics",
			Description: "Analog-Digital Converters in FeCIM",
			Module:      "module4-circuits",
			Difficulty:  LevelBeginner,
			Tags:        []string{"ADC", "DAC", "circuits"},
			Frames: []AnimationFrame{
				{
					Title:    "1. The Analog-Digital Interface",
					Content:  "Digital inputs must be converted to analog voltages (DAC). Analog outputs must be digitized (ADC).",
					Duration: 4 * time.Second,
				},
				{
					Title:    "2. DAC: Digital-to-Analog",
					Content:  "Converts N-bit digital input to one of 2^N voltage levels. Applied to crossbar word lines.",
					Duration: 4 * time.Second,
				},
				{
					Title:    "3. ADC: Analog-to-Digital",
					Content:  "Measures column currents and converts to digital. Precision determines output quality.",
					Duration: 4 * time.Second,
				},
				{
					Title:    "4. Bit Precision Tradeoffs",
					Content:  "More bits = better precision but more area, power, and latency. 8-bit is common for inference.",
					Duration: 4 * time.Second,
				},
				{
					Title:    "5. Flash ADC",
					Content:  "Fastest type: 2^N comparators in parallel. Used when speed matters most.",
					Duration: 4 * time.Second,
				},
				{
					Title:    "6. SAR ADC",
					Content:  "Successive approximation: binary search over N cycles. Area-efficient, commonly used.",
					Duration: 4 * time.Second,
				},
			},
		},
		{
			Name:        "sense-amplifier-design",
			Description: "Reading Tiny Currents with Sense Amplifiers",
			Module:      "module4-circuits",
			Difficulty:  LevelAdvanced,
			Tags:        []string{"sense-amp", "circuits", "design"},
			Frames: []AnimationFrame{
				{
					Title:    "1. The Challenge",
					Content:  "Crossbar column currents can be nanoamps. How do we measure such tiny signals reliably?",
					Duration: 4 * time.Second,
				},
				{
					Title:    "2. Transimpedance Amplifier",
					Content:  "Converts current to voltage: V_out = I_in × R_feedback. Gain set by feedback resistor.",
					Duration: 4 * time.Second,
				},
				{
					Title:    "3. Noise Considerations",
					Content:  "Thermal noise, shot noise, flicker noise all compete with signal. SNR must be high enough to distinguish levels.",
					Duration: 4 * time.Second,
				},
				{
					Title:    "4. Bandwidth vs Noise",
					Content:  "Higher bandwidth = faster but noisier. Optimize for just enough bandwidth to meet timing.",
					Duration: 4 * time.Second,
				},
				{
					Title:    "5. Column Capacitance",
					Content:  "Long bit lines have parasitic capacitance. This sets integration time and affects speed.",
					Duration: 4 * time.Second,
				},
				{
					Title:    "6. Design Example",
					Content:  "128 rows × 1µA/cell = 128µA max. With 100mV swing needed, R_feedback ≈ 800Ω.",
					Duration: 5 * time.Second,
				},
			},
		},
	}
}

// GetAllAnimations returns all available educational animations.
func GetAllAnimations() []AnimationPreset {
	var all []AnimationPreset
	all = append(all, GetHysteresisAnimations()...)
	all = append(all, GetCrossbarAnimations()...)
	all = append(all, GetMNISTAnimations()...)
	all = append(all, GetCircuitsAnimations()...)
	return all
}

// GetAnimationsByModule returns animations for a specific module.
func GetAnimationsByModule(module string) []AnimationPreset {
	var result []AnimationPreset
	for _, anim := range GetAllAnimations() {
		if anim.Module == module {
			result = append(result, anim)
		}
	}
	return result
}

// GetAnimationsByTag returns animations with a specific tag.
func GetAnimationsByTag(tag string) []AnimationPreset {
	var result []AnimationPreset
	for _, anim := range GetAllAnimations() {
		for _, t := range anim.Tags {
			if t == tag {
				result = append(result, anim)
				break
			}
		}
	}
	return result
}

// GetAnimationsByLevel returns animations at or below a difficulty level.
func GetAnimationsByLevel(maxLevel TutorialLevel) []AnimationPreset {
	var result []AnimationPreset
	for _, anim := range GetAllAnimations() {
		if anim.Difficulty <= maxLevel {
			result = append(result, anim)
		}
	}
	return result
}

// AnimationPresetToController creates an AnimationController from a preset.
func AnimationPresetToController(preset AnimationPreset) *AnimationController {
	return NewAnimationController(preset.Frames)
}

// CreateAnimationFramesFromTutorial converts tutorial steps to animation frames.
func CreateAnimationFramesFromTutorial(steps []TutorialStep) []AnimationFrame {
	frames := make([]AnimationFrame, len(steps))
	for i, step := range steps {
		frames[i] = AnimationFrame{
			Title:     step.Title,
			Content:   step.Explanation,
			Highlight: step.HighlightElement,
			Duration:  step.Duration,
		}
	}
	return frames
}

// FrameCounter formats the current frame position.
func FrameCounter(current, total int) string {
	return fmt.Sprintf("Frame %d of %d", current+1, total)
}

// ProgressPercentage calculates progress as a percentage.
func ProgressPercentage(current, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(current) / float64(total) * 100
}
