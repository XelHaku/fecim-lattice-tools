// Package layers provides simulation of emerging ferroelectric materials
// and comprehensive endurance testing methodologies for CIM devices.
package layers

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
)

// =============================================================================
// EMERGING FERROELECTRIC MATERIALS
// =============================================================================

// MaterialClass represents the class of ferroelectric material
type MaterialClass int

const (
	MaterialClassFluorite  MaterialClass = iota // HfO2, HZO
	MaterialClassWurtzite                       // AlScN, AlN
	MaterialClass2DVdW                          // CuInP2S6, In2Se3
	MaterialClassPerovskite                     // PZT, BaTiO3
)

// FerroelectricMaterial represents a ferroelectric material's properties
type FerroelectricMaterial struct {
	Name               string
	Class              MaterialClass
	RemanentPolarization float64 // Pr in µC/cm²
	CoerciveField      float64   // Ec in MV/cm
	CurieTemperature   float64   // Tc in K
	DielectricConstant float64   // εr
	BandGap            float64   // in eV
	Endurance          int64     // cycles
	RetentionTime      float64   // in seconds
	CMOSCompatible     bool
	Scalable           bool      // sub-10nm scalability
}

// =============================================================================
// WURTZITE FERROELECTRICS (AlScN)
// =============================================================================

// AlScNConfig configures AlScN material simulation
type AlScNConfig struct {
	ScandiumContent   float64 // x in Al(1-x)Sc(x)N, typically 0.27-0.43
	Thickness         float64 // in nm
	GrowthTemperature float64 // in °C
	SubstrateType     string  // "Si", "SiC", "GaN", "Sapphire"
}

// AlScNMaterial models AlScN wurtzite ferroelectric
type AlScNMaterial struct {
	Config             *AlScNConfig
	Properties         *FerroelectricMaterial
	CAxisOrientation   float64 // (002) texture quality
	GrainSize          float64 // in nm
	InterfaceQuality   float64 // 0-1 scale
}

// NewAlScNMaterial creates a new AlScN material model
func NewAlScNMaterial(config *AlScNConfig) *AlScNMaterial {
	alscn := &AlScNMaterial{
		Config: config,
	}

	// Calculate properties based on Sc content
	// Pr increases with Sc content (up to saturation ~40%)
	x := config.ScandiumContent
	pr := 80.0 + 120.0*x // Pr: 80-130 µC/cm² typical
	if x > 0.4 {
		pr = 130.0 // Saturation
	}

	// Ec decreases with Sc content
	ec := 5.0 - 3.0*x // Ec: 2-5 MV/cm

	// Tc very high for wurtzite (>1000K)
	tc := 1200.0 - 200.0*x

	alscn.Properties = &FerroelectricMaterial{
		Name:               fmt.Sprintf("Al%.2fSc%.2fN", 1-x, x),
		Class:              MaterialClassWurtzite,
		RemanentPolarization: pr,
		CoerciveField:      ec,
		CurieTemperature:   tc,
		DielectricConstant: 10.0 + 5.0*x,
		BandGap:            6.2 - 2.5*x, // AlN: 6.2eV, decreases with Sc
		Endurance:          1e10,        // High endurance reported
		RetentionTime:      1e10,        // Excellent retention
		CMOSCompatible:     true,
		Scalable:           true,
	}

	// Calculate structural properties
	alscn.CAxisOrientation = 0.95 - 0.1*rand.Float64()
	alscn.GrainSize = 20.0 + 30.0*(1-x) // Larger grains at lower Sc
	alscn.InterfaceQuality = 0.9 - 0.2*x

	return alscn
}

// CalculateSwitchingVoltage computes switching voltage for given thickness
func (alscn *AlScNMaterial) CalculateSwitchingVoltage() float64 {
	// V = Ec × thickness
	return alscn.Properties.CoerciveField * alscn.Config.Thickness * 1e-7 * 1e6 // MV/cm × nm → V
}

// =============================================================================
// 2D VAN DER WAALS FERROELECTRICS
// =============================================================================

// CIPS2DConfig configures CuInP2S6 simulation
type CIPS2DConfig struct {
	Thickness       float64 // in nm (can be as thin as 4nm)
	StackingLayers  int
	SubstrateType   string // "SiO2", "hBN", "Graphene"
	Temperature     float64 // in K
}

// CIPSMaterial models CuInP2S6 2D ferroelectric
type CIPSMaterial struct {
	Config           *CIPS2DConfig
	Properties       *FerroelectricMaterial
	CuIonMobility    float64 // Cu+ ion contribution
	Electroresistance float64 // ON/OFF ratio
	RectificationRatio float64
}

