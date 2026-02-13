package validate

import (
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// DRCRules defines minimum geometry requirements in microns.
type DRCRules struct {
	MinMetalWidth   float64
	MinMetalSpacing float64
	MinViaEnclosure float64
}

// DefaultSKY130DRCRules provides conservative SKY130 defaults for met1.
func DefaultSKY130DRCRules() DRCRules {
	return DRCRules{
		MinMetalWidth:   0.14,
		MinMetalSpacing: 0.14,
		MinViaEnclosure: 0.06,
	}
}

type lefRect struct {
	Layer string
	X1    float64
	Y1    float64
	X2    float64
	Y2    float64
	Pin   string
}

func (r lefRect) width() float64 { return math.Min(r.X2-r.X1, r.Y2-r.Y1) }

func (r lefRect) isVia() bool {
	l := strings.ToLower(r.Layer)
	return strings.Contains(l, "via") || strings.HasPrefix(l, "li")
}

func (r lefRect) isMetal() bool {
	l := strings.ToLower(r.Layer)
	return strings.HasPrefix(l, "met") || l == "metal1" || l == "metal2"
}

func ValidateLEFDRCFile(lefPath string, rules DRCRules) error {
	content, err := os.ReadFile(lefPath)
	if err != nil {
		return fmt.Errorf("read lef: %w", err)
	}
	return ValidateLEFDRC(string(content), rules)
}

func ValidateLEFWithPDKConstraintsFile(lefPath string, rules DRCRules) error {
	content, err := os.ReadFile(lefPath)
	if err != nil {
		return fmt.Errorf("read lef: %w", err)
	}
	return ValidateLEFWithPDKConstraints(string(content), rules)
}

// ValidateLEFDRC checks basic LEF geometry against provided rules.
func ValidateLEFDRC(lefContent string, rules DRCRules) error {
	rects, err := parseLEFRects(lefContent)
	if err != nil {
		return err
	}

	for _, r := range rects {
		if !r.isMetal() {
			continue
		}
		if r.width() < rules.MinMetalWidth {
			return fmt.Errorf("metal width violation on %s rect %.3f %.3f %.3f %.3f: width %.3fum < %.3fum",
				r.Layer, r.X1, r.Y1, r.X2, r.Y2, r.width(), rules.MinMetalWidth)
		}
	}

	for i := 0; i < len(rects); i++ {
		for j := i + 1; j < len(rects); j++ {
			a, b := rects[i], rects[j]
			if !a.isMetal() || !b.isMetal() {
				continue
			}
			if strings.ToLower(a.Layer) != strings.ToLower(b.Layer) {
				continue
			}
			if overlap2D(a, b) {
				continue
			}
			spacing := rectSpacing(a, b)
			if spacing < rules.MinMetalSpacing {
				return fmt.Errorf("metal spacing violation on %s: spacing %.3fum < %.3fum", a.Layer, spacing, rules.MinMetalSpacing)
			}
		}
	}

	for _, via := range rects {
		if !via.isVia() {
			continue
		}
		ok := false
		for _, metal := range rects {
			if !metal.isMetal() {
				continue
			}
			if encloses(metal, via, rules.MinViaEnclosure) {
				ok = true
				break
			}
		}
		if !ok {
			return fmt.Errorf("via enclosure violation on %s rect %.3f %.3f %.3f %.3f", via.Layer, via.X1, via.Y1, via.X2, via.Y2)
		}
	}

	return nil
}

// ValidateLEFWithPDKConstraints extends DRC with pin-within-bounds checks.
func ValidateLEFWithPDKConstraints(lefContent string, rules DRCRules) error {
	if err := ValidateLEFDRC(lefContent, rules); err != nil {
		return err
	}
	rects, err := parseLEFRects(lefContent)
	if err != nil {
		return err
	}
	w, h, ok := parseLEFMacroSize(lefContent)
	if !ok {
		return fmt.Errorf("missing MACRO SIZE in LEF")
	}
	for _, r := range rects {
		if r.Pin == "" {
			continue
		}
		if r.X1 < 0 || r.Y1 < 0 || r.X2 > w || r.Y2 > h {
			return fmt.Errorf("pin %s rectangle out of bounds: rect %.3f %.3f %.3f %.3f outside macro %.3f x %.3f", r.Pin, r.X1, r.Y1, r.X2, r.Y2, w, h)
		}
	}
	return nil
}

func parseLEFRects(lef string) ([]lefRect, error) {
	lines := strings.Split(lef, "\n")
	layerRe := regexp.MustCompile(`^LAYER\s+(\S+)`)
	rectRe := regexp.MustCompile(`^RECT\s+([\d.\-]+)\s+([\d.\-]+)\s+([\d.\-]+)\s+([\d.\-]+)`) // LEF rect
	pinRe := regexp.MustCompile(`^PIN\s+(\S+)`)

	currentLayer := ""
	currentPin := ""
	var rects []lefRect

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if m := pinRe.FindStringSubmatch(line); m != nil {
			currentPin = m[1]
			continue
		}
		if strings.HasPrefix(line, "END ") && strings.TrimSpace(strings.TrimPrefix(line, "END ")) == currentPin {
			currentPin = ""
		}
		if m := layerRe.FindStringSubmatch(line); m != nil {
			currentLayer = m[1]
			continue
		}
		if m := rectRe.FindStringSubmatch(line); m != nil {
			x1, _ := strconv.ParseFloat(m[1], 64)
			y1, _ := strconv.ParseFloat(m[2], 64)
			x2, _ := strconv.ParseFloat(m[3], 64)
			y2, _ := strconv.ParseFloat(m[4], 64)
			if x2 < x1 || y2 < y1 {
				return nil, fmt.Errorf("invalid rect coordinates: %s", line)
			}
			rects = append(rects, lefRect{Layer: currentLayer, X1: x1, Y1: y1, X2: x2, Y2: y2, Pin: currentPin})
		}
	}

	return rects, nil
}

func overlap2D(a, b lefRect) bool {
	return a.X1 < b.X2 && a.X2 > b.X1 && a.Y1 < b.Y2 && a.Y2 > b.Y1
}

func rectSpacing(a, b lefRect) float64 {
	dx := math.Max(0, math.Max(a.X1-b.X2, b.X1-a.X2))
	dy := math.Max(0, math.Max(a.Y1-b.Y2, b.Y1-a.Y2))
	if dx == 0 {
		return dy
	}
	if dy == 0 {
		return dx
	}
	return math.Hypot(dx, dy)
}

func parseLEFMacroSize(lef string) (float64, float64, bool) {
	sizeRe := regexp.MustCompile(`(?m)^\s*SIZE\s+([\d.]+)\s+BY\s+([\d.]+)`) // SIZE w BY h
	m := sizeRe.FindStringSubmatch(lef)
	if m == nil {
		return 0, 0, false
	}
	w, errW := strconv.ParseFloat(m[1], 64)
	h, errH := strconv.ParseFloat(m[2], 64)
	if errW != nil || errH != nil {
		return 0, 0, false
	}
	return w, h, true
}

func encloses(metal, via lefRect, enclosure float64) bool {
	return metal.X1 <= via.X1-enclosure &&
		metal.Y1 <= via.Y1-enclosure &&
		metal.X2 >= via.X2+enclosure &&
		metal.Y2 >= via.Y2+enclosure
}
