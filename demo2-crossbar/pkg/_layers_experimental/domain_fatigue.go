// Package layers provides domain engineering and wake-up/fatigue simulation
// for HZO-based ferroelectric CIM devices.
//
// Based on research:
// - Polarization switching pathways (47 in orthorhombic HfO2)
// - Wake-up effect via oxygen vacancy redistribution
// - Fatigue mechanisms (degradation-type, breakdown-type)
// - Fatigue-free designs (CeO2-x oxygen sponge, superlattices)
//
// References:
// - Zhang et al., "Phase Transformation Driven by Oxygen Vacancy Redistribution"
//   Advanced Electronic Materials 2024
// - Jan et al., "Resetting the Drift of Oxygen Vacancies" Small Science 2024
// - Nature Communications 2025, "Fatigue-free ferroelectricity via interfacial design"
package layers

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sort"
)

// ============================================================================
// FERROELECTRIC DOMAIN MODELING
// ============================================================================

// DomainOrientation represents ferroelectric domain polarization direction
type DomainOrientation int

const (
	DomainUp       DomainOrientation = iota // Polarization pointing up (+z)
	DomainDown                               // Polarization pointing down (-z)
	DomainInPlane0                           // In-plane 0°
	DomainInPlane90                          // In-plane 90°
	DomainInPlane180                         // In-plane 180°
	DomainInPlane270                         // In-plane 270°
)

// SwitchingPathway represents a polarization reversal pathway
type SwitchingPathway struct {
	PathwayID     int               // Unique identifier (0-46 for HfO2)
	InitialState  DomainOrientation // Starting polarization
	FinalState    DomainOrientation // Ending polarization
	AngleChange   float64           // Degrees (180°, 90°, etc.)
	EnergyBarrier float64           // eV (0.32-0.57 for unconventional paths)
	Description   string            // Path type description
}

// SwitchingPathwayDatabase contains all 47 HfO2 switching pathways
type SwitchingPathwayDatabase struct {
	Pathways          []SwitchingPathway
	Material          string
	LatticeParameters struct {
		A float64 // Angstrom
		B float64 // Angstrom
		C float64 // Angstrom
	}
}

// NewHfO2PathwayDatabase creates the complete HfO2 switching pathway database
func NewHfO2PathwayDatabase() *SwitchingPathwayDatabase {
	db := &SwitchingPathwayDatabase{
		Material: "HfO2-orthorhombic",
		Pathways: make([]SwitchingPathway, 47),
	}

	// Lattice parameters for orthorhombic Pca21 phase
	db.LatticeParameters.A = 5.25
	db.LatticeParameters.B = 5.05
	db.LatticeParameters.C = 5.07

	// Generate pathways based on group theory analysis
	// 180° reversal pathways (primary)
	for i := 0; i < 8; i++ {
		db.Pathways[i] = SwitchingPathway{
			PathwayID:     i,
			InitialState:  DomainUp,
			FinalState:    DomainDown,
			AngleChange:   180.0,
			EnergyBarrier: 0.32 + float64(i)*0.03, // 0.32-0.56 eV range
			Description:   fmt.Sprintf("180° reversal via suboxygen shift %d", i),
		}
	}

	// 90° switching pathways
	for i := 8; i < 24; i++ {
		db.Pathways[i] = SwitchingPathway{
			PathwayID:     i,
			InitialState:  DomainUp,
			FinalState:    DomainInPlane0 + DomainOrientation((i-8)%4),
			AngleChange:   90.0,
			EnergyBarrier: 0.35 + float64(i-8)*0.015,
			Description:   fmt.Sprintf("90° rotation pathway %d", i-7),
		}
	}

	// Complex multi-step pathways
	for i := 24; i < 47; i++ {
		db.Pathways[i] = SwitchingPathway{
			PathwayID:     i,
			InitialState:  DomainUp,
			FinalState:    DomainDown,
			AngleChange:   180.0,
			EnergyBarrier: 0.40 + float64(i-24)*0.007,
			Description:   fmt.Sprintf("Multi-step pathway via intermediate states %d", i-23),
		}
	}

	return db
}

// FindOptimalPathway finds the lowest energy switching pathway
func (db *SwitchingPathwayDatabase) FindOptimalPathway(initial, final DomainOrientation) *SwitchingPathway {
	var optimal *SwitchingPathway
	lowestEnergy := math.MaxFloat64

	for i := range db.Pathways {
		p := &db.Pathways[i]
		if p.InitialState == initial && p.FinalState == final {
			if p.EnergyBarrier < lowestEnergy {
				lowestEnergy = p.EnergyBarrier
				optimal = p
			}
		}
	}

	return optimal
}

// FerroelectricDomain represents a single ferroelectric domain
type FerroelectricDomain struct {
	ID           int
	CenterX      float64           // nm
	CenterY      float64           // nm
	Radius       float64           // nm (equivalent circular radius)
	Orientation  DomainOrientation
	Polarization float64           // μC/cm²
	Pinned       bool              // Whether domain is pinned by defects
	WallEnergy   float64           // J/m² domain wall energy
}

// DomainWall represents the boundary between two domains
type DomainWall struct {
	Domain1ID    int
	Domain2ID    int
	Length       float64 // nm
	Width        float64 // nm (typical 1-5 nm for HfO2)
	WallType     string  // "180°", "90°", "head-to-head", "tail-to-tail"
	MobilityCoeff float64 // Domain wall mobility coefficient
	PinningDensity float64 // Pinning sites per nm
}

// DomainStructure represents the complete domain configuration
type DomainStructure struct {
	Domains       []*FerroelectricDomain
	Walls         []*DomainWall
	FilmThickness float64 // nm
	FilmArea      float64 // nm²
	AverageGrainSize float64 // nm (10-20 nm typical for HZO)
}

// NewDomainStructure creates a multi-domain ferroelectric structure
func NewDomainStructure(thickness, area float64, grainSize float64, numDomains int) *DomainStructure {
	ds := &DomainStructure{
		Domains:          make([]*FerroelectricDomain, numDomains),
		FilmThickness:    thickness,
		FilmArea:         area,
		AverageGrainSize: grainSize,
	}

	// Create random domain distribution
	sqrtArea := math.Sqrt(area)
	for i := 0; i < numDomains; i++ {
		orientation := DomainUp
		if rand.Float64() < 0.5 {
			orientation = DomainDown
		}

		ds.Domains[i] = &FerroelectricDomain{
			ID:           i,
			CenterX:      rand.Float64() * sqrtArea,
			CenterY:      rand.Float64() * sqrtArea,
			Radius:       (0.5 + rand.Float64()) * grainSize / 2,
			Orientation:  orientation,
			Polarization: 20.0 + rand.Float64()*20.0, // 20-40 μC/cm²
			Pinned:       rand.Float64() < 0.2,       // 20% initially pinned
			WallEnergy:   0.1 + rand.Float64()*0.1,   // 0.1-0.2 J/m²
		}
	}

	// Create domain walls
	ds.calculateDomainWalls()

	return ds
}

