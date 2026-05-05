# Papers To Read

Move records here after a paper file exists but before it has been read.

## Core Verification (read first)

- [ ] park2015_advmat_hzo — verify Pr=24, Ec=1.0 from Fig 2a against digitized data
  - Evidence match: `experimental-data/hzo/pe-loops/park2015_advmat_hzo_10nm_fig2a.json`
- [ ] materlik2015_jap_hfo2_origin — verify β=-6.72e8, γ=1.95e10 from Table I
  - Evidence match: LKSolver defaults in `shared/physics/landau.go`
- [ ] guo2018_apl_nls — verify NLSSigma=1.5, TauInf=100ps
  - Evidence match: NLS defaults in `shared/physics/landau.go`
- [ ] alessandri2018_ieee_edl_switching — verify NLS activation field Ea=1.9 MV/cm
  - Evidence match: LKSolver.ActivationField default

## Calibration Data Verification (read second)

- [ ] cheema2020_nature_hzo_superlattice — verify Pr=22 µC/cm² at 5nm
  - Feed: LiteratureSuperlattice material preset
- [ ] kim2020_materials_tin_hzo — verify wake-up cycle count
  - Feed: HZOStandard32 material preset
- [ ] jaiswal2021_crystals_bto — verify BTO Pr
  - Feed: BTO material preset calibration
- [ ] bi2024_nano_pzt_thinfilm — verify PZT thin film Pr
  - Feed: PZT material preset calibration

## External Validation (read third)

- [ ] crosssim2024_sandia — verify SOR solver output format
  - Feed: `crosssim_reference_8x8.json` test vectors
- [ ] soliman2023_ncomms_multilevel — verify 4-state FeFET crossbar
  - Feed: `module2-crossbar` multi-level programming model
