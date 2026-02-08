// Package widgets provides shared UI components for FeCIM visualizers.
// interactive_tutorials.go provides pre-built interactive tutorial sequences
// with quizzes, user prompts, and hands-on exercises for each module.
package widgets

import (
	"time"
)

// InteractiveTutorial represents a complete interactive learning experience.
type InteractiveTutorial struct {
	ID          string
	Name        string
	Description string
	Module      string
	Difficulty  TutorialLevel
	Duration    time.Duration // Estimated duration
	Steps       []TutorialStep
	Prerequisites []string // IDs of tutorials that should be completed first
	LearningGoals []string
}

// GetHysteresisTutorials returns interactive tutorials for Module 1.
func GetHysteresisTutorials() []InteractiveTutorial {
	return []InteractiveTutorial{
		{
			ID:          "hys-intro",
			Name:        "Introduction to Ferroelectricity",
			Description: "Learn the fundamentals of ferroelectric materials and why they're perfect for memory applications.",
			Module:      "module1-hysteresis",
			Difficulty:  LevelBeginner,
			Duration:    10 * time.Minute,
			LearningGoals: []string{
				"Understand what makes a material ferroelectric",
				"Know the key parameters: Ps, Pr, Ec",
				"Explain why polarization persists without power",
			},
			Steps: []TutorialStep{
				{
					Title:       "Welcome to Ferroelectrics!",
					Explanation: "Ferroelectric materials are the foundation of FeCIM technology. In this tutorial, you'll learn why these materials can 'remember' information without power.",
					Duration:    5 * time.Second,
					DifficultyLevel: LevelBeginner,
				},
				{
					Title:       "Electric Dipoles",
					Explanation: "Inside ferroelectric crystals, atoms are arranged asymmetrically, creating tiny electric dipoles. Each dipole has a positive and negative end.\n\n🔬 In HfO₂-ZrO₂, oxygen atoms shift position to create these dipoles.",
					Duration:    6 * time.Second,
					HighlightElement: "dipole-diagram",
					DifficultyLevel: LevelBeginner,
				},
				{
					Title:       "Collective Switching",
					Explanation: "When we apply a strong enough electric field, ALL dipoles in a region flip together. This is the key insight - it's collective behavior, not individual atoms.\n\n⚡ The field needed is called the Coercive Field (Ec).",
					Duration:    6 * time.Second,
					Action: func() {
						// Animation trigger would go here
					},
					DifficultyLevel: LevelBeginner,
				},
				{
					Title:       "Memory Without Power",
					Explanation: "Once flipped, the dipoles STAY in their new orientation even after the field is removed. This 'remanent' state is stable for years!\n\n🔋 This is why ferroelectric memory is non-volatile.",
					Duration:    6 * time.Second,
					DifficultyLevel: LevelBeginner,
				},
				{
					Title:       "Quick Check ✓",
					Explanation: "Time to test your understanding!",
					Duration:    0, // Wait for quiz
					QuizQuestion: &QuizQuestion{
						Question:    "What happens to ferroelectric polarization when power is removed?",
						Options:     []string{"It disappears immediately", "It remains stable", "It gradually fades over seconds", "It reverses direction"},
						CorrectIdx:  1,
						Explanation: "Correct! Ferroelectric polarization is remanent - it stays in place without power, making it perfect for non-volatile memory.",
					},
					DifficultyLevel: LevelBeginner,
				},
				{
					Title:       "Saturation Polarization (Ps)",
					Explanation: "When we apply a very strong field, we reach maximum polarization - called Saturation Polarization (Ps).\n\n📊 For HZO superlattice: Ps ≈ 45 µC/cm²",
					Duration:    5 * time.Second,
					HighlightElement: "ps-value",
					DifficultyLevel: LevelBeginner,
				},
				{
					Title:       "Remanent Polarization (Pr)",
					Explanation: "At zero field, polarization settles to the Remanent Polarization (Pr). This is slightly less than Ps due to relaxation.\n\n📊 For HZO: Pr ≈ 35 µC/cm²\n\n💡 Pr determines the signal strength when reading memory!",
					Duration:    6 * time.Second,
					HighlightElement: "pr-value",
					DifficultyLevel: LevelBeginner,
				},
				{
					Title:       "Coercive Field (Ec)",
					Explanation: "The Coercive Field is the minimum field needed to START switching. It's the 'activation energy' for polarization reversal.\n\n📊 For HZO: Ec ≈ 1-2 MV/cm\n\n⚡ Lower Ec = easier to write = lower power!",
					Duration:    6 * time.Second,
					HighlightElement: "ec-value",
					DifficultyLevel: LevelBeginner,
				},
				{
					Title:       "Quick Check ✓",
					Explanation: "Let's verify you understand the key parameters!",
					Duration:    0,
					QuizQuestion: &QuizQuestion{
						Question:    "Which parameter determines how easy it is to write to ferroelectric memory?",
						Options:     []string{"Saturation Polarization (Ps)", "Remanent Polarization (Pr)", "Coercive Field (Ec)", "Film Thickness"},
						CorrectIdx:  2,
						Explanation: "Correct! Lower Coercive Field (Ec) means less voltage is needed to switch the polarization, reducing write energy.",
					},
					DifficultyLevel: LevelBeginner,
				},
				{
					Title:       "🎉 Tutorial Complete!",
					Explanation: "You now understand the basics of ferroelectric materials!\n\n✅ Dipoles create stable polarization states\n✅ Ps, Pr, Ec are the key parameters\n✅ Non-volatility comes from stable remanent states\n\n➡️ Next: Try the 'P-E Hysteresis Loop' tutorial!",
					Duration:    8 * time.Second,
					DifficultyLevel: LevelBeginner,
				},
			},
		},
		{
			ID:          "hys-loop",
			Name:        "Understanding the P-E Loop",
			Description: "Explore the hysteresis loop interactively and understand its shape.",
			Module:      "module1-hysteresis",
			Difficulty:  LevelBeginner,
			Duration:    15 * time.Minute,
			Prerequisites: []string{"hys-intro"},
			LearningGoals: []string{
				"Read and interpret P-E hysteresis loops",
				"Understand why the loop has its characteristic S-shape",
				"Identify key points on the loop: Ps, Pr, Ec, -Ec",
			},
			Steps: []TutorialStep{
				{
					Title:       "The Hysteresis Loop",
					Explanation: "The P-E loop shows how Polarization (P) responds to Electric field (E). It's the 'fingerprint' of a ferroelectric material.\n\n🎯 Goal: Learn to read this plot and extract key information.",
					Duration:    5 * time.Second,
					DifficultyLevel: LevelBeginner,
				},
				{
					Title:       "Try It: Start from Zero",
					Explanation: "We'll start with an unpoled material (P ≈ 0). Watch what happens as we increase the electric field.\n\n👆 Use the E-field slider to increase the field slowly.",
					Duration:    0, // Wait for user
					UserPrompt:  "Drag the E-field slider to +2 MV/cm",
					HighlightElement: "e-field-slider",
					DifficultyLevel: LevelBeginner,
				},
				{
					Title:       "Approaching Saturation",
					Explanation: "As E increases, more dipoles align. Notice how the curve starts to flatten - we're approaching saturation!\n\n📈 The S-shape comes from the statistical distribution of switching thresholds.",
					Duration:    5 * time.Second,
					DifficultyLevel: LevelBeginner,
				},
				{
					Title:       "Try It: Remove the Field",
					Explanation: "Now let's see the 'memory' effect. Bring the field back to zero.\n\n👆 Drag the slider back to E = 0",
					Duration:    0,
					UserPrompt:  "Reduce E-field to zero",
					HighlightElement: "e-field-slider",
					DifficultyLevel: LevelBeginner,
				},
				{
					Title:       "Remanent Polarization",
					Explanation: "Notice: P didn't go back to zero! It stayed at +Pr. This is the memory!\n\n🔋 The material 'remembers' it was polarized positively.",
					Duration:    5 * time.Second,
					HighlightElement: "p-value-display",
					DifficultyLevel: LevelBeginner,
				},
				{
					Title:       "Try It: Reverse the Field",
					Explanation: "Now apply a negative field. Watch for the coercive point where polarization starts to switch.\n\n👆 Drag the slider to negative values",
					Duration:    0,
					UserPrompt:  "Apply negative field (E < 0)",
					DifficultyLevel: LevelBeginner,
				},
				{
					Title:       "The Coercive Point",
					Explanation: "At E = -Ec, the polarization crosses zero and begins switching to negative. This is the coercive field!\n\n⚡ In real memory, we apply pulses that exceed Ec to write data.",
					Duration:    6 * time.Second,
					HighlightElement: "ec-point",
					DifficultyLevel: LevelBeginner,
				},
				{
					Title:       "Complete the Loop",
					Explanation: "Continue to negative saturation, then sweep back positive. The loop closes!\n\n🔄 The area inside the loop represents energy dissipated per cycle.",
					Duration:    6 * time.Second,
					DifficultyLevel: LevelBeginner,
				},
				{
					Title:       "Quick Check ✓",
					Explanation: "Test your loop-reading skills!",
					Duration:    0,
					QuizQuestion: &QuizQuestion{
						Question:    "Looking at a P-E loop, where would you find the remanent polarization Pr?",
						Options:     []string{"At the maximum E-field", "Where the loop crosses P = 0", "Where the loop crosses E = 0", "At the center of the loop"},
						CorrectIdx:  2,
						Explanation: "Correct! Pr is the polarization value when E = 0 (on the P-axis). There are two Pr points: +Pr and -Pr.",
					},
					DifficultyLevel: LevelBeginner,
				},
				{
					Title:       "🎉 Tutorial Complete!",
					Explanation: "You can now read P-E hysteresis loops!\n\n✅ The loop shows path-dependent (hysteretic) behavior\n✅ Pr is found at E = 0, Ec is found at P = 0\n✅ The S-shape comes from statistical dipole switching\n\n➡️ Next: Try the 'Analog Levels' tutorial!",
					Duration:    8 * time.Second,
					DifficultyLevel: LevelBeginner,
				},
			},
		},
		{
			ID:          "hys-analog",
			Name:        "Programming Analog Levels",
			Description: "Learn how 30 discrete polarization levels enable analog computing.",
			Module:      "module1-hysteresis",
			Difficulty:  LevelIntermediate,
			Duration:    15 * time.Minute,
			Prerequisites: []string{"hys-loop"},
			LearningGoals: []string{
				"Understand partial polarization switching",
				"Program specific polarization levels using pulse trains",
				"Calculate information density (bits per cell)",
			},
			Steps: []TutorialStep{
				{
					Title:       "Beyond Binary",
					Explanation: "Traditional memory stores binary (0 or 1). But ferroelectrics can store MANY intermediate states by partially switching the polarization.\n\n🔢 FeCIM uses 30 discrete levels for neural network weights!",
					Duration:    6 * time.Second,
					DifficultyLevel: LevelIntermediate,
				},
				{
					Title:       "How Partial Switching Works",
					Explanation: "Not all dipoles switch at exactly the same field. By applying a field just above Ec, we switch SOME dipoles, creating an intermediate polarization.\n\n📊 This is statistical domain switching.",
					Duration:    6 * time.Second,
					DifficultyLevel: LevelIntermediate,
				},
				{
					Title:       "Incremental Step Pulse Programming",
					Explanation: "ISPP applies a sequence of increasing pulses:\n\n1. Apply pulse\n2. Read current polarization\n3. If target not reached, apply slightly stronger pulse\n4. Repeat until target achieved\n\n🎯 This achieves precise level placement.",
					Duration:    8 * time.Second,
					DifficultyLevel: LevelIntermediate,
				},
				{
					Title:       "Try It: Program a Level",
					Explanation: "Let's program level 15 (middle of the 30-level range). Use the pulse control to apply incremental pulses.\n\n👆 Apply write pulses until P reaches the target zone.",
					Duration:    0,
					UserPrompt:  "Program the cell to Level 15 using ISPP",
					HighlightElement: "ispp-controls",
					DifficultyLevel: LevelIntermediate,
				},
				{
					Title:       "Level Separation",
					Explanation: "With 30 levels across the Pr to Ps range, each level is separated by:\n\nΔP = 2×Pr / 30 ≈ 2.3 µC/cm²\n\n📏 We need good SNR to distinguish adjacent levels!",
					Duration:    6 * time.Second,
					DifficultyLevel: LevelIntermediate,
				},
				{
					Title:       "Information Density",
					Explanation: "How many bits does 30 levels give us?\n\nbits = log₂(30) ≈ 4.91 bits/cell\n\n🎉 Almost 5× more than binary! Each cell stores one neural network weight.",
					Duration:    6 * time.Second,
					DifficultyLevel: LevelIntermediate,
				},
				{
					Title:       "Quick Check ✓",
					Explanation: "Test your understanding of analog storage!",
					Duration:    0,
					QuizQuestion: &QuizQuestion{
						Question:    "If we increase from 30 to 60 levels, how many bits per cell would we have?",
						Options:     []string{"~4.9 bits", "~5.9 bits", "~6.9 bits", "~7.9 bits"},
						CorrectIdx:  1,
						Explanation: "Correct! log₂(60) ≈ 5.9 bits. Doubling levels adds about 1 bit. But more levels = tighter tolerances and more noise sensitivity!",
					},
					DifficultyLevel: LevelIntermediate,
				},
				{
					Title:       "Retention Challenge",
					Explanation: "Intermediate states are less stable than fully saturated states. Thermal relaxation can cause levels to drift over time.\n\n⏰ Periodic refresh or error correction may be needed for long-term storage.",
					Duration:    6 * time.Second,
					DifficultyLevel: LevelIntermediate,
				},
				{
					Title:       "🎉 Tutorial Complete!",
					Explanation: "You understand analog level programming!\n\n✅ Partial switching creates intermediate states\n✅ ISPP achieves precise level targeting\n✅ 30 levels ≈ 4.9 bits per cell\n\n➡️ Next: Explore the crossbar array tutorials!",
					Duration:    8 * time.Second,
					DifficultyLevel: LevelIntermediate,
				},
			},
		},
	}
}