// calculateDomainWalls identifies and creates domain wall structures
func (ds *DomainStructure) calculateDomainWalls() {
	ds.Walls = make([]*DomainWall, 0)

	for i := 0; i < len(ds.Domains); i++ {
		for j := i + 1; j < len(ds.Domains); j++ {
			d1 := ds.Domains[i]
			d2 := ds.Domains[j]

			// Check if domains are adjacent
			distance := math.Sqrt(math.Pow(d1.CenterX-d2.CenterX, 2) +
				math.Pow(d1.CenterY-d2.CenterY, 2))

			if distance < d1.Radius+d2.Radius+5.0 { // Within 5nm proximity
				wallType := "180°"
				if d1.Orientation == d2.Orientation {
					continue // No wall needed
				}
				if (d1.Orientation == DomainUp && d2.Orientation == DomainDown) ||
					(d1.Orientation == DomainDown && d2.Orientation == DomainUp) {
					wallType = "180°"
				} else {
					wallType = "90°"
				}

				wall := &DomainWall{
					Domain1ID:      d1.ID,
					Domain2ID:      d2.ID,
					Length:         2.0 * math.Min(d1.Radius, d2.Radius),
					Width:          2.0, // ~2 nm typical for HfO2
					WallType:       wallType,
					MobilityCoeff:  1e-7, // m²/Vs
					PinningDensity: 0.1 + rand.Float64()*0.2,
				}
				ds.Walls = append(ds.Walls, wall)
			}
		}
	}
}

// GetNetPolarization calculates the net polarization of the structure
func (ds *DomainStructure) GetNetPolarization() float64 {
	var netP float64
	var totalArea float64

	for _, d := range ds.Domains {
		area := math.Pi * d.Radius * d.Radius
		switch d.Orientation {
		case DomainUp:
			netP += d.Polarization * area
		case DomainDown:
			netP -= d.Polarization * area
		}
		totalArea += area
	}

	if totalArea > 0 {
		return netP / totalArea
	}
	return 0
}

// SwitchDomains simulates domain switching under applied field
func (ds *DomainStructure) SwitchDomains(electricField float64, pathwayDB *SwitchingPathwayDatabase) int {
	switchedCount := 0
	kT := 0.026 // Thermal energy at room temperature (eV)

	for _, d := range ds.Domains {
		if d.Pinned {
			continue
		}

		// Determine if field favors switching
		shouldSwitch := false
		var newOrientation DomainOrientation

		if electricField > 0 && d.Orientation == DomainDown {
			shouldSwitch = true
			newOrientation = DomainUp
		} else if electricField < 0 && d.Orientation == DomainUp {
			shouldSwitch = true
			newOrientation = DomainDown
		}

		if shouldSwitch {
			// Find switching pathway
			pathway := pathwayDB.FindOptimalPathway(d.Orientation, newOrientation)
			if pathway == nil {
				continue
			}

			// Calculate switching probability (thermally activated)
			fieldEnergy := math.Abs(electricField) * d.Polarization * 1e-6 // Convert units
			effectiveBarrier := pathway.EnergyBarrier - fieldEnergy
			if effectiveBarrier < 0 {
				effectiveBarrier = 0
			}

			probability := math.Exp(-effectiveBarrier / kT)

			if rand.Float64() < probability {
				d.Orientation = newOrientation
				switchedCount++
			}
		}
	}

	ds.calculateDomainWalls()
	return switchedCount
}

// ============================================================================
// OXYGEN VACANCY MODELING
// ============================================================================

// OxygenVacancy represents a single oxygen vacancy defect
type OxygenVacancy struct {
	ID          int
	X           float64 // nm (lateral position)
	Y           float64 // nm (lateral position)
	Z           float64 // nm (depth, 0=bottom electrode)
	Charge      int     // +2 for doubly ionized V_O^{2+}
	Mobile      bool    // Whether vacancy can migrate
	Formation   float64 // Formation energy (eV)
	Migration   float64 // Migration barrier (eV)
}

// OxygenVacancyDistribution represents the spatial distribution of vacancies
type OxygenVacancyDistribution struct {
	Vacancies     []*OxygenVacancy
	TotalConc     float64 // Total concentration (%)
	InterfaceConc float64 // Concentration at interfaces (%)
	BulkConc      float64 // Concentration in bulk (%)
	FilmThickness float64 // nm
}

// NewOxygenVacancyDistribution creates initial vacancy distribution
// Initially vacancies concentrate at electrode interfaces
func NewOxygenVacancyDistribution(thickness float64, totalConc float64, numVacancies int) *OxygenVacancyDistribution {
	ovd := &OxygenVacancyDistribution{
		Vacancies:     make([]*OxygenVacancy, numVacancies),
		TotalConc:     totalConc,
		FilmThickness: thickness,
	}

	// Create vacancies concentrated at interfaces (pristine state)
	for i := 0; i < numVacancies; i++ {
		// 80% at interfaces, 20% in bulk
		var z float64
		if rand.Float64() < 0.8 {
			// Interface region (within 2nm of electrodes)
			if rand.Float64() < 0.5 {
				z = rand.Float64() * 2.0 // Bottom electrode
			} else {
				z = thickness - rand.Float64()*2.0 // Top electrode
			}
		} else {
			z = 2.0 + rand.Float64()*(thickness-4.0) // Bulk region
		}

		ovd.Vacancies[i] = &OxygenVacancy{
			ID:        i,
			X:         rand.Float64() * 100.0, // 100nm lateral extent
			Y:         rand.Float64() * 100.0,
			Z:         z,
			Charge:    2, // V_O^{2+}
			Mobile:    true,
			Formation: 1.5 + rand.Float64()*0.5, // 1.5-2.0 eV
			Migration: 0.3 + rand.Float64()*0.2, // 0.3-0.5 eV
		}
	}

	ovd.updateConcentrations()
	return ovd
}

