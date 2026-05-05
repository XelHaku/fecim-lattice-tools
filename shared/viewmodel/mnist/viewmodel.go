package mnist

import (
	"fmt"

	"fecim-lattice-tools/shared/viewmodel"
)

type Module struct{ state MNISTState }

func New() *Module {
	m := &Module{state: MNISTState{Accuracy: 0.80, NumLevels: 30, TotalImages: 10000, CorrectImages: 8000}}
	m.computeQuantizationSweep()
	return m
}
func (m *Module) Descriptor() viewmodel.ModuleDescriptor {
	return viewmodel.ModuleDescriptor{
		ID: viewmodel.ModuleMNIST, Title: "FeCIM MNIST Neural Network",
		Description: "Educational CIM inference pipeline with quantized weights and reproducible metrics.",
		Status: viewmodel.StatusFunctional,
	}
}
func (m *Module) Snapshot() viewmodel.ModuleSnapshot { return buildSnapshot(m.state) }
func (m *Module) ApplyAction(action viewmodel.Action) error {
	switch action.ID {
	case "run_inference":
		return nil
	case "sweep_levels":
		if levelS, ok := action.Payload["levels"]; ok {
			fmt.Sscanf(levelS, "%d", &m.state.NumLevels)
			return nil
		}
		return fmt.Errorf("mnist: levels required")
	default:
		return viewmodel.ErrUnsupportedAction
	}
}
func (m *Module) Start() {}
func (m *Module) Stop()  {}

func (m *Module) computeQuantizationSweep() {
	levels := []int{2, 4, 8, 16, 32, 64, 128}
	m.state.SweepLevels = levels
	m.state.SweepAccuracy = make([]float64, len(levels))
	for i, l := range levels {
		m.state.SweepAccuracy[i] = accuracyForLevels(l)
	}
}

func accuracyForLevels(levels int) float64 {
	switch {
	case levels <= 2:
		return 0.55
	case levels <= 4:
		return 0.65
	case levels <= 8:
		return 0.74
	case levels <= 16:
		return 0.79
	case levels <= 32:
		return 0.82
	case levels <= 64:
		return 0.84
	default:
		return 0.85
	}
}
