# Configuration Reference

**Config schema version:** 1.0.0
**Go package:** `config/physics` (`physics.Config` struct)
**Source of truth:** `config/physics/physics.go`

## Overview

FeCIM Lattice Tools uses split YAML configuration files under `config/`. Each file controls a specific simulation domain and maps to a top-level field in the `physics.Config` struct. The config system supports three loading modes:

1. **Split files** (`config/*.yaml`) -- preferred for production and experimentation.
2. **Legacy monolith** (`config/physics.yaml`) -- deprecated; parsed into the same `Config` struct.
3. **Embedded defaults** (`config/physics/defaults/*.yaml`) -- automatic fallback with `[WARN]` log.

The loader probes `config/`, `../config/`, and `../../config/` relative to the working directory. The first directory that contains any split file (other than `materials.yaml`) wins.

---

## Config Files

### 1. constants.yaml

Global physics constants shared by all modules.

**Go struct:** `physics.Constants`
**YAML root key:** `constants`

| Field | YAML key | Type | Default | Unit | Simulation effect |
|-------|----------|------|---------|------|-------------------|
| FeCIMLevels | `fecim_levels` | int | *(omitted; fallback 30)* | -- | Number of discrete conductance levels for quantization. Controls `crossbar.QuantizeTo30Levels()` granularity and level-mapping in the Preisach calibration loop. When absent, `Material.GetNumLevels()` returns 30. |
| BitsPerCell | `bits_per_cell` | float64 | 4.907 | bits | Information density per cell (`log2(30) = 4.907`). Used in energy-per-bit and storage-density calculations. |
| BoltzmannEV | `boltzmann_ev` | float64 | 8.617e-5 | eV/K | Boltzmann constant in eV. Used in thermal activation (KAI model), Arrhenius switching, and temperature-dependent Landau coefficient scaling. |
| Epsilon0 | `epsilon_0` | float64 | 8.854e-12 | F/m | Vacuum permittivity. Enters dielectric field calculations, depolarization field computation, and capacitance scaling. |
| RoomTemperature | `room_temperature` | float64 | 300 | K | Reference temperature for all temperature-dependent models (L-K solver alpha coefficient, KAI switching dynamics, retention extrapolation). |

**Notes:**
- `fecim_levels` is intentionally omitted from the default YAML. The per-material `analog_states` field takes priority; when that is also zero, the code falls back to 30.

---

### 2. materials.yaml

Ferroelectric material parameter sets. Each key under `materials:` defines a named material profile that can be selected at runtime via `cfg.GetMaterial("name")`.

**Go struct:** `physics.Material` (map value)
**YAML root key:** `materials`

The file ships with 7 preset materials:

| Key | Name | States | Description |
|-----|------|--------|-------------|
| `default_hzo` | HZO (Si-doped) | 30 | Baseline standard HZO from literature (Park 2015) |
| `fecim_hzo` | FeCIM HZO | 30 | Dr. Tour's FeCIM demonstrated values (conservative) |
| `fecim_hzo_target` | FeCIM HZO (TARGET) | 30 | Aspirational targets (NOT demonstrated) |
| `literature_superlattice` | Literature Superlattice | 64 | Best academic HfO2/ZrO2 superlattice (Cheema 2020) |
| `cryogenic_hzo` | Cryogenic HZO | 48 | HZO at 4 K for quantum computing integration |
| `hzo_standard_32` | HZO Standard (32 states) | 32 | Oh IEEE EDL 2017 peer-reviewed 32-state demo |
| `hzo_ftj_140` | HZO FTJ (140 states) | 140 | Song Adv. Science 2024 FTJ with >7-bit operation |
| `alscn` | AlScN (8-16 states) | 16 | High-Pr AlScN, limited granularity due to high Ec |

#### Material field reference

**Metadata fields:**

| Field | YAML key | Type | Simulation effect |
|-------|----------|------|-------------------|
| Name | `name` | string | Display label in GUI material selector |
| Description | `description` | string | Tooltip / info text |
| Reference | `reference` | string | Literature citation |
| AnalogStates | `analog_states` | int | Overrides global `fecim_levels` for this material; determines quantization grid |
| TargetRangeFrac | `target_range_frac` | float64 | Fraction of Ps used for outer-level targets (0..1); prevents saturation clipping |
| TRLLevel | `trl_level` | int | Technology Readiness Level (1-9); informational |
| CMOSCompatible | `cmos_compatible` | bool | Flags CMOS fabrication compatibility; informational |

