# M4-INV-04 Results — ENOB Refinement (target ≈12.75)

Source tests:
- `shared/peripherals/m4_investigations_test.go::TestM4INV04_ThermalNoiseVsADCRefine`
- `shared/peripherals/noise_vs_adc_analysis_test.go::TestNoiseVsADC_M4INV04_UsefulADCCeilingRecommendation`

Conditions: `Vrange=1.8V`, `R_TIA=10kΩ`, `BW=10MHz`, `T=300K`.

## Key outputs

- ENOB sweep includes:
  - 12-bit ADC → `ENOB = 11.931`
  - 13-bit ADC → `ENOB = 12.753`
  - 14-bit ADC → `ENOB = 13.299`
- Useful ADC ceiling message from noise analysis:
  - **"13 bits (ENOB=12.75), diminishing returns above 14 bits"**

Conclusion: refined investigation confirms prior target **ENOB ≈ 12.75** at the practical 13-bit operating point under current noise assumptions.