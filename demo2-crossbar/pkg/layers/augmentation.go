// Package layers provides neural network layer implementations for crossbar-based CIM.
// augmentation.go implements data augmentation utilities for training neural networks.
//
// Data augmentation is critical for improving model generalization, especially when
// training models destined for CIM deployment where noise and quantization effects
// require robust feature learning.
//
// CIM-specific considerations:
// - Augmentations should produce values within quantized ranges
// - Noise injection augmentation mimics crossbar variability
// - Careful with intensity transformations that may push values out of ADC range

package layers

import (
	"math"
	"math/rand"
)

// Augmentation represents a data augmentation operation
type Augmentation interface {
	Apply(image [][]float64) [][]float64
	Name() string
}

// AugmentationPipeline chains multiple augmentations
type AugmentationPipeline struct {
	Augmentations []Augmentation
	Probability   float64 // Overall probability of applying pipeline
}

// NewAugmentationPipeline creates a new augmentation pipeline
func NewAugmentationPipeline(probability float64) *AugmentationPipeline {
	return &AugmentationPipeline{
		Augmentations: make([]Augmentation, 0),
		Probability:   probability,
	}
}

// Add adds an augmentation to the pipeline
func (ap *AugmentationPipeline) Add(aug Augmentation) *AugmentationPipeline {
	ap.Augmentations = append(ap.Augmentations, aug)
	return ap
}

// Apply applies all augmentations in sequence
func (ap *AugmentationPipeline) Apply(image [][]float64) [][]float64 {
	if rand.Float64() > ap.Probability {
		return image
	}
	result := copyImage(image)
	for _, aug := range ap.Augmentations {
		result = aug.Apply(result)
	}
	return result
}

// RandomChoice applies one random augmentation from the list
type RandomChoice struct {
	Augmentations []Augmentation
}

// NewRandomChoice creates a random choice augmentation
func NewRandomChoice(augmentations ...Augmentation) *RandomChoice {
	return &RandomChoice{Augmentations: augmentations}
}

// Apply applies one random augmentation
func (rc *RandomChoice) Apply(image [][]float64) [][]float64 {
	if len(rc.Augmentations) == 0 {
		return image
	}
	idx := rand.Intn(len(rc.Augmentations))
	return rc.Augmentations[idx].Apply(image)
}

// Name returns the name
func (rc *RandomChoice) Name() string {
	return "RandomChoice"
}

// ============================================================================
// Geometric Augmentations
// ============================================================================

// HorizontalFlip flips image horizontally
type HorizontalFlip struct {
	Probability float64
}

// NewHorizontalFlip creates a horizontal flip augmentation
func NewHorizontalFlip(probability float64) *HorizontalFlip {
	return &HorizontalFlip{Probability: probability}
}

// Apply applies horizontal flip
func (hf *HorizontalFlip) Apply(image [][]float64) [][]float64 {
	if rand.Float64() > hf.Probability {
		return image
	}
	h := len(image)
	if h == 0 {
		return image
	}
	w := len(image[0])
	result := make([][]float64, h)
	for i := 0; i < h; i++ {
		result[i] = make([]float64, w)
		for j := 0; j < w; j++ {
			result[i][j] = image[i][w-1-j]
		}
	}
	return result
}

// Name returns the name
func (hf *HorizontalFlip) Name() string {
	return "HorizontalFlip"
}

// VerticalFlip flips image vertically
type VerticalFlip struct {
	Probability float64
}

// NewVerticalFlip creates a vertical flip augmentation
func NewVerticalFlip(probability float64) *VerticalFlip {
	return &VerticalFlip{Probability: probability}
}

// Apply applies vertical flip
func (vf *VerticalFlip) Apply(image [][]float64) [][]float64 {
	if rand.Float64() > vf.Probability {
		return image
	}
	h := len(image)
	if h == 0 {
		return image
	}
	w := len(image[0])
	result := make([][]float64, h)
	for i := 0; i < h; i++ {
		result[i] = make([]float64, w)
		copy(result[i], image[h-1-i])
	}
	return result
}

// Name returns the name
func (vf *VerticalFlip) Name() string {
	return "VerticalFlip"
}

// Rotation rotates image by 90-degree increments
type Rotation struct {
	Angles      []int   // Valid: 90, 180, 270
	Probability float64
}

// NewRotation creates a rotation augmentation
func NewRotation(probability float64, angles ...int) *Rotation {
	if len(angles) == 0 {
		angles = []int{90, 180, 270}
	}
	return &Rotation{Angles: angles, Probability: probability}
}

// Apply applies rotation
func (r *Rotation) Apply(image [][]float64) [][]float64 {
	if rand.Float64() > r.Probability {
		return image
	}
	angle := r.Angles[rand.Intn(len(r.Angles))]
	return rotateImage(image, angle)
}

// Name returns the name
func (r *Rotation) Name() string {
	return "Rotation"
}

// rotateImage rotates by 90-degree increments
func rotateImage(image [][]float64, angle int) [][]float64 {
	h := len(image)
	if h == 0 {
		return image
	}
	w := len(image[0])

	switch angle % 360 {
	case 90, -270:
		result := make([][]float64, w)
		for i := 0; i < w; i++ {
			result[i] = make([]float64, h)
			for j := 0; j < h; j++ {
				result[i][j] = image[h-1-j][i]
			}
		}
		return result
	case 180, -180:
		result := make([][]float64, h)
		for i := 0; i < h; i++ {
			result[i] = make([]float64, w)
			for j := 0; j < w; j++ {
				result[i][j] = image[h-1-i][w-1-j]
			}
		}
		return result
	case 270, -90:
		result := make([][]float64, w)
		for i := 0; i < w; i++ {
			result[i] = make([]float64, h)
			for j := 0; j < h; j++ {
				result[i][j] = image[j][w-1-i]
			}
		}
		return result
	default:
		return image
	}
}

