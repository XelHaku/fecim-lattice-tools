package docs

import "fecim-lattice-tools/shared/viewmodel"

type Module struct{ state DocsState }

func New() *Module { return &Module{} }
func (m *Module) Descriptor() viewmodel.ModuleDescriptor {
	return viewmodel.ModuleDescriptor{
		ID: viewmodel.ModuleDocs, Title: "Documentation",
		Description: "Curriculum, validation references, trust boundaries, and research notes.",
		Status:      viewmodel.StatusFunctional,
	}
}
func (m *Module) Snapshot() viewmodel.ModuleSnapshot { return buildSnapshot(m.state) }
func (m *Module) ApplyAction(action viewmodel.Action) error {
	switch action.ID {
	case "search":
		if q, ok := action.Payload["query"]; ok {
			m.state.SearchQuery = q
			return nil
		}
		return viewmodel.ErrUnsupportedAction
	case "start_curriculum":
		m.state.ActivePage = "curriculum"
		return nil
	default:
		return viewmodel.ErrUnsupportedAction
	}
}
func (m *Module) Start() {}
func (m *Module) Stop()  {}
