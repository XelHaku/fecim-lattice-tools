package mnistcli

import (
	"fmt"
	"math"
	"math/rand"
	"strings"

	"fecim-lattice-tools/module3-mnist/pkg/training"
)

func showImage(img []float64) {
	fmt.Println("\nInput Image:")
	for row := 0; row < 28; row++ {
		fmt.Print("  ")
		for col := 0; col < 28; col++ {
			val := img[row*28+col]
			if val > 0.75 {
				fmt.Print("██")
			} else if val > 0.5 {
				fmt.Print("▓▓")
			} else if val > 0.25 {
				fmt.Print("░░")
			} else {
				fmt.Print("  ")
			}
		}
		fmt.Println()
	}
}

func showPrediction(net *training.MNISTNetwork, img []float64, pred int, conf float64) {
	fmt.Println("\n=== Crossbar Inference Result ===")

	// Show layer-by-layer activations
	showLayerActivations(net, img)

	// Show output probabilities
	probs := net.GetOutputProbabilities(img)
	fmt.Println("\nOutput probabilities (softmax):")
	for i := 0; i < 10; i++ {
		barLen := int(probs[i] * 40)
		bar := strings.Repeat("█", barLen) + strings.Repeat("░", 40-barLen)
		marker := " "
		if i == pred {
			marker = "→"
		}
		fmt.Printf("  %s %d: %s %.1f%%\n", marker, i, bar, probs[i]*100)
	}

	fmt.Printf("\nPredicted digit: %d (confidence: %.1f%%)\n", pred, conf*100)
}

func showLayerActivations(net *training.MNISTNetwork, img []float64) {
	input, hidden, output := net.GetLayerActivations(img)

	fmt.Println("\n─── Layer-by-Layer Activations ───")

	// Input layer summary
	activePixels := 0
	for _, v := range input {
		if v > 0.5 {
			activePixels++
		}
	}
	fmt.Printf("\nInput Layer (784 pixels):\n")
	fmt.Printf("  Active pixels: %d / 784 (%.1f%%)\n", activePixels, float64(activePixels)/784*100)

	// Hidden layer visualization (show first 64 neurons)
	fmt.Printf("\nHidden Layer Activations (%d neurons):\n", len(hidden))
	fmt.Print("  ")
	maxShow := min(64, len(hidden))
	for i := 0; i < maxShow; i++ {
		char := activationToChar(hidden[i])
		fmt.Print(char)
		if (i+1)%32 == 0 {
			fmt.Println()
			if i < maxShow-1 {
				fmt.Print("  ")
			}
		}
	}
	if len(hidden) > maxShow {
		fmt.Printf("... (%d more)\n", len(hidden)-maxShow)
	}

	// Stats
	activeHidden := 0
	maxAct := 0.0
	sumAct := 0.0
	for _, h := range hidden {
		if h > 0.1 {
			activeHidden++
		}
		if h > maxAct {
			maxAct = h
		}
		sumAct += h
	}
	fmt.Printf("  Active neurons: %d / %d (%.1f%%)\n", activeHidden, len(hidden), float64(activeHidden)/float64(len(hidden))*100)
	fmt.Printf("  Max activation: %.3f, Mean: %.3f\n", maxAct, sumAct/float64(len(hidden)))

	// Output layer raw values (before softmax, for insight)
	fmt.Println("\nOutput Layer (10 classes):")
	fmt.Print("  Pre-softmax: ")
	for i := 0; i < 10; i++ {
		// Compute pre-softmax value
		rawVal := 0.0
		for j := range hidden {
			rawVal += hidden[j] * 0.5 // Simplified
		}
		fmt.Printf("%d:%.2f ", i, output[i])
	}
	fmt.Println()
}

func activationToChar(value float64) string {
	if value < 0.1 {
		return "·"
	} else if value < 0.3 {
		return "░"
	} else if value < 0.5 {
		return "▒"
	} else if value < 0.7 {
		return "▓"
	}
	return "█"
}

func matrix10ToSlice(matrix [10][10]int) [][]int {
	out := make([][]int, 10)
	for i := 0; i < 10; i++ {
		out[i] = make([]int, 10)
		for j := 0; j < 10; j++ {
			out[i][j] = matrix[i][j]
		}
	}
	return out
}

