// Package layers provides pre-trained model zoo and end-to-end inference demonstration.
//
// This module implements:
// - Pre-trained model definitions for MNIST and CIFAR-10/100
// - Model architectures: LeNet, VGG, ResNet, MobileNet
// - Weight initialization and loading utilities
// - End-to-end inference pipeline with CIM mapping
// - Visualization: activation heatmaps, crossbar state, CAM
// - Benchmark suite for accuracy and performance evaluation
//
// Based on:
// - MemTorch: Memristive Deep Learning Simulation (arXiv 2407.13410)
// - PyTorch CIFAR Models (chenyaofo/pytorch-cifar-models)
// - IBM 64-core PCM inference chip
// - Dual-domain CIM system (Nature Electronics 2024)
package layers

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
	"sort"
)

// =============================================================================
// Part 1: Dataset Definitions
// =============================================================================

// DatasetType represents supported datasets
type DatasetType string

const (
	DatasetMNIST     DatasetType = "MNIST"
	DatasetCIFAR10   DatasetType = "CIFAR10"
	DatasetCIFAR100  DatasetType = "CIFAR100"
	DatasetImageNet  DatasetType = "ImageNet"
)

// DatasetInfo contains dataset metadata
type DatasetInfo struct {
	Name        DatasetType
	NumClasses  int
	InputHeight int
	InputWidth  int
	NumChannels int
	TrainSize   int
	TestSize    int
	ClassNames  []string
}

// GetDatasetInfo returns metadata for a dataset
func GetDatasetInfo(dataset DatasetType) DatasetInfo {
	switch dataset {
	case DatasetMNIST:
		return DatasetInfo{
			Name:        DatasetMNIST,
			NumClasses:  10,
			InputHeight: 28,
			InputWidth:  28,
			NumChannels: 1,
			TrainSize:   60000,
			TestSize:    10000,
			ClassNames:  []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"},
		}
	case DatasetCIFAR10:
		return DatasetInfo{
			Name:        DatasetCIFAR10,
			NumClasses:  10,
			InputHeight: 32,
			InputWidth:  32,
			NumChannels: 3,
			TrainSize:   50000,
			TestSize:    10000,
			ClassNames: []string{
				"airplane", "automobile", "bird", "cat", "deer",
				"dog", "frog", "horse", "ship", "truck",
			},
		}
	case DatasetCIFAR100:
		return DatasetInfo{
			Name:        DatasetCIFAR100,
			NumClasses:  100,
			InputHeight: 32,
			InputWidth:  32,
			NumChannels: 3,
			TrainSize:   50000,
			TestSize:    10000,
			ClassNames:  generateCIFAR100Classes(),
		}
	case DatasetImageNet:
		return DatasetInfo{
			Name:        DatasetImageNet,
			NumClasses:  1000,
			InputHeight: 224,
			InputWidth:  224,
			NumChannels: 3,
			TrainSize:   1281167,
			TestSize:    50000,
		}
	default:
		return DatasetInfo{}
	}
}

func generateCIFAR100Classes() []string {
	// Simplified: return class indices as strings
	classes := make([]string, 100)
	for i := range classes {
		classes[i] = fmt.Sprintf("class_%d", i)
	}
	return classes
}

// Sample represents a single data sample
type Sample struct {
	Data     []float64 // Flattened input data
	Label    int       // Ground truth label
	Height   int
	Width    int
	Channels int
}

// GenerateSyntheticSample creates synthetic data for testing
func GenerateSyntheticSample(dataset DatasetType, rng *rand.Rand) Sample {
	info := GetDatasetInfo(dataset)
	size := info.InputHeight * info.InputWidth * info.NumChannels
	data := make([]float64, size)
	for i := range data {
		data[i] = rng.Float64()*2 - 1 // [-1, 1]
	}
	return Sample{
		Data:     data,
		Label:    rng.Intn(info.NumClasses),
		Height:   info.InputHeight,
		Width:    info.InputWidth,
		Channels: info.NumChannels,
	}
}

// =============================================================================
// Part 2: Model Architecture Definitions
// =============================================================================

// ModelArchitecture represents supported model architectures
type ModelArchitecture string

const (
	ArchLeNet5       ModelArchitecture = "LeNet5"
	ArchVGG11        ModelArchitecture = "VGG11"
	ArchVGG16        ModelArchitecture = "VGG16"
	ArchResNet18     ModelArchitecture = "ResNet18"
	ArchResNet32     ModelArchitecture = "ResNet32"
	ArchMobileNetV2  ModelArchitecture = "MobileNetV2"
	ArchSimpleMLP    ModelArchitecture = "SimpleMLP"
)

// LayerConfig defines a single layer configuration
type LayerConfig struct {
	Type       string // "conv", "fc", "pool", "bn", "relu", "dropout"
	InChannels int
	OutChannels int
	KernelSize int
	Stride     int
	Padding    int
	Activation string // "relu", "sigmoid", "none"
}

// ModelConfig defines complete model configuration
type ModelConfig struct {
	Architecture ModelArchitecture
	Dataset      DatasetType
	Layers       []LayerConfig
	NumParams    int
	FLOPs        int
	Accuracy     float64 // Expected accuracy on test set
}