// NewCIPSMaterial creates a new CIPS material model
func NewCIPSMaterial(config *CIPS2DConfig) *CIPSMaterial {
	cips := &CIPSMaterial{
		Config: config,
	}

	// CIPS properties
	cips.Properties = &FerroelectricMaterial{
		Name:               "CuInP2S6",
		Class:              MaterialClass2DVdW,
		RemanentPolarization: 4.0, // ~4 µC/cm² for thin films
		CoerciveField:      0.2,   // ~200 kV/cm = 0.2 MV/cm
		CurieTemperature:   320.0, // ~320 K (room temperature)
		DielectricConstant: 15.0,
		BandGap:            2.9,   // ~2.9 eV
		Endurance:          1e6,   // 10^6 cycles demonstrated
		RetentionTime:      5.2e6, // 2 months demonstrated
		CMOSCompatible:     true,
		Scalable:           true,  // Can scale to 4nm
	}

	// Thickness-dependent properties
	if config.Thickness < 10 {
		cips.Properties.RemanentPolarization *= 0.8 // Reduced for ultrathin
	}

	// Calculate device metrics
	cips.CuIonMobility = 1e-10 * math.Exp(-0.5/(8.617e-5*config.Temperature))
	cips.Electroresistance = 1e6       // 10^6 reported
	cips.RectificationRatio = 2500.0   // Reported for CIPS/graphene

	return cips
}

// CalculateOnOffRatio computes memory ON/OFF ratio
func (cips *CIPSMaterial) CalculateOnOffRatio() float64 {
	// Combining ferroelectric and ionic contributions
	// Au/CIPS/Au devices: >10^8 reported
	ferroContribution := 1e4
	ionicContribution := 1e4 * cips.CuIonMobility / 1e-10

	return ferroContribution * ionicContribution / 1e4
}

// =============================================================================
// IN2SE3 FERROELECTRIC
// =============================================================================

// In2Se3Config configures In2Se3 simulation
type In2Se3Config struct {
	Phase           string  // "alpha", "beta", "gamma"
	Thickness       float64 // in nm
	SubstrateType   string
	Temperature     float64
}

// In2Se3Material models In2Se3 2D ferroelectric
type In2Se3Material struct {
	Config           *In2Se3Config
	Properties       *FerroelectricMaterial
	InPlanePolarization float64
	OutOfPlanePolarization float64
	MemristiveRatio  float64
}

// NewIn2Se3Material creates In2Se3 material model
func NewIn2Se3Material(config *In2Se3Config) *In2Se3Material {
	in2se3 := &In2Se3Material{
		Config: config,
	}

	in2se3.Properties = &FerroelectricMaterial{
		Name:               "α-In2Se3",
		Class:              MaterialClass2DVdW,
		RemanentPolarization: 0.5, // Lower than CIPS
		CoerciveField:      0.15,  // ~150 kV/cm
		CurieTemperature:   475.0, // Higher than CIPS
		DielectricConstant: 20.0,
		BandGap:            1.3,   // ~1.3 eV (semiconductor)
		Endurance:          1e7,
		RetentionTime:      1e8,
		CMOSCompatible:     true,
		Scalable:           true,
	}

	// α-In2Se3 has both IP and OOP polarization
	in2se3.InPlanePolarization = 0.3
	in2se3.OutOfPlanePolarization = 0.5
	in2se3.MemristiveRatio = 100.0 // Gate-tunable

	return in2se3
}

// =============================================================================
// MATERIAL COMPARISON DATABASE
// =============================================================================

// MaterialDatabase holds various ferroelectric materials
type MaterialDatabase struct {
	Materials map[string]*FerroelectricMaterial
}