**Polarization (C/m^2):**

| Field | YAML key | Type | Simulation effect |
|-------|----------|------|-------------------|
| PrCM2 | `pr_c_m2` | float64 | Remanent polarization. Sets the stable storage states in the hysteresis loop; directly scales the P-E curve height and level separation. |
| PsCM2 | `ps_c_m2` | float64 | Saturation polarization. Upper bound of the L-K free-energy potential; defines the maximum achievable polarization and the full-scale conductance window. |

**Field (V/m):**

| Field | YAML key | Type | Simulation effect |
|-------|----------|------|-------------------|
| EcVM | `ec_v_m` | float64 | Coercive field. Determines the field threshold for polarization switching; sets Preisach distribution center and calibration voltage range. `CoerciveVoltage() = Ec * thickness`. |
| MemoryWindowV | `memory_window_v` | float64 | Memory window voltage. Informational for FTJ/cryo materials; used in voltage margin calculations. |

**Dielectric properties:**

| Field | YAML key | Type | Simulation effect |
|-------|----------|------|-------------------|
| EpsilonHF | `epsilon_hf` | float64 | High-frequency permittivity. Enters dielectric displacement computation and depolarization field model. |
| EpsilonLF | `epsilon_lf` | float64 | Low-frequency permittivity. Used in quasi-static capacitance and L-K alpha coefficient calculation. |
| LossTangent | `loss_tangent` | float64 | Dielectric loss tan-delta. Adds dissipative damping in the AC hysteresis loop; affects energy loss per cycle. |

**Geometry:**

| Field | YAML key | Type | Simulation effect |
|-------|----------|------|-------------------|
| ThicknessM | `thickness_m` | float64 | Film thickness (m). Converts field to voltage (`V = E * d`); scales capacitance and switching energy. |
| AreaM2 | `area_m2` | float64 | Cell area (m^2). Scales charge (`Q = P * A`), energy, and current in device-level simulations. |
| CellPitchNm | `cell_pitch_nm` | float64 | Cell pitch (nm). Informational; used in area-efficiency metrics. |

**Switching dynamics:**

| Field | YAML key | Type | Simulation effect |
|-------|----------|------|-------------------|
| TauS | `tau_s` | float64 | Characteristic switching time (s). KAI model time constant; governs transient domain nucleation/growth speed. |
| Tau0S | `tau0_s` | float64 | Attempt frequency inverse (s). Pre-exponential in Arrhenius expression for thermally activated switching. |
| ActivationEnergyEV | `activation_energy_ev` | float64 | Thermal activation barrier (eV). Exponential factor in switching probability; higher values slow switching at a given temperature. |
| KAIExponent | `kai_exponent` | float64 | KAI/Avrami exponent. Domain growth dimensionality (1=1D, 2=2D, 3=3D); shapes the switching transient curve. |

**Temperature:**

| Field | YAML key | Type | Simulation effect |
|-------|----------|------|-------------------|
| CurieTempK | `curie_temp_k` | float64 | Curie temperature (K). Phase transition point; enters L-K alpha coefficient as `alpha = (T - Tc) / C`. |
| TempCoeffEc | `temp_coeff_ec` | float64 | dEc/dT (V/m/K). Shifts coercive field with temperature in thermal sweep simulations. |
| TempCoeffPr | `temp_coeff_pr` | float64 | dPr/dT (C/m^2/K). Shifts remanent polarization with temperature. |
| OperatingTempK | `operating_temp_k` | float64 | Operating temperature (K). For cryogenic profiles; overrides `room_temperature` in thermal models when set. |

**Reliability:**