// GetModelConfig returns configuration for a model-dataset pair
func GetModelConfig(arch ModelArchitecture, dataset DatasetType) ModelConfig {
	info := GetDatasetInfo(dataset)

	switch arch {
	case ArchLeNet5:
		return ModelConfig{
			Architecture: ArchLeNet5,
			Dataset:      dataset,
			Layers: []LayerConfig{
				{Type: "conv", InChannels: info.NumChannels, OutChannels: 6, KernelSize: 5, Stride: 1, Padding: 0, Activation: "relu"},
				{Type: "pool", KernelSize: 2, Stride: 2},
				{Type: "conv", InChannels: 6, OutChannels: 16, KernelSize: 5, Stride: 1, Padding: 0, Activation: "relu"},
				{Type: "pool", KernelSize: 2, Stride: 2},
				{Type: "fc", InChannels: 400, OutChannels: 120, Activation: "relu"},
				{Type: "fc", InChannels: 120, OutChannels: 84, Activation: "relu"},
				{Type: "fc", InChannels: 84, OutChannels: info.NumClasses, Activation: "none"},
			},
			NumParams: 61706,
			FLOPs:     416000,
			Accuracy:  99.0, // MNIST
		}
	case ArchVGG11:
		return ModelConfig{
			Architecture: ArchVGG11,
			Dataset:      dataset,
			Layers:       buildVGG11Layers(info),
			NumParams:    9756426,
			FLOPs:        153000000,
			Accuracy:     92.0, // CIFAR-10
		}
	case ArchVGG16:
		return ModelConfig{
			Architecture: ArchVGG16,
			Dataset:      dataset,
			Layers:       buildVGG16Layers(info),
			NumParams:    14728266,
			FLOPs:        313000000,
			Accuracy:     93.5, // CIFAR-10
		}
	case ArchResNet18:
		return ModelConfig{
			Architecture: ArchResNet18,
			Dataset:      dataset,
			Layers:       buildResNet18Layers(info),
			NumParams:    11173962,
			FLOPs:        556000000,
			Accuracy:     93.0, // CIFAR-10
		}
	case ArchResNet32:
		return ModelConfig{
			Architecture: ArchResNet32,
			Dataset:      dataset,
			Layers:       buildResNet32Layers(info),
			NumParams:    470000,
			FLOPs:        70000000,
			Accuracy:     92.5, // CIFAR-10
		}
	case ArchMobileNetV2:
		return ModelConfig{
			Architecture: ArchMobileNetV2,
			Dataset:      dataset,
			Layers:       buildMobileNetV2Layers(info),
			NumParams:    2236682,
			FLOPs:        91000000,
			Accuracy:     91.5, // CIFAR-10
		}
	case ArchSimpleMLP:
		inputSize := info.InputHeight * info.InputWidth * info.NumChannels
		return ModelConfig{
			Architecture: ArchSimpleMLP,
			Dataset:      dataset,
			Layers: []LayerConfig{
				{Type: "fc", InChannels: inputSize, OutChannels: 512, Activation: "relu"},
				{Type: "fc", InChannels: 512, OutChannels: 256, Activation: "relu"},
				{Type: "fc", InChannels: 256, OutChannels: info.NumClasses, Activation: "none"},
			},
			NumParams: inputSize*512 + 512*256 + 256*info.NumClasses,
			FLOPs:     inputSize*512*2 + 512*256*2 + 256*info.NumClasses*2,
			Accuracy:  98.0, // MNIST
		}
	default:
		return ModelConfig{}
	}
}

func buildVGG11Layers(info DatasetInfo) []LayerConfig {
	return []LayerConfig{
		{Type: "conv", InChannels: info.NumChannels, OutChannels: 64, KernelSize: 3, Stride: 1, Padding: 1, Activation: "relu"},
		{Type: "pool", KernelSize: 2, Stride: 2},
		{Type: "conv", InChannels: 64, OutChannels: 128, KernelSize: 3, Stride: 1, Padding: 1, Activation: "relu"},
		{Type: "pool", KernelSize: 2, Stride: 2},
		{Type: "conv", InChannels: 128, OutChannels: 256, KernelSize: 3, Stride: 1, Padding: 1, Activation: "relu"},
		{Type: "conv", InChannels: 256, OutChannels: 256, KernelSize: 3, Stride: 1, Padding: 1, Activation: "relu"},
		{Type: "pool", KernelSize: 2, Stride: 2},
		{Type: "conv", InChannels: 256, OutChannels: 512, KernelSize: 3, Stride: 1, Padding: 1, Activation: "relu"},
		{Type: "conv", InChannels: 512, OutChannels: 512, KernelSize: 3, Stride: 1, Padding: 1, Activation: "relu"},
		{Type: "pool", KernelSize: 2, Stride: 2},
		{Type: "conv", InChannels: 512, OutChannels: 512, KernelSize: 3, Stride: 1, Padding: 1, Activation: "relu"},
		{Type: "conv", InChannels: 512, OutChannels: 512, KernelSize: 3, Stride: 1, Padding: 1, Activation: "relu"},
		{Type: "pool", KernelSize: 2, Stride: 2},
		{Type: "fc", InChannels: 512, OutChannels: 512, Activation: "relu"},
		{Type: "fc", InChannels: 512, OutChannels: info.NumClasses, Activation: "none"},
	}
}

func buildVGG16Layers(info DatasetInfo) []LayerConfig {
	return []LayerConfig{
		{Type: "conv", InChannels: info.NumChannels, OutChannels: 64, KernelSize: 3, Stride: 1, Padding: 1, Activation: "relu"},
		{Type: "conv", InChannels: 64, OutChannels: 64, KernelSize: 3, Stride: 1, Padding: 1, Activation: "relu"},
		{Type: "pool", KernelSize: 2, Stride: 2},
		{Type: "conv", InChannels: 64, OutChannels: 128, KernelSize: 3, Stride: 1, Padding: 1, Activation: "relu"},
		{Type: "conv", InChannels: 128, OutChannels: 128, KernelSize: 3, Stride: 1, Padding: 1, Activation: "relu"},
		{Type: "pool", KernelSize: 2, Stride: 2},
		{Type: "conv", InChannels: 128, OutChannels: 256, KernelSize: 3, Stride: 1, Padding: 1, Activation: "relu"},
		{Type: "conv", InChannels: 256, OutChannels: 256, KernelSize: 3, Stride: 1, Padding: 1, Activation: "relu"},
		{Type: "conv", InChannels: 256, OutChannels: 256, KernelSize: 3, Stride: 1, Padding: 1, Activation: "relu"},
		{Type: "pool", KernelSize: 2, Stride: 2},
		{Type: "conv", InChannels: 256, OutChannels: 512, KernelSize: 3, Stride: 1, Padding: 1, Activation: "relu"},
		{Type: "conv", InChannels: 512, OutChannels: 512, KernelSize: 3, Stride: 1, Padding: 1, Activation: "relu"},
		{Type: "conv", InChannels: 512, OutChannels: 512, KernelSize: 3, Stride: 1, Padding: 1, Activation: "relu"},
		{Type: "pool", KernelSize: 2, Stride: 2},
		{Type: "conv", InChannels: 512, OutChannels: 512, KernelSize: 3, Stride: 1, Padding: 1, Activation: "relu"},
		{Type: "conv", InChannels: 512, OutChannels: 512, KernelSize: 3, Stride: 1, Padding: 1, Activation: "relu"},
		{Type: "conv", InChannels: 512, OutChannels: 512, KernelSize: 3, Stride: 1, Padding: 1, Activation: "relu"},
		{Type: "pool", KernelSize: 2, Stride: 2},
		{Type: "fc", InChannels: 512, OutChannels: 4096, Activation: "relu"},
		{Type: "fc", InChannels: 4096, OutChannels: 4096, Activation: "relu"},
		{Type: "fc", InChannels: 4096, OutChannels: info.NumClasses, Activation: "none"},
	}
}

