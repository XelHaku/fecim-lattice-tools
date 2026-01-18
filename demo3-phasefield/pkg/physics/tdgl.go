// Package physics provides TDGL phase-field physics models for ferroelectrics.
//
// The Time-Dependent Ginzburg-Landau (TDGL) equation governs domain evolution:
//
//   ∂P/∂t = -L · δF/δP
//
// where F is the Landau-Ginzburg free energy functional:
//
//   F = ∫[α·P² + β·P⁴ + γ·P⁶ + κ|∇P|² - E·P] dV
//
// This is the continuum limit of the Landau-Khalatnikov equation (Eq. 3):
//   E = (∂U/∂P)_S + ρ(dP/dt)
//
// Reference: Sivasubramanian & Widom, arXiv:cond-mat/0108189v1 (2001)
package physics

import (
	"fmt"
	"math"
)

// Grid3D represents a 3D computational grid for phase-field simulation.
type Grid3D struct {
	Nx, Ny, Nz int       // Grid dimensions
	dx, dy, dz float64   // Grid spacing (m)
	P          []float64 // Polarization field (flattened)
	Pnew       []float64 // Updated polarization (double buffering)
}

// NewGrid3D creates a new 3D grid with specified dimensions.
func NewGrid3D(nx, ny, nz int, dx float64) *Grid3D {
	size := nx * ny * nz
	return &Grid3D{
		Nx:   nx,
		Ny:   ny,
		Nz:   nz,
		dx:   dx,
		dy:   dx,
		dz:   dx,
		P:    make([]float64, size),
		Pnew: make([]float64, size),
	}
}

// Index converts 3D coordinates to linear index.
func (g *Grid3D) Index(x, y, z int) int {
	return z*g.Ny*g.Nx + y*g.Nx + x
}

// Coords converts linear index to 3D coordinates.
func (g *Grid3D) Coords(idx int) (x, y, z int) {
	z = idx / (g.Ny * g.Nx)
	idx = idx % (g.Ny * g.Nx)
	y = idx / g.Nx
	x = idx % g.Nx
	return
}

// Get returns polarization at (x, y, z) with periodic boundary conditions.
func (g *Grid3D) Get(x, y, z int) float64 {
	// Periodic boundaries
	x = ((x % g.Nx) + g.Nx) % g.Nx
	y = ((y % g.Ny) + g.Ny) % g.Ny
	z = ((z % g.Nz) + g.Nz) % g.Nz
	return g.P[g.Index(x, y, z)]
}

// Set sets polarization at (x, y, z).
func (g *Grid3D) Set(x, y, z int, val float64) {
	g.P[g.Index(x, y, z)] = val
}

// Size returns total number of grid points.
func (g *Grid3D) Size() int {
	return g.Nx * g.Ny * g.Nz
}

// TDGLSolver implements the Time-Dependent Ginzburg-Landau equation solver.
type TDGLSolver struct {
	grid     *Grid3D
	material *HZOMaterial

	// Simulation parameters
	dt   float64 // Time step (s)
	time float64 // Current simulation time (s)
	step int64   // Current step number

	// External conditions
	Temperature  float64 // T (K)
	ExternalField float64 // E_ext (V/m)

	// Effective Landau coefficient at current temperature
	alphaEff float64
}

// NewTDGLSolver creates a new TDGL solver with given grid and material.
func NewTDGLSolver(grid *Grid3D, material *HZOMaterial) *TDGLSolver {
	solver := &TDGLSolver{
		grid:        grid,
		material:    material,
		dt:          1e-12, // 1 ps default
		Temperature: 300,    // Room temperature
	}
	solver.updateAlpha()
	return solver
}

// updateAlpha recalculates the effective α from temperature.
func (s *TDGLSolver) updateAlpha() {
	s.alphaEff = s.material.AlphaTemperature(s.Temperature)
}

// SetTemperature updates the simulation temperature.
func (s *TDGLSolver) SetTemperature(T float64) {
	s.Temperature = T
	s.updateAlpha()
}

// SetTimeStep sets the integration time step.
func (s *TDGLSolver) SetTimeStep(dt float64) {
	s.dt = dt
}

