package layout

import (
	"fmt"
	"math"
	"sort"

	"fecim-lattice-tools/shared/mathutil"
)

// MacroBlock is a coarse placement object (e.g., crossbar macro/peripheral macro).
type MacroBlock struct {
	Name   string
	Width  int // DBU
	Height int // DBU
}

// Placement keeps lower-left coordinates for a macro block.
type Placement struct {
	X int
	Y int
}

// Net defines a connection between macro blocks.
type Net struct {
	Name  string
	Nodes []string
}

// RouteSegment is a Manhattan segment for DEF-style routes.
type RouteSegment struct {
	X1, Y1 int
	X2, Y2 int
	Layer  string
}

// RoutePath is one net's routed geometry.
type RoutePath struct {
	NetName  string
	Segments []RouteSegment
}

// PlaceForceDirected places macros using a basic force-directed solver.
// The output is snapped to sitePitchX/sitePitchY and constrained to die bounds.
func PlaceForceDirected(macros []MacroBlock, nets []Net, dieWidth, dieHeight, sitePitchX, sitePitchY int, iterations int) map[string]Placement {
	if len(macros) == 0 {
		return map[string]Placement{}
	}
	if iterations <= 0 {
		iterations = 80
	}
	if sitePitchX <= 0 {
		sitePitchX = 1
	}
	if sitePitchY <= 0 {
		sitePitchY = 1
	}

	macroByName := make(map[string]MacroBlock, len(macros))
	for _, m := range macros {
		macroByName[m.Name] = m
	}

	adj := make(map[string]map[string]struct{})
	for _, n := range nets {
		for i := 0; i < len(n.Nodes); i++ {
			for j := i + 1; j < len(n.Nodes); j++ {
				a, b := n.Nodes[i], n.Nodes[j]
				if adj[a] == nil {
					adj[a] = map[string]struct{}{}
				}
				if adj[b] == nil {
					adj[b] = map[string]struct{}{}
				}
				adj[a][b] = struct{}{}
				adj[b][a] = struct{}{}
			}
		}
	}

	type vec struct{ x, y float64 }
	p := make(map[string]vec, len(macros))

	// Deterministic seed layout: row-major packing.
	x, y := 0, 0
	rowH := 0
	for _, m := range macros {
		if x+m.Width > dieWidth && x > 0 {
			x = 0
			y += rowH + sitePitchY
			rowH = 0
		}
		if y+m.Height > dieHeight {
			y = max(0, dieHeight-m.Height)
		}
		p[m.Name] = vec{float64(x + m.Width/2), float64(y + m.Height/2)}
		x += m.Width + sitePitchX
		if m.Height > rowH {
			rowH = m.Height
		}
	}

	kRepel := 6e5
	kSpring := 0.08
	temperature := 1200.0

	for it := 0; it < iterations; it++ {
		forces := make(map[string]vec, len(macros))

		// Repulsion for overlap reduction.
		for i := 0; i < len(macros); i++ {
			for j := i + 1; j < len(macros); j++ {
				a, b := macros[i], macros[j]
				pa, pb := p[a.Name], p[b.Name]
				dx := pa.x - pb.x
				dy := pa.y - pb.y
				d2 := dx*dx + dy*dy
				if d2 < 1 {
					d2 = 1
					dx += 0.7
					dy += 0.3
				}
				f := kRepel / d2
				invD := 1.0 / math.Sqrt(d2)
				fx := f * dx * invD
				fy := f * dy * invD
				fa, fb := forces[a.Name], forces[b.Name]
				forces[a.Name] = vec{fa.x + fx, fa.y + fy}
				forces[b.Name] = vec{fb.x - fx, fb.y - fy}
			}
		}

		// Springs on connected macros.
		for name, nbrs := range adj {
			pn := p[name]
			for other := range nbrs {
				po := p[other]
				dx := po.x - pn.x
				dy := po.y - pn.y
				fx := kSpring * dx
				fy := kSpring * dy
				f := forces[name]
				forces[name] = vec{f.x + fx, f.y + fy}
			}
		}

		for _, m := range macros {
			pt := p[m.Name]
			f := forces[m.Name]
			mag := math.Sqrt(f.x*f.x + f.y*f.y)
			if mag > temperature && mag > 0 {
				s := temperature / mag
				f.x *= s
				f.y *= s
			}
			pt.x += f.x
			pt.y += f.y

			minX := float64(m.Width) / 2
			maxX := float64(max(m.Width/2, dieWidth-m.Width/2))
			minY := float64(m.Height) / 2
			maxY := float64(max(m.Height/2, dieHeight-m.Height/2))
			if pt.x < minX {
				pt.x = minX
			}
			if pt.x > maxX {
				pt.x = maxX
			}
			if pt.y < minY {
				pt.y = minY
			}
			if pt.y > maxY {
				pt.y = maxY
			}
			p[m.Name] = pt
		}

		temperature *= 0.95
	}

	placed := make(map[string]Placement, len(macros))
	for _, m := range macros {
		c := p[m.Name]
		x0 := snapInt(int(math.Round(c.x))-m.Width/2, sitePitchX)
		y0 := snapInt(int(math.Round(c.y))-m.Height/2, sitePitchY)
		if x0 < 0 {
			x0 = 0
		}
		if y0 < 0 {
			y0 = 0
		}
		if x0+m.Width > dieWidth {
			x0 = max(0, dieWidth-m.Width)
		}
		if y0+m.Height > dieHeight {
			y0 = max(0, dieHeight-m.Height)
		}
		placed[m.Name] = Placement{X: x0, Y: y0}
	}

	resolveOverlaps(placed, macros, dieWidth, dieHeight, sitePitchX, sitePitchY)
	return placed
}

