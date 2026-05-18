//go:build !cgo

package gogpuapp

import (
	"testing"

	"fecim-lattice-tools/shared/viewmodel"
	circuitsvm "fecim-lattice-tools/shared/viewmodel/circuits"

	"github.com/gogpu/gg"
)

func TestCircuitsOverlayStateIncludesHalfSelectStress(t *testing.T) {
	vm := circuitsvm.New()
	if err := vm.ApplyAction(viewmodel.Action{
		ID:      circuitsvm.ActionSetOperationMode,
		Kind:    viewmodel.ActionSelect,
		Payload: map[string]string{"mode": circuitsvm.OperationWrite},
	}); err != nil {
		t.Fatalf("set write mode: %v", err)
	}
	state := circuitsOverlayStateFromSnapshot(vm.Snapshot())

	if state.halfSelectState != circuitsvm.HalfSelectStateColumnWriteActive {
		t.Fatalf("halfSelectState = %q, want %q", state.halfSelectState, circuitsvm.HalfSelectStateColumnWriteActive)
	}
	if state.halfSelectCells != 7 {
		t.Fatalf("halfSelectCells = %d, want 7", state.halfSelectCells)
	}
	if state.stressBudget != "400 pulses/level" {
		t.Fatalf("stressBudget = %q, want 400 pulses/level", state.stressBudget)
	}
}

func TestCircuitsOverlayStateIncludesPVTInvestigationSummaries(t *testing.T) {
	vm := circuitsvm.New()
	state := circuitsOverlayStateFromSnapshot(vm.Snapshot())

	if state.pvtProcessYield != "100.0% (20/20)" {
		t.Fatalf("pvtProcessYield = %q, want 100.0%% (20/20)", state.pvtProcessYield)
	}
	if state.pvtTempSweep != "pass -40/25/85/125 C" {
		t.Fatalf("pvtTempSweep = %q, want temperature sweep summary", state.pvtTempSweep)
	}
	if state.pvtCornerENOB != "FF 4.51 / TT 4.42 / SS 4.30 bits" {
		t.Fatalf("pvtCornerENOB = %q, want FF/TT/SS summary", state.pvtCornerENOB)
	}
	if state.pvtNoiseCeiling != "13.61 bits at 16-bit ADC" {
		t.Fatalf("pvtNoiseCeiling = %q, want thermal-noise ceiling", state.pvtNoiseCeiling)
	}
}

func TestCircuitsOverlayStateIncludesReferenceSpecSummaries(t *testing.T) {
	vm := circuitsvm.New()
	state := circuitsOverlayStateFromSnapshot(vm.Snapshot())

	if state.specPowerLatency != "5.8 mW / 76 ns" {
		t.Fatalf("specPowerLatency = %q, want reference power/latency", state.specPowerLatency)
	}
	if state.specThroughput != "0.84 GOPS / 145 GOPS/W" {
		t.Fatalf("specThroughput = %q, want reference throughput", state.specThroughput)
	}
	if state.specCompliance != "OK: DAC/ADC cover 30 levels" {
		t.Fatalf("specCompliance = %q, want compliance summary", state.specCompliance)
	}
}

func TestCircuitsOverlayStateIncludesReferenceTimingSummaries(t *testing.T) {
	vm := circuitsvm.New()
	state := circuitsOverlayStateFromSnapshot(vm.Snapshot())

	if state.timingActive != "READ 76 ns total" {
		t.Fatalf("timingActive = %q, want READ timing summary", state.timingActive)
	}
	if state.timingActivePhases != "DAC 10 / Array 5 / TIA 11 / ADC 50 ns" {
		t.Fatalf("timingActivePhases = %q, want read timing phases", state.timingActivePhases)
	}
}

func TestCircuitsOverlayStateIncludesReferenceTimingWaveformMetadata(t *testing.T) {
	vm := circuitsvm.New()
	if err := vm.ApplyAction(viewmodel.Action{
		ID:      circuitsvm.ActionSetTimingOperation,
		Kind:    viewmodel.ActionSelect,
		Payload: map[string]string{"operation": "COMPUTE"},
	}); err != nil {
		t.Fatalf("set compute timing operation: %v", err)
	}

	state := circuitsOverlayStateFromSnapshot(vm.Snapshot())
	if state.timingWaveformSignals != "COMPUTE: CLK, INPUT_VALID, DAC_ALL, ARRAY_SETTLE, ADC_ALL, OUTPUT_VALID" {
		t.Fatalf("timingWaveformSignals = %q, want compute signal list", state.timingWaveformSignals)
	}
	if state.timingWaveformMarkers != "COMPUTE markers: 0ns, 19ns, 38ns, 57ns, 76ns" {
		t.Fatalf("timingWaveformMarkers = %q, want compute time markers", state.timingWaveformMarkers)
	}
	if state.timingWaveformPhases != "COMPUTE phases: DAC 10ns, Array 5ns, TIA+ADC 61ns" {
		t.Fatalf("timingWaveformPhases = %q, want compute phase markers", state.timingWaveformPhases)
	}
}

