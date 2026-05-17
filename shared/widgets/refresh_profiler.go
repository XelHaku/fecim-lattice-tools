//go:build legacy_fyne

package widgets

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Refresher is the minimal Refresh-capable interface implemented by Fyne widgets.
type Refresher interface {
	Refresh()
}

// RefreshProfiler counts Refresh() calls grouped by a component key (component/tab).
type RefreshProfiler struct {
	mu     sync.Mutex
	counts map[string]uint64
}

// NewRefreshProfiler creates a new refresh profiler instance.
func NewRefreshProfiler() *RefreshProfiler {
	return &RefreshProfiler{counts: make(map[string]uint64)}
}

// Wrap returns a Refresher wrapper that increments the key counter before delegating Refresh().
func (p *RefreshProfiler) Wrap(key string, target Refresher) Refresher {
	return refreshWrapper{key: key, target: target, profiler: p}
}

// Count returns the current refresh count for key.
func (p *RefreshProfiler) Count(key string) uint64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.counts[key]
}

// Snapshot returns a copy of all counters.
func (p *RefreshProfiler) Snapshot() map[string]uint64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make(map[string]uint64, len(p.counts))
	for k, v := range p.counts {
		out[k] = v
	}
	return out
}

// Reset clears all counters.
func (p *RefreshProfiler) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.counts = make(map[string]uint64)
}

// HotspotReport emits a descending refresh hotspot report.
func (p *RefreshProfiler) HotspotReport(topN int) string {
	type pair struct {
		key   string
		count uint64
	}
	snap := p.Snapshot()
	rows := make([]pair, 0, len(snap))
	for k, v := range snap {
		rows = append(rows, pair{key: k, count: v})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].count == rows[j].count {
			return rows[i].key < rows[j].key
		}
		return rows[i].count > rows[j].count
	})
	if topN > 0 && len(rows) > topN {
		rows = rows[:topN]
	}
	if len(rows) == 0 {
		return "refresh hotspots: no samples"
	}
	var b strings.Builder
	b.WriteString("refresh hotspots:\n")
	for i, r := range rows {
		_, _ = fmt.Fprintf(&b, "%d) %s -> %d\n", i+1, r.key, r.count)
	}
	return strings.TrimRight(b.String(), "\n")
}

type refreshWrapper struct {
	key      string
	target   Refresher
	profiler *RefreshProfiler
}

func (w refreshWrapper) Refresh() {
	if w.profiler != nil {
		w.profiler.mu.Lock()
		w.profiler.counts[w.key]++
		w.profiler.mu.Unlock()
	}
	if w.target != nil {
		w.target.Refresh()
	}
}
