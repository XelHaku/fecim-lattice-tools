package gui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fecim-lattice-tools/module3-mnist/pkg/core"
)

func writeTestWeights(t *testing.T, dir, filename string) {
	t.Helper()
	wf := core.WeightsFile{
		Layer1Weights: [][]float64{{0.1, -0.1}, {0.05, -0.05}},
		Layer2Weights: [][]float64{{0.2, -0.2}},
		Biases1:       []float64{0, 0},
		Biases2:       []float64{0},
		L1Scale:       1,
		L1Offset:      0,
		L2Scale:       1,
		L2Offset:      0,
	}
	data, err := json.Marshal(wf)
	if err != nil {
		t.Fatalf("marshal weights: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, filename), data, 0o644); err != nil {
		t.Fatalf("write weights: %v", err)
	}
}

func TestTryLoadQATWeights_FallbackDefaultWarns(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestWeights(t, tmpDir, "pretrained_weights.json")

	nc := &NetworkController{
		network:             core.NewDualModeNetwork(2, 2, 1),
		dataDir:             tmpDir,
		weightsDir:          tmpDir,
		currentQATLevel:     30,
		warnedMissingLevels: map[int]bool{},
	}

	result, err := nc.TryLoadQATWeights(17)
	if err != nil {
		t.Fatalf("TryLoadQATWeights returned error: %v", err)
	}
	if result != QATFallbackDefaultFirstWarning {
		t.Fatalf("expected QATFallbackDefaultFirstWarning, got %v", result)
	}
}

func TestNetworkController_SetNumLevelsClampNotice(t *testing.T) {
	nc := NewNetworkController(4, 3, 2)
	var notices []string
	nc.SetOnNotice(func(message string) {
		notices = append(notices, message)
	})

	nc.SetNumLevels(core.MaxDemoLevels + 10)

	if len(notices) == 0 {
		t.Fatal("expected notice for clamped levels, got none")
	}
	if !strings.Contains(notices[0], "clamped") {
		t.Fatalf("expected clamp notice, got %q", notices[0])
	}
}