| Field | YAML key | Type | Simulation effect |
|-------|----------|------|-------------------|
| EnduranceCycles | `endurance_cycles` | float64 | Maximum cycling before degradation. Used by Preisach fatigue model to scale hysteron density over lifetime. |
| RetentionTimeS | `retention_time_s` | float64 | Data retention at operating temp (s). Informational; enters retention-extrapolation calculations. |
| ImprintFieldVM | `imprint_field_v_m` | float64 | DC-bias-induced asymmetry (V/m). Shifts the hysteresis loop center horizontally; models aging. |

**Depolarization:**

| Field | YAML key | Type | Simulation effect |
|-------|----------|------|-------------------|
| DepolarizationFactorVMC | `depolarization_factor_vm_c` | float64 | Depolarization coefficient (V*m/C). Creates the "slanted" P-E curve characteristic of polycrystalline films; `E_dep = -k_dep * P`. Essential for multi-level analog operation. |

**FTJ-specific:**

| Field | YAML key | Type | Simulation effect |
|-------|----------|------|-------------------|
| TERRatio | `ter_ratio` | float64 | Tunneling electroresistance ratio. On/off resistance ratio for tunnel junctions. |
| GmaxGminRatio | `gmax_gmin_ratio` | float64 | Conductance on/off ratio. Alternative to explicit Gmin/Gmax for FTJ devices. |

**AlScN-specific:**

| Field | YAML key | Type | Simulation effect |
|-------|----------|------|-------------------|
| ScFraction | `sc_fraction` | float64 | Scandium fraction in AlScN (e.g. 0.36 for Al0.64Sc0.36N). |

**In2Se3-specific (2D ferroelectric):**

| Field | YAML key | Type | Simulation effect |
|-------|----------|------|-------------------|
| PrInplaneCM2 | `pr_inplane_c_m2` | float64 | In-plane Pr for 2D materials |
| EcThinVM | `ec_thin_v_m` | float64 | Ec for ultrathin films |
| EcThickVM | `ec_thick_v_m` | float64 | Ec for thicker films |
| BandgapEV | `bandgap_ev` | float64 | Bandgap for semiconductor FE |
| MinThicknessM | `min_thickness_m` | float64 | Minimum viable thickness |
| QuintupleLayerNm | `quintuple_layer_nm` | float64 | Quintuple-layer thickness for 2D |
| VdWMaterial | `vdw_material` | bool | Van der Waals layered flag |
| Stacking | `stacking` | string | Stacking type (3R, 2H) |
| Phase | `phase` | string | Crystal phase (alpha, beta) |
| AlphaToBetaTempK | `alpha_to_beta_temp_k` | float64 | Phase transition temperature (K) |
| LinearityImprovement | `linearity_improvement` | bool | Better linearity at cryo |

**Sub-structs (nested under each material):**

**`thermodynamics:` (Landau-Khalatnikov)**

| Field | YAML key | Type | Default (HZO) | Simulation effect |
|-------|----------|------|----------------|-------------------|
| BetaLandau | `beta_landau` | float64 | -6.720e8 | First-order barrier coefficient (J*m^5/C^4). Negative value creates the double-well free energy landscape; depth controls bistability. |
| GammaLandau | `gamma_landau` | float64 | 1.950e10 | Sixth-order stability coefficient (J*m^9/C^6). Positive value prevents runaway polarization at high fields. |
| RhoViscosity | `rho_viscosity` | float64 | 0.05 | Viscous damping (Ohm*m). Controls L-K solver transient speed; lower values give faster switching. |
| CurieConstK | `curie_const_k` | float64 | 1.5e5 | Curie constant (K). Scales the temperature-dependent alpha coefficient. |

**`coupling:` (Electrostriction)**

| Field | YAML key | Type | Default (HZO) | Simulation effect |
|-------|----------|------|----------------|-------------------|
| Q11Electrostriction | `q11_electrostriction` | float64 | 0.089 | Longitudinal electrostriction (m^4/C^2). Couples polarization to in-plane strain; affects stress-dependent switching. |
| Q12Electrostriction | `q12_electrostriction` | float64 | -0.026 | Transverse electrostriction (m^4/C^2). Cross-coupling term. |
| StressGPa | `stress_gpa` | float64 | 1.0 | In-plane biaxial stress (GPa). **Note:** The L-K solver uses Pa internally; multiply by 1e9 when converting. |

**`circuit:` (Parasitics)**