// updateConcentrations calculates interface and bulk concentrations
func (ovd *OxygenVacancyDistribution) updateConcentrations() {
	interfaceCount := 0
	bulkCount := 0
	interfaceThreshold := 2.0 // nm

	for _, v := range ovd.Vacancies {
		if v.Z < interfaceThreshold || v.Z > ovd.FilmThickness-interfaceThreshold {
			interfaceCount++
		} else {
			bulkCount++
		}
	}

	total := len(ovd.Vacancies)
	if total > 0 {
		ovd.InterfaceConc = ovd.TotalConc * float64(interfaceCount) / float64(total)
		ovd.BulkConc = ovd.TotalConc * float64(bulkCount) / float64(total)
	}
}

// MigrateUnderField simulates vacancy migration under electric field
func (ovd *OxygenVacancyDistribution) MigrateUnderField(electricField float64, temperature float64, dt float64) {
	kB := 8.617e-5 // eV/K
	kT := kB * temperature

	// Attempt frequency for ionic hopping
	attemptFreq := 1e13 // Hz (typical phonon frequency)
	hopDistance := 0.3  // nm (typical O-O distance)

	for _, v := range ovd.Vacancies {
		if !v.Mobile {
			continue
		}

		// Calculate drift velocity under field
		// Enhanced migration in field direction for positive charge
		effectiveBarrier := v.Migration
		if electricField > 0 {
			// Field lowers barrier in +z direction
			effectiveBarrier -= float64(v.Charge) * electricField * hopDistance * 0.01
		} else {
			// Field lowers barrier in -z direction
			effectiveBarrier += float64(v.Charge) * electricField * hopDistance * 0.01
		}

		if effectiveBarrier < 0 {
			effectiveBarrier = 0.01
		}

		// Hopping rate
		rate := attemptFreq * math.Exp(-effectiveBarrier/kT)

		// Probability of hopping in this timestep
		hopProb := rate * dt
		if hopProb > 1 {
			hopProb = 1
		}

		if rand.Float64() < hopProb {
			// Hop in field-preferred direction with some randomness
			dz := hopDistance
			if electricField < 0 {
				dz = -hopDistance
			}
			dz += (rand.Float64() - 0.5) * hopDistance * 0.5 // Random component

			v.Z += dz

			// Boundary conditions
			if v.Z < 0 {
				v.Z = 0
			}
			if v.Z > ovd.FilmThickness {
				v.Z = ovd.FilmThickness
			}
		}
	}

	ovd.updateConcentrations()
}

// GetDepthProfile returns vacancy concentration vs depth
func (ovd *OxygenVacancyDistribution) GetDepthProfile(numBins int) []float64 {
	profile := make([]float64, numBins)
	binWidth := ovd.FilmThickness / float64(numBins)

	for _, v := range ovd.Vacancies {
		bin := int(v.Z / binWidth)
		if bin >= numBins {
			bin = numBins - 1
		}
		profile[bin]++
	}

	// Normalize
	maxCount := 0.0
	for _, c := range profile {
		if c > maxCount {
			maxCount = c
		}
	}
	if maxCount > 0 {
		for i := range profile {
			profile[i] /= maxCount
		}
	}

	return profile
}

// ============================================================================
// WAKE-UP AND FATIGUE SIMULATION
// ============================================================================

// WakeUpState represents the current wake-up status
type WakeUpState int

const (
	Pristine   WakeUpState = iota // Initial state
	WakingUp                      // During wake-up process
	WokenUp                       // Fully woken up
	Fatiguing                     // During fatigue
	Fatigued                      // Severely fatigued
	BrokenDown                    // Dielectric breakdown
)

// WakeUpFatigueConfig configures wake-up and fatigue simulation
type WakeUpFatigueConfig struct {
	InitialPr           float64 // Initial remanent polarization (μC/cm²)
	MaxPr               float64 // Maximum achievable Pr (μC/cm²)
	WakeUpCycles        int     // Cycles to reach max Pr
	FatigueOnsetCycles  int     // Cycles when fatigue begins
	BreakdownCycles     int     // Cycles at breakdown (if applicable)
	VacancyConc         float64 // Initial oxygen vacancy concentration (%)
	Temperature         float64 // Operating temperature (K)
	AppliedField        float64 // Cycling field amplitude (MV/cm)
	FatigueType         string  // "degradation" or "breakdown"
}

// DefaultWakeUpFatigueConfig returns typical HZO configuration
func DefaultWakeUpFatigueConfig() *WakeUpFatigueConfig {
	return &WakeUpFatigueConfig{
		InitialPr:          10.0,
		MaxPr:              35.0,
		WakeUpCycles:       1000,
		FatigueOnsetCycles: 1e8,
		BreakdownCycles:    1e10,
		VacancyConc:        2.5,
		Temperature:        300.0,
		AppliedField:       3.0,
		FatigueType:        "degradation",
	}
}

// WakeUpFatigueSimulator simulates wake-up and fatigue effects
type WakeUpFatigueSimulator struct {
	Config        *WakeUpFatigueConfig
	State         WakeUpState
	CurrentCycles int64
	CurrentPr     float64
	CoerciveField float64 // Ec (MV/cm)
	LeakageCurrent float64 // A/cm²

	// Internal state
	vacancies     *OxygenVacancyDistribution
	domains       *DomainStructure
	pathwayDB     *SwitchingPathwayDatabase

	// History tracking
	PrHistory     []float64
	EcHistory     []float64
	CycleHistory  []int64
}

// NewWakeUpFatigueSimulator creates a new wake-up/fatigue simulator
func NewWakeUpFatigueSimulator(config *WakeUpFatigueConfig) *WakeUpFatigueSimulator {
	if config == nil {
		config = DefaultWakeUpFatigueConfig()
	}

	sim := &WakeUpFatigueSimulator{
		Config:        config,
		State:         Pristine,
		CurrentCycles: 0,
		CurrentPr:     config.InitialPr,
		CoerciveField: 1.5, // Initial Ec
		LeakageCurrent: 1e-9,

		vacancies:  NewOxygenVacancyDistribution(10.0, config.VacancyConc, 1000),
		domains:    NewDomainStructure(10.0, 10000.0, 15.0, 100),
		pathwayDB:  NewHfO2PathwayDatabase(),

		PrHistory:    make([]float64, 0),
		EcHistory:    make([]float64, 0),
		CycleHistory: make([]int64, 0),
	}

	return sim
}

// Cycle performs one or more switching cycles
func (sim *WakeUpFatigueSimulator) Cycle(numCycles int64) {
	for i := int64(0); i < numCycles; i++ {
		sim.singleCycle()
	}
}

