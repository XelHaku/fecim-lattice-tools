// Package layers provides neural network layer implementations for CIM simulation.
// optimizer.go implements common optimizers for training neural networks.
package layers

import (
	"math"
)

// =============================================================================
// Optimizer Interface
// =============================================================================

// Optimizer defines the interface for optimization algorithms
type Optimizer interface {
	// Update modifies parameters based on gradients
	Update(params, grads []float64) []float64

	// Step increments the internal step counter
	Step()

	// GetLR returns the current learning rate
	GetLR() float64

	// SetLR sets the learning rate
	SetLR(lr float64)

	// Name returns the optimizer name
	Name() string
}

// =============================================================================
// SGD (Stochastic Gradient Descent)
// =============================================================================

// SGD implements vanilla stochastic gradient descent with optional momentum
type SGD struct {
	LR           float64   // Learning rate
	Momentum     float64   // Momentum factor (0 = no momentum)
	WeightDecay  float64   // L2 regularization
	Nesterov     bool      // Use Nesterov momentum
	Velocity     []float64 // Momentum buffer
	StepCount    int
}

// NewSGD creates a new SGD optimizer
func NewSGD(lr float64) *SGD {
	return &SGD{
		LR:          lr,
		Momentum:    0.0,
		WeightDecay: 0.0,
		Nesterov:    false,
	}
}

// NewSGDWithMomentum creates SGD with momentum
func NewSGDWithMomentum(lr, momentum float64) *SGD {
	return &SGD{
		LR:          lr,
		Momentum:    momentum,
		WeightDecay: 0.0,
		Nesterov:    false,
	}
}

// Update performs SGD update
func (o *SGD) Update(params, grads []float64) []float64 {
	n := len(params)
	result := make([]float64, n)

	// Initialize velocity if needed
	if o.Momentum > 0 && len(o.Velocity) != n {
		o.Velocity = make([]float64, n)
	}

	for i := 0; i < n; i++ {
		g := grads[i]

		// Apply weight decay
		if o.WeightDecay > 0 {
			g += o.WeightDecay * params[i]
		}

		if o.Momentum > 0 {
			// Update velocity
			o.Velocity[i] = o.Momentum*o.Velocity[i] + g

			if o.Nesterov {
				// Nesterov momentum
				g = g + o.Momentum*o.Velocity[i]
			} else {
				g = o.Velocity[i]
			}
		}

		result[i] = params[i] - o.LR*g
	}

	return result
}

func (o *SGD) Step()           { o.StepCount++ }
func (o *SGD) GetLR() float64  { return o.LR }
func (o *SGD) SetLR(lr float64) { o.LR = lr }
func (o *SGD) Name() string    { return "sgd" }

// =============================================================================
// Adam (Adaptive Moment Estimation)
// =============================================================================

// Adam implements the Adam optimizer
type Adam struct {
	LR          float64   // Learning rate
	Beta1       float64   // Exponential decay for 1st moment
	Beta2       float64   // Exponential decay for 2nd moment
	Epsilon     float64   // Numerical stability
	WeightDecay float64   // L2 regularization (decoupled)
	M           []float64 // First moment estimate
	V           []float64 // Second moment estimate
	StepCount   int
}

// NewAdam creates a new Adam optimizer with default parameters
func NewAdam(lr float64) *Adam {
	return &Adam{
		LR:          lr,
		Beta1:       0.9,
		Beta2:       0.999,
		Epsilon:     1e-8,
		WeightDecay: 0.0,
	}
}

// NewAdamW creates Adam with decoupled weight decay (AdamW)
func NewAdamW(lr, weightDecay float64) *Adam {
	return &Adam{
		LR:          lr,
		Beta1:       0.9,
		Beta2:       0.999,
		Epsilon:     1e-8,
		WeightDecay: weightDecay,
	}
}