// RandomCrop crops a random region from the image
type RandomCrop struct {
	CropH       int
	CropW       int
	Probability float64
}

// NewRandomCrop creates a random crop augmentation
func NewRandomCrop(cropH, cropW int, probability float64) *RandomCrop {
	return &RandomCrop{CropH: cropH, CropW: cropW, Probability: probability}
}

// Apply applies random crop
func (rc *RandomCrop) Apply(image [][]float64) [][]float64 {
	if rand.Float64() > rc.Probability {
		return image
	}
	h := len(image)
	if h == 0 {
		return image
	}
	w := len(image[0])

	if rc.CropH >= h || rc.CropW >= w {
		return image
	}

	startH := rand.Intn(h - rc.CropH)
	startW := rand.Intn(w - rc.CropW)

	result := make([][]float64, rc.CropH)
	for i := 0; i < rc.CropH; i++ {
		result[i] = make([]float64, rc.CropW)
		copy(result[i], image[startH+i][startW:startW+rc.CropW])
	}
	return result
}

// Name returns the name
func (rc *RandomCrop) Name() string {
	return "RandomCrop"
}

// CenterCrop crops from the center of the image
type CenterCrop struct {
	CropH int
	CropW int
}

// NewCenterCrop creates a center crop
func NewCenterCrop(cropH, cropW int) *CenterCrop {
	return &CenterCrop{CropH: cropH, CropW: cropW}
}

// Apply applies center crop
func (cc *CenterCrop) Apply(image [][]float64) [][]float64 {
	h := len(image)
	if h == 0 {
		return image
	}
	w := len(image[0])

	if cc.CropH >= h || cc.CropW >= w {
		return image
	}

	startH := (h - cc.CropH) / 2
	startW := (w - cc.CropW) / 2

	result := make([][]float64, cc.CropH)
	for i := 0; i < cc.CropH; i++ {
		result[i] = make([]float64, cc.CropW)
		copy(result[i], image[startH+i][startW:startW+cc.CropW])
	}
	return result
}

// Name returns the name
func (cc *CenterCrop) Name() string {
	return "CenterCrop"
}

// RandomResizedCrop crops and resizes to original size
type RandomResizedCrop struct {
	ScaleMin    float64 // Minimum crop scale (0.08 typical)
	ScaleMax    float64 // Maximum crop scale (1.0 typical)
	RatioMin    float64 // Minimum aspect ratio (0.75 typical)
	RatioMax    float64 // Maximum aspect ratio (1.33 typical)
	Probability float64
}

// NewRandomResizedCrop creates a random resized crop
func NewRandomResizedCrop(scaleMin, scaleMax, ratioMin, ratioMax, probability float64) *RandomResizedCrop {
	return &RandomResizedCrop{
		ScaleMin:    scaleMin,
		ScaleMax:    scaleMax,
		RatioMin:    ratioMin,
		RatioMax:    ratioMax,
		Probability: probability,
	}
}

// Apply applies random resized crop
func (rrc *RandomResizedCrop) Apply(image [][]float64) [][]float64 {
	if rand.Float64() > rrc.Probability {
		return image
	}
	h := len(image)
	if h == 0 {
		return image
	}
	w := len(image[0])

	area := float64(h * w)
	scale := rrc.ScaleMin + rand.Float64()*(rrc.ScaleMax-rrc.ScaleMin)
	ratio := math.Exp(math.Log(rrc.RatioMin) + rand.Float64()*(math.Log(rrc.RatioMax)-math.Log(rrc.RatioMin)))

	targetArea := area * scale
	cropW := int(math.Sqrt(targetArea * ratio))
	cropH := int(math.Sqrt(targetArea / ratio))

	if cropW > w {
		cropW = w
	}
	if cropH > h {
		cropH = h
	}
	if cropW < 1 {
		cropW = 1
	}
	if cropH < 1 {
		cropH = 1
	}

	startH := rand.Intn(h - cropH + 1)
	startW := rand.Intn(w - cropW + 1)

	// Crop
	cropped := make([][]float64, cropH)
	for i := 0; i < cropH; i++ {
		cropped[i] = make([]float64, cropW)
		copy(cropped[i], image[startH+i][startW:startW+cropW])
	}

	// Resize back to original
	return bilinearResize(cropped, h, w)
}

// Name returns the name
func (rrc *RandomResizedCrop) Name() string {
	return "RandomResizedCrop"
}

// bilinearResize resizes image using bilinear interpolation
func bilinearResize(image [][]float64, newH, newW int) [][]float64 {
	h := len(image)
	if h == 0 {
		return image
	}
	w := len(image[0])

	result := make([][]float64, newH)
	for i := 0; i < newH; i++ {
		result[i] = make([]float64, newW)
		for j := 0; j < newW; j++ {
			srcY := float64(i) * float64(h-1) / float64(newH-1)
			srcX := float64(j) * float64(w-1) / float64(newW-1)

			y0 := int(srcY)
			x0 := int(srcX)
			y1 := y0 + 1
			x1 := x0 + 1

			if y1 >= h {
				y1 = h - 1
			}
			if x1 >= w {
				x1 = w - 1
			}

			dy := srcY - float64(y0)
			dx := srcX - float64(x0)

			result[i][j] = (1-dy)*(1-dx)*image[y0][x0] +
				(1-dy)*dx*image[y0][x1] +
				dy*(1-dx)*image[y1][x0] +
				dy*dx*image[y1][x1]
		}
	}
	return result
}

// Translate shifts image by random offset
type Translate struct {
	MaxH        float64 // Max translation as fraction of height
	MaxW        float64 // Max translation as fraction of width
	FillValue   float64
	Probability float64
}

