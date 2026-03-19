package layout

import (
	"fmt"
	"testing"

	"fecim-lattice-tools/shared/mathutil"
)

func TestPlaceForceDirected_WithinDieBoundsAndNoOverlap(t *testing.T) {
	const dieW, dieH = 2200, 1800
	macros := []MacroBlock{
		{Name: "XBAR0", Width: 280, Height: 240},
		{Name: "XBAR1", Width: 260, Height: 260},
		{Name: "DAC", Width: 180, Height: 160},
		{Name: "ADC", Width: 180, Height: 160},
		{Name: "CTRL", Width: 220, Height: 180},
		{Name: "BUF", Width: 160, Height: 140},
	}
	nets := []Net{
		{Name: "n0", Nodes: []string{"XBAR0", "DAC", "CTRL"}},
		{Name: "n1", Nodes: []string{"XBAR0", "ADC"}},
		{Name: "n2", Nodes: []string{"XBAR0", "XBAR1"}},
		{Name: "n3", Nodes: []string{"XBAR1", "BUF", "CTRL"}},
	}

	placed := PlaceForceDirected(macros, nets, dieW, dieH, 20, 20, 80)
	if len(placed) != len(macros) {
		t.Fatalf("expected %d placed macros, got %d", len(macros), len(placed))
	}

	for _, m := range macros {
		p, ok := placed[m.Name]
		if !ok {
			t.Fatalf("missing placement for %s", m.Name)
		}
		if p.X < 0 || p.Y < 0 {
			t.Fatalf("negative placement for %s: %+v", m.Name, p)
		}
		if p.X+m.Width > dieW || p.Y+m.Height > dieH {
			t.Fatalf("placement out of die bounds for %s: %+v (die=%dx%d)", m.Name, p, dieW, dieH)
		}
	}

	for i := 0; i < len(macros); i++ {
		for j := i + 1; j < len(macros); j++ {
			ma, mb := macros[i], macros[j]
			pa, pb := placed[ma.Name], placed[mb.Name]
			if rectOverlap(pa.X, pa.Y, ma.Width, ma.Height, pb.X, pb.Y, mb.Width, mb.Height) {
				t.Fatalf("overlap detected between %s and %s", ma.Name, mb.Name)
			}
		}
	}
}

func TestRouteManhattan_ConnectivityLayersAndWirelength(t *testing.T) {
	macros := []MacroBlock{
		{Name: "A", Width: 200, Height: 200},
		{Name: "B", Width: 200, Height: 200},
		{Name: "C", Width: 200, Height: 200},
		{Name: "D", Width: 200, Height: 200},
		{Name: "E", Width: 200, Height: 200},
		{Name: "OBS", Width: 240, Height: 240},
	}
	placements := map[string]Placement{
		"A":   {X: 100, Y: 100},
		"B":   {X: 1100, Y: 100},
		"C":   {X: 1100, Y: 900},
		"D":   {X: 100, Y: 900},
		"E":   {X: 1800, Y: 500},
		"OBS": {X: 700, Y: 400},
	}
	nets := []Net{
		{Name: "n_data", Nodes: []string{"A", "B", "C", "D"}},
		{Name: "n_ctrl", Nodes: []string{"A", "E"}},
	}

	const gridStep = 100
	routes, err := RouteManhattan(macros, placements, nets, gridStep)
	if err != nil {
		t.Fatalf("RouteManhattan failed: %v", err)
	}
	if len(routes) != len(nets) {
		t.Fatalf("expected %d routed nets, got %d", len(nets), len(routes))
	}

	macroByName := make(map[string]MacroBlock, len(macros))
	for _, m := range macros {
		macroByName[m.Name] = m
	}
	routeByName := make(map[string]RoutePath, len(routes))
	for _, r := range routes {
		routeByName[r.NetName] = r
	}

	for _, n := range nets {
		r, ok := routeByName[n.Name]
		if !ok {
			t.Fatalf("missing route for required net %s", n.Name)
		}
		if len(r.Segments) == 0 {
			t.Fatalf("route for net %s has no segments", n.Name)
		}

		polylines := splitIntoPolylines(r.Segments)
		requiredConnections := len(n.Nodes) - 1
		if len(polylines) != requiredConnections {
			t.Fatalf("net %s: expected %d routed connections, got %d", n.Name, requiredConnections, len(polylines))
		}

		src := n.Nodes[0]
		for i := 1; i < len(n.Nodes); i++ {
			dst := n.Nodes[i]
			start, goal := edgeTerminals(placements[src], macroByName[src], placements[dst], macroByName[dst], gridStep)
			wantStart := pointFromGrid(start, gridStep)
			wantGoal := pointFromGrid(goal, gridStep)
			gotStart, gotGoal := polylineEndpoints(polylines[i-1])
			if gotStart != wantStart || gotGoal != wantGoal {
				t.Fatalf("net %s connection %s->%s endpoints mismatch: got (%v -> %v), want (%v -> %v)",
					n.Name, src, dst, gotStart, gotGoal, wantStart, wantGoal)
			}
		}

		for _, s := range r.Segments {
			if s.X1 != s.X2 && s.Y1 != s.Y2 {
				t.Fatalf("net %s contains non-Manhattan segment: %+v", n.Name, s)
			}
			if s.X1 == s.X2 {
				if s.Layer != "met2" {
					t.Fatalf("vertical segment must be on met2, got %q: %+v", s.Layer, s)
				}
			} else {
				if s.Layer != "met1" {
					t.Fatalf("horizontal segment must be on met1, got %q: %+v", s.Layer, s)
				}
			}
		}

		wireLen := routeWireLength(r)
		mstLen := manhattanMSTLength(n.Nodes, macroByName, placements)
		if mstLen == 0 {
			t.Fatalf("unexpected zero MST length for net %s", n.Name)
		}
		if wireLen > 2*mstLen {
			t.Fatalf("net %s wirelength too large: got %d, mst=%d, ratio=%.2f (>2.0)",
				n.Name, wireLen, mstLen, float64(wireLen)/float64(mstLen))
		}
	}
}