// NewMaterialDatabase creates a comprehensive material database
func NewMaterialDatabase() *MaterialDatabase {
	db := &MaterialDatabase{
		Materials: make(map[string]*FerroelectricMaterial),
	}

	// Add standard materials
	db.Materials["HfO2"] = &FerroelectricMaterial{
		Name:               "HfO2",
		Class:              MaterialClassFluorite,
		RemanentPolarization: 20.0,
		CoerciveField:      1.5,
		CurieTemperature:   723.0,
		DielectricConstant: 25.0,
		BandGap:            5.7,
		Endurance:          1e8,
		RetentionTime:      1e10,
		CMOSCompatible:     true,
		Scalable:           true,
	}

	db.Materials["HZO"] = &FerroelectricMaterial{
		Name:               "Hf0.5Zr0.5O2",
		Class:              MaterialClassFluorite,
		RemanentPolarization: 25.0,
		CoerciveField:      1.2,
		CurieTemperature:   673.0,
		DielectricConstant: 30.0,
		BandGap:            5.5,
		Endurance:          1e11, // 10^11 with optimization
		RetentionTime:      1e10,
		CMOSCompatible:     true,
		Scalable:           true,
	}

	db.Materials["AlScN_27"] = &FerroelectricMaterial{
		Name:               "Al0.73Sc0.27N",
		Class:              MaterialClassWurtzite,
		RemanentPolarization: 100.0,
		CoerciveField:      4.0,
		CurieTemperature:   1150.0,
		DielectricConstant: 12.0,
		BandGap:            5.5,
		Endurance:          1e10,
		RetentionTime:      1e12,
		CMOSCompatible:     true,
		Scalable:           true,
	}

	db.Materials["CIPS"] = &FerroelectricMaterial{
		Name:               "CuInP2S6",
		Class:              MaterialClass2DVdW,
		RemanentPolarization: 4.0,
		CoerciveField:      0.2,
		CurieTemperature:   320.0,
		DielectricConstant: 15.0,
		BandGap:            2.9,
		Endurance:          1e6,
		RetentionTime:      5e6,
		CMOSCompatible:     true,
		Scalable:           true,
	}

	db.Materials["In2Se3"] = &FerroelectricMaterial{
		Name:               "α-In2Se3",
		Class:              MaterialClass2DVdW,
		RemanentPolarization: 0.5,
		CoerciveField:      0.15,
		CurieTemperature:   475.0,
		DielectricConstant: 20.0,
		BandGap:            1.3,
		Endurance:          1e7,
		RetentionTime:      1e8,
		CMOSCompatible:     true,
		Scalable:           true,
	}

	db.Materials["PZT"] = &FerroelectricMaterial{
		Name:               "Pb(Zr,Ti)O3",
		Class:              MaterialClassPerovskite,
		RemanentPolarization: 40.0,
		CoerciveField:      0.5,
		CurieTemperature:   620.0,
		DielectricConstant: 1000.0,
		BandGap:            3.4,
		Endurance:          1e12,
		RetentionTime:      1e10,
		CMOSCompatible:     false, // Lead contamination
		Scalable:           false, // Scaling issues
	}

	return db
}

// CompareMaterials compares materials for CIM application
func (db *MaterialDatabase) CompareMaterials(names []string) []map[string]interface{} {
	var results []map[string]interface{}

	for _, name := range names {
		if mat, ok := db.Materials[name]; ok {
			score := db.calculateCIMScore(mat)
			results = append(results, map[string]interface{}{
				"name":       mat.Name,
				"Pr":         mat.RemanentPolarization,
				"Ec":         mat.CoerciveField,
				"Tc":         mat.CurieTemperature,
				"Endurance":  mat.Endurance,
				"CMOSCompat": mat.CMOSCompatible,
				"CIMScore":   score,
			})
		}
	}

	return results
}

// calculateCIMScore computes a figure of merit for CIM applications
func (db *MaterialDatabase) calculateCIMScore(mat *FerroelectricMaterial) float64 {
	score := 0.0

	// Higher Pr is better for memory window
	score += mat.RemanentPolarization / 100.0 * 25.0

	// Lower Ec is better for low power
	score += (5.0 - mat.CoerciveField) / 5.0 * 20.0

	// Higher endurance is critical
	score += math.Log10(float64(mat.Endurance)) / 12.0 * 30.0

	// CMOS compatibility is essential
	if mat.CMOSCompatible {
		score += 15.0
	}

	// Scalability important for density
	if mat.Scalable {
		score += 10.0
	}

	return score
}

// =============================================================================
// ENDURANCE TESTING METHODOLOGY
// =============================================================================

// EnduranceTestConfig configures endurance testing
type EnduranceTestConfig struct {
	DeviceType       string  // "FeFET", "FTJ", "ReRAM"
	TargetCycles     int64
	Temperature      float64 // in °C
	StressVoltage    float64
	PulseWidth       float64 // in ns
	MeasureInterval  int64   // Cycles between measurements
	AcceleratedMode  bool
}

// EnduranceTestResult holds test results
type EnduranceTestResult struct {
	CyclesTested      int64
	FinalMemoryWindow float64
	InitialMemoryWindow float64
	WindowDegradation float64 // percentage
	FailureCycle      int64   // 0 if no failure
	FailureMode       string
	VthShift          float64
	TrappedCharge     float64
}

