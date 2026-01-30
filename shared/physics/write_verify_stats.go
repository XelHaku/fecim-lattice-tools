// Package physics provides shared physics utilities for FeCIM simulations.
// This file implements write-verify statistics tracking for ISPP operations.
//
// C12: Per Dr. Tour critique - add write-verify statistics visualization
// showing pulses to converge, failure rate vs cycles.
//
// Physics basis:
//   - ISPP (Incremental Step Pulse Programming) uses iterative write-verify
//   - Typical convergence: 1-5 pulses for well-characterized devices
//   - Failure modes: overshoot, max iterations, stuck bits
//   - Failure rate increases with cycle count (fatigue)
//
// References:
//   - IEEE IEDM 2022 (FeFET programming statistics)
//   - Nature Electronics 2023 (analog state precision)
package physics

import (
	"math"
	"sync"
)

// WriteVerifyStats tracks statistics for write-verify (ISPP) operations.
type WriteVerifyStats struct {
	mu sync.RWMutex

	// Total counts
	TotalWrites     int // Total write operations attempted
	SuccessfulWrites int // Writes that converged to target
	FailedWrites    int // Writes that failed (max iterations, stuck)

	// Pulses to converge histogram (index = pulses, value = count)
	// PulsesHistogram[0] = writes that converged in 1 pulse
	// PulsesHistogram[4] = writes that converged in 5 pulses
	PulsesHistogram [10]int // Support up to 10 pulses

	// Overshoot statistics
	OvershootCount int // Number of overshoots requiring reset
	ResetCount     int // Number of reset operations

	// Cycle-dependent failure tracking
	CycleCount       int       // Total write cycles on array
	FailureRateHistory []float64 // Failure rate at each 100-cycle checkpoint

	// Per-level statistics (which levels are hardest to hit)
	LevelAttempts  [32]int // Attempts per target level
	LevelSuccesses [32]int // Successes per target level

	// Timing statistics (in microseconds)
	TotalWriteTimeUs float64 // Total time spent in write operations
	AvgPulsesPerWrite float64 // Running average pulses per write
}

// NewWriteVerifyStats creates a new statistics tracker.
func NewWriteVerifyStats() *WriteVerifyStats {
	return &WriteVerifyStats{
		FailureRateHistory: make([]float64, 0, 100),
	}
}

// RecordWrite records a completed write operation.
func (s *WriteVerifyStats) RecordWrite(targetLevel int, pulsesUsed int, success bool, hadOvershoot bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.TotalWrites++
	s.CycleCount++

	if targetLevel >= 0 && targetLevel < 32 {
		s.LevelAttempts[targetLevel]++
		if success {
			s.LevelSuccesses[targetLevel]++
		}
	}

	if success {
		s.SuccessfulWrites++
		if pulsesUsed > 0 && pulsesUsed <= 10 {
			s.PulsesHistogram[pulsesUsed-1]++
		}
	} else {
		s.FailedWrites++
	}

	if hadOvershoot {
		s.OvershootCount++
	}

	// Update running average
	s.AvgPulsesPerWrite = (s.AvgPulsesPerWrite*float64(s.TotalWrites-1) + float64(pulsesUsed)) / float64(s.TotalWrites)

	// Record failure rate every 100 cycles
	if s.CycleCount%100 == 0 && s.TotalWrites > 0 {
		failureRate := float64(s.FailedWrites) / float64(s.TotalWrites)
		s.FailureRateHistory = append(s.FailureRateHistory, failureRate)
	}
}

// RecordReset records a reset operation (due to overshoot).
func (s *WriteVerifyStats) RecordReset() {
	s.mu.Lock()
	s.ResetCount++
	s.mu.Unlock()
}

// GetSuccessRate returns the overall success rate (0-1).
func (s *WriteVerifyStats) GetSuccessRate() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.TotalWrites == 0 {
		return 1.0 // No data yet
	}
	return float64(s.SuccessfulWrites) / float64(s.TotalWrites)
}

// GetFailureRate returns the overall failure rate (0-1).
func (s *WriteVerifyStats) GetFailureRate() float64 {
	return 1.0 - s.GetSuccessRate()
}

// GetAveragePulses returns the average pulses per successful write.
func (s *WriteVerifyStats) GetAveragePulses() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.SuccessfulWrites == 0 {
		return 0
	}

	totalPulses := 0
	for i, count := range s.PulsesHistogram {
		totalPulses += (i + 1) * count
	}
	return float64(totalPulses) / float64(s.SuccessfulWrites)
}

// GetPulsesHistogram returns a copy of the pulses histogram.
func (s *WriteVerifyStats) GetPulsesHistogram() [10]int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.PulsesHistogram
}

// GetLevelSuccessRates returns success rate for each level (0-31).
func (s *WriteVerifyStats) GetLevelSuccessRates() [32]float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var rates [32]float64
	for i := 0; i < 32; i++ {
		if s.LevelAttempts[i] > 0 {
			rates[i] = float64(s.LevelSuccesses[i]) / float64(s.LevelAttempts[i])
		} else {
			rates[i] = 1.0 // No attempts = assume perfect
		}
	}
	return rates
}