type point struct{ x, y int }

func pointFromGrid(p gridPoint, step int) point {
	return point{x: p.x * step, y: p.y * step}
}

func splitIntoPolylines(segs []RouteSegment) [][]RouteSegment {
	if len(segs) == 0 {
		return nil
	}
	out := make([][]RouteSegment, 0, 4)
	cur := []RouteSegment{segs[0]}
	for i := 1; i < len(segs); i++ {
		prev := segs[i-1]
		next := segs[i]
		if prev.X2 == next.X1 && prev.Y2 == next.Y1 {
			cur = append(cur, next)
			continue
		}
		out = append(out, cur)
		cur = []RouteSegment{next}
	}
	out = append(out, cur)
	return out
}

func polylineEndpoints(poly []RouteSegment) (point, point) {
	if len(poly) == 0 {
		return point{}, point{}
	}
	first := poly[0]
	last := poly[len(poly)-1]
	return point{first.X1, first.Y1}, point{last.X2, last.Y2}
}

func routeWireLength(r RoutePath) int {
	total := 0
	for _, s := range r.Segments {
		total += mathutil.AbsInt(s.X2-s.X1) + mathutil.AbsInt(s.Y2-s.Y1)
	}
	return total
}

func manhattanMSTLength(nodes []string, macros map[string]MacroBlock, placements map[string]Placement) int {
	if len(nodes) <= 1 {
		return 0
	}

	centers := make([]point, len(nodes))
	for i, n := range nodes {
		m, okM := macros[n]
		p, okP := placements[n]
		if !okM || !okP {
			panic(fmt.Sprintf("missing macro/placement for node %s", n))
		}
		centers[i] = point{x: p.X + m.Width/2, y: p.Y + m.Height/2}
	}

	inTree := make([]bool, len(nodes))
	best := make([]int, len(nodes))
	const inf = int(^uint(0) >> 1)
	for i := range best {
		best[i] = inf
	}
	best[0] = 0
	mst := 0

	for added := 0; added < len(nodes); added++ {
		u := -1
		for i := 0; i < len(nodes); i++ {
			if inTree[i] {
				continue
			}
			if u == -1 || best[i] < best[u] {
				u = i
			}
		}
		if u == -1 {
			break
		}
		inTree[u] = true
		mst += best[u]
		for v := 0; v < len(nodes); v++ {
			if inTree[v] {
				continue
			}
			d := mathutil.AbsInt(centers[u].x-centers[v].x) + mathutil.AbsInt(centers[u].y-centers[v].y)
			if d < best[v] {
				best[v] = d
			}
		}
	}
	return mst
}