// NewTranslate creates a translation augmentation
func NewTranslate(maxH, maxW, fillValue, probability float64) *Translate {
	return &Translate{MaxH: maxH, MaxW: maxW, FillValue: fillValue, Probability: probability}
}

// Apply applies translation
func (t *Translate) Apply(image [][]float64) [][]float64 {
	if rand.Float64() > t.Probability {
		return image
	}
	h := len(image)
	if h == 0 {
		return image
	}
	w := len(image[0])

	shiftH := int(float64(h) * t.MaxH * (2*rand.Float64() - 1))
	shiftW := int(float64(w) * t.MaxW * (2*rand.Float64() - 1))

	result := make([][]float64, h)
	for i := 0; i < h; i++ {
		result[i] = make([]float64, w)
		for j := 0; j < w; j++ {
			srcI := i - shiftH
			srcJ := j - shiftW
			if srcI >= 0 && srcI < h && srcJ >= 0 && srcJ < w {
				result[i][j] = image[srcI][srcJ]
			} else {
				result[i][j] = t.FillValue
			}
		}
	}
	return result
}

// Name returns the name
func (t *Translate) Name() string {
	return "Translate"
}

// Scale scales image by random factor
type Scale struct {
	ScaleMin    float64
	ScaleMax    float64
	Probability float64
}

// NewScale creates a scale augmentation
func NewScale(scaleMin, scaleMax, probability float64) *Scale {
	return &Scale{ScaleMin: scaleMin, ScaleMax: scaleMax, Probability: probability}
}

// Apply applies scaling
func (s *Scale) Apply(image [][]float64) [][]float64 {
	if rand.Float64() > s.Probability {
		return image
	}
	h := len(image)
	if h == 0 {
		return image
	}
	w := len(image[0])

	scale := s.ScaleMin + rand.Float64()*(s.ScaleMax-s.ScaleMin)
	newH := int(float64(h) * scale)
	newW := int(float64(w) * scale)

	if newH < 1 {
		newH = 1
	}
	if newW < 1 {
		newW = 1
	}

	scaled := bilinearResize(image, newH, newW)

	// Pad or crop to original size
	result := make([][]float64, h)
	for i := 0; i < h; i++ {
		result[i] = make([]float64, w)
		for j := 0; j < w; j++ {
			srcI := i - (h-newH)/2
			srcJ := j - (w-newW)/2
			if srcI >= 0 && srcI < newH && srcJ >= 0 && srcJ < newW {
				result[i][j] = scaled[srcI][srcJ]
			}
		}
	}
	return result
}

// Name returns the name
func (s *Scale) Name() string {
	return "Scale"
}

// ElasticTransform applies elastic deformation
type ElasticTransform struct {
	Alpha       float64 // Intensity of deformation
	Sigma       float64 // Smoothness (Gaussian std)
	Probability float64
}

// NewElasticTransform creates elastic transform
func NewElasticTransform(alpha, sigma, probability float64) *ElasticTransform {
	return &ElasticTransform{Alpha: alpha, Sigma: sigma, Probability: probability}
}

// Apply applies elastic transform
func (et *ElasticTransform) Apply(image [][]float64) [][]float64 {
	if rand.Float64() > et.Probability {
		return image
	}
	h := len(image)
	if h == 0 {
		return image
	}
	w := len(image[0])

	// Generate random displacement fields
	dx := make([][]float64, h)
	dy := make([][]float64, h)
	for i := 0; i < h; i++ {
		dx[i] = make([]float64, w)
		dy[i] = make([]float64, w)
		for j := 0; j < w; j++ {
			dx[i][j] = (2*rand.Float64() - 1) * et.Alpha
			dy[i][j] = (2*rand.Float64() - 1) * et.Alpha
		}
	}

	// Apply Gaussian smoothing to displacement fields
	dx = gaussianBlur2D(dx, et.Sigma)
	dy = gaussianBlur2D(dy, et.Sigma)

	// Apply displacement
	result := make([][]float64, h)
	for i := 0; i < h; i++ {
		result[i] = make([]float64, w)
		for j := 0; j < w; j++ {
			srcY := float64(i) + dy[i][j]
			srcX := float64(j) + dx[i][j]

			// Bilinear interpolation
			y0 := int(math.Floor(srcY))
			x0 := int(math.Floor(srcX))
			y1 := y0 + 1
			x1 := x0 + 1

			if y0 < 0 || y1 >= h || x0 < 0 || x1 >= w {
				result[i][j] = 0
				continue
			}

			fy := srcY - float64(y0)
			fx := srcX - float64(x0)

			result[i][j] = (1-fy)*(1-fx)*image[y0][x0] +
				(1-fy)*fx*image[y0][x1] +
				fy*(1-fx)*image[y1][x0] +
				fy*fx*image[y1][x1]
		}
	}
	return result
}

// Name returns the name
func (et *ElasticTransform) Name() string {
	return "ElasticTransform"
}

// gaussianBlur2D applies Gaussian blur
func gaussianBlur2D(image [][]float64, sigma float64) [][]float64 {
	h := len(image)
	if h == 0 {
		return image
	}
	w := len(image[0])

	kernelSize := int(6*sigma) | 1 // Ensure odd
	if kernelSize < 3 {
		kernelSize = 3
	}
	half := kernelSize / 2

	// Create Gaussian kernel
	kernel := make([]float64, kernelSize)
	sum := 0.0
	for i := 0; i < kernelSize; i++ {
		x := float64(i - half)
		kernel[i] = math.Exp(-x * x / (2 * sigma * sigma))
		sum += kernel[i]
	}
	for i := range kernel {
		kernel[i] /= sum
	}

	// Horizontal pass
	temp := make([][]float64, h)
	for i := 0; i < h; i++ {
		temp[i] = make([]float64, w)
		for j := 0; j < w; j++ {
			val := 0.0
			for k := 0; k < kernelSize; k++ {
				jj := j + k - half
				if jj < 0 {
					jj = 0
				}
				if jj >= w {
					jj = w - 1
				}
				val += image[i][jj] * kernel[k]
			}
			temp[i][j] = val
		}
	}

	// Vertical pass
	result := make([][]float64, h)
	for i := 0; i < h; i++ {
		result[i] = make([]float64, w)
		for j := 0; j < w; j++ {
			val := 0.0
			for k := 0; k < kernelSize; k++ {
				ii := i + k - half
				if ii < 0 {
					ii = 0
				}
				if ii >= h {
					ii = h - 1
				}
				val += temp[ii][j] * kernel[k]
			}
			result[i][j] = val
		}
	}
	return result
}

