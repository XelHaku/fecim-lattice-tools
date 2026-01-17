// Package layers provides neural network layer implementations for crossbar-based CIM.
// batch.go implements batch processing utilities for efficient neural network training
// and inference on crossbar arrays.
//
// CIM-specific considerations:
// - Batching improves crossbar utilization
// - Multiple inputs can share weight reads
// - Batch normalization statistics require batch processing
// - Gradient averaging over batch for stable training

package layers

import (
	"math"
	"math/rand"
	"sort"
	"sync"
)

// Batch represents a batch of samples
type Batch struct {
	Data   [][][]float64 // [batch_size][height][width] for images
	Labels []int         // [batch_size] for classification
	Size   int
}

// NewBatch creates a new batch from data and labels
func NewBatch(data [][][]float64, labels []int) *Batch {
	return &Batch{
		Data:   data,
		Labels: labels,
		Size:   len(data),
	}
}

// FlatBatch represents a batch of flattened samples
type FlatBatch struct {
	Data   [][]float64 // [batch_size][features]
	Labels []int       // [batch_size]
	Size   int
}

// NewFlatBatch creates a new flat batch
func NewFlatBatch(data [][]float64, labels []int) *FlatBatch {
	return &FlatBatch{
		Data:   data,
		Labels: labels,
		Size:   len(data),
	}
}

// Flatten converts Batch to FlatBatch
func (b *Batch) Flatten() *FlatBatch {
	if b.Size == 0 {
		return &FlatBatch{Data: [][]float64{}, Labels: []int{}, Size: 0}
	}

	h := len(b.Data[0])
	w := 0
	if h > 0 {
		w = len(b.Data[0][0])
	}
	features := h * w

	flatData := make([][]float64, b.Size)
	for i := 0; i < b.Size; i++ {
		flatData[i] = make([]float64, features)
		idx := 0
		for row := 0; row < h; row++ {
			for col := 0; col < w; col++ {
				flatData[i][idx] = b.Data[i][row][col]
				idx++
			}
		}
	}

	return &FlatBatch{
		Data:   flatData,
		Labels: b.Labels,
		Size:   b.Size,
	}
}

// DataLoader handles batching and shuffling of datasets
type DataLoader struct {
	Data       [][][]float64 // All data samples
	Labels     []int         // All labels
	BatchSize  int
	Shuffle    bool
	DropLast   bool  // Drop last incomplete batch
	indices    []int // Current epoch indices
	position   int   // Current position in indices
	NumSamples int
	NumBatches int
	mu         sync.Mutex
}

// NewDataLoader creates a new data loader
func NewDataLoader(data [][][]float64, labels []int, batchSize int, shuffle, dropLast bool) *DataLoader {
	numSamples := len(data)
	numBatches := numSamples / batchSize
	if !dropLast && numSamples%batchSize > 0 {
		numBatches++
	}

	dl := &DataLoader{
		Data:       data,
		Labels:     labels,
		BatchSize:  batchSize,
		Shuffle:    shuffle,
		DropLast:   dropLast,
		NumSamples: numSamples,
		NumBatches: numBatches,
	}
	dl.Reset()
	return dl
}

// Reset resets the data loader for a new epoch
func (dl *DataLoader) Reset() {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	dl.indices = make([]int, dl.NumSamples)
	for i := 0; i < dl.NumSamples; i++ {
		dl.indices[i] = i
	}

	if dl.Shuffle {
		rand.Shuffle(len(dl.indices), func(i, j int) {
			dl.indices[i], dl.indices[j] = dl.indices[j], dl.indices[i]
		})
	}

	dl.position = 0
}

// Next returns the next batch, or nil if no more batches
func (dl *DataLoader) Next() *Batch {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	if dl.position >= len(dl.indices) {
		return nil
	}

	endPos := dl.position + dl.BatchSize
	if endPos > len(dl.indices) {
		if dl.DropLast {
			return nil
		}
		endPos = len(dl.indices)
	}

	batchIndices := dl.indices[dl.position:endPos]
	dl.position = endPos

	batchData := make([][][]float64, len(batchIndices))
	batchLabels := make([]int, len(batchIndices))

	for i, idx := range batchIndices {
		batchData[i] = dl.Data[idx]
		batchLabels[i] = dl.Labels[idx]
	}

	return NewBatch(batchData, batchLabels)
}

