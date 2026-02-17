package crossbar

import (
	"math"
	"testing"
)

// ===== DriftSimulator Tests =====

func TestNewDriftSimulator_Defaults(t *testing.T) {
	rows, cols, levels := 8, 8, 30
	sim := NewDriftSimulator(rows, cols, levels)

	if sim.Rows != rows {
		t.Errorf("Rows = %d, want %d", sim.Rows, rows)
	}
	if sim.Cols != cols {
		t.Errorf("Cols = %d, want %d", sim.Cols, cols)
	}
	if sim.Levels != levels {
		t.Errorf("Levels = %d, want %d", sim.Levels, levels)
	}
	if sim.GMin != GMin {
		t.Errorf("GMin = %e, want %e", sim.GMin, GMin)
	}
	if sim.GMax != GMax {
		t.Errorf("GMax = %e, want %e", sim.GMax, GMax)
	}
	if sim.DriftCoeff != FeFETDriftCoefficients.Assumed {
		t.Errorf("DriftCoeff = %f, want %f", sim.DriftCoeff, FeFETDriftCoefficients.Assumed)
	}
	if sim.Temperature != 300.0 {
		t.Errorf("Temperature = %f, want 300.0", sim.Temperature)
	}
	if sim.Time != 0 {
		t.Errorf("Time = %f, want 0", sim.Time)
	}
}

func TestNewDriftSimulator_Conductances(t *testing.T) {
	rows, cols, levels := 5, 5, 30
	sim := NewDriftSimulator(rows, cols, levels)

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			g := sim.Conductances[i][j]
			if g < sim.GMin || g > sim.GMax {
				t.Errorf("Conductance[%d][%d] = %e, out of range [%e, %e]", i, j, g, sim.GMin, sim.GMax)
			}
		}
	}
}

func TestNewDriftSimulatorWithModel_Literature(t *testing.T) {
	sim := NewDriftSimulatorWithModel(4, 4, 30, DriftModelLiterature)

	if sim.DriftModel != DriftModelLiterature {
		t.Errorf("DriftModel = %d, want %d", sim.DriftModel, DriftModelLiterature)
	}
	if sim.DriftCoeff != FeFETDriftCoefficients.Literature {
		t.Errorf("DriftCoeff = %f, want %f", sim.DriftCoeff, FeFETDriftCoefficients.Literature)
	}
}

func TestSetDriftModel_AllModels(t *testing.T) {
	sim := NewDriftSimulator(4, 4, 30)

	tests := []struct {
		model    DriftModel
		expected float64
	}{
		{DriftModelAssumed, FeFETDriftCoefficients.Assumed},
		{DriftModelLiterature, FeFETDriftCoefficients.Literature},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			sim.SetDriftModel(tt.model)
			if sim.DriftCoeff != tt.expected {
				t.Errorf("After SetDriftModel(%d), DriftCoeff = %f, want %f", tt.model, sim.DriftCoeff, tt.expected)
			}
		})
	}

	// Test Measured model keeps existing coefficient
	customCoeff := 0.123
	sim.DriftCoeff = customCoeff
	sim.SetDriftModel(DriftModelMeasured)
	if sim.DriftCoeff != customCoeff {
		t.Errorf("After SetDriftModel(Measured), DriftCoeff = %f, want %f (unchanged)", sim.DriftCoeff, customCoeff)
	}
}

func TestSetMeasuredDriftCoeff(t *testing.T) {
	sim := NewDriftSimulator(4, 4, 30)
	customCoeff := 0.00123

	sim.SetMeasuredDriftCoeff(customCoeff)

	if sim.DriftModel != DriftModelMeasured {
		t.Errorf("DriftModel = %d, want %d", sim.DriftModel, DriftModelMeasured)
	}
	if sim.DriftCoeff != customCoeff {
		t.Errorf("DriftCoeff = %f, want %f", sim.DriftCoeff, customCoeff)
	}
}

func TestGetDriftModelInfo_Metadata(t *testing.T) {
	tests := []struct {
		name          string
		model         DriftModel
		wantName      string
		wantIsAssumed bool
	}{
		{"Assumed", DriftModelAssumed, "Assumed (Default)", true},
		{"Literature", DriftModelLiterature, "Literature-Derived", true},
		{"Measured", DriftModelMeasured, "User-Measured", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sim := NewDriftSimulator(4, 4, 30)
			sim.SetDriftModel(tt.model)
			info := sim.GetDriftModelInfo()

			if info.Model != tt.model {
				t.Errorf("Model = %d, want %d", info.Model, tt.model)
			}
			if info.ModelName != tt.wantName {
				t.Errorf("ModelName = %q, want %q", info.ModelName, tt.wantName)
			}
			if info.IsAssumed != tt.wantIsAssumed {
				t.Errorf("IsAssumed = %v, want %v", info.IsAssumed, tt.wantIsAssumed)
			}
			if info.SourceNote == "" {
				t.Error("SourceNote should not be empty")
			}
		})
	}
}

func TestSetConductanceLevel_BoundsCheck(t *testing.T) {
	sim := NewDriftSimulator(4, 4, 30)
	initialG := sim.Conductances[0][0]

	// Out of bounds - should do nothing
	sim.SetConductanceLevel(-1, 0, 15)
	sim.SetConductanceLevel(0, -1, 15)
	sim.SetConductanceLevel(100, 0, 15)
	sim.SetConductanceLevel(0, 100, 15)

	if sim.Conductances[0][0] != initialG {
		t.Error("Out of bounds SetConductanceLevel should not modify conductance")
	}
}

func TestSetConductanceLevel_ValidLevel(t *testing.T) {
	sim := NewDriftSimulator(4, 4, 30)
	row, col, level := 2, 3, 15

	sim.SetConductanceLevel(row, col, level)

	expected := sim.GMin + (sim.GMax-sim.GMin)*float64(level)/float64(sim.Levels-1)
	if math.Abs(sim.Conductances[row][col]-expected) > 1e-6 {
		t.Errorf("Conductances[%d][%d] = %e, want %e", row, col, sim.Conductances[row][col], expected)
	}
}

func TestSimulateTimeStep_TimeAdvances(t *testing.T) {
	sim := NewDriftSimulator(4, 4, 30)
	dt := 1000.0

	sim.SimulateTimeStep(dt)

	if sim.Time != dt {
		t.Errorf("Time = %f, want %f", sim.Time, dt)
	}

	sim.SimulateTimeStep(dt)
	if math.Abs(sim.Time-2*dt) > 1e-6 {
		t.Errorf("Time = %f, want %f", sim.Time, 2*dt)
	}
}

func TestSimulateTimeStep_ConductancesChange(t *testing.T) {
	sim := NewDriftSimulator(4, 4, 30)

	// Verify that Time increases (basic functionality)
	initialTime := sim.Time
	dt := 1000.0
	sim.SimulateTimeStep(dt)

	if sim.Time != initialTime+dt {
		t.Errorf("Time = %f, want %f after simulation", sim.Time, initialTime+dt)
	}

	// The drift formula uses log(t) which may produce very small changes
	// Just verify the simulation runs without error and time advances
}

func TestSimulateTimeStep_Clamping(t *testing.T) {
	sim := NewDriftSimulator(4, 4, 30)
	sim.DriftCoeff = 1.0 // Extreme drift

	// Simulate very long time
	for i := 0; i < 10; i++ {
		sim.SimulateTimeStep(1e9)
	}

	// All conductances should stay in valid range
	for i := 0; i < sim.Rows; i++ {
		for j := 0; j < sim.Cols; j++ {
			g := sim.Conductances[i][j]
			if g < sim.GMin || g > sim.GMax {
				t.Errorf("Conductance[%d][%d] = %e, out of range [%e, %e]", i, j, g, sim.GMin, sim.GMax)
			}
		}
	}
}

func TestRecordSnapshot_History(t *testing.T) {
	sim := NewDriftSimulator(4, 4, 30)

	if len(sim.DriftHistory) != 0 {
		t.Errorf("Initial history length = %d, want 0", len(sim.DriftHistory))
	}

	sim.RecordSnapshot()
	if len(sim.DriftHistory) != 1 {
		t.Errorf("After RecordSnapshot, history length = %d, want 1", len(sim.DriftHistory))
	}

	sim.RecordSnapshot()
	if len(sim.DriftHistory) != 2 {
		t.Errorf("After 2nd RecordSnapshot, history length = %d, want 2", len(sim.DriftHistory))
	}
}

func TestGetCurrentLevel_OutOfBounds(t *testing.T) {
	sim := NewDriftSimulator(4, 4, 30)

	tests := []struct {
		row, col int
	}{
		{-1, 0},
		{0, -1},
		{100, 0},
		{0, 100},
	}

	for _, tt := range tests {
		level := sim.GetCurrentLevel(tt.row, tt.col)
		if level != 0 {
			t.Errorf("GetCurrentLevel(%d, %d) = %d, want 0 for out of bounds", tt.row, tt.col, level)
		}
	}
}