// EnduranceTester implements endurance testing protocols
type EnduranceTester struct {
	Config           *EnduranceTestConfig
	Results          []*EnduranceTestResult
	CurrentCycle     int64
	MemoryWindowHistory []float64
	VthHistory       []float64
}

// NewEnduranceTester creates endurance tester
func NewEnduranceTester(config *EnduranceTestConfig) *EnduranceTester {
	return &EnduranceTester{
		Config: config,
	}
}

// RunStandardEnduranceTest executes standard endurance test
func (et *EnduranceTester) RunStandardEnduranceTest(initialWindow float64) *EnduranceTestResult {
	result := &EnduranceTestResult{
		InitialMemoryWindow: initialWindow,
		FinalMemoryWindow:   initialWindow,
	}

	window := initialWindow
	vth := 0.0

	for cycle := int64(1); cycle <= et.Config.TargetCycles; cycle++ {
		et.CurrentCycle = cycle

		// Simulate degradation
		degradation := et.simulateCycleDegradation(cycle)
		window *= (1 - degradation)
		vth += et.simulateVthShift(cycle)

		// Record at intervals
		if cycle%et.Config.MeasureInterval == 0 {
			et.MemoryWindowHistory = append(et.MemoryWindowHistory, window)
			et.VthHistory = append(et.VthHistory, vth)
		}

		// Check for failure
		if window < initialWindow*0.1 { // 90% degradation
			result.FailureCycle = cycle
			result.FailureMode = "memory_window_collapse"
			break
		}

		if math.Abs(vth) > 2.0 { // 2V Vth shift
			result.FailureCycle = cycle
			result.FailureMode = "vth_shift_failure"
			break
		}
	}

	result.CyclesTested = et.CurrentCycle
	result.FinalMemoryWindow = window
	result.WindowDegradation = (1 - window/initialWindow) * 100
	result.VthShift = vth
	result.TrappedCharge = vth * 1e-6 // Approximate

	et.Results = append(et.Results, result)
	return result
}

// simulateCycleDegradation models degradation per cycle
func (et *EnduranceTester) simulateCycleDegradation(cycle int64) float64 {
	// Power-law degradation with temperature acceleration
	baseRate := 1e-10

	// Temperature acceleration (Arrhenius)
	tempK := et.Config.Temperature + 273.15
	activation := 0.7 // eV, typical for oxide trapping
	acceleration := math.Exp(-activation / (8.617e-5 * tempK))

	// Voltage acceleration
	voltageAccel := math.Pow(et.Config.StressVoltage/3.0, 2)

	// Fatigue onset (accelerates after ~10^6 cycles)
	fatigueMultiplier := 1.0
	if cycle > 1e6 {
		fatigueMultiplier = 1.0 + math.Log10(float64(cycle)/1e6)
	}

	return baseRate * acceleration * voltageAccel * fatigueMultiplier
}

// simulateVthShift models threshold voltage shift
func (et *EnduranceTester) simulateVthShift(cycle int64) float64 {
	// Log-linear Vth shift
	if cycle < 100 {
		return 0
	}
	return 1e-8 * math.Log10(float64(cycle))
}

// =============================================================================
// ACCELERATED LIFETIME TESTING
// =============================================================================

// HTOLConfig configures High Temperature Operating Life test
type HTOLConfig struct {
	Temperature    float64 // in °C (125, 150, 175 typical)
	Duration       float64 // in hours
	BiasVoltage    float64
	ReadInterval   float64 // hours between reads
	ActivationEnergy float64 // in eV
}

// HTOLTest implements HTOL testing
type HTOLTest struct {
	Config           *HTOLConfig
	AccelerationFactor float64
	ProjectedLifetime float64 // at 25°C in years
	PassFail         bool
	RetentionData    []float64
}

// NewHTOLTest creates HTOL test
func NewHTOLTest(config *HTOLConfig) *HTOLTest {
	htol := &HTOLTest{
		Config: config,
	}

	// Calculate acceleration factor (Arrhenius)
	// AF = exp(Ea/k * (1/T_use - 1/T_stress))
	tUse := 25.0 + 273.15   // 25°C use temperature
	tStress := config.Temperature + 273.15
	k := 8.617e-5 // Boltzmann constant in eV/K

	htol.AccelerationFactor = math.Exp(config.ActivationEnergy / k * (1/tUse - 1/tStress))

	return htol
}