// HasNext returns true if there are more batches
func (dl *DataLoader) HasNext() bool {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	if dl.position >= len(dl.indices) {
		return false
	}
	remaining := len(dl.indices) - dl.position
	if dl.DropLast && remaining < dl.BatchSize {
		return false
	}
	return true
}

// FlatDataLoader handles flattened data
type FlatDataLoader struct {
	Data       [][]float64
	Labels     []int
	BatchSize  int
	Shuffle    bool
	DropLast   bool
	indices    []int
	position   int
	NumSamples int
	NumBatches int
	mu         sync.Mutex
}

// NewFlatDataLoader creates a new flat data loader
func NewFlatDataLoader(data [][]float64, labels []int, batchSize int, shuffle, dropLast bool) *FlatDataLoader {
	numSamples := len(data)
	numBatches := numSamples / batchSize
	if !dropLast && numSamples%batchSize > 0 {
		numBatches++
	}

	dl := &FlatDataLoader{
		Data:       data,
		Labels:     labels,
		BatchSize:  batchSize,
		Shuffle:    shuffle,
		DropLast:   dropLast,
		NumSamples: numSamples,
		NumBatches: numBatches,
	}
	dl.Reset()
	return dl
}

// Reset resets the loader for a new epoch
func (dl *FlatDataLoader) Reset() {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	dl.indices = make([]int, dl.NumSamples)
	for i := 0; i < dl.NumSamples; i++ {
		dl.indices[i] = i
	}

	if dl.Shuffle {
		rand.Shuffle(len(dl.indices), func(i, j int) {
			dl.indices[i], dl.indices[j] = dl.indices[j], dl.indices[i]
		})
	}

	dl.position = 0
}

// Next returns the next batch
func (dl *FlatDataLoader) Next() *FlatBatch {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	if dl.position >= len(dl.indices) {
		return nil
	}

	endPos := dl.position + dl.BatchSize
	if endPos > len(dl.indices) {
		if dl.DropLast {
			return nil
		}
		endPos = len(dl.indices)
	}

	batchIndices := dl.indices[dl.position:endPos]
	dl.position = endPos

	batchData := make([][]float64, len(batchIndices))
	batchLabels := make([]int, len(batchIndices))

	for i, idx := range batchIndices {
		batchData[i] = dl.Data[idx]
		batchLabels[i] = dl.Labels[idx]
	}

	return NewFlatBatch(batchData, batchLabels)
}

// HasNext returns true if there are more batches
func (dl *FlatDataLoader) HasNext() bool {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	if dl.position >= len(dl.indices) {
		return false
	}
	remaining := len(dl.indices) - dl.position
	if dl.DropLast && remaining < dl.BatchSize {
		return false
	}
	return true
}

// ============================================================================
// Batch Operations
// ============================================================================

// BatchMean computes mean across batch dimension
func BatchMean(batch [][]float64) []float64 {
	if len(batch) == 0 {
		return nil
	}
	features := len(batch[0])
	mean := make([]float64, features)

	for _, sample := range batch {
		for j, val := range sample {
			mean[j] += val
		}
	}

	n := float64(len(batch))
	for j := range mean {
		mean[j] /= n
	}
	return mean
}

// BatchVar computes variance across batch dimension
func BatchVar(batch [][]float64, mean []float64) []float64 {
	if len(batch) == 0 {
		return nil
	}
	features := len(batch[0])
	variance := make([]float64, features)

	for _, sample := range batch {
		for j, val := range sample {
			diff := val - mean[j]
			variance[j] += diff * diff
		}
	}

	n := float64(len(batch))
	for j := range variance {
		variance[j] /= n
	}
	return variance
}

// BatchStd computes standard deviation across batch dimension
func BatchStd(batch [][]float64, mean []float64) []float64 {
	variance := BatchVar(batch, mean)
	std := make([]float64, len(variance))
	for i, v := range variance {
		std[i] = math.Sqrt(v)
	}
	return std
}

// BatchNormalize normalizes a batch to zero mean and unit variance
func BatchNormalize(batch [][]float64, eps float64) ([][]float64, []float64, []float64) {
	mean := BatchMean(batch)
	std := BatchStd(batch, mean)

	normalized := make([][]float64, len(batch))
	for i, sample := range batch {
		normalized[i] = make([]float64, len(sample))
		for j, val := range sample {
			normalized[i][j] = (val - mean[j]) / (std[j] + eps)
		}
	}

	return normalized, mean, std
}

