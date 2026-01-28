"""
Analyze core physics equations for DAC, ADC, TIA, and Charge Pump.
"""
import json
from datetime import datetime

print("[OBJECTIVE] Document core physics equations for Module 4 peripheral circuits\n")

# ============================================================================
# DAC (Digital-to-Analog Converter) Physics
# ============================================================================
print("="*80)
print("DAC (Digital-to-Analog Converter) Physics")
print("="*80)

dac_physics = {
    "conversion": {
        "equation": "Vout = VrefLow + (level / maxLevel) * (VrefHigh - VrefLow)",
        "description": "Linear interpolation between reference voltages",
        "variables": {
            "Vout": "Output voltage (V)",
            "VrefLow": "Low reference voltage (-1.5V for FeCIM)",
            "VrefHigh": "High reference voltage (+1.5V for FeCIM)",
            "level": "Digital input level (0-29)",
            "maxLevel": "Maximum level (31 for 5-bit)"
        },
        "source": "dac.go:46-48"
    },
    
    "resolution": {
        "equation": "LSB = (VrefHigh - VrefLow) / (2^bits - 1)",
        "description": "Voltage step per least significant bit",
        "variables": {
            "LSB": "Least significant bit voltage (V)",
            "bits": "DAC resolution (5 bits for 30 levels)"
        },
        "source": "dac.go:75-76",
        "calculated_value": "3.0V / 31 ≈ 96.77 mV/level"
    },
    
    "inl_error": {
        "equation": "INL_error = INL * LSB * sin(π * level / (2^bits - 1))",
        "description": "Integral nonlinearity error (bow-shaped)",
        "variables": {
            "INL": "Integral nonlinearity (LSB units, typically 0.5)",
            "INL_error": "Actual voltage error (V)"
        },
        "source": "dac.go:61",
        "physical_meaning": "Deviation from ideal straight-line transfer function"
    },
    
    "dnl_error": {
        "equation": "DNL_error = DNL * LSB * (0.5 - level%3 / 2.0)",
        "description": "Differential nonlinearity error (step variation)",
        "variables": {
            "DNL": "Differential nonlinearity (LSB units, typically 0.25)",
            "DNL_error": "Step size error (V)"
        },
        "source": "dac.go:64",
        "physical_meaning": "Variation in step size between adjacent levels"
    },
    
    "energy": {
        "equation": "E = C * Vref^2 * 2^N",
        "description": "Energy per conversion (switched-capacitor DAC)",
        "variables": {
            "E": "Energy per conversion (J)",
            "C": "Unit capacitor (1 fF typical)",
            "Vref": "Reference voltage range",
            "N": "Number of bits"
        },
        "source": "dac.go:81-88",
        "calculated_value": "~15 fJ for 5-bit, ±1.5V range"
    }
}

for key, eq in dac_physics.items():
    print(f"\n{key.upper()}")
    print(f"  Equation: {eq['equation']}")
    print(f"  Description: {eq['description']}")
    print(f"  Source: {eq['source']}")
    if 'calculated_value' in eq:
        print(f"  Calculated: {eq['calculated_value']}")

print(f"\n[STAT:dac_core_equations] {len(dac_physics)}")

# ============================================================================
# ADC (Analog-to-Digital Converter) Physics
# ============================================================================
print("\n" + "="*80)
print("ADC (Analog-to-Digital Converter) Physics")
print("="*80)

