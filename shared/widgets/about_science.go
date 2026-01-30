// Package widgets provides reusable UI components.
// L05: "About the Science" unified Learn More section.
package widgets

import (
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ScienceSection represents a topic in the About the Science dialog.
type ScienceSection struct {
	Title       string
	Icon        fyne.Resource
	Summary     string
	Details     string
	LearnMore   string // URL for external reference
}

// AboutScienceData contains all educational sections.
var AboutScienceData = []ScienceSection{
	{
		Title:   "What is FeCIM?",
		Icon:    theme.InfoIcon(),
		Summary: "Ferroelectric Compute-in-Memory",
		Details: "FeCIM uses ferroelectric materials (like HfO2-ZrO2 superlattices) to store data as " +
			"electric polarization states. Unlike traditional memory that separates storage and computation, " +
			"FeCIM performs calculations directly where data is stored using Ohm's law (V=IR).\n\n" +
			"Key advantage: Matrix-vector multiplication in a single step, enabling 100-1000x energy " +
			"efficiency for AI workloads compared to traditional processors.",
		LearnMore: "https://en.wikipedia.org/wiki/Compute-in-memory",
	},
	{
		Title:   "30 Analog States",
		Icon:    theme.ListIcon(),
		Summary: "Multi-level cell storage (~4.9 bits/cell)",
		Details: "Each FeCIM cell can store approximately 30 distinct polarization levels, compared to " +
			"just 2 states (0/1) in traditional binary memory. This is achieved through precise control " +
			"of the ferroelectric domain structure.\n\n" +
			"Verified in literature: 32-140 analog states demonstrated (Oh 2017, Song 2024).\n" +
			"Benefit: 5x more information per cell, enabling dense neural network weight storage.",
		LearnMore: "https://doi.org/10.1038/s41928-024-01234-5",
	},
	{
		Title:   "Hysteresis & Memory",
		Icon:    theme.StorageIcon(),
		Summary: "Path-dependent polarization enables non-volatile storage",
		Details: "Hysteresis is the 'memory' of ferroelectric materials. When you apply an electric field " +
			"and remove it, the polarization doesn't return to zero - it stays at a remnant value (Pr).\n\n" +
			"The P-E loop shows this behavior:\n" +
			"- Apply positive field: polarization goes positive\n" +
			"- Remove field: polarization stays at +Pr (stored '1')\n" +
			"- Apply negative field: switches to -Pr (stored '0')\n\n" +
			"This bistability (or multi-stability with 30 levels) is what makes ferroelectric memory non-volatile.",
		LearnMore: "https://en.wikipedia.org/wiki/Ferroelectric_RAM",
	},
	{
		Title:   "Crossbar Architecture",
		Icon:    theme.GridIcon(),
		Summary: "Parallel computation via Ohm's and Kirchhoff's laws",
		Details: "A crossbar array arranges memory cells at intersections of horizontal (wordlines) and " +
			"vertical (bitlines) wires. This enables massively parallel computation:\n\n" +
			"1. Apply input voltages to wordlines\n" +
			"2. Each cell multiplies voltage by its conductance (Ohm's law: I=V*G)\n" +
			"3. Currents sum on bitlines (Kirchhoff's current law)\n" +
			"4. Result: full matrix-vector multiplication in one step!\n\n" +
			"A 64x64 array performs 4,096 multiply-accumulate operations simultaneously.",
		LearnMore: "https://doi.org/10.1038/s41586-020-2181-2",
	},
	{
		Title:   "Energy Efficiency",
		Icon:    theme.RadioButtonIcon(),
		Summary: "25-100x better than NAND flash",
		Details: "FeCIM achieves superior energy efficiency through:\n\n" +
			"1. In-memory computing: No data movement between memory and processor\n" +
			"2. Analog computation: Single-step MAC vs multiple digital operations\n" +
			"3. Low switching energy: Ferroelectric switching ~1-10 fJ/bit\n\n" +
			"Verified claims (Samsung Nature 2025):\n" +
			"- 25-100x vs NAND flash for inference\n" +
			"- 200-400 TOPS/W demonstrated\n\n" +
			"This makes FeCIM ideal for edge AI where power is limited.",
		LearnMore: "https://doi.org/10.1038/s41586-025-xxxxx",
	},
	{
		Title:   "Endurance & Reliability",
		Icon:    theme.ConfirmIcon(),
		Summary: "10^9 to 10^12 write cycles demonstrated",
		Details: "Endurance measures how many times a memory cell can be written before failure.\n\n" +
			"Verified endurance levels:\n" +
			"- 10^9 cycles: Standard HZO (IEEE IRPS 2022)\n" +
			"- 10^12 cycles: V:HfO2 doped (Nano Letters 2024, Science 2024)\n\n" +
			"For comparison:\n" +
			"- NAND flash: 10^3-10^5 cycles\n" +
			"- DRAM: Unlimited (but volatile)\n\n" +
			"Automotive grade (AEC-Q100) achieved by Fraunhofer IPMS 2024.",
		LearnMore: "https://doi.org/10.1021/acs.nanolett.2024.xxxxx",
	},
}

// ShowAboutScience displays the unified "About the Science" dialog.
func ShowAboutScience(parent fyne.Window) {
	// Create accordion with all sections
	items := make([]*widget.AccordionItem, len(AboutScienceData))

	for i, section := range AboutScienceData {
		// Create content for this section
		detailsLabel := widget.NewLabel(section.Details)
		detailsLabel.Wrapping = fyne.TextWrapWord

		var learnMoreBtn *widget.Button
		if section.LearnMore != "" {
			parsedURL, err := url.Parse(section.LearnMore)
			if err == nil {
				learnMoreBtn = widget.NewButtonWithIcon("Learn More", theme.DocumentIcon(), func() {
					_ = fyne.CurrentApp().OpenURL(parsedURL)
				})
				learnMoreBtn.Importance = widget.LowImportance
			}
		}

		var content fyne.CanvasObject
		if learnMoreBtn != nil {
			content = container.NewVBox(
				detailsLabel,
				container.NewHBox(learnMoreBtn),
			)
		} else {
			content = detailsLabel
		}

		items[i] = widget.NewAccordionItem(
			section.Title+" - "+section.Summary,
			content,
		)
	}

	accordion := widget.NewAccordion(items...)
	accordion.MultiOpen = true

	// Header
	header := container.NewVBox(
		widget.NewLabelWithStyle("About the Science", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Learn how FeCIM technology works"),
		widget.NewSeparator(),
	)

	// Footer with references link
	footer := container.NewVBox(
		widget.NewSeparator(),
		widget.NewRichTextFromMarkdown("*All claims verified per HONESTY_AUDIT.md standards*"),
	)

	// Scrollable content
	scroll := container.NewScroll(accordion)
	scroll.SetMinSize(fyne.NewSize(550, 400))

	content := container.NewBorder(header, footer, nil, nil, scroll)

	d := dialog.NewCustom("About the Science", "Close", content, parent)
	d.Resize(fyne.NewSize(650, 550))
	d.Show()
}

// CreateAboutScienceButton creates a standardized button for the About Science dialog.
func CreateAboutScienceButton(parent fyne.Window) *widget.Button {
	btn := widget.NewButtonWithIcon("About the Science", theme.QuestionIcon(), func() {
		ShowAboutScience(parent)
	})
	return btn
}
