package viewmodel

import "errors"

var ErrUnsupportedAction = errors.New("viewmodel: unsupported action")

type StaticModule struct {
	descriptor ModuleDescriptor
	sections   []Section
	metrics    []Metric
	actions    []Action
}

func NewStaticModule(descriptor ModuleDescriptor, sections []Section) *StaticModule {
	return &StaticModule{descriptor: descriptor, sections: append([]Section(nil), sections...)}
}

func (m *StaticModule) Descriptor() ModuleDescriptor { return m.descriptor }

func (m *StaticModule) Snapshot() ModuleSnapshot {
	return ModuleSnapshot{
		Descriptor: m.descriptor,
		Metrics:    append([]Metric(nil), m.metrics...),
		Sections:   append([]Section(nil), m.sections...),
		Actions:    append([]Action(nil), m.actions...),
	}
}

func (m *StaticModule) ApplyAction(Action) error { return ErrUnsupportedAction }
func (m *StaticModule) Start()                   {}
func (m *StaticModule) Stop()                    {}
