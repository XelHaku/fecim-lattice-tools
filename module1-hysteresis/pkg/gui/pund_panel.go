package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	sharedphysics "fecim-lattice-tools/shared/physics"
)

// PUNDPanel provides PUND measurement mode visualization.
type PUNDPanel struct {
	resultsLabel *widget.Label
	runButton    *widget.Button
	content      fyne.CanvasObject
}

// NewPUNDPanel creates a new PUND measurement panel.
func NewPUNDPanel() *PUNDPanel {
	p := &PUNDPanel{
		resultsLabel: widget.NewLabel("Press 'Run PUND' to execute measurement"),
	}

	p.runButton = widget.NewButton("Run PUND", func() {
		go p.runPUND()
	})

	p.content = container.NewVBox(
		widget.NewLabelWithStyle("PUND Measurement", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel("6-pulse protocol: Preset -> P -> U -> N -> D"),
		p.runButton,
		widget.NewSeparator(),
		p.resultsLabel,
	)
	return p
}

func (p *PUNDPanel) runPUND() {
	// Default HZO parameters
	ec := 3e7  // V/m
	ps := 0.25 // C/m²
	stack := sharedphysics.NewPreisachStack(ec, &sharedphysics.TanhEverett{Ec: ec, Ps: ps, Delta: ec * 0.3})

	result, _, err := sharedphysics.RunPUNDSimulation(stack, 5*ec, 100e-9, 5e-9, 1e-12)

	var text string
	if err != nil {
		text = fmt.Sprintf("PUND error: %v", err)
	} else {
		text = fmt.Sprintf(
			"QP: %.3e C\nQU: %.3e C\nQN: %.3e C\nQD: %.3e C\n\nPsw+ = QP-QU: %.3e C\nPsw- = QN-QD: %.3e C",
			result.QP_C, result.QU_C, result.QN_C, result.QD_C,
			result.SwitchingPositive_C, result.SwitchingNegative_C,
		)
	}

	fyne.Do(func() {
		p.resultsLabel.SetText(text)
	})
}

// Content returns the panel's Fyne canvas object.
func (p *PUNDPanel) Content() fyne.CanvasObject {
	return p.content
}
