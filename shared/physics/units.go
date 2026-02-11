// Package physics provides shared physics utilities for FeCIM simulations.
// This includes unit formatting, conductance calculations, and physical constants.
package physics

import "fmt"

// Electric field unit conversions.
//
// Internally, simulations store electric field in V/m (SI).
// UI/logs often display in MV/cm, common in ferroelectric literature.
//
// 1 MV/cm = 10^6 V/cm = 10^6 V per 10^-2 m = 10^8 V/m.
const VPerMPerMVPerCm = 1e8

// VPerMToMVPerCm converts electric field from V/m to MV/cm.
func VPerMToMVPerCm(vPerM float64) float64 { return vPerM / VPerMPerMVPerCm }

// MVPerCmToVPerM converts electric field from MV/cm to V/m.
func MVPerCmToVPerM(mvPerCm float64) float64 { return mvPerCm * VPerMPerMVPerCm }

// FormatElectricField formats an electric field given in V/m as MV/cm.
func FormatElectricField(vPerM float64) string {
	return fmt.Sprintf("%.3f MV/cm", VPerMToMVPerCm(vPerM))
}

// FormatEnergy formats energy in Joules with appropriate SI prefix.
// Automatically scales from fJ (femtojoules) to J (joules).
//
// Example:
//
//	FormatEnergy(1.5e-15) // "1.50 fJ"
//	FormatEnergy(2.3e-12) // "2.30 pJ"
//	FormatEnergy(4.5e-9)  // "4.50 nJ"
//	FormatEnergy(6.7e-6)  // "6.70 µJ"
//	FormatEnergy(8.9e-3)  // "8.90 mJ"
//	FormatEnergy(1.2)     // "1.20 J"
func FormatEnergy(joules float64) string {
	switch {
	case joules <= 0:
		return "0 J"
	case joules < 1e-12:
		return fmt.Sprintf("%.2f fJ", joules*1e15)
	case joules < 1e-9:
		return fmt.Sprintf("%.2f pJ", joules*1e12)
	case joules < 1e-6:
		return fmt.Sprintf("%.2f nJ", joules*1e9)
	case joules < 1e-3:
		return fmt.Sprintf("%.2f µJ", joules*1e6)
	case joules < 1:
		return fmt.Sprintf("%.2f mJ", joules*1e3)
	default:
		return fmt.Sprintf("%.2f J", joules)
	}
}

// FormatEnergyMJ formats energy given in millijoules with appropriate SI prefix.
// Convenience wrapper for data already in mJ.
func FormatEnergyMJ(mj float64) string {
	return FormatEnergy(mj * 1e-3)
}

// FormatEnergyUJ formats energy given in microjoules with appropriate SI prefix.
// Convenience wrapper for data already in µJ.
func FormatEnergyUJ(uj float64) string {
	return FormatEnergy(uj * 1e-6)
}

// FormatConductance formats conductance in Siemens with appropriate SI prefix.
// Automatically scales from nS (nanosiemens) to S (siemens).
//
// Example:
//
//	FormatConductance(1e-9)  // "1.00 nS"
//	FormatConductance(50e-6) // "50.00 µS"
//	FormatConductance(1e-3)  // "1.00 mS"
func FormatConductance(siemens float64) string {
	switch {
	case siemens <= 0:
		return "0 S"
	case siemens < 1e-6:
		return fmt.Sprintf("%.2f nS", siemens*1e9)
	case siemens < 1e-3:
		return fmt.Sprintf("%.2f µS", siemens*1e6)
	case siemens < 1:
		return fmt.Sprintf("%.2f mS", siemens*1e3)
	default:
		return fmt.Sprintf("%.2f S", siemens)
	}
}

// FormatCurrent formats current in Amperes with appropriate SI prefix.
// Automatically scales from pA (picoamperes) to A (amperes).
//
// Example:
//
//	FormatCurrent(1e-12) // "1.00 pA"
//	FormatCurrent(50e-9) // "50.00 nA"
//	FormatCurrent(1e-6)  // "1.00 µA"
//	FormatCurrent(1e-3)  // "1.00 mA"
func FormatCurrent(amperes float64) string {
	switch {
	case amperes <= 0:
		return "0 A"
	case amperes < 1e-9:
		return fmt.Sprintf("%.2f pA", amperes*1e12)
	case amperes < 1e-6:
		return fmt.Sprintf("%.2f nA", amperes*1e9)
	case amperes < 1e-3:
		return fmt.Sprintf("%.2f µA", amperes*1e6)
	case amperes < 1:
		return fmt.Sprintf("%.2f mA", amperes*1e3)
	default:
		return fmt.Sprintf("%.2f A", amperes)
	}
}

