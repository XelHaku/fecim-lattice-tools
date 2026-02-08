package progress_test

import (
	"context"
	"fmt"
	"time"

	"fecim-lattice-tools/shared/progress"
)

// ExampleProgress demonstrates basic progress tracking for a simulation
func ExampleProgress() {
	// Create a progress tracker for a simulation with 1000 steps
	p := progress.NewProgress("Hysteresis Simulation", 1000)

	// Register callbacks (optional)
	p.OnComplete(func(p *progress.Progress) {
		fmt.Printf("Simulation completed in %v\n", p.Elapsed())
	})

	// Start the simulation
	p.Start()
	p.SetPhase("Initializing")

	// Simulate work
	for i := 0; i < 1000; i++ {
		// Check for cancellation
		select {
		case <-p.Context().Done():
			fmt.Println("Simulation cancelled")
			return
		default:
		}

		// Update progress with status
		if i%100 == 0 {
			p.UpdateWithStatus(int64(i), "Computing",
				fmt.Sprintf("Processing step %d of 1000", i))
		} else {
			p.Increment()
		}

		// Simulate some work
		time.Sleep(time.Microsecond)
	}

	p.Complete()
}

// ExampleProgress_withCancellation shows how to handle cancellation
func ExampleProgress_withCancellation() {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Create progress with external context
	p := progress.NewProgressWithContext(ctx, "Long Simulation", 10000)
	p.Start()

	for i := 0; i < 10000; i++ {
		// Check context
		select {
		case <-p.Context().Done():
			fmt.Println("Operation timed out or cancelled")
			return
		default:
		}

		p.Increment()
		time.Sleep(time.Millisecond)
	}

	p.Complete()
}

// ExampleCLIProgress demonstrates CLI progress bar usage
func ExampleCLIProgress() {
	p := progress.NewProgress("Data Export", 100)
	cli := progress.NewCLIProgress(p)

	p.Start()
	cli.Start()

	for i := 0; i < 100; i++ {
		p.UpdateWithStatus(int64(i), "Exporting",
			fmt.Sprintf("Writing row %d", i))
		time.Sleep(10 * time.Millisecond)
	}

	p.Complete()
	cli.Stop()
}

// ExampleSimpleProgress demonstrates the simple one-liner API
func ExampleSimpleProgress() {
	err := progress.SimpleProgress("Processing", 100, func(p *progress.Progress) error {
		for i := 0; i < 100; i++ {
			p.Increment()
			time.Sleep(time.Millisecond)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

// ExampleSpinnerProgress demonstrates indeterminate progress
func ExampleSpinnerProgress() {
	err := progress.SpinnerProgress("Loading configuration", func(p *progress.Progress) error {
		// Simulate loading with unknown duration
		time.Sleep(100 * time.Millisecond)
		return nil
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

// ExampleIterProgress demonstrates the iterator-style API
func ExampleIterProgress() {
	items := []string{"file1.dat", "file2.dat", "file3.dat"}

	ip := progress.NewIterProgress("Processing files", int64(len(items)))
	ip.Start()

	for i, file := range items {
		if ip.IsCancelled() {
			break
		}

		ip.SetPhase(fmt.Sprintf("Processing %s", file))
		ip.SetDetail(fmt.Sprintf("File %d of %d", i+1, len(items)))

		// Process file...
		time.Sleep(50 * time.Millisecond)

		ip.Increment()
	}

	ip.Complete()
}

// ExampleMultiCLIProgress demonstrates multiple concurrent progress bars
func ExampleMultiCLIProgress() {
	multi := progress.NewMultiCLIProgress(nil) // nil = os.Stderr

	p1 := progress.NewProgress("Download", 1000)
	p2 := progress.NewProgress("Processing", 500)

	multi.Add(p1, progress.WithPrefix("[DL]"))
	multi.Add(p2, progress.WithPrefix("[PR]"))

	p1.Start()
	p2.Start()
	multi.Start()

	// Simulate concurrent work
	go func() {
		for i := 0; i < 1000; i++ {
			p1.Increment()
			time.Sleep(time.Millisecond)
		}
		p1.Complete()
	}()

	go func() {
		for i := 0; i < 500; i++ {
			p2.Increment()
			time.Sleep(2 * time.Millisecond)
		}
		p2.Complete()
	}()

	// Wait for completion (in real code, use WaitGroup or channels)
	time.Sleep(1100 * time.Millisecond)
	multi.Stop()
}