// BatchAdd adds a vector to each sample in batch
func BatchAdd(batch [][]float64, vec []float64) [][]float64 {
	result := make([][]float64, len(batch))
	for i, sample := range batch {
		result[i] = make([]float64, len(sample))
		for j, val := range sample {
			result[i][j] = val + vec[j]
		}
	}
	return result
}

// BatchMul multiplies each sample by a vector (element-wise)
func BatchMul(batch [][]float64, vec []float64) [][]float64 {
	result := make([][]float64, len(batch))
	for i, sample := range batch {
		result[i] = make([]float64, len(sample))
		for j, val := range sample {
			result[i][j] = val * vec[j]
		}
	}
	return result
}

// BatchScale scales all values in batch by a scalar
func BatchScale(batch [][]float64, scalar float64) [][]float64 {
	result := make([][]float64, len(batch))
	for i, sample := range batch {
		result[i] = make([]float64, len(sample))
		for j, val := range sample {
			result[i][j] = val * scalar
		}
	}
	return result
}

// ============================================================================
// Batch Matrix Operations (for Crossbar)
// ============================================================================

// BatchMVM performs matrix-vector multiply for a batch
// Each sample is multiplied by the weight matrix
func BatchMVM(batch [][]float64, weights [][]float64) [][]float64 {
	if len(batch) == 0 || len(weights) == 0 {
		return nil
	}

	outFeatures := len(weights)
	inFeatures := len(weights[0])

	results := make([][]float64, len(batch))
	for i, sample := range batch {
		results[i] = make([]float64, outFeatures)
		for j := 0; j < outFeatures; j++ {
			sum := 0.0
			for k := 0; k < inFeatures && k < len(sample); k++ {
				sum += weights[j][k] * sample[k]
			}
			results[i][j] = sum
		}
	}
	return results
}

// BatchMVMParallel performs parallel batch MVM using goroutines
func BatchMVMParallel(batch [][]float64, weights [][]float64, numWorkers int) [][]float64 {
	if len(batch) == 0 || len(weights) == 0 {
		return nil
	}

	results := make([][]float64, len(batch))
	var wg sync.WaitGroup

	batchesPerWorker := (len(batch) + numWorkers - 1) / numWorkers

	for w := 0; w < numWorkers; w++ {
		start := w * batchesPerWorker
		end := start + batchesPerWorker
		if end > len(batch) {
			end = len(batch)
		}
		if start >= end {
			continue
		}

		wg.Add(1)
		go func(startIdx, endIdx int) {
			defer wg.Done()
			outFeatures := len(weights)
			inFeatures := len(weights[0])

			for i := startIdx; i < endIdx; i++ {
				results[i] = make([]float64, outFeatures)
				sample := batch[i]
				for j := 0; j < outFeatures; j++ {
					sum := 0.0
					for k := 0; k < inFeatures && k < len(sample); k++ {
						sum += weights[j][k] * sample[k]
					}
					results[i][j] = sum
				}
			}
		}(start, end)
	}

	wg.Wait()
	return results
}

// BatchMVMWithNoise performs batch MVM with crossbar noise simulation
func BatchMVMWithNoise(batch [][]float64, weights [][]float64, noiseStd float64) [][]float64 {
	if len(batch) == 0 || len(weights) == 0 {
		return nil
	}

	outFeatures := len(weights)
	inFeatures := len(weights[0])

	results := make([][]float64, len(batch))
	for i, sample := range batch {
		results[i] = make([]float64, outFeatures)
		for j := 0; j < outFeatures; j++ {
			sum := 0.0
			for k := 0; k < inFeatures && k < len(sample); k++ {
				// Add per-element noise (simulates crossbar variability)
				noise := noiseStd * rand.NormFloat64()
				sum += (weights[j][k] + noise) * sample[k]
			}
			// Add output noise (ADC noise)
			results[i][j] = sum + noiseStd*rand.NormFloat64()
		}
	}
	return results
}

// ============================================================================
// Batch Gradient Operations
// ============================================================================

// BatchGradientSum sums gradients across batch
func BatchGradientSum(gradients [][]float64) []float64 {
	if len(gradients) == 0 {
		return nil
	}
	features := len(gradients[0])
	sum := make([]float64, features)

	for _, grad := range gradients {
		for j, val := range grad {
			sum[j] += val
		}
	}
	return sum
}

