package gogpuapp

import "fecim-lattice-tools/shared/viewmodel"

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
	ports := BuildAppPorts()
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

func (m *AppModel) SelectModule(id viewmodel.ModuleID) bool {
	for i, p := range m.Ports {
		if p.Descriptor().ID == id {
			m.ActiveIndex = i
			m.ActiveModuleID = id
			return true
		}
	}
	return false
}

func BuildAppPorts() []viewmodel.ModulePort {
	entries := ModuleEngineRegistry()
	ports := make([]viewmodel.ModulePort, 0, len(entries))
	for _, entry := range entries {
		ports = append(ports, entry.New())
	}
	return ports
}