func buildResNet18Layers(info DatasetInfo) []LayerConfig {
	// Simplified ResNet-18 structure
	return []LayerConfig{
		{Type: "conv", InChannels: info.NumChannels, OutChannels: 64, KernelSize: 7, Stride: 2, Padding: 3, Activation: "relu"},
		{Type: "pool", KernelSize: 3, Stride: 2},
		// Block 1
		{Type: "conv", InChannels: 64, OutChannels: 64, KernelSize: 3, Stride: 1, Padding: 1, Activation: "relu"},
		{Type: "conv", InChannels: 64, OutChannels: 64, KernelSize: 3, Stride: 1, Padding: 1, Activation: "relu"},
		// Block 2
		{Type: "conv", InChannels: 64, OutChannels: 128, KernelSize: 3, Stride: 2, Padding: 1, Activation: "relu"},
		{Type: "conv", InChannels: 128, OutChannels: 128, KernelSize: 3, Stride: 1, Padding: 1, Activation: "relu"},
		// Block 3
		{Type: "conv", InChannels: 128, OutChannels: 256, KernelSize: 3, Stride: 2, Padding: 1, Activation: "relu"},
		{Type: "conv", InChannels: 256, OutChannels: 256, KernelSize: 3, Stride: 1, Padding: 1, Activation: "relu"},
		// Block 4
		{Type: "conv", InChannels: 256, OutChannels: 512, KernelSize: 3, Stride: 2, Padding: 1, Activation: "relu"},
		{Type: "conv", InChannels: 512, OutChannels: 512, KernelSize: 3, Stride: 1, Padding: 1, Activation: "relu"},
		// Classifier
		{Type: "pool", KernelSize: 7, Stride: 1}, // Global average pool
		{Type: "fc", InChannels: 512, OutChannels: info.NumClasses, Activation: "none"},
	}
}

func buildResNet32Layers(info DatasetInfo) []LayerConfig {
	// CIFAR-specific ResNet-32
	layers := []LayerConfig{
		{Type: "conv", InChannels: info.NumChannels, OutChannels: 16, KernelSize: 3, Stride: 1, Padding: 1, Activation: "relu"},
	}
	// 5 blocks of 2 convs each for each stage (16, 32, 64 channels)
	channels := []int{16, 32, 64}
	for i, ch := range channels {
		stride := 1
		if i > 0 {
			stride = 2
		}
		inCh := 16
		if i > 0 {
			inCh = channels[i-1]
		}
		for j := 0; j < 5; j++ {
			s := 1
			if j == 0 {
				s = stride
			}
			layers = append(layers, LayerConfig{Type: "conv", InChannels: inCh, OutChannels: ch, KernelSize: 3, Stride: s, Padding: 1, Activation: "relu"})
			layers = append(layers, LayerConfig{Type: "conv", InChannels: ch, OutChannels: ch, KernelSize: 3, Stride: 1, Padding: 1, Activation: "relu"})
			inCh = ch
		}
	}
	layers = append(layers, LayerConfig{Type: "pool", KernelSize: 8, Stride: 1})
	layers = append(layers, LayerConfig{Type: "fc", InChannels: 64, OutChannels: info.NumClasses, Activation: "none"})
	return layers
}

func buildMobileNetV2Layers(info DatasetInfo) []LayerConfig {
	// Simplified MobileNetV2
	return []LayerConfig{
		{Type: "conv", InChannels: info.NumChannels, OutChannels: 32, KernelSize: 3, Stride: 2, Padding: 1, Activation: "relu"},
		// Inverted residual blocks (simplified)
		{Type: "conv", InChannels: 32, OutChannels: 16, KernelSize: 1, Stride: 1, Padding: 0, Activation: "relu"},
		{Type: "conv", InChannels: 16, OutChannels: 24, KernelSize: 3, Stride: 2, Padding: 1, Activation: "relu"},
		{Type: "conv", InChannels: 24, OutChannels: 32, KernelSize: 3, Stride: 2, Padding: 1, Activation: "relu"},
		{Type: "conv", InChannels: 32, OutChannels: 64, KernelSize: 3, Stride: 2, Padding: 1, Activation: "relu"},
		{Type: "conv", InChannels: 64, OutChannels: 96, KernelSize: 3, Stride: 1, Padding: 1, Activation: "relu"},
		{Type: "conv", InChannels: 96, OutChannels: 160, KernelSize: 3, Stride: 2, Padding: 1, Activation: "relu"},
		{Type: "conv", InChannels: 160, OutChannels: 320, KernelSize: 3, Stride: 1, Padding: 1, Activation: "relu"},
		{Type: "conv", InChannels: 320, OutChannels: 1280, KernelSize: 1, Stride: 1, Padding: 0, Activation: "relu"},
		{Type: "pool", KernelSize: 7, Stride: 1},
		{Type: "fc", InChannels: 1280, OutChannels: info.NumClasses, Activation: "none"},
	}
}

// =============================================================================
// Part 3: Pre-trained Model Zoo
// =============================================================================

// PretrainedModel represents a pre-trained model with weights
type PretrainedModel struct {
	Config      ModelConfig
	Weights     map[string][][]float64 // Layer name -> weight matrix
	Biases      map[string][]float64   // Layer name -> bias vector
	BNParams    map[string]BNParameters // BatchNorm parameters
	Quantized   bool
	QuantBits   int
	CIMReady    bool // Weights formatted for CIM mapping
}

// BNParameters holds BatchNorm parameters
type BNParameters struct {
	Gamma    []float64
	Beta     []float64
	Mean     []float64
	Variance []float64
}

// ModelZoo manages pre-trained models
type ModelZoo struct {
	Models      map[string]*PretrainedModel
	BasePath    string
	Initialized bool
}