// BatchGradientMean averages gradients across batch
func BatchGradientMean(gradients [][]float64) []float64 {
	sum := BatchGradientSum(gradients)
	n := float64(len(gradients))
	for i := range sum {
		sum[i] /= n
	}
	return sum
}

// ComputeWeightGradients computes weight gradients from batch
// dL/dW = (1/N) * sum(dL/dy * x^T)
func ComputeWeightGradients(inputs, outputGrads [][]float64) [][]float64 {
	if len(inputs) == 0 || len(outputGrads) == 0 {
		return nil
	}

	outFeatures := len(outputGrads[0])
	inFeatures := len(inputs[0])
	batchSize := len(inputs)

	gradients := make([][]float64, outFeatures)
	for i := 0; i < outFeatures; i++ {
		gradients[i] = make([]float64, inFeatures)
		for j := 0; j < inFeatures; j++ {
			sum := 0.0
			for b := 0; b < batchSize; b++ {
				sum += outputGrads[b][i] * inputs[b][j]
			}
			gradients[i][j] = sum / float64(batchSize)
		}
	}
	return gradients
}

// ComputeInputGradients computes input gradients for backprop
// dL/dx = W^T * dL/dy
func ComputeInputGradients(outputGrads, weights [][]float64) [][]float64 {
	if len(outputGrads) == 0 || len(weights) == 0 {
		return nil
	}

	outFeatures := len(weights)
	inFeatures := len(weights[0])
	batchSize := len(outputGrads)

	inputGrads := make([][]float64, batchSize)
	for b := 0; b < batchSize; b++ {
		inputGrads[b] = make([]float64, inFeatures)
		for j := 0; j < inFeatures; j++ {
			sum := 0.0
			for i := 0; i < outFeatures; i++ {
				sum += weights[i][j] * outputGrads[b][i]
			}
			inputGrads[b][j] = sum
		}
	}
	return inputGrads
}

// ============================================================================
// Batch Sampling
// ============================================================================

// StratifiedSampler ensures class balance in batches
type StratifiedSampler struct {
	ClassIndices map[int][]int
	Classes      []int
	NumClasses   int
}

// NewStratifiedSampler creates a stratified sampler
func NewStratifiedSampler(labels []int) *StratifiedSampler {
	classIndices := make(map[int][]int)
	for i, label := range labels {
		classIndices[label] = append(classIndices[label], i)
	}

	classes := make([]int, 0, len(classIndices))
	for c := range classIndices {
		classes = append(classes, c)
	}
	sort.Ints(classes)

	return &StratifiedSampler{
		ClassIndices: classIndices,
		Classes:      classes,
		NumClasses:   len(classes),
	}
}

// Sample returns a stratified sample of indices
func (ss *StratifiedSampler) Sample(samplesPerClass int) []int {
	indices := make([]int, 0, samplesPerClass*ss.NumClasses)

	for _, class := range ss.Classes {
		classIdx := ss.ClassIndices[class]
		n := len(classIdx)
		if n == 0 {
			continue
		}

		// Sample with replacement if needed
		for i := 0; i < samplesPerClass; i++ {
			idx := classIdx[rand.Intn(n)]
			indices = append(indices, idx)
		}
	}

	// Shuffle the result
	rand.Shuffle(len(indices), func(i, j int) {
		indices[i], indices[j] = indices[j], indices[i]
	})

	return indices
}

// WeightedSampler samples based on weights
type WeightedSampler struct {
	Weights    []float64
	CumWeights []float64
	Total      float64
}

// NewWeightedSampler creates a weighted sampler
func NewWeightedSampler(weights []float64) *WeightedSampler {
	cumWeights := make([]float64, len(weights))
	total := 0.0
	for i, w := range weights {
		total += w
		cumWeights[i] = total
	}

	return &WeightedSampler{
		Weights:    weights,
		CumWeights: cumWeights,
		Total:      total,
	}
}

// Sample returns n weighted random indices
func (ws *WeightedSampler) Sample(n int) []int {
	indices := make([]int, n)
	for i := 0; i < n; i++ {
		r := rand.Float64() * ws.Total
		// Binary search for the index
		idx := sort.Search(len(ws.CumWeights), func(j int) bool {
			return ws.CumWeights[j] > r
		})
		if idx >= len(ws.Weights) {
			idx = len(ws.Weights) - 1
		}
		indices[i] = idx
	}
	return indices
}