func resolveOverlaps(placed map[string]Placement, macros []MacroBlock, dieWidth, dieHeight, pitchX, pitchY int) {
	byName := map[string]MacroBlock{}
	for _, m := range macros {
		byName[m.Name] = m
	}
	names := make([]string, 0, len(macros))
	for _, m := range macros {
		names = append(names, m.Name)
	}
	sort.Strings(names)

	for iter := 0; iter < len(names)*8; iter++ {
		changed := false
		for i := 0; i < len(names); i++ {
			for j := i + 1; j < len(names); j++ {
				a, b := names[i], names[j]
				pa, pb := placed[a], placed[b]
				ma, mb := byName[a], byName[b]
				if !rectOverlap(pa.X, pa.Y, ma.Width, ma.Height, pb.X, pb.Y, mb.Width, mb.Height) {
					continue
				}
				// Push lexicographically later macro to the right/down.
				newX := snapInt(pb.X+mb.Width+pitchX, pitchX)
				if newX+mb.Width > dieWidth {
					newX = 0
					pb.Y = snapInt(pb.Y+mb.Height+pitchY, pitchY)
				}
				pb.X = newX
				if pb.Y+mb.Height > dieHeight {
					pb.Y = max(0, dieHeight-mb.Height)
				}
				placed[b] = pb
				changed = true
			}
		}
		if !changed {
			return
		}
	}
}

func rectOverlap(ax, ay, aw, ah, bx, by, bw, bh int) bool {
	return ax < bx+bw && ax+aw > bx && ay < by+bh && ay+ah > by
}

func snapInt(v, step int) int {
	if step <= 1 {
		return v
	}
	return (v / step) * step
}

// RouteManhattan routes all nets on a coarse Manhattan grid using BFS with block obstacles.
func RouteManhattan(macros []MacroBlock, placements map[string]Placement, nets []Net, gridStep int) ([]RoutePath, error) {
	if gridStep <= 0 {
		gridStep = 80
	}

	byName := map[string]MacroBlock{}
	for _, m := range macros {
		byName[m.Name] = m
	}

	// Grid bounds from placed geometry.
	maxX, maxY := 0, 0
	for _, m := range macros {
		p := placements[m.Name]
		if p.X+m.Width > maxX {
			maxX = p.X + m.Width
		}
		if p.Y+m.Height > maxY {
			maxY = p.Y + m.Height
		}
	}
	w := maxX/gridStep + 4
	h := maxY/gridStep + 4

	blocked := make([][]bool, h)
	for y := 0; y < h; y++ {
		blocked[y] = make([]bool, w)
	}
	for _, m := range macros {
		p := placements[m.Name]
		x0 := p.X / gridStep
		y0 := p.Y / gridStep
		x1 := (p.X + m.Width) / gridStep
		y1 := (p.Y + m.Height) / gridStep
		for y := y0; y <= y1 && y < h; y++ {
			for x := x0; x <= x1 && x < w; x++ {
				if x >= 0 && y >= 0 {
					blocked[y][x] = true
				}
			}
		}
	}

	var routes []RoutePath
	for _, net := range nets {
		if len(net.Nodes) < 2 {
			continue
		}
		src := net.Nodes[0]
		path := RoutePath{NetName: net.Name}
		for i := 1; i < len(net.Nodes); i++ {
			dst := net.Nodes[i]
			pm := byName[src]
			pn := byName[dst]
			ps := placements[src]
			pd := placements[dst]
			start, goal := edgeTerminals(ps, pm, pd, pn, gridStep)

			gBlocked := copyBlocked(blocked)
			if inBounds(start.x, start.y, w, h) {
				gBlocked[start.y][start.x] = false
			}
			if inBounds(goal.x, goal.y, w, h) {
				gBlocked[goal.y][goal.x] = false
			}

			pts, ok := bfsPath(start, goal, gBlocked, w, h)
			if !ok {
				return nil, fmt.Errorf("failed to route net %s (%s->%s)", net.Name, src, dst)
			}
			path.Segments = append(path.Segments, compressToSegments(pts, gridStep)...)
		}
		routes = append(routes, path)
	}
	return routes, nil
}