// ============================================================================
// Color/Intensity Augmentations
// ============================================================================

// Normalize normalizes image to specified mean and std
type Normalize struct {
	Mean float64
	Std  float64
}

// NewNormalize creates a normalize augmentation
func NewNormalize(mean, std float64) *Normalize {
	return &Normalize{Mean: mean, Std: std}
}

// Apply applies normalization
func (n *Normalize) Apply(image [][]float64) [][]float64 {
	h := len(image)
	if h == 0 {
		return image
	}
	w := len(image[0])

	result := make([][]float64, h)
	for i := 0; i < h; i++ {
		result[i] = make([]float64, w)
		for j := 0; j < w; j++ {
			result[i][j] = (image[i][j] - n.Mean) / n.Std
		}
	}
	return result
}

// Name returns the name
func (n *Normalize) Name() string {
	return "Normalize"
}

// RandomBrightness adjusts brightness randomly
type RandomBrightness struct {
	MaxDelta    float64
	Probability float64
}

// NewRandomBrightness creates brightness augmentation
func NewRandomBrightness(maxDelta, probability float64) *RandomBrightness {
	return &RandomBrightness{MaxDelta: maxDelta, Probability: probability}
}

// Apply applies brightness adjustment
func (rb *RandomBrightness) Apply(image [][]float64) [][]float64 {
	if rand.Float64() > rb.Probability {
		return image
	}
	delta := rb.MaxDelta * (2*rand.Float64() - 1)

	h := len(image)
	if h == 0 {
		return image
	}
	w := len(image[0])

	result := make([][]float64, h)
	for i := 0; i < h; i++ {
		result[i] = make([]float64, w)
		for j := 0; j < w; j++ {
			result[i][j] = clamp(image[i][j]+delta, 0, 1)
		}
	}
	return result
}

// Name returns the name
func (rb *RandomBrightness) Name() string {
	return "RandomBrightness"
}

// RandomContrast adjusts contrast randomly
type RandomContrast struct {
	MinFactor   float64
	MaxFactor   float64
	Probability float64
}

// NewRandomContrast creates contrast augmentation
func NewRandomContrast(minFactor, maxFactor, probability float64) *RandomContrast {
	return &RandomContrast{MinFactor: minFactor, MaxFactor: maxFactor, Probability: probability}
}

// Apply applies contrast adjustment
func (rc *RandomContrast) Apply(image [][]float64) [][]float64 {
	if rand.Float64() > rc.Probability {
		return image
	}
	factor := rc.MinFactor + rand.Float64()*(rc.MaxFactor-rc.MinFactor)

	h := len(image)
	if h == 0 {
		return image
	}
	w := len(image[0])

	// Compute mean
	sum := 0.0
	count := 0
	for i := 0; i < h; i++ {
		for j := 0; j < w; j++ {
			sum += image[i][j]
			count++
		}
	}
	mean := sum / float64(count)

	result := make([][]float64, h)
	for i := 0; i < h; i++ {
		result[i] = make([]float64, w)
		for j := 0; j < w; j++ {
			result[i][j] = clamp(mean+(image[i][j]-mean)*factor, 0, 1)
		}
	}
	return result
}

// Name returns the name
func (rc *RandomContrast) Name() string {
	return "RandomContrast"
}

// GammaCorrection applies gamma correction
type GammaCorrection struct {
	GammaMin    float64
	GammaMax    float64
	Probability float64
}

// NewGammaCorrection creates gamma correction augmentation
func NewGammaCorrection(gammaMin, gammaMax, probability float64) *GammaCorrection {
	return &GammaCorrection{GammaMin: gammaMin, GammaMax: gammaMax, Probability: probability}
}

// Apply applies gamma correction
func (gc *GammaCorrection) Apply(image [][]float64) [][]float64 {
	if rand.Float64() > gc.Probability {
		return image
	}
	gamma := gc.GammaMin + rand.Float64()*(gc.GammaMax-gc.GammaMin)

	h := len(image)
	if h == 0 {
		return image
	}
	w := len(image[0])

	result := make([][]float64, h)
	for i := 0; i < h; i++ {
		result[i] = make([]float64, w)
		for j := 0; j < w; j++ {
			result[i][j] = math.Pow(clamp(image[i][j], 0, 1), gamma)
		}
	}
	return result
}

// Name returns the name
func (gc *GammaCorrection) Name() string {
	return "GammaCorrection"
}

// Invert inverts pixel values
type Invert struct {
	Probability float64
}

// NewInvert creates invert augmentation
func NewInvert(probability float64) *Invert {
	return &Invert{Probability: probability}
}

// Apply applies inversion
func (inv *Invert) Apply(image [][]float64) [][]float64 {
	if rand.Float64() > inv.Probability {
		return image
	}
	h := len(image)
	if h == 0 {
		return image
	}
	w := len(image[0])

	result := make([][]float64, h)
	for i := 0; i < h; i++ {
		result[i] = make([]float64, w)
		for j := 0; j < w; j++ {
			result[i][j] = 1.0 - image[i][j]
		}
	}
	return result
}

