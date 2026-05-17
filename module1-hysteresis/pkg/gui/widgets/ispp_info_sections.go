//go:build legacy_fyne

package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

func buildIsppOverviewSection() fyne.CanvasObject {
	return container.NewVBox(
		sectionTitle("ISPP (Incremental Step Pulse Programming)"),
		bodyLabel("ISPP is a write-verify loop: apply a programming pulse, then VERIFY by reading the discrete level at E≈0."),
		bodyLabel("Unlike the L-K and Preisach tabs, ISPP is a controller (state machine), not a closed-form constitutive equation."),
		bodyLabel("The key invariant: the UI target highlight should reflect the *active controller target* during WRITE/VERIFY/SUCCESS until the field settles."),
	)
}

func buildIsppStatesSection() fyne.CanvasObject {
	return container.NewVBox(
		sectionTitle("Controller States"),
		bodyLabel(bullets([]string{
			"RESET/PREP: drive toward a known branch (saturation) to make the mapping monotonic.",
			"WRITE: apply Vpulse (E-field) with adaptive step sizing.",
			"WAIT: hold briefly before VERIFY.",
			"VERIFY: measure level at E≈0; update bounds/step; detect overshoot.",
			"SUCCESS/FAILED: terminal states for this target.",
		})),
	)
}

func buildIsppStabilitySection() fyne.CanvasObject {
	return container.NewVBox(
		sectionTitle("Stability & Deadlock Breakers"),
		bodyLabel(bullets([]string{
			"Bounds must remain valid: Vmin < Vmax; clear/reset on invalid brackets.",
			"Stuck detection: if level does not change across VERIFY cycles, raise a step floor and clear bounds.",
			"No-improvement detection: if error does not shrink, escalate step + clear bounds.",
			"Overshoot recovery: apply a short reverse pulse to get back near the target branch, then continue.",
		})),
	)
}

func buildIsppCodeRefsSection() fyne.CanvasObject {
	return container.NewVBox(
		sectionTitle("Code References"),
		bodyLabel(bullets([]string{
			"GUI controller: module1-hysteresis/pkg/controller (WriteController)",
			"GUI orchestration + highlight: module1-hysteresis/pkg/gui/simulation.go (refreshGUI WRD target highlight)",
			"Headless evidence runner: cmd/fecim-lattice-tools/mode.go (ISPP write-verify sequence)",
			"Physics engines: shared/physics/landau.go (LK) and module1-hysteresis/pkg/ferroelectric (Preisach)",
		})),
	)
}
