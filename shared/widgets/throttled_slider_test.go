//go:build legacy_fyne

package widgets

import (
	"sync"
	"testing"
	"time"
)

func TestThrottledSliderPreviewDuringDragAndCommitOnRelease(t *testing.T) {
	var mu sync.Mutex
	previewValues := make([]float64, 0)
	commitValues := make([]float64, 0)

	ts := NewThrottledSlider(0, 1, 30*time.Millisecond,
		func(v float64) {
			mu.Lock()
			previewValues = append(previewValues, v)
			mu.Unlock()
		},
		func(v float64) {
			mu.Lock()
			commitValues = append(commitValues, v)
			mu.Unlock()
		},
	)

	for _, v := range []float64{0.1, 0.2, 0.3, 0.4, 0.5} {
		ts.Slider.OnChanged(v)
		time.Sleep(5 * time.Millisecond)
	}

	ts.Slider.OnChangeEnded(0.5)

	mu.Lock()
	defer mu.Unlock()

	if len(previewValues) == 0 {
		t.Fatal("expected at least one preview update during drag")
	}
	if len(previewValues) >= 5 {
		t.Fatalf("expected throttled preview (<5 calls), got %d", len(previewValues))
	}
	if len(commitValues) != 1 || commitValues[0] != 0.5 {
		t.Fatalf("expected exactly one commit with final value 0.5, got %#v", commitValues)
	}
}

func TestThrottledSliderAllowsPreviewAgainAfterRelease(t *testing.T) {
	previewCalls := 0
	ts := NewThrottledSlider(0, 100, 50*time.Millisecond,
		func(float64) { previewCalls++ },
		nil,
	)

	ts.Slider.OnChanged(10)
	ts.Slider.OnChangeEnded(10)
	ts.Slider.OnChanged(20)

	if previewCalls != 2 {
		t.Fatalf("expected preview to reset after release, got %d calls", previewCalls)
	}
}