| Field | YAML key | Type | Default (HZO) | Simulation effect |
|-------|----------|------|----------------|-------------------|
| SeriesResistanceOhm | `series_resistance_ohm` | float64 | 50 | Series resistance (Ohm). Creates voltage divider with ferroelectric capacitor; adds IR drop to applied voltage. |

**`nls:` (Nucleation-Limited Switching)**

| Field | YAML key | Type | Default (HZO) | Simulation effect |
|-------|----------|------|----------------|-------------------|
| ActivationFieldVM | `activation_field_v_m` | float64 | 1.2e9 | Merz-law activation field (V/m). Exponential switching barrier in NLS model. |
| TauInfS | `tau_inf_s` | float64 | 1.0e-10 | Attempt time (s). Pre-exponential in NLS switching rate. |
| Sigma | `sigma` | float64 | *(optional)* | Log-normal spread of switching field distribution. |

**`conductance:` (Polarization-to-conductance mapping)**

| Field | YAML key | Type | Default (HZO) | Simulation effect |
|-------|----------|------|----------------|-------------------|
| GminS | `gmin_s` | float64 | 1.0e-6 | Minimum conductance / HRS (S). Sets the floor of the weight range in crossbar MVM. |
| GmaxS | `gmax_s` | float64 | 100.0e-6 | Maximum conductance / LRS (S). Sets the ceiling of the weight range. |
| OnOffRatio | `on_off_ratio` | float64 | *(optional)* | Gmax/Gmin ratio. |
| Model | `model` | string | `"linear"` | Conductance model: `linear`, `subthreshold`, or `saturation`. |
| KvT | `k_vt` | float64 | *(optional)* | Threshold voltage scale (V) for P-to-Vt coupling. |
| VGSReadV | `vgs_read_v` | float64 | *(optional)* | Effective read gate voltage (V) for saturation model. |
| VT0V | `vt0_v` | float64 | *(optional)* | Zero-polarization threshold voltage (V). |

**`synaptic:` (Neuromorphic device parameters)**

| Field | YAML key | Type | Simulation effect |
|-------|----------|------|-------------------|
| PotentiationPulses | `potentiation_pulses` | int | Number of LTP pulses |
| DepressionPulses | `depression_pulses` | int | Number of LTD pulses |
| PulseWidthS | `pulse_width_s` | float64 | Programming pulse width (s) |
| PulseVoltageV | `pulse_voltage_v` | float64 | Programming voltage (V) |
| NonlinearityLTP | `nonlinearity_ltp` | float64 | LTP nonlinearity factor |
| NonlinearityLTD | `nonlinearity_ltd` | float64 | LTD nonlinearity factor |

**`synthesis:` (Fabrication method)**

| Field | YAML key | Type | Simulation effect |
|-------|----------|------|-------------------|
| Method | `method` | string | Synthesis method (FWF, ALD, etc.) |
| Scale | `scale` | string | Production scale |
| Precursors | `precursors` | []string | Precursor materials |
| SynthesisTimeS | `synthesis_time_s` | float64 | Synthesis time (s) |
| EnergyReduction | `energy_reduction` | float64 | Energy reduction vs conventional |

---

### 3. crossbar.yaml

Crossbar array configuration for matrix-vector multiply (MVM) simulations.

**Go struct:** `physics.Crossbar`
**YAML root key:** `crossbar`

