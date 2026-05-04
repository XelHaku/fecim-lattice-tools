package main

import (
	"fecim-lattice-tools/shared/viewmodel"
	comparisonvm "fecim-lattice-tools/shared/viewmodel/comparison"
	crossbarvm "fecim-lattice-tools/shared/viewmodel/crossbar"
	hysteresisvm "fecim-lattice-tools/shared/viewmodel/hysteresis"
)

type AppSpec struct {
	Title   string
	Command string
	Width   int
	Height  int
}

func DefaultAppSpec() AppSpec {
	return AppSpec{
		Title:   "FeCIM Lattice Tools Next",
		Command: "fecim-lattice-tools-next",
		Width:   1400,
		Height:  900,
	}
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