// Name returns the name
func (inv *Invert) Name() string {
	return "Invert"
}

// ============================================================================
// Noise Augmentations (CIM-specific)
// ============================================================================

// GaussianNoise adds Gaussian noise (simulates crossbar thermal noise)
type GaussianNoise struct {
	Mean        float64
	Std         float64
	Probability float64
}

// NewGaussianNoise creates Gaussian noise augmentation
func NewGaussianNoise(mean, std, probability float64) *GaussianNoise {
	return &GaussianNoise{Mean: mean, Std: std, Probability: probability}
}

// Apply applies Gaussian noise
func (gn *GaussianNoise) Apply(image [][]float64) [][]float64 {
	if rand.Float64() > gn.Probability {
		return image
	}
	h := len(image)
	if h == 0 {
		return image
	}
	w := len(image[0])

	result := make([][]float64, h)
	for i := 0; i < h; i++ {
		result[i] = make([]float64, w)
		for j := 0; j < w; j++ {
			noise := gn.Mean + gn.Std*rand.NormFloat64()
			result[i][j] = image[i][j] + noise
		}
	}
	return result
}

// Name returns the name
func (gn *GaussianNoise) Name() string {
	return "GaussianNoise"
}

// SaltAndPepperNoise adds salt and pepper noise
type SaltAndPepperNoise struct {
	Amount      float64 // Fraction of pixels affected
	SaltRatio   float64 // Ratio of salt to pepper (0.5 = equal)
	Probability float64
}

// NewSaltAndPepperNoise creates salt and pepper noise
func NewSaltAndPepperNoise(amount, saltRatio, probability float64) *SaltAndPepperNoise {
	return &SaltAndPepperNoise{Amount: amount, SaltRatio: saltRatio, Probability: probability}
}

// Apply applies salt and pepper noise
func (sp *SaltAndPepperNoise) Apply(image [][]float64) [][]float64 {
	if rand.Float64() > sp.Probability {
		return image
	}
	h := len(image)
	if h == 0 {
		return image
	}
	w := len(image[0])

	result := make([][]float64, h)
	for i := 0; i < h; i++ {
		result[i] = make([]float64, w)
		for j := 0; j < w; j++ {
			if rand.Float64() < sp.Amount {
				if rand.Float64() < sp.SaltRatio {
					result[i][j] = 1.0 // Salt
				} else {
					result[i][j] = 0.0 // Pepper
				}
			} else {
				result[i][j] = image[i][j]
			}
		}
	}
	return result
}

// Name returns the name
func (sp *SaltAndPepperNoise) Name() string {
	return "SaltAndPepperNoise"
}

// CrossbarNoise simulates CIM crossbar array noise characteristics
type CrossbarNoise struct {
	ThermalNoise  float64 // Thermal noise std (typically 1-5%)
	ShotNoise     float64 // Shot noise coefficient
	ReadNoise     float64 // Read circuit noise std
	LineRes       float64 // Wire resistance variation (0-10%)
	Probability   float64
}

// NewCrossbarNoise creates crossbar-specific noise augmentation
func NewCrossbarNoise(thermal, shot, readNoise, lineRes, probability float64) *CrossbarNoise {
	return &CrossbarNoise{
		ThermalNoise: thermal,
		ShotNoise:    shot,
		ReadNoise:    readNoise,
		LineRes:      lineRes,
		Probability:  probability,
	}
}

// Apply applies crossbar noise model
func (cn *CrossbarNoise) Apply(image [][]float64) [][]float64 {
	if rand.Float64() > cn.Probability {
		return image
	}
	h := len(image)
	if h == 0 {
		return image
	}
	w := len(image[0])

	result := make([][]float64, h)
	for i := 0; i < h; i++ {
		result[i] = make([]float64, w)
		// Position-dependent wire resistance (increases with distance)
		rowFactor := 1.0 + cn.LineRes*float64(i)/float64(h)
		for j := 0; j < w; j++ {
			colFactor := 1.0 + cn.LineRes*float64(j)/float64(w)
			wireNoise := (rowFactor + colFactor - 2.0) * (rand.Float64() - 0.5)

			// Signal-dependent shot noise
			signal := image[i][j]
			shotComponent := cn.ShotNoise * math.Sqrt(math.Abs(signal)) * rand.NormFloat64()

			// Add all noise components
			totalNoise := cn.ThermalNoise*rand.NormFloat64() +
				shotComponent +
				cn.ReadNoise*rand.NormFloat64() +
				wireNoise

			result[i][j] = signal + totalNoise
		}
	}
	return result
}

// Name returns the name
func (cn *CrossbarNoise) Name() string {
	return "CrossbarNoise"
}

// QuantizationNoise simulates ADC/DAC quantization (CIM-specific)
type QuantizationNoise struct {
	Bits        int     // Number of bits
	AddNoise    bool    // Add uniform noise within LSB
	Probability float64
}

// NewQuantizationNoise creates quantization noise augmentation
func NewQuantizationNoise(bits int, addNoise bool, probability float64) *QuantizationNoise {
	return &QuantizationNoise{Bits: bits, AddNoise: addNoise, Probability: probability}
}

// Apply applies quantization
func (qn *QuantizationNoise) Apply(image [][]float64) [][]float64 {
	if rand.Float64() > qn.Probability {
		return image
	}
	h := len(image)
	if h == 0 {
		return image
	}
	w := len(image[0])

	levels := float64(int(1) << qn.Bits)
	step := 1.0 / (levels - 1)

	result := make([][]float64, h)
	for i := 0; i < h; i++ {
		result[i] = make([]float64, w)
		for j := 0; j < w; j++ {
			// Quantize
			level := math.Round(image[i][j] * (levels - 1))
			if level < 0 {
				level = 0
			}
			if level >= levels {
				level = levels - 1
			}
			quantized := level / (levels - 1)

			if qn.AddNoise {
				// Add uniform noise within one LSB
				noise := (rand.Float64() - 0.5) * step
				result[i][j] = quantized + noise
			} else {
				result[i][j] = quantized
			}
		}
	}
	return result
}