// GetCrossbarTutorials returns interactive tutorials for Module 2.
func GetCrossbarTutorials() []InteractiveTutorial {
	return []InteractiveTutorial{
		{
			ID:          "xbar-intro",
			Name:        "Crossbar Array Fundamentals",
			Description: "Learn how crossbar arrays perform matrix-vector multiplication.",
			Module:      "module2-crossbar",
			Difficulty:  LevelBeginner,
			Duration:    12 * time.Minute,
			LearningGoals: []string{
				"Understand crossbar array structure",
				"Explain how MVM works in hardware",
				"Calculate simple crossbar outputs",
			},
			Steps: []TutorialStep{
				{
					Title:       "Welcome to Crossbar Arrays!",
					Explanation: "A crossbar array is the heart of FeCIM computation. It performs matrix-vector multiplication (MVM) in a single step!\n\n🎯 In this tutorial, you'll learn how this magical hardware works.",
					Duration:    5 * time.Second,
					DifficultyLevel: LevelBeginner,
				},
				{
					Title:       "Array Structure",
					Explanation: "Imagine a grid:\n• Horizontal lines = Word Lines (WL) = Rows\n• Vertical lines = Bit Lines (BL) = Columns\n• At each intersection = one memory cell\n\n📐 A 128×128 array has 16,384 cells!",
					Duration:    6 * time.Second,
					HighlightElement: "crossbar-structure",
					DifficultyLevel: LevelBeginner,
				},
				{
					Title:       "Cells Store Conductance",
					Explanation: "Each cell has a conductance G (the inverse of resistance R).\n\nG = 1/R\n\nIn FeCIM, conductance is set by the polarization state. High P = High G.",
					Duration:    6 * time.Second,
					DifficultyLevel: LevelBeginner,
				},
				{
					Title:       "Try It: Explore the Array",
					Explanation: "Click on cells to see their conductance values. Brighter = higher conductance.\n\n👆 Click a few cells to explore.",
					Duration:    0,
					UserPrompt:  "Click on cells to view conductance values",
					HighlightElement: "crossbar-grid",
					DifficultyLevel: LevelBeginner,
				},
				{
					Title:       "The MVM Operation",
					Explanation: "Here's where physics helps us compute!\n\n1. Apply voltages V₁, V₂, ... to rows\n2. By Ohm's Law: I = G × V\n3. Currents sum at each column: I_col = Σ(G × V)\n\n🧮 This IS matrix-vector multiplication!",
					Duration:    8 * time.Second,
					DifficultyLevel: LevelBeginner,
				},
				{
					Title:       "Visualizing MVM",
					Explanation: "Input vector V = [V₁, V₂, V₃, ...] applied to rows\nWeight matrix W stored in cells (as G values)\nOutput vector I = W × V read from columns\n\n⚡ All multiplications happen IN PARALLEL!",
					Duration:    6 * time.Second,
					DifficultyLevel: LevelBeginner,
				},
				{
					Title:       "Try It: Run an MVM",
					Explanation: "Apply an input vector and watch the output currents.\n\n👆 Set input values and click 'Compute MVM'.",
					Duration:    0,
					UserPrompt:  "Run a matrix-vector multiplication",
					HighlightElement: "mvm-controls",
					DifficultyLevel: LevelBeginner,
				},
				{
					Title:       "Quick Check ✓",
					Explanation: "Time to test your understanding!",
					Duration:    0,
					QuizQuestion: &QuizQuestion{
						Question:    "In a 100×100 crossbar, how many multiply-add operations happen in one MVM?",
						Options:     []string{"100", "200", "10,000", "1,000,000"},
						CorrectIdx:  2,
						Explanation: "Correct! 100×100 = 10,000 cells, each performing one multiplication. All happen in parallel in O(1) time!",
					},
					DifficultyLevel: LevelBeginner,
				},
				{
					Title:       "Why This Matters",
					Explanation: "Traditional computing: MVM takes O(n²) operations\nCrossbar computing: MVM takes O(1) time!\n\n🚀 This is the fundamental speedup of compute-in-memory.",
					Duration:    6 * time.Second,
					DifficultyLevel: LevelBeginner,
				},
				{
					Title:       "🎉 Tutorial Complete!",
					Explanation: "You understand crossbar MVM!\n\n✅ Arrays store weights as conductance\n✅ Ohm's Law + Kirchhoff's current law = MVM\n✅ Parallel operation = O(1) compute\n\n➡️ Next: Learn about crossbar non-idealities!",
					Duration:    8 * time.Second,
					DifficultyLevel: LevelBeginner,
				},
			},
		},
		{
			ID:          "xbar-nonideal",
			Name:        "Crossbar Non-Idealities",
			Description: "Understand and visualize IR drop, sneak paths, and other real-world effects.",
			Module:      "module2-crossbar",
			Difficulty:  LevelIntermediate,
			Duration:    15 * time.Minute,
			Prerequisites: []string{"xbar-intro"},
			LearningGoals: []string{
				"Identify sources of error in crossbar arrays",
				"Visualize IR drop patterns",
				"Understand sneak path currents",
				"Know mitigation strategies",
			},
			Steps: []TutorialStep{
				{
					Title:       "Real Crossbars Aren't Perfect",
					Explanation: "Our idealized MVM assumed perfect wires with zero resistance. Real arrays have parasitic effects that cause errors.\n\n🔍 Let's explore the main non-idealities.",
					Duration:    5 * time.Second,
					DifficultyLevel: LevelIntermediate,
				},
				{
					Title:       "IR Drop: The Wire Problem",
					Explanation: "Metal lines have small but non-zero resistance. As current flows, voltage drops along the line.\n\nV_actual = V_applied - I × R_wire\n\n📉 Cells far from the voltage source see lower effective voltage!",
					Duration:    6 * time.Second,
					DifficultyLevel: LevelIntermediate,
				},
				{
					Title:       "Try It: Visualize IR Drop",
					Explanation: "Enable the IR drop overlay to see voltage distribution across the array.\n\n👆 Toggle the IR drop visualization.",
					Duration:    0,
					UserPrompt:  "Enable IR drop visualization",
					HighlightElement: "ir-drop-toggle",
					DifficultyLevel: LevelIntermediate,
				},
				{
					Title:       "IR Drop Pattern",
					Explanation: "Notice the gradient! Cells near the corner (far from drivers) have the lowest effective voltage.\n\n🗺️ This creates a systematic, position-dependent error map.",
					Duration:    6 * time.Second,
					HighlightElement: "ir-drop-heatmap",
					DifficultyLevel: LevelIntermediate,
				},
				{
					Title:       "IR Drop Mitigation",
					Explanation: "Solutions:\n\n1. Lower wire resistance (thicker metals, shorter paths)\n2. Break large arrays into smaller tiles\n3. Software compensation using calibrated error maps\n4. Differential signaling",
					Duration:    7 * time.Second,
					DifficultyLevel: LevelIntermediate,
				},
				{
					Title:       "Sneak Paths",
					Explanation: "In passive arrays (no transistor per cell), current can flow through unintended paths.\n\n🐍 Example: Trying to read cell (0,0) but current sneaks through (0,1)→(1,1)→(1,0).",
					Duration:    6 * time.Second,
					DifficultyLevel: LevelIntermediate,
				},
				{
					Title:       "Try It: See Sneak Paths",
					Explanation: "Select a cell and visualize potential sneak current paths.\n\n👆 Click a cell to see its sneak path analysis.",
					Duration:    0,
					UserPrompt:  "Select a cell to view sneak paths",
					HighlightElement: "sneak-path-viz",
					DifficultyLevel: LevelIntermediate,
				},
				{
					Title:       "1T1R Architecture",
					Explanation: "Solution: Add one Transistor per cell (1T1R).\n\nThe transistor acts as a switch:\n• ON: allows current for computation\n• OFF: blocks sneak paths completely\n\n📐 Tradeoff: More area per cell",
					Duration:    6 * time.Second,
					DifficultyLevel: LevelIntermediate,
				},
				{
					Title:       "Quick Check ✓",
					Explanation: "Test your understanding!",
					Duration:    0,
					QuizQuestion: &QuizQuestion{
						Question:    "Which cells in a crossbar are MOST affected by IR drop?",
						Options:     []string{"Cells near the voltage drivers", "Cells in the center of the array", "Cells far from the voltage drivers", "All cells equally"},
						CorrectIdx:  2,
						Explanation: "Correct! Cells far from the voltage drivers see the most accumulated voltage drop along the metal lines.",
					},
					DifficultyLevel: LevelIntermediate,
				},
				{
					Title:       "Other Non-Idealities",
					Explanation: "Additional effects to consider:\n\n• Read noise (thermal, shot, flicker)\n• Write variability (cycle-to-cycle, device-to-device)\n• Stuck-at faults (cells that can't switch)\n• Temperature sensitivity",
					Duration:    6 * time.Second,
					DifficultyLevel: LevelIntermediate,
				},
				{
					Title:       "🎉 Tutorial Complete!",
					Explanation: "You understand crossbar non-idealities!\n\n✅ IR drop causes position-dependent errors\n✅ Sneak paths corrupt passive array reads\n✅ 1T1R and compensation mitigate effects\n\n➡️ Next: MNIST neural network tutorial!",
					Duration:    8 * time.Second,
					DifficultyLevel: LevelIntermediate,
				},
			},
		},
	}
}