// RunHTOL executes HTOL test
func (htol *HTOLTest) RunHTOL(initialRetention float64) {
	retention := initialRetention
	hours := 0.0

	for hours < htol.Config.Duration {
		hours += htol.Config.ReadInterval

		// Simulate retention degradation
		degradation := 1e-4 * hours / htol.Config.Duration
		retention *= (1 - degradation)

		htol.RetentionData = append(htol.RetentionData, retention)
	}

	// Project lifetime at use temperature
	testLifetimeHours := htol.Config.Duration
	htol.ProjectedLifetime = testLifetimeHours * htol.AccelerationFactor / (24 * 365)

	// Pass if >10 years projected lifetime
	htol.PassFail = htol.ProjectedLifetime > 10
}

// =============================================================================
// AEC-Q100 QUALIFICATION
// =============================================================================

// AECQ100Config configures AEC-Q100 automotive qualification
type AECQ100Config struct {
	Grade            int     // 0, 1, 2, or 3
	MaxTemperature   float64 // Grade 0: 150°C, Grade 1: 125°C, etc.
	MinTemperature   float64
	EnduranceCycles  int64   // 100K typical
	RetentionHours   float64 // 1000 hours typical
}

// AECQ100Qualification implements AEC-Q100 testing
type AECQ100Qualification struct {
	Config           *AECQ100Config
	HTOLResult       bool
	EnduranceResult  bool
	TCResult         bool // Temperature cycling
	HHBResult        bool // High humidity bias
	OverallPass      bool
	Timestamp        string
}

// NewAECQ100Qualification creates AEC-Q100 qualification test
func NewAECQ100Qualification(grade int) *AECQ100Qualification {
	var maxTemp float64
	switch grade {
	case 0:
		maxTemp = 150
	case 1:
		maxTemp = 125
	case 2:
		maxTemp = 105
	case 3:
		maxTemp = 85
	default:
		maxTemp = 85
	}

	return &AECQ100Qualification{
		Config: &AECQ100Config{
			Grade:           grade,
			MaxTemperature:  maxTemp,
			MinTemperature:  -40,
			EnduranceCycles: 100000,
			RetentionHours:  1000,
		},
	}
}

// RunFullQualification runs all AEC-Q100 tests
func (q *AECQ100Qualification) RunFullQualification() {
	// Simulate test results based on Grade 0 requirements
	// (150°C, 100K cycles, -40 to 150°C range)

	// HTOL: 1000 hours at max temp
	q.HTOLResult = rand.Float64() > 0.05 // 95% pass rate

	// Endurance: 100K cycles at max temp
	q.EnduranceResult = rand.Float64() > 0.03 // 97% pass rate

	// Temperature cycling: -40°C to max temp, 1000 cycles
	q.TCResult = rand.Float64() > 0.04 // 96% pass rate

	// High humidity bias: 85°C/85%RH, 1000 hours
	q.HHBResult = rand.Float64() > 0.06 // 94% pass rate

	// Overall pass requires all tests to pass
	q.OverallPass = q.HTOLResult && q.EnduranceResult && q.TCResult && q.HHBResult
}

// =============================================================================
// STATISTICAL ENDURANCE ANALYSIS
// =============================================================================

// EnduranceStatistics holds statistical analysis of endurance data
type EnduranceStatistics struct {
	SampleSize       int
	MeanEndurance    float64
	StdDevEndurance  float64
	MedianEndurance  float64
	WeibullShape     float64 // β (shape parameter)
	WeibullScale     float64 // η (scale parameter)
	B10Life          float64 // 10% failure lifetime
	B50Life          float64 // 50% failure lifetime
	LogNormalMu      float64
	LogNormalSigma   float64
}

// EnduranceAnalyzer performs statistical analysis
type EnduranceAnalyzer struct {
	RawData    []float64 // Endurance cycles per device
	Statistics *EnduranceStatistics
}

// NewEnduranceAnalyzer creates analyzer
func NewEnduranceAnalyzer(data []float64) *EnduranceAnalyzer {
	return &EnduranceAnalyzer{
		RawData: data,
	}
}