// Name returns the name
func (qn *QuantizationNoise) Name() string {
	return "QuantizationNoise"
}

// ============================================================================
// Cutout/Erasing Augmentations
// ============================================================================

// Cutout randomly erases rectangular regions
type Cutout struct {
	NumHoles    int     // Number of holes
	HoleSize    int     // Size of each hole
	FillValue   float64 // Value to fill holes with
	Probability float64
}

// NewCutout creates cutout augmentation
func NewCutout(numHoles, holeSize int, fillValue, probability float64) *Cutout {
	return &Cutout{NumHoles: numHoles, HoleSize: holeSize, FillValue: fillValue, Probability: probability}
}

// Apply applies cutout
func (c *Cutout) Apply(image [][]float64) [][]float64 {
	if rand.Float64() > c.Probability {
		return image
	}
	result := copyImage(image)
	h := len(result)
	if h == 0 {
		return result
	}
	w := len(result[0])

	for n := 0; n < c.NumHoles; n++ {
		centerH := rand.Intn(h)
		centerW := rand.Intn(w)

		startH := max(0, centerH-c.HoleSize/2)
		endH := min(h, centerH+c.HoleSize/2)
		startW := max(0, centerW-c.HoleSize/2)
		endW := min(w, centerW+c.HoleSize/2)

		for i := startH; i < endH; i++ {
			for j := startW; j < endW; j++ {
				result[i][j] = c.FillValue
			}
		}
	}
	return result
}

// Name returns the name
func (c *Cutout) Name() string {
	return "Cutout"
}

// RandomErasing erases random rectangles (more configurable than Cutout)
type RandomErasing struct {
	AreaMin     float64 // Min area ratio to erase
	AreaMax     float64 // Max area ratio to erase
	RatioMin    float64 // Min aspect ratio
	RatioMax    float64 // Max aspect ratio
	FillMode    string  // "constant", "random", "mean"
	FillValue   float64
	Probability float64
}

// NewRandomErasing creates random erasing augmentation
func NewRandomErasing(areaMin, areaMax, ratioMin, ratioMax float64, fillMode string, fillValue, probability float64) *RandomErasing {
	return &RandomErasing{
		AreaMin:     areaMin,
		AreaMax:     areaMax,
		RatioMin:    ratioMin,
		RatioMax:    ratioMax,
		FillMode:    fillMode,
		FillValue:   fillValue,
		Probability: probability,
	}
}

// Apply applies random erasing
func (re *RandomErasing) Apply(image [][]float64) [][]float64 {
	if rand.Float64() > re.Probability {
		return image
	}
	result := copyImage(image)
	h := len(result)
	if h == 0 {
		return result
	}
	w := len(result[0])
	area := float64(h * w)

	// Try to find valid erase region
	for attempts := 0; attempts < 10; attempts++ {
		eraseArea := area * (re.AreaMin + rand.Float64()*(re.AreaMax-re.AreaMin))
		ratio := re.RatioMin + rand.Float64()*(re.RatioMax-re.RatioMin)

		eraseH := int(math.Sqrt(eraseArea / ratio))
		eraseW := int(math.Sqrt(eraseArea * ratio))

		if eraseH < h && eraseW < w {
			startH := rand.Intn(h - eraseH)
			startW := rand.Intn(w - eraseW)

			// Compute mean if needed
			fillVal := re.FillValue
			if re.FillMode == "mean" {
				sum := 0.0
				count := 0
				for i := 0; i < h; i++ {
					for j := 0; j < w; j++ {
						sum += image[i][j]
						count++
					}
				}
				fillVal = sum / float64(count)
			}

			// Apply erasing
			for i := startH; i < startH+eraseH; i++ {
				for j := startW; j < startW+eraseW; j++ {
					switch re.FillMode {
					case "random":
						result[i][j] = rand.Float64()
					default:
						result[i][j] = fillVal
					}
				}
			}
			break
		}
	}
	return result
}

// Name returns the name
func (re *RandomErasing) Name() string {
	return "RandomErasing"
}

// GridMask applies grid-pattern masking
type GridMask struct {
	GridSize    int     // Size of grid cells
	Ratio       float64 // Ratio of masked area
	FillValue   float64
	Probability float64
}

// NewGridMask creates grid mask augmentation
func NewGridMask(gridSize int, ratio, fillValue, probability float64) *GridMask {
	return &GridMask{GridSize: gridSize, Ratio: ratio, FillValue: fillValue, Probability: probability}
}

// Apply applies grid mask
func (gm *GridMask) Apply(image [][]float64) [][]float64 {
	if rand.Float64() > gm.Probability {
		return image
	}
	result := copyImage(image)
	h := len(result)
	if h == 0 {
		return result
	}
	w := len(result[0])

	maskSize := int(float64(gm.GridSize) * gm.Ratio)

	for i := 0; i < h; i++ {
		gridI := i % gm.GridSize
		for j := 0; j < w; j++ {
			gridJ := j % gm.GridSize
			if gridI < maskSize && gridJ < maskSize {
				result[i][j] = gm.FillValue
			}
		}
	}
	return result
}

// Name returns the name
func (gm *GridMask) Name() string {
	return "GridMask"
}

// ============================================================================
// Mixing Augmentations
// ============================================================================

// MixUp mixes two images (for batch-level augmentation)
type MixUp struct {
	Alpha float64 // Beta distribution parameter
}

