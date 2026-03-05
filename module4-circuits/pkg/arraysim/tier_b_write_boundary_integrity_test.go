package arraysim

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"
)

type writeBoundaryArtifact struct {
	Version       string  `json:"version"`
	GeneratedUnix int64   `json:"generated_unix"`
	TestName      string  `json:"test_name"`
	MaxNodeDelta  float64 `json:"max_node_delta"`
	TopNodeDelta  float64 `json:"top_node_delta"`
	FarEndShift   float64 `json:"far_end_shift"`
	MiddleShift   float64 `json:"middle_shift"`
	Pass          bool    `json:"pass"`
}

func TestTierBWriteBoundaryIntegrity_TopInjectedAndSolverDerivedInternals(t *testing.T) {
	params := SolveParams{
		WLVoltages: []float64{1.20, 1.05, 0.90},
		BLVoltages: []float64{0.05, 0.10, 0.15, 0.20},
		Conductance: [][]float64{
			{120e-6, 80e-6, 100e-6, 140e-6},
			{90e-6, 110e-6, 70e-6, 130e-6},
			{85e-6, 95e-6, 125e-6, 105e-6},
		},
		SelectorMode: SelectorWrite,
		WriteMask: [][]bool{
			{true, false, true, false},
			{true, true, false, false},
			{false, true, true, false},
		},
		Wire: WireParams{RWordLine: 1.8, RBitLine: 2.2},
		Boundary: BoundaryParams{
			WLDriveResistance:       4.0,
			BLDriveResistance:       3.5,
			WLTerminationResistance: 9.0,
			WLTerminationVoltage:    0.30,
			BLTerminationResistance: 8.5,
			BLTerminationVoltage:    0.25,
		},
	}

	solverRes, err := NewTierBSolver().SolveDC(params)
	if err != nil {
		t.Fatalf("tier-b solve: %v", err)
	}
	refRes, err := referenceSolveDense(params)
	if err != nil {
		t.Fatalf("reference solve: %v", err)
	}

	maxNodeDelta := 0.0
	for r := range solverRes.WLNodes {
		for c := range solverRes.WLNodes[r] {
			maxNodeDelta = math.Max(maxNodeDelta, math.Abs(solverRes.WLNodes[r][c]-refRes.WLNodes[r][c]))
			maxNodeDelta = math.Max(maxNodeDelta, math.Abs(solverRes.BLNodes[r][c]-refRes.BLNodes[r][c]))
		}
	}
	if maxNodeDelta > 1e-8 {
		t.Fatalf("write-path node mismatch tier-b vs dense reference: max delta=%g", maxNodeDelta)
	}

	topNodeDelta := 0.0
	for r, drive := range params.WLVoltages {
		topNodeDelta = math.Max(topNodeDelta, math.Abs(solverRes.WLNodes[r][0]-drive))
	}
	if !(topNodeDelta > 1e-5) {
		t.Fatalf("expected finite source impedance to prevent hard WL top-node assignment, got max delta=%g", topNodeDelta)
	}

	art := writeBoundaryArtifact{
		Version:       "v1",
		GeneratedUnix: 0,
		TestName:      t.Name(),
		MaxNodeDelta:  maxNodeDelta,
		TopNodeDelta:  topNodeDelta,
		Pass:          true,
	}
	writeWriteBoundaryArtifact(t, art)
}

func TestTierBWriteBoundaryIntegrity_NoDirectInternalAssignmentPath(t *testing.T) {
	base := SolveParams{
		WLVoltages: []float64{1.10, 0.95, 0.80},
		BLVoltages: []float64{0.00, 0.05, 0.10, 0.15},
		Conductance: [][]float64{
			{100e-6, 90e-6, 80e-6, 70e-6},
			{95e-6, 85e-6, 75e-6, 65e-6},
			{88e-6, 78e-6, 68e-6, 58e-6},
		},
		SelectorMode: SelectorWrite,
		WriteMask: [][]bool{
			{true, true, true, true},
			{true, false, true, false},
			{false, true, false, true},
		},
		Wire: WireParams{RWordLine: 2.0, RBitLine: 2.4},
		Boundary: BoundaryParams{
			WLDriveResistance:       3.8,
			BLDriveResistance:       3.8,
			WLTerminationResistance: 7.5,
			WLTerminationVoltage:    0.20,
			BLTerminationResistance: 7.5,
			BLTerminationVoltage:    0.10,
		},
	}

	shifted := base
	shifted.Boundary.WLTerminationVoltage = 0.55

	baseRes, err := NewTierBSolver().SolveDC(base)
	if err != nil {
		t.Fatalf("base solve: %v", err)
	}
	shiftRes, err := NewTierBSolver().SolveDC(shifted)
	if err != nil {
		t.Fatalf("shifted solve: %v", err)
	}

	farEndShift := math.Abs(shiftRes.WLNodes[1][3] - baseRes.WLNodes[1][3])
	middleShift := math.Abs(shiftRes.WLNodes[1][1] - baseRes.WLNodes[1][1])

	if !(farEndShift > 1e-4) {
		t.Fatalf("expected far-end node to move when WL termination voltage changes, shift=%g", farEndShift)
	}
	if !(middleShift > 1e-5) {
		t.Fatalf("expected internal WL node to move via solver coupling, shift=%g", middleShift)
	}

	art := writeBoundaryArtifact{
		Version:       "v1",
		GeneratedUnix: 0,
		TestName:      t.Name(),
		FarEndShift:   farEndShift,
		MiddleShift:   middleShift,
		Pass:          true,
	}
	writeWriteBoundaryArtifact(t, art)
}

func writeWriteBoundaryArtifact(t *testing.T, art writeBoundaryArtifact) {
	t.Helper()
	base := os.Getenv("FECIM_M4_REGRESSION_JSON_DIR")
	if base == "" {
		base = filepath.Join("output", "regression", "module4")
	}
	path := filepath.Join(base, "write_boundary_integrity.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Logf("write artifact mkdir failed: %v", err)
		return
	}

	payload := map[string]any{}
	if b, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(b, &payload)
	}
	payload[art.TestName] = art

	out, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		t.Logf("write artifact marshal failed: %v", err)
		return
	}
	if err := os.WriteFile(path, out, 0o644); err != nil {
		t.Logf("write artifact write failed: %v", err)
		return
	}
	t.Logf("write boundary artifact written: %s", path)
}