func array10ToSlice(v [10]float64) []float64 {
	out := make([]float64, 10)
	for i := 0; i < 10; i++ {
		out[i] = v[i]
	}
	return out
}

func showConfusionMatrix(matrix [][]int) {
	fmt.Println("\n═══════════════════════════════════════════════════")
	fmt.Println("              Confusion Matrix")
	fmt.Println("═══════════════════════════════════════════════════")

	// Header
	fmt.Print("       Predicted\n")
	fmt.Print("       ")
	for i := 0; i < 10; i++ {
		fmt.Printf("%4d", i)
	}
	fmt.Println("  │ Total")
	fmt.Println("      +" + strings.Repeat("────", 10) + "──┼──────")

	// Matrix rows
	totalCorrect := 0
	totalSamples := 0

	for i := 0; i < 10; i++ {
		rowTotal := 0
		for j := 0; j < 10; j++ {
			rowTotal += matrix[i][j]
		}
		totalSamples += rowTotal

		if i == 5 {
			fmt.Printf("A %d │ ", i)
		} else {
			fmt.Printf("  %d │ ", i)
		}

		for j := 0; j < 10; j++ {
			val := matrix[i][j]
			if i == j {
				// Diagonal (correct predictions)
				fmt.Printf("\033[92m%4d\033[0m", val)
				totalCorrect += val
			} else if val > rowTotal/10 {
				// High error
				fmt.Printf("\033[91m%4d\033[0m", val)
			} else if val > 0 {
				// Some error
				fmt.Printf("\033[93m%4d\033[0m", val)
			} else {
				fmt.Printf("%4d", val)
			}
		}
		fmt.Printf("  │ %4d\n", rowTotal)
	}

	fmt.Println("      +" + strings.Repeat("────", 10) + "──┼──────")

	// Column totals
	fmt.Print("Total ")
	for j := 0; j < 10; j++ {
		colTotal := 0
		for i := 0; i < 10; i++ {
			colTotal += matrix[i][j]
		}
		fmt.Printf("%4d", colTotal)
	}
	fmt.Printf("  │ %4d\n", totalSamples)

	// Summary
	accuracy := float64(totalCorrect) / float64(totalSamples) * 100
	fmt.Printf("\nOverall Accuracy: %.1f%% (%d/%d correct)\n", accuracy, totalCorrect, totalSamples)

	// Legend
	fmt.Println("\nLegend: \033[92m■ Correct\033[0m  \033[93m■ Minor error\033[0m  \033[91m■ Major error\033[0m")
}

func showPerClassMetrics(precision, recall, f1 []float64) {
	fmt.Println("\n─── Per-Class Performance Metrics ───")
	fmt.Println("Class  Precision  Recall    F1-Score")
	fmt.Println("─────  ─────────  ────────  ────────")

	for i := 0; i < 10; i++ {
		pBar := strings.Repeat("█", int(precision[i]*10)) + strings.Repeat("░", 10-int(precision[i]*10))
		rBar := strings.Repeat("█", int(recall[i]*10)) + strings.Repeat("░", 10-int(recall[i]*10))

		fmt.Printf("  %d    %s %.1f%%  %s %.1f%%  %.3f\n",
			i, pBar, precision[i]*100, rBar, recall[i]*100, f1[i])
	}

	// Macro averages
	var avgP, avgR, avgF1 float64
	for i := 0; i < 10; i++ {
		avgP += precision[i]
		avgR += recall[i]
		avgF1 += f1[i]
	}
	avgP /= 10
	avgR /= 10
	avgF1 /= 10

	fmt.Println("─────────────────────────────────────")
	fmt.Printf("Macro  %-23s %.1f%%  %-11s %.1f%%  %.3f\n", "", avgP*100, "", avgR*100, avgF1)
}

