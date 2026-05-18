package mnist

import (
	"fecim-lattice-tools/shared/viewmodel"
	"fmt"
)

func buildSnapshot(state MNISTState) viewmodel.ModuleSnapshot {
	metrics := []viewmodel.Metric{
		{ID: "accuracy", Label: "Accuracy", Value: fmt.Sprintf("%.1f%%", state.Accuracy*100)},
		{ID: "levels", Label: "Quantization", Value: fmt.Sprintf("%d levels", state.NumLevels)},
		{ID: "correct", Label: "Correct", Value: fmt.Sprintf("%d/%d", state.CorrectImages, state.TotalImages)},
	}
	sections := []viewmodel.Section{
		{ID: "pipeline", Title: "Inference Pipeline", Body: fmt.Sprintf("Image → Quantize → %d-level MVM → Softmax → Prediction. Baseline: %.0f%% at %d levels.", state.NumLevels, state.Accuracy*100, state.NumLevels), Category: "research"},
		{ID: "nonideality", Title: "Non-Ideality Impact", Body: "IR drop and conductance drift modeled at array level. Quantization error increases at lower level counts.", Category: "research"},
	}
	sections = append(sections, viewmodel.Section{
		ID: "edu_pipeline", Title: "CIM Inference Pipeline",
		Body:     "Image pixels → quantize to voltage levels → apply to crossbar rows → currents sum at columns (MVM) → softmax activation → digit prediction. The crossbar performs the matrix multiplication in O(1) analog time instead of O(n³) digital.",
		Category: "education",
	})
	sections = append(sections, viewmodel.Section{
		ID: "research_benchmark", Title: "Benchmark Reference",
		Body:     "80% baseline on MNIST test set (10,000 images). Educational, not validated device claim. Compare against: HZO FTJ reservoir computing (98.24%, J. Alloys Compounds 2025) — note that this is a different architecture, not FeCIM. Crossbar non-idealities reduce accuracy from ideal baseline.",
		Category: "research",
	})
	sections = append(sections, viewmodel.Section{
		ID: "design_tradeoff", Title: "Accuracy vs. Quantization",
		Body:     fmt.Sprintf("Design sweep: vary quantization levels (8–128). More levels = higher accuracy but harder to program. At %d levels, expect ~%.0f%% accuracy. At 64 levels, expect ~85-90%% (projected, not validated). Cross-reference: Module 2 for array sizing vs. accuracy.", state.NumLevels, state.Accuracy*100),
		Category: "design",
	})
	actions := []viewmodel.Action{
		{ID: "run_inference", Label: "Run Inference", Kind: viewmodel.ActionCommand},
		{ID: "sweep_levels", Label: "Sweep Levels", Kind: viewmodel.ActionCommand},
	}
	return viewmodel.ModuleSnapshot{
		Descriptor: viewmodel.ModuleDescriptor{
			ID: viewmodel.ModuleMNIST, Title: "FeCIM MNIST Neural Network",
			Description:    "Educational CIM inference pipeline with quantized weights and reproducible metrics.",
			Status:         viewmodel.StatusFunctional,
			BoundaryNotice: "EDUCATIONAL SIMULATION — Not a validated device measurement. 80% baseline is an educational target on MNIST; not comparable to silicon accuracy claims. 98.24% reference (J. Alloys Comp. 2025) uses HZO FTJ reservoir computing, a different architecture.",
		},
		Metrics: metrics, Sections: sections, Actions: actions,
		Plots: buildAccuracySweepPlots(state),
	}
}

func buildAccuracySweepPlots(state MNISTState) []viewmodel.PlotData {
	if len(state.SweepLevels) == 0 {
		return nil
	}
	pts := make([]viewmodel.PlotPoint, len(state.SweepLevels))
	for i := range state.SweepLevels {
		pts[i] = viewmodel.PlotPoint{X: float64(state.SweepLevels[i]), Y: state.SweepAccuracy[i] * 100}
	}
	return []viewmodel.PlotData{{
		ID:     "accuracy_sweep",
		Title:  "Accuracy vs Quantization Levels",
		XLabel: "Quantization Levels",
		YLabel: "Accuracy (%)",
		Series: []viewmodel.PlotSeries{{Name: "CIM accuracy", Points: pts}},
	}}
}