// Analyze performs full statistical analysis
func (ea *EnduranceAnalyzer) Analyze() *EnduranceStatistics {
	stats := &EnduranceStatistics{
		SampleSize: len(ea.RawData),
	}

	if len(ea.RawData) == 0 {
		ea.Statistics = stats
		return stats
	}

	// Basic statistics
	sum := 0.0
	for _, v := range ea.RawData {
		sum += v
	}
	stats.MeanEndurance = sum / float64(len(ea.RawData))

	variance := 0.0
	for _, v := range ea.RawData {
		variance += (v - stats.MeanEndurance) * (v - stats.MeanEndurance)
	}
	stats.StdDevEndurance = math.Sqrt(variance / float64(len(ea.RawData)))

	// Median
	sorted := make([]float64, len(ea.RawData))
	copy(sorted, ea.RawData)
	sort.Float64s(sorted)
	stats.MedianEndurance = sorted[len(sorted)/2]

	// Log-normal fit
	logSum := 0.0
	for _, v := range ea.RawData {
		if v > 0 {
			logSum += math.Log(v)
		}
	}
	stats.LogNormalMu = logSum / float64(len(ea.RawData))

	logVariance := 0.0
	for _, v := range ea.RawData {
		if v > 0 {
			logVariance += (math.Log(v) - stats.LogNormalMu) * (math.Log(v) - stats.LogNormalMu)
		}
	}
	stats.LogNormalSigma = math.Sqrt(logVariance / float64(len(ea.RawData)))

	// Weibull estimation (simplified method of moments)
	// β ≈ 1.28 * mean / stddev for many distributions
	if stats.StdDevEndurance > 0 {
		stats.WeibullShape = 1.28 * stats.MeanEndurance / stats.StdDevEndurance
		if stats.WeibullShape < 0.5 {
			stats.WeibullShape = 0.5
		}
		if stats.WeibullShape > 10 {
			stats.WeibullShape = 10
		}
	} else {
		stats.WeibullShape = 3.0 // Default
	}

	// η = mean / Γ(1 + 1/β)
	stats.WeibullScale = stats.MeanEndurance / math.Gamma(1+1/stats.WeibullShape)

	// B10 and B50 lives from Weibull
	stats.B10Life = stats.WeibullScale * math.Pow(-math.Log(0.9), 1/stats.WeibullShape)
	stats.B50Life = stats.WeibullScale * math.Pow(-math.Log(0.5), 1/stats.WeibullShape)

	ea.Statistics = stats
	return stats
}

// GenerateReliabilityReport creates a reliability report
func (ea *EnduranceAnalyzer) GenerateReliabilityReport() map[string]interface{} {
	if ea.Statistics == nil {
		ea.Analyze()
	}

	return map[string]interface{}{
		"sample_size":     ea.Statistics.SampleSize,
		"mean_endurance":  ea.Statistics.MeanEndurance,
		"stddev":          ea.Statistics.StdDevEndurance,
		"median":          ea.Statistics.MedianEndurance,
		"weibull_beta":    ea.Statistics.WeibullShape,
		"weibull_eta":     ea.Statistics.WeibullScale,
		"B10_life":        ea.Statistics.B10Life,
		"B50_life":        ea.Statistics.B50Life,
		"lognormal_mu":    ea.Statistics.LogNormalMu,
		"lognormal_sigma": ea.Statistics.LogNormalSigma,
	}
}

// =============================================================================
// EXTENDED MEASURE-STRESS-MEASURE (eMSM)
// =============================================================================

// EMSMConfig configures extended MSM sequence
type EMSMConfig struct {
	StressTime       float64 // seconds
	MeasureTime      float64 // seconds
	NumCycles        int
	StressVoltage    float64
	MeasureVoltage   float64
	TrapAnalysis     bool
}

// TrapCharacteristics holds trap analysis results
type TrapCharacteristics struct {
	FastTrapDensity  float64 // RA-type (interlayer)
	SlowTrapDensity  float64 // RB-type (HZO layer)
	FastTimeConstant float64 // seconds
	SlowTimeConstant float64 // seconds
	TotalTrappedCharge float64
}

// EMSMSequence implements eMSM testing
type EMSMSequence struct {
	Config           *EMSMConfig
	VthMeasurements  []float64
	DeltaVth         []float64
	Traps            *TrapCharacteristics
}

// NewEMSMSequence creates eMSM sequence
func NewEMSMSequence(config *EMSMConfig) *EMSMSequence {
	return &EMSMSequence{
		Config: config,
	}
}

// RunEMSM executes eMSM sequence
func (emsm *EMSMSequence) RunEMSM(initialVth float64) {
	vth := initialVth

	for i := 0; i < emsm.Config.NumCycles; i++ {
		// Stress phase
		vthAfterStress := vth + emsm.simulateStress()

		// Measure phase (some recovery)
		vthAfterMeasure := vthAfterStress - emsm.simulateRecovery()

		emsm.VthMeasurements = append(emsm.VthMeasurements, vthAfterMeasure)
		emsm.DeltaVth = append(emsm.DeltaVth, vthAfterMeasure-initialVth)

		vth = vthAfterMeasure
	}

	// Trap analysis
	if emsm.Config.TrapAnalysis {
		emsm.analyzeTrap()
	}
}

