package neural

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetNumLevels_ClampEmitsNotification(t *testing.T) {
	net := NewDualModeNetwork(4, 3, 2)
	var notices []string
	net.SetNotificationHandler(func(message string) {
		notices = append(notices, message)
	})

	net.SetNumLevels(MaxDemoLevels + 99)

	if len(notices) == 0 {
		t.Fatal("expected clamp notification, got none")
	}
	if !strings.Contains(notices[0], "clamped") {
		t.Fatalf("expected clamp message, got %q", notices[0])
	}
}

func TestForwardFP_GPUFallbackEmitsNotification(t *testing.T) {
	net := NewDualModeNetwork(128, 2, 1)
	net.useGPU = true // force GPU attempt

	var notices []string
	net.SetNotificationHandler(func(message string) {
		notices = append(notices, message)
	})

	input := make([]float64, 128)
	weights := [][]float64{make([]float64, 128)}
	bias := []float64{0}

	_ = net.forwardFP(input, weights, bias)

	if len(notices) == 0 {
		t.Fatal("expected GPU fallback notification, got none")
	}
	if !strings.Contains(notices[0], "Falling back to CPU") {
		t.Fatalf("expected fallback message, got %q", notices[0])
	}
}

func TestLoadWeightsForLevel_FallbackEmitsNotification(t *testing.T) {
	tmp := t.TempDir()
	wf := createValidTestWeights()
	wf.QuantLevels = 20
	path := filepath.Join(tmp, "pretrained_weights_20.json")
	if err := os.WriteFile(path, []byte(marshalWeightsFile(t, wf)), 0o644); err != nil {
		t.Fatalf("write weights: %v", err)
	}

	net := NewDualModeNetwork(784, 128, 10)
	var notices []string
	net.SetNotificationHandler(func(message string) {
		notices = append(notices, message)
	})

	if err := net.LoadWeightsForLevel(tmp, 17); err != nil {
		t.Fatalf("LoadWeightsForLevel failed: %v", err)
	}

	found := false
	for _, n := range notices {
		if strings.Contains(n, "requested 17") && strings.Contains(n, "nearest available 20") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected nearest-level fallback notification, got %v", notices)
	}
}

func marshalWeightsFile(t *testing.T, wf *WeightsFile) string {
	t.Helper()
	b, err := json.Marshal(wf)
	if err != nil {
		t.Fatalf("marshal weights: %v", err)
	}
	return string(b)
}