type gridPoint struct{ x, y int }

func bfsPath(start, goal gridPoint, blocked [][]bool, w, h int) ([]gridPoint, bool) {
	if !inBounds(start.x, start.y, w, h) || !inBounds(goal.x, goal.y, w, h) {
		return nil, false
	}
	q := []gridPoint{start}
	seen := map[gridPoint]bool{start: true}
	prev := map[gridPoint]gridPoint{}
	dirs := []gridPoint{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}

	for len(q) > 0 {
		cur := q[0]
		q = q[1:]
		if cur == goal {
			break
		}
		for _, d := range dirs {
			nx, ny := cur.x+d.x, cur.y+d.y
			np := gridPoint{nx, ny}
			if !inBounds(nx, ny, w, h) || blocked[ny][nx] || seen[np] {
				continue
			}
			seen[np] = true
			prev[np] = cur
			q = append(q, np)
		}
	}

	if !seen[goal] {
		return nil, false
	}
	var path []gridPoint
	for cur := goal; ; {
		path = append(path, cur)
		if cur == start {
			break
		}
		cur = prev[cur]
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path, true
}

func compressToSegments(points []gridPoint, step int) []RouteSegment {
	if len(points) < 2 {
		return nil
	}
	var segs []RouteSegment
	start := points[0]
	prev := points[1]
	dx := prev.x - start.x
	dy := prev.y - start.y
	for i := 2; i < len(points); i++ {
		ndx := points[i].x - prev.x
		ndy := points[i].y - prev.y
		if ndx != dx || ndy != dy {
			segs = append(segs, RouteSegment{
				X1:    start.x * step,
				Y1:    start.y * step,
				X2:    prev.x * step,
				Y2:    prev.y * step,
				Layer: layerFor(dx, dy),
			})
			start = prev
			dx, dy = ndx, ndy
		}
		prev = points[i]
	}
	segs = append(segs, RouteSegment{X1: start.x * step, Y1: start.y * step, X2: prev.x * step, Y2: prev.y * step, Layer: layerFor(dx, dy)})
	return segs
}

func layerFor(dx, dy int) string {
	if dy != 0 {
		return "met2"
	}
	return "met1"
}

func edgeTerminals(ps Placement, ms MacroBlock, pd Placement, md MacroBlock, step int) (gridPoint, gridPoint) {
	sx, sy := ps.X+ms.Width/2, ps.Y+ms.Height/2
	dx, dy := pd.X+md.Width/2, pd.Y+md.Height/2

	if mathutil.AbsInt(dx-sx) >= mathutil.AbsInt(dy-sy) {
		if dx >= sx {
			sx = ps.X + ms.Width + step
			dx = pd.X - step
		} else {
			sx = ps.X - step
			dx = pd.X + md.Width + step
		}
	} else {
		if dy >= sy {
			sy = ps.Y + ms.Height + step
			dy = pd.Y - step
		} else {
			sy = ps.Y - step
			dy = pd.Y + md.Height + step
		}
	}
	return gridPoint{sx / step, sy / step}, gridPoint{dx / step, dy / step}
}

func copyBlocked(src [][]bool) [][]bool {
	out := make([][]bool, len(src))
	for i := range src {
		out[i] = make([]bool, len(src[i]))
		copy(out[i], src[i])
	}
	return out
}

func inBounds(x, y, w, h int) bool {
	return x >= 0 && y >= 0 && x < w && y < h
}