// NewMixUp creates mixup augmentation
func NewMixUp(alpha float64) *MixUp {
	return &MixUp{Alpha: alpha}
}

// Mix mixes two images with sampled lambda
func (m *MixUp) Mix(image1, image2 [][]float64) ([][]float64, float64) {
	// Sample lambda from Beta(alpha, alpha)
	lambda := betaSample(m.Alpha, m.Alpha)

	h := len(image1)
	if h == 0 {
		return image1, lambda
	}
	w := len(image1[0])

	result := make([][]float64, h)
	for i := 0; i < h; i++ {
		result[i] = make([]float64, w)
		for j := 0; j < w; j++ {
			result[i][j] = lambda*image1[i][j] + (1-lambda)*image2[i][j]
		}
	}
	return result, lambda
}

// Apply applies identity (mixup needs two images)
func (m *MixUp) Apply(image [][]float64) [][]float64 {
	return image
}

// Name returns the name
func (m *MixUp) Name() string {
	return "MixUp"
}

// CutMix cuts and pastes between images
type CutMix struct {
	Alpha float64 // Beta distribution parameter
}

// NewCutMix creates cutmix augmentation
func NewCutMix(alpha float64) *CutMix {
	return &CutMix{Alpha: alpha}
}

// Mix performs cutmix on two images
func (cm *CutMix) Mix(image1, image2 [][]float64) ([][]float64, float64) {
	lambda := betaSample(cm.Alpha, cm.Alpha)

	h := len(image1)
	if h == 0 {
		return image1, lambda
	}
	w := len(image1[0])

	// Compute cut region
	cutRatio := math.Sqrt(1 - lambda)
	cutH := int(float64(h) * cutRatio)
	cutW := int(float64(w) * cutRatio)

	centerH := rand.Intn(h)
	centerW := rand.Intn(w)

	startH := max(0, centerH-cutH/2)
	endH := min(h, centerH+cutH/2)
	startW := max(0, centerW-cutW/2)
	endW := min(w, centerW+cutW/2)

	result := copyImage(image1)
	for i := startH; i < endH; i++ {
		for j := startW; j < endW; j++ {
			if i < len(image2) && j < len(image2[0]) {
				result[i][j] = image2[i][j]
			}
		}
	}

	// Adjust lambda based on actual cut size
	actualLambda := 1.0 - float64((endH-startH)*(endW-startW))/float64(h*w)
	return result, actualLambda
}

// Apply applies identity
func (cm *CutMix) Apply(image [][]float64) [][]float64 {
	return image
}

// Name returns the name
func (cm *CutMix) Name() string {
	return "CutMix"
}

// betaSample samples from Beta distribution using gamma sampling
func betaSample(alpha, beta float64) float64 {
	// Simple approximation using uniform when alpha=beta=1
	if alpha == 1 && beta == 1 {
		return rand.Float64()
	}
	// For other cases, use gamma sampling
	x := gammaSample(alpha)
	y := gammaSample(beta)
	return x / (x + y)
}

// gammaSample samples from Gamma distribution (Marsaglia and Tsang method)
func gammaSample(shape float64) float64 {
	if shape < 1 {
		return gammaSample(shape+1) * math.Pow(rand.Float64(), 1/shape)
	}
	d := shape - 1.0/3.0
	c := 1.0 / math.Sqrt(9.0*d)
	for {
		var x, v float64
		for {
			x = rand.NormFloat64()
			v = 1.0 + c*x
			if v > 0 {
				break
			}
		}
		v = v * v * v
		u := rand.Float64()
		if u < 1.0-0.0331*(x*x)*(x*x) {
			return d * v
		}
		if math.Log(u) < 0.5*x*x+d*(1.0-v+math.Log(v)) {
			return d * v
		}
	}
}

// ============================================================================
// Morphological Augmentations
// ============================================================================

// Erosion applies morphological erosion
type Erosion struct {
	KernelSize  int
	Probability float64
}

// NewErosion creates erosion augmentation
func NewErosion(kernelSize int, probability float64) *Erosion {
	return &Erosion{KernelSize: kernelSize, Probability: probability}
}

// Apply applies erosion
func (e *Erosion) Apply(image [][]float64) [][]float64 {
	if rand.Float64() > e.Probability {
		return image
	}
	h := len(image)
	if h == 0 {
		return image
	}
	w := len(image[0])
	half := e.KernelSize / 2

	result := make([][]float64, h)
	for i := 0; i < h; i++ {
		result[i] = make([]float64, w)
		for j := 0; j < w; j++ {
			minVal := image[i][j]
			for ki := -half; ki <= half; ki++ {
				for kj := -half; kj <= half; kj++ {
					ni := i + ki
					nj := j + kj
					if ni >= 0 && ni < h && nj >= 0 && nj < w {
						if image[ni][nj] < minVal {
							minVal = image[ni][nj]
						}
					}
				}
			}
			result[i][j] = minVal
		}
	}
	return result
}

// Name returns the name
func (e *Erosion) Name() string {
	return "Erosion"
}

// Dilation applies morphological dilation
type Dilation struct {
	KernelSize  int
	Probability float64
}

// NewDilation creates dilation augmentation
func NewDilation(kernelSize int, probability float64) *Dilation {
	return &Dilation{KernelSize: kernelSize, Probability: probability}
}

// Apply applies dilation
func (d *Dilation) Apply(image [][]float64) [][]float64 {
	if rand.Float64() > d.Probability {
		return image
	}
	h := len(image)
	if h == 0 {
		return image
	}
	w := len(image[0])
	half := d.KernelSize / 2

	result := make([][]float64, h)
	for i := 0; i < h; i++ {
		result[i] = make([]float64, w)
		for j := 0; j < w; j++ {
			maxVal := image[i][j]
			for ki := -half; ki <= half; ki++ {
				for kj := -half; kj <= half; kj++ {
					ni := i + ki
					nj := j + kj
					if ni >= 0 && ni < h && nj >= 0 && nj < w {
						if image[ni][nj] > maxVal {
							maxVal = image[ni][nj]
						}
					}
				}
			}
			result[i][j] = maxVal
		}
	}
	return result
}

