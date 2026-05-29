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

// ModuleEngineEntry binds a module descriptor to its Default UI Surface port factory.
type ModuleEngineEntry struct {
	Descriptor viewmodel.ModuleDescriptor
	New        func() viewmodel.ModulePort
}

// ModuleEngineRegistry returns module engine entries in KnownDescriptors order.
func ModuleEngineRegistry() []ModuleEngineEntry {
	factories := map[viewmodel.ModuleID]func() viewmodel.ModulePort{
		viewmodel.ModuleComparison: func() viewmodel.ModulePort { return comparisonvm.New() },
		viewmodel.ModuleHysteresis: func() viewmodel.ModulePort { return hysteresisvm.New() },
		viewmodel.ModuleCrossbar:   func() viewmodel.ModulePort { return crossbarvm.New(8, 8) },
		viewmodel.ModuleCircuits:   func() viewmodel.ModulePort { return circuitsvm.New() },
		viewmodel.ModuleEDA:        func() viewmodel.ModulePort { return edavm.New() },
		viewmodel.ModuleMNIST:      func() viewmodel.ModulePort { return mnistvm.New() },
		viewmodel.ModuleDocs:       func() viewmodel.ModulePort { return docsvm.New() },
	}

	descriptors := viewmodel.KnownDescriptors()
	entries := make([]ModuleEngineEntry, 0, len(descriptors))
	for _, descriptor := range descriptors {
		factory := factories[descriptor.ID]
		if factory == nil {
			continue
		}
		entries = append(entries, ModuleEngineEntry{Descriptor: descriptor, New: factory})
	}
	return entries
}