// SetExternalField sets the applied electric field.
func (s *TDGLSolver) SetExternalField(E float64) {
	s.ExternalField = E
}

// InitializeRandom initializes the grid with random polarization.
func (s *TDGLSolver) InitializeRandom(amplitude float64) {
	for i := range s.grid.P {
		// Random polarization centered at zero
		s.grid.P[i] = amplitude * (2*randFloat() - 1)
	}
}

// InitializeUniform initializes the grid with uniform polarization.
func (s *TDGLSolver) InitializeUniform(P0 float64) {
	for i := range s.grid.P {
		s.grid.P[i] = P0
	}
}

// InitializeDomainPattern creates a simple domain pattern.
func (s *TDGLSolver) InitializeDomainPattern(Ps float64) {
	for z := 0; z < s.grid.Nz; z++ {
		for y := 0; y < s.grid.Ny; y++ {
			for x := 0; x < s.grid.Nx; x++ {
				// Stripe pattern alternating along x
				if (x/8)%2 == 0 {
					s.grid.Set(x, y, z, Ps)
				} else {
					s.grid.Set(x, y, z, -Ps)
				}
			}
		}
	}
}

// Step advances the simulation by one time step using forward Euler.
//
// Implements the TDGL equation:
//   P_new = P - dt·L·(dF/dP)
//
// where dF/dP = 2αP + 4βP³ + 6γP⁵ - κ∇²P - E_ext
//
// This is the discrete form of Eq. (18) from Sivasubramanian & Widom (2001):
//   dy/dθ + ηy(y² - 1) = z·cos(θ)
//
// The 6-point Laplacian stencil provides O(dx²) accuracy.
func (s *TDGLSolver) Step() {
	g := s.grid
	m := s.material
	dt := s.dt
	L := m.L
	kappa := m.Kappa
	alpha := s.alphaEff
	beta := m.Beta
	gamma := m.Gamma
	Eext := s.ExternalField
	dx2 := g.dx * g.dx

	// Compute updated polarization
	for z := 0; z < g.Nz; z++ {
		for y := 0; y < g.Ny; y++ {
			for x := 0; x < g.Nx; x++ {
				idx := g.Index(x, y, z)
				P := g.P[idx]

				// Laplacian using 6-point stencil with periodic BC
				Pxp := g.Get(x+1, y, z)
				Pxm := g.Get(x-1, y, z)
				Pyp := g.Get(x, y+1, z)
				Pym := g.Get(x, y-1, z)
				Pzp := g.Get(x, y, z+1)
				Pzm := g.Get(x, y, z-1)

				laplacian := (Pxp + Pxm + Pyp + Pym + Pzp + Pzm - 6*P) / dx2

				// Free energy derivative: dF/dP
				// dF/dP = 2αP + 4βP³ + 6γP⁵ - κ∇²P - E_ext
				P2 := P * P
				P3 := P2 * P
				P5 := P3 * P2

				dFdP := 2*alpha*P + 4*beta*P3 + 6*gamma*P5 - kappa*laplacian - Eext

				// TDGL time evolution: dP/dt = -L * dF/dP
				Pnew := P - dt*L*dFdP

				// Store in buffer
				g.Pnew[idx] = Pnew
			}
		}
	}

	// Swap buffers
	g.P, g.Pnew = g.Pnew, g.P

	s.time += dt
	s.step++
}

// StepMultiple advances the simulation by n time steps.
func (s *TDGLSolver) StepMultiple(n int) {
	for i := 0; i < n; i++ {
		s.Step()
	}
}

// GetTime returns the current simulation time.
func (s *TDGLSolver) GetTime() float64 {
	return s.time
}

// GetStep returns the current step number.
func (s *TDGLSolver) GetStep() int64 {
	return s.step
}

