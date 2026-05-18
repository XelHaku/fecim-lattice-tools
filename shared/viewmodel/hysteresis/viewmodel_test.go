package hysteresis

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fecim-lattice-tools/shared/viewmodel"
)

func TestModuleImplementsModulePort(t *testing.T) {
	var m viewmodel.ModulePort = New()
	if m == nil {
		t.Fatal("New() returned nil")
	}
}

func TestDescriptorHasCorrectID(t *testing.T) {
	m := New()
	d := m.Descriptor()
	if d.ID != viewmodel.ModuleHysteresis {
		t.Errorf("Descriptor().ID = %v, want %v", d.ID, viewmodel.ModuleHysteresis)
	}
	if d.Status != viewmodel.StatusFunctional {
		t.Errorf("Descriptor().Status = %v, want %v", d.Status, viewmodel.StatusFunctional)
	}
}

func TestSnapshotContainsMetrics(t *testing.T) {
	m := New()
	s := m.Snapshot()
	if len(s.Metrics) == 0 {
		t.Error("Snapshot().Metrics is empty, expected material metrics")
	}
	if s.Descriptor.ID != viewmodel.ModuleHysteresis {
		t.Errorf("Snapshot().Descriptor.ID = %v", s.Descriptor.ID)
	}
}

func TestSnapshotHasSections(t *testing.T) {
	m := New()
	s := m.Snapshot()
	if len(s.Sections) == 0 {
		t.Error("Snapshot().Sections is empty")
	}
}

func TestApplyActionUnsupported(t *testing.T) {
	m := New()
	err := m.ApplyAction(viewmodel.Action{ID: "unknown", Kind: viewmodel.ActionCommand})
	if err != viewmodel.ErrUnsupportedAction {
		t.Errorf("ApplyAction error = %v, want %v", err, viewmodel.ErrUnsupportedAction)
	}
}

func TestApplyActionSelectMaterial(t *testing.T) {
	m := New()
	matName := m.Descriptor().Title
	_ = matName
	s := m.Snapshot()
	var firstMaterial string
	for _, m := range s.Metrics {
		if m.ID == "material" {
			firstMaterial = m.Value
			break
		}
	}
	if firstMaterial == "" {
		t.Fatal("no initial material found")
	}
	err := m.ApplyAction(viewmodel.Action{
		ID:      EventSelectMaterial,
		Kind:    viewmodel.ActionSelect,
		Payload: map[string]string{"material": firstMaterial},
	})
	if err != nil {
		t.Fatalf("ApplyAction select_material(%q): %v", firstMaterial, err)
	}
}

func TestApplyActionToggleSimulation(t *testing.T) {
	m := New()
	m.ApplyAction(viewmodel.Action{ID: EventToggleSimulation, Kind: viewmodel.ActionToggle})
	m.ApplyAction(viewmodel.Action{ID: EventToggleSimulation, Kind: viewmodel.ActionToggle})
}

func TestSnapshotExposesWaveformAndCSVExportActions(t *testing.T) {
	m := New()
	s := m.Snapshot()
	actions := snapshotActionsByID(s)

	if _, ok := actions[EventSetWaveform]; !ok {
		t.Fatalf("snapshot actions missing %q", EventSetWaveform)
	}
	if _, ok := actions[EventExportCSV]; !ok {
		t.Fatalf("snapshot actions missing %q", EventExportCSV)
	}
}

func TestSnapshotExposesPUNDAndFORCActions(t *testing.T) {
	m := New()
	actions := snapshotActionsByID(m.Snapshot())

	if _, ok := actions["run_pund"]; !ok {
		t.Fatalf("snapshot actions missing run_pund")
	}
	if _, ok := actions["run_forc"]; !ok {
		t.Fatalf("snapshot actions missing run_forc")
	}
	for _, id := range []string{"export_pund_csv", "export_forc_sweep_csv", "export_forc_matrix_csv", "export_forc_metadata_json"} {
		if _, ok := actions[id]; !ok {
			t.Fatalf("snapshot actions missing %s", id)
		}
	}
}

