# M4-INV-05 Results — Dickson Charge Pump Efficiency at 3V (SKY130)

Source test: `shared/peripherals/m4_investigations_test.go::TestM4INV05_ChargePumpDicksonEfficiencyAt3V`

Configuration used:
- `Vin=1.8V`, `Vout_target=3.0V`
- Dickson stages: `2`
- `DiodeDrop=0.25V`, `fclk=100MHz`
- `Iload=50µA`, `Cfly=200pF`

## Results

- Actual boosted output: `Vout_actual = 3.000V`
- Charge-transfer efficiency: `stage_eff = 0.556`
- System efficiency (`Pout/Pin`): `system_eff = 0.720`

Conclusion: under the current compact SKY130 model and load point, 3V generation is achievable with ~72% system efficiency and ~55.6% transfer efficiency relative to ideal stage gain.