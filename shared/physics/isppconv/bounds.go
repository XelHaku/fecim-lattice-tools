// Package isppconv contains UI-neutral ISPP convergence policy helpers.
package isppconv

// Bounds describes the established non-negative search interval for pulse
// magnitude. Min/Max are magnitude values in the caller's native units
// (electric field for Module 1, voltage for LK adapters).
type Bounds struct {
	Min    float64
	Max    float64
	MinSet bool
	MaxSet bool
}

// RecoveryInput describes one collapsed-bounds recovery decision.
type RecoveryInput struct {
	NeedMore         bool
	NeedLess         bool
	CurrentMagnitude float64
	MaxMagnitude     float64
	MinimumWidth     float64
}

// RecoveryReceipt reports the recovered bounds and why they changed.
type RecoveryReceipt struct {
	Bounds           Bounds
	Changed          bool
	ResetToFullRange bool
	Reason           string
}

// RecoverCollapsedBounds widens a collapsed search interval using directional
// evidence when available. It preserves locality for the ISPP guard rule that
// collapsed bounds should be widened minimally, not reset to the full range.
func RecoverCollapsedBounds(bounds Bounds, input RecoveryInput) RecoveryReceipt {
	receipt := RecoveryReceipt{Bounds: bounds}
	if !bounds.MinSet || !bounds.MaxSet || bounds.Min < bounds.Max {
		return receipt
	}

	width := input.MinimumWidth
	if width <= 0 {
		width = 1
	}

	maxMagnitude := input.MaxMagnitude
	if maxMagnitude <= 0 {
		maxMagnitude = bounds.Max
	}

	switch {
	case input.NeedMore:
		min := bounds.Min
		if input.CurrentMagnitude > min {
			min = input.CurrentMagnitude
		}
		max := min + width
		if maxMagnitude > 0 && max > maxMagnitude {
			max = maxMagnitude
			min = max - width
			if min < 0 {
				min = 0
			}
		}
		receipt.Bounds = Bounds{Min: min, Max: max, MinSet: true, MaxSet: true}
		receipt.Changed = true
		receipt.Reason = "directional need-more recovery"
	case input.NeedLess:
		max := bounds.Max
		if input.CurrentMagnitude > 0 && input.CurrentMagnitude < max {
			max = input.CurrentMagnitude
		}
		min := max - width
		if min < 0 {
			min = 0
			max = width
			if maxMagnitude > 0 && max > maxMagnitude {
				max = maxMagnitude
			}
		}
		receipt.Bounds = Bounds{Min: min, Max: max, MinSet: true, MaxSet: true}
		receipt.Changed = true
		receipt.Reason = "directional need-less recovery"
	default:
		receipt.Bounds = Bounds{Min: 0, Max: maxMagnitude, MinSet: false, MaxSet: false}
		receipt.Changed = true
		receipt.ResetToFullRange = true
		receipt.Reason = "unknown direction full-range recovery"
	}
	return receipt
}