// GetHardestLevels returns the indices of levels with lowest success rates.
func (s *WriteVerifyStats) GetHardestLevels(n int) []int {
	rates := s.GetLevelSuccessRates()

	// Find n levels with lowest success rate (that have been attempted)
	type levelRate struct {
		level int
		rate  float64
		attempts int
	}

	s.mu.RLock()
	attempts := s.LevelAttempts
	s.mu.RUnlock()

	var items []levelRate
	for i, rate := range rates {
		if attempts[i] > 0 {
			items = append(items, levelRate{i, rate, attempts[i]})
		}
	}

	// Sort by rate ascending (lowest success = hardest)
	for i := 0; i < len(items)-1; i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].rate < items[i].rate {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	result := make([]int, 0, n)
	for i := 0; i < len(items) && i < n; i++ {
		result = append(result, items[i].level)
	}
	return result
}

// GetFailureRateVsCycles returns the failure rate history.
func (s *WriteVerifyStats) GetFailureRateVsCycles() []float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]float64, len(s.FailureRateHistory))
	copy(result, s.FailureRateHistory)
	return result
}

// GetOvershootRate returns the fraction of writes that required overshoot handling.
func (s *WriteVerifyStats) GetOvershootRate() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.TotalWrites == 0 {
		return 0
	}
	return float64(s.OvershootCount) / float64(s.TotalWrites)
}

// GetSummary returns a human-readable summary string.
func (s *WriteVerifyStats) GetSummary() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return formatSummary(s)
}

// formatSummary formats the statistics summary (called with lock held).
func formatSummary(s *WriteVerifyStats) string {
	if s.TotalWrites == 0 {
		return "No writes recorded"
	}

	successRate := float64(s.SuccessfulWrites) / float64(s.TotalWrites) * 100
	avgPulses := s.AvgPulsesPerWrite
	overshootRate := float64(s.OvershootCount) / float64(s.TotalWrites) * 100

	return formatStatsString(s.TotalWrites, s.SuccessfulWrites, successRate, avgPulses, overshootRate)
}

// formatStatsString builds the formatted string.
func formatStatsString(total, successful int, successRate, avgPulses, overshootRate float64) string {
	// Using simple concatenation to avoid fmt import issues
	return "Writes: " + itoa(total) +
		" | Success: " + ftoa(successRate, 1) + "%" +
		" | Avg pulses: " + ftoa(avgPulses, 1) +
		" | Overshoot: " + ftoa(overshootRate, 1) + "%"
}

// Simple int to string
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		return "-" + string(digits)
	}
	return string(digits)
}

// Simple float to string with precision
func ftoa(f float64, prec int) string {
	if math.IsNaN(f) {
		return "NaN"
	}
	if math.IsInf(f, 0) {
		return "Inf"
	}

	neg := f < 0
	if neg {
		f = -f
	}

	// Multiply by 10^prec and round
	mult := math.Pow(10, float64(prec))
	rounded := int(f*mult + 0.5)

	intPart := rounded / int(mult)
	fracPart := rounded % int(mult)

	result := itoa(intPart)
	if prec > 0 {
		fracStr := itoa(fracPart)
		// Pad with leading zeros
		for len(fracStr) < prec {
			fracStr = "0" + fracStr
		}
		result += "." + fracStr
	}

	if neg {
		return "-" + result
	}
	return result
}

// Reset clears all statistics.
func (s *WriteVerifyStats) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.TotalWrites = 0
	s.SuccessfulWrites = 0
	s.FailedWrites = 0
	s.PulsesHistogram = [10]int{}
	s.OvershootCount = 0
	s.ResetCount = 0
	s.CycleCount = 0
	s.FailureRateHistory = make([]float64, 0, 100)
	s.LevelAttempts = [32]int{}
	s.LevelSuccesses = [32]int{}
	s.TotalWriteTimeUs = 0
	s.AvgPulsesPerWrite = 0
}

// SimulateFailureRateProgression simulates how failure rate increases with cycles.
// Uses a stretched exponential model based on FeFET fatigue literature.
// Returns failure rate at given cycle count.
func SimulateFailureRateProgression(cycles int, enduranceLimit float64) float64 {
	// Stretched exponential fatigue model
	// FailureRate = baseline + (1-baseline) * (1 - exp(-(N/N0)^beta))
	// where beta ≈ 0.3-0.5 for typical FeFET fatigue

	baseline := 0.001 // 0.1% baseline failure rate
	beta := 0.35

	if float64(cycles) >= enduranceLimit {
		return 1.0 // Complete failure at endurance limit
	}

	fatigueContribution := 1.0 - math.Exp(-math.Pow(float64(cycles)/enduranceLimit, beta))
	return baseline + (1.0-baseline)*fatigueContribution*0.1 // Cap at 10% max failure rate before endurance limit
}
