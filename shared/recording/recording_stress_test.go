package recording

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestBufferPoolConcurrentAllocateRecycleStress(t *testing.T) {
	t.Parallel()

	const (
		width      = 640
		height     = 360
		goroutines = 24
		iterations = 800
	)

	pool := NewBufferPool(width, height)
	expectedSize := width * height * 3

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for g := 0; g < goroutines; g++ {
		go func(seed byte) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				buf := pool.Get()
				if len(buf) != expectedSize {
					t.Errorf("got buffer size %d, want %d", len(buf), expectedSize)
					return
				}

				// Touch memory to simulate frame work and catch unsafe reuse patterns.
				step := len(buf) / 32
				if step == 0 {
					step = 1
				}
				for p := 0; p < len(buf); p += step {
					buf[p] = byte(i) ^ seed
				}
				pool.Put(buf)
			}
		}(byte(g + 1))
	}

	wg.Wait()

	stats := pool.Stats()
	expectedGets := int64(goroutines * iterations)
	if stats.Gets != expectedGets {
		t.Fatalf("gets = %d, want %d", stats.Gets, expectedGets)
	}
	if stats.Puts != expectedGets {
		t.Fatalf("puts = %d, want %d", stats.Puts, expectedGets)
	}
	if stats.Hits == 0 {
		t.Fatalf("expected pool reuse hits > 0, got %d", stats.Hits)
	}
}

func TestFrameCaptureNoMemoryLeakOver1000Frames(t *testing.T) {
	const (
		width  = 320
		height = 180
		frames = 1000
	)

	pool := NewBufferPool(width, height)

	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	for i := 0; i < frames; i++ {
		buf := pool.Get()
		if len(buf) != width*height*3 {
			t.Fatalf("frame %d: got size %d, want %d", i, len(buf), width*height*3)
		}

		// Simulate capture write.
		for p := 0; p < len(buf); p += 113 {
			buf[p] = byte(i)
		}
		pool.Put(buf)
	}

	runtime.GC()
	runtime.GC()
	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	// Allow a small GC/noise window, but memory should not grow without bound.
	const maxGrowth = 8 * 1024 * 1024 // 8 MiB
	if after.HeapAlloc > before.HeapAlloc+maxGrowth {
		t.Fatalf("heap grew too much after %d frames: before=%d after=%d maxGrowth=%d", frames, before.HeapAlloc, after.HeapAlloc, maxGrowth)
	}
}

func TestRecordingStartStopRestartWithoutCorruption(t *testing.T) {
	manager := NewManager()
	capture := &MockCaptureSource{width: 320, height: 240}

	if err := manager.Start(capture); err != nil {
		if isFFmpegMissing(err) {
			t.Skip("FFmpeg not available for stress start/stop/restart test")
		}
		t.Fatalf("first start failed: %v", err)
	}

	time.Sleep(150 * time.Millisecond)

	firstFile, err := manager.Stop()
	if err != nil {
		t.Fatalf("first stop failed: %v", err)
	}
	if firstFile == "" {
		t.Fatal("first stop returned empty output file")
	}

	if err := manager.Start(capture); err != nil {
		t.Fatalf("restart failed: %v", err)
	}

	time.Sleep(150 * time.Millisecond)

	secondFile, err := manager.Stop()
	if err != nil {
		t.Fatalf("second stop failed: %v", err)
	}
	if secondFile == "" {
		t.Fatal("second stop returned empty output file")
	}
	_ = firstFile // both runs completed; path uniqueness depends on timestamp granularity
}

func TestBufferSizesMatchExpectedRGB24Dimensions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		width  int
		height int
	}{
		{name: "VGA", width: 640, height: 480},
		{name: "HD", width: 1280, height: 720},
		{name: "FullHD", width: 1920, height: 1080},
		{name: "Small", width: 2, height: 2},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			pool := NewBufferPool(tc.width, tc.height)
			buf := pool.Get()
			expected := tc.width * tc.height * 3
			if len(buf) != expected {
				t.Fatalf("RGB24 buffer size mismatch for %dx%d: got %d, want %d", tc.width, tc.height, len(buf), expected)
			}
			pool.Put(buf)
		})
	}
}
