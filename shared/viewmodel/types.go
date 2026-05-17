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
	ID             ModuleID
	Title          string
	Description    string
	Status         string
	BoundaryNotice string
}

type Metric struct {
	ID         string
	Label      string
	Value      string
	Unit       string
	Confidence string
}

type Section struct {
	ID       string
	Title    string
	Body     string
	Category string
}

type PlotPoint struct {
	X float64
	Y float64
	V float64
}

type PlotSeries struct {
	Name   string
	Points []PlotPoint
}

type PlotData struct {
	ID     string
	Title  string
	XLabel string
	YLabel string
	Series []PlotSeries
}

type EvidenceClass string

const (
	EvidenceSimulation   EvidenceClass = "simulation"
	EvidenceLiterature   EvidenceClass = "literature-benchmark"
	EvidenceMeasured     EvidenceClass = "measured-data"
	EvidencePredicted    EvidenceClass = "predicted-structure"
	EvidencePeerReviewed EvidenceClass = "peer-reviewed-model"
)

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
	Plots      []PlotData
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
		{ID: ModuleHysteresis, Title: "FeCIM Hysteresis Simulation", Description: "P-E curves, Preisach model, Landau-Khalatnikov solver, and material presets.", Status: StatusFunctional},
		{ID: ModuleCrossbar, Title: "FeCIM Crossbar Array Visualization", Description: "Matrix-vector multiply, IR drop, sneak paths, drift, and conductance quantization.", Status: StatusFunctional},
		{ID: ModuleMNIST, Title: "FeCIM MNIST Neural Network", Description: "Educational CIM inference pipeline with quantized weights and reproducible metrics.", Status: StatusFunctional},
		{ID: ModuleCircuits, Title: "FeCIM Peripheral Circuits Visualizer", Description: "DAC, ADC, TIA, read path, write path, and ISPP circuit behavior.", Status: StatusFunctional},
		{ID: ModuleComparison, Title: "FeCIM Comparison", Description: "Evidence-first technology comparison and scenario analysis.", Status: StatusFunctional},
		{ID: ModuleEDA, Title: "FeCIM EDA Design Suite", Description: "SPICE, Verilog, Liberty, DEF, LEF, and OpenLane-oriented export workflows.", Status: StatusFunctional},
		{ID: ModuleDocs, Title: "Documentation", Description: "Curriculum, validation references, trust boundaries, and research notes.", Status: StatusFunctional},
	}
}