// Update performs Adam update
func (o *Adam) Update(params, grads []float64) []float64 {
	n := len(params)
	result := make([]float64, n)

	// Initialize moment estimates if needed
	if len(o.M) != n {
		o.M = make([]float64, n)
		o.V = make([]float64, n)
	}

	t := float64(o.StepCount + 1)

	// Bias correction factors
	bc1 := 1.0 - math.Pow(o.Beta1, t)
	bc2 := 1.0 - math.Pow(o.Beta2, t)

	for i := 0; i < n; i++ {
		g := grads[i]

		// Update biased first moment estimate
		o.M[i] = o.Beta1*o.M[i] + (1-o.Beta1)*g

		// Update biased second moment estimate
		o.V[i] = o.Beta2*o.V[i] + (1-o.Beta2)*g*g

		// Bias-corrected estimates
		mHat := o.M[i] / bc1
		vHat := o.V[i] / bc2

		// Update parameters
		update := o.LR * mHat / (math.Sqrt(vHat) + o.Epsilon)

		// Apply decoupled weight decay (AdamW)
		if o.WeightDecay > 0 {
			result[i] = params[i] - update - o.LR*o.WeightDecay*params[i]
		} else {
			result[i] = params[i] - update
		}
	}

	return result
}

func (o *Adam) Step()           { o.StepCount++ }
func (o *Adam) GetLR() float64  { return o.LR }
func (o *Adam) SetLR(lr float64) { o.LR = lr }
func (o *Adam) Name() string    { return "adam" }

// =============================================================================
// RMSprop
// =============================================================================

// RMSprop implements the RMSprop optimizer
type RMSprop struct {
	LR          float64   // Learning rate
	Alpha       float64   // Smoothing constant
	Epsilon     float64   // Numerical stability
	WeightDecay float64   // L2 regularization
	Momentum    float64   // Momentum factor
	Centered    bool      // Use centered RMSprop
	V           []float64 // Running average of squared gradients
	G           []float64 // Running average of gradients (if centered)
	Buf         []float64 // Momentum buffer
	StepCount   int
}

// NewRMSprop creates a new RMSprop optimizer
func NewRMSprop(lr float64) *RMSprop {
	return &RMSprop{
		LR:       lr,
		Alpha:    0.99,
		Epsilon:  1e-8,
		Momentum: 0.0,
		Centered: false,
	}
}

// Update performs RMSprop update
func (o *RMSprop) Update(params, grads []float64) []float64 {
	n := len(params)
	result := make([]float64, n)

	// Initialize buffers if needed
	if len(o.V) != n {
		o.V = make([]float64, n)
		if o.Centered {
			o.G = make([]float64, n)
		}
		if o.Momentum > 0 {
			o.Buf = make([]float64, n)
		}
	}

	for i := 0; i < n; i++ {
		g := grads[i]

		// Apply weight decay
		if o.WeightDecay > 0 {
			g += o.WeightDecay * params[i]
		}

		// Update running average of squared gradients
		o.V[i] = o.Alpha*o.V[i] + (1-o.Alpha)*g*g

		var avg float64
		if o.Centered {
			// Update running average of gradients
			o.G[i] = o.Alpha*o.G[i] + (1-o.Alpha)*g
			avg = math.Sqrt(o.V[i]-o.G[i]*o.G[i]+o.Epsilon)
		} else {
			avg = math.Sqrt(o.V[i]) + o.Epsilon
		}

		if o.Momentum > 0 {
			o.Buf[i] = o.Momentum*o.Buf[i] + g/avg
			result[i] = params[i] - o.LR*o.Buf[i]
		} else {
			result[i] = params[i] - o.LR*g/avg
		}
	}

	return result
}

func (o *RMSprop) Step()           { o.StepCount++ }
func (o *RMSprop) GetLR() float64  { return o.LR }
func (o *RMSprop) SetLR(lr float64) { o.LR = lr }
func (o *RMSprop) Name() string    { return "rmsprop" }

// =============================================================================
// Adagrad
// =============================================================================

// Adagrad implements the Adagrad optimizer
type Adagrad struct {
	LR          float64   // Learning rate
	Epsilon     float64   // Numerical stability
	WeightDecay float64   // L2 regularization
	SumSq       []float64 // Sum of squared gradients
	StepCount   int
}

// NewAdagrad creates a new Adagrad optimizer
func NewAdagrad(lr float64) *Adagrad {
	return &Adagrad{
		LR:      lr,
		Epsilon: 1e-10,
	}
}

// Update performs Adagrad update
func (o *Adagrad) Update(params, grads []float64) []float64 {
	n := len(params)
	result := make([]float64, n)

	// Initialize sum of squares if needed
	if len(o.SumSq) != n {
		o.SumSq = make([]float64, n)
	}

	for i := 0; i < n; i++ {
		g := grads[i]

		// Apply weight decay
		if o.WeightDecay > 0 {
			g += o.WeightDecay * params[i]
		}

		// Accumulate squared gradients
		o.SumSq[i] += g * g

		// Update parameters
		result[i] = params[i] - o.LR*g/(math.Sqrt(o.SumSq[i])+o.Epsilon)
	}

	return result
}

