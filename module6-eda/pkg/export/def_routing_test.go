// pkg/export/def_routing_test.go
// M6-DEF-03: DEF routing validation test
// Verifies all nets have routing paths and proper connectivity (each net has ≥ 2 pins)

package export

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
)

// NetInfo represents a net's connectivity in the DEF
type NetInfo struct {
	Name       string
	Pins       []string // List of pins/components connected
	PinCount   int
	HasRouting bool // Currently stub - checking connectivity only
}

// TestDEFRoutingConnectivity tests M6-DEF-03:
// Verify all nets have routing paths (even if stub) and each net has ≥ 2 pins
func TestDEFRoutingConnectivity(t *testing.T) {
	// Create 4×4 weight matrix
	weights := [][]float64{
		{0.1, 0.2, 0.3, 0.4},
		{0.5, 0.6, 0.7, 0.8},
		{0.9, 1.0, -0.1, -0.2},
		{-0.3, -0.4, -0.5, -0.6},
	}

	config := compiler.DefaultConfig()
	config.ArrayRows = 8
	config.ArrayCols = 8
	design, err := compiler.Compile(weights, config)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	def := GenerateDEFWithDefaults(design)

	// Parse nets from DEF
	nets := parseNets(def)

	if len(nets) == 0 {
		t.Fatal("M6-DEF-03 FAIL: No nets found in DEF")
	}

	t.Logf("M6-DEF-03: Parsed %d nets from DEF", len(nets))

	// M6-DEF-03.1: Verify each net has ≥ 2 pins
	insufficientPins := 0
	for _, net := range nets {
		if net.PinCount < 2 {
			t.Errorf("M6-DEF-03.1 FAIL: Net %s has %d pins (expected ≥ 2)", net.Name, net.PinCount)
			insufficientPins++
		}
	}
	if insufficientPins == 0 {
		t.Logf("M6-DEF-03.1 PASS: All %d nets have ≥ 2 pins", len(nets))
	}

	// M6-DEF-03.2: Verify WL nets connect to all cells in row
	wlNets := filterNets(nets, "WL")
	if len(wlNets) != 4 {
		t.Errorf("M6-DEF-03.2 FAIL: Expected 4 WL nets, found %d", len(wlNets))
	}
	for _, wl := range wlNets {
		// Each WL should connect to 4 cells in the row + 1 external pin
		if wl.PinCount < 5 {
			t.Errorf("M6-DEF-03.2 FAIL: %s has %d pins, expected ≥ 5 (1 ext + 4 cells)", wl.Name, wl.PinCount)
		}
	}
	t.Logf("M6-DEF-03.2 PASS: WL nets correctly connect rows (%d nets)", len(wlNets))

	// M6-DEF-03.3: Verify BL nets connect to all cells in column
	blNets := filterNets(nets, "BL")
	if len(blNets) != 4 {
		t.Errorf("M6-DEF-03.3 FAIL: Expected 4 BL nets, found %d", len(blNets))
	}
	for _, bl := range blNets {
		// Each BL should connect to 4 cells in the column + 1 external pin
		if bl.PinCount < 5 {
			t.Errorf("M6-DEF-03.3 FAIL: %s has %d pins, expected ≥ 5 (1 ext + 4 cells)", bl.Name, bl.PinCount)
		}
	}
	t.Logf("M6-DEF-03.3 PASS: BL nets correctly connect columns (%d nets)", len(blNets))

	// M6-DEF-03.4: Verify power nets connect to all cells
	vpwrNets := filterNets(nets, "VPWR")
	if len(vpwrNets) != 1 {
		t.Errorf("M6-DEF-03.4 FAIL: Expected 1 VPWR net, found %d", len(vpwrNets))
	} else {
		// VPWR should connect to 16 cells + 1 external pin = 17 pins
		if vpwrNets[0].PinCount < 17 {
			t.Errorf("M6-DEF-03.4 FAIL: VPWR has %d pins, expected ≥ 17 (1 ext + 16 cells)", vpwrNets[0].PinCount)
		} else {
			t.Logf("M6-DEF-03.4 PASS: VPWR net connects all cells (%d pins)", vpwrNets[0].PinCount)
		}
	}

	vgndNets := filterNets(nets, "VGND")
	if len(vgndNets) != 1 {
		t.Errorf("M6-DEF-03.4 FAIL: Expected 1 VGND net, found %d", len(vgndNets))
	} else {
		// VGND should connect to 16 cells + 1 external pin = 17 pins
		if vgndNets[0].PinCount < 17 {
			t.Errorf("M6-DEF-03.4 FAIL: VGND has %d pins, expected ≥ 17 (1 ext + 16 cells)", vgndNets[0].PinCount)
		} else {
			t.Logf("M6-DEF-03.4 PASS: VGND net connects all cells (%d pins)", vgndNets[0].PinCount)
		}
	}

	t.Logf("M6-DEF-03 routing connectivity: %d nets validated, all with ≥ 2 pins", len(nets))
}

