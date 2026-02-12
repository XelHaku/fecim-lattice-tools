package core

import (
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