// simulateStress simulates stress-induced Vth shift
func (emsm *EMSMSequence) simulateStress() float64 {
	// Power-law stress
	return 0.001 * math.Pow(emsm.Config.StressVoltage, 2) * math.Sqrt(emsm.Config.StressTime)
}

// simulateRecovery simulates recovery during measurement
func (emsm *EMSMSequence) simulateRecovery() float64 {
	// Log recovery
	return 0.0001 * math.Log(1+emsm.Config.MeasureTime)
}

// analyzeTrap performs trap analysis
func (emsm *EMSMSequence) analyzeTrap() {
	emsm.Traps = &TrapCharacteristics{
		FastTrapDensity:  1e11, // /cm²
		SlowTrapDensity:  5e10, // /cm²
		FastTimeConstant: 1e-6, // 1 µs
		SlowTimeConstant: 1e-3, // 1 ms
	}

	// Estimate total trapped charge from Vth shift
	if len(emsm.DeltaVth) > 0 {
		finalDelta := emsm.DeltaVth[len(emsm.DeltaVth)-1]
		emsm.Traps.TotalTrappedCharge = finalDelta * 1e-6 / 1.6e-19 // Approximate
	}
}

// =============================================================================
// VOLTAGE DIVIDER FOR VARIABILITY REDUCTION
// =============================================================================

// VoltageDividerConfig configures voltage divider analysis
type VoltageDividerConfig struct {
	SeriesResistance float64 // Ω
	MemristorRon     float64 // Ω
	MemristorRoff    float64 // Ω
	AppliedVoltage   float64 // V
}

// VoltageDividerAnalysis analyzes voltage divider effect
type VoltageDividerAnalysis struct {
	Config              *VoltageDividerConfig
	VoltageOnMemristorSet   float64
	VoltageOnMemristorReset float64
	VariabilityReduction   float64
	EnduranceImprovement   float64
}

// NewVoltageDividerAnalysis creates analysis
func NewVoltageDividerAnalysis(config *VoltageDividerConfig) *VoltageDividerAnalysis {
	vda := &VoltageDividerAnalysis{
		Config: config,
	}

	// Calculate voltage on memristor
	// V_mem = V_applied * R_mem / (R_s + R_mem)
	vda.VoltageOnMemristorSet = config.AppliedVoltage * config.MemristorRon /
		(config.SeriesResistance + config.MemristorRon)

	vda.VoltageOnMemristorReset = config.AppliedVoltage * config.MemristorRoff /
		(config.SeriesResistance + config.MemristorRoff)

	// Dynamic divider suppresses voltage variability
	// Higher Rs → more uniform voltage → lower variability
	ratio := config.SeriesResistance / ((config.MemristorRon + config.MemristorRoff) / 2)
	vda.VariabilityReduction = 1 - 1/(1+ratio)

	// Endurance improves with lower peak voltage on memristor
	vda.EnduranceImprovement = math.Pow(config.AppliedVoltage/vda.VoltageOnMemristorSet, 2)

	return vda
}

// =============================================================================
// ARRAY-LEVEL ENDURANCE TESTING
// =============================================================================

// ArrayEnduranceConfig configures array-level testing
type ArrayEnduranceConfig struct {
	ArrayRows        int
	ArrayCols        int
	TargetCycles     int64
	WearLevelingEnabled bool
	SparingEnabled   bool
	SpareRowPercent  float64
	SpareColPercent  float64
}

// ArrayEnduranceTester tests full crossbar arrays
type ArrayEnduranceTester struct {
	Config           *ArrayEnduranceConfig
	CellEndurance    [][]int64 // Per-cell endurance tracking
	FailedCells      int
	ActiveCells      int
	ArrayLifetime    int64 // Until first uncorrectable failure
	WearDistribution []int64
}

// NewArrayEnduranceTester creates array tester
func NewArrayEnduranceTester(config *ArrayEnduranceConfig) *ArrayEnduranceTester {
	aet := &ArrayEnduranceTester{
		Config:        config,
		CellEndurance: make([][]int64, config.ArrayRows),
		ActiveCells:   config.ArrayRows * config.ArrayCols,
	}

	// Initialize with random endurance per cell (log-normal)
	for i := 0; i < config.ArrayRows; i++ {
		aet.CellEndurance[i] = make([]int64, config.ArrayCols)
		for j := 0; j < config.ArrayCols; j++ {
			// Log-normal with mean 8.6e5, stddev 4.3e4
			mu := math.Log(8.6e5)
			sigma := 0.05
			endurance := math.Exp(mu + sigma*rand.NormFloat64())
			aet.CellEndurance[i][j] = int64(endurance)
		}
	}

	return aet
}

