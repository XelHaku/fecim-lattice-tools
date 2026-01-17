// Package layers provides neural network layer implementations for crossbar-based CIM.
// checkpoint.go implements model checkpointing and state management utilities.
//
// Checkpointing is essential for:
// - Saving/loading trained models
// - Resuming interrupted training
// - Model versioning and comparison
// - Deploying models to crossbar hardware
//
// CIM-specific considerations:
// - Quantized weight checkpoints for hardware deployment
// - Crossbar configuration metadata
// - Noise model parameters

package layers

import (
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// ModelCheckpoint represents a complete model state
type ModelCheckpoint struct {
	Version     string                 `json:"version"`
	Timestamp   time.Time              `json:"timestamp"`
	Epoch       int                    `json:"epoch"`
	Step        int                    `json:"step"`
	Layers      []LayerCheckpoint      `json:"layers"`
	Optimizer   *OptimizerCheckpoint   `json:"optimizer,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
	Metrics     map[string]float64     `json:"metrics,omitempty"`
	CrossbarCfg *CrossbarCheckpoint    `json:"crossbar_config,omitempty"`
}

// LayerCheckpoint represents a single layer's state
type LayerCheckpoint struct {
	Name       string             `json:"name"`
	Type       string             `json:"type"`
	Shape      []int              `json:"shape"`
	Weights    [][]float64        `json:"weights,omitempty"`
	Biases     []float64          `json:"biases,omitempty"`
	Params     map[string]float64 `json:"params,omitempty"`
	Quantized  bool               `json:"quantized"`
	QuantBits  int                `json:"quant_bits,omitempty"`
	BufferData map[string][]byte  `json:"-"` // For binary format
}

// OptimizerCheckpoint stores optimizer state
type OptimizerCheckpoint struct {
	Type      string               `json:"type"`
	LR        float64              `json:"lr"`
	Step      int                  `json:"step"`
	Momentum  [][]float64          `json:"momentum,omitempty"`
	Velocity  [][]float64          `json:"velocity,omitempty"`
	AdamState map[string][]float64 `json:"adam_state,omitempty"`
}

// CrossbarCheckpoint stores crossbar configuration
type CrossbarCheckpoint struct {
	ArraySize    int     `json:"array_size"`
	ADCBits      int     `json:"adc_bits"`
	DACBits      int     `json:"dac_bits"`
	NoiseLevel   float64 `json:"noise_level"`
	WeightBits   int     `json:"weight_bits"`
	MinCond      float64 `json:"min_conductance"`
	MaxCond      float64 `json:"max_conductance"`
	TileSize     int     `json:"tile_size"`
	NumCrossbars int     `json:"num_crossbars"`
}

// CheckpointManager handles saving and loading checkpoints
type CheckpointManager struct {
	Directory     string
	Prefix        string
	MaxCheckpoints int
	SaveOptimizer bool
	Compress      bool
}

// NewCheckpointManager creates a new checkpoint manager
func NewCheckpointManager(directory, prefix string) *CheckpointManager {
	return &CheckpointManager{
		Directory:     directory,
		Prefix:        prefix,
		MaxCheckpoints: 5, // Keep last 5 checkpoints
		SaveOptimizer: true,
		Compress:      true,
	}
}

// Save saves a model checkpoint
func (cm *CheckpointManager) Save(checkpoint *ModelCheckpoint) error {
	// Create directory if needed
	if err := os.MkdirAll(cm.Directory, 0755); err != nil {
		return fmt.Errorf("failed to create checkpoint directory: %w", err)
	}

	// Generate filename
	filename := fmt.Sprintf("%s_epoch%d_step%d.ckpt", cm.Prefix, checkpoint.Epoch, checkpoint.Step)
	if cm.Compress {
		filename += ".gz"
	}
	filepath := filepath.Join(cm.Directory, filename)

	// Create file
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create checkpoint file: %w", err)
	}
	defer file.Close()

	var writer io.Writer = file
	if cm.Compress {
		gzWriter := gzip.NewWriter(file)
		defer gzWriter.Close()
		writer = gzWriter
	}

	// Encode checkpoint
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(checkpoint); err != nil {
		return fmt.Errorf("failed to encode checkpoint: %w", err)
	}

	// Cleanup old checkpoints
	if err := cm.cleanupOldCheckpoints(); err != nil {
		// Log but don't fail
		fmt.Printf("Warning: failed to cleanup old checkpoints: %v\n", err)
	}

	return nil
}

// Load loads a checkpoint from file
func (cm *CheckpointManager) Load(filename string) (*ModelCheckpoint, error) {
	filepath := filepath.Join(cm.Directory, filename)

	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open checkpoint file: %w", err)
	}
	defer file.Close()

	var reader io.Reader = file
	if filepath[len(filepath)-3:] == ".gz" {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	var checkpoint ModelCheckpoint
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&checkpoint); err != nil {
		return nil, fmt.Errorf("failed to decode checkpoint: %w", err)
	}

	return &checkpoint, nil
}

// LoadLatest loads the most recent checkpoint
func (cm *CheckpointManager) LoadLatest() (*ModelCheckpoint, error) {
	files, err := cm.listCheckpoints()
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no checkpoints found in %s", cm.Directory)
	}

	// Files are sorted by modification time, get latest
	return cm.Load(files[len(files)-1])
}

// LoadBest loads checkpoint with best metric value
func (cm *CheckpointManager) LoadBest(metricName string, higher bool) (*ModelCheckpoint, error) {
	files, err := cm.listCheckpoints()
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no checkpoints found")
	}

	var bestCheckpoint *ModelCheckpoint
	var bestValue float64
	first := true

	for _, file := range files {
		ckpt, err := cm.Load(file)
		if err != nil {
			continue
		}

		if val, ok := ckpt.Metrics[metricName]; ok {
			if first {
				bestValue = val
				bestCheckpoint = ckpt
				first = false
			} else if (higher && val > bestValue) || (!higher && val < bestValue) {
				bestValue = val
				bestCheckpoint = ckpt
			}
		}
	}

	if bestCheckpoint == nil {
		return nil, fmt.Errorf("no checkpoint with metric %s found", metricName)
	}

	return bestCheckpoint, nil
}

// listCheckpoints returns sorted list of checkpoint files
func (cm *CheckpointManager) listCheckpoints() ([]string, error) {
	entries, err := os.ReadDir(cm.Directory)
	if err != nil {
		return nil, fmt.Errorf("failed to read checkpoint directory: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".ckpt" ||
			(len(entry.Name()) > 8 && entry.Name()[len(entry.Name())-8:] == ".ckpt.gz") {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

// cleanupOldCheckpoints removes old checkpoints beyond MaxCheckpoints
func (cm *CheckpointManager) cleanupOldCheckpoints() error {
	files, err := cm.listCheckpoints()
	if err != nil {
		return err
	}

	if len(files) <= cm.MaxCheckpoints {
		return nil
	}

	// Remove oldest checkpoints
	toRemove := files[:len(files)-cm.MaxCheckpoints]
	for _, file := range toRemove {
		path := filepath.Join(cm.Directory, file)
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("failed to remove old checkpoint %s: %w", file, err)
		}
	}

	return nil
}

// ============================================================================
// Checkpoint Creation Helpers
// ============================================================================

// NewModelCheckpoint creates a new checkpoint with basic info
func NewModelCheckpoint(epoch, step int) *ModelCheckpoint {
	return &ModelCheckpoint{
		Version:   "1.0",
		Timestamp: time.Now(),
		Epoch:     epoch,
		Step:      step,
		Layers:    make([]LayerCheckpoint, 0),
		Config:    make(map[string]interface{}),
		Metrics:   make(map[string]float64),
	}
}

// AddLayer adds a layer checkpoint
func (mc *ModelCheckpoint) AddLayer(name, layerType string, weights [][]float64, biases []float64) {
	shape := make([]int, 0)
	if len(weights) > 0 {
		shape = append(shape, len(weights))
		if len(weights[0]) > 0 {
			shape = append(shape, len(weights[0]))
		}
	}

	mc.Layers = append(mc.Layers, LayerCheckpoint{
		Name:    name,
		Type:    layerType,
		Shape:   shape,
		Weights: weights,
		Biases:  biases,
		Params:  make(map[string]float64),
	})
}

// AddQuantizedLayer adds a quantized layer checkpoint
func (mc *ModelCheckpoint) AddQuantizedLayer(name, layerType string, weights [][]float64, biases []float64, bits int) {
	shape := make([]int, 0)
	if len(weights) > 0 {
		shape = append(shape, len(weights))
		if len(weights[0]) > 0 {
			shape = append(shape, len(weights[0]))
		}
	}

	mc.Layers = append(mc.Layers, LayerCheckpoint{
		Name:      name,
		Type:      layerType,
		Shape:     shape,
		Weights:   weights,
		Biases:    biases,
		Params:    make(map[string]float64),
		Quantized: true,
		QuantBits: bits,
	})
}

// SetCrossbarConfig sets the crossbar configuration
func (mc *ModelCheckpoint) SetCrossbarConfig(arraySize, adcBits, dacBits, weightBits int, noiseLevel float64) {
	mc.CrossbarCfg = &CrossbarCheckpoint{
		ArraySize:  arraySize,
		ADCBits:    adcBits,
		DACBits:    dacBits,
		NoiseLevel: noiseLevel,
		WeightBits: weightBits,
		MinCond:    0.1,
		MaxCond:    1.0,
		TileSize:   arraySize,
	}
}

// AddMetric adds a metric value
func (mc *ModelCheckpoint) AddMetric(name string, value float64) {
	mc.Metrics[name] = value
}

// GetLayer returns a layer by name
func (mc *ModelCheckpoint) GetLayer(name string) *LayerCheckpoint {
	for i := range mc.Layers {
		if mc.Layers[i].Name == name {
			return &mc.Layers[i]
		}
	}
	return nil
}

// ============================================================================
// Binary Checkpoint Format (for large models)
// ============================================================================

// BinaryCheckpoint provides efficient storage for large models
type BinaryCheckpoint struct {
	HeaderSize  int64
	NumLayers   int
	LayerInfo   []BinaryLayerInfo
	WeightData  []byte
}

// BinaryLayerInfo describes a layer in binary format
type BinaryLayerInfo struct {
	Name       string
	Type       string
	Rows       int
	Cols       int
	BiasLen    int
	WeightOff  int64 // Offset in WeightData
	BiasOff    int64 // Offset in WeightData
	Quantized  bool
	QuantBits  int
}

// SaveBinary saves checkpoint in binary format
func SaveBinary(filename string, checkpoint *ModelCheckpoint, compress bool) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var writer io.Writer = file
	if compress {
		gzWriter := gzip.NewWriter(file)
		defer gzWriter.Close()
		writer = gzWriter
	}

	// Write header
	header := struct {
		Magic     [4]byte
		Version   uint32
		NumLayers uint32
		Epoch     uint32
		Step      uint32
	}{
		Magic:     [4]byte{'C', 'K', 'P', 'T'},
		Version:   1,
		NumLayers: uint32(len(checkpoint.Layers)),
		Epoch:     uint32(checkpoint.Epoch),
		Step:      uint32(checkpoint.Step),
	}

	if err := binary.Write(writer, binary.LittleEndian, header); err != nil {
		return err
	}

	// Write layer info
	for _, layer := range checkpoint.Layers {
		rows := 0
		cols := 0
		if len(layer.Shape) >= 1 {
			rows = layer.Shape[0]
		}
		if len(layer.Shape) >= 2 {
			cols = layer.Shape[1]
		}

		// Write layer header
		nameBytes := []byte(layer.Name)
		if err := binary.Write(writer, binary.LittleEndian, uint32(len(nameBytes))); err != nil {
			return err
		}
		if _, err := writer.Write(nameBytes); err != nil {
			return err
		}

		typeBytes := []byte(layer.Type)
		if err := binary.Write(writer, binary.LittleEndian, uint32(len(typeBytes))); err != nil {
			return err
		}
		if _, err := writer.Write(typeBytes); err != nil {
			return err
		}

		layerHeader := struct {
			Rows      uint32
			Cols      uint32
			BiasLen   uint32
			Quantized uint8
			QuantBits uint8
			Padding   [2]byte
		}{
			Rows:      uint32(rows),
			Cols:      uint32(cols),
			BiasLen:   uint32(len(layer.Biases)),
			Quantized: boolToUint8(layer.Quantized),
			QuantBits: uint8(layer.QuantBits),
		}

		if err := binary.Write(writer, binary.LittleEndian, layerHeader); err != nil {
			return err
		}

		// Write weights
		for _, row := range layer.Weights {
			for _, val := range row {
				if err := binary.Write(writer, binary.LittleEndian, val); err != nil {
					return err
				}
			}
		}

		// Write biases
		for _, val := range layer.Biases {
			if err := binary.Write(writer, binary.LittleEndian, val); err != nil {
				return err
			}
		}
	}

	return nil
}

// LoadBinary loads checkpoint from binary format
func LoadBinary(filename string) (*ModelCheckpoint, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var reader io.Reader = file
	if len(filename) > 3 && filename[len(filename)-3:] == ".gz" {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return nil, err
		}
		defer gzReader.Close()
		reader = gzReader
	}

	// Read header
	var header struct {
		Magic     [4]byte
		Version   uint32
		NumLayers uint32
		Epoch     uint32
		Step      uint32
	}

	if err := binary.Read(reader, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	if header.Magic != [4]byte{'C', 'K', 'P', 'T'} {
		return nil, fmt.Errorf("invalid checkpoint file format")
	}

	checkpoint := &ModelCheckpoint{
		Version: fmt.Sprintf("%d", header.Version),
		Epoch:   int(header.Epoch),
		Step:    int(header.Step),
		Layers:  make([]LayerCheckpoint, header.NumLayers),
		Config:  make(map[string]interface{}),
		Metrics: make(map[string]float64),
	}

	// Read layers
	for i := uint32(0); i < header.NumLayers; i++ {
		var nameLen uint32
		if err := binary.Read(reader, binary.LittleEndian, &nameLen); err != nil {
			return nil, err
		}
		nameBytes := make([]byte, nameLen)
		if _, err := io.ReadFull(reader, nameBytes); err != nil {
			return nil, err
		}

		var typeLen uint32
		if err := binary.Read(reader, binary.LittleEndian, &typeLen); err != nil {
			return nil, err
		}
		typeBytes := make([]byte, typeLen)
		if _, err := io.ReadFull(reader, typeBytes); err != nil {
			return nil, err
		}

		var layerHeader struct {
			Rows      uint32
			Cols      uint32
			BiasLen   uint32
			Quantized uint8
			QuantBits uint8
			Padding   [2]byte
		}

		if err := binary.Read(reader, binary.LittleEndian, &layerHeader); err != nil {
			return nil, err
		}

		// Read weights
		weights := make([][]float64, layerHeader.Rows)
		for r := uint32(0); r < layerHeader.Rows; r++ {
			weights[r] = make([]float64, layerHeader.Cols)
			for c := uint32(0); c < layerHeader.Cols; c++ {
				if err := binary.Read(reader, binary.LittleEndian, &weights[r][c]); err != nil {
					return nil, err
				}
			}
		}

		// Read biases
		biases := make([]float64, layerHeader.BiasLen)
		for b := uint32(0); b < layerHeader.BiasLen; b++ {
			if err := binary.Read(reader, binary.LittleEndian, &biases[b]); err != nil {
				return nil, err
			}
		}

		checkpoint.Layers[i] = LayerCheckpoint{
			Name:      string(nameBytes),
			Type:      string(typeBytes),
			Shape:     []int{int(layerHeader.Rows), int(layerHeader.Cols)},
			Weights:   weights,
			Biases:    biases,
			Quantized: layerHeader.Quantized != 0,
			QuantBits: int(layerHeader.QuantBits),
			Params:    make(map[string]float64),
		}
	}

	return checkpoint, nil
}

func boolToUint8(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}

// ============================================================================
// Checkpoint Conversion for Hardware Deployment
// ============================================================================

// ConvertToHardwareFormat converts checkpoint for crossbar deployment
func ConvertToHardwareFormat(checkpoint *ModelCheckpoint, config *CrossbarCheckpoint) (*ModelCheckpoint, error) {
	if config == nil {
		return nil, fmt.Errorf("crossbar config required")
	}

	hwCheckpoint := &ModelCheckpoint{
		Version:     checkpoint.Version,
		Timestamp:   time.Now(),
		Epoch:       checkpoint.Epoch,
		Step:        checkpoint.Step,
		Layers:      make([]LayerCheckpoint, len(checkpoint.Layers)),
		Config:      checkpoint.Config,
		Metrics:     checkpoint.Metrics,
		CrossbarCfg: config,
	}

	// Convert each layer
	for i, layer := range checkpoint.Layers {
		hwLayer := LayerCheckpoint{
			Name:      layer.Name,
			Type:      layer.Type,
			Shape:     layer.Shape,
			Params:    layer.Params,
			Quantized: true,
			QuantBits: config.WeightBits,
		}

		// Quantize weights
		if len(layer.Weights) > 0 {
			hwLayer.Weights = quantizeWeightsForHardware(layer.Weights, config)
		}

		// Quantize biases
		if len(layer.Biases) > 0 {
			hwLayer.Biases = quantizeBiasesForHardware(layer.Biases, config)
		}

		hwCheckpoint.Layers[i] = hwLayer
	}

	return hwCheckpoint, nil
}

// quantizeWeightsForHardware quantizes weights for crossbar deployment
func quantizeWeightsForHardware(weights [][]float64, config *CrossbarCheckpoint) [][]float64 {
	levels := float64(int(1) << config.WeightBits)

	// Find min/max
	minVal, maxVal := weights[0][0], weights[0][0]
	for _, row := range weights {
		for _, val := range row {
			if val < minVal {
				minVal = val
			}
			if val > maxVal {
				maxVal = val
			}
		}
	}

	// Symmetric quantization
	absMax := maxVal
	if -minVal > absMax {
		absMax = -minVal
	}
	if absMax < 1e-10 {
		absMax = 1e-10
	}

	scale := absMax / ((levels - 1) / 2)

	quantized := make([][]float64, len(weights))
	for i, row := range weights {
		quantized[i] = make([]float64, len(row))
		for j, val := range row {
			// Quantize
			level := val / scale
			if level > (levels-1)/2 {
				level = (levels - 1) / 2
			}
			if level < -(levels-1)/2 {
				level = -(levels - 1) / 2
			}
			// Map to conductance range
			normalized := (level + (levels-1)/2) / (levels - 1)
			quantized[i][j] = config.MinCond + normalized*(config.MaxCond-config.MinCond)
		}
	}

	return quantized
}

// quantizeBiasesForHardware quantizes biases
func quantizeBiasesForHardware(biases []float64, config *CrossbarCheckpoint) []float64 {
	levels := float64(int(1) << config.WeightBits)

	minVal, maxVal := biases[0], biases[0]
	for _, val := range biases {
		if val < minVal {
			minVal = val
		}
		if val > maxVal {
			maxVal = val
		}
	}

	absMax := maxVal
	if -minVal > absMax {
		absMax = -minVal
	}
	if absMax < 1e-10 {
		absMax = 1e-10
	}

	scale := absMax / ((levels - 1) / 2)

	quantized := make([]float64, len(biases))
	for i, val := range biases {
		level := val / scale
		if level > (levels-1)/2 {
			level = (levels - 1) / 2
		}
		if level < -(levels-1)/2 {
			level = -(levels - 1) / 2
		}
		quantized[i] = level * scale
	}

	return quantized
}

// ============================================================================
// Training State Management
// ============================================================================

// TrainingState tracks training progress
type TrainingState struct {
	CurrentEpoch int
	CurrentStep  int
	BestMetric   float64
	BestEpoch    int
	History      map[string][]float64
	EarlyStopping *EarlyStoppingState
}

// EarlyStoppingState tracks early stopping
type EarlyStoppingState struct {
	Patience     int
	Counter      int
	BestValue    float64
	Mode         string // "min" or "max"
	ShouldStop   bool
}

// NewTrainingState creates new training state
func NewTrainingState() *TrainingState {
	return &TrainingState{
		History: make(map[string][]float64),
	}
}

// Update updates training state with new metrics
func (ts *TrainingState) Update(epoch int, step int, metrics map[string]float64) {
	ts.CurrentEpoch = epoch
	ts.CurrentStep = step

	for name, value := range metrics {
		ts.History[name] = append(ts.History[name], value)
	}
}

// SetEarlyStopping configures early stopping
func (ts *TrainingState) SetEarlyStopping(metricName string, patience int, mode string) {
	ts.EarlyStopping = &EarlyStoppingState{
		Patience:   patience,
		Counter:    0,
		Mode:       mode,
		ShouldStop: false,
	}
}

// CheckEarlyStopping checks if training should stop
func (ts *TrainingState) CheckEarlyStopping(value float64) bool {
	if ts.EarlyStopping == nil {
		return false
	}

	es := ts.EarlyStopping

	improved := false
	if es.Mode == "min" {
		if value < es.BestValue || es.Counter == 0 {
			es.BestValue = value
			improved = true
		}
	} else {
		if value > es.BestValue || es.Counter == 0 {
			es.BestValue = value
			improved = true
		}
	}

	if improved {
		es.Counter = 0
		ts.BestMetric = value
		ts.BestEpoch = ts.CurrentEpoch
	} else {
		es.Counter++
	}

	es.ShouldStop = es.Counter >= es.Patience
	return es.ShouldStop
}

// SaveToCheckpoint adds training state to checkpoint
func (ts *TrainingState) SaveToCheckpoint(checkpoint *ModelCheckpoint) {
	checkpoint.Epoch = ts.CurrentEpoch
	checkpoint.Step = ts.CurrentStep
	checkpoint.Config["best_metric"] = ts.BestMetric
	checkpoint.Config["best_epoch"] = ts.BestEpoch
}

// LoadFromCheckpoint loads training state from checkpoint
func (ts *TrainingState) LoadFromCheckpoint(checkpoint *ModelCheckpoint) {
	ts.CurrentEpoch = checkpoint.Epoch
	ts.CurrentStep = checkpoint.Step

	if bm, ok := checkpoint.Config["best_metric"].(float64); ok {
		ts.BestMetric = bm
	}
	if be, ok := checkpoint.Config["best_epoch"].(int); ok {
		ts.BestEpoch = be
	}
}

// ============================================================================
// Model Comparison Utilities
// ============================================================================

// CompareCheckpoints compares two checkpoints
func CompareCheckpoints(a, b *ModelCheckpoint) map[string]interface{} {
	comparison := make(map[string]interface{})

	comparison["epochs"] = map[string]int{"a": a.Epoch, "b": b.Epoch}
	comparison["steps"] = map[string]int{"a": a.Step, "b": b.Step}
	comparison["num_layers"] = map[string]int{"a": len(a.Layers), "b": len(b.Layers)}

	// Compare metrics
	metricDiff := make(map[string]float64)
	for name, valA := range a.Metrics {
		if valB, ok := b.Metrics[name]; ok {
			metricDiff[name] = valB - valA
		}
	}
	comparison["metric_diff"] = metricDiff

	// Compare layer shapes
	layerDiffs := make([]map[string]interface{}, 0)
	for _, layerA := range a.Layers {
		for _, layerB := range b.Layers {
			if layerA.Name == layerB.Name {
				if !equalShapes(layerA.Shape, layerB.Shape) {
					layerDiffs = append(layerDiffs, map[string]interface{}{
						"name":    layerA.Name,
						"shape_a": layerA.Shape,
						"shape_b": layerB.Shape,
					})
				}
			}
		}
	}
	comparison["layer_shape_diffs"] = layerDiffs

	return comparison
}

func equalShapes(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