// ============================================================================
// Batch Collation
// ============================================================================

// CollateFn is a function that collates samples into a batch
type CollateFn func(samples [][][]float64, labels []int) (*Batch, error)

// DefaultCollate is the default collation function
func DefaultCollate(samples [][][]float64, labels []int) (*Batch, error) {
	return NewBatch(samples, labels), nil
}

// PaddedCollate pads samples to the same size
func PaddedCollate(samples [][][]float64, labels []int, padValue float64) (*Batch, error) {
	if len(samples) == 0 {
		return NewBatch(samples, labels), nil
	}

	// Find max dimensions
	maxH, maxW := 0, 0
	for _, sample := range samples {
		if len(sample) > maxH {
			maxH = len(sample)
		}
		for _, row := range sample {
			if len(row) > maxW {
				maxW = len(row)
			}
		}
	}

	// Pad samples
	paddedSamples := make([][][]float64, len(samples))
	for i, sample := range samples {
		paddedSamples[i] = make([][]float64, maxH)
		for j := 0; j < maxH; j++ {
			paddedSamples[i][j] = make([]float64, maxW)
			for k := 0; k < maxW; k++ {
				if j < len(sample) && k < len(sample[j]) {
					paddedSamples[i][j][k] = sample[j][k]
				} else {
					paddedSamples[i][j][k] = padValue
				}
			}
		}
	}

	return NewBatch(paddedSamples, labels), nil
}

// ============================================================================
// Batch Utilities
// ============================================================================

// SplitBatch splits a batch into smaller chunks
func SplitBatch(batch *Batch, chunkSize int) []*Batch {
	numChunks := (batch.Size + chunkSize - 1) / chunkSize
	chunks := make([]*Batch, 0, numChunks)

	for i := 0; i < batch.Size; i += chunkSize {
		end := i + chunkSize
		if end > batch.Size {
			end = batch.Size
		}

		chunkData := batch.Data[i:end]
		chunkLabels := batch.Labels[i:end]
		chunks = append(chunks, NewBatch(chunkData, chunkLabels))
	}

	return chunks
}

// ConcatBatches concatenates multiple batches
func ConcatBatches(batches ...*Batch) *Batch {
	totalSize := 0
	for _, b := range batches {
		totalSize += b.Size
	}

	data := make([][][]float64, 0, totalSize)
	labels := make([]int, 0, totalSize)

	for _, b := range batches {
		data = append(data, b.Data...)
		labels = append(labels, b.Labels...)
	}

	return NewBatch(data, labels)
}

// ShuffleBatch shuffles the samples in a batch
func ShuffleBatch(batch *Batch) *Batch {
	indices := make([]int, batch.Size)
	for i := range indices {
		indices[i] = i
	}
	rand.Shuffle(len(indices), func(i, j int) {
		indices[i], indices[j] = indices[j], indices[i]
	})

	shuffledData := make([][][]float64, batch.Size)
	shuffledLabels := make([]int, batch.Size)
	for i, idx := range indices {
		shuffledData[i] = batch.Data[idx]
		shuffledLabels[i] = batch.Labels[idx]
	}

	return NewBatch(shuffledData, shuffledLabels)
}

// FilterBatch filters samples based on a predicate
func FilterBatch(batch *Batch, predicate func(sample [][]float64, label int) bool) *Batch {
	filteredData := make([][][]float64, 0)
	filteredLabels := make([]int, 0)

	for i := 0; i < batch.Size; i++ {
		if predicate(batch.Data[i], batch.Labels[i]) {
			filteredData = append(filteredData, batch.Data[i])
			filteredLabels = append(filteredLabels, batch.Labels[i])
		}
	}

	return NewBatch(filteredData, filteredLabels)
}

// MapBatch applies a function to each sample
func MapBatch(batch *Batch, fn func(sample [][]float64) [][]float64) *Batch {
	mappedData := make([][][]float64, batch.Size)
	for i := 0; i < batch.Size; i++ {
		mappedData[i] = fn(batch.Data[i])
	}
	return NewBatch(mappedData, batch.Labels)
}

// ============================================================================
// Crossbar-Aware Batch Processing
// ============================================================================

