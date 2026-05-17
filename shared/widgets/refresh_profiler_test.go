//go:build legacy_fyne

package widgets

import (
	"strings"
	"sync"
	"testing"
)

type mockRefresher struct {
	mu    sync.Mutex
	calls int
}

func (m *mockRefresher) Refresh() {
	m.mu.Lock()
	m.calls++
	m.mu.Unlock()
}

func (m *mockRefresher) Calls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls
}

func TestRefreshProfilerWrapCountsAndDelegates(t *testing.T) {
	p := NewRefreshProfiler()
	m := &mockRefresher{}
	w := p.Wrap("tab:eda/component:plot", m)

	w.Refresh()
	w.Refresh()

	if got := p.Count("tab:eda/component:plot"); got != 2 {
		t.Fatalf("count mismatch: got %d want 2", got)
	}
	if got := m.Calls(); got != 2 {
		t.Fatalf("delegate refresh mismatch: got %d want 2", got)
	}
}

func TestRefreshProfilerHotspotReportSorted(t *testing.T) {
	p := NewRefreshProfiler()
	for i := 0; i < 5; i++ {
		p.Wrap("tab:mnist/component:canvas", nil).Refresh()
	}
	for i := 0; i < 2; i++ {
		p.Wrap("tab:eda/component:status", nil).Refresh()
	}

	report := p.HotspotReport(0)
	first := strings.Index(report, "tab:mnist/component:canvas -> 5")
	second := strings.Index(report, "tab:eda/component:status -> 2")
	if first < 0 || second < 0 || first > second {
		t.Fatalf("unexpected report order:\n%s", report)
	}
}
