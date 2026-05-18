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
	if state.CSVExportStatus != "" {
		metrics = append(metrics,
			viewmodel.Metric{ID: "csv_export", Label: "CSV Export", Value: state.CSVExportStatus},
			viewmodel.Metric{ID: "csv_export_path", Label: "CSV Export Path", Value: state.CSVExportPath},
			viewmodel.Metric{ID: "csv_export_bytes", Label: "CSV Export Bytes", Value: fmt.Sprintf("%d bytes", state.CSVExportBytes)},
		)
	}
	if state.PUND.Available {
		metrics = append(metrics,
			viewmodel.Metric{ID: "pund_switching_positive", Label: "PUND Qsw+", Value: fmt.Sprintf("%.3e C", state.PUND.SwitchingPositive)},
			viewmodel.Metric{ID: "pund_switching_negative", Label: "PUND Qsw-", Value: fmt.Sprintf("%.3e C", state.PUND.SwitchingNegative)},
			viewmodel.Metric{ID: "pund_switching_ratio", Label: "PUND Ratio", Value: fmt.Sprintf("%.3f", state.PUND.SwitchingRatio)},
		)
		if state.PUND.ExportStatus != "" {
			metrics = append(metrics,
				viewmodel.Metric{ID: "pund_export", Label: "PUND Export", Value: state.PUND.ExportStatus},
				viewmodel.Metric{ID: "pund_export_path", Label: "PUND Export Path", Value: state.PUND.ExportPath},
				viewmodel.Metric{ID: "pund_export_bytes", Label: "PUND Export Bytes", Value: fmt.Sprintf("%d bytes", state.PUND.ExportBytes)},
			)
		}
	}
	if state.FORC.Available {
		metrics = append(metrics,
			viewmodel.Metric{ID: "forc_curves", Label: "FORC Curves", Value: fmt.Sprintf("%d", state.FORC.Curves)},
			viewmodel.Metric{ID: "forc_density_peak", Label: "FORC Peak Density", Value: fmt.Sprintf("%.3e", state.FORC.PeakDensity)},
			viewmodel.Metric{ID: "forc_density_location", Label: "FORC Peak Location", Value: fmt.Sprintf("Ea %.3e V/m, Eb %.3e V/m", state.FORC.PeakEa_Vm, state.FORC.PeakEb_Vm)},
		)
		if state.FORC.SweepExportStatus != "" {
			metrics = append(metrics,
				viewmodel.Metric{ID: "forc_sweep_export", Label: "FORC Sweep Export", Value: state.FORC.SweepExportStatus},
				viewmodel.Metric{ID: "forc_sweep_export_path", Label: "FORC Sweep Export Path", Value: state.FORC.SweepExportPath},
				viewmodel.Metric{ID: "forc_sweep_export_bytes", Label: "FORC Sweep Export Bytes", Value: fmt.Sprintf("%d bytes", state.FORC.SweepExportBytes)},
			)
		}
		if state.FORC.MatrixExportStatus != "" {
			metrics = append(metrics,
				viewmodel.Metric{ID: "forc_matrix_export", Label: "FORC Matrix Export", Value: state.FORC.MatrixExportStatus},
				viewmodel.Metric{ID: "forc_matrix_export_path", Label: "FORC Matrix Export Path", Value: state.FORC.MatrixExportPath},
				viewmodel.Metric{ID: "forc_matrix_export_bytes", Label: "FORC Matrix Export Bytes", Value: fmt.Sprintf("%d bytes", state.FORC.MatrixExportBytes)},
			)
		}
		if state.FORC.MetaExportStatus != "" {
			metrics = append(metrics,
				viewmodel.Metric{ID: "forc_metadata_export", Label: "FORC Metadata Export", Value: state.FORC.MetaExportStatus},
				viewmodel.Metric{ID: "forc_metadata_export_path", Label: "FORC Metadata Export Path", Value: state.FORC.MetaExportPath},
				viewmodel.Metric{ID: "forc_metadata_export_bytes", Label: "FORC Metadata Export Bytes", Value: fmt.Sprintf("%d bytes", state.FORC.MetaExportBytes)},
			)
		}
	}

	sections := []viewmodel.Section{}
	if state.PUND.Available {
		sections = append(sections, viewmodel.Section{
			ID:       "diagnostic_pund",
			Title:    "PUND Measurement Summary",
			Body:     state.PUND.Summary,
			Category: "research",
		})
	}
	if state.FORC.Available {
		sections = append(sections, viewmodel.Section{
			ID:       "diagnostic_forc",
			Title:    "FORC Density Summary",
			Body:     state.FORC.Summary,
			Category: "research",
		})
	}
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
		{ID: EventSetWaveform, Label: "Set Waveform", Kind: viewmodel.ActionSelect, Payload: map[string]string{"waveform": state.Waveform}},
		{ID: EventToggleSimulation, Label: "Run/Pause", Kind: viewmodel.ActionToggle},
		{ID: EventExportCSV, Label: "Export CSV", Kind: viewmodel.ActionCommand},
		{ID: EventRunPUND, Label: "Run PUND", Kind: viewmodel.ActionCommand},
		{ID: EventRunFORC, Label: "Run FORC", Kind: viewmodel.ActionCommand},
		{ID: EventExportPUNDCSV, Label: "Export PUND CSV", Kind: viewmodel.ActionCommand},
		{ID: EventExportFORCSweep, Label: "Export FORC Sweep CSV", Kind: viewmodel.ActionCommand},
		{ID: EventExportFORCMatrix, Label: "Export FORC Matrix CSV", Kind: viewmodel.ActionCommand},
		{ID: EventExportFORCMeta, Label: "Export FORC Metadata JSON", Kind: viewmodel.ActionCommand},
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
	if len(state.RetentionTimes) > 0 && len(state.RetentionPr) > 0 {
		pts := make([]viewmodel.PlotPoint, len(state.RetentionTimes))
		for i := range state.RetentionTimes {
			pts[i] = viewmodel.PlotPoint{X: state.RetentionTimes[i], Y: state.RetentionPr[i]}
		}
		plots = append(plots, viewmodel.PlotData{
			ID:     "retention",
			Title:  "Retention Prediction (Power-Law)",
			XLabel: "Time (seconds)",
			YLabel: "Pr (µC/cm²)",
			Series: []viewmodel.PlotSeries{{Name: "retention", Points: pts}},
		})
	}
	if plot, ok := buildPUNDWaveformPlot(state.PUND); ok {
		plots = append(plots, plot)
	}
	if plot, ok := buildFORCDensityHeatmapPlot(state.FORC); ok {
		plots = append(plots, plot)
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

func buildPUNDWaveformPlot(summary PUNDSummary) (viewmodel.PlotData, bool) {
	if !summary.Available || len(summary.TraceSamples) == 0 {
		return viewmodel.PlotData{}, false
	}
	pointsByPulse := map[string][]viewmodel.PlotPoint{
		"P": {},
		"U": {},
		"N": {},
		"D": {},
	}
	for _, sample := range summary.TraceSamples {
		if _, ok := pointsByPulse[sample.Pulse]; !ok {
			continue
		}
		pointsByPulse[sample.Pulse] = append(pointsByPulse[sample.Pulse], viewmodel.PlotPoint{
			X: sample.TimeS * 1e9,
			Y: sample.CurrentA,
		})
	}

	series := make([]viewmodel.PlotSeries, 0, 4)
	for _, pulse := range []string{"P", "U", "N", "D"} {
		points := pointsByPulse[pulse]
		if len(points) == 0 {
			continue
		}
		series = append(series, viewmodel.PlotSeries{Name: pulse, Points: points})
	}
	if len(series) == 0 {
		return viewmodel.PlotData{}, false
	}
	return viewmodel.PlotData{
		ID:     "pund_current_waveforms",
		Title:  "PUND Current Waveforms",
		XLabel: "Time (ns)",
		YLabel: "Current (A)",
		Series: series,
	}, true
}

func buildFORCDensityHeatmapPlot(summary FORCSummary) (viewmodel.PlotData, bool) {
	if !summary.Available || len(summary.DensitySamples) == 0 {
		return viewmodel.PlotData{}, false
	}
	points := make([]viewmodel.PlotPoint, len(summary.DensitySamples))
	for i, sample := range summary.DensitySamples {
		points[i] = viewmodel.PlotPoint{
			X: sample.Ea_Vm,
			Y: sample.Eb_Vm,
			V: sample.Density,
		}
	}
	return viewmodel.PlotData{
		ID:     "forc_density_heatmap",
		Title:  "FORC Density Heatmap",
		XLabel: "Ea (V/m)",
		YLabel: "Eb (V/m)",
		Series: []viewmodel.PlotSeries{{Name: "rho(Ea,Eb)", Points: points}},
	}, true
}