// RunArrayTest executes array-level endurance test
func (aet *ArrayEnduranceTester) RunArrayTest() {
	totalCells := aet.Config.ArrayRows * aet.Config.ArrayCols
	spareRows := int(float64(aet.Config.ArrayRows) * aet.Config.SpareRowPercent)
	spareCols := int(float64(aet.Config.ArrayCols) * aet.Config.SpareColPercent)
	totalSpares := spareRows*aet.Config.ArrayCols + spareCols*aet.Config.ArrayRows

	for cycle := int64(1); cycle <= aet.Config.TargetCycles; cycle++ {
		// Simulate random writes
		row := rand.Intn(aet.Config.ArrayRows)
		col := rand.Intn(aet.Config.ArrayCols)

		// Decrement endurance
		aet.CellEndurance[row][col]--

		// Check for failure
		if aet.CellEndurance[row][col] <= 0 {
			aet.FailedCells++

			if aet.Config.SparingEnabled && aet.FailedCells <= totalSpares {
				// Repair with spare
				aet.CellEndurance[row][col] = 1e6 // New spare cell
			} else if !aet.Config.SparingEnabled || aet.FailedCells > totalSpares {
				// Array failure
				aet.ArrayLifetime = cycle
				break
			}
		}

		// Track wear distribution
		if cycle%1000 == 0 {
			minWear := int64(1e12)
			for i := 0; i < aet.Config.ArrayRows; i++ {
				for j := 0; j < aet.Config.ArrayCols; j++ {
					if aet.CellEndurance[i][j] < minWear {
						minWear = aet.CellEndurance[i][j]
					}
				}
			}
			aet.WearDistribution = append(aet.WearDistribution, minWear)
		}
	}

	if aet.ArrayLifetime == 0 {
		aet.ArrayLifetime = aet.Config.TargetCycles
	}

	aet.ActiveCells = totalCells - aet.FailedCells
}

// GetArrayStatistics returns array-level statistics
func (aet *ArrayEnduranceTester) GetArrayStatistics() map[string]interface{} {
	return map[string]interface{}{
		"array_size":     aet.Config.ArrayRows * aet.Config.ArrayCols,
		"failed_cells":   aet.FailedCells,
		"active_cells":   aet.ActiveCells,
		"array_lifetime": aet.ArrayLifetime,
		"yield":          float64(aet.ActiveCells) / float64(aet.Config.ArrayRows*aet.Config.ArrayCols) * 100,
	}
}

// =============================================================================
// BENCHMARKS AND COMPARISON
// =============================================================================

// MaterialBenchmark compares materials for CIM
type MaterialBenchmark struct {
	Material         string
	EnduranceCycles  int64
	SwitchingSpeed   float64 // ns
	EnergyPerSwitch  float64 // fJ
	ScalabilityNm    float64 // Minimum feature size
	CMOSTemp         float64 // Max process temperature °C
	Score            float64
}

// RunMaterialBenchmarks compares different materials
func RunMaterialBenchmarks() []*MaterialBenchmark {
	benchmarks := []*MaterialBenchmark{
		{
			Material:        "HZO FeFET",
			EnduranceCycles: 1e11,
			SwitchingSpeed:  10,
			EnergyPerSwitch: 10,
			ScalabilityNm:   5,
			CMOSTemp:        400,
			Score:           92,
		},
		{
			Material:        "AlScN FeCAP",
			EnduranceCycles: 1e10,
			SwitchingSpeed:  5,
			EnergyPerSwitch: 5,
			ScalabilityNm:   10,
			CMOSTemp:        350,
			Score:           88,
		},
		{
			Material:        "CIPS FTJ",
			EnduranceCycles: 1e6,
			SwitchingSpeed:  100,
			EnergyPerSwitch: 50,
			ScalabilityNm:   4,
			CMOSTemp:        150, // Low temp process
			Score:           70,
		},
		{
			Material:        "TaOx ReRAM",
			EnduranceCycles: 1e10,
			SwitchingSpeed:  1,
			EnergyPerSwitch: 1,
			ScalabilityNm:   10,
			CMOSTemp:        400,
			Score:           85,
		},
		{
			Material:        "a-SiC ReRAM",
			EnduranceCycles: 1e9,
			SwitchingSpeed:  10,
			EnergyPerSwitch: 10,
			ScalabilityNm:   28,
			CMOSTemp:        400,
			Score:           80,
		},
	}

	return benchmarks
}
