package crossbar

import (
	"math"
	"runtime"
	"sync"
	"testing"
)

func newConcurrentStressArray(t *testing.T, rows, cols int) *Array {
	t.Helper()

	arr, err := NewArray(&Config{
		Rows:       rows,
		Cols:       cols,
		NoiseLevel: 0.0,
		ADCBits:    8,
		DACBits:    8,
		UseGPU:     false,
	})
	if err != nil {
		t.Fatalf("failed to create array: %v", err)
	}

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			w := float64((i+j)%DefaultQuantizationLevels) / float64(DefaultQuantizationLevels-1)
			if err := arr.ProgramWeight(i, j, w); err != nil {
				t.Fatalf("failed to program weight at (%d,%d): %v", i, j, err)
			}
		}
	}

	return arr
}

// (1) Run 100 concurrent MVM operations on a shared 16x16 array and verify no corruption.
func TestConcurrentStressSharedArray100MVM(t *testing.T) {
	const size = 16
	const ops = 100

	arr := newConcurrentStressArray(t, size, size)
	defer arr.Destroy()

	input := make([]float64, size)
	for i := range input {
		input[i] = float64(i+1) / float64(size)
	}

	reference, err := arr.MVM(input)
	if err != nil {
		t.Fatalf("reference MVM failed: %v", err)
	}

	errCh := make(chan error, ops)
	var wg sync.WaitGroup

	for i := 0; i < ops; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			out, err := arr.MVM(input)
			if err != nil {
				errCh <- err
				return
			}
			if len(out) != len(reference) {
				errCh <- err
				return
			}
			for k := range out {
				if out[k] != reference[k] {
					errCh <- &mismatchError{index: k, got: out[k], want: reference[k]}
					return
				}
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatalf("concurrent MVM verification failed: %v", err)
		}
	}
}

// (2) Run concurrent read+write operations and verify consistency.
// Test uses an external lock to model coordinated concurrent access with deterministic consistency checks.
func TestConcurrentStressReadWriteConsistency(t *testing.T) {
	const size = 16
	const writers = 8
	const readers = 8
	const iterations = 200

	arr := newConcurrentStressArray(t, size, size)
	defer arr.Destroy()

	shadow := make([][]float64, size)
	for i := range shadow {
		shadow[i] = make([]float64, size)
		for j := 0; j < size; j++ {
			shadow[i][j] = arr.cells[i][j].Conductance
		}
	}

	var mu sync.RWMutex
	var wg sync.WaitGroup
	errCh := make(chan error, writers+readers)

	for w := 0; w < writers; w++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()
			for step := 0; step < iterations; step++ {
				row := (writerID*7 + step*3) % size
				col := (writerID*11 + step*5) % size
				value := float64((writerID+step)%DefaultQuantizationLevels) / float64(DefaultQuantizationLevels-1)
				expected := QuantizeToLevels(value)

				mu.Lock()
				err := arr.ProgramWeight(row, col, value)
				if err == nil {
					shadow[row][col] = expected
				}
				mu.Unlock()

				if err != nil {
					errCh <- err
					return
				}
			}
		}(w)
	}

	for r := 0; r < readers; r++ {
		wg.Add(1)
		go func(readerID int) {
			defer wg.Done()
			input := make([]float64, size)
			for i := range input {
				input[i] = float64((readerID+i)%size+1) / float64(size)
			}

			for step := 0; step < iterations; step++ {
				mu.RLock()
				out, err := arr.MVM(input)
				mu.RUnlock()
				if err != nil {
					errCh <- err
					return
				}
				for _, v := range out {
					if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 1 {
						errCh <- &rangeError{value: v}
						return
					}
				}
			}
		}(r)
	}

	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatalf("concurrent read/write failed: %v", err)
		}
	}

	mu.RLock()
	matrix := arr.GetConductanceMatrix()
	mu.RUnlock()
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			if matrix[i][j] != shadow[i][j] {
				t.Fatalf("consistency mismatch at (%d,%d): got %.8f want %.8f", i, j, matrix[i][j], shadow[i][j])
			}
		}
	}
}

// (3) Run concurrent ISPP-like programming on different cells simultaneously.
// ISPP is modeled as multiple incremental ProgramWeight pulses per target cell.
func TestConcurrentStressISPPDifferentCells(t *testing.T) {
	const size = 16
	const cellsToProgram = 64
	const pulses = 20

	arr := newConcurrentStressArray(t, size, size)
	defer arr.Destroy()

	type target struct {
		row  int
		col  int
		want float64
	}
	targets := make([]target, 0, cellsToProgram)
	for idx := 0; idx < cellsToProgram; idx++ {
		row := idx / size
		col := idx % size
		want := float64((idx%DefaultQuantizationLevels)) / float64(DefaultQuantizationLevels-1)
		targets = append(targets, target{row: row, col: col, want: QuantizeToLevels(want)})
	}

	var wg sync.WaitGroup
	errCh := make(chan error, cellsToProgram)

	for _, tg := range targets {
		tg := tg
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := 1; p <= pulses; p++ {
				intermediate := tg.want * float64(p) / float64(pulses)
				if err := arr.ProgramWeight(tg.row, tg.col, intermediate); err != nil {
					errCh <- err
					return
				}
			}
		}()
	}

	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatalf("concurrent ISPP programming failed: %v", err)
		}
	}

	matrix := arr.GetConductanceMatrix()
	for _, tg := range targets {
		got := matrix[tg.row][tg.col]
		if got != tg.want {
			t.Fatalf("ISPP final mismatch at (%d,%d): got %.8f want %.8f", tg.row, tg.col, got, tg.want)
		}
	}
}

// (4) Verify memory usage stays bounded during concurrent access.
func TestConcurrentStressMemoryBounded(t *testing.T) {
	const size = 16
	const goroutines = 32
	const opsPerGoroutine = 1000
	const maxGrowthBytes = 64 << 20 // 64 MiB guardrail for this stress test.

	arr := newConcurrentStressArray(t, size, size)
	defer arr.Destroy()

	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	var wg sync.WaitGroup
	errCh := make(chan error, goroutines)

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			input := make([]float64, size)
			for i := range input {
				input[i] = float64((id+i)%size+1) / float64(size)
			}
			for op := 0; op < opsPerGoroutine; op++ {
				out, err := arr.MVM(input)
				if err != nil {
					errCh <- err
					return
				}
				if len(out) != size {
					errCh <- &lengthError{got: len(out), want: size}
					return
				}
			}
		}(g)
	}

	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatalf("concurrent memory stress failed: %v", err)
		}
	}

	runtime.GC()
	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	growth := int64(after.Alloc) - int64(before.Alloc)
	if growth > maxGrowthBytes {
		t.Fatalf("memory growth too high: +%d bytes (limit %d)", growth, maxGrowthBytes)
	}
}

type mismatchError struct {
	index    int
	got, want float64
}

func (e *mismatchError) Error() string {
	return "output mismatch"
}

type rangeError struct{ value float64 }

func (e *rangeError) Error() string { return "output out of range" }

type lengthError struct{ got, want int }

func (e *lengthError) Error() string { return "unexpected output length" }