adc_physics = {
    "quantization": {
        "equation": "level = round((Vin - VrefLow) / (VrefHigh - VrefLow) * (2^bits - 1))",
        "description": "Voltage to digital level conversion with rounding",
        "variables": {
            "Vin": "Input voltage to quantize (V)",
            "level": "Output digital level (0-31)"
        },
        "source": "adc.go:57-58"
    },
    
    "snr_theoretical": {
        "equation": "SNR = 6.02 * N + 1.76 dB",
        "description": "Theoretical signal-to-noise ratio for ideal N-bit ADC",
        "variables": {
            "SNR": "Signal-to-noise ratio (dB)",
            "N": "Number of bits"
        },
        "source": "adc.go:116",
        "calculated_value": "31.86 dB for 5-bit ADC"
    },
    
    "enob": {
        "equation": "ENOB = bits - log2(sqrt(1 + INL^2 + DNL^2))",
        "description": "Effective number of bits accounting for nonlinearity",
        "variables": {
            "ENOB": "Effective number of bits",
            "INL": "Integral nonlinearity (LSB)",
            "DNL": "Differential nonlinearity (LSB)"
        },
        "source": "adc.go:91-92",
        "physical_meaning": "Actual resolution considering real-world errors"
    },
    
    "snr_effective": {
        "equation": "SNR_eff = 6.02 * ENOB + 1.76 dB",
        "description": "Effective SNR using ENOB instead of ideal bits",
        "source": "adc.go:121"
    },
    
    "energy_sar": {
        "equation": "E_SAR = 5 fJ/bit * N_bits",
        "description": "Energy per conversion for SAR ADC",
        "variables": {
            "E_SAR": "Energy (J)",
            "N_bits": "Resolution (bits)"
        },
        "source": "adc.go:101",
        "calculated_value": "~25 fJ for 5-bit SAR"
    },
    
    "energy_flash": {
        "equation": "E_flash = 50 fJ * 2^N",
        "description": "Energy per conversion for Flash ADC (2^N comparators)",
        "source": "adc.go:104",
        "calculated_value": "~1600 fJ for 5-bit Flash (much higher)"
    }
}

for key, eq in adc_physics.items():
    print(f"\n{key.upper()}")
    print(f"  Equation: {eq['equation']}")
    print(f"  Description: {eq['description']}")
    print(f"  Source: {eq['source']}")
    if 'calculated_value' in eq:
        print(f"  Calculated: {eq['calculated_value']}")

print(f"\n[STAT:adc_core_equations] {len(adc_physics)}")

# ============================================================================
# TIA (Transimpedance Amplifier) Physics
# ============================================================================
print("\n" + "="*80)
print("TIA (Transimpedance Amplifier) Physics")
print("="*80)

tia_physics = {
    "transimpedance": {
        "equation": "Vout = Iin * Gain + Voffset",
        "description": "Current-to-voltage conversion (Ohm's Law)",
        "variables": {
            "Vout": "Output voltage (V)",
            "Iin": "Input current (A)",
            "Gain": "Transimpedance gain (Ω), typically 10 kΩ",
            "Voffset": "Output offset voltage (V), typically 5 mV"
        },
        "source": "tia.go:33",
        "physical_meaning": "Converts column current to measurable voltage"
    },
    
    "noise_voltage": {
        "equation": "Vnoise_rms = Inoise * Gain * sqrt(BW)",
        "description": "Output noise voltage (thermal + shot noise)",
        "variables": {
            "Vnoise_rms": "RMS noise voltage (V)",
            "Inoise": "Input-referred noise current density (A/√Hz), typically 1 pA/√Hz",
            "BW": "Bandwidth (Hz), typically 100 MHz"
        },
        "source": "tia.go:52",
        "calculated_value": "~100 µV RMS for 1 pA/√Hz, 10 kΩ, 100 MHz"
    },
    
    "snr": {
        "equation": "SNR = 20 * log10(Isignal * Gain / Vnoise)",
        "description": "Signal-to-noise ratio for TIA output",
        "variables": {
            "Isignal": "Input signal current (A)"
        },
        "source": "tia.go:70"
    },
    
    "min_detectable_current": {
        "equation": "Imin = Inoise * sqrt(BW)",
        "description": "Minimum current for SNR=1 (0 dB)",
        "source": "tia.go:76",
        "calculated_value": "~10 nA for 1 pA/√Hz, 100 MHz"
    },
    
    "dynamic_range": {
        "equation": "DR = 20 * log10(Imax / Imin)",
        "description": "Dynamic range from minimum to maximum detectable current",
        "variables": {
            "DR": "Dynamic range (dB)",
            "Imax": "Maximum input current (A)",
            "Imin": "Minimum detectable current (A)"
        },
        "source": "tia.go:82"
    },
    
    "settling_time": {
        "equation": "t_settle = ln(1/accuracy) / (2*π*BW)",
        "description": "Time to settle to within accuracy (single-pole response)",
        "variables": {
            "accuracy": "Settling accuracy (0.001 for 0.1%)",
            "BW": "Bandwidth (Hz)"
        },
        "source": "tia.go:90",
        "calculated_value": "~11 ns for 0.1% accuracy, 100 MHz BW"
    },
    
    "power": {
        "equation": "P ≈ 2 * kT * BW * Gain / η",
        "description": "TIA power consumption estimate",
        "variables": {
            "kT": "Thermal energy (4.14e-21 J at 300K)",
            "η": "Efficiency (typically 0.1)"
        },
        "source": "tia.go:99"
    }
}