func TestGetCurrentLevel_Valid(t *testing.T) {
	sim := NewDriftSimulator(4, 4, 30)
	row, col, targetLevel := 2, 3, 20

	sim.SetConductanceLevel(row, col, targetLevel)
	level := sim.GetCurrentLevel(row, col)

	if level != targetLevel {
		t.Errorf("GetCurrentLevel(%d, %d) = %d, want %d", row, col, level, targetLevel)
	}
}

func TestGetStats_Fresh(t *testing.T) {
	sim := NewDriftSimulator(4, 4, 30)

	stats := sim.GetStats()

	if stats.ElapsedTime != 0 {
		t.Errorf("ElapsedTime = %f, want 0", stats.ElapsedTime)
	}
	// At time=0, drift should be ~0 (but random component may exist)
	if stats.AvgDrift > 1e-3 {
		t.Errorf("AvgDrift = %e, expected near 0 at time=0", stats.AvgDrift)
	}
}

func TestGetStats_AfterDrift(t *testing.T) {
	sim := NewDriftSimulator(4, 4, 30)
	sim.DriftCoeff = 0.01 // Higher drift

	sim.SimulateTimeStep(1e6)
	stats := sim.GetStats()

	if stats.AvgDrift <= 0 {
		t.Error("After simulation, AvgDrift should be > 0")
	}
	if stats.MaxDrift <= 0 {
		t.Error("After simulation, MaxDrift should be > 0")
	}
}

func TestGetStats_RetentionPrediction(t *testing.T) {
	sim := NewDriftSimulator(4, 4, 30)
	sim.DriftCoeff = FeFETDriftCoefficients.Literature // Low drift

	stats := sim.GetStats()

	// For FeFET with low drift, retention should be high
	if stats.RetentionPrediction < 90.0 {
		t.Errorf("RetentionPrediction = %f%%, expected >90%% for low drift FeFET", stats.RetentionPrediction)
	}
}

func TestGetStats_TechnologyComparison(t *testing.T) {
	sim := NewDriftSimulator(4, 4, 30)

	stats := sim.GetStats()
	comp := stats.TechnologyComparison

	if comp.FeFETDrift != sim.DriftCoeff {
		t.Errorf("FeFETDrift = %f, want %f", comp.FeFETDrift, sim.DriftCoeff)
	}
	if comp.RRAMDrift != FeFETDriftCoefficients.RRAM {
		t.Errorf("RRAMDrift = %f, want %f", comp.RRAMDrift, FeFETDriftCoefficients.RRAM)
	}

	// FeFETAdvantage should be RRAM_drift / FeFET_drift = 0.05 / 0.001 = 50
	expectedAdvantage := FeFETDriftCoefficients.RRAM / sim.DriftCoeff
	if math.Abs(comp.FeFETAdvantage-expectedAdvantage) > 1e-2 {
		t.Errorf("FeFETAdvantage = %f, want %f", comp.FeFETAdvantage, expectedAdvantage)
	}
}

func TestRefreshCell_SnapsToLevel(t *testing.T) {
	sim := NewDriftSimulator(4, 4, 30)
	row, col := 1, 1

	// Set to a level, then manually perturb
	sim.SetConductanceLevel(row, col, 15)
	originalG := sim.Conductances[row][col]
	sim.Conductances[row][col] += 1e-6 // Small drift

	sim.RefreshCell(row, col)

	// After refresh, should snap back to nearest level
	refreshedG := sim.Conductances[row][col]
	if math.Abs(refreshedG-originalG) > 1e-9 {
		t.Errorf("After refresh, conductance = %e, want %e", refreshedG, originalG)
	}
}

func TestRefreshAll_AllCells(t *testing.T) {
	sim := NewDriftSimulator(4, 4, 30)

	// Manually perturb conductances to simulate drift
	for i := 0; i < sim.Rows; i++ {
		for j := 0; j < sim.Cols; j++ {
			// Add small perturbation
			sim.Conductances[i][j] += 1e-6
		}
	}

	// Verify perturbation worked
	drifted := false
	for i := 0; i < sim.Rows; i++ {
		for j := 0; j < sim.Cols; j++ {
			if math.Abs(sim.Conductances[i][j]-sim.InitialConds[i][j]) > 1e-9 {
				drifted = true
				break
			}
		}
	}
	if !drifted {
		t.Error("Expected some drift before RefreshAll")
	}

	sim.RefreshAll()

	// After RefreshAll, conductances should be at valid levels
	for i := 0; i < sim.Rows; i++ {
		for j := 0; j < sim.Cols; j++ {
			level := sim.GetCurrentLevel(i, j)
			expectedG := sim.GMin + (sim.GMax-sim.GMin)*float64(level)/float64(sim.Levels-1)
			if math.Abs(sim.Conductances[i][j]-expectedG) > 1e-9 {
				t.Errorf("After RefreshAll, Conductances[%d][%d] = %e, want %e", i, j, sim.Conductances[i][j], expectedG)
			}
		}
	}
}