| Field | YAML key | Type | Default | Unit | Simulation effect |
|-------|----------|------|---------|------|-------------------|
| DefaultRows | `default_rows` | int | 128 | -- | Number of word lines in the crossbar. Sets input vector dimension for MVM. |
| DefaultCols | `default_cols` | int | 128 | -- | Number of bit lines. Sets output vector dimension for MVM. |
| QuantizationLevels | `quantization_levels` | int | 30 | -- | Conductance quantization resolution. Must match `fecim_levels` / `analog_states`. Controls weight precision. |
| ADCBits | `adc_bits` | int | 8 | bits | ADC resolution for reading column currents. Limits output precision; adds quantization noise. |
| DACBits | `dac_bits` | int | 8 | bits | DAC resolution for input voltage encoding. Limits input precision. |
| ConductanceMinS | `conductance_min_s` | float64 | 1.0e-6 | S | Minimum device conductance (HRS). Floor of the weight range; lower values improve on/off ratio. |
| ConductanceMaxS | `conductance_max_s` | float64 | 100.0e-6 | S | Maximum device conductance (LRS). Ceiling of the weight range; higher values increase column current. |
| ConductanceRatio | `conductance_ratio` | float64 | 100 | -- | Gmax/Gmin on/off ratio. Determines weight dynamic range. |
| DeviceVariation | `device_variation` | float64 | 0.05 | -- | Device-to-device sigma (fraction). Adds Gaussian noise to programmed conductance values; models fabrication variability. |
| ReadNoise | `read_noise` | float64 | 0.02 | -- | Read noise sigma (fraction). Adds per-read Gaussian noise; models thermal/shot noise. |
| WriteNoise | `write_noise` | float64 | 0.03 | -- | Write noise sigma (fraction). Adds noise to programmed values during write; models stochastic switching. |
| WordLineResistanceOhm | `word_line_resistance_ohm` | float64 | 10 | Ohm | Word line resistance per segment. Drives IR-drop simulation along rows; larger arrays see more voltage droop. |
| BitLineResistanceOhm | `bit_line_resistance_ohm` | float64 | 10 | Ohm | Bit line resistance per segment. Drives IR-drop simulation along columns. |
| SneakPathEnabled | `sneak_path_enabled` | bool | true | -- | Enables sneak-path leakage model. When true, unselected cells contribute parasitic current to column sums. |
| SneakConductanceRatio | `sneak_conductance_ratio` | float64 | 0.01 | -- | Leakage-to-on conductance ratio for sneak paths. Lower values reduce MVM error from unselected cells. |

**Notes:**
- The default YAML also includes informational fields (`max_demonstrated_rows`, `mlc_levels`, `ter_ratio`, `fcm_sneak_free`) that are not mapped to Go struct fields. These are documentation-only entries.

---

### 4. training.yaml

Neural network training configuration for the MNIST and crossbar-aware training demos.

**Go struct:** `physics.Training`
**YAML root key:** `training`

| Field | YAML key | Type | Default | Simulation effect |
|-------|----------|------|---------|-------------------|
| LearningRate | `learning_rate` | float64 | 0.01 | Optimizer step size per batch. Controls convergence speed vs. stability. |
| WeightDecay | `weight_decay` | float64 | 1.0e-4 | L2 regularization coefficient. Penalizes large weights; improves generalization. |
| Momentum | `momentum` | float64 | 0.9 | SGD momentum factor. Accelerates convergence along consistent gradient directions. 0 disables. |
| DefaultBatchSize | `default_batch_size` | int | 32 | Samples per optimizer step. Larger batches reduce gradient noise; smaller batches improve generalization. |
| GradientClip | `gradient_clip` | float64 | 1.0 | Maximum gradient L2 norm. Prevents exploding gradients; clips above this threshold. |
| WeightClipMin | `weight_clip_min` | float64 | 0.0 | Minimum weight value, maps to Gmin. Clamps weights to the hardware-achievable conductance floor. |
| WeightClipMax | `weight_clip_max` | float64 | 1.0 | Maximum weight value, maps to Gmax. Clamps weights to the hardware-achievable conductance ceiling. |
| UpdateNoiseSigma | `update_noise_sigma` | float64 | 0.01 | Noise standard deviation added to weight updates. Models stochastic write variation in hardware. |
| AsymmetryRatio | `asymmetry_ratio` | float64 | 1.0 | Potentiation/depression asymmetry. 1.0 = symmetric; <1 means depression is weaker than potentiation. |
| QuantizeForward | `quantize_forward` | bool | true | Quantize weights in the forward pass. Simulates inference with discrete conductance levels. |
| QuantizeBackward | `quantize_backward` | bool | false | Quantize weights in the backward pass. Usually false to preserve gradient precision. |
| StraightThrough | `straight_through` | bool | true | Use Straight-Through Estimator (STE) for quantization gradients. Allows gradient flow through non-differentiable quantization. |

---

### 5. energy.yaml

Per-operation energy parameters and technology comparison baselines.