// singleCycle performs one complete switching cycle
func (sim *WakeUpFatigueSimulator) singleCycle() {
	sim.CurrentCycles++

	// Simulate vacancy migration during each half-cycle
	dt := 1e-6 // 1 μs per half-cycle
	sim.vacancies.MigrateUnderField(sim.Config.AppliedField, sim.Config.Temperature, dt)
	sim.vacancies.MigrateUnderField(-sim.Config.AppliedField, sim.Config.Temperature, dt)

	// Update state based on cycle count
	sim.updateState()

	// Calculate current Pr based on state
	sim.calculatePr()

	// Update coercive field
	sim.calculateEc()

	// Track history periodically
	if sim.CurrentCycles%1000 == 0 || sim.CurrentCycles == 1 {
		sim.PrHistory = append(sim.PrHistory, sim.CurrentPr)
		sim.EcHistory = append(sim.EcHistory, sim.CoerciveField)
		sim.CycleHistory = append(sim.CycleHistory, sim.CurrentCycles)
	}
}

// updateState updates the wake-up/fatigue state
func (sim *WakeUpFatigueSimulator) updateState() {
	switch {
	case sim.CurrentCycles < int64(sim.Config.WakeUpCycles):
		sim.State = WakingUp
	case sim.CurrentCycles < int64(sim.Config.FatigueOnsetCycles):
		sim.State = WokenUp
	case sim.CurrentCycles < int64(sim.Config.BreakdownCycles):
		sim.State = Fatiguing
	default:
		if sim.Config.FatigueType == "breakdown" {
			sim.State = BrokenDown
		} else {
			sim.State = Fatigued
		}
	}
}

// calculatePr calculates current remanent polarization
func (sim *WakeUpFatigueSimulator) calculatePr() {
	cycle := float64(sim.CurrentCycles)

	switch sim.State {
	case Pristine:
		sim.CurrentPr = sim.Config.InitialPr

	case WakingUp:
		// Logarithmic increase during wake-up
		// Pr increases as vacancies redistribute from interfaces to bulk
		wakeUpProgress := cycle / float64(sim.Config.WakeUpCycles)
		sim.CurrentPr = sim.Config.InitialPr +
			(sim.Config.MaxPr-sim.Config.InitialPr)*math.Log(1+9*wakeUpProgress)/math.Log(10)

		// Correlate with vacancy distribution
		bulkFraction := sim.vacancies.BulkConc / sim.Config.VacancyConc
		sim.CurrentPr *= (0.7 + 0.3*bulkFraction)

	case WokenUp:
		sim.CurrentPr = sim.Config.MaxPr

	case Fatiguing:
		// Gradual degradation
		fatigueProgress := (cycle - float64(sim.Config.FatigueOnsetCycles)) /
			(float64(sim.Config.BreakdownCycles) - float64(sim.Config.FatigueOnsetCycles))

		// Stretched exponential decay (typical for ferroelectric fatigue)
		beta := 0.5 // Stretch exponent
		tau := 0.3  // Characteristic decay
		sim.CurrentPr = sim.Config.MaxPr * math.Exp(-math.Pow(fatigueProgress/tau, beta))

	case Fatigued:
		sim.CurrentPr = sim.Config.MaxPr * 0.3 // 70% degradation

	case BrokenDown:
		sim.CurrentPr = 0 // Complete loss
	}

	// Add noise
	sim.CurrentPr *= (1.0 + (rand.Float64()-0.5)*0.02)
}

// calculateEc calculates current coercive field
func (sim *WakeUpFatigueSimulator) calculateEc() {
	// Ec typically decreases during wake-up then increases during fatigue
	initialEc := 1.5 // MV/cm
	minEc := 1.0     // MV/cm (after wake-up)

	switch sim.State {
	case WakingUp:
		progress := float64(sim.CurrentCycles) / float64(sim.Config.WakeUpCycles)
		sim.CoerciveField = initialEc - (initialEc-minEc)*progress

	case WokenUp:
		sim.CoerciveField = minEc

	case Fatiguing:
		// Ec increases due to defect generation
		progress := (float64(sim.CurrentCycles) - float64(sim.Config.FatigueOnsetCycles)) /
			(float64(sim.Config.BreakdownCycles) - float64(sim.Config.FatigueOnsetCycles))
		sim.CoerciveField = minEc + (initialEc-minEc)*progress*2

	case Fatigued:
		sim.CoerciveField = initialEc * 1.5

	case BrokenDown:
		sim.CoerciveField = 0 // Undefined
	}
}

// GetStatistics returns current simulation statistics
func (sim *WakeUpFatigueSimulator) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"cycles":          sim.CurrentCycles,
		"state":           sim.State.String(),
		"Pr_uC_cm2":       sim.CurrentPr,
		"Ec_MV_cm":        sim.CoerciveField,
		"interface_conc":  sim.vacancies.InterfaceConc,
		"bulk_conc":       sim.vacancies.BulkConc,
		"leakage_A_cm2":   sim.LeakageCurrent,
	}
}

// String returns string representation of WakeUpState
func (s WakeUpState) String() string {
	states := []string{"Pristine", "WakingUp", "WokenUp", "Fatiguing", "Fatigued", "BrokenDown"}
	if int(s) < len(states) {
		return states[s]
	}
	return "Unknown"
}

// ============================================================================
// ELECTRICAL PULSE ENGINEERING
// ============================================================================

// PulseType defines the type of electrical pulse
type PulseType int

const (
	SquarePulse PulseType = iota
	TriangularPulse
	SinusoidalPulse
	TrainPulse
	BipolarPulse
)

// ElectricalPulse represents an engineered electrical pulse
type ElectricalPulse struct {
	Type        PulseType
	Amplitude   float64 // V
	Duration    float64 // s
	RiseTime    float64 // s
	FallTime    float64 // s
	Frequency   float64 // Hz (for train pulses)
	DutyCycle   float64 // 0-1
	NumPulses   int     // For train
}

// PulseEngineeringConfig configures pulse-based wake-up/fatigue control
type PulseEngineeringConfig struct {
	WakeUpPulse        ElectricalPulse // Pulse to accelerate wake-up
	FatigueResetPulse  ElectricalPulse // Pulse to retard fatigue
	LeakageResetPulse  ElectricalPulse // Pulse to reduce leakage
}

// DefaultPulseEngineeringConfig returns optimized pulse configuration
// Based on Jan et al., Small Science 2024
func DefaultPulseEngineeringConfig() *PulseEngineeringConfig {
	return &PulseEngineeringConfig{
		WakeUpPulse: ElectricalPulse{
			Type:      SquarePulse,
			Amplitude: 3.5,     // V
			Duration:  100e-9,  // 100 ns (short pulses speed up wake-up)
			RiseTime:  5e-9,
			FallTime:  5e-9,
		},
		FatigueResetPulse: ElectricalPulse{
			Type:      BipolarPulse,
			Amplitude: 2.5,     // V (lower amplitude retards fatigue)
			Duration:  1e-6,    // 1 μs
			RiseTime:  10e-9,
			FallTime:  10e-9,
		},
		LeakageResetPulse: ElectricalPulse{
			Type:      TrainPulse,
			Amplitude: 1.5,     // V
			Duration:  50e-9,
			Frequency: 1e6,     // 1 MHz
			NumPulses: 10,
		},
	}
}