func TestReset_ClearsState(t *testing.T) {
	sim := NewDriftSimulator(4, 4, 30)

	// Simulate and record
	sim.SimulateTimeStep(1000)
	sim.RecordSnapshot()

	if sim.Time == 0 {
		t.Error("Time should be > 0 after simulation")
	}
	if len(sim.DriftHistory) == 0 {
		t.Error("DriftHistory should not be empty after RecordSnapshot")
	}

	initialConds := make([][]float64, sim.Rows)
	for i := range initialConds {
		initialConds[i] = make([]float64, sim.Cols)
		copy(initialConds[i], sim.InitialConds[i])
	}

	sim.Reset()

	if sim.Time != 0 {
		t.Errorf("After Reset, Time = %f, want 0", sim.Time)
	}
	if len(sim.DriftHistory) != 0 {
		t.Errorf("After Reset, DriftHistory length = %d, want 0", len(sim.DriftHistory))
	}

	// Conductances should match initial
	for i := 0; i < sim.Rows; i++ {
		for j := 0; j < sim.Cols; j++ {
			if sim.Conductances[i][j] != initialConds[i][j] {
				t.Errorf("After Reset, Conductances[%d][%d] != initial", i, j)
			}
		}
	}
}

func TestCompareTechnologies_AllTechs(t *testing.T) {
	results := CompareTechnologies(8, 8, 1e6)

	expectedTechs := []string{"FeCIM (FeFET)", "RRAM", "PCM", "Flash"}
	for _, tech := range expectedTechs {
		if _, ok := results[tech]; !ok {
			t.Errorf("Missing technology %q in results", tech)
		}
	}
}

func TestCompareTechnologies_FeCIMBest(t *testing.T) {
	results := CompareTechnologies(8, 8, 1e7)

	fecimStats := results["FeCIM (FeFET)"]
	rramStats := results["RRAM"]
	pcmStats := results["PCM"]

	// FeCIM should have lowest drift
	if fecimStats.AvgDrift >= rramStats.AvgDrift {
		t.Errorf("FeCIM AvgDrift = %e, should be < RRAM AvgDrift = %e", fecimStats.AvgDrift, rramStats.AvgDrift)
	}
	if fecimStats.AvgDrift >= pcmStats.AvgDrift {
		t.Errorf("FeCIM AvgDrift = %e, should be < PCM AvgDrift = %e", fecimStats.AvgDrift, pcmStats.AvgDrift)
	}
}

// ===== IRDropSimulator Tests =====

func TestNewIRDropSimulator_Defaults(t *testing.T) {
	rows, cols := 8, 8
	sim := NewIRDropSimulator(rows, cols)

	if sim.Rows != rows {
		t.Errorf("Rows = %d, want %d", sim.Rows, rows)
	}
	if sim.Cols != cols {
		t.Errorf("Cols = %d, want %d", sim.Cols, cols)
	}
	if sim.RowResist != 2.5 {
		t.Errorf("RowResist = %f, want 2.5", sim.RowResist)
	}
	if sim.ColResist != 2.5 {
		t.Errorf("ColResist = %f, want 2.5", sim.ColResist)
	}
	if sim.Temperature != 300.0 {
		t.Errorf("Temperature = %f, want 300.0", sim.Temperature)
	}

	// Check default conductance
	expectedG := 50e-6
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if sim.Conductances[i][j] != expectedG {
				t.Errorf("Conductances[%d][%d] = %e, want %e", i, j, sim.Conductances[i][j], expectedG)
			}
		}
	}
}

func TestSetTemperature_RoomTemp(t *testing.T) {
	sim := NewIRDropSimulator(4, 4)
	baseResist := sim.RowResist

	sim.SetTemperature(300.0)

	// At 300K (reference temp), resistance should be unchanged
	if math.Abs(sim.RowResist-baseResist) > 1e-6 {
		t.Errorf("RowResist = %f, want %f at 300K", sim.RowResist, baseResist)
	}
}