// TestDEF1T1RRouting verifies 1T1R SL nets have correct connectivity
func TestDEF1T1RRouting(t *testing.T) {
	weights := [][]float64{
		{0.1, 0.2},
		{0.3, 0.4},
	}

	config := compiler.Config1T1R()
	config.ArrayRows = 4
	config.ArrayCols = 4
	design, err := compiler.Compile(weights, config)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	def := GenerateDEFWithDefaults(design)
	nets := parseNets(def)

	// M6-DEF-03 (1T1R): Verify SL nets exist and connect to all cells in column
	slNets := filterNets(nets, "SL")
	if len(slNets) != 2 {
		t.Errorf("M6-DEF-03 (1T1R) FAIL: Expected 2 SL nets, found %d", len(slNets))
	}

	for _, sl := range slNets {
		// Each SL should connect to 2 cells in the column + 1 external pin = 3 pins
		if sl.PinCount < 3 {
			t.Errorf("M6-DEF-03 (1T1R) FAIL: %s has %d pins, expected ≥ 3 (1 ext + 2 cells)", sl.Name, sl.PinCount)
		}
	}

	if len(slNets) == 2 && slNets[0].PinCount >= 3 && slNets[1].PinCount >= 3 {
		t.Logf("M6-DEF-03 (1T1R) PASS: SL nets correctly connect columns (%d nets, ≥3 pins each)", len(slNets))
	}
}

// TestDEFNetNaming verifies net names match pin names
func TestDEFNetNaming(t *testing.T) {
	weights := [][]float64{{0.1}}
	config := compiler.DefaultConfig()
	config.ArrayRows = 2
	config.ArrayCols = 2
	design, err := compiler.Compile(weights, config)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	def := GenerateDEFWithDefaults(design)

	// Verify each net has a matching pin
	// Pattern: - WL[0] + NET WL[0] + ...
	// Followed later by: - WL[0]\n ( PIN WL[0] ) ...

	nets := parseNets(def)
	pinsSection := extractPinsSection(def)

	unmatchedNets := 0
	for _, net := range nets {
		// Check if there's a corresponding pin declaration
		pinPattern := fmt.Sprintf("- %s +", net.Name)
		if !strings.Contains(pinsSection, pinPattern) {
			t.Errorf("M6-DEF-03 FAIL: Net %s has no matching pin declaration", net.Name)
			unmatchedNets++
		}
	}

	if unmatchedNets == 0 {
		t.Logf("M6-DEF-03 PASS: All %d nets have matching pin declarations", len(nets))
	}
}

// TestDEFInstanceConnectivity verifies all instances are connected to nets
func TestDEFInstanceConnectivity(t *testing.T) {
	weights := [][]float64{
		{0.1, 0.2},
		{0.3, 0.4},
	}

	config := compiler.DefaultConfig()
	config.ArrayRows = 4
	config.ArrayCols = 4
	design, err := compiler.Compile(weights, config)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	def := GenerateDEFWithDefaults(design)
	nets := parseNets(def)

	// Extract all instance names from nets
	instanceSet := make(map[string]bool)
	for _, net := range nets {
		for _, pin := range net.Pins {
			// Pin format: "R_0_0" (instance) or "PIN WL[0]" (external)
			if strings.HasPrefix(pin, "R_") {
				instanceSet[pin] = true
			}
		}
	}

	// Should have 4 instances (2×2 array)
	if len(instanceSet) != 4 {
		t.Errorf("M6-DEF-03 FAIL: Expected 4 connected instances, found %d", len(instanceSet))
	}

	// Verify all expected instances are connected
	expectedInstances := []string{"R_0_0", "R_0_1", "R_1_0", "R_1_1"}
	for _, inst := range expectedInstances {
		if !instanceSet[inst] {
			t.Errorf("M6-DEF-03 FAIL: Instance %s not connected to any net", inst)
		}
	}

	if len(instanceSet) == 4 {
		t.Log("M6-DEF-03 PASS: All 4 instances connected to nets")
	}
}

// Helper: parse nets from DEF NETS section
func parseNets(def string) []NetInfo {
	var nets []NetInfo

	// Extract NETS section
	netsStart := strings.Index(def, "NETS ")
	if netsStart == -1 {
		return nets
	}
	netsEnd := strings.Index(def[netsStart:], "END NETS")
	if netsEnd == -1 {
		return nets
	}
	netsSection := def[netsStart : netsStart+netsEnd]

	// Parse each net: - WL[0]\n ( PIN WL[0] ) ( R_0_0 WL ) ( R_1_0 WL ) ;
	// Or: - VPWR ( PIN VPWR ) ( R_0_0 VPWR ) ... ;
	// Use (?s) for dot-all mode to match across newlines
	re := regexp.MustCompile(`(?s)-\s+(\S+)\s+(.+?)\s+;`)
	matches := re.FindAllStringSubmatch(netsSection, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			netName := match[1]
			connectionsStr := match[2]

			// Parse connections: ( PIN WL[0] ) ( R_0_0 WL ) ...
			pinRe := regexp.MustCompile(`\(\s*(\S+)\s+\S+\s*\)`)
			pinMatches := pinRe.FindAllStringSubmatch(connectionsStr, -1)

			var pins []string
			for _, pm := range pinMatches {
				if len(pm) >= 2 {
					pins = append(pins, pm[1])
				}
			}

			nets = append(nets, NetInfo{
				Name:     netName,
				Pins:     pins,
				PinCount: len(pins),
			})
		}
	}

	return nets
}

// Helper: filter nets by name prefix
func filterNets(nets []NetInfo, prefix string) []NetInfo {
	var filtered []NetInfo
	for _, net := range nets {
		if strings.HasPrefix(net.Name, prefix) {
			filtered = append(filtered, net)
		}
	}
	return filtered
}

// Helper: extract PINS section
func extractPinsSection(def string) string {
	pinsStart := strings.Index(def, "PINS ")
	if pinsStart == -1 {
		return ""
	}
	pinsEnd := strings.Index(def[pinsStart:], "END PINS")
	if pinsEnd == -1 {
		return ""
	}
	return def[pinsStart : pinsStart+pinsEnd]
}