func createSampleDigit(digit int) []float64 {
	img := make([]float64, 784)

	switch digit {
	case 0:
		// Draw a 0
		for row := 6; row < 24; row++ {
			for col := 8; col < 20; col++ {
				// Circle outline
				dr := float64(row - 14)
				dc := float64(col - 14)
				dist := math.Sqrt(dr*dr + dc*dc)
				if dist > 5 && dist < 9 {
					img[row*28+col] = 1.0
				}
			}
		}
	case 1:
		// Draw a 1
		for row := 6; row < 24; row++ {
			img[row*28+14] = 1.0
			img[row*28+15] = 1.0
		}
		// Top serif
		img[6*28+12] = 1.0
		img[6*28+13] = 1.0
		// Bottom line
		for col := 11; col < 19; col++ {
			img[23*28+col] = 1.0
		}
	case 2:
		// Draw a 2
		for col := 8; col < 20; col++ {
			img[6*28+col] = 1.0  // Top
			img[14*28+col] = 1.0 // Middle
			img[23*28+col] = 1.0 // Bottom
		}
		for row := 6; row < 14; row++ {
			img[row*28+19] = 1.0 // Right top
		}
		for row := 14; row < 24; row++ {
			img[row*28+8] = 1.0 // Left bottom
		}
	case 3:
		// Draw a 3
		for col := 8; col < 20; col++ {
			img[6*28+col] = 1.0  // Top
			img[14*28+col] = 1.0 // Middle
			img[23*28+col] = 1.0 // Bottom
		}
		for row := 6; row < 24; row++ {
			img[row*28+19] = 1.0 // Right side
		}
	case 4:
		// Draw a 4
		for row := 6; row < 15; row++ {
			img[row*28+8] = 1.0 // Left top
		}
		for col := 8; col < 20; col++ {
			img[14*28+col] = 1.0 // Middle
		}
		for row := 6; row < 24; row++ {
			img[row*28+16] = 1.0 // Right
		}
	case 5:
		// Draw a 5
		for col := 8; col < 20; col++ {
			img[6*28+col] = 1.0  // Top
			img[14*28+col] = 1.0 // Middle
			img[23*28+col] = 1.0 // Bottom
		}
		for row := 6; row < 14; row++ {
			img[row*28+8] = 1.0 // Left top
		}
		for row := 14; row < 24; row++ {
			img[row*28+19] = 1.0 // Right bottom
		}
	case 6:
		// Draw a 6
		for col := 8; col < 20; col++ {
			img[6*28+col] = 1.0  // Top
			img[14*28+col] = 1.0 // Middle
			img[23*28+col] = 1.0 // Bottom
		}
		for row := 6; row < 24; row++ {
			img[row*28+8] = 1.0 // Left
		}
		for row := 14; row < 24; row++ {
			img[row*28+19] = 1.0 // Right bottom
		}
	case 7:
		// Draw a 7
		for col := 8; col < 22; col++ {
			for row := 4; row < 7; row++ {
				img[row*28+col] = 1.0
			}
		}
		for i := 0; i < 20; i++ {
			row := 6 + i
			col := 20 - i/2
			if row < 28 && col >= 0 && col < 28 {
				img[row*28+col] = 1.0
				if col > 0 {
					img[row*28+col-1] = 1.0
				}
			}
		}
	case 8:
		// Draw an 8
		for col := 8; col < 20; col++ {
			img[6*28+col] = 1.0  // Top
			img[14*28+col] = 1.0 // Middle
			img[23*28+col] = 1.0 // Bottom
		}
		for row := 6; row < 24; row++ {
			img[row*28+8] = 1.0  // Left
			img[row*28+19] = 1.0 // Right
		}
	case 9:
		// Draw a 9
		for col := 8; col < 20; col++ {
			img[6*28+col] = 1.0  // Top
			img[14*28+col] = 1.0 // Middle
			img[23*28+col] = 1.0 // Bottom
		}
		for row := 6; row < 14; row++ {
			img[row*28+8] = 1.0 // Left top
		}
		for row := 6; row < 24; row++ {
			img[row*28+19] = 1.0 // Right
		}
	}

	return img
}

func generateSyntheticData(n int) ([][]float64, []int) {
	images := make([][]float64, n)
	labels := make([]int, n)

	for i := 0; i < n; i++ {
		digit := rand.Intn(10)
		labels[i] = digit
		images[i] = createSampleDigit(digit)

		// Add some noise for variation
		for j := range images[i] {
			if rand.Float64() < 0.05 {
				images[i][j] = 1.0 - images[i][j]
			}
		}
	}

	return images, labels
}

func checkMark(correct bool) string {
	if correct {
		return "✓"
	}
	return "✗"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
