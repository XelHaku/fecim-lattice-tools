package physics

// CharacterizationResult captures transiently characterized operation metrics
// for downstream timing/power export (e.g., Liberty generation).
type CharacterizationResult struct {
	WriteTimeNs    float64 // Time for P to reach 90% of Pr
	ReadTimeNs     float64 // Time for TIA output to settle within 1 LSB
	WriteEnergy_fJ float64
	ReadEnergy_fJ  float64
}