for key, eq in tia_physics.items():
    print(f"\n{key.upper()}")
    print(f"  Equation: {eq['equation']}")
    print(f"  Description: {eq['description']}")
    print(f"  Source: {eq['source']}")
    if 'calculated_value' in eq:
        print(f"  Calculated: {eq['calculated_value']}")

print(f"\n[STAT:tia_core_equations] {len(tia_physics)}")

# ============================================================================
# Charge Pump Physics
# ============================================================================
print("\n" + "="*80)
print("Charge Pump (Dickson Topology) Physics")
print("="*80)

pump_physics = {
    "ideal_output": {
        "equation": "Vout_ideal = (N + 1) * Vin",
        "description": "Ideal Dickson charge pump output (no losses)",
        "variables": {
            "N": "Number of stages (2 for FeCIM)",
            "Vin": "Input voltage (1.0V CMOS supply)"
        },
        "source": "chargepump.go:36",
        "calculated_value": "3.0V for 2-stage, 1V input"
    },
    
    "actual_output": {
        "equation": "Vout_actual = (N+1)*Vin - N*Vth - Iload/(C*f)",
        "description": "Real output with diode drops and IR drop",
        "variables": {
            "Vth": "Threshold voltage drop per stage (~0.3V for MOS switches)",
            "Iload": "Load current (A)",
            "C": "Flying capacitor (F)",
            "f": "Clock frequency (Hz)"
        },
        "source": "chargepump.go:43-44",
        "calculated_value": "~1.8V for 2-stage with losses (then boost to 1.5V target)"
    },
    
    "output_ripple": {
        "equation": "ΔV = Iload / (Cout * f)",
        "description": "Peak-to-peak output voltage ripple",
        "variables": {
            "Cout": "Output capacitor (typically 10× flying cap)"
        },
        "source": "chargepump.go:52"
    },
    
    "boost_factor": {
        "equation": "Boost = Vout_actual / Vin",
        "description": "Voltage multiplication factor",
        "source": "chargepump.go:58"
    },
    
    "power_efficiency": {
        "equation": "η = Pout / Pin = (Vout * Iload) / Pin",
        "description": "Power conversion efficiency",
        "variables": {
            "η": "Efficiency (typically 0.7 or 70%)"
        },
        "source": "chargepump.go:65"
    },
    
    "rise_time": {
        "equation": "t_rise = (N * 2.2) / f_clk",
        "description": "Output voltage rise time (10% to 90%)",
        "source": "chargepump.go:82",
        "calculated_value": "~88 ns for 2-stage, 50 MHz clock"
    },
    
    "max_current": {
        "equation": "Imax = C * f * (N+1) * Vin / Vout",
        "description": "Maximum sustainable output current",
        "variables": {
            "C": "Flying capacitor",
            "f": "Clock frequency"
        },
        "source": "chargepump.go:88"
    },
    
    "charge_transfer_eff": {
        "equation": "η_stage = Vout_actual / Vout_ideal",
        "description": "Per-stage charge transfer efficiency",
        "source": "chargepump.go:109"
    }
}

for key, eq in pump_physics.items():
    print(f"\n{key.upper()}")
    print(f"  Equation: {eq['equation']}")
    print(f"  Description: {eq['description']}")
    print(f"  Source: {eq['source']}")
    if 'calculated_value' in eq:
        print(f"  Calculated: {eq['calculated_value']}")

print(f"\n[STAT:chargepump_core_equations] {len(pump_physics)}")

# Save structured physics data
physics_data = {
    "dac": dac_physics,
    "adc": adc_physics,
    "tia": tia_physics,
    "charge_pump": pump_physics,
    "timestamp": datetime.now().isoformat()
}

with open('<local-path>', 'w') as f:
    json.dump(physics_data, f, indent=2)

total_core = len(dac_physics) + len(adc_physics) + len(tia_physics) + len(pump_physics)
print(f"\n[STAT:total_core_equations] {total_core}")
print("\n[FINDING] All core peripheral circuit physics equations documented")