func (o *Adagrad) Step()           { o.StepCount++ }
func (o *Adagrad) GetLR() float64  { return o.LR }
func (o *Adagrad) SetLR(lr float64) { o.LR = lr }
func (o *Adagrad) Name() string    { return "adagrad" }

// =============================================================================
// Adadelta
// =============================================================================

// Adadelta implements the Adadelta optimizer (no learning rate required)
type Adadelta struct {
	Rho         float64   // Decay rate
	Epsilon     float64   // Numerical stability
	WeightDecay float64   // L2 regularization
	AvgSqGrad   []float64 // Running average of squared gradients
	AvgSqDelta  []float64 // Running average of squared updates
	StepCount   int
}

// NewAdadelta creates a new Adadelta optimizer
func NewAdadelta() *Adadelta {
	return &Adadelta{
		Rho:     0.9,
		Epsilon: 1e-6,
	}
}

// Update performs Adadelta update
func (o *Adadelta) Update(params, grads []float64) []float64 {
	n := len(params)
	result := make([]float64, n)

	// Initialize buffers if needed
	if len(o.AvgSqGrad) != n {
		o.AvgSqGrad = make([]float64, n)
		o.AvgSqDelta = make([]float64, n)
	}

	for i := 0; i < n; i++ {
		g := grads[i]

		// Apply weight decay
		if o.WeightDecay > 0 {
			g += o.WeightDecay * params[i]
		}

		// Update running average of squared gradients
		o.AvgSqGrad[i] = o.Rho*o.AvgSqGrad[i] + (1-o.Rho)*g*g

		// Compute update
		rmsGrad := math.Sqrt(o.AvgSqGrad[i] + o.Epsilon)
		rmsDelta := math.Sqrt(o.AvgSqDelta[i] + o.Epsilon)
		delta := (rmsDelta / rmsGrad) * g

		// Update running average of squared updates
		o.AvgSqDelta[i] = o.Rho*o.AvgSqDelta[i] + (1-o.Rho)*delta*delta

		result[i] = params[i] - delta
	}

	return result
}

func (o *Adadelta) Step()           { o.StepCount++ }
func (o *Adadelta) GetLR() float64  { return 1.0 } // Adadelta doesn't use LR
func (o *Adadelta) SetLR(lr float64) {}            // No-op
func (o *Adadelta) Name() string    { return "adadelta" }

// =============================================================================
// AdaMax
// =============================================================================

// AdaMax implements the AdaMax optimizer (Adam with infinity norm)
type AdaMax struct {
	LR          float64   // Learning rate
	Beta1       float64   // Exponential decay for 1st moment
	Beta2       float64   // Exponential decay for infinity norm
	Epsilon     float64   // Numerical stability
	WeightDecay float64   // L2 regularization
	M           []float64 // First moment estimate
	U           []float64 // Infinity norm
	StepCount   int
}

// NewAdaMax creates a new AdaMax optimizer
func NewAdaMax(lr float64) *AdaMax {
	return &AdaMax{
		LR:      lr,
		Beta1:   0.9,
		Beta2:   0.999,
		Epsilon: 1e-8,
	}
}

// Update performs AdaMax update
func (o *AdaMax) Update(params, grads []float64) []float64 {
	n := len(params)
	result := make([]float64, n)

	// Initialize buffers if needed
	if len(o.M) != n {
		o.M = make([]float64, n)
		o.U = make([]float64, n)
	}

	t := float64(o.StepCount + 1)
	bc1 := 1.0 - math.Pow(o.Beta1, t)

	for i := 0; i < n; i++ {
		g := grads[i]

		// Apply weight decay
		if o.WeightDecay > 0 {
			g += o.WeightDecay * params[i]
		}

		// Update biased first moment estimate
		o.M[i] = o.Beta1*o.M[i] + (1-o.Beta1)*g

		// Update infinity norm
		o.U[i] = math.Max(o.Beta2*o.U[i], math.Abs(g))

		// Bias-corrected first moment
		mHat := o.M[i] / bc1

		// Update parameters
		result[i] = params[i] - o.LR*mHat/(o.U[i]+o.Epsilon)
	}

	return result
}

func (o *AdaMax) Step()           { o.StepCount++ }
func (o *AdaMax) GetLR() float64  { return o.LR }
func (o *AdaMax) SetLR(lr float64) { o.LR = lr }
func (o *AdaMax) Name() string    { return "adamax" }

// =============================================================================
// NAdam (Nesterov Adam)
// =============================================================================

// NAdam implements Nesterov-accelerated Adam
type NAdam struct {
	LR          float64   // Learning rate
	Beta1       float64   // Exponential decay for 1st moment
	Beta2       float64   // Exponential decay for 2nd moment
	Epsilon     float64   // Numerical stability
	WeightDecay float64   // L2 regularization
	M           []float64 // First moment estimate
	V           []float64 // Second moment estimate
	StepCount   int
}

// NewNAdam creates a new NAdam optimizer
func NewNAdam(lr float64) *NAdam {
	return &NAdam{
		LR:      lr,
		Beta1:   0.9,
		Beta2:   0.999,
		Epsilon: 1e-8,
	}
}

// Update performs NAdam update
func (o *NAdam) Update(params, grads []float64) []float64 {
	n := len(params)
	result := make([]float64, n)

	// Initialize buffers if needed
	if len(o.M) != n {
		o.M = make([]float64, n)
		o.V = make([]float64, n)
	}

	t := float64(o.StepCount + 1)
	bc1 := 1.0 - math.Pow(o.Beta1, t)
	bc2 := 1.0 - math.Pow(o.Beta2, t)

	for i := 0; i < n; i++ {
		g := grads[i]

		// Apply weight decay
		if o.WeightDecay > 0 {
			g += o.WeightDecay * params[i]
		}

		// Update biased first moment estimate
		o.M[i] = o.Beta1*o.M[i] + (1-o.Beta1)*g

		// Update biased second moment estimate
		o.V[i] = o.Beta2*o.V[i] + (1-o.Beta2)*g*g

		// Bias-corrected estimates
		mHat := o.M[i] / bc1
		vHat := o.V[i] / bc2

		// Nesterov momentum: look ahead using gradient
		mNesterov := o.Beta1*mHat + (1-o.Beta1)*g/bc1

		// Update parameters
		result[i] = params[i] - o.LR*mNesterov/(math.Sqrt(vHat)+o.Epsilon)
	}

	return result
}

func (o *NAdam) Step()           { o.StepCount++ }
func (o *NAdam) GetLR() float64  { return o.LR }
func (o *NAdam) SetLR(lr float64) { o.LR = lr }
func (o *NAdam) Name() string    { return "nadam" }

// =============================================================================
// Learning Rate Schedulers
// =============================================================================

// LRScheduler defines the interface for learning rate schedulers
type LRScheduler interface {
	// Step updates the learning rate
	Step(optimizer Optimizer)

	// GetLR returns the current scheduled learning rate
	GetLR() float64
}

// StepLR decreases LR by gamma every step_size epochs
type StepLR struct {
	BaseLR    float64
	StepSize  int
	Gamma     float64
	StepCount int
}

// NewStepLR creates a step learning rate scheduler
func NewStepLR(baseLR float64, stepSize int, gamma float64) *StepLR {
	return &StepLR{
		BaseLR:   baseLR,
		StepSize: stepSize,
		Gamma:    gamma,
	}
}

// Step updates the optimizer's learning rate
func (s *StepLR) Step(optimizer Optimizer) {
	s.StepCount++
	optimizer.SetLR(s.GetLR())
}

// GetLR returns the current learning rate
func (s *StepLR) GetLR() float64 {
	numDecays := s.StepCount / s.StepSize
	return s.BaseLR * math.Pow(s.Gamma, float64(numDecays))
}

// ExponentialLR decreases LR by gamma every epoch
type ExponentialLR struct {
	BaseLR    float64
	Gamma     float64
	StepCount int
}

// NewExponentialLR creates an exponential learning rate scheduler
func NewExponentialLR(baseLR, gamma float64) *ExponentialLR {
	return &ExponentialLR{
		BaseLR: baseLR,
		Gamma:  gamma,
	}
}

// Step updates the learning rate
func (s *ExponentialLR) Step(optimizer Optimizer) {
	s.StepCount++
	optimizer.SetLR(s.GetLR())
}

// GetLR returns the current learning rate
func (s *ExponentialLR) GetLR() float64 {
	return s.BaseLR * math.Pow(s.Gamma, float64(s.StepCount))
}

