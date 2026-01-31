package validation

import (
	"math"
	"sort"
)

// KolmogorovSmirnovTest performs a two-sample KS test
// Returns the KS statistic and approximate p-value
func KolmogorovSmirnovTest(sample1, sample2 []float64) (statistic, pValue float64) {
	n1 := float64(len(sample1))
	n2 := float64(len(sample2))

	// Sort both samples
	sorted1 := make([]float64, len(sample1))
	sorted2 := make([]float64, len(sample2))
	copy(sorted1, sample1)
	copy(sorted2, sample2)
	sort.Float64s(sorted1)
	sort.Float64s(sorted2)

	// Compute empirical CDFs and find maximum difference
	i, j := 0, 0
	maxDiff := 0.0

	for i < len(sorted1) && j < len(sorted2) {
		cdf1 := float64(i+1) / n1
		cdf2 := float64(j+1) / n2

		diff := math.Abs(cdf1 - cdf2)
		if diff > maxDiff {
			maxDiff = diff
		}

		if sorted1[i] < sorted2[j] {
			i++
		} else {
			j++
		}
	}

	statistic = maxDiff

	// Approximate p-value using Kolmogorov distribution
	// For large samples, this is approximately valid
	effectiveN := math.Sqrt((n1 * n2) / (n1 + n2))
	lambda := (effectiveN + 0.12 + 0.11/effectiveN) * statistic

	// Smirnov's approximation for p-value
	pValue = 2.0 * math.Exp(-2.0*lambda*lambda)
	if pValue > 1.0 {
		pValue = 1.0
	}

	return statistic, pValue
}

// ChiSquaredTest performs a chi-squared goodness-of-fit test
// observed: observed frequencies
// expected: expected frequencies
// Returns chi-squared statistic and degrees of freedom
func ChiSquaredTest(observed, expected []float64) (chiSquared float64, df int) {
	if len(observed) != len(expected) {
		return 0, 0
	}

	chiSquared = 0.0
	for i := range observed {
		if expected[i] > 0 {
			diff := observed[i] - expected[i]
			chiSquared += (diff * diff) / expected[i]
		}
	}

	df = len(observed) - 1
	return chiSquared, df
}

// MeanAbsoluteError computes MAE between two samples
func MeanAbsoluteError(measured, expected []float64) float64 {
	if len(measured) != len(expected) || len(measured) == 0 {
		return 0
	}

	sum := 0.0
	for i := range measured {
		sum += math.Abs(measured[i] - expected[i])
	}
	return sum / float64(len(measured))
}

// RootMeanSquaredError computes RMSE between two samples
func RootMeanSquaredError(measured, expected []float64) float64 {
	if len(measured) != len(expected) || len(measured) == 0 {
		return 0
	}

	sum := 0.0
	for i := range measured {
		diff := measured[i] - expected[i]
		sum += diff * diff
	}
	return math.Sqrt(sum / float64(len(measured)))
}

// Mean computes the arithmetic mean of a sample
func Mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// StandardDeviation computes the sample standard deviation
func StandardDeviation(values []float64) float64 {
	if len(values) <= 1 {
		return 0
	}

	mean := Mean(values)
	sumSq := 0.0
	for _, v := range values {
		diff := v - mean
		sumSq += diff * diff
	}
	return math.Sqrt(sumSq / float64(len(values)-1))
}

// RelativeError computes the relative error between measured and expected
func RelativeError(measured, expected float64) float64 {
	if expected == 0 {
		return 0
	}
	return math.Abs((measured - expected) / expected)
}

// WithinTolerance checks if measured value is within tolerance of expected
func WithinTolerance(measured, expected, tolerancePct float64) bool {
	relErr := RelativeError(measured, expected)
	return relErr <= tolerancePct/100.0
}