func TestCircuitsOverlayFindsActiveTimingWaveformPlot(t *testing.T) {
	vm := circuitsvm.New()
	if err := vm.ApplyAction(viewmodel.Action{
		ID:      circuitsvm.ActionSetTimingOperation,
		Kind:    viewmodel.ActionSelect,
		Payload: map[string]string{"operation": "WRITE"},
	}); err != nil {
		t.Fatalf("set write timing operation: %v", err)
	}

	plot := activeTimingWaveformPlot(vm.Snapshot())
	if plot == nil {
		t.Fatal("expected active timing waveform plot")
	}
	if plot.Title != "WRITE Timing Waveform" {
		t.Fatalf("timing waveform plot title = %q, want WRITE Timing Waveform", plot.Title)
	}
	if len(plot.Series) != 6 || plot.Series[4].Name != "V_PROG" {
		t.Fatalf("timing waveform series = %+v, want WRITE signal plot", plot.Series)
	}
}

func TestDrawTimingWaveformStripDrawsActivePlot(t *testing.T) {
	vm := circuitsvm.New()
	plot := activeTimingWaveformPlot(vm.Snapshot())
	if plot == nil {
		t.Fatal("expected active timing waveform plot")
	}

	dc := gg.NewContext(360, 160)
	defer dc.Close()
	dc.SetRGBA(0.04, 0.07, 0.06, 1)
	dc.DrawRectangle(0, 0, 360, 160)
	dc.Fill()
	if err := dc.FlushGPU(); err != nil {
		t.Fatalf("flush base waveform strip frame: %v", err)
	}
	before := imageSignature(dc.Image())

	drawTimingWaveformStrip(dc, *plot, 12, 12, 320, 120, "READ")
	if err := dc.FlushGPU(); err != nil {
		t.Fatalf("flush waveform strip frame: %v", err)
	}
	after := imageSignature(dc.Image())
	if after == before {
		t.Fatal("timing waveform strip did not draw into the frame")
	}
}

func TestCircuitsOverlayStateIncludesOperationLogSummaries(t *testing.T) {
	vm := circuitsvm.New()
	if err := vm.ApplyAction(viewmodel.Action{
		ID:      circuitsvm.ActionSetOperationMode,
		Kind:    viewmodel.ActionSelect,
		Payload: map[string]string{"mode": circuitsvm.OperationWrite},
	}); err != nil {
		t.Fatalf("set write mode: %v", err)
	}
	if err := vm.ApplyAction(viewmodel.Action{ID: circuitsvm.ActionRunWrite, Kind: viewmodel.ActionCommand}); err != nil {
		t.Fatalf("run write: %v", err)
	}

	state := circuitsOverlayStateFromSnapshot(vm.Snapshot())
	if state.operationLogLatest != "operation #2: WRITE level 15 to cell [0,0] using Preisach (Level-based)" {
		t.Fatalf("operationLogLatest = %q, want latest write entry", state.operationLogLatest)
	}
	if state.operationLogRecent != "control #1: Operation mode set to write | operation #2: WRITE level 15 to cell [0,0] using Preisach (Level-based)" {
		t.Fatalf("operationLogRecent = %q, want compact recent log", state.operationLogRecent)
	}
}

func TestCircuitsOverlayStateIncludesLoggerVerbosity(t *testing.T) {
	vm := circuitsvm.New()
	if err := vm.ApplyAction(viewmodel.Action{
		ID:      circuitsvm.ActionSetLoggerVerbosity,
		Kind:    viewmodel.ActionSelect,
		Payload: map[string]string{"verbosity": "trace"},
	}); err != nil {
		t.Fatalf("set logger verbosity: %v", err)
	}

	state := circuitsOverlayStateFromSnapshot(vm.Snapshot())
	if state.loggerVerbosity != "trace" {
		t.Fatalf("loggerVerbosity = %q, want trace", state.loggerVerbosity)
	}
	if state.loggerDetail != "trace: every UI update and simulation tick" {
		t.Fatalf("loggerDetail = %q, want trace detail", state.loggerDetail)
	}
}