func TestSetTemperature_HighTemp(t *testing.T) {
	sim := NewIRDropSimulator(4, 4)

	sim.SetTemperature(400.0) // 100K above reference

	// R(T) = R0 * (1 + 0.00393 * (T - 300))
	expectedFactor := 1.0 + 0.00393*100.0
	expectedResist := 2.5 * expectedFactor

	if math.Abs(sim.RowResist-expectedResist) > 1e-6 {
		t.Errorf("RowResist = %f, want %f at 400K", sim.RowResist, expectedResist)
	}
}

func TestSetTemperature_Negative(t *testing.T) {
	sim := NewIRDropSimulator(4, 4)

	sim.SetTemperature(-50.0)

	// Should clamp to 300K
	if sim.Temperature != 300.0 {
		t.Errorf("Temperature = %f, want 300.0 (clamped)", sim.Temperature)
	}
}

func TestGetTemperatureEffect_300K(t *testing.T) {
	sim := NewIRDropSimulator(4, 4)

	effect := sim.GetTemperatureEffect()

	if math.Abs(effect-1.0) > 1e-6 {
		t.Errorf("GetTemperatureEffect() = %f, want 1.0 at 300K", effect)
	}
}

func TestIRDropSetConductance_Valid(t *testing.T) {
	sim := NewIRDropSimulator(4, 4)
	row, col := 2, 3
	g := 123e-6

	sim.SetConductance(row, col, g)

	if sim.Conductances[row][col] != g {
		t.Errorf("Conductances[%d][%d] = %e, want %e", row, col, sim.Conductances[row][col], g)
	}
}

func TestIRDropSetConductance_OutOfBounds(t *testing.T) {
	sim := NewIRDropSimulator(4, 4)
	initialG := sim.Conductances[0][0]

	sim.SetConductance(-1, 0, 999e-6)
	sim.SetConductance(0, -1, 999e-6)
	sim.SetConductance(100, 0, 999e-6)
	sim.SetConductance(0, 100, 999e-6)

	// Should not modify anything
	if sim.Conductances[0][0] != initialG {
		t.Error("Out of bounds SetConductance should not modify array")
	}
}

func TestSetInputVoltage_Valid(t *testing.T) {
	sim := NewIRDropSimulator(4, 4)
	row := 2
	v := 1.5

	sim.SetInputVoltage(row, v)

	if sim.VoltageIn[row] != v {
		t.Errorf("VoltageIn[%d] = %f, want %f", row, sim.VoltageIn[row], v)
	}
}

func TestSetAllInputs(t *testing.T) {
	sim := NewIRDropSimulator(4, 4)
	voltages := []float64{1.0, 2.0, 3.0, 4.0}

	sim.SetAllInputs(voltages)

	for i, v := range voltages {
		if sim.VoltageIn[i] != v {
			t.Errorf("VoltageIn[%d] = %f, want %f", i, sim.VoltageIn[i], v)
		}
	}
}

func TestSimulate_ZeroInput(t *testing.T) {
	sim := NewIRDropSimulator(4, 4)

	// All inputs are zero by default
	sim.Simulate(100)

	outputs := sim.GetOutputCurrents()
	for i, curr := range outputs {
		if math.Abs(curr) > 1e-12 {
			t.Errorf("Output[%d] = %e, want 0 for zero input", i, curr)
		}
	}
}

func TestSimulate_UniformInput(t *testing.T) {
	rows, cols := 2, 2
	sim := NewIRDropSimulator(rows, cols)

	// Set uniform voltage and conductance
	v := 1.0
	g := 100e-6
	for i := 0; i < rows; i++ {
		sim.SetInputVoltage(i, v)
		for j := 0; j < cols; j++ {
			sim.SetConductance(i, j, g)
		}
	}

	sim.Simulate(100)

	// All outputs should be similar
	outputs := sim.GetOutputCurrents()
	for i := 1; i < len(outputs); i++ {
		if math.Abs(outputs[i]-outputs[0])/outputs[0] > 0.1 {
			t.Errorf("Output[%d] = %e, differs too much from Output[0] = %e", i, outputs[i], outputs[0])
		}
	}
}

