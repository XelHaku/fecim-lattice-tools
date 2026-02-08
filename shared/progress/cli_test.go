package progress

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestCLIProgress(t *testing.T) {
	var buf bytes.Buffer
	p := NewProgress("Test", 100)
	cli := NewCLIProgress(p, WithWriter(&buf), WithWidth(20))

	p.Start()
	cli.Start()

	p.Update(50)
	time.Sleep(150 * time.Millisecond) // Wait for render

	cli.Stop()

	output := buf.String()
	if len(output) == 0 {
		t.Error("CLI progress should produce output")
	}
}

func TestCLIProgressOptions(t *testing.T) {
	var buf bytes.Buffer
	p := NewProgress("Test", 100)

	cli := NewCLIProgress(p,
		WithWriter(&buf),
		WithWidth(30),
		WithShowRate(true),
		WithShowETA(true),
		WithPrefix("➤"),
	)

	if cli.width != 30 {
		t.Errorf("Width = %d, want 30", cli.width)
	}
	if cli.showRate != true {
		t.Error("ShowRate should be true")
	}
	if cli.showETA != true {
		t.Error("ShowETA should be true")
	}
	if cli.prefix != "➤" {
		t.Errorf("Prefix = %q, want %q", cli.prefix, "➤")
	}
}

func TestCLIProgressBar(t *testing.T) {
	p := NewProgress("Test", 100)
	cli := &CLIProgress{
		progress: p,
		width:    10,
		showRate: false,
		showETA:  false,
	}

	p.Start()
	p.Update(50)

	info := p.Info()
	bar := cli.buildBar(info)

	// Should have proper characters
	if !strings.Contains(bar, "█") {
		t.Error("Progress bar should contain filled blocks")
	}
	if !strings.Contains(bar, "░") {
		t.Error("Progress bar should contain empty blocks")
	}
	if !strings.HasPrefix(bar, "[") || !strings.HasSuffix(bar, "]") {
		t.Error("Progress bar should have brackets")
	}
}

func TestCLIProgressLine(t *testing.T) {
	p := NewProgress("Test Op", 100)
	cli := &CLIProgress{
		progress: p,
		width:    10,
		showRate: true,
		showETA:  true,
		prefix:   ">>",
	}

	p.Start()
	p.SetPhase("Phase 1")
	p.Update(25)

	info := p.Info()
	line := cli.buildLine(info)

	// Should contain prefix
	if !strings.HasPrefix(line, ">>") {
		t.Error("Line should start with prefix")
	}

	// Should contain phase
	if !strings.Contains(line, "Phase 1") {
		t.Error("Line should contain phase")
	}

	// Should contain percentage
	if !strings.Contains(line, "25") {
		t.Error("Line should contain percentage")
	}
}

func TestIterProgress(t *testing.T) {
	ip := NewIterProgress("Test", 10)
	ip.Start()

	for i := 0; i < 10; i++ {
		ip.Increment()
		ip.SetPhase("Step " + string(rune('0'+i)))
	}

	if ip.Progress().Current() != 10 {
		t.Errorf("Current = %d, want 10", ip.Progress().Current())
	}

	ip.Complete()

	if ip.Progress().State() != StateCompleted {
		t.Error("State should be Completed")
	}
}

func TestIterProgressCancellation(t *testing.T) {
	ip := NewIterProgress("Test", 100)
	ip.Start()

	ip.Cancel()

	if !ip.IsCancelled() {
		t.Error("IsCancelled should be true")
	}
}

func TestSimpleProgress(t *testing.T) {
	counter := 0
	err := SimpleProgress("Test", 10, func(p *Progress) error {
		for i := 0; i < 10; i++ {
			counter++
			p.Increment()
		}
		return nil
	})

	if err != nil {
		t.Errorf("SimpleProgress error: %v", err)
	}
	if counter != 10 {
		t.Errorf("Counter = %d, want 10", counter)
	}
}

func TestSpinnerProgress(t *testing.T) {
	called := false
	err := SpinnerProgress("Loading", func(p *Progress) error {
		called = true
		time.Sleep(50 * time.Millisecond)
		return nil
	})

	if err != nil {
		t.Errorf("SpinnerProgress error: %v", err)
	}
	if !called {
		t.Error("Work function should have been called")
	}
}

func TestMultiCLIProgress(t *testing.T) {
	var buf bytes.Buffer
	multi := NewMultiCLIProgress(&buf)

	p1 := NewProgress("Task 1", 100)
	p2 := NewProgress("Task 2", 50)

	multi.Add(p1, WithPrefix("[1]"))
	multi.Add(p2, WithPrefix("[2]"))

	p1.Start()
	p2.Start()

	multi.Start()

	p1.Update(50)
	p2.Update(25)

	time.Sleep(150 * time.Millisecond)

	multi.Stop()

	// Should have output
	if buf.Len() == 0 {
		t.Error("Multi CLI progress should produce output")
	}
}

func TestFormatRate(t *testing.T) {
	tests := []struct {
		rate     float64
		total    int64
		expected string
	}{
		{0.5, 100, "0.50 /sec"},
		{10, 100, "10.0 /sec"},
		{1500, 10000, "1.5k /sec"},
	}

	for _, tt := range tests {
		got := formatRate(tt.rate, tt.total)
		if got != tt.expected {
			t.Errorf("formatRate(%f, %d) = %q, want %q", tt.rate, tt.total, got, tt.expected)
		}
	}
}
