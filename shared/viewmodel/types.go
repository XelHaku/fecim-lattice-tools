package viewmodel

import "time"

type ModuleID string

const (
	ModuleHysteresis ModuleID = "hysteresis"
	ModuleCrossbar   ModuleID = "crossbar"
	ModuleMNIST      ModuleID = "mnist"
	ModuleCircuits   ModuleID = "circuits"
	ModuleComparison ModuleID = "comparison"
	ModuleEDA        ModuleID = "eda"
	ModuleDocs       ModuleID = "docs"
)

const (
	StatusPlaceholder = "placeholder"
	StatusFunctional  = "functional"
	StatusFallback    = "fyne-fallback"
)

type ModuleDescriptor struct {
	ID          ModuleID
	Title       string
	Description string
	Status      string
}

type Metric struct {
	ID         string
	Label      string
	Value      string
	Unit       string
	Confidence string
}

type Section struct {
	ID    string
	Title string
	Body  string
}

type ActionKind string

const (
	ActionCommand ActionKind = "command"
	ActionToggle  ActionKind = "toggle"
	ActionSelect  ActionKind = "select"
)

type Action struct {
	ID      string
	Label   string
	Kind    ActionKind
	Payload map[string]string
}

type ModuleSnapshot struct {
	Descriptor ModuleDescriptor
	Metrics    []Metric
	Sections   []Section
	Actions    []Action
	UpdatedAt  time.Time
}

type ModulePort interface {
	Descriptor() ModuleDescriptor
	Snapshot() ModuleSnapshot
	ApplyAction(Action) error
	Start()
	Stop()
}

func KnownDescriptors() []ModuleDescriptor {
	return []ModuleDescriptor{
		{ID: ModuleHysteresis, Title: "FeCIM Hysteresis Simulation", Description: "P-E curves, Preisach model, Landau-Khalatnikov solver, and material presets.", Status: StatusPlaceholder},
		{ID: ModuleCrossbar, Title: "FeCIM Crossbar Array Visualization", Description: "Matrix-vector multiply, IR drop, sneak paths, drift, and conductance quantization.", Status: StatusPlaceholder},
		{ID: ModuleMNIST, Title: "FeCIM MNIST Neural Network", Description: "Educational CIM inference pipeline with quantized weights and reproducible metrics.", Status: StatusPlaceholder},
		{ID: ModuleCircuits, Title: "FeCIM Peripheral Circuits Visualizer", Description: "DAC, ADC, TIA, read path, write path, and ISPP circuit behavior.", Status: StatusPlaceholder},
		{ID: ModuleComparison, Title: "FeCIM Comparison", Description: "Evidence-first technology comparison and scenario analysis.", Status: StatusPlaceholder},
		{ID: ModuleEDA, Title: "FeCIM EDA Design Suite", Description: "SPICE, Verilog, Liberty, DEF, LEF, and OpenLane-oriented export workflows.", Status: StatusPlaceholder},
		{ID: ModuleDocs, Title: "Documentation", Description: "Curriculum, validation references, trust boundaries, and research notes.", Status: StatusPlaceholder},
	}
}
