//go:build legacy_fyne

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
	Title     string
	IconName  fyne.ThemeIconName
	Summary   string
	Details   string
	LearnMore string // URL for external reference
}

func aboutScienceData() []ScienceSection {
	return []ScienceSection{
		{
			Title:    "What is FeCIM?",
			IconName: theme.IconNameInfo,
			Summary:  "Compute-in-memory with ferroelectric devices",
			Details: "FeCIM (Ferroelectric Compute-in-Memory) stores analog weights inside memory cells and computes in place.\n\n" +
				"Rows apply voltages, columns sum currents, and the array performs matrix-vector multiplication in parallel using Ohm's and Kirchhoff's laws.",
			LearnMore: "https://en.wikipedia.org/wiki/Compute-in-memory",
		},
		{
			Title:    "HZO material",
			IconName: theme.IconNameStorage,
			Summary:  "Why hafnium-zirconium oxide matters",
			Details: "HZO (Hf0.5Zr0.5O2) is a CMOS-compatible ferroelectric material used to build non-volatile analog cells.\n\n" +
				"Its polarization can be switched with electric fields, enabling compact memory-plus-compute primitives for FeCIM hardware.",
			LearnMore: "https://en.wikipedia.org/wiki/Hafnium(IV)_oxide",
		},
		{
			Title:    "Hysteresis",
			IconName: theme.IconNameMediaReplay,
			Summary:  "The memory effect in the P-E loop",
			Details: "Ferroelectrics show a polarization-electric-field (P-E) hysteresis loop.\n\n" +
				"When field returns to zero, remanent polarization remains. This is the physical basis of non-volatile states and multi-level programming.",
			LearnMore: "https://en.wikipedia.org/wiki/Ferroelectricity",
		},
		{
			Title:    "Crossbar arrays",
			IconName: theme.IconNameGrid,
			Summary:  "Hardware matrix math at array scale",
			Details: "Crossbar arrays map neural weights to cell conductances.\n\n" +
				"A single analog operation can evaluate many multiply-accumulate terms in parallel, while non-idealities (IR drop, variation, leakage) set practical limits.",
			LearnMore: "https://doi.org/10.1038/s41586-020-2181-2",
		},
		{
			Title:    "Neuromorphic computing",
			IconName: theme.IconNameComputer,
			Summary:  "Brain-inspired efficiency goals",
			Details: "Neuromorphic systems co-locate memory and compute to reduce data movement energy.\n\n" +
				"FeCIM crossbars are one route to this goal: analog in-memory MACs with non-volatile weights for low-power AI inference.",
			LearnMore: "https://en.wikipedia.org/wiki/Neuromorphic_engineering",
		},
	}
}

// ShowAboutScience displays the unified "About the Science" dialog.
func ShowAboutScience(parent fyne.Window) {
	sections := aboutScienceData()
	items := make([]*widget.AccordionItem, len(sections))

	for i, section := range sections {
		detailsLabel := widget.NewLabel(section.Details)
		detailsLabel.Wrapping = fyne.TextWrapWord

		contentItems := []fyne.CanvasObject{detailsLabel}
		if section.LearnMore != "" {
			parsedURL, err := url.Parse(section.LearnMore)
			if err == nil {
				learnMoreBtn := widget.NewButtonWithIcon("Learn More", theme.DocumentIcon(), func() {
					_ = fyne.CurrentApp().OpenURL(parsedURL)
				})
				learnMoreBtn.Importance = widget.LowImportance
				contentItems = append(contentItems, container.NewHBox(learnMoreBtn))
			}
		}

		items[i] = widget.NewAccordionItem(
			section.Title+" — "+section.Summary,
			container.NewVBox(contentItems...),
		)
	}

	accordion := widget.NewAccordion(items...)
	accordion.MultiOpen = true

	header := container.NewVBox(
		widget.NewLabelWithStyle("About the Science", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Unified Learn More primer for ferroelectric computing"),
		widget.NewSeparator(),
	)

	footer := container.NewVBox(
		widget.NewSeparator(),
		widget.NewRichTextFromMarkdown("*Concise overview; use each section's Learn More link for deeper reading.*"),
	)

	scroll := container.NewScroll(accordion)
	scroll.SetMinSize(fyne.NewSize(580, 420))

	content := container.NewBorder(header, footer, nil, nil, scroll)
	d := dialog.NewCustom("About the Science", "Close", content, parent)
	d.Resize(fyne.NewSize(700, 600))
	d.Show()
}

// CreateAboutScienceButton creates a standardized button for the About Science dialog.
func CreateAboutScienceButton(parent fyne.Window) *widget.Button {
	btn := widget.NewButtonWithIcon("About the Science", theme.QuestionIcon(), func() {
		ShowAboutScience(parent)
	})
	return btn
}