// GetMNISTTutorials returns interactive tutorials for Module 3.
func GetMNISTTutorials() []InteractiveTutorial {
	return []InteractiveTutorial{
		{
			ID:          "mnist-inference",
			Name:        "MNIST Digit Recognition",
			Description: "See how FeCIM crossbars recognize handwritten digits in real-time.",
			Module:      "module3-mnist",
			Difficulty:  LevelBeginner,
			Duration:    15 * time.Minute,
			Prerequisites: []string{"xbar-intro"},
			LearningGoals: []string{
				"Understand neural network inference flow",
				"See MVM in action for digit recognition",
				"Compare floating-point vs quantized accuracy",
			},
			Steps: []TutorialStep{
				{
					Title:       "Neural Networks on FeCIM",
					Explanation: "Neural networks are layers of matrix multiplications. Perfect for crossbar arrays!\n\n🧠 In this tutorial, you'll see a real network classify handwritten digits.",
					Duration:    5 * time.Second,
					DifficultyLevel: LevelBeginner,
				},
				{
					Title:       "The MNIST Dataset",
					Explanation: "MNIST: 70,000 handwritten digits (0-9)\n• Each image: 28×28 = 784 pixels\n• Each pixel: 0 (white) to 255 (black)\n\n📝 A classic benchmark for neural networks since 1998!",
					Duration:    5 * time.Second,
					DifficultyLevel: LevelBeginner,
				},
				{
					Title:       "Network Architecture",
					Explanation: "Our network:\n\n784 inputs (pixels)\n  ↓ Layer 1: 784×128 weights\n128 hidden neurons (ReLU)\n  ↓ Layer 2: 128×10 weights\n10 outputs (digit probabilities)\n\n🏗️ Two crossbar arrays: 784×128 and 128×10",
					Duration:    7 * time.Second,
					DifficultyLevel: LevelBeginner,
				},
				{
					Title:       "Try It: Draw a Digit",
					Explanation: "Use the canvas to draw a digit (0-9). The network will try to recognize it!\n\n✏️ Draw a digit in the input box.",
					Duration:    0,
					UserPrompt:  "Draw a digit (0-9) on the canvas",
					HighlightElement: "draw-canvas",
					DifficultyLevel: LevelBeginner,
				},
				{
					Title:       "Layer 1 Computation",
					Explanation: "Your drawing is flattened to 784 values and sent to the first crossbar.\n\n🧮 128 neurons compute in parallel - watch the hidden layer activate!",
					Duration:    5 * time.Second,
					HighlightElement: "layer1-viz",
					DifficultyLevel: LevelBeginner,
				},
				{
					Title:       "Layer 2 Computation",
					Explanation: "The 128 hidden values go to the second crossbar.\n\n🎯 10 outputs emerge - each representing a digit's 'score'.",
					Duration:    5 * time.Second,
					HighlightElement: "layer2-viz",
					DifficultyLevel: LevelBeginner,
				},
				{
					Title:       "The Prediction",
					Explanation: "Softmax converts scores to probabilities. The highest probability wins!\n\n🏆 Check if the network recognized your digit correctly.",
					Duration:    5 * time.Second,
					HighlightElement: "prediction-display",
					DifficultyLevel: LevelBeginner,
				},
				{
					Title:       "Try It: Compare Modes",
					Explanation: "Toggle between 'FP32' (perfect math) and 'CIM 30-level' (realistic hardware).\n\n🔍 Watch how quantization affects the prediction.",
					Duration:    0,
					UserPrompt:  "Compare floating-point and CIM modes",
					HighlightElement: "mode-toggle",
					DifficultyLevel: LevelBeginner,
				},
				{
					Title:       "Quick Check ✓",
					Explanation: "Test your understanding!",
					Duration:    0,
					QuizQuestion: &QuizQuestion{
						Question:    "Why might FP32 and 30-level CIM give different predictions?",
						Options:     []string{"Different random seeds", "Weight quantization error", "Different input preprocessing", "Hardware runs faster"},
						CorrectIdx:  1,
						Explanation: "Correct! The 30-level CIM quantizes weights, introducing small errors that can accumulate through layers and occasionally change predictions.",
					},
					DifficultyLevel: LevelBeginner,
				},
				{
					Title:       "🎉 Tutorial Complete!",
					Explanation: "You've seen FeCIM neural network inference!\n\n✅ MNIST images flow through crossbar layers\n✅ MVM computes all neurons in parallel\n✅ 30-level quantization maintains good accuracy\n\n➡️ Next: Explore the circuits module!",
					Duration:    8 * time.Second,
					DifficultyLevel: LevelBeginner,
				},
			},
		},
	}
}

// GetAllTutorials returns all available interactive tutorials.
func GetAllTutorials() []InteractiveTutorial {
	var all []InteractiveTutorial
	all = append(all, GetHysteresisTutorials()...)
	all = append(all, GetCrossbarTutorials()...)
	all = append(all, GetMNISTTutorials()...)
	return all
}

// GetTutorialByID finds a tutorial by its ID.
func GetTutorialByID(id string) *InteractiveTutorial {
	for _, t := range GetAllTutorials() {
		if t.ID == id {
			return &t
		}
	}
	return nil
}

// GetTutorialsByModule returns tutorials for a specific module.
func GetTutorialsByModule(module string) []InteractiveTutorial {
	var result []InteractiveTutorial
	for _, t := range GetAllTutorials() {
		if t.Module == module {
			result = append(result, t)
		}
	}
	return result
}

// TutorialToController converts an InteractiveTutorial to a TutorialController.
func TutorialToController(tutorial InteractiveTutorial) *TutorialController {
	return NewTutorialController(tutorial.Steps)
}
