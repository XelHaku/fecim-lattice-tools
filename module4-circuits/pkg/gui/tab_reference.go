// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// This file contains the main REFERENCE tab that combines TIMING and SPECS sections.
package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// ============================================================================
// UNIFIED REFERENCE TAB: TIMING DIAGRAMS + SPECIFICATIONS
// ============================================================================

// createReferenceTab creates the unified REFERENCE view
func (ca *CircuitsApp) createReferenceTab() fyne.CanvasObject {
	// Create sections FIRST (before selector triggers callback)
	timingSection := ca.createReferenceTimingSection()
	specsSection := ca.createReferenceSpecsSection()

	// Assign to struct BEFORE SetSelected triggers callback
	ca.refTimingSection = timingSection
	ca.refSpecsSection = specsSection
	specsSection.Hide()

	// Section selector (callback now safe - sections are assigned)
	sectionSelect := widget.NewSelect([]string{"TIMING DIAGRAMS", "SPECIFICATIONS"}, func(s string) {
		ca.onReferenceSectionChanged(s)
	})
	sectionSelect.SetSelected("TIMING DIAGRAMS")

	contentStack := container.NewStack(timingSection, specsSection)

	header := container.NewHBox(
		widget.NewLabel("Reference:"),
		sectionSelect,
		layout.NewSpacer(),
	)

	return container.NewBorder(
		container.NewVBox(header, widget.NewSeparator()),
		nil, nil, nil,
		contentStack,
	)
}

func (ca *CircuitsApp) onReferenceSectionChanged(section string) {
	// Safety check - sections may not be initialized yet
	if ca.refTimingSection == nil || ca.refSpecsSection == nil {
		return
	}
	if section == "TIMING DIAGRAMS" {
		ca.refTimingSection.Show()
		ca.refSpecsSection.Hide()
	} else {
		ca.refTimingSection.Hide()
		ca.refSpecsSection.Show()
	}
}