// PulseEngineer implements pulse-based reliability engineering
type PulseEngineer struct {
	Config    *PulseEngineeringConfig
	Simulator *WakeUpFatigueSimulator

	// Performance improvements
	WakeUpSpeedup     float64 // Factor (e.g., 1.5x)
	FatigueRetardation float64 // Factor (e.g., 4x)
	LeakageReduction  float64 // Percentage
}

// NewPulseEngineer creates a pulse engineering controller
func NewPulseEngineer(config *PulseEngineeringConfig, sim *WakeUpFatigueSimulator) *PulseEngineer {
	if config == nil {
		config = DefaultPulseEngineeringConfig()
	}

	return &PulseEngineer{
		Config:             config,
		Simulator:         sim,
		WakeUpSpeedup:     1.5,
		FatigueRetardation: 4.0,
		LeakageReduction:  59.0, // 59% reduction
	}
}

// ApplyWakeUpPulse applies optimized wake-up pulse sequence
func (pe *PulseEngineer) ApplyWakeUpPulse(numPulses int) {
	pulse := pe.Config.WakeUpPulse

	for i := 0; i < numPulses; i++ {
		// Short pulses accelerate vacancy redistribution
		effectiveCycles := int64(pulse.Duration * 1e9) // Convert to equivalent cycles

		// Apply enhanced migration due to pulse characteristics
		pe.Simulator.vacancies.MigrateUnderField(
			pulse.Amplitude*0.3, // Convert V to MV/cm approximately
			pe.Simulator.Config.Temperature,
			pulse.Duration,
		)

		pe.Simulator.CurrentCycles += effectiveCycles
	}

	// Accelerated wake-up effect
	pe.Simulator.Config.WakeUpCycles = int(float64(pe.Simulator.Config.WakeUpCycles) / pe.WakeUpSpeedup)
}

// ApplyFatigueResetPulse applies pulse sequence to retard fatigue
func (pe *PulseEngineer) ApplyFatigueResetPulse() {
	pulse := pe.Config.FatigueResetPulse

	// Bipolar pulses help redistribute vacancies more uniformly
	// This prevents accumulation at interfaces
	for i := 0; i < 100; i++ {
		// Positive half
		pe.Simulator.vacancies.MigrateUnderField(
			pulse.Amplitude*0.2,
			pe.Simulator.Config.Temperature,
			pulse.Duration/2,
		)
		// Negative half
		pe.Simulator.vacancies.MigrateUnderField(
			-pulse.Amplitude*0.2,
			pe.Simulator.Config.Temperature,
			pulse.Duration/2,
		)
	}

	// Extend fatigue onset
	pe.Simulator.Config.FatigueOnsetCycles = int(float64(pe.Simulator.Config.FatigueOnsetCycles) * pe.FatigueRetardation)
}

// ============================================================================
// FATIGUE-FREE INTERFACE DESIGN
// ============================================================================

// InterfaceType defines the electrode/ferroelectric interface type
type InterfaceType int

const (
	StandardInterface InterfaceType = iota // TiN/HZO/TiN
	OxygenSponge                           // CeO2-x/HZO
	Superlattice                           // HfO2/ZrO2 superlattice
	GradedInterface                        // Composition-graded
	ScavengingLayer                        // With oxygen scavenging layer
)

// FatigueFreeInterface models interface engineering for fatigue-free operation
type FatigueFreeInterface struct {
	Type             InterfaceType
	Material         string
	Thickness        float64  // nm
	OxygenActivity   float64  // 0-1, ability to absorb/release O
	CoerciveFieldRed float64  // Ec reduction factor
	EnduranceBoost   float64  // Endurance improvement factor
	Description      string
}

// NewCeO2Interface creates a CeO2-x oxygen sponge interface
// Based on Nature Communications 2025 fatigue-free design
func NewCeO2Interface() *FatigueFreeInterface {
	return &FatigueFreeInterface{
		Type:             OxygenSponge,
		Material:         "CeO2-x",
		Thickness:        3.0,
		OxygenActivity:   0.9,  // High oxygen activity
		CoerciveFieldRed: 0.6,  // 40% Ec reduction
		EnduranceBoost:   100.0, // 100x endurance improvement
		Description:      "Coherent CeO2-x/HZO heterointerface - oxygen sponge design",
	}
}

// NewSuperlatticeInterface creates HfO2/ZrO2 superlattice interface
func NewSuperlatticeInterface(numPeriods int, periodThickness float64) *FatigueFreeInterface {
	return &FatigueFreeInterface{
		Type:             Superlattice,
		Material:         fmt.Sprintf("(HfO2)%d/(ZrO2)%d superlattice", numPeriods, numPeriods),
		Thickness:        float64(numPeriods) * periodThickness * 2,
		OxygenActivity:   0.5,
		CoerciveFieldRed: 0.85, // Ec = 0.85 MV/cm
		EnduranceBoost:   10.0, // 10^9 cycles demonstrated
		Description:      "Epitaxial HfO2/ZrO2 superlattice with optimized period",
	}
}

// FatigueFreeDesign implements comprehensive fatigue-free device design
type FatigueFreeDesign struct {
	TopInterface    *FatigueFreeInterface
	BottomInterface *FatigueFreeInterface
	FerroThickness  float64 // nm
	VacancyConc     float64 // % (target low, e.g., 1.9%)
	ExpectedEndurance int64  // Cycles
	ExpectedPr      float64 // μC/cm²
}

// NewFatigueFreeDesign creates an optimized fatigue-free device design
func NewFatigueFreeDesign() *FatigueFreeDesign {
	return &FatigueFreeDesign{
		TopInterface:     NewCeO2Interface(),
		BottomInterface:  NewCeO2Interface(),
		FerroThickness:   8.0,
		VacancyConc:      1.9, // Low vacancy concentration
		ExpectedEndurance: 1e10,
		ExpectedPr:       36.4, // μC/cm² demonstrated
	}
}