**Go struct:** `physics.Energy`
**YAML root key:** `energy`

| Field | YAML key | Type | Default | Unit | Simulation effect |
|-------|----------|------|---------|------|-------------------|
| ReadEnergyJ | `read_energy_j` | float64 | 1.0e-15 | J | Energy per read operation (~1 fJ). Used in energy-efficiency reporting and technology comparison charts. |
| WriteEnergyJ | `write_energy_j` | float64 | 10.0e-15 | J | Energy per write operation (~10 fJ). Dominates training energy budget; shown in per-epoch energy summaries. |
| MACEnergyJ | `mac_energy_j` | float64 | 5.0e-15 | J | Energy per multiply-accumulate (~5 fJ). Primary metric for inference energy efficiency (TOPS/W). |
| NANDWriteEnergyJ | `nand_write_energy_j` | float64 | 500.0e-15 | J | NAND flash write energy (~500 fJ). Comparison baseline for module5 technology charts. |
| DRAMAccessEnergyJ | `dram_access_energy_j` | float64 | 5000.0e-15 | J | DRAM access energy (~5 pJ). Comparison baseline. |
| SRAMAccessEnergyJ | `sram_access_energy_j` | float64 | 50.0e-15 | J | SRAM access energy (~50 fJ). Comparison baseline. |
| StandbyPowerWPerCell | `standby_power_w_per_cell` | float64 | 1.0e-12 | W | Standby leakage per cell (~1 pW). Used in idle-power and array-level power estimates. |

**Notes:**
- The default YAML also includes informational fields (`puf_readout_energy_j`, `cim_area_efficiency_gops_mm2`, `cim_energy_efficiency_tops_w`) that are present in the YAML but not mapped to Go struct fields. These serve as reference benchmarks in comments.

---

### 6. timing.yaml

Operation latency and programming pulse parameters.

**Go struct:** `physics.Timing`
**YAML root key:** `timing`

| Field | YAML key | Type | Default | Unit | Simulation effect |
|-------|----------|------|---------|------|-------------------|
| ReadLatencyS | `read_latency_s` | float64 | 10.0e-9 | s | Read latency (10 ns). Sets the time cost of inference reads in throughput calculations. |
| WriteLatencyS | `write_latency_s` | float64 | 100.0e-9 | s | Write latency (100 ns). Includes program-verify; dominates training wall-clock time. |
| MACLatencyS | `mac_latency_s` | float64 | 10.0e-9 | s | MAC operation latency (10 ns). Sets the per-operation time for MVM throughput estimates. |
| PulseWidthS | `pulse_width_s` | float64 | 10.0e-9 | s | Programming pulse width (10 ns). Duration of each voltage pulse in ISPP write sequences. |
| VerifyDelayS | `verify_delay_s` | float64 | 5.0e-9 | s | Verify delay after each program pulse (5 ns). Settling time before read-back verification. |
| MaxProgramCycles | `max_program_cycles` | int | 10 | -- | Maximum program-verify iterations. Limits the ISPP loop; if the target level is not reached in this many cycles, the write is abandoned. |

**Notes:**
- The default YAML also includes informational fields (`switching_time_typical_s`, `switching_time_record_s`) that document literature records but are not mapped to Go struct fields.

---

### 7. preisach.yaml

Classical Preisach hysteresis model parameters.

**Go struct:** `physics.Preisach`
**YAML root key:** `preisach`

| Field | YAML key | Type | Default | Unit | Simulation effect |
|-------|----------|------|---------|------|-------------------|
| GridSize | `grid_size` | int | 30 | -- | Hysteron grid resolution (N x N). Higher values give smoother hysteresis curves but cost O(N^2) memory and computation. |
| AlphaSigmaRatio | `alpha_sigma_ratio` | float64 | 0.20 | -- | sigma_alpha / Ec ratio (20%). Width of the switching-up field distribution; wider distributions create more gradual (slanted) loops. |
| BetaSigmaRatio | `beta_sigma_ratio` | float64 | 0.20 | -- | sigma_beta / Ec ratio (20%). Width of the switching-down field distribution. |
| Correlation | `correlation` | float64 | 0.5 | -- | Alpha-beta correlation coefficient (0..1). Controls coupling between up and down switching fields in the Preisach plane. |
| FatigueRate | `fatigue_rate` | float64 | 1.0e-10 | 1/cycle | Degradation rate per switching cycle. Reduces hysteron density over lifetime; affects long-term endurance simulations. |
| WakeupCycles | `wakeup_cycles` | int | 100 | cycles | Number of initial cycles for wake-up effect. During wake-up, the Preisach distribution evolves toward steady state. |
| InitialWakeup | `initial_wakeup` | float64 | 0.8 | -- | Initial wake-up factor (0..1). Fraction of full polarization available at cycle zero; ramps to 1.0 over `wakeup_cycles`. |

