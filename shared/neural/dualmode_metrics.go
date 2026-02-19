package neural

// ModeMetrics holds confusion matrix and derived metrics for one inference mode.
type ModeMetrics struct {
	Confusion [10][10]int
	Accuracy  float64
	Precision [10]float64
	Recall    [10]float64
	F1        [10]float64
}

// DualModeDatasetMetrics captures FP vs CIM dataset-level behavior.
type DualModeDatasetMetrics struct {
	FP        ModeMetrics
	CIM       ModeMetrics
	Agreement float64
	AvgKL     float64
	AvgEnergy float64 // μJ per inference
	Samples   int
}

// EvaluateDualModeDataset runs inference over a dataset and returns FP/CIM confusion matrices
// plus shared dual-mode metrics (agreement, KL, energy).
func EvaluateDualModeDataset(net *DualModeNetwork, images [][]float64, labels []int) DualModeDatasetMetrics {
	metrics := DualModeDatasetMetrics{}
	if net == nil {
		return metrics
	}

	var agreeCount, fpCorrect, cimCorrect int
	var totalKL, totalEnergy float64

	maxN := len(images)
	if len(labels) < maxN {
		maxN = len(labels)
	}

	for i := 0; i < maxN; i++ {
		label := labels[i]
		if label < 0 || label > 9 {
			continue
		}

		result := net.Infer(images[i])
		if result == nil {
			continue
		}
		if result.FPPrediction < 0 || result.FPPrediction > 9 || result.CIMPrediction < 0 || result.CIMPrediction > 9 {
			continue
		}

		metrics.Samples++
		metrics.FP.Confusion[label][result.FPPrediction]++
		metrics.CIM.Confusion[label][result.CIMPrediction]++

		if result.FPPrediction == label {
			fpCorrect++
		}
		if result.CIMPrediction == label {
			cimCorrect++
		}
		if result.Agree {
			agreeCount++
		}
		totalKL += result.Disagreement
		totalEnergy += result.EnergyUsed
	}

	if metrics.Samples == 0 {
		return metrics
	}

	den := float64(metrics.Samples)
	metrics.FP.Accuracy = float64(fpCorrect) / den
	metrics.CIM.Accuracy = float64(cimCorrect) / den
	metrics.Agreement = float64(agreeCount) / den
	metrics.AvgKL = totalKL / den
	metrics.AvgEnergy = totalEnergy / den

	metrics.FP.Precision, metrics.FP.Recall, metrics.FP.F1 = computeModePRF1(metrics.FP.Confusion)
	metrics.CIM.Precision, metrics.CIM.Recall, metrics.CIM.F1 = computeModePRF1(metrics.CIM.Confusion)
	return metrics
}

func computeModePRF1(conf [10][10]int) (precision, recall, f1 [10]float64) {
	for class := 0; class < 10; class++ {
		tp := float64(conf[class][class])
		fp := 0.0
		fn := 0.0
		for i := 0; i < 10; i++ {
			if i != class {
				fp += float64(conf[i][class])
				fn += float64(conf[class][i])
			}
		}
		if tp+fp > 0 {
			precision[class] = tp / (tp + fp)
		}
		if tp+fn > 0 {
			recall[class] = tp / (tp + fn)
		}
		if precision[class]+recall[class] > 0 {
			f1[class] = 2 * precision[class] * recall[class] / (precision[class] + recall[class])
		}
	}
	return
}