func TestApplyActionSetWaveformUpdatesSnapshotAndLoop(t *testing.T) {
	m := New()
	before := m.Snapshot()
	beforeFirst := before.Plots[0].Series[0].Points[0]

	err := m.ApplyAction(viewmodel.Action{
		ID:      EventSetWaveform,
		Kind:    viewmodel.ActionSelect,
		Payload: map[string]string{"waveform": "triangle"},
	})
	if err != nil {
		t.Fatalf("ApplyAction set_waveform: %v", err)
	}

	after := m.Snapshot()
	if got := snapshotMetricValue(after, "waveform"); got != "triangle" {
		t.Fatalf("waveform metric = %q, want triangle", got)
	}
	afterFirst := after.Plots[0].Series[0].Points[0]
	if afterFirst == beforeFirst {
		t.Fatal("first P-E loop point did not change after selecting triangle waveform")
	}
}

func TestApplyActionSetWaveformRejectsUnknownWaveform(t *testing.T) {
	m := New()

	err := m.ApplyAction(viewmodel.Action{
		ID:      EventSetWaveform,
		Kind:    viewmodel.ActionSelect,
		Payload: map[string]string{"waveform": "sawtooth"},
	})
	if err == nil {
		t.Fatal("ApplyAction accepted unknown waveform")
	}
	if got := snapshotMetricValue(m.Snapshot(), "waveform"); got != "sine" {
		t.Fatalf("waveform metric after invalid action = %q, want sine", got)
	}
}

func TestApplyActionExportCSVBuffersLoopData(t *testing.T) {
	m := New()

	if err := m.ApplyAction(viewmodel.Action{ID: EventExportCSV, Kind: viewmodel.ActionCommand}); err != nil {
		t.Fatalf("ApplyAction export_csv: %v", err)
	}

	s := m.Snapshot()
	if got := snapshotMetricValue(s, "csv_export_path"); got != "artifact buffer" {
		t.Fatalf("csv_export_path = %q, want artifact buffer", got)
	}
	if got := snapshotMetricValue(s, "csv_export"); !strings.Contains(got, "buffered 200 points") {
		t.Fatalf("csv_export = %q, want buffered point count", got)
	}
	if got := snapshotMetricValue(s, "csv_export_bytes"); got == "" || got == "0 bytes" {
		t.Fatalf("csv_export_bytes = %q, want non-zero bytes", got)
	}
}

func TestApplyActionExportCSVWritesValidatedPath(t *testing.T) {
	m := New()
	path := filepath.Join(t.TempDir(), "exports", "pe-loop.csv")

	if err := m.ApplyAction(viewmodel.Action{
		ID:      EventExportCSV,
		Kind:    viewmodel.ActionCommand,
		Payload: map[string]string{"path": path},
	}); err != nil {
		t.Fatalf("ApplyAction export_csv path: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read exported CSV: %v", err)
	}
	text := string(content)
	if !strings.HasPrefix(text, "Index,E_field_kV_cm,Polarization_uC_cm2\n") {
		t.Fatalf("CSV header = %q", strings.SplitN(text, "\n", 2)[0])
	}
	if got := strings.Count(text, "\n"); got != 201 {
		t.Fatalf("CSV newline count = %d, want 201", got)
	}
	if got := snapshotMetricValue(m.Snapshot(), "csv_export_path"); got != path {
		t.Fatalf("csv_export_path = %q, want %q", got, path)
	}
}

func TestApplyActionExportCSVRejectsTraversalPath(t *testing.T) {
	m := New()

	err := m.ApplyAction(viewmodel.Action{
		ID:      EventExportCSV,
		Kind:    viewmodel.ActionCommand,
		Payload: map[string]string{"path": "../escape.csv"},
	})
	if err == nil {
		t.Fatal("ApplyAction accepted traversal export path")
	}
}

func TestApplyActionRunPUNDPopulatesDiagnosticSummary(t *testing.T) {
	m := New()

	if err := m.ApplyAction(viewmodel.Action{ID: "run_pund", Kind: viewmodel.ActionCommand}); err != nil {
		t.Fatalf("ApplyAction run_pund: %v", err)
	}

	s := m.Snapshot()
	if got := snapshotMetricValue(s, "pund_switching_positive"); got == "" || got == "0.000e+00 C" {
		t.Fatalf("pund_switching_positive = %q, want non-zero charge", got)
	}
	if got := snapshotMetricValue(s, "pund_switching_negative"); got == "" || got == "0.000e+00 C" {
		t.Fatalf("pund_switching_negative = %q, want non-zero charge", got)
	}
	if got := snapshotMetricValue(s, "pund_switching_ratio"); got == "" {
		t.Fatal("missing pund_switching_ratio metric")
	}
	section := snapshotSectionBody(s, "diagnostic_pund")
	if !strings.Contains(section, "QP=") || !strings.Contains(section, "Switching ratio") {
		t.Fatalf("PUND diagnostic section = %q, want charge and ratio summary", section)
	}
}