// SimulateFatigueFree runs fatigue-free device simulation
func (ffd *FatigueFreeDesign) SimulateFatigueFree() *WakeUpFatigueSimulator {
	// Configure with fatigue-free parameters
	config := &WakeUpFatigueConfig{
		InitialPr:          ffd.ExpectedPr * 0.9, // Minimal wake-up needed
		MaxPr:              ffd.ExpectedPr,
		WakeUpCycles:       100,                  // Very fast wake-up
		FatigueOnsetCycles: int(ffd.ExpectedEndurance),
		BreakdownCycles:    int(ffd.ExpectedEndurance * 10),
		VacancyConc:        ffd.VacancyConc,
		Temperature:        300.0,
		AppliedField:       2.0, // Lower field due to reduced Ec
		FatigueType:        "degradation",
	}

	sim := NewWakeUpFatigueSimulator(config)
	return sim
}

// ============================================================================
// DOMAIN-AWARE CIM CROSSBAR
// ============================================================================

// DomainAwareCIMCell represents a CIM cell with domain-level modeling
type DomainAwareCIMCell struct {
	Row          int
	Col          int
	Domains      *DomainStructure
	Vacancies    *OxygenVacancyDistribution
	Cycles       int64
	State        WakeUpState
	Conductance  float64 // Current conductance (S)
	MaxCond      float64 // Maximum conductance
	MinCond      float64 // Minimum conductance
}

// DomainAwareCIMArray implements domain-aware CIM crossbar
type DomainAwareCIMArray struct {
	Rows         int
	Cols         int
	Cells        [][]*DomainAwareCIMCell
	PathwayDB    *SwitchingPathwayDatabase
	TotalCycles  int64
	ArrayState   WakeUpState
}

// NewDomainAwareCIMArray creates a domain-aware CIM crossbar array
func NewDomainAwareCIMArray(rows, cols int) *DomainAwareCIMArray {
	array := &DomainAwareCIMArray{
		Rows:        rows,
		Cols:        cols,
		Cells:       make([][]*DomainAwareCIMCell, rows),
		PathwayDB:   NewHfO2PathwayDatabase(),
		TotalCycles: 0,
		ArrayState:  Pristine,
	}

	for i := 0; i < rows; i++ {
		array.Cells[i] = make([]*DomainAwareCIMCell, cols)
		for j := 0; j < cols; j++ {
			array.Cells[i][j] = &DomainAwareCIMCell{
				Row:         i,
				Col:         j,
				Domains:     NewDomainStructure(10.0, 1000.0, 15.0, 20),
				Vacancies:   NewOxygenVacancyDistribution(10.0, 2.5, 100),
				Cycles:      0,
				State:       Pristine,
				MaxCond:     1e-4, // 0.1 mS
				MinCond:     1e-7, // 0.1 μS
			}
			array.Cells[i][j].updateConductance()
		}
	}

	return array
}

// updateConductance updates cell conductance based on domain structure
func (cell *DomainAwareCIMCell) updateConductance() {
	netP := cell.Domains.GetNetPolarization()
	maxP := 40.0 // μC/cm²

	// Map polarization to conductance (FTJ-like behavior)
	normalizedP := (netP + maxP) / (2 * maxP) // 0 to 1
	if normalizedP < 0 {
		normalizedP = 0
	}
	if normalizedP > 1 {
		normalizedP = 1
	}

	// Logarithmic conductance mapping
	cell.Conductance = cell.MinCond * math.Pow(cell.MaxCond/cell.MinCond, normalizedP)
}

// ProgramCell programs a cell to target conductance
func (array *DomainAwareCIMArray) ProgramCell(row, col int, targetCond float64) error {
	if row < 0 || row >= array.Rows || col < 0 || col >= array.Cols {
		return fmt.Errorf("cell index out of range")
	}

	cell := array.Cells[row][col]

	// Determine required polarization state
	targetNormalized := math.Log(targetCond/cell.MinCond) / math.Log(cell.MaxCond/cell.MinCond)
	maxP := 40.0
	targetP := targetNormalized*2*maxP - maxP

	// Apply field to switch domains
	currentP := cell.Domains.GetNetPolarization()

	// Iterative programming with verification
	maxIterations := 100
	tolerance := 2.0 // μC/cm²

	for iter := 0; iter < maxIterations; iter++ {
		diff := targetP - currentP
		if math.Abs(diff) < tolerance {
			break
		}

		// Apply field proportional to required change
		field := diff * 0.1 // MV/cm per μC/cm² difference
		cell.Domains.SwitchDomains(field, array.PathwayDB)

		// Simulate vacancy migration during programming
		cell.Vacancies.MigrateUnderField(field, 300.0, 1e-6)

		cell.Cycles++
		currentP = cell.Domains.GetNetPolarization()
	}

	cell.updateConductance()
	return nil
}

// MatrixVectorMultiply performs analog MVM with domain-aware accuracy
func (array *DomainAwareCIMArray) MatrixVectorMultiply(input []float64) ([]float64, error) {
	if len(input) != array.Cols {
		return nil, fmt.Errorf("input size mismatch: got %d, expected %d", len(input), array.Cols)
	}

	output := make([]float64, array.Rows)

	for i := 0; i < array.Rows; i++ {
		var sum float64
		for j := 0; j < array.Cols; j++ {
			cell := array.Cells[i][j]

			// Conductance-based computation with wake-up/fatigue effects
			effectiveCond := cell.Conductance

			// Apply state-dependent degradation
			switch cell.State {
			case Fatiguing:
				effectiveCond *= 0.9
			case Fatigued:
				effectiveCond *= 0.7
			case BrokenDown:
				effectiveCond = cell.MinCond // Stuck at minimum
			}

			sum += input[j] * effectiveCond
		}
		output[i] = sum
	}

	return output, nil
}

// CycleArray performs cycling on entire array (for wear simulation)
func (array *DomainAwareCIMArray) CycleArray(numCycles int64) {
	for i := 0; i < array.Rows; i++ {
		for j := 0; j < array.Cols; j++ {
			cell := array.Cells[i][j]
			cell.Cycles += numCycles

			// Update state based on cycles
			switch {
			case cell.Cycles < 1000:
				cell.State = WakingUp
			case cell.Cycles < 1e8:
				cell.State = WokenUp
			case cell.Cycles < 1e10:
				cell.State = Fatiguing
			default:
				cell.State = Fatigued
			}

			// Simulate vacancy evolution
			for c := int64(0); c < numCycles; c += 1000 {
				cell.Vacancies.MigrateUnderField(3.0, 300.0, 1e-6)
				cell.Vacancies.MigrateUnderField(-3.0, 300.0, 1e-6)
			}

			cell.updateConductance()
		}
	}

	array.TotalCycles += numCycles
}

