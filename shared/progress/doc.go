// Package progress provides progress tracking for long-running operations in FeCIM tools.
//
// Features:
//   - Progress bars with percentage and ETA calculation
//   - Cancellation support via context
//   - Pause/resume capability
//   - Detailed status messages (phase + detail)
//   - Callbacks for progress updates, completion, and errors
//   - CLI progress bars for terminal output
//   - Fyne GUI widgets for graphical display
//
// Basic Usage:
//
//	p := progress.NewProgress("Simulation", 1000)
//	p.Start()
//	for i := 0; i < 1000; i++ {
//	    p.Increment()
//	}
//	p.Complete()
//
// With Status Messages:
//
//	p.UpdateWithStatus(500, "Processing", "Batch 5 of 10")
//
// With Cancellation:
//
//	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
//	p := progress.NewProgressWithContext(ctx, "Export", 1000)
//	p.Start()
//	for i := 0; i < 1000; i++ {
//	    select {
//	    case <-p.Context().Done():
//	        return // Cancelled
//	    default:
//	        p.Increment()
//	    }
//	}
//	p.Complete()
//
// CLI Progress Bar:
//
//	p := progress.NewProgress("Export", 1000)
//	cli := progress.NewCLIProgress(p)
//	p.Start()
//	cli.Start()
//	// ... work ...
//	cli.Stop()
//
// Simple One-Liner:
//
//	err := progress.SimpleProgress("Processing", 100, func(p *progress.Progress) error {
//	    for i := 0; i < 100; i++ {
//	        p.Increment()
//	    }
//	    return nil
//	})
//
// Fyne Widget:
//
//	p := progress.NewProgress("Simulation", 1000)
//	widget := progress.NewProgressWidget(p, progress.WithCancel(true))
//	// Add widget to your Fyne container
//
// ETA Calculation:
//
// The progress tracker maintains a ring buffer of recent progress samples
// and uses linear regression to estimate the completion time. The ETA
// becomes more accurate as more samples are collected.
//
// Thread Safety:
//
// All Progress methods are safe for concurrent use. The widget uses
// fyne.Do() to ensure UI updates happen on the main thread.
package progress
