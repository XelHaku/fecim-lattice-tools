package validation

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	crossbar "fecim-lattice-tools/module2-crossbar/pkg/crossbar"
)

func TestResourceCleanup_NewArrayDoesNotLeakGoroutines(t *testing.T) {
	baseline := runtime.NumGoroutine()

	for i := 0; i < 10; i++ {
		arr, err := crossbar.NewArray(&crossbar.Config{
			Rows:       16,
			Cols:       16,
			NoiseLevel: 0.0,
			ADCBits:    8,
			DACBits:    8,
			UseGPU:     false,
		})
		if err != nil {
			t.Fatalf("NewArray failed: %v", err)
		}
		arr.Destroy()
	}

	// Give any asynchronous cleanup a moment to settle.
	time.Sleep(100 * time.Millisecond)
	runtime.GC()
	time.Sleep(50 * time.Millisecond)

	after := runtime.NumGoroutine()
	if diff := after - baseline; diff > 2 {
		t.Fatalf("possible goroutine leak: baseline=%d after=%d diff=+%d", baseline, after, diff)
	}
}

func TestResourceCleanup_ExportOperationsCloseFileHandles(t *testing.T) {
	tmpDir := t.TempDir()
	arr, err := crossbar.NewArray(&crossbar.Config{
		Rows:       8,
		Cols:       8,
		NoiseLevel: 0.0,
		ADCBits:    8,
		DACBits:    8,
		UseGPU:     false,
	})
	if err != nil {
		t.Fatalf("NewArray failed: %v", err)
	}
	defer arr.Destroy()

	weightsPath := filepath.Join(tmpDir, "weights.csv")
	analysisPath := filepath.Join(tmpDir, "analysis.json")

	if err := arr.ExportWeightsCSV(weightsPath); err != nil {
		t.Fatalf("ExportWeightsCSV failed: %v", err)
	}
	if err := arr.ExportAnalysisJSON(analysisPath, nil); err != nil {
		t.Fatalf("ExportAnalysisJSON failed: %v", err)
	}

	openInTmp, err := countOpenFDsInPath(tmpDir)
	if err != nil {
		t.Fatalf("countOpenFDsInPath failed: %v", err)
	}
	if openInTmp != 0 {
		t.Fatalf("detected %d open file descriptor(s) under %s after exports", openInTmp, tmpDir)
	}

	if err := os.Remove(weightsPath); err != nil {
		t.Fatalf("remove weights file failed (possible open handle): %v", err)
	}
	if err := os.Remove(analysisPath); err != nil {
		t.Fatalf("remove analysis file failed (possible open handle): %v", err)
	}
}

func TestResourceCleanup_LargeArrayMVMHeapUnder100MB(t *testing.T) {
	arr, err := crossbar.NewArray(&crossbar.Config{
		Rows:       128,
		Cols:       128,
		NoiseLevel: 0.0,
		ADCBits:    8,
		DACBits:    8,
		UseGPU:     false,
	})
	if err != nil {
		t.Fatalf("NewArray failed: %v", err)
	}
	defer arr.Destroy()

	input := make([]float64, 128)
	for i := range input {
		input[i] = 0.5
	}

	for i := 0; i < 20; i++ {
		if _, err := arr.MVM(input); err != nil {
			t.Fatalf("MVM failed: %v", err)
		}
	}

	runtime.GC()
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	const heapLimit = 100 * 1024 * 1024 // 100 MB
	if ms.HeapAlloc > heapLimit {
		t.Fatalf("heap alloc too high after 128x128 MVM: got=%d bytes (%.2f MB), limit=%d bytes (100 MB)",
			ms.HeapAlloc, float64(ms.HeapAlloc)/(1024*1024), heapLimit)
	}
	if ms.HeapInuse > heapLimit {
		t.Fatalf("heap in-use too high after 128x128 MVM: got=%d bytes (%.2f MB), limit=%d bytes (100 MB)",
			ms.HeapInuse, float64(ms.HeapInuse)/(1024*1024), heapLimit)
	}
}

func TestResourceCleanup_GPUContextDestroyIsIdempotent(t *testing.T) {
	arr, err := crossbar.NewArray(&crossbar.Config{
		Rows:       8,
		Cols:       8,
		NoiseLevel: 0.0,
		ADCBits:    8,
		DACBits:    8,
		UseGPU:     true,
	})
	if err != nil {
		t.Fatalf("NewArray failed: %v", err)
	}

	// Trigger lazy init path (falls back to CPU safely if GPU unavailable).
	_, _ = arr.MVM(make([]float64, 8))

	for i := 0; i < 5; i++ {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("Destroy panicked on call %d: %v", i+1, r)
				}
			}()
			arr.Destroy()
		}()
	}
}

func countOpenFDsInPath(pathPrefix string) (int, error) {
	entries, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		return 0, fmt.Errorf("read /proc/self/fd: %w", err)
	}

	count := 0
	for _, entry := range entries {
		linkPath, err := os.Readlink(filepath.Join("/proc/self/fd", entry.Name()))
		if err != nil {
			// Raced with fd close/open; skip noisy descriptors.
			continue
		}
		if strings.HasPrefix(linkPath, pathPrefix) {
			count++
		}
	}
	return count, nil
}