func TestCircuitsOverlayStateIncludesOperationLogExportStatus(t *testing.T) {
	vm := circuitsvm.New()
	if err := vm.ApplyAction(viewmodel.Action{ID: circuitsvm.ActionRunCompute, Kind: viewmodel.ActionCommand}); err != nil {
		t.Fatalf("run compute: %v", err)
	}
	if err := vm.ApplyAction(viewmodel.Action{ID: circuitsvm.ActionExportOperationLog, Kind: viewmodel.ActionCommand}); err != nil {
		t.Fatalf("export operation log: %v", err)
	}

	state := circuitsOverlayStateFromSnapshot(vm.Snapshot())
	if state.operationLogExport != "buffered 1 entries" {
		t.Fatalf("operationLogExport = %q, want buffered export status", state.operationLogExport)
	}
}

func TestCircuitsOverlayStateIncludesReferenceSpecExportStatus(t *testing.T) {
	vm := circuitsvm.New()
	if err := vm.ApplyAction(viewmodel.Action{ID: circuitsvm.ActionExportReferenceSpecs, Kind: viewmodel.ActionCommand}); err != nil {
		t.Fatalf("export reference specs: %v", err)
	}

	state := circuitsOverlayStateFromSnapshot(vm.Snapshot())
	if state.referenceSpecExport != "buffered 64 cells" {
		t.Fatalf("referenceSpecExport = %q, want buffered export status", state.referenceSpecExport)
	}
}

func TestCircuitsOverlayStateIncludesReferenceTimingExportStatus(t *testing.T) {
	vm := circuitsvm.New()
	if err := vm.ApplyAction(viewmodel.Action{ID: circuitsvm.ActionExportReferenceTiming, Kind: viewmodel.ActionCommand}); err != nil {
		t.Fatalf("export reference timing: %v", err)
	}

	state := circuitsOverlayStateFromSnapshot(vm.Snapshot())
	if state.referenceTimingExport != "buffered 3 operations" {
		t.Fatalf("referenceTimingExport = %q, want buffered export status", state.referenceTimingExport)
	}
}

func TestCircuitsOverlayStateIncludesReferenceTimingSVGExportStatus(t *testing.T) {
	vm := circuitsvm.New()
	if err := vm.ApplyAction(viewmodel.Action{
		ID:      circuitsvm.ActionSetTimingOperation,
		Kind:    viewmodel.ActionSelect,
		Payload: map[string]string{"operation": "WRITE"},
	}); err != nil {
		t.Fatalf("set write timing operation: %v", err)
	}
	if err := vm.ApplyAction(viewmodel.Action{ID: circuitsvm.ActionExportReferenceTimingSVG, Kind: viewmodel.ActionCommand}); err != nil {
		t.Fatalf("export reference timing SVG: %v", err)
	}

	state := circuitsOverlayStateFromSnapshot(vm.Snapshot())
	if state.referenceTimingSVGExport != "buffered WRITE waveform" {
		t.Fatalf("referenceTimingSVGExport = %q, want buffered SVG export status", state.referenceTimingSVGExport)
	}
}

func TestCircuitsOverlayStateIncludesReferenceTimingAnimationStatus(t *testing.T) {
	vm := circuitsvm.New()
	if err := vm.ApplyAction(viewmodel.Action{
		ID:      circuitsvm.ActionSetTimingOperation,
		Kind:    viewmodel.ActionSelect,
		Payload: map[string]string{"operation": "COMPUTE"},
	}); err != nil {
		t.Fatalf("set compute timing operation: %v", err)
	}
	if err := vm.ApplyAction(viewmodel.Action{ID: circuitsvm.ActionAnimateReferenceTiming, Kind: viewmodel.ActionCommand}); err != nil {
		t.Fatalf("animate reference timing: %v", err)
	}

	state := circuitsOverlayStateFromSnapshot(vm.Snapshot())
	want := "COMPUTE timing animation step 1/6: Phase 1: INPUT_VALID asserted (0ns)..."
	if state.referenceTimingAnimation != want {
		t.Fatalf("referenceTimingAnimation = %q, want %q", state.referenceTimingAnimation, want)
	}
}

func TestCircuitsOverlayStateIncludesComputeRunSummary(t *testing.T) {
	vm := circuitsvm.New()
	if err := vm.ApplyAction(viewmodel.Action{ID: circuitsvm.ActionRunCompute, Kind: viewmodel.ActionCommand}); err != nil {
		t.Fatalf("run compute: %v", err)
	}

	state := circuitsOverlayStateFromSnapshot(vm.Snapshot())
	if state.computeRun != "8x8 / 8 rows / 64 cells" {
		t.Fatalf("computeRun = %q, want compute-run summary", state.computeRun)
	}
	if state.computeRunPeak == "not evaluated" || state.computeRunPeak == "" {
		t.Fatalf("computeRunPeak = %q, want peak-current summary", state.computeRunPeak)
	}
}