// Name returns the name
func (d *Dilation) Name() string {
	return "Dilation"
}

// ============================================================================
// Utility Functions
// ============================================================================

// copyImage creates a deep copy of a 2D image
func copyImage(image [][]float64) [][]float64 {
	h := len(image)
	if h == 0 {
		return image
	}
	result := make([][]float64, h)
	for i := 0; i < h; i++ {
		result[i] = make([]float64, len(image[i]))
		copy(result[i], image[i])
	}
	return result
}

// clamp restricts value to [minVal, maxVal]
func clamp(x, minVal, maxVal float64) float64 {
	if x < minVal {
		return minVal
	}
	if x > maxVal {
		return maxVal
	}
	return x
}

// ============================================================================
// Standard Augmentation Presets
// ============================================================================

// MNISTAugmentation returns standard augmentation for MNIST
func MNISTAugmentation() *AugmentationPipeline {
	return NewAugmentationPipeline(1.0).
		Add(NewTranslate(0.1, 0.1, 0, 0.5)).
		Add(NewScale(0.9, 1.1, 0.5)).
		Add(NewRotation(0.3, 90, 270)).
		Add(NewGaussianNoise(0, 0.05, 0.3)).
		Add(NewElasticTransform(5, 2, 0.3))
}

// CIFARAugmentation returns standard augmentation for CIFAR
func CIFARAugmentation() *AugmentationPipeline {
	return NewAugmentationPipeline(1.0).
		Add(NewHorizontalFlip(0.5)).
		Add(NewRandomResizedCrop(0.08, 1.0, 0.75, 1.33, 0.8)).
		Add(NewRandomBrightness(0.2, 0.5)).
		Add(NewRandomContrast(0.8, 1.2, 0.5)).
		Add(NewCutout(1, 8, 0, 0.5))
}

// CIMAugmentation returns CIM-specific augmentation for hardware robustness
func CIMAugmentation(adcBits int, noiseLevel float64) *AugmentationPipeline {
	return NewAugmentationPipeline(1.0).
		Add(NewCrossbarNoise(noiseLevel, noiseLevel*0.5, noiseLevel*0.3, noiseLevel, 0.8)).
		Add(NewQuantizationNoise(adcBits, true, 0.7)).
		Add(NewGaussianNoise(0, noiseLevel*0.5, 0.5))
}

// StrongAugmentation returns aggressive augmentation for regularization
func StrongAugmentation() *AugmentationPipeline {
	return NewAugmentationPipeline(1.0).
		Add(NewRandomChoice(
			NewHorizontalFlip(1.0),
			NewRotation(1.0, 90, 180, 270),
			NewInvert(1.0),
		)).
		Add(NewRandomResizedCrop(0.5, 1.0, 0.5, 2.0, 1.0)).
		Add(NewRandomChoice(
			NewRandomBrightness(0.4, 1.0),
			NewRandomContrast(0.5, 1.5, 1.0),
			NewGammaCorrection(0.5, 2.0, 1.0),
		)).
		Add(NewRandomChoice(
			NewCutout(2, 12, 0, 1.0),
			NewRandomErasing(0.02, 0.4, 0.3, 3.3, "random", 0, 1.0),
			NewGridMask(8, 0.5, 0, 1.0),
		))
}

// AugmentationRegistry provides named access to augmentations
var AugmentationRegistry = map[string]func() Augmentation{
	"horizontal_flip":   func() Augmentation { return NewHorizontalFlip(0.5) },
	"vertical_flip":     func() Augmentation { return NewVerticalFlip(0.5) },
	"rotation":          func() Augmentation { return NewRotation(0.5) },
	"random_crop":       func() Augmentation { return NewRandomCrop(24, 24, 0.5) },
	"center_crop":       func() Augmentation { return NewCenterCrop(24, 24) },
	"translate":         func() Augmentation { return NewTranslate(0.1, 0.1, 0, 0.5) },
	"scale":             func() Augmentation { return NewScale(0.8, 1.2, 0.5) },
	"elastic":           func() Augmentation { return NewElasticTransform(10, 3, 0.3) },
	"normalize":         func() Augmentation { return NewNormalize(0.5, 0.5) },
	"brightness":        func() Augmentation { return NewRandomBrightness(0.2, 0.5) },
	"contrast":          func() Augmentation { return NewRandomContrast(0.8, 1.2, 0.5) },
	"gamma":             func() Augmentation { return NewGammaCorrection(0.8, 1.2, 0.5) },
	"invert":            func() Augmentation { return NewInvert(0.1) },
	"gaussian_noise":    func() Augmentation { return NewGaussianNoise(0, 0.1, 0.5) },
	"salt_pepper":       func() Augmentation { return NewSaltAndPepperNoise(0.02, 0.5, 0.3) },
	"crossbar_noise":    func() Augmentation { return NewCrossbarNoise(0.02, 0.01, 0.01, 0.05, 0.5) },
	"quantization":      func() Augmentation { return NewQuantizationNoise(6, true, 0.5) },
	"cutout":            func() Augmentation { return NewCutout(1, 8, 0, 0.5) },
	"random_erasing":    func() Augmentation { return NewRandomErasing(0.02, 0.33, 0.3, 3.3, "constant", 0, 0.5) },
	"grid_mask":         func() Augmentation { return NewGridMask(8, 0.5, 0, 0.3) },
	"erosion":           func() Augmentation { return NewErosion(3, 0.3) },
	"dilation":          func() Augmentation { return NewDilation(3, 0.3) },
}