// NewModelZoo creates a new model zoo
func NewModelZoo(basePath string) *ModelZoo {
	return &ModelZoo{
		Models:   make(map[string]*PretrainedModel),
		BasePath: basePath,
	}
}

// RegisterModel adds a model to the zoo
func (mz *ModelZoo) RegisterModel(name string, model *PretrainedModel) {
	mz.Models[name] = model
}

// GetModel retrieves a model by name
func (mz *ModelZoo) GetModel(name string) (*PretrainedModel, bool) {
	model, exists := mz.Models[name]
	return model, exists
}

// InitializeWithRandomWeights creates a model with random weights
func (mz *ModelZoo) InitializeWithRandomWeights(arch ModelArchitecture, dataset DatasetType) *PretrainedModel {
	config := GetModelConfig(arch, dataset)
	rng := rand.New(rand.NewSource(42))

	model := &PretrainedModel{
		Config:   config,
		Weights:  make(map[string][][]float64),
		Biases:   make(map[string][]float64),
		BNParams: make(map[string]BNParameters),
	}

	for i, layer := range config.Layers {
		layerName := fmt.Sprintf("layer_%d", i)

		switch layer.Type {
		case "conv":
			// Conv weights: [OutChannels, InChannels * K * K]
			fanIn := layer.InChannels * layer.KernelSize * layer.KernelSize
			fanOut := layer.OutChannels
			std := math.Sqrt(2.0 / float64(fanIn)) // He initialization

			weights := make([][]float64, fanOut)
			for j := range weights {
				weights[j] = make([]float64, fanIn)
				for k := range weights[j] {
					weights[j][k] = rng.NormFloat64() * std
				}
			}
			model.Weights[layerName] = weights

			// Bias
			bias := make([]float64, fanOut)
			model.Biases[layerName] = bias

		case "fc":
			// FC weights: [OutChannels, InChannels]
			fanIn := layer.InChannels
			fanOut := layer.OutChannels
			std := math.Sqrt(2.0 / float64(fanIn))

			weights := make([][]float64, fanOut)
			for j := range weights {
				weights[j] = make([]float64, fanIn)
				for k := range weights[j] {
					weights[j][k] = rng.NormFloat64() * std
				}
			}
			model.Weights[layerName] = weights

			bias := make([]float64, fanOut)
			model.Biases[layerName] = bias
		}
	}

	name := fmt.Sprintf("%s_%s", arch, dataset)
	mz.RegisterModel(name, model)

	return model
}

