// Package core re-exports shared/neural for backward compatibility.
// New code should import "fecim-lattice-tools/shared/neural" directly.
package core

import neural "fecim-lattice-tools/shared/neural"

// Type aliases for backward compatibility.
type Network = neural.Network
type Inferer = neural.Inferer
type WeightLoader = neural.WeightLoader
type WeightProvider = neural.WeightProvider
type NetworkConfigurer = neural.NetworkConfigurer
type DataLoader = neural.DataLoader
type DualModeNetwork = neural.DualModeNetwork
type InferenceResult = neural.InferenceResult
type NetworkConfig = neural.NetworkConfig
type WeightsFile = neural.WeightsFile
type QuantizationStats = neural.QuantizationStats
type EnergyEstimate = neural.EnergyEstimate
type ModeMetrics = neural.ModeMetrics
type DualModeDatasetMetrics = neural.DualModeDatasetMetrics
type CIMNoiseComponents = neural.CIMNoiseComponents
type TIAModel = neural.TIAModel
type RandomSource = neural.RandomSource

// Constants re-exported for backward compatibility.
const FeCIMLevels = neural.FeCIMLevels
const MaxDemoLevels = neural.MaxDemoLevels
const MaxNoiseLevel = neural.MaxNoiseLevel
const EnergyPerBitPerMACJ = neural.EnergyPerBitPerMACJ
const EnergyPerDACJ = neural.EnergyPerDACJ
const EnergyPerADCJ = neural.EnergyPerADCJ

// Function aliases for backward compatibility.
var NewDualModeNetwork = neural.NewDualModeNetwork
var DefaultNetworkConfig = neural.DefaultNetworkConfig
var ScanAvailableQATLevels = neural.ScanAvailableQATLevels
var GetWeightsFilename = neural.GetWeightsFilename
var GetBestMatchingWeightsLevel = neural.GetBestMatchingWeightsLevel
var QuantizeWeights = neural.QuantizeWeights
var QuantizeBias = neural.QuantizeBias
var ComputeQuantizationStats = neural.ComputeQuantizationStats
var AddGaussianNoise = neural.AddGaussianNoise
var AddGaussianNoiseInPlace = neural.AddGaussianNoiseInPlace
var NewRandomSource = neural.NewRandomSource
var EnergyPerMACJ = neural.EnergyPerMACJ
var EstimateInferenceEnergyJ = neural.EstimateInferenceEnergyJ
var EstimateInferenceEnergyMicroJ = neural.EstimateInferenceEnergyMicroJ
var EvaluateDualModeDataset = neural.EvaluateDualModeDataset
var InitGPU = neural.InitGPU
var IsGPUAvailable = neural.IsGPUAvailable
var DestroyGPU = neural.DestroyGPU