// ComputeFreeEnergy calculates the total Landau-Ginzburg free energy.
func (s *TDGLSolver) ComputeFreeEnergy() float64 {
	g := s.grid
	m := s.material
	alpha := s.alphaEff
	beta := m.Beta
	gamma := m.Gamma
	kappa := m.Kappa
	dx := g.dx
	_ = kappa // Used in gradient calculation below

	var totalEnergy float64
	cellVolume := dx * dx * dx

	for z := 0; z < g.Nz; z++ {
		for y := 0; y < g.Ny; y++ {
			for x := 0; x < g.Nx; x++ {
				P := g.Get(x, y, z)
				P2 := P * P
				P4 := P2 * P2
				P6 := P4 * P2

				// Landau energy density: α P² + β P⁴ + γ P⁶
				landauEnergy := alpha*P2 + beta*P4 + gamma*P6

				// Gradient energy density: κ |∇P|²
				Pxp := g.Get(x+1, y, z)
				Pxm := g.Get(x-1, y, z)
				Pyp := g.Get(x, y+1, z)
				Pym := g.Get(x, y-1, z)
				Pzp := g.Get(x, y, z+1)
				Pzm := g.Get(x, y, z-1)

				dPdx := (Pxp - Pxm) / (2 * dx)
				dPdy := (Pyp - Pym) / (2 * dx)
				dPdz := (Pzp - Pzm) / (2 * dx)

				gradEnergy := 0.5 * kappa * (dPdx*dPdx + dPdy*dPdy + dPdz*dPdz)

				// Electric energy: -E * P
				elecEnergy := -s.ExternalField * P

				totalEnergy += (landauEnergy + gradEnergy + elecEnergy) * cellVolume
			}
		}
	}

	return totalEnergy
}

// ComputeAveragePolarization returns the volume-averaged polarization.
func (s *TDGLSolver) ComputeAveragePolarization() float64 {
	var sum float64
	for _, p := range s.grid.P {
		sum += p
	}
	return sum / float64(len(s.grid.P))
}

// ComputeDomainFraction returns fraction of positive vs negative domains.
func (s *TDGLSolver) ComputeDomainFraction() (positive, negative float64) {
	var posCount, negCount int
	for _, p := range s.grid.P {
		if p > 0 {
			posCount++
		} else {
			negCount++
		}
	}
	total := float64(len(s.grid.P))
	return float64(posCount) / total, float64(negCount) / total
}

// GetPolarizationSlice returns a 2D slice of the polarization field at z=zSlice.
func (s *TDGLSolver) GetPolarizationSlice(zSlice int) [][]float64 {
	g := s.grid
	slice := make([][]float64, g.Ny)
	for y := 0; y < g.Ny; y++ {
		slice[y] = make([]float64, g.Nx)
		for x := 0; x < g.Nx; x++ {
			slice[y][x] = g.Get(x, y, zSlice)
		}
	}
	return slice
}

// Stats returns simulation statistics.
func (s *TDGLSolver) Stats() string {
	avgP := s.ComputeAveragePolarization()
	posF, _ := s.ComputeDomainFraction()
	energy := s.ComputeFreeEnergy()

	return fmt.Sprintf(
		"Step: %d, Time: %.3e s, <P>: %.4e, +domain: %.1f%%, Energy: %.3e",
		s.step, s.time, avgP, posF*100, energy,
	)
}

// randFloat returns a random float in [0, 1).
// Simple linear congruential generator for reproducibility.
var randState uint64 = 12345

func randFloat() float64 {
	randState = randState*6364136223846793005 + 1442695040888963407
	return float64(randState>>33) / float64(1<<31)
}

// SetRandomSeed sets the random seed for reproducibility.
func SetRandomSeed(seed uint64) {
	randState = seed
}

// StabilityLimit returns the maximum stable time step based on grid spacing.
// For explicit Euler on diffusion: dt < dx² / (2 * D * dim)
// where D ≈ L * κ for TDGL
func (s *TDGLSolver) StabilityLimit() float64 {
	D := s.material.L * s.material.Kappa
	return s.grid.dx * s.grid.dx / (6 * D) * 0.5 // Safety factor
}

// AutoSetTimeStep sets time step based on stability analysis.
func (s *TDGLSolver) AutoSetTimeStep() {
	limit := s.StabilityLimit()
	s.dt = math.Min(limit, 1e-12) // At most 1 ps
}