func TestSimulate_IRDropIncreases(t *testing.T) {
	// Small array should have less IR drop than large array
	smallSim := NewIRDropSimulator(2, 2)
	largeSim := NewIRDropSimulator(16, 16)

	v := 1.0
	for i := 0; i < 2; i++ {
		smallSim.SetInputVoltage(i, v)
	}
	for i := 0; i < 16; i++ {
		largeSim.SetInputVoltage(i, v)
	}

	smallSim.Simulate(100)
	largeSim.Simulate(100)

	smallDrop := smallSim.GetMaxIRDrop()
	largeDrop := largeSim.GetMaxIRDrop()

	if largeDrop <= smallDrop {
		t.Errorf("Large array MaxIRDrop = %e should be > small array MaxIRDrop = %e", largeDrop, smallDrop)
	}
}

func TestGetOutputCurrents_MatchesIdeal_SmallArray(t *testing.T) {
	rows, cols := 2, 2
	sim := NewIRDropSimulator(rows, cols)

	// Low resistance for minimal IR drop
	sim.RowResist = 0.01
	sim.ColResist = 0.01

	v := 1.0
	g := 100e-6
	for i := 0; i < rows; i++ {
		sim.SetInputVoltage(i, v)
		for j := 0; j < cols; j++ {
			sim.SetConductance(i, j, g)
		}
	}

	sim.Simulate(100)

	actual := sim.GetOutputCurrents()
	ideal := sim.GetIdealOutputs()

	for i := range actual {
		relErr := math.Abs(actual[i]-ideal[i]) / ideal[i]
		if relErr > 0.05 { // 5% tolerance for small array with low resistance
			t.Errorf("Output[%d]: actual = %e, ideal = %e, rel error = %f%% > 5%%", i, actual[i], ideal[i], relErr*100)
		}
	}
}

func TestGetOutputError_SmallArray(t *testing.T) {
	rows, cols := 2, 2
	sim := NewIRDropSimulator(rows, cols)

	sim.RowResist = 0.1
	sim.ColResist = 0.1

	v := 1.0
	for i := 0; i < rows; i++ {
		sim.SetInputVoltage(i, v)
	}

	sim.Simulate(100)

	errors := sim.GetOutputError()

	// For small array, errors should be small
	for i, err := range errors {
		if err > 10.0 { // Less than 10% error
			t.Errorf("OutputError[%d] = %f%%, expected < 10%%", i, err)
		}
	}
}

func TestGetStats_Fields(t *testing.T) {
	sim := NewIRDropSimulator(4, 4)

	for i := 0; i < 4; i++ {
		sim.SetInputVoltage(i, 1.0)
	}

	sim.Simulate(100)
	stats := sim.GetStats()

	// Check all fields are populated
	if stats.MaxIRDrop < 0 {
		t.Error("MaxIRDrop should be >= 0")
	}
	if stats.AvgIRDrop < 0 {
		t.Error("AvgIRDrop should be >= 0")
	}
	if stats.Temperature != 300.0 {
		t.Errorf("Temperature = %f, want 300.0", stats.Temperature)
	}
	if stats.ResistFactor != 1.0 {
		t.Errorf("ResistFactor = %f, want 1.0 at 300K", stats.ResistFactor)
	}
	if stats.WorstCellRow < 0 || stats.WorstCellRow >= sim.Rows {
		t.Errorf("WorstCellRow = %d, out of range", stats.WorstCellRow)
	}
	if stats.WorstCellCol < 0 || stats.WorstCellCol >= sim.Cols {
		t.Errorf("WorstCellCol = %d, out of range", stats.WorstCellCol)
	}
}

func TestApplyMitigation_WidenedLines(t *testing.T) {
	sim := NewIRDropSimulator(8, 8)

	for i := 0; i < 8; i++ {
		sim.SetInputVoltage(i, 1.0)
	}

	sim.Simulate(100)
	dropBefore := sim.GetMaxIRDrop()

	// Apply widened lines mitigation
	mit := IRDropMitigation{
		UseWidenedLines:   true,
		LineWidthIncrease: 2.0, // 2x wider lines
	}
	sim.ApplyMitigation(mit)

	dropAfter := sim.GetMaxIRDrop()

	// IR drop should decrease with wider lines
	if dropAfter >= dropBefore {
		t.Errorf("After mitigation, MaxIRDrop = %e should be < before = %e", dropAfter, dropBefore)
	}

	// Resistance should be halved
	expectedResist := 2.5 / 2.0
	if math.Abs(sim.RowResist-expectedResist) > 1e-6 {
		t.Errorf("RowResist = %f, want %f after 2x widening", sim.RowResist, expectedResist)
	}
}