// FormatVoltage formats voltage in Volts with appropriate SI prefix.
// Automatically scales from µV (microvolts) to V (volts).
//
// Example:
//
//	FormatVoltage(1e-6) // "1.00 µV"
//	FormatVoltage(1e-3) // "1.00 mV"
//	FormatVoltage(1.5)  // "1.50 V"
func FormatVoltage(volts float64) string {
	switch {
	case volts <= 0:
		return "0 V"
	case volts < 1e-3:
		return fmt.Sprintf("%.2f µV", volts*1e6)
	case volts < 1:
		return fmt.Sprintf("%.2f mV", volts*1e3)
	default:
		return fmt.Sprintf("%.2f V", volts)
	}
}

// FormatTime formats time in seconds with appropriate SI prefix.
// Automatically scales from ps (picoseconds) to s (seconds).
//
// Example:
//
//	FormatTime(1e-12) // "1.00 ps"
//	FormatTime(1e-9)  // "1.00 ns"
//	FormatTime(1e-6)  // "1.00 µs"
//	FormatTime(1e-3)  // "1.00 ms"
func FormatTime(seconds float64) string {
	switch {
	case seconds <= 0:
		return "0 s"
	case seconds < 1e-9:
		return fmt.Sprintf("%.2f ps", seconds*1e12)
	case seconds < 1e-6:
		return fmt.Sprintf("%.2f ns", seconds*1e9)
	case seconds < 1e-3:
		return fmt.Sprintf("%.2f µs", seconds*1e6)
	case seconds < 1:
		return fmt.Sprintf("%.2f ms", seconds*1e3)
	default:
		return fmt.Sprintf("%.2f s", seconds)
	}
}

// FormatFrequency formats frequency in Hertz with appropriate SI prefix.
// Automatically scales from Hz to GHz.
//
// Example:
//
//	FormatFrequency(1e3) // "1.00 kHz"
//	FormatFrequency(1e6) // "1.00 MHz"
//	FormatFrequency(1e9) // "1.00 GHz"
func FormatFrequency(hz float64) string {
	switch {
	case hz <= 0:
		return "0 Hz"
	case hz < 1e3:
		return fmt.Sprintf("%.2f Hz", hz)
	case hz < 1e6:
		return fmt.Sprintf("%.2f kHz", hz/1e3)
	case hz < 1e9:
		return fmt.Sprintf("%.2f MHz", hz/1e6)
	default:
		return fmt.Sprintf("%.2f GHz", hz/1e9)
	}
}

// FormatResistance formats resistance in Ohms with appropriate SI prefix.
// Automatically scales from mΩ (milliohms) to GΩ (gigaohms).
//
// Example:
//
//	FormatResistance(0.001)  // "1.00 mΩ"
//	FormatResistance(100)    // "100.00 Ω"
//	FormatResistance(4700)   // "4.70 kΩ"
//	FormatResistance(1e6)    // "1.00 MΩ"
//	FormatResistance(1e9)    // "1.00 GΩ"
func FormatResistance(ohms float64) string {
	switch {
	case ohms <= 0:
		return "0 Ω"
	case ohms < 1:
		return fmt.Sprintf("%.2f mΩ", ohms*1e3)
	case ohms < 1e3:
		return fmt.Sprintf("%.2f Ω", ohms)
	case ohms < 1e6:
		return fmt.Sprintf("%.2f kΩ", ohms/1e3)
	case ohms < 1e9:
		return fmt.Sprintf("%.2f MΩ", ohms/1e6)
	default:
		return fmt.Sprintf("%.2f GΩ", ohms/1e9)
	}
}

// FormatCapacitance formats capacitance in Farads with appropriate SI prefix.
// Automatically scales from aF (attofarads) to F (farads).
//
// Example:
//
//	FormatCapacitance(1e-18) // "1.00 aF"
//	FormatCapacitance(1e-15) // "1.00 fF"
//	FormatCapacitance(1e-12) // "1.00 pF"
//	FormatCapacitance(1e-9)  // "1.00 nF"
//	FormatCapacitance(1e-6)  // "1.00 µF"
//	FormatCapacitance(1e-3)  // "1.00 mF"
func FormatCapacitance(farads float64) string {
	switch {
	case farads <= 0:
		return "0 F"
	case farads < 1e-15:
		return fmt.Sprintf("%.2f aF", farads*1e18)
	case farads < 1e-12:
		return fmt.Sprintf("%.2f fF", farads*1e15)
	case farads < 1e-9:
		return fmt.Sprintf("%.2f pF", farads*1e12)
	case farads < 1e-6:
		return fmt.Sprintf("%.2f nF", farads*1e9)
	case farads < 1e-3:
		return fmt.Sprintf("%.2f µF", farads*1e6)
	case farads < 1:
		return fmt.Sprintf("%.2f mF", farads*1e3)
	default:
		return fmt.Sprintf("%.2f F", farads)
	}
}

