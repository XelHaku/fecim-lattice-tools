//go:build ignore

package gui

import (
	"fmt"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	"fecim-lattice-tools/module4-circuits/pkg/arraysim"
	sharedphysics "fecim-lattice-tools/shared/physics"
	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

type paritySnapshot struct {
	EffectiveCellVoltage float64
	RowCurrentsUA        []float64
	RowLevels            []int
}

type parityWriteTrace struct {
	RequestedVoltage float64
	AppliedVoltage   float64
	EffectiveVoltage float64
	DACCode          int
	Trajectory       [3]int // [start, next, post-iterate]
	Result           ISPPResult
}

type parityStepArtifact struct {
	Material           string  `json:"material"`
	Scenario           string  `json:"scenario"`  // 0T1R / 1T1R
	StepType           string  `json:"step_type"` // READ / COMPUTE / WRITE_STEP
	Tolerance          float64 `json:"tolerance"`
	MaxObservedDelta   float64 `json:"max_observed_delta"`
	Pass               bool    `json:"pass"`
	FailingIndex       int     `json:"failing_index,omitempty"`
	DeltaKind          string  `json:"delta_kind"` // voltage|current_ua|level|write
	AdditionalMetadata any     `json:"additional_metadata,omitempty"`
}

type parityArtifact struct {
	Version       string               `json:"version"`
	Profile       string               `json:"profile"`
	GeneratedUnix int64                `json:"generated_unix"`
	Records       []parityStepArtifact `json:"records"`
}

func TestHeadlessPhysicsParity_GUIVsHeadless_ReadComputeWriteStep_MaterialAware(t *testing.T) {
	materials := []struct {
		name string
		mat  *sharedphysics.HZOMaterial
	}{
		{name: "fecim_hzo", mat: sharedphysics.FeCIMMaterial()},
		{name: "literature_superlattice", mat: sharedphysics.LiteratureSuperlattice()},
	}
	if sel := strings.TrimSpace(os.Getenv("FECIM_PARITY_MATERIALS")); sel != "" {
		want := map[string]bool{}
		for _, part := range strings.Split(sel, ",") {
			want[strings.TrimSpace(part)] = true
		}
		filtered := materials[:0]
		for _, m := range materials {
			if want[m.name] {
				filtered = append(filtered, m)
			}
		}
		materials = filtered
	}

	artifact := parityArtifact{
		Version:       "v1",
		Profile:       os.Getenv("FECIM_MATERIAL_PROFILE"),
		GeneratedUnix: time.Now().Unix(),
		Records:       nil,
	}

	architectures := []struct {
		name      string
		mode      string
		isPassive bool
		coupling  arraysim.CouplingMode
	}{
		{name: "0T1R", mode: sharedwidgets.Architecture0T1R, isPassive: true, coupling: arraysim.CouplingTierA},
		{name: "1T1R", mode: sharedwidgets.Architecture1T1R, isPassive: false, coupling: arraysim.CouplingTierA},
	}

	defer writeParityArtifact(t, &artifact)

	for _, mc := range materials {
		mc := mc
		for _, ac := range architectures {
			ac := ac
			t.Run(fmt.Sprintf("%s_%s", mc.name, ac.name), func(t *testing.T) {
				// RG-VAL-03: scripts check for presence of required material verdicts in logs.
				t.Logf("VERDICT material=%s", mc.name)

				embedded, app, win := setupUnifiedTestApp(t)
				defer app.Quit()
				defer win.Close()
				defer embedded.Stop()

				ca := embedded.CircuitsApp
				if ca == nil || ca.deviceState == nil {
					t.Fatal("expected circuits app with initialized device state")
				}
				if mc.mat == nil {
					t.Fatal("test setup: material must be non-nil")
				}

				ca.architecture = ac.mode
				ca.deviceState.SetPassiveMode(ac.isPassive)
				ca.deviceState.SetCouplingMode(ac.coupling)
				ca.deviceState.SetMaterial(mc.mat)

				rows, cols := ca.arrayRows, ca.arrayCols
				row, col := rows/2, cols/2
				ca.deviceState.SetSelectedCell(row, col)

				weights := makeParityWeights(rows, cols, ca.quantLevels)
				setParityWeights(ca, weights)
				inputs := makeParityInputs(cols)
				setParityInputs(ca, inputs)

				// READ parity: GUI action path vs headless harness path.
				guiRead := runGUIReadParity(ca, row, col)
				headlessReadDS := newHeadlessParityState(ca, mc.mat, ac.isPassive, ac.coupling)
				headlessRead := runHeadlessReadParity(headlessReadDS, cloneInt2D(weights), ca.quantLevels, row, col)
				artifact.Records = append(artifact.Records, compareSnapshotAndRecord(t, &artifact, mc.name, ac.name, "READ", guiRead, headlessRead, 1e-6, 1e-3)...)

				// COMPUTE parity: GUI action path vs headless harness path.
				guiCompute := runGUIComputeParity(ca, row, col, inputs)
				headlessComputeDS := newHeadlessParityState(ca, mc.mat, ac.isPassive, ac.coupling)
				headlessCompute := runHeadlessComputeParity(headlessComputeDS, cloneInt2D(weights), ca.quantLevels, row, col, inputs)
				artifact.Records = append(artifact.Records, compareSnapshotAndRecord(t, &artifact, mc.name, ac.name, "COMPUTE", guiCompute, headlessCompute, 1e-6, 1e-3)...)

				// WRITE one-step trajectory parity: same ISPP pulse request and resulting next-level update.
				current := weights[row][col]
				target := current + 4
				if target >= ca.quantLevels {
					target = ca.quantLevels - 1
				}
				if target == current {
					target = current - 2
					if target < 0 {
						target = 0
					}
				}

				guiWrite := runGUIWriteStepParity(ca, row, col, target)
				headlessWriteDS := newHeadlessParityState(ca, mc.mat, ac.isPassive, ac.coupling)
				headlessWrite := runHeadlessWriteStepParity(headlessWriteDS, cloneInt2D(weights), ca.quantLevels, row, col, target)
				artifact.Records = append(artifact.Records, compareWriteAndRecord(t, mc.name, ac.name, guiWrite, headlessWrite, 1e-6))
			})
		}
	}
}

func makeParityWeights(rows, cols, levels int) [][]int {
	w := make([][]int, rows)
	for r := 0; r < rows; r++ {
		w[r] = make([]int, cols)
		for c := 0; c < cols; c++ {
			w[r][c] = (r*7 + c*11 + 3) % levels
		}
	}
	return w
}

func makeParityInputs(cols int) []int {
	out := make([]int, cols)
	for c := 0; c < cols; c++ {
		out[c] = (c*47 + 29) % 256
	}
	return out
}

func cloneInt2D(src [][]int) [][]int {
	out := make([][]int, len(src))
	for i := range src {
		out[i] = append([]int(nil), src[i]...)
	}
	return out
}

func setParityWeights(ca *CircuitsApp, weights [][]int) {
	ca.mu.Lock()
	ca.arrayWeights = cloneInt2D(weights)
	ca.mu.Unlock()
}

func setParityInputs(ca *CircuitsApp, inputs []int) {
	ca.mu.Lock()
	ca.inputVector = append([]int(nil), inputs...)
	ca.mu.Unlock()
}

func newHeadlessParityState(ca *CircuitsApp, mat *sharedphysics.HZOMaterial, passive bool, coupling arraysim.CouplingMode) *DeviceState {
	ds := NewDeviceState(ca.arrayRows, ca.arrayCols, ca.tia, ca.adc)
	src := ca.deviceState

	// Clone GUI device-state physics/peripheral configuration so headless path
	// solves with the same geometry, selector, and conversion settings.
	src.mu.RLock()
	geom := src.cellGeometry
	wire := src.wireParams
	selectorEnabled := src.selectorEnabled
	selectorRon := src.selectorRon
	selectorLeakage := src.selectorLeakageConductance
	enableDACNonlinearity := src.enableDACNonlinearity
	peripheralTemp := src.peripheralTemperature
	processCorner := src.processCorner
	isppEngine := src.isppEngine
	dacBits := 0
	adcBits := 0
	if src.dac != nil {
		dacBits = src.dac.Bits
	}
	if src.adc != nil {
		adcBits = src.adc.Bits
	}
	src.mu.RUnlock()

	if dacBits > 0 {
		ds.SetDACBits(dacBits)
	}
	if adcBits > 0 {
		ds.SetADCBits(adcBits)
	}
	ds.SetCellGeometry(geom)
	ds.SetWireParams(wire)
	ds.SetSelectorSeriesParams(selectorEnabled, selectorRon, selectorLeakage)
	ds.SetDACNonlinearity(enableDACNonlinearity)
	ds.SetPeripheralPVT(peripheralTemp, processCorner)
	ds.SetISPPEngine(isppEngine)

	ds.SetMaterial(mat)
	ds.SetPassiveMode(passive)
	ds.SetCouplingMode(coupling)
	return ds
}

func captureParitySnapshot(ds *DeviceState, row, col, rows int) paritySnapshot {
	s := paritySnapshot{
		EffectiveCellVoltage: ds.GetEffectiveCellVoltage(row, col),
		RowCurrentsUA:        make([]float64, rows),
		RowLevels:            make([]int, rows),
	}
	for r := 0; r < rows; r++ {
		s.RowCurrentsUA[r] = ds.GetRowCurrent(r)
		s.RowLevels[r] = ds.GetRowLevel(r)
	}
	return s
}

func runGUIReadParity(ca *CircuitsApp, row, col int) paritySnapshot {
	ca.deviceState.SetSelectedCell(row, col)
	ca.setOperationMode(OpModeRead)
	// Avoid throttled deferred recompute in onUnifiedRead so parity captures settled state.
	time.Sleep(uiRefreshMinInterval + 5*time.Millisecond)
	ca.onUnifiedRead()
	return captureParitySnapshot(ca.deviceState, row, col, ca.arrayRows)
}

func runHeadlessReadParity(ds *DeviceState, weights [][]int, quantLevels, row, col int) paritySnapshot {
	ds.SetSelectedCell(row, col)
	ds.SetOperationMode(OpModeRead)
	if !ds.IsPassiveMode() {
		ds.SetWLSingle(row)
	}
	readVoltage := ds.GetReadRange().Max * 0.4
	if readVoltage < 0.1 {
		readVoltage = 0.2
	}
	// Match GUI mode-switch recompute path.
	ds.SetAllDACVoltages(0)
	ds.SetDACVoltage(col, readVoltage)
	ds.SetDACRangeMode(DACRangeRead)
	ds.Compute(weights, quantLevels)

	// Match onUnifiedRead action path.
	ds.SetAllDACVoltages(0)
	ds.SetDACVoltage(col, readVoltage)
	ds.SetDACRangeMode(DACRangeRead)
	ds.Compute(weights, quantLevels)
	return captureParitySnapshot(ds, row, col, len(weights))
}

func runGUIComputeParity(ca *CircuitsApp, row, col int, inputs []int) paritySnapshot {
	setParityInputs(ca, inputs)
	ca.deviceState.SetSelectedCell(row, col)
	ca.setOperationMode(OpModeCompute)
	// Avoid throttled deferred recompute in onUnifiedCompute so parity captures settled state.
	time.Sleep(uiRefreshMinInterval + 5*time.Millisecond)
	ca.onUnifiedCompute()
	return captureParitySnapshot(ca.deviceState, row, col, ca.arrayRows)
}

func runHeadlessComputeParity(ds *DeviceState, weights [][]int, quantLevels, row, col int, inputs []int) paritySnapshot {
	ds.SetSelectedCell(row, col)
	ds.SetOperationMode(OpModeCompute)
	if !ds.IsPassiveMode() {
		ds.SetWLAll()
	}
	params := make([]float64, len(inputs))
	for i, v := range inputs {
		params[i] = float64(v)
	}
	// Match GUI mode-switch recompute path.
	ds.SetDACRangeMode(DACRangeRead)
	ds.SetDACPreset(DACInputVector, params...)
	ds.Compute(weights, quantLevels)

	// Match onUnifiedCompute action path.
	ds.SetDACRangeMode(DACRangeRead)
	ds.SetDACPreset(DACInputVector, params...)
	ds.Compute(weights, quantLevels)
	return captureParitySnapshot(ds, row, col, len(weights))
}

func runGUIWriteStepParity(ca *CircuitsApp, row, col, targetLevel int) parityWriteTrace {
	ca.deviceState.SetSelectedCell(row, col)
	ca.setOperationMode(OpModeWrite)

	ca.mu.RLock()
	current := ca.arrayWeights[row][col]
	ca.mu.RUnlock()

	ca.deviceState.StartISPP(row, col, targetLevel, current)
	status := ca.deviceState.GetISPPStatus()

	applied, code := ca.applyWriteVoltages(row, col, status.Voltage)
	ca.recomputeAndRefreshNow()

	effective := ca.deviceState.GetEffectiveCellVoltage(row, col)
	if effective == 0 {
		effective = status.Voltage
	}
	next := ca.deviceState.programLevelFromCoupledVoltage(
		current, effective, float64(PhaseWriteDurationNs)*1e-9, ca.quantLevels,
	)
	next = clampISPPNextLevel(next, current, targetLevel, status.Direction)
	result := ca.deviceState.ISPPIterate(next)
	post := ca.deviceState.GetISPPStatus().CurrentLevel

	ca.deviceState.CancelISPP()
	ca.deviceState.ResetWriteVoltages()

	return parityWriteTrace{
		RequestedVoltage: status.Voltage,
		AppliedVoltage:   applied,
		EffectiveVoltage: effective,
		DACCode:          code,
		Trajectory:       [3]int{current, next, post},
		Result:           result,
	}
}

func runHeadlessWriteStepParity(ds *DeviceState, weights [][]int, quantLevels, row, col, targetLevel int) parityWriteTrace {
	ds.SetSelectedCell(row, col)
	ds.SetOperationMode(OpModeWrite)
	if !ds.IsPassiveMode() {
		ds.SetWLSingle(row)
	}
	ds.SetDACRangeMode(DACRangeWrite)
	ds.SetAllDACVoltages(0)
	// Match GUI mode-switch recompute path before explicit program pulse.
	ds.Compute(weights, quantLevels)

	current := weights[row][col]
	ds.StartISPP(row, col, targetLevel, current)
	status := ds.GetISPPStatus()

	applied, code := applyWriteVoltagesHeadless(ds, row, col, status.Voltage)
	ds.Compute(weights, quantLevels)

	effective := ds.GetEffectiveCellVoltage(row, col)
	if effective == 0 {
		effective = status.Voltage
	}
	next := ds.programLevelFromCoupledVoltage(
		current, effective, float64(PhaseWriteDurationNs)*1e-9, quantLevels,
	)
	next = clampISPPNextLevel(next, current, targetLevel, status.Direction)
	result := ds.ISPPIterate(next)
	post := ds.GetISPPStatus().CurrentLevel

	ds.CancelISPP()
	ds.ResetWriteVoltages()

	return parityWriteTrace{
		RequestedVoltage: status.Voltage,
		AppliedVoltage:   applied,
		EffectiveVoltage: effective,
		DACCode:          code,
		Trajectory:       [3]int{current, next, post},
		Result:           result,
	}
}

func applyWriteVoltagesHeadless(ds *DeviceState, row, col int, targetVoltage float64) (float64, int) {
	applied, dacCode := ds.DACWriteVoltage(targetVoltage)
	if ds.IsPassiveMode() {
		ds.ApplyHalfSelectWrite(row, col, applied)
	} else {
		ds.SetAllDACVoltages(0)
		ds.SetDACVoltage(col, applied)
	}
	ds.SetDACRangeMode(DACRangeWrite)
	return applied, dacCode
}

func clampISPPNextLevel(next, current, target int, direction HysteresisDirection) int {
	switch direction {
	case DirectionAscending:
		if next < current {
			next = current
		}
		if next > target {
			next = target
		}
	case DirectionDescending:
		if next > current {
			next = current
		}
		if next < target {
			next = target
		}
	}
	return next
}

func assertParitySnapshotEqual(t *testing.T, label string, gui, headless paritySnapshot, voltageTol, currentTol float64) {
	t.Helper()
	assertParityFloatClose(t, label+" effective V", gui.EffectiveCellVoltage, headless.EffectiveCellVoltage, voltageTol)
	if len(gui.RowCurrentsUA) != len(headless.RowCurrentsUA) {
		t.Fatalf("%s row currents length mismatch: gui=%d headless=%d", label, len(gui.RowCurrentsUA), len(headless.RowCurrentsUA))
	}
	if len(gui.RowLevels) != len(headless.RowLevels) {
		t.Fatalf("%s row levels length mismatch: gui=%d headless=%d", label, len(gui.RowLevels), len(headless.RowLevels))
	}
	for i := range gui.RowCurrentsUA {
		assertParityFloatClose(t, fmt.Sprintf("%s row[%d] current", label, i), gui.RowCurrentsUA[i], headless.RowCurrentsUA[i], currentTol)
		if gui.RowLevels[i] != headless.RowLevels[i] {
			t.Fatalf("%s row[%d] adc level mismatch: gui=%d headless=%d", label, i, gui.RowLevels[i], headless.RowLevels[i])
		}
	}
}

func assertParityWriteTraceEqual(t *testing.T, gui, headless parityWriteTrace, tol float64) {
	t.Helper()
	assertParityFloatClose(t, "write requested voltage", gui.RequestedVoltage, headless.RequestedVoltage, tol)
	assertParityFloatClose(t, "write applied voltage", gui.AppliedVoltage, headless.AppliedVoltage, tol)
	assertParityFloatClose(t, "write effective voltage", gui.EffectiveVoltage, headless.EffectiveVoltage, tol)
	if gui.DACCode != headless.DACCode {
		t.Fatalf("write DAC code mismatch: gui=%d headless=%d", gui.DACCode, headless.DACCode)
	}
	if gui.Result != headless.Result {
		t.Fatalf("write ISPP result mismatch: gui=%v headless=%v", gui.Result, headless.Result)
	}
	if gui.Trajectory != headless.Trajectory {
		t.Fatalf("write trajectory mismatch: gui=%v headless=%v", gui.Trajectory, headless.Trajectory)
	}
}

func assertParityFloatClose(t *testing.T, label string, got, want, tol float64) {
	t.Helper()
	if math.Abs(got-want) > tol {
		t.Fatalf("%s mismatch: got=%.9f want=%.9f tol=%.9f", label, got, want, tol)
	}
}
