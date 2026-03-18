// Package export provides unified data export utilities for FeCIM tools.
// This file defines standard simulation metadata and disclaimers that must
// accompany all exported files to prevent misuse of educational defaults.
package export

import (
	"strings"
	"time"
)

// SimulationMetadata provides standard metadata for all exported files.
type SimulationMetadata struct {
	Tool        string `json:"tool"`
	Version     string `json:"version"`
	Disclaimer  string `json:"disclaimer"`
	GeneratedAt string `json:"generated_at"`
	GitCommit   string `json:"git_commit,omitempty"`
}

// DefaultSimulationMetadata returns standard metadata for research exports.
//
// The disclaimer clearly marks all outputs as simulation-only, directing
// users to the honesty audit document for parameter provenance details.
func DefaultSimulationMetadata() SimulationMetadata {
	return SimulationMetadata{
		Tool:        "FeCIM Lattice Tools",
		Version:     "0.1.0-education",
		Disclaimer:  "SIMULATION ONLY — not validated against silicon measurements. Parameters are educational defaults unless otherwise noted. See docs/4-research/honesty-audit.md.",
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// SimulationDisclaimer returns a plain-text disclaimer for non-SPICE text exports.
func SimulationDisclaimer() string {
	return "SIMULATION ONLY — FeCIM Lattice Tools v0.1.0-education\n" +
		"Parameters are educational defaults, NOT validated against silicon.\n" +
		"See: docs/4-research/honesty-audit.md\n"
}

// SPICEDisclaimer returns a standard comment block for SPICE netlist headers.
//
// The block uses SPICE comment syntax (* prefix) and contains a clear
// warning that the netlist parameters are educational defaults.
func SPICEDisclaimer() string {
	var b strings.Builder
	b.WriteString("* ============================================================\n")
	b.WriteString("* SIMULATION ONLY — FeCIM Lattice Tools\n")
	b.WriteString("* Version: 0.1.0-education\n")
	b.WriteString("*\n")
	b.WriteString("* Parameters are educational defaults, NOT validated against\n")
	b.WriteString("* silicon measurements. Do not use for tapeout or hardware\n")
	b.WriteString("* correlation without independent calibration.\n")
	b.WriteString("*\n")
	b.WriteString("* See: docs/4-research/honesty-audit.md\n")
	b.WriteString("* ============================================================\n")
	return b.String()
}