// FormatPower formats power in Watts with appropriate SI prefix.
// Automatically scales from fW (femtowatts) to kW (kilowatts).
//
// Example:
//
//	FormatPower(1e-15) // "1.00 fW"
//	FormatPower(1e-12) // "1.00 pW"
//	FormatPower(1e-9)  // "1.00 nW"
//	FormatPower(1e-6)  // "1.00 µW"
//	FormatPower(1e-3)  // "1.00 mW"
//	FormatPower(1.5)   // "1.50 W"
//	FormatPower(1500)  // "1.50 kW"
func FormatPower(watts float64) string {
	switch {
	case watts <= 0:
		return "0 W"
	case watts < 1e-12:
		return fmt.Sprintf("%.2f fW", watts*1e15)
	case watts < 1e-9:
		return fmt.Sprintf("%.2f pW", watts*1e12)
	case watts < 1e-6:
		return fmt.Sprintf("%.2f nW", watts*1e9)
	case watts < 1e-3:
		return fmt.Sprintf("%.2f µW", watts*1e6)
	case watts < 1:
		return fmt.Sprintf("%.2f mW", watts*1e3)
	case watts < 1e3:
		return fmt.Sprintf("%.2f W", watts)
	default:
		return fmt.Sprintf("%.2f kW", watts/1e3)
	}
}

// FormatCharge formats electric charge in Coulombs with appropriate SI prefix.
// Automatically scales from fC (femtocoulombs) to C (coulombs).
//
// Example:
//
//	FormatCharge(1e-15) // "1.00 fC"
//	FormatCharge(1e-12) // "1.00 pC"
//	FormatCharge(1e-9)  // "1.00 nC"
//	FormatCharge(1e-6)  // "1.00 µC"
//	FormatCharge(1e-3)  // "1.00 mC"
func FormatCharge(coulombs float64) string {
	switch {
	case coulombs <= 0:
		return "0 C"
	case coulombs < 1e-12:
		return fmt.Sprintf("%.2f fC", coulombs*1e15)
	case coulombs < 1e-9:
		return fmt.Sprintf("%.2f pC", coulombs*1e12)
	case coulombs < 1e-6:
		return fmt.Sprintf("%.2f nC", coulombs*1e9)
	case coulombs < 1e-3:
		return fmt.Sprintf("%.2f µC", coulombs*1e6)
	case coulombs < 1:
		return fmt.Sprintf("%.2f mC", coulombs*1e3)
	default:
		return fmt.Sprintf("%.2f C", coulombs)
	}
}

// FormatPolarization formats polarization in C/m² as µC/cm² (standard ferroelectric units).
// This is a domain-specific format commonly used in ferroelectric literature.
//
// Example:
//
//	FormatPolarization(0.20)  // "20.0 µC/cm²"
//	FormatPolarization(0.35)  // "35.0 µC/cm²"
//	FormatPolarization(0.001) // "0.1 µC/cm²"
func FormatPolarization(cm2 float64) string {
	// Convert C/m² to µC/cm²: multiply by 100 (1 C/m² = 100 µC/cm²)
	microCcm2 := cm2 * 100
	if microCcm2 <= 0 {
		return "0 µC/cm²"
	}
	if microCcm2 >= 100 {
		return fmt.Sprintf("%.0f µC/cm²", microCcm2)
	}
	if microCcm2 >= 10 {
		return fmt.Sprintf("%.1f µC/cm²", microCcm2)
	}
	return fmt.Sprintf("%.2f µC/cm²", microCcm2)
}

// FormatElectricField formats electric field in V/m as MV/cm or kV/cm.
// This is the standard format for ferroelectric coercive fields.
//
// Example:
//
//	FormatElectricField(1e8)   // "1.00 MV/cm"
//	FormatElectricField(5e7)   // "500.00 kV/cm"
//	FormatElectricField(1.5e8) // "1.50 MV/cm"
func FormatElectricField(vm float64) string {
	if vm <= 0 {
		return "0 V/m"
	}
	// Convert V/m to MV/cm: divide by 1e8 (1 MV/cm = 1e8 V/m)
	mvCm := vm / 1e8
	if mvCm >= 1 {
		return fmt.Sprintf("%.2f MV/cm", mvCm)
	}
	// Use kV/cm for smaller fields
	kvCm := vm / 1e5
	if kvCm >= 1 {
		return fmt.Sprintf("%.2f kV/cm", kvCm)
	}
	return fmt.Sprintf("%.2f V/cm", vm/1e2)
}