func TestApplyActionRunPUNDExposesCurrentWaveformPlot(t *testing.T) {
	m := New()

	if err := m.ApplyAction(viewmodel.Action{ID: "run_pund", Kind: viewmodel.ActionCommand}); err != nil {
		t.Fatalf("ApplyAction run_pund: %v", err)
	}

	plot := snapshotPlotByID(m.Snapshot(), "pund_current_waveforms")
	if plot.ID == "" {
		t.Fatal("missing pund_current_waveforms plot")
	}
	if plot.XLabel != "Time (ns)" || plot.YLabel != "Current (A)" {
		t.Fatalf("PUND plot labels = %q/%q, want Time (ns)/Current (A)", plot.XLabel, plot.YLabel)
	}
	if len(plot.Series) != 4 {
		t.Fatalf("PUND plot series count = %d, want P/U/N/D", len(plot.Series))
	}
	for _, wantName := range []string{"P", "U", "N", "D"} {
		series := plotSeriesByName(plot, wantName)
		if len(series.Points) < 2 {
			t.Fatalf("PUND %s series point count = %d, want at least 2", wantName, len(series.Points))
		}
		if series.Points[1].X <= series.Points[0].X {
			t.Fatalf("PUND %s series time did not increase: %+v", wantName, series.Points[:2])
		}
	}
}

func TestApplyActionRunFORCPopulatesDensitySummary(t *testing.T) {
	m := New()

	if err := m.ApplyAction(viewmodel.Action{
		ID:      "run_forc",
		Kind:    viewmodel.ActionCommand,
		Payload: map[string]string{"reversals": "13"},
	}); err != nil {
		t.Fatalf("ApplyAction run_forc: %v", err)
	}

	s := m.Snapshot()
	if got := snapshotMetricValue(s, "forc_curves"); got != "13" {
		t.Fatalf("forc_curves = %q, want 13", got)
	}
	if got := snapshotMetricValue(s, "forc_density_peak"); got == "" || got == "0.000e+00" {
		t.Fatalf("forc_density_peak = %q, want non-zero density peak", got)
	}
	section := snapshotSectionBody(s, "diagnostic_forc")
	if !strings.Contains(section, "peak_density=") || !strings.Contains(section, "density_range=") {
		t.Fatalf("FORC diagnostic section = %q, want density summary", section)
	}
}

func TestApplyActionRunFORCExposesDensityHeatmapPlot(t *testing.T) {
	m := New()

	if err := m.ApplyAction(viewmodel.Action{
		ID:      "run_forc",
		Kind:    viewmodel.ActionCommand,
		Payload: map[string]string{"reversals": "13"},
	}); err != nil {
		t.Fatalf("ApplyAction run_forc: %v", err)
	}

	plot := snapshotPlotByID(m.Snapshot(), "forc_density_heatmap")
	if plot.ID == "" {
		t.Fatal("missing forc_density_heatmap plot")
	}
	if plot.XLabel != "Ea (V/m)" || plot.YLabel != "Eb (V/m)" {
		t.Fatalf("FORC plot labels = %q/%q, want Ea (V/m)/Eb (V/m)", plot.XLabel, plot.YLabel)
	}
	if len(plot.Series) != 1 {
		t.Fatalf("FORC plot series count = %d, want one density series", len(plot.Series))
	}
	points := plot.Series[0].Points
	if len(points) != 13*13 {
		t.Fatalf("FORC density point count = %d, want 169", len(points))
	}
	foundDensity := false
	for _, point := range points {
		if point.V != 0 {
			foundDensity = true
			break
		}
	}
	if !foundDensity {
		t.Fatal("FORC density heatmap plot has no non-zero density values")
	}
}

