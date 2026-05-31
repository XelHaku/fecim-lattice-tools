//go:build legacy_fyne

// Package widgets provides shared UI components for FeCIM visualizers.
package widgets

import (
	"image/color"

	"fecim-lattice-tools/shared/widgets/display"
)

// ClaimStatusValue represents the verification outcome for a claim.
type ClaimStatusValue = display.ClaimStatusValue

const (
	ClaimPass     = display.ClaimPass
	ClaimFail     = display.ClaimFail
	ClaimUntested = display.ClaimUntested
)

// ClaimStatus holds one scientific claim's verification metadata.
type ClaimStatus = display.ClaimStatus

// ValidationDashboard is a Fyne container that shows a scrollable list of
// scientific claims with provenance badges and pass/fail status icons.
type ValidationDashboard = display.ValidationDashboard

// NewValidationDashboard builds the dashboard pre-populated with claims from
// the CLAIMS-MATRIX (docs/4-research/validation/claims/scientific-claims-matrix.md).
func NewValidationDashboard() *ValidationDashboard { return display.NewValidationDashboard() }

func statusColor(s ClaimStatusValue) color.Color { return display.StatusColor(s) }