// GetArrayStatistics returns array-level statistics
func (array *DomainAwareCIMArray) GetArrayStatistics() map[string]interface{} {
	stateCount := make(map[WakeUpState]int)
	totalCond := 0.0
	minCond := math.MaxFloat64
	maxCond := 0.0

	for i := 0; i < array.Rows; i++ {
		for j := 0; j < array.Cols; j++ {
			cell := array.Cells[i][j]
			stateCount[cell.State]++
			totalCond += cell.Conductance
			if cell.Conductance < minCond {
				minCond = cell.Conductance
			}
			if cell.Conductance > maxCond {
				maxCond = cell.Conductance
			}
		}
	}

	numCells := float64(array.Rows * array.Cols)

	return map[string]interface{}{
		"rows":           array.Rows,
		"cols":           array.Cols,
		"total_cycles":   array.TotalCycles,
		"avg_conductance": totalCond / numCells,
		"min_conductance": minCond,
		"max_conductance": maxCond,
		"state_distribution": stateCount,
	}
}

// ============================================================================
// FECIM DOMAIN-AWARE INTEGRATION
// ============================================================================

// FeCIMDomainConfig configures FeCIM domain-aware simulation
type FeCIMDomainConfig struct {
	// Material parameters (HZO superlattice)
	FerroThickness float64 // nm
	GrainSize      float64 // nm
	VacancyConc    float64 // %

	// Interface design
	InterfaceType    InterfaceType
	OxygenActivity   float64

	// Reliability targets
	TargetEndurance  int64   // Cycles
	TargetPr         float64 // μC/cm²
	MaxVariation     float64 // %

	// Pulse engineering
	EnablePulseEng   bool
}

// DefaultFeCIMDomainConfig returns FeCIM-optimized configuration
func DefaultFeCIMDomainConfig() *FeCIMDomainConfig {
	return &FeCIMDomainConfig{
		FerroThickness:  8.0,
		GrainSize:       15.0, // 10-20 nm typical
		VacancyConc:     2.0,
		InterfaceType:   OxygenSponge,
		OxygenActivity:  0.8,
		TargetEndurance: 1e10,
		TargetPr:        35.0,
		MaxVariation:    3.0,
		EnablePulseEng:  true,
	}
}

// FeCIMDomainArray implements FeCIM domain-aware CIM array
type FeCIMDomainArray struct {
	Config       *FeCIMDomainConfig
	Array        *DomainAwareCIMArray
	PulseEng     *PulseEngineer
	FatigueDesign *FatigueFreeDesign
	PathwayDB    *SwitchingPathwayDatabase

	// Performance metrics
	Accuracy     float64
	EnergyPerOp  float64 // fJ
	Throughput   float64 // TOPS
}

// NewFeCIMDomainArray creates an FeCIM domain-aware array
func NewFeCIMDomainArray(rows, cols int, config *FeCIMDomainConfig) *FeCIMDomainArray {
	if config == nil {
		config = DefaultFeCIMDomainConfig()
	}

	array := NewDomainAwareCIMArray(rows, cols)
	fatigueDesign := NewFatigueFreeDesign()
	fatigueDesign.VacancyConc = config.VacancyConc
	fatigueDesign.FerroThickness = config.FerroThickness

	sim := fatigueDesign.SimulateFatigueFree()

	var pulseEng *PulseEngineer
	if config.EnablePulseEng {
		pulseEng = NewPulseEngineer(DefaultPulseEngineeringConfig(), sim)
	}

	return &FeCIMDomainArray{
		Config:        config,
		Array:         array,
		PulseEng:      pulseEng,
		FatigueDesign: fatigueDesign,
		PathwayDB:     NewHfO2PathwayDatabase(),
		EnergyPerOp:   50.0, // fJ
		Throughput:    100.0, // TOPS
	}
}

// ProgramWeights programs weight matrix with domain-aware precision
func (ila *FeCIMDomainArray) ProgramWeights(weights [][]float64) error {
	if len(weights) != ila.Array.Rows {
		return fmt.Errorf("weight matrix row count mismatch")
	}

	for i, row := range weights {
		if len(row) != ila.Array.Cols {
			return fmt.Errorf("weight matrix column count mismatch at row %d", i)
		}

		for j, w := range row {
			// Map weight to conductance
			// Assume weights normalized to [-1, 1]
			normalizedW := (w + 1.0) / 2.0
			cell := ila.Array.Cells[i][j]
			targetCond := cell.MinCond * math.Pow(cell.MaxCond/cell.MinCond, normalizedW)

			if err := ila.Array.ProgramCell(i, j, targetCond); err != nil {
				return err
			}
		}
	}

	return nil
}

// Inference performs inference with domain-aware degradation modeling
func (ila *FeCIMDomainArray) Inference(input []float64) ([]float64, error) {
	// Apply pulse engineering if needed
	if ila.PulseEng != nil && ila.Array.TotalCycles > int64(ila.PulseEng.Simulator.Config.WakeUpCycles) {
		// Periodic fatigue reset
		if ila.Array.TotalCycles%1000000 == 0 {
			ila.PulseEng.ApplyFatigueResetPulse()
		}
	}

	output, err := ila.Array.MatrixVectorMultiply(input)
	if err != nil {
		return nil, err
	}

	// Apply wake-up/fatigue correction
	ila.applyReliabilityCorrection(output)

	return output, nil
}

// applyReliabilityCorrection applies corrections based on array state
func (ila *FeCIMDomainArray) applyReliabilityCorrection(output []float64) {
	stats := ila.Array.GetArrayStatistics()
	stateDistribution := stats["state_distribution"].(map[WakeUpState]int)

	totalCells := ila.Array.Rows * ila.Array.Cols
	fatiguedFraction := float64(stateDistribution[Fatiguing]+stateDistribution[Fatigued]) / float64(totalCells)

	// Apply correction factor based on fatigue level
	if fatiguedFraction > 0.1 {
		correctionFactor := 1.0 / (1.0 - fatiguedFraction*0.3)
		for i := range output {
			output[i] *= correctionFactor
		}
	}
}