func TestApplyActionExportPUNDCSVBuffersSummary(t *testing.T) {
	m := New()

	if err := m.ApplyAction(viewmodel.Action{ID: "export_pund_csv", Kind: viewmodel.ActionCommand}); err != nil {
		t.Fatalf("ApplyAction export_pund_csv: %v", err)
	}

	s := m.Snapshot()
	if got := snapshotMetricValue(s, "pund_export_path"); got != "artifact buffer" {
		t.Fatalf("pund_export_path = %q, want artifact buffer", got)
	}
	if got := snapshotMetricValue(s, "pund_export"); !strings.Contains(got, "buffered") {
		t.Fatalf("pund_export = %q, want buffered status", got)
	}
	if got := snapshotMetricValue(s, "pund_export_bytes"); got == "" || got == "0 bytes" {
		t.Fatalf("pund_export_bytes = %q, want non-zero bytes", got)
	}
}

func TestApplyActionExportFORCArtifactsWritesValidatedPaths(t *testing.T) {
	m := New()
	dir := t.TempDir()
	sweepPath := filepath.Join(dir, "forc-sweep.csv")
	matrixPath := filepath.Join(dir, "forc-matrix.csv")
	metaPath := filepath.Join(dir, "forc-meta.json")

	for _, tc := range []struct {
		id     string
		path   string
		header string
	}{
		{"export_forc_sweep_csv", sweepPath, "reversal_field_vm,applied_field_vm,polarization_cm2\n"},
		{"export_forc_matrix_csv", matrixPath, "Ea_Vm,Eb_Vm,density\n"},
		{"export_forc_metadata_json", metaPath, "{\n"},
	} {
		if err := m.ApplyAction(viewmodel.Action{
			ID:      tc.id,
			Kind:    viewmodel.ActionCommand,
			Payload: map[string]string{"path": tc.path, "reversals": "9"},
		}); err != nil {
			t.Fatalf("ApplyAction %s: %v", tc.id, err)
		}
		content, err := os.ReadFile(tc.path)
		if err != nil {
			t.Fatalf("read %s: %v", tc.path, err)
		}
		if !strings.HasPrefix(string(content), tc.header) {
			t.Fatalf("%s header = %q, want prefix %q", tc.id, strings.SplitN(string(content), "\n", 2)[0], strings.TrimSpace(tc.header))
		}
	}

	s := m.Snapshot()
	if got := snapshotMetricValue(s, "forc_sweep_export_path"); got != sweepPath {
		t.Fatalf("forc_sweep_export_path = %q, want %q", got, sweepPath)
	}
	if got := snapshotMetricValue(s, "forc_matrix_export_path"); got != matrixPath {
		t.Fatalf("forc_matrix_export_path = %q, want %q", got, matrixPath)
	}
	if got := snapshotMetricValue(s, "forc_metadata_export_path"); got != metaPath {
		t.Fatalf("forc_metadata_export_path = %q, want %q", got, metaPath)
	}
}

func TestMaterialSummary(t *testing.T) {
	summary := materialSummary(nil)
	if summary != "N/A" {
		t.Errorf("materialSummary(nil) = %q, want N/A", summary)
	}
}

func TestStateDefaults(t *testing.T) {
	m := New()
	s := m.Snapshot()
	var matMetric *viewmodel.Metric
	for i := range s.Metrics {
		if s.Metrics[i].ID == "material" {
			matMetric = &s.Metrics[i]
			break
		}
	}
	if matMetric == nil {
		t.Fatal("no material metric found")
	}
	if matMetric.Value == "" {
		t.Error("material metric value is empty")
	}
}

func snapshotActionsByID(snapshot viewmodel.ModuleSnapshot) map[string]viewmodel.Action {
	actions := make(map[string]viewmodel.Action, len(snapshot.Actions))
	for _, action := range snapshot.Actions {
		actions[action.ID] = action
	}
	return actions
}

func snapshotMetricValue(snapshot viewmodel.ModuleSnapshot, id string) string {
	for _, metric := range snapshot.Metrics {
		if metric.ID == id {
			return metric.Value
		}
	}
	return ""
}

func snapshotSectionBody(snapshot viewmodel.ModuleSnapshot, id string) string {
	for _, section := range snapshot.Sections {
		if section.ID == id {
			return section.Body
		}
	}
	return ""
}

func snapshotPlotByID(snapshot viewmodel.ModuleSnapshot, id string) viewmodel.PlotData {
	for _, plot := range snapshot.Plots {
		if plot.ID == id {
			return plot
		}
	}
	return viewmodel.PlotData{}
}

func plotSeriesByName(plot viewmodel.PlotData, name string) viewmodel.PlotSeries {
	for _, series := range plot.Series {
		if series.Name == name {
			return series
		}
	}
	return viewmodel.PlotSeries{}
}
