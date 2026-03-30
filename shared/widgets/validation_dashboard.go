// Package widgets provides shared UI components for FeCIM visualizers.
//
// validation_dashboard.go implements a scrollable dashboard widget that displays
// physics claim pass/fail status from the CLAIMS-MATRIX.
package widgets

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ClaimStatusValue represents the verification outcome for a claim.
type ClaimStatusValue string

const (
	ClaimPass     ClaimStatusValue = "Pass"
	ClaimFail     ClaimStatusValue = "Fail"
	ClaimUntested ClaimStatusValue = "Untested"
)

// ClaimStatus holds one scientific claim's verification metadata.
type ClaimStatus struct {
	Name        string
	Description string
	Provenance  ConfidenceLevel
	TestName    string
	Status      ClaimStatusValue
	Tolerance   string
}

// ValidationDashboard is a Fyne container that shows a scrollable list of
// scientific claims with provenance badges and pass/fail status icons.
type ValidationDashboard struct {
	Claims    []ClaimStatus
	Container *fyne.Container
}

// NewValidationDashboard builds the dashboard pre-populated with claims from
// the CLAIMS-MATRIX (docs/4-research/validation/CLAIMS-MATRIX.md).
func NewValidationDashboard() *ValidationDashboard {
	vd := &ValidationDashboard{
		Claims: defaultClaims(),
	}
	vd.Container = vd.build()
	return vd
}

// build constructs the scrollable VBox layout.
func (vd *ValidationDashboard) build() *fyne.Container {
	rows := container.NewVBox()
	for i, c := range vd.Claims {
		rows.Add(vd.buildRow(i, c))
		rows.Add(widget.NewSeparator())
	}

	scroll := container.NewVScroll(rows)
	scroll.SetMinSize(fyne.NewSize(500, 300))

	title := widget.NewLabelWithStyle(
		"Validation Dashboard", fyne.TextAlignCenter, fyne.TextStyle{Bold: true},
	)
	subtitle := widget.NewLabel(
		fmt.Sprintf("%d claims (%d pass, %d untested)",
			len(vd.Claims), vd.countStatus(ClaimPass), vd.countStatus(ClaimUntested)),
	)
	subtitle.Alignment = fyne.TextAlignCenter

	return container.NewBorder(
		container.NewVBox(title, subtitle, widget.NewSeparator()),
		nil, nil, nil, scroll,
	)
}

// countStatus returns how many claims have the given status.
func (vd *ValidationDashboard) countStatus(s ClaimStatusValue) int {
	n := 0
	for _, c := range vd.Claims {
		if c.Status == s {
			n++
		}
	}
	return n
}

// buildRow renders a single claim row: badge dot + name + status dot + tolerance.
func (vd *ValidationDashboard) buildRow(_ int, c ClaimStatus) *fyne.Container {
	provDot := canvas.NewCircle(confidenceColor(c.Provenance))
	provDot.Resize(fyne.NewSize(10, 10))
	provLabel := widget.NewLabel(string(c.Provenance))
	provLabel.TextStyle = fyne.TextStyle{Bold: true}

	statusDot := canvas.NewCircle(statusColor(c.Status))
	statusDot.Resize(fyne.NewSize(10, 10))
	statusLabel := widget.NewLabel(string(c.Status))

	nameLabel := widget.NewLabel(c.Name)
	nameLabel.TextStyle = fyne.TextStyle{Bold: true}

	tolLabel := widget.NewLabel(c.Tolerance)

	left := container.NewHBox(
		container.NewGridWrap(fyne.NewSize(10, 10), provDot),
		provLabel,
	)
	right := container.NewHBox(
		container.NewGridWrap(fyne.NewSize(10, 10), statusDot),
		statusLabel,
		widget.NewSeparator(),
		tolLabel,
	)

	return container.NewVBox(
		container.NewHBox(nameLabel),
		container.NewHBox(left, widget.NewSeparator(), right),
	)
}

// statusColor maps a ClaimStatusValue to a display color.
func statusColor(s ClaimStatusValue) color.Color {
	switch s {
	case ClaimPass:
		return color.RGBA{60, 180, 75, 255} // green
	case ClaimFail:
		return color.RGBA{220, 80, 80, 255} // red
	default:
		return color.RGBA{160, 160, 160, 255} // gray
	}
}

// defaultClaims returns the hardcoded set from CLAIMS-MATRIX.md.
func defaultClaims() []ClaimStatus { //nolint:funlen // data table
	c := func(name, desc string, p ConfidenceLevel, test string, s ClaimStatusValue, tol string) ClaimStatus {
		return ClaimStatus{Name: name, Description: desc, Provenance: p, TestName: test, Status: s, Tolerance: tol}
	}
	return []ClaimStatus{
		c("Pr = 19.17 uC/cm2 (HZO)", "Remnant polarization (Materlik 2015)", Measured, "TestPhysicsRegressionCurves", ClaimPass, "+/-5%"),
		c("Ec = 1.16 MV/cm", "Coercive field for HZO", Measured, "TestPhysicsRegressionCurves", ClaimPass, "+/-5%"),
		c("MNIST >= 80% accuracy", "Full-stack neural inference on crossbar", Calibrated, "TestFullStackMNIST", ClaimUntested, ">=80%"),
		c("Energy 44.94 fJ/cell", "Transient pulse energy per cell", Estimated, "TestTransientPulse (removed)", ClaimUntested, "10-100 fJ"),
		c("M4: Power conservation", "Power conservation below 1%", Measured, "TestThermodynamics*", ClaimPass, "<1%"),
		c("M4: KCL residual", "Kirchhoff current law residual", Measured, "TestKirchhoff*", ClaimPass, "<1e-12"),
		c("M4: MVM accuracy", "Matrix-vector multiply (Tier-A)", Calibrated, "TestComputeMVM*", ClaimPass, "<5%"),
		c("M4: BER", "Bit error rate", Calibrated, "TestComputeBER*", ClaimUntested, "<5%"),
		c("M4: Read margin", "Read margin above 3 sigma", Calibrated, "TestReadMarginBER*", ClaimUntested, ">3 sigma"),
		c("M4: DAC INL", "DAC integral nonlinearity", Measured, "TestPeripheralsINLDNL*", ClaimPass, "<1 LSB"),
		c("M4: Retention", "Conductance drift over time", Estimated, "TestRetention*", ClaimUntested, "dG/G <1%"),
		c("M4: Write disturb", "Write disturb bounded", Estimated, "TestWriteDisturb*", ClaimUntested, "bounded"),
		c("0 spurious discontinuities", "All discontinuities physical", Measured, "TestHeadlessISPPContinuityValidation", ClaimPass, "0 spurious"),
		c("ISPP convergence", "All target levels reached", Calibrated, "TestISPPConverges_*", ClaimPass, "all levels"),
		c("Literature RMSE", "RMSE vs experimental datasets", Measured, "TestExperimentalDataValidation", ClaimPass, "RMSE thresh"),
		c("Array ISPP disturb", "Array ISPP with disturb tracking", Calibrated, "TestArrayISPP", ClaimPass, "MaxDisturb<0.3"),
	}
}