---

### 8. calibration.yaml

Level calibration parameters for the Preisach-based multi-level programming loop.

**Go struct:** `physics.Calibration`
**YAML root key:** `calibration`

| Field | YAML key | Type | Default | Unit | Simulation effect |
|-------|----------|------|---------|------|-------------------|
| Iterations | `iterations` | int | 15 | -- | Binary search iterations for finding the voltage that produces each target polarization level. More iterations improve precision at the cost of calibration time. |
| FieldMinRatio | `field_min_ratio` | float64 | 0.7 | -- | Read voltage ceiling as fraction of coercive voltage (`Vread_max = 0.7 * Vc`). Ensures non-destructive reads with a 30% safety margin below the coercive voltage. |
| FieldMaxRatio | `field_max_ratio` | float64 | 2.5 | -- | Maximum applied field as fraction of Ec (`Emax = 2.5 * Ec`). Upper bound of the calibration search range; must exceed saturation. |
| AdjustmentRate | `adjustment_rate` | float64 | 0.06 | -- | Adaptive adjustment rate (6% of Ec per level error). Controls how aggressively the calibration corrects for level mismatch on each iteration. |
| LevelTolerance | `level_tolerance` | int | 1 | levels | Acceptable level error. A programmed state within +/-1 level of the target is considered a successful write. |

---

### 9. simulation.yaml

Simulation runtime parameters for the GUI and waveform engine.

**Go struct:** `physics.Simulation`
**YAML root key:** `simulation`

| Field | YAML key | Type | Default | Unit | Simulation effect |
|-------|----------|------|---------|------|-------------------|
| FrameRateHz | `frame_rate_hz` | int | 60 | Hz | UI refresh rate. Controls how frequently the Fyne canvas redraws; higher values give smoother animation but cost more CPU. |
| DtS | `dt_s` | float64 | 0.0167 | s | Simulation time step (1/60 s). The L-K solver and Preisach engine advance by this delta each frame. Smaller values improve accuracy but slow real-time playback. |
| MaxHistoryPoints | `max_history_points` | int | 500 | -- | P-E plot trail length. Number of (E, P) samples retained for the hysteresis loop trace. Older points are discarded in FIFO order. |
| DefaultFrequencyHz | `default_frequency_hz` | float64 | 0.5 | Hz | Default waveform frequency. The initial frequency of the applied electric field oscillation in the hysteresis module. |
| DefaultAmplitudeRatio | `default_amplitude_ratio` | float64 | 1.5 | -- | Default E-field amplitude as a multiple of Ec (`E_peak = 1.5 * Ec`). Ensures the field is large enough to drive full switching without excessive over-saturation. |

---

### 10. mnist.yaml

MNIST handwritten digit recognition demo parameters.

**Go struct:** `physics.MNIST`
**YAML root key:** `mnist`

| Field | YAML key | Type | Default | Simulation effect |
|-------|----------|------|---------|-------------------|
| InputSize | `input_size` | int | 784 | Input layer size (28x28 pixels flattened). Must match MNIST image dimensions. |
| HiddenSizes | `hidden_sizes` | []int | [128, 64] | Hidden layer sizes. Defines the MLP architecture; each entry adds a fully-connected layer. |
| OutputSize | `output_size` | int | 10 | Output layer size (10 digit classes, 0-9). |
| Epochs | `epochs` | int | 10 | Number of full training passes over the dataset. |
| BatchSize | `batch_size` | int | 64 | Samples per optimization batch (overrides `training.default_batch_size` for the MNIST demo). |
| LearningRate | `learning_rate` | float64 | 0.001 | Optimizer learning rate for the MNIST demo (overrides `training.learning_rate`). |
| BaselineAccuracy | `baseline_accuracy` | float64 | 0.966 | Reference accuracy from peer-reviewed literature (Nature Commun. 2023). Displayed as a comparison baseline; NOT a target for this simulator. |
| TourClaimedAccuracy | `tour_claimed_accuracy` | float64 | *(removed)* | Previously 0.87; removed as unverified conference claim. Field exists in Go struct but is no longer set in default YAML. |