// CosineAnnealingLR implements cosine annealing schedule
type CosineAnnealingLR struct {
	BaseLR    float64
	MinLR     float64
	TMax      int // Maximum number of iterations
	StepCount int
}

// NewCosineAnnealingLR creates a cosine annealing scheduler
func NewCosineAnnealingLR(baseLR, minLR float64, tMax int) *CosineAnnealingLR {
	return &CosineAnnealingLR{
		BaseLR: baseLR,
		MinLR:  minLR,
		TMax:   tMax,
	}
}

// Step updates the learning rate
func (s *CosineAnnealingLR) Step(optimizer Optimizer) {
	s.StepCount++
	optimizer.SetLR(s.GetLR())
}

// GetLR returns the current learning rate
func (s *CosineAnnealingLR) GetLR() float64 {
	t := s.StepCount % s.TMax
	return s.MinLR + 0.5*(s.BaseLR-s.MinLR)*(1+math.Cos(math.Pi*float64(t)/float64(s.TMax)))
}

// WarmupLR implements linear warmup followed by constant LR
type WarmupLR struct {
	BaseLR      float64
	WarmupSteps int
	StepCount   int
}

// NewWarmupLR creates a warmup scheduler
func NewWarmupLR(baseLR float64, warmupSteps int) *WarmupLR {
	return &WarmupLR{
		BaseLR:      baseLR,
		WarmupSteps: warmupSteps,
	}
}

// Step updates the learning rate
func (s *WarmupLR) Step(optimizer Optimizer) {
	s.StepCount++
	optimizer.SetLR(s.GetLR())
}

// GetLR returns the current learning rate
func (s *WarmupLR) GetLR() float64 {
	if s.StepCount < s.WarmupSteps {
		return s.BaseLR * float64(s.StepCount) / float64(s.WarmupSteps)
	}
	return s.BaseLR
}

// =============================================================================
// Optimizer Registry
// =============================================================================

// OptimizerRegistry provides name-based optimizer lookup
var OptimizerRegistry = map[string]func(lr float64) Optimizer{
	"sgd":      func(lr float64) Optimizer { return NewSGD(lr) },
	"adam":     func(lr float64) Optimizer { return NewAdam(lr) },
	"adamw":    func(lr float64) Optimizer { return NewAdamW(lr, 0.01) },
	"rmsprop":  func(lr float64) Optimizer { return NewRMSprop(lr) },
	"adagrad":  func(lr float64) Optimizer { return NewAdagrad(lr) },
	"adadelta": func(lr float64) Optimizer { return NewAdadelta() },
	"adamax":   func(lr float64) Optimizer { return NewAdaMax(lr) },
	"nadam":    func(lr float64) Optimizer { return NewNAdam(lr) },
}

// GetOptimizer returns an optimizer by name
func GetOptimizer(name string, lr float64) Optimizer {
	if creator, ok := OptimizerRegistry[name]; ok {
		return creator(lr)
	}
	return NewSGD(lr) // Default to SGD
}

// =============================================================================
// Hardware-Aware Optimizer Extensions
// =============================================================================

// QuantizedSGD implements SGD with quantized weight updates for CIM
type QuantizedSGD struct {
	*SGD
	WeightBits  int     // Weight precision
	UpdateBits  int     // Update precision
	ScaleFactor float64 // Scaling for quantization
}

// NewQuantizedSGD creates a quantized SGD optimizer
func NewQuantizedSGD(lr float64, weightBits, updateBits int) *QuantizedSGD {
	return &QuantizedSGD{
		SGD:         NewSGD(lr),
		WeightBits:  weightBits,
		UpdateBits:  updateBits,
		ScaleFactor: math.Pow(2, float64(weightBits-1)) - 1,
	}
}

// Update performs quantized SGD update
func (o *QuantizedSGD) Update(params, grads []float64) []float64 {
	// Get standard SGD update
	result := o.SGD.Update(params, grads)

	// Quantize weights
	for i := range result {
		// Quantize to weight precision
		result[i] = o.quantize(result[i])
	}

	return result
}

func (o *QuantizedSGD) quantize(x float64) float64 {
	// Symmetric quantization
	maxVal := 1.0
	x = math.Max(-maxVal, math.Min(maxVal, x))
	scale := o.ScaleFactor / maxVal
	return math.Round(x*scale) / scale
}

func (o *QuantizedSGD) Name() string { return "quantized_sgd" }
