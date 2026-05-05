package hysteresis

import (
	"fmt"
	"math"

	"fecim-lattice-tools/shared/physics"
	"fecim-lattice-tools/shared/viewmodel"
)

type Module struct {
	state HysteresisState
}

func New() *Module {
		materials := physics.AllMaterials()
		defaultMat := "HZO (Si-doped, Park 2015 midpoint)"
		if len(materials) > 0 && materials[0] != nil {
		defaultMat = materials[0].Name
	}
	m := &Module{
		state: HysteresisState{
			SelectedMaterial: defaultMat,
			Materials:        materials,
			FieldRange:       FieldRange{MinField: -3000, MaxField: 3000},
			Waveform:         "sine",
		},
	}
	m.computeLoopForCurrentMaterial()
	return m
}

func (m *Module) Descriptor() viewmodel.ModuleDescriptor {
	return viewmodel.ModuleDescriptor{
		ID:          viewmodel.ModuleHysteresis,
		Title:       "FeCIM Hysteresis Simulation",
		Description: "P-E curves, Preisach model, Landau-Khalatnikov solver, and material presets.",
		Status:      viewmodel.StatusFunctional,
	}
}

func (m *Module) Snapshot() viewmodel.ModuleSnapshot { return buildSnapshot(m.state) }

func (m *Module) ApplyAction(action viewmodel.Action) error {
	switch action.ID {
	case EventSelectMaterial:
		if name, ok := action.Payload["material"]; ok {
			for _, mat := range m.state.Materials {
				if mat != nil && mat.Name == name {
					m.state.SelectedMaterial = name
					m.computeLoopForCurrentMaterial()
					return nil
				}
			}
		}
		return fmt.Errorf("hysteresis: material %q not found", action.Payload["material"])
	case EventToggleSimulation:
		m.state.IsRunning = !m.state.IsRunning
		return nil
	case EventSetFieldRange:
		if minS, ok := action.Payload["min"]; ok {
			fmt.Sscanf(minS, "%f", &m.state.FieldRange.MinField)
		}
		if maxS, ok := action.Payload["max"]; ok {
			fmt.Sscanf(maxS, "%f", &m.state.FieldRange.MaxField)
		}
		m.computeLoopForCurrentMaterial()
		return nil
	default:
		return viewmodel.ErrUnsupportedAction
	}
}

func (m *Module) Start() {}
func (m *Module) Stop()  {}

func (m *Module) computeLoopForCurrentMaterial() {
	var mat *physics.HZOMaterial
	for _, candidate := range m.state.Materials {
		if candidate != nil && candidate.Name == m.state.SelectedMaterial {
			mat = candidate
			break
		}
	}
	if mat == nil {
		return
	}

	solver := physics.NewLKSolver()
	solver.ConfigureFromMaterial(mat)

	maxFieldKVcm := math.Max(math.Abs(m.state.FieldRange.MinField), math.Abs(m.state.FieldRange.MaxField))
	ecKVcm := mat.Ec * 1e-5
	if maxFieldKVcm < ecKVcm*2 {
		maxFieldKVcm = ecKVcm * 2
	}
	maxFieldSI := maxFieldKVcm * 1e5

	const numPoints = 200
	dt := 5e-5 // ~100Hz equivalent quasi-static timestep

	for cycle := 0; cycle < 2; cycle++ {
		for i := 0; i < numPoints; i++ {
			t := float64(i) * 2 * math.Pi / float64(numPoints-1)
			fieldSI := maxFieldSI * math.Sin(t)
			solver.Step(fieldSI, dt)
		}
	}

	pts := make([]LoopPoint, numPoints)
	for i := 0; i < numPoints; i++ {
		t := float64(i) * 2 * math.Pi / float64(numPoints-1)
		fieldSI := maxFieldSI * math.Sin(t)
		polSI := solver.Step(fieldSI, dt)
		pts[i] = LoopPoint{
			Field:        fieldSI * 1e-5, // V/m → kV/cm
			Polarization: polSI * 100,    // C/m² → µC/cm²
		}
	}
	m.state.LoopPoints = pts
}