**Notes:**
- The default YAML also includes informational fields (`best_accuracy`, `current_limited_accuracy`, `fashion_mnist_rc`) that document external benchmarks but are not mapped to Go struct fields.

---

### 11. benchmarks.yaml

Arbitrary key-value benchmark data for reference and comparison charts.

**Go struct:** `map[string]any`
**YAML root key:** `benchmarks`

| Key (example) | Type | Default | Description |
|----------------|------|---------|-------------|
| `matrix_precision` | string | `"fp32"` | Precision achieved in analog matrix operations |
| `matrix_size_max` | int | 16 | Maximum matrix dimension demonstrated with high precision |
| `rc_energy_vs_gpu` | int | 10000 | Energy reduction factor vs. GPU for reservoir computing |
| `rc_endurance_cycles` | int | 100000 | Endurance cycles demonstrated in reservoir computing |
| `puf_ml_attack_resistance` | int | 10000000 | PUF resistance to ML attacks (number of training samples) |
| `puf_reconfigurability` | string | `"near_ideal"` | Qualitative PUF reconfigurability assessment |

This file uses a free-form `map[string]any` schema. You may add arbitrary key-value pairs without changing the Go code. Values are accessed via `cfg.Benchmarks["key"]`.

---

## Loading Priority

The `physics.Load()` function resolves configuration in this order:

1. **Split config files** in `config/` (relative to working directory, also probes `../config/` and `../../config/`). The loader considers split mode active if **any** probe file exists (all files except `materials.yaml` are probed). Missing individual files within a split set fall through to embedded defaults for that section.
2. **Legacy `config/physics.yaml`** monolith. Parsed into the same `Config` struct. Supported but deprecated.
3. **Embedded defaults** (`config/physics/defaults/*.yaml`, compiled into the binary via `//go:embed`). Used when no external config directory is found. Logged with `[WARN]` level.

Within a single split-file load, if an external file is missing for a given section, the embedded default for that section is used as a silent fallback.

### Load functions

| Function | Behavior on error |
|----------|-------------------|
| `Load()` | Returns `(*Config, error)`. Singleton; cached after first call. |
| `MustLoad()` | Calls `log.Fatalf` on error. **Deprecated** -- use `Load()` instead. |
| `LoadWithDefaults()` | Returns a valid `*Config` even on error, falling back through embedded defaults to a minimal hardcoded config (30 levels, 4.9 bits/cell, 300 K). Suitable for GUI code. |
| `Reload()` | Resets the singleton and re-loads from disk. Useful in tests. |

---

## For Reproducibility

To ensure another researcher can reproduce your simulation results:

1. Record the config schema version: `cfg.Version()` returns `"1.0.0"`.
2. Include a copy of all 11 active YAML files (or the legacy monolith) in your reproducibility pack.
3. Use the `--seed` flag (when available) for deterministic RNG in stochastic models (device variation, read/write noise, Preisach fatigue).
4. Document the material profile name used (e.g., `default_hzo`, `fecim_hzo`).
5. Note any overrides applied at runtime (e.g., environment variables, CLI flags).

A minimal reproducibility manifest:

```
repro-pack/
  config/
    constants.yaml
    materials.yaml
    crossbar.yaml
    training.yaml
    energy.yaml
    timing.yaml
    preisach.yaml
    calibration.yaml
    simulation.yaml
    mnist.yaml
    benchmarks.yaml
  config_version.txt       # Contains "1.0.0"
  go.sum                   # Dependency lockfile
  seed.txt                 # RNG seed used
  results/
    ...
```