// ListModels returns all registered model names
func (mz *ModelZoo) ListModels() []string {
	names := make([]string, 0, len(mz.Models))
	for name := range mz.Models {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// =============================================================================
// Part 4: Model Quantization for CIM
// =============================================================================

// QuantizationConfig configures model quantization
type QuantizationConfig struct {
	WeightBits     int     // Bits for weights
	ActivationBits int     // Bits for activations
	SymmetricQuant bool    // Use symmetric quantization
	PerChannel     bool    // Per-channel vs per-tensor
	CalibrationSamples int // Number of calibration samples
}

// DefaultQuantizationConfig returns typical CIM quantization settings
func DefaultQuantizationConfig() QuantizationConfig {
	return QuantizationConfig{
		WeightBits:         6,
		ActivationBits:     8,
		SymmetricQuant:     true,
		PerChannel:         true,
		CalibrationSamples: 1000,
	}
}

// QuantizationScale stores quantization parameters
type QuantizationScale struct {
	Scale     float64
	ZeroPoint int
	MinVal    float64
	MaxVal    float64
}

// ModelQuantizer handles model quantization
type ModelQuantizer struct {
	Config QuantizationConfig
	Scales map[string]QuantizationScale
}

// NewModelQuantizer creates a model quantizer
func NewModelQuantizer(config QuantizationConfig) *ModelQuantizer {
	return &ModelQuantizer{
		Config: config,
		Scales: make(map[string]QuantizationScale),
	}
}

// QuantizeModel quantizes a pre-trained model
func (mq *ModelQuantizer) QuantizeModel(model *PretrainedModel) *PretrainedModel {
	quantized := &PretrainedModel{
		Config:    model.Config,
		Weights:   make(map[string][][]float64),
		Biases:    make(map[string][]float64),
		BNParams:  model.BNParams,
		Quantized: true,
		QuantBits: mq.Config.WeightBits,
	}

	levels := float64(1 << mq.Config.WeightBits)

	for name, weights := range model.Weights {
		// Find min/max
		minVal, maxVal := weights[0][0], weights[0][0]
		for _, row := range weights {
			for _, w := range row {
				if w < minVal {
					minVal = w
				}
				if w > maxVal {
					maxVal = w
				}
			}
		}

		// Compute scale
		var scale QuantizationScale
		if mq.Config.SymmetricQuant {
			absMax := math.Max(math.Abs(minVal), math.Abs(maxVal))
			scale.Scale = absMax / (levels / 2)
			scale.ZeroPoint = 0
			scale.MinVal = -absMax
			scale.MaxVal = absMax
		} else {
			scale.Scale = (maxVal - minVal) / (levels - 1)
			scale.ZeroPoint = int(-minVal / scale.Scale)
			scale.MinVal = minVal
			scale.MaxVal = maxVal
		}
		if scale.Scale == 0 {
			scale.Scale = 1
		}
		mq.Scales[name] = scale

		// Quantize weights
		quantWeights := make([][]float64, len(weights))
		for i, row := range weights {
			quantWeights[i] = make([]float64, len(row))
			for j, w := range row {
				// Quantize and dequantize
				q := math.Round((w - scale.MinVal) / scale.Scale)
				q = math.Max(0, math.Min(levels-1, q))
				quantWeights[i][j] = q*scale.Scale + scale.MinVal
			}
		}
		quantized.Weights[name] = quantWeights
	}

	// Copy biases (typically kept in higher precision)
	for name, bias := range model.Biases {
		quantized.Biases[name] = make([]float64, len(bias))
		copy(quantized.Biases[name], bias)
	}

	return quantized
}

// =============================================================================
// Part 5: End-to-End Inference Pipeline
// =============================================================================

// InferencePipeline orchestrates end-to-end inference
type InferencePipeline struct {
	Model           *PretrainedModel
	CIMSimulator    *FASTSimulator
	Activations     map[string][]float64 // Intermediate activations for visualization
	LastPrediction  int
	LastConfidences []float64
	InferenceCount  int
	TotalLatencyUs  float64
	TotalEnergyFJ   float64
}

// NewInferencePipeline creates an inference pipeline
func NewInferencePipeline(model *PretrainedModel) *InferencePipeline {
	// Create FAST simulator for CIM inference
	fastConfig := FASTConfig{
		CrossbarRows:      64,
		CrossbarCols:      64,
		NonIdealityCfg:    DefaultNonIdealityConfig(),
		EnableSparsity:    true,
		SparsityThreshold: 0.01,
	}

	return &InferencePipeline{
		Model:        model,
		CIMSimulator: NewFASTSimulator(fastConfig),
		Activations:  make(map[string][]float64),
	}
}

// RunInference performs inference on a single sample
func (ip *InferencePipeline) RunInference(sample Sample) (int, []float64) {
	activation := sample.Data

	// Process each layer
	for i, layer := range ip.Model.Config.Layers {
		layerName := fmt.Sprintf("layer_%d", i)

		switch layer.Type {
		case "conv":
			activation = ip.convForward(layerName, activation, layer)
		case "fc":
			activation = ip.fcForward(layerName, activation, layer)
		case "pool":
			activation = ip.poolForward(activation, layer)
		}

		// Apply activation function
		if layer.Activation == "relu" {
			for j := range activation {
				if activation[j] < 0 {
					activation[j] = 0
				}
			}
		}

		// Store for visualization
		ip.Activations[layerName] = make([]float64, len(activation))
		copy(ip.Activations[layerName], activation)
	}

	// Softmax for probabilities
	confidences := softmax(activation)
	ip.LastConfidences = confidences

	// Find prediction
	predicted := 0
	maxConf := confidences[0]
	for i, c := range confidences {
		if c > maxConf {
			maxConf = c
			predicted = i
		}
	}
	ip.LastPrediction = predicted
	ip.InferenceCount++

	return predicted, confidences
}

func (ip *InferencePipeline) convForward(name string, input []float64, layer LayerConfig) []float64 {
	weights, exists := ip.Model.Weights[name]
	if !exists {
		return input
	}

	// Simplified: treat as matrix multiplication (im2col style)
	outSize := layer.OutChannels
	output := make([]float64, outSize)

	// Map to CIM and compute
	ip.CIMSimulator.MapWeights(weights)
	output = ip.CIMSimulator.SimulateInference(input[:len(weights[0])])

	// Add bias
	if bias, ok := ip.Model.Biases[name]; ok {
		for i := range output {
			if i < len(bias) {
				output[i] += bias[i]
			}
		}
	}

	return output
}

func (ip *InferencePipeline) fcForward(name string, input []float64, layer LayerConfig) []float64 {
	weights, exists := ip.Model.Weights[name]
	if !exists {
		return input
	}

	output := make([]float64, len(weights))

	// Matrix-vector multiplication
	for i := range weights {
		for j := range weights[i] {
			if j < len(input) {
				output[i] += weights[i][j] * input[j]
			}
		}
	}

	// Add bias
	if bias, ok := ip.Model.Biases[name]; ok {
		for i := range output {
			if i < len(bias) {
				output[i] += bias[i]
			}
		}
	}

	return output
}

func (ip *InferencePipeline) poolForward(input []float64, layer LayerConfig) []float64 {
	// Simplified pooling: reduce by factor of kernel^2
	poolFactor := layer.KernelSize * layer.KernelSize
	if poolFactor <= 0 {
		poolFactor = 4
	}

	outSize := len(input) / poolFactor
	if outSize < 1 {
		outSize = 1
	}

	output := make([]float64, outSize)
	for i := range output {
		start := i * poolFactor
		end := start + poolFactor
		if end > len(input) {
			end = len(input)
		}

		// Max pooling
		maxVal := input[start]
		for j := start + 1; j < end; j++ {
			if input[j] > maxVal {
				maxVal = input[j]
			}
		}
		output[i] = maxVal
	}

	return output
}

func softmax(input []float64) []float64 {
	maxVal := input[0]
	for _, v := range input {
		if v > maxVal {
			maxVal = v
		}
	}

	output := make([]float64, len(input))
	var sum float64
	for i, v := range input {
		output[i] = math.Exp(v - maxVal)
		sum += output[i]
	}
	for i := range output {
		output[i] /= sum
	}

	return output
}

// =============================================================================
// Part 6: Visualization: Activation Heatmaps and CAM
// =============================================================================

// VisualizationConfig configures visualization output
type VisualizationConfig struct {
	ColorMap    string // "jet", "viridis", "grayscale"
	Width       int
	Height      int
	ShowOverlay bool
}

// ActivationVisualizer creates visualizations of network activations
type ActivationVisualizer struct {
	Config     VisualizationConfig
	Pipeline   *InferencePipeline
}

// NewActivationVisualizer creates an activation visualizer
func NewActivationVisualizer(pipeline *InferencePipeline, config VisualizationConfig) *ActivationVisualizer {
	return &ActivationVisualizer{
		Config:   config,
		Pipeline: pipeline,
	}
}

// GenerateActivationHeatmap creates a heatmap image for a layer
func (av *ActivationVisualizer) GenerateActivationHeatmap(layerName string) *image.RGBA {
	activations, exists := av.Pipeline.Activations[layerName]
	if !exists || len(activations) == 0 {
		return nil
	}

	// Determine dimensions (try to make square-ish)
	size := len(activations)
	width := int(math.Sqrt(float64(size)))
	height := (size + width - 1) / width

	if av.Config.Width > 0 {
		width = av.Config.Width
	}
	if av.Config.Height > 0 {
		height = av.Config.Height
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Normalize activations
	minAct, maxAct := activations[0], activations[0]
	for _, a := range activations {
		if a < minAct {
			minAct = a
		}
		if a > maxAct {
			maxAct = a
		}
	}
	actRange := maxAct - minAct
	if actRange == 0 {
		actRange = 1
	}

	// Draw heatmap
	for i := 0; i < width*height && i < len(activations); i++ {
		x := i % width
		y := i / width
		normalized := (activations[i] - minAct) / actRange
		c := av.getColor(normalized)
		img.Set(x, y, c)
	}

	return img
}

// GenerateCAM creates Class Activation Map
func (av *ActivationVisualizer) GenerateCAM(sample Sample, targetClass int) []float64 {
	// Run inference to get activations
	av.Pipeline.RunInference(sample)

	// Get last conv layer activations
	var lastConvAct []float64
	var lastConvName string
	for i := len(av.Pipeline.Model.Config.Layers) - 1; i >= 0; i-- {
		layer := av.Pipeline.Model.Config.Layers[i]
		if layer.Type == "conv" {
			lastConvName = fmt.Sprintf("layer_%d", i)
			lastConvAct = av.Pipeline.Activations[lastConvName]
			break
		}
	}

	if lastConvAct == nil {
		return nil
	}

	// Get classifier weights for target class
	var classWeights []float64
	for name, weights := range av.Pipeline.Model.Weights {
		if len(weights) > targetClass {
			classWeights = weights[targetClass]
			break
		}
		_ = name
	}

	if classWeights == nil {
		return lastConvAct // Return raw activations as fallback
	}

	// Compute weighted combination
	cam := make([]float64, len(lastConvAct))
	for i := range cam {
		if i < len(classWeights) {
			cam[i] = lastConvAct[i] * classWeights[i]
		}
	}

	// ReLU
	for i := range cam {
		if cam[i] < 0 {
			cam[i] = 0
		}
	}

	return cam
}

// getColor maps a normalized value [0,1] to a color
func (av *ActivationVisualizer) getColor(value float64) color.RGBA {
	value = math.Max(0, math.Min(1, value))

	switch av.Config.ColorMap {
	case "jet":
		return jetColorMap(value)
	case "viridis":
		return viridisColorMap(value)
	default:
		// Grayscale
		g := uint8(value * 255)
		return color.RGBA{g, g, g, 255}
	}
}

func jetColorMap(v float64) color.RGBA {
	// Jet colormap: blue -> cyan -> green -> yellow -> red
	var r, g, b float64
	if v < 0.25 {
		r = 0
		g = 4 * v
		b = 1
	} else if v < 0.5 {
		r = 0
		g = 1
		b = 1 - 4*(v-0.25)
	} else if v < 0.75 {
		r = 4 * (v - 0.5)
		g = 1
		b = 0
	} else {
		r = 1
		g = 1 - 4*(v-0.75)
		b = 0
	}
	return color.RGBA{uint8(r * 255), uint8(g * 255), uint8(b * 255), 255}
}

func viridisColorMap(v float64) color.RGBA {
	// Simplified viridis: purple -> blue -> green -> yellow
	var r, g, b float64
	if v < 0.33 {
		t := v / 0.33
		r = 0.267 + t*0.015
		g = 0.004 + t*0.325
		b = 0.329 + t*0.214
	} else if v < 0.66 {
		t := (v - 0.33) / 0.33
		r = 0.282 + t*(-0.100)
		g = 0.329 + t*0.345
		b = 0.543 + t*(-0.220)
	} else {
		t := (v - 0.66) / 0.34
		r = 0.182 + t*0.811
		g = 0.674 + t*0.271
		b = 0.323 + t*(-0.278)
	}
	return color.RGBA{uint8(r * 255), uint8(g * 255), uint8(b * 255), 255}
}

// =============================================================================
// Part 7: Crossbar State Visualization
// =============================================================================

// CrossbarVisualizer visualizes CIM crossbar state
type CrossbarVisualizer struct {
	Simulator *FASTSimulator
	Config    VisualizationConfig
}

// NewCrossbarVisualizer creates a crossbar visualizer
func NewCrossbarVisualizer(sim *FASTSimulator) *CrossbarVisualizer {
	return &CrossbarVisualizer{
		Simulator: sim,
		Config: VisualizationConfig{
			ColorMap: "viridis",
			Width:    64,
			Height:   64,
		},
	}
}

// GenerateConductanceMap creates a heatmap of conductance values
func (cv *CrossbarVisualizer) GenerateConductanceMap() *image.RGBA {
	crossbar := cv.Simulator.Crossbar
	img := image.NewRGBA(image.Rect(0, 0, crossbar.Cols, crossbar.Rows))

	// Find conductance range
	minG, maxG := crossbar.GMin, crossbar.GMax

	for i := 0; i < crossbar.Rows; i++ {
		for j := 0; j < crossbar.Cols; j++ {
			g := crossbar.Cells[i][j].ActualG
			normalized := (g - minG) / (maxG - minG)
			c := viridisColorMap(normalized)

			// Mark faulty cells
			if crossbar.Cells[i][j].FaultType != NoFault {
				c = color.RGBA{255, 0, 0, 255} // Red for faults
			}

			img.Set(j, i, c)
		}
	}

	return img
}

// GenerateFaultMap creates a map showing stuck-at-faults
func (cv *CrossbarVisualizer) GenerateFaultMap() *image.RGBA {
	crossbar := cv.Simulator.Crossbar
	img := image.NewRGBA(image.Rect(0, 0, crossbar.Cols, crossbar.Rows))

	for i := 0; i < crossbar.Rows; i++ {
		for j := 0; j < crossbar.Cols; j++ {
			var c color.RGBA
			switch crossbar.Cells[i][j].FaultType {
			case NoFault:
				c = color.RGBA{0, 255, 0, 255} // Green: healthy
			case StuckAtHRS:
				c = color.RGBA{255, 0, 0, 255} // Red: stuck high
			case StuckAtLRS:
				c = color.RGBA{0, 0, 255, 255} // Blue: stuck low
			}
			img.Set(j, i, c)
		}
	}

	return img
}

// GenerateVariationMap shows D2D variation factors
func (cv *CrossbarVisualizer) GenerateVariationMap() *image.RGBA {
	crossbar := cv.Simulator.Crossbar
	img := image.NewRGBA(image.Rect(0, 0, crossbar.Cols, crossbar.Rows))

	// Find variation range
	minVar, maxVar := 1.0, 1.0
	for i := 0; i < crossbar.Rows; i++ {
		for j := 0; j < crossbar.Cols; j++ {
			v := crossbar.Cells[i][j].D2DFactor
			if v < minVar {
				minVar = v
			}
			if v > maxVar {
				maxVar = v
			}
		}
	}
	varRange := maxVar - minVar
	if varRange == 0 {
		varRange = 1
	}

	for i := 0; i < crossbar.Rows; i++ {
		for j := 0; j < crossbar.Cols; j++ {
			v := crossbar.Cells[i][j].D2DFactor
			normalized := (v - minVar) / varRange
			c := viridisColorMap(normalized)
			img.Set(j, i, c)
		}
	}

	return img
}

// =============================================================================
// Part 8: Inference Demo Runner
// =============================================================================

// DemoConfig configures the inference demonstration
type DemoConfig struct {
	Architecture ModelArchitecture
	Dataset      DatasetType
	NumSamples   int
	Quantize     bool
	QuantBits    int
	Visualize    bool
	Verbose      bool
}

// DefaultDemoConfig returns typical demo configuration
func DefaultDemoConfig() DemoConfig {
	return DemoConfig{
		Architecture: ArchLeNet5,
		Dataset:      DatasetMNIST,
		NumSamples:   100,
		Quantize:     true,
		QuantBits:    6,
		Visualize:    true,
		Verbose:      true,
	}
}

// InferenceDemo runs an end-to-end inference demonstration
type InferenceDemo struct {
	Config       DemoConfig
	ModelZoo     *ModelZoo
	Model        *PretrainedModel
	Pipeline     *InferencePipeline
	Visualizer   *ActivationVisualizer
	Results      DemoResults
}

// DemoResults stores demonstration results
type DemoResults struct {
	TotalSamples    int
	CorrectCount    int
	Accuracy        float64
	AvgLatencyUs    float64
	TotalEnergyFJ   float64
	EnergyPerInf    float64
	TopKAccuracies  map[int]float64
	ConfusionMatrix [][]int
	ClassAccuracies []float64
}

// NewInferenceDemo creates an inference demonstration
func NewInferenceDemo(config DemoConfig) *InferenceDemo {
	demo := &InferenceDemo{
		Config:   config,
		ModelZoo: NewModelZoo("./models"),
		Results: DemoResults{
			TopKAccuracies: make(map[int]float64),
		},
	}

	// Initialize model
	demo.Model = demo.ModelZoo.InitializeWithRandomWeights(config.Architecture, config.Dataset)

	// Quantize if requested
	if config.Quantize {
		quantizer := NewModelQuantizer(QuantizationConfig{
			WeightBits:     config.QuantBits,
			ActivationBits: 8,
			SymmetricQuant: true,
		})
		demo.Model = quantizer.QuantizeModel(demo.Model)
	}

	// Create pipeline
	demo.Pipeline = NewInferencePipeline(demo.Model)

	// Create visualizer
	if config.Visualize {
		demo.Visualizer = NewActivationVisualizer(demo.Pipeline, VisualizationConfig{
			ColorMap: "jet",
			Width:    64,
			Height:   64,
		})
	}

	// Initialize confusion matrix
	numClasses := GetDatasetInfo(config.Dataset).NumClasses
	demo.Results.ConfusionMatrix = make([][]int, numClasses)
	for i := range demo.Results.ConfusionMatrix {
		demo.Results.ConfusionMatrix[i] = make([]int, numClasses)
	}
	demo.Results.ClassAccuracies = make([]float64, numClasses)

	return demo
}

// Run executes the inference demonstration
func (demo *InferenceDemo) Run() DemoResults {
	rng := rand.New(rand.NewSource(42))
	datasetInfo := GetDatasetInfo(demo.Config.Dataset)

	classCounts := make([]int, datasetInfo.NumClasses)
	classCorrect := make([]int, datasetInfo.NumClasses)

	for i := 0; i < demo.Config.NumSamples; i++ {
		// Generate synthetic sample
		sample := GenerateSyntheticSample(demo.Config.Dataset, rng)

		// Run inference
		predicted, confidences := demo.Pipeline.RunInference(sample)

		// Update results
		demo.Results.TotalSamples++
		if predicted == sample.Label {
			demo.Results.CorrectCount++
			classCorrect[sample.Label]++
		}
		classCounts[sample.Label]++

		// Update confusion matrix
		if sample.Label < datasetInfo.NumClasses && predicted < datasetInfo.NumClasses {
			demo.Results.ConfusionMatrix[sample.Label][predicted]++
		}

		// Top-K accuracy
		for _, k := range []int{1, 3, 5} {
			if demo.isTopK(confidences, sample.Label, k) {
				demo.Results.TopKAccuracies[k]++
			}
		}
	}

	// Calculate final metrics
	demo.Results.Accuracy = float64(demo.Results.CorrectCount) / float64(demo.Results.TotalSamples) * 100

	for k := range demo.Results.TopKAccuracies {
		demo.Results.TopKAccuracies[k] = demo.Results.TopKAccuracies[k] / float64(demo.Results.TotalSamples) * 100
	}

	for i := range demo.Results.ClassAccuracies {
		if classCounts[i] > 0 {
			demo.Results.ClassAccuracies[i] = float64(classCorrect[i]) / float64(classCounts[i]) * 100
		}
	}

	// Energy and latency from simulator
	metrics := demo.Pipeline.CIMSimulator.GetMetrics()
	demo.Results.TotalEnergyFJ = metrics["energy_J"] * 1e15
	demo.Results.AvgLatencyUs = metrics["latency_us"] / float64(demo.Results.TotalSamples)
	demo.Results.EnergyPerInf = demo.Results.TotalEnergyFJ / float64(demo.Results.TotalSamples)

	return demo.Results
}

func (demo *InferenceDemo) isTopK(confidences []float64, label int, k int) bool {
	// Create index-value pairs and sort
	type pair struct {
		idx int
		val float64
	}
	pairs := make([]pair, len(confidences))
	for i, v := range confidences {
		pairs[i] = pair{i, v}
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].val > pairs[j].val
	})

	// Check if label is in top-k
	for i := 0; i < k && i < len(pairs); i++ {
		if pairs[i].idx == label {
			return true
		}
	}
	return false
}

// GenerateReport creates a text report of the demonstration
func (demo *InferenceDemo) GenerateReport() string {
	var report string
	report += "=== CIM Inference Demonstration Report ===\n\n"

	report += fmt.Sprintf("Model: %s\n", demo.Config.Architecture)
	report += fmt.Sprintf("Dataset: %s\n", demo.Config.Dataset)
	report += fmt.Sprintf("Quantization: %d-bit\n\n", demo.Config.QuantBits)

	report += "Performance Metrics:\n"
	report += fmt.Sprintf("  Accuracy: %.2f%%\n", demo.Results.Accuracy)
	report += fmt.Sprintf("  Top-3 Accuracy: %.2f%%\n", demo.Results.TopKAccuracies[3])
	report += fmt.Sprintf("  Top-5 Accuracy: %.2f%%\n", demo.Results.TopKAccuracies[5])
	report += fmt.Sprintf("  Avg Latency: %.2f µs\n", demo.Results.AvgLatencyUs)
	report += fmt.Sprintf("  Energy/Inference: %.2f fJ\n", demo.Results.EnergyPerInf)
	report += fmt.Sprintf("  Total Samples: %d\n\n", demo.Results.TotalSamples)

	report += "Per-Class Accuracy:\n"
	datasetInfo := GetDatasetInfo(demo.Config.Dataset)
	for i, acc := range demo.Results.ClassAccuracies {
		className := fmt.Sprintf("%d", i)
		if i < len(datasetInfo.ClassNames) {
			className = datasetInfo.ClassNames[i]
		}
		report += fmt.Sprintf("  %s: %.2f%%\n", className, acc)
	}

	return report
}

// ToJSON exports results as JSON
func (demo *InferenceDemo) ToJSON() ([]byte, error) {
	return json.MarshalIndent(demo.Results, "", "  ")
}

// =============================================================================
// Part 9: Benchmark Suite
// =============================================================================

// ModelBenchmark benchmarks multiple models
type ModelBenchmark struct {
	Models  []ModelArchitecture
	Dataset DatasetType
	Results map[string]DemoResults
}

// NewModelBenchmark creates a benchmark suite
func NewModelBenchmark(dataset DatasetType) *ModelBenchmark {
	return &ModelBenchmark{
		Models: []ModelArchitecture{
			ArchSimpleMLP,
			ArchLeNet5,
			ArchVGG11,
			ArchResNet18,
		},
		Dataset: dataset,
		Results: make(map[string]DemoResults),
	}
}

// RunAll benchmarks all models
func (mb *ModelBenchmark) RunAll(numSamples int) {
	for _, arch := range mb.Models {
		config := DemoConfig{
			Architecture: arch,
			Dataset:      mb.Dataset,
			NumSamples:   numSamples,
			Quantize:     true,
			QuantBits:    6,
			Visualize:    false,
		}

		demo := NewInferenceDemo(config)
		results := demo.Run()
		mb.Results[string(arch)] = results
	}
}

// CompareResults generates comparison report
func (mb *ModelBenchmark) CompareResults() string {
	var report string
	report += "=== Model Comparison Benchmark ===\n\n"
	report += fmt.Sprintf("Dataset: %s\n\n", mb.Dataset)

	report += fmt.Sprintf("%-15s %10s %12s %12s\n", "Model", "Accuracy", "Latency(µs)", "Energy(fJ)")
	report += fmt.Sprintf("%-15s %10s %12s %12s\n", "-----", "--------", "-----------", "----------")

	for _, arch := range mb.Models {
		if result, exists := mb.Results[string(arch)]; exists {
			report += fmt.Sprintf("%-15s %9.2f%% %12.2f %12.2f\n",
				arch, result.Accuracy, result.AvgLatencyUs, result.EnergyPerInf)
		}
	}

	return report
}

// =============================================================================
// Part 10: Export Utilities
// =============================================================================

// ModelExporter exports models for deployment
type ModelExporter struct {
	Model *PretrainedModel
}

// NewModelExporter creates a model exporter
func NewModelExporter(model *PretrainedModel) *ModelExporter {
	return &ModelExporter{Model: model}
}

// ExportToJSON exports model weights as JSON
func (me *ModelExporter) ExportToJSON() ([]byte, error) {
	export := struct {
		Architecture string                 `json:"architecture"`
		Dataset      string                 `json:"dataset"`
		Quantized    bool                   `json:"quantized"`
		QuantBits    int                    `json:"quant_bits"`
		NumParams    int                    `json:"num_params"`
		Weights      map[string][][]float64 `json:"weights"`
		Biases       map[string][]float64   `json:"biases"`
	}{
		Architecture: string(me.Model.Config.Architecture),
		Dataset:      string(me.Model.Config.Dataset),
		Quantized:    me.Model.Quantized,
		QuantBits:    me.Model.QuantBits,
		NumParams:    me.Model.Config.NumParams,
		Weights:      me.Model.Weights,
		Biases:       me.Model.Biases,
	}

	return json.MarshalIndent(export, "", "  ")
}

// ExportToCIMFormat exports weights formatted for CIM mapping
func (me *ModelExporter) ExportToCIMFormat(crossbarSize int) []CrossbarTile {
	var tiles []CrossbarTile
	tileID := 0

	for name, weights := range me.Model.Weights {
		rows := len(weights)
		cols := 0
		if rows > 0 {
			cols = len(weights[0])
		}

		// Tile the weight matrix
		rowTiles := (rows + crossbarSize - 1) / crossbarSize
		colTiles := (cols + crossbarSize - 1) / crossbarSize

		for rt := 0; rt < rowTiles; rt++ {
			for ct := 0; ct < colTiles; ct++ {
				tile := CrossbarTile{
					ID:        tileID,
					LayerName: name,
					RowOffset: rt * crossbarSize,
					ColOffset: ct * crossbarSize,
					Data:      make([][]float64, crossbarSize),
				}

				for i := 0; i < crossbarSize; i++ {
					tile.Data[i] = make([]float64, crossbarSize)
					srcRow := rt*crossbarSize + i
					if srcRow >= rows {
						continue
					}
					for j := 0; j < crossbarSize; j++ {
						srcCol := ct*crossbarSize + j
						if srcCol < cols {
							tile.Data[i][j] = weights[srcRow][srcCol]
						}
					}
				}

				tiles = append(tiles, tile)
				tileID++
			}
		}
	}

	return tiles
}

// CrossbarTile represents a weight tile for CIM mapping
type CrossbarTile struct {
	ID        int
	LayerName string
	RowOffset int
	ColOffset int
	Data      [][]float64
}