// CrossbarBatchConfig configures crossbar-aware batch processing
type CrossbarBatchConfig struct {
	ArraySize    int     // Crossbar array size (e.g., 64)
	TileSize     int     // For tiled operations
	NoiseLevel   float64 // Simulation noise
	QuantBits    int     // Input quantization bits
	ParallelRows int     // Number of rows to activate in parallel
}

// DefaultCrossbarBatchConfig returns default config
func DefaultCrossbarBatchConfig() *CrossbarBatchConfig {
	return &CrossbarBatchConfig{
		ArraySize:    64,
		TileSize:     64,
		NoiseLevel:   0.02,
		QuantBits:    8,
		ParallelRows: 1,
	}
}

// TileBatch splits batch operations for tiled crossbar execution
func TileBatch(batch [][]float64, weights [][]float64, config *CrossbarBatchConfig) [][]float64 {
	if len(batch) == 0 || len(weights) == 0 {
		return nil
	}

	outFeatures := len(weights)
	inFeatures := len(weights[0])

	// Calculate number of tiles needed
	numInTiles := (inFeatures + config.TileSize - 1) / config.TileSize
	numOutTiles := (outFeatures + config.TileSize - 1) / config.TileSize

	results := make([][]float64, len(batch))
	for b := range results {
		results[b] = make([]float64, outFeatures)
	}

	// Process each output tile
	for outTile := 0; outTile < numOutTiles; outTile++ {
		outStart := outTile * config.TileSize
		outEnd := outStart + config.TileSize
		if outEnd > outFeatures {
			outEnd = outFeatures
		}

		// Process each input tile
		for inTile := 0; inTile < numInTiles; inTile++ {
			inStart := inTile * config.TileSize
			inEnd := inStart + config.TileSize
			if inEnd > inFeatures {
				inEnd = inFeatures
			}

			// Perform tiled MVM
			for b, sample := range batch {
				for o := outStart; o < outEnd; o++ {
					for i := inStart; i < inEnd; i++ {
						if i < len(sample) {
							// Add noise per element
							noise := config.NoiseLevel * rand.NormFloat64()
							results[b][o] += (weights[o][i] + noise) * sample[i]
						}
					}
				}
			}
		}
	}

	return results
}

// QuantizeBatchInputs quantizes batch inputs to specified bits
func QuantizeBatchInputs(batch [][]float64, bits int) [][]float64 {
	levels := float64(int(1) << bits)

	quantized := make([][]float64, len(batch))
	for i, sample := range batch {
		quantized[i] = make([]float64, len(sample))
		for j, val := range sample {
			// Assume input in [0, 1]
			level := math.Round(val * (levels - 1))
			if level < 0 {
				level = 0
			}
			if level >= levels {
				level = levels - 1
			}
			quantized[i][j] = level / (levels - 1)
		}
	}
	return quantized
}

// BatchAccuracyTopK computes top-k accuracy for a batch
func BatchAccuracyTopK(outputs [][]float64, labels []int, k int) float64 {
	if len(outputs) == 0 || len(labels) == 0 {
		return 0.0
	}

	correct := 0
	for i, output := range outputs {
		if i >= len(labels) {
			break
		}

		// Get top-k predictions
		type indexedValue struct {
			idx int
			val float64
		}
		indexed := make([]indexedValue, len(output))
		for j, val := range output {
			indexed[j] = indexedValue{j, val}
		}
		sort.Slice(indexed, func(a, b int) bool {
			return indexed[a].val > indexed[b].val
		})

		// Check if true label is in top-k
		trueLabel := labels[i]
		for j := 0; j < k && j < len(indexed); j++ {
			if indexed[j].idx == trueLabel {
				correct++
				break
			}
		}
	}

	return float64(correct) / float64(len(outputs))
}

// BatchConfusionMatrix computes confusion matrix for batch predictions
func BatchConfusionMatrix(outputs [][]float64, labels []int, numClasses int) [][]int {
	confusion := make([][]int, numClasses)
	for i := range confusion {
		confusion[i] = make([]int, numClasses)
	}

	for i, output := range outputs {
		if i >= len(labels) {
			break
		}

		// Get predicted class
		predicted := 0
		maxVal := output[0]
		for j, val := range output {
			if val > maxVal {
				maxVal = val
				predicted = j
			}
		}

		trueClass := labels[i]
		if trueClass >= 0 && trueClass < numClasses && predicted >= 0 && predicted < numClasses {
			confusion[trueClass][predicted]++
		}
	}

	return confusion
}
