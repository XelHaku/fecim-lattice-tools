//go:build legacy_fyne

package widgets

import "time"

const (
	FrameWarnThreshold     = 16 * time.Millisecond
	FrameCriticalThreshold = 33 * time.Millisecond
)

type FrameSeverity string

const (
	FrameOK       FrameSeverity = "ok"
	FrameWarn     FrameSeverity = "warn"
	FrameCritical FrameSeverity = "critical"
)

type FrameSample struct {
	Duration time.Duration
	Severity FrameSeverity
}

type FrameWatchdog struct {
	now     func() time.Time
	start   time.Time
	samples []FrameSample
}

func NewFrameWatchdog() *FrameWatchdog {
	return &FrameWatchdog{now: time.Now}
}

func (w *FrameWatchdog) StartFrame() {
	w.start = w.now()
}

func (w *FrameWatchdog) EndFrame() FrameSample {
	d := w.now().Sub(w.start)
	s := FrameSample{Duration: d, Severity: classifyFrame(d)}
	w.samples = append(w.samples, s)
	return s
}

func classifyFrame(d time.Duration) FrameSeverity {
	if d > FrameCriticalThreshold {
		return FrameCritical
	}
	if d > FrameWarnThreshold {
		return FrameWarn
	}
	return FrameOK
}

func (w *FrameWatchdog) Samples() []FrameSample {
	out := make([]FrameSample, len(w.samples))
	copy(out, w.samples)
	return out
}