// GetReliabilityReport returns comprehensive reliability report
func (ila *FeCIMDomainArray) GetReliabilityReport() map[string]interface{} {
	arrayStats := ila.Array.GetArrayStatistics()

	// Calculate expected remaining lifetime
	currentCycles := ila.Array.TotalCycles
	remainingCycles := ila.Config.TargetEndurance - currentCycles
	if remainingCycles < 0 {
		remainingCycles = 0
	}

	// Estimate accuracy degradation
	stateDistribution := arrayStats["state_distribution"].(map[WakeUpState]int)
	totalCells := ila.Array.Rows * ila.Array.Cols
	healthyFraction := float64(stateDistribution[WokenUp]) / float64(totalCells)

	estimatedAccuracy := 98.0 * healthyFraction // Baseline 98% accuracy

	return map[string]interface{}{
		"array_stats":        arrayStats,
		"total_cycles":       currentCycles,
		"remaining_cycles":   remainingCycles,
		"lifetime_percent":   float64(currentCycles) / float64(ila.Config.TargetEndurance) * 100,
		"estimated_accuracy": estimatedAccuracy,
		"healthy_cell_percent": healthyFraction * 100,
		"pulse_engineering":  ila.PulseEng != nil,
		"interface_type":     ila.Config.InterfaceType,
		"energy_per_op_fJ":   ila.EnergyPerOp,
		"throughput_TOPS":    ila.Throughput,
	}
}

// ExportConfiguration exports configuration for documentation
func (ila *FeCIMDomainArray) ExportConfiguration() ([]byte, error) {
	export := map[string]interface{}{
		"config":           ila.Config,
		"array_size":       []int{ila.Array.Rows, ila.Array.Cols},
		"fatigue_design":   ila.FatigueDesign,
		"reliability":      ila.GetReliabilityReport(),
		"num_pathways":     len(ila.PathwayDB.Pathways),
	}

	return json.MarshalIndent(export, "", "  ")
}

// ============================================================================
// DOMAIN VISUALIZATION UTILITIES
// ============================================================================

// DomainVisualization provides domain structure visualization
type DomainVisualization struct {
	Domains     *DomainStructure
	Resolution  int // Grid resolution
}

// NewDomainVisualization creates visualization for domain structure
func NewDomainVisualization(domains *DomainStructure, resolution int) *DomainVisualization {
	return &DomainVisualization{
		Domains:    domains,
		Resolution: resolution,
	}
}

// GenerateASCIIMap generates ASCII art representation of domains
func (dv *DomainVisualization) GenerateASCIIMap() string {
	sqrtArea := math.Sqrt(dv.Domains.FilmArea)
	cellSize := sqrtArea / float64(dv.Resolution)

	grid := make([][]rune, dv.Resolution)
	for i := range grid {
		grid[i] = make([]rune, dv.Resolution)
		for j := range grid[i] {
			grid[i][j] = '.'
		}
	}

	// Map domains to grid
	for _, d := range dv.Domains.Domains {
		x := int(d.CenterX / cellSize)
		y := int(d.CenterY / cellSize)

		if x >= 0 && x < dv.Resolution && y >= 0 && y < dv.Resolution {
			switch d.Orientation {
			case DomainUp:
				grid[y][x] = '↑'
			case DomainDown:
				grid[y][x] = '↓'
			case DomainInPlane0:
				grid[y][x] = '→'
			case DomainInPlane180:
				grid[y][x] = '←'
			default:
				grid[y][x] = '●'
			}
		}
	}

	// Convert to string
	result := ""
	for _, row := range grid {
		result += string(row) + "\n"
	}

	return result
}

// GenerateVacancyProfile generates ASCII vacancy depth profile
func (dv *DomainVisualization) GenerateVacancyProfile(vacancies *OxygenVacancyDistribution) string {
	profile := vacancies.GetDepthProfile(20)

	result := "Vacancy Depth Profile:\n"
	result += "Interface (bottom) ←→ Interface (top)\n"

	for i, val := range profile {
		bars := int(val * 50)
		result += fmt.Sprintf("%2d | %s\n", i, repeatChar('█', bars))
	}

	return result
}

// repeatChar repeats a character n times
func repeatChar(c rune, n int) string {
	result := make([]rune, n)
	for i := range result {
		result[i] = c
	}
	return string(result)
}

// ============================================================================
// BENCHMARK AND ANALYSIS
// ============================================================================

// DomainFatigueBenchmark runs comprehensive domain/fatigue benchmarks
type DomainFatigueBenchmark struct {
	Config    *FeCIMDomainConfig
	Results   []BenchmarkResult
}

// BenchmarkResult stores single benchmark result
type BenchmarkResult struct {
	Name           string
	Cycles         int64
	Pr             float64
	Ec             float64
	Accuracy       float64
	HealthyPercent float64
}

// NewDomainFatigueBenchmark creates benchmark suite
func NewDomainFatigueBenchmark(config *FeCIMDomainConfig) *DomainFatigueBenchmark {
	return &DomainFatigueBenchmark{
		Config:  config,
		Results: make([]BenchmarkResult, 0),
	}
}

// RunEnduranceBenchmark tests endurance over cycling
func (dfb *DomainFatigueBenchmark) RunEnduranceBenchmark(maxCycles int64, checkpoints []int64) {
	array := NewFeCIMDomainArray(32, 32, dfb.Config)

	// Sort checkpoints
	sort.Slice(checkpoints, func(i, j int) bool {
		return checkpoints[i] < checkpoints[j]
	})

	for _, checkpoint := range checkpoints {
		cyclesToRun := checkpoint - array.Array.TotalCycles
		if cyclesToRun > 0 {
			array.Array.CycleArray(cyclesToRun)
		}

		report := array.GetReliabilityReport()

		dfb.Results = append(dfb.Results, BenchmarkResult{
			Name:           fmt.Sprintf("Endurance_%d", checkpoint),
			Cycles:         checkpoint,
			Accuracy:       report["estimated_accuracy"].(float64),
			HealthyPercent: report["healthy_cell_percent"].(float64),
		})
	}
}

// GenerateReport generates markdown benchmark report
func (dfb *DomainFatigueBenchmark) GenerateReport() string {
	report := "# Domain Engineering & Fatigue Benchmark Report\n\n"
	report += "## Configuration\n"
	report += fmt.Sprintf("- Ferroelectric thickness: %.1f nm\n", dfb.Config.FerroThickness)
	report += fmt.Sprintf("- Grain size: %.1f nm\n", dfb.Config.GrainSize)
	report += fmt.Sprintf("- Vacancy concentration: %.1f%%\n", dfb.Config.VacancyConc)
	report += fmt.Sprintf("- Target endurance: %d cycles\n", dfb.Config.TargetEndurance)
	report += fmt.Sprintf("- Pulse engineering: %v\n\n", dfb.Config.EnablePulseEng)

	report += "## Results\n\n"
	report += "| Benchmark | Cycles | Accuracy (%) | Healthy Cells (%) |\n"
	report += "|-----------|--------|--------------|-------------------|\n"

	for _, r := range dfb.Results {
		report += fmt.Sprintf("| %s | %d | %.2f | %.2f |\n",
			r.Name, r.Cycles, r.Accuracy, r.HealthyPercent)
	}

	return report
}
