package gogpuapp

import (
	"fecim-lattice-tools/shared/viewmodel"
	circuitsvm "fecim-lattice-tools/shared/viewmodel/circuits"
	comparisonvm "fecim-lattice-tools/shared/viewmodel/comparison"
	crossbarvm "fecim-lattice-tools/shared/viewmodel/crossbar"
	docsvm "fecim-lattice-tools/shared/viewmodel/docs"
	edavm "fecim-lattice-tools/shared/viewmodel/eda"
	hysteresisvm "fecim-lattice-tools/shared/viewmodel/hysteresis"
	mnistvm "fecim-lattice-tools/shared/viewmodel/mnist"
)

type AppSpec struct {
	Title   string
	Command string
	Width   int
	Height  int
}

type AppModel struct {
	Spec           AppSpec
	Ports          []viewmodel.ModulePort
	ActiveModuleID viewmodel.ModuleID
	ActiveIndex    int
}

func DefaultAppSpec() AppSpec {
	return AppSpec{
		Title:   "FeCIM Lattice Tools",
		Command: "fecim-lattice-tools",
		Width:   1400,
		Height:  900,
	}
}

func NewAppModel(active viewmodel.ModuleID) AppModel {
	ports := BuildPlaceholderPorts()
	activeIndex := 0
	if active != "" {
		for i, p := range ports {
			if p.Descriptor().ID == active {
				activeIndex = i
				break
			}
		}
	}
	return AppModel{
		Spec:           DefaultAppSpec(),
		Ports:          ports,
		ActiveModuleID: ports[activeIndex].Descriptor().ID,
		ActiveIndex:    activeIndex,
	}
}

func (m AppModel) ActivePort() viewmodel.ModulePort {
	if m.ActiveIndex >= 0 && m.ActiveIndex < len(m.Ports) {
		return m.Ports[m.ActiveIndex]
	}
	if len(m.Ports) == 0 {
		return nil
	}
	return m.Ports[0]
}

func BuildPlaceholderPorts() []viewmodel.ModulePort {
	descriptors := viewmodel.KnownDescriptors()
	ports := make([]viewmodel.ModulePort, 0, len(descriptors))
	for _, descriptor := range descriptors {
		switch descriptor.ID {
		case viewmodel.ModuleComparison:
			ports = append(ports, comparisonvm.New())
		case viewmodel.ModuleHysteresis:
			ports = append(ports, hysteresisvm.New())
		case viewmodel.ModuleCrossbar:
			ports = append(ports, crossbarvm.New(8, 8))
		case viewmodel.ModuleCircuits:
			ports = append(ports, circuitsvm.New())
		case viewmodel.ModuleEDA:
			ports = append(ports, edavm.New())
		case viewmodel.ModuleMNIST:
			ports = append(ports, mnistvm.New())
		case viewmodel.ModuleDocs:
			ports = append(ports, docsvm.New())
		default:
			ports = append(ports, viewmodel.NewStaticModule(descriptor, []viewmodel.Section{
				{
					ID:    "migration-status",
					Title: "Migration Status",
					Body:  "This module is represented by a UI-neutral placeholder while the gogpu/ui shell reaches parity with the current Fyne implementation.",
				},
			}))
		}
	}
	return ports
}
