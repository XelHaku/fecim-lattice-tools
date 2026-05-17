//go:build legacy_fyne

// Package gui provides Fyne-based GUI components for MNIST visualization.
// dialogs.go provides educational info dialogs.
package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ShowInfoDialog displays a tabbed dialog with all information sections.
func ShowInfoDialog(window fyne.Window) {
	// Create content for each tab (reusing existing dialog content functions)
	why30Content := createWhy30LevelsContent()
	hardwareContent := createHardwareRealityContent()
	failuresContent := createFailureModesContent()
	aboutContent := createAboutContent()

	// Create tabs
	tabs := container.NewAppTabs(
		container.NewTabItem("Why 30 Levels?", container.NewVScroll(why30Content)),
		container.NewTabItem("Hardware Reality", container.NewVScroll(hardwareContent)),
		container.NewTabItem("Failure Modes", container.NewVScroll(failuresContent)),
		container.NewTabItem("About", container.NewVScroll(aboutContent)),
	)

	// Create dialog with tabbed content
	d := dialog.NewCustom("Information", "Close", tabs, window)
	d.Resize(fyne.NewSize(600, 550))
	d.Show()
}

// createWhy30LevelsContent creates the content for the Why 30 Levels section.
func createWhy30LevelsContent() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabelWithStyle("Why 30 Analog Levels?", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),

		// CRIT-003: Verification status section
		widget.NewLabelWithStyle("⚠️ Verification Status", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("• 30-state conference baseline: UNVERIFIED (2025 conference report, not peer-reviewed)"),
		widget.NewLabel("• Peer-reviewed range: 32-140 states demonstrated"),
		widget.NewLabel("  - 32 states: Oh et al., IEEE EDL 2017 (VERIFIED)"),
		widget.NewLabel("  - 140 states: Song et al., Adv. Science 2024 (VERIFIED)"),
		widget.NewLabel("• Conclusion: 30 states is a demo baseline from a conference report (plausible within demonstrated range)"),
		widget.NewSeparator(),

		widget.NewLabelWithStyle("Physics Justification", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("• Material: HfO₂-ZrO₂ (HZO) ferroelectric superlattice"),
		widget.NewLabel("• Mechanism: dozens of stable polarization states from domain wall pinning"),
		widget.NewLabel("• Crystal defects: Oxygen vacancies create energy barriers"),
		widget.NewLabel("• ADC limitation: finite ADC resolution (default 8-bit / 256 levels in this demo) → fewer reliably distinguishable levels in practice"),
		widget.NewLabel("• Separation: 3σ noise margin between adjacent levels"),
		widget.NewSeparator(),

		widget.NewLabelWithStyle("Peer-Reviewed Technology Comparison", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Technology     | Levels | Bits/Cell | Verification"),
		widget.NewLabel("---------------|--------|-----------|------------------"),
		widget.NewLabel("SRAM           | 2      | 1 bit     | Industry standard"),
		widget.NewLabel("Flash (NAND)   | 2-4    | 1-2 bits  | Industry standard"),
		widget.NewLabel("ReRAM          | 4-16   | 2-4 bits  | Peer-reviewed"),
		widget.NewLabel("FeFET (Oh 2017)| 32     | 5.0 bits  | IEEE EDL (VERIFIED)"),
		widget.NewLabel("FeFET (Song'24)| 140    | 7.1 bits  | Adv. Science (VERIFIED)"),
		widget.NewLabel("Conference ref.| 30     | ~4.9 bits | 2025 report (UNVERIFIED, within 32-140 range)"),
		widget.NewSeparator(),

		widget.NewLabelWithStyle("Impact on MNIST Accuracy", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("• 2 levels (binary): ~50% (PROJECTED - simulation)"),
		widget.NewLabel("• 4 levels (2-bit): ~65% (PROJECTED - simulation)"),
		widget.NewLabel("• 7 levels (FeFET): 96.6% (VERIFIED - Nature Commun. 2023)"),
		widget.NewLabel("• 30+ levels: 92-98% (PROJECTED - simulation, varies)"),
		widget.NewLabel("• HZO-FTJ reservoir: 98.24% (VERIFIED - ScienceDirect 2025)"),
		widget.NewLabel("• Float32 (digital): ~98% (theoretical baseline)"),
		widget.NewLabel(""),
		widget.NewLabel("Note: Accuracy depends on implementation quality and non-idealities"),
		widget.NewSeparator(),

		widget.NewLabelWithStyle("Why Not 64+ Levels (ADC/Noise limited in practice)?", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("This demo assumes 30 levels as a conservative baseline (conference claim); practical limits vary by process:"),
		widget.NewLabel("1. Device-to-device variation: ~2.75% (process variation)"),
		widget.NewLabel("2. Cycle-to-cycle variation: ~1.5% (retention drift)"),
		widget.NewLabel("3. Read noise: ~0.5% σ/μ (thermal Johnson noise)"),
		widget.NewLabel("4. Temperature sensitivity: ±1% per 10°C"),
		widget.NewLabel(""),
		widget.NewLabel("Statistical requirement: 3σ separation between adjacent levels"),
		widget.NewLabel("Result: 30 levels is a reasonable demo baseline, not a universal manufacturing limit"),
	)
}

// ShowWhy30LevelsDialog displays information about the 30 analog levels.
// CRIT-003 fix: Add verification status and peer-reviewed context
func ShowWhy30LevelsDialog(window fyne.Window) {
	content := createWhy30LevelsContent()
	scroll := container.NewVScroll(content)
	scroll.SetMinSize(fyne.NewSize(500, 400))

	d := dialog.NewCustom("Why 30 Levels?", "Close", scroll, window)
	d.Resize(fyne.NewSize(550, 500))
	d.Show()
}

// createHardwareRealityContent creates the content for the Hardware Reality section.
func createHardwareRealityContent() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabelWithStyle("Hardware Reality Check", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),

		widget.NewLabelWithStyle("Accuracy Depends on Configuration", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(""),
		widget.NewLabel("Simulation (this demo): 85-98% depending on noise, levels, ADC bits"),
		widget.NewLabel("Peer-reviewed analog results: 96.6% FeFET (Nature Commun. 2023), 98.24% FTJ (ScienceDirect 2025, non-FeCIM)"),
		widget.NewLabel("Software baseline: 98-99% (FP32); CIM: 92-98% (peer-reviewed literature)"),
		widget.NewSeparator(),

		widget.NewLabelWithStyle("The Accuracy Gap Explained", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Non-Ideality          | This Sim | Real HW | Accuracy Impact"),
		widget.NewLabel("----------------------|----------|---------|----------------"),
		widget.NewLabel("Weight quantization   | ✓ Yes    | ✓ Yes   | -1% to -2%"),
		widget.NewLabel("Read noise (Gaussian) | ✓ Yes    | ✓ Yes   | -2% to -3%"),
		widget.NewLabel("IR drop (resistive)   | ✗ No     | ✓ Yes   | -2% to -3%"),
		widget.NewLabel("Sneak paths (leakage) | ✗ No     | ✓ Yes   | -1% to -2%"),
		widget.NewLabel("ADC non-linearity     | ✗ No     | ✓ Yes   | -1%"),
		widget.NewLabel("Retention drift       | ✗ No     | ✓ Yes   | -1%"),
		widget.NewLabel("Cycle-to-cycle var.   | ✗ No     | ✓ Yes   | -1% to -2%"),
		widget.NewLabel(""),
		widget.NewLabel("Gap varies: Peer-reviewed achieves 96-98%; gap depends on implementation quality"),
		widget.NewSeparator(),

		widget.NewLabelWithStyle("How to Match Real Hardware in This Simulation", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("To approximate real hardware behavior:"),
		widget.NewLabel("• Adjust noise level (1-15% standard deviation)"),
		widget.NewLabel("• This captures the combined effect of hardware non-idealities"),
		widget.NewLabel("• Result: Accuracy varies from ~70% (high noise) to ~98% (ideal)"),
		widget.NewSeparator(),

		widget.NewLabelWithStyle("Energy Efficiency Advantage", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Reported in literature: 25-100× vs NAND/GPU (not verified here)"),
		widget.NewLabel("This UI uses model inputs (Horowitz 2014 baseline)"),
		widget.NewLabel(""),
		widget.NewLabel("FeCIM Energy: ~10 fJ/bit/MAC × log2(levels) + ADC/DAC overhead (shared core model)"),
		widget.NewLabel(""),
		widget.NewLabel("Note: Exact improvement depends on workload and comparison baseline."),
	)
}

// ShowHardwareRealityDialog displays the hardware reality check information.
func ShowHardwareRealityDialog(window fyne.Window) {
	content := createHardwareRealityContent()
	scroll := container.NewVScroll(content)
	scroll.SetMinSize(fyne.NewSize(500, 400))

	d := dialog.NewCustom("Hardware Reality Check", "Close", scroll, window)
	d.Resize(fyne.NewSize(550, 500))
	d.Show()
}

// createFailureModesContent creates the content for the Failure Modes section.
func createFailureModesContent() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabelWithStyle("Failure Modes Explained", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),

		widget.NewLabelWithStyle("1. Quantization Cliff (< 4 levels)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Trigger: Quick Demo step 4 (forces 2 levels via PTQ)"),
		widget.NewLabel("Settings: Levels=2, Noise=0.01, ADC=8-bit (fixed in UI)"),
		widget.NewLabel("Result: Accuracy collapses to ~50% (worse than random guessing!)"),
		widget.NewLabel(""),
		widget.NewLabel("Root Cause: Binary weights {-1, +1} create severe quantization error."),
		widget.NewLabel("The 128-dimensional weight space collapses to only 2 discrete values,"),
		widget.NewLabel("destroying the network's ability to represent learned features."),
		widget.NewLabel("Observable: Weight heatmap shows only blue and red (no gradation)."),
		widget.NewSeparator(),

		widget.NewLabelWithStyle("2. Noise Wall (> 0.10 noise)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Preset: Click 'Noisy' button to test"),
		widget.NewLabel("Settings: Levels=30, Noise=0.15 (15% std dev), ADC=8-bit (fixed in UI)"),
		widget.NewLabel("Result: Accuracy degrades to ~70%, confidence drops to 40-60%"),
		widget.NewLabel(""),
		widget.NewLabel("Root Cause: Gaussian noise corrupts analog currents during MVM."),
		widget.NewLabel("Sources: Thermal (Johnson) noise in sense amplifiers, device variation,"),
		widget.NewLabel("and retention drift combine to cause the ADC to read incorrect values."),
		widget.NewLabel("Observable: Digit '8' frequently misclassified as '3' or '5'."),
		widget.NewSeparator(),

		widget.NewLabelWithStyle("3. ADC Quantization Artifacts (< 4-bit)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Advanced (not exposed in Dual-Mode UI)"),
		widget.NewLabel("Settings: Levels=30, Noise=0.01, ADC=3-bit (only 8 levels!)"),
		widget.NewLabel("Result: Accuracy degrades to ~65% with visible staircase artifacts"),
		widget.NewLabel(""),
		widget.NewLabel("Root Cause: 3-bit ADC provides only 8 discrete output levels."),
		widget.NewLabel("Hidden layer activations are coarsely quantized, losing fine-grained"),
		widget.NewLabel("information needed to distinguish similar digits (e.g., '3' vs '8')."),
		widget.NewLabel("Observable: Probability distribution shows staircase-like discontinuities."),
		widget.NewSeparator(),

		widget.NewLabelWithStyle("4. Confidence Collapse (Extreme Failure)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Advanced (not exposed in Dual-Mode UI)"),
		widget.NewLabel("Settings: Levels=2, Noise=0.20 (20% std dev), ADC=3-bit"),
		widget.NewLabel("Result: All output probabilities → ~10% (uniform distribution)"),
		widget.NewLabel(""),
		widget.NewLabel("Root Cause: Catastrophic combination of three failure modes:"),
		widget.NewLabel("1. Binary quantization (information loss in weights)"),
		widget.NewLabel("2. Extreme noise (20% corrupts analog computation)"),
		widget.NewLabel("3. Coarse ADC (only 8 discrete outputs)"),
		widget.NewLabel(""),
		widget.NewLabel("Result: Network completely loses ability to extract meaningful features."),
		widget.NewLabel("All digits produce nearly uniform probability distributions."),
	)
}

// ShowFailureModesDialog displays information about failure modes.
func ShowFailureModesDialog(window fyne.Window) {
	content := createFailureModesContent()
	scroll := container.NewVScroll(content)
	scroll.SetMinSize(fyne.NewSize(500, 400))

	d := dialog.NewCustom("Failure Modes", "Close", scroll, window)
	d.Resize(fyne.NewSize(550, 500))
	d.Show()
}

// createAboutContent creates the content for the About section.
func createAboutContent() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabelWithStyle("MNIST FeCIM Demo", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Educational Visualization of Ferroelectric Compute-in-Memory"),
		widget.NewLabel("Literature accuracy: 96.6% FeFET, 98.24% FTJ (non-FeCIM)"),
		widget.NewSeparator(),

		widget.NewLabelWithStyle("This Demo Answers Four Key Questions:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("1. What are 30 analog levels? (Physics + competitive advantage)"),
		widget.NewLabel("2. How does accuracy depend on configuration? (Noise, levels, ADC)"),
		widget.NewLabel("3. What happens when hardware fails? (Quantization, noise, ADC)"),
		widget.NewLabel("4. Why does this matter? (model-based energy gains reported in literature)"),
		widget.NewSeparator(),

		widget.NewLabelWithStyle("Neural Network Architecture", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("• Input Layer:  784 neurons (28×28 grayscale pixels)"),
		widget.NewLabel("• Hidden Layer: 128 neurons (ReLU activation)"),
		widget.NewLabel("• Output Layer: 10 neurons (Softmax, digits 0-9)"),
		widget.NewLabel("• Total MACs:   101,632 multiply-accumulate ops per inference"),
		widget.NewLabel("• Topology:     Fully connected (784→128→10)"),
		widget.NewSeparator(),

		widget.NewLabelWithStyle("Key References", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("• 2025 conference presentation (ferroelectric CIM baseline notes)"),
		widget.NewLabel("• Jerry et al., IEDM 2017 - FeFET Synapse (DOI: 10.1109/IEDM.2017.8268338)"),
		widget.NewLabel("• MNIST Dataset - Yann LeCun et al. (http://yann.lecun.com/exdb/mnist/)"),
		widget.NewSeparator(),

		widget.NewLabel("GitHub: your-org/fecim-lattice-tools (Open Source)"),
	)
}

// ShowAboutDialog displays information about the demo.
func ShowAboutDialog(window fyne.Window) {
	content := createAboutContent()
	scroll := container.NewVScroll(content)
	scroll.SetMinSize(fyne.NewSize(450, 400))

	d := dialog.NewCustom("About", "Close", scroll, window)
	d.Resize(fyne.NewSize(500, 500))
	d.Show()
}
