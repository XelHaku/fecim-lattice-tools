package hysteresis

import (
	"fmt"

	"fecim-lattice-tools/shared/viewmodel"
)

func buildSnapshot(state HysteresisState) viewmodel.ModuleSnapshot {
	metrics := []viewmodel.Metric{
		{ID: "material", Label: "Material", Value: state.SelectedMaterial},
		{ID: "field_min", Label: "Min Field", Value: fmt.Sprintf("%.0f kV/cm", state.FieldRange.MinField)},
		{ID: "field_max", Label: "Max Field", Value: fmt.Sprintf("%.0f kV/cm", state.FieldRange.MaxField)},
		{ID: "waveform", Label: "Waveform", Value: state.Waveform},
	}
	if state.Pr > 0 {
		metrics = append(metrics,
			viewmodel.Metric{ID: "pr", Label: "Pr (P at E=0)", Value: fmt.Sprintf("%.1f µC/cm²", state.Pr)},
			viewmodel.Metric{ID: "ec_plus", Label: "+Ec", Value: fmt.Sprintf("%.0f kV/cm", state.EcPlus)},
			viewmodel.Metric{ID: "ec_minus", Label: "−Ec", Value: fmt.Sprintf("%.0f kV/cm", state.EcMinus)},
			viewmodel.Metric{ID: "psat", Label: "Psat", Value: fmt.Sprintf("%.1f µC/cm²", state.Psat)},
			viewmodel.Metric{ID: "loop_area", Label: "Loop Area", Value: fmt.Sprintf("%.1f J/m³", state.LoopArea)},
		)
	}

	sections := []viewmodel.Section{}
	for _, mat := range state.Materials {
		if mat == nil {
			continue
		}
		sections = append(sections, viewmodel.Section{
			ID:       "material_" + mat.Name,
			Title:    mat.Name,
			Body:     materialSummary(mat),
			Category: "research",
		})
	}

	sections = append(sections, viewmodel.Section{
		ID:       "edu_pe_loop",
		Title:    "Understanding P-E Loops",
		Body:     "The P-E (Polarization-Electric Field) hysteresis loop shows how a ferroelectric material's polarization changes with applied field. Key landmarks: Ec (coercive field — where P crosses zero), Pr (remanent polarization — P at E=0), Ps (saturation). The loop area represents energy lost per cycle.",
		Category: "education",
	})
	sections = append(sections, viewmodel.Section{
		ID:       "edu_preisach",
		Title:    "Preisach Model",
		Body:     "The Preisach model decomposes hysteresis into a distribution of elementary bistable units (hysterons) on the (α,β) half-plane. The Everett function integrates over the Preisach density to compute polarization. Used for minor loop and history-dependent behavior.",
		Category: "education",
	})
	sections = append(sections, viewmodel.Section{
		ID:       "edu_landau",
		Title:    "Landau-Khalatnikov Equation",
		Body:     "γ·dP/dt = −∂G/∂P + E(t) — a time-domain ODE capturing switching dynamics. G is the Landau free energy: G = αP²/2 + βP⁴/4 + γP⁶/6. The coefficients α, β, γ are material-specific and determine loop shape.",
		Category: "education",
	})

	sections = append(sections, viewmodel.Section{
		ID:       "research_citations",
		Title:    "Literature Citations",
		Body:     "HZO parameters drawn from: Materlik et al., J. Appl. Phys. 117, 134109 (2015) — LGD coefficients for orthorhombic HfO₂. Park et al., Adv. Mater. (2015) — HZO ferroelectricity confirmation. All values are educational baselines unless marked 'validated'.",
		Category: "research",
	})

	sections = append(sections, viewmodel.Section{
		ID:       "design_sweep",
		Title:    "Design Exploration",
		Body:     "Parameter sweep guidance: vary thickness (1–20 nm) to shift Ec; vary α/β/γ Landau coefficients to change loop shape. Use Ec sensitivity analysis to match target operating voltage. Cross-reference with Module 2 for conductance mapping from polarization.",
		Category: "design",
	})

	actions := []viewmodel.Action{
		{ID: EventSelectMaterial, Label: "Change Material", Kind: viewmodel.ActionSelect},
		{ID: EventSetFieldRange, Label: "Set Field Range", Kind: viewmodel.ActionCommand},
		{ID: EventToggleSimulation, Label: "Run/Pause", Kind: viewmodel.ActionToggle},
	}

	plots := []viewmodel.PlotData{}
	if len(state.LoopPoints) > 0 {
		pts := make([]viewmodel.PlotPoint, len(state.LoopPoints))
		for i, lp := range state.LoopPoints {
			pts[i] = viewmodel.PlotPoint{X: lp.Field, Y: lp.Polarization}
		}
		plots = append(plots, viewmodel.PlotData{
			ID:     "pe_loop",
			Title:  "P-E Hysteresis Loop",
			XLabel: "Electric Field (kV/cm)",
			YLabel: "Polarization (µC/cm²)",
			Series: []viewmodel.PlotSeries{{Name: "P-E", Points: pts}},
		})
	}

	return viewmodel.ModuleSnapshot{
		Descriptor: viewmodel.ModuleDescriptor{
			ID:             viewmodel.ModuleHysteresis,
			Title:          "FeCIM Hysteresis Simulation",
			Description:    "P-E curves, Preisach model, Landau-Khalatnikov solver, and material presets.",
			Status:         viewmodel.StatusFunctional,
			BoundaryNotice: "SIMULATION OUTPUT — Not measured device data. HZO parameters from Materlik (2015), Park (2015). Results are educational estimates based on published physics models.",
		},
		Metrics:  metrics,
		Sections: sections,
		Actions:  actions,
		Plots:    plots,
	}
}
