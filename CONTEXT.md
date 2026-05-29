# FeCIM Lattice Tools

Shared language for the FeCIM Lattice Tools UI migration and scientific simulation workspace.

## Language

**Default UI Surface**:
The supported user-facing application surface for normal use, testing, screenshots, and documentation. It is the place where migrated module behavior must appear.
_Avoid_: next UI, future shell, experimental shell

**Legacy Fyne Surface**:
The deprecated historical UI surface retained only as a temporary migration and parity reference. It must not be presented as the default, recommended, or released user path.
_Avoid_: default GUI, recommended Fyne UI, current desktop app

**Migration Parity**:
A behavior is available in the Default UI Surface with equivalent user-observable workflow, state, controls, outputs, and boundary notices before the Legacy Fyne Surface is removed.
_Avoid_: visual similarity only, screenshot match, placeholder parity, computation-only parity

**Level Calibration Workflow**:
A user workflow for mapping discrete polarization or conductance levels to write/read field values for a selected material, level count, target range, and temperature. The first migration slice exposes status and recalculation before persisted calibration artifacts. It is not literature fitting and not device validation.
_Avoid_: calibration proof, literature fitting, hardware calibration, measured-device confirmation

**Level Calibration Engine**:
A UI-neutral domain service responsible for computing Level Calibration Workflow data independently of any user interface surface. It provides deterministic, summary-oriented results before persisted calibration artifacts are part of the workflow.
_Avoid_: Fyne-backed calibration, GUI service, hidden legacy dependency

**Level Calibration State**:
The user-visible condition of the Level Calibration Workflow for the current material, level count, target range, and calibration temperature: not calibrated, stale, or fresh. Stale means a previous Level Calibration Summary no longer matches the current inputs.
_Avoid_: validation status, hardware readiness, measurement confidence

**Level Calibration Inputs**:
The selected material, level count, target range, and calibration temperature that define a Level Calibration Workflow run. Inputs must be meaningful simulation conditions before a Level Calibration Summary can be fresh.
_Avoid_: hidden defaults, silently corrected calibration, measured setup

**Level Count**:
The number of discrete simulated levels used for Module 1 level mapping. It is a simulation discretization input, not a claim about demonstrated hardware states.
_Avoid_: hardware states, measured levels

**Target Range**:
The fraction of the simulated polarization range used as the outer target span for level mapping. It controls headroom for level calibration and write/read demonstrations.
_Avoid_: operating guarantee, measured voltage margin

**Calibration Temperature**:
The temperature input used by the Level Calibration Workflow for temperature-dependent level mapping. It is a simulation condition, not an environmental qualification claim.
_Avoid_: qualified temperature, validated operating temperature

**Level Calibration Summary**:
A concise user-visible description of the current level calibration inputs, state, method, and result quality. It communicates that a simulation mapping was computed without exposing full per-level lookup tables as the primary workflow.
_Avoid_: full calibration dataset, measured calibration report, device characterization table

**Level Calibration Detail**:
A compact user-visible explanation of a Level Calibration Summary, optionally including representative first, middle, and last level rows. It supports understanding and export confidence without becoming the authoritative full lookup table.
_Avoid_: full lookup-table parity, measured calibration table, device characterization dataset

**Level Calibration Method**:
The declared simulation model used to compute a Level Calibration Summary. It must be visible enough that users do not confuse one simulated mapping method with another or with measured-device calibration.
_Avoid_: hidden calibration path, measured method, device-proven method

**Level Calibration Error**:
A user-visible explanation that a Level Calibration Workflow run could not produce a fresh summary for the selected inputs. It does not replace Level Calibration State and does not erase a prior result.
_Avoid_: failed validation, failed measurement, invalid hardware

**Level Calibration Export**:
An explicit user action that writes Level Calibration Workflow artifacts for review or reuse. It is separate from running calibration and must carry the same simulation boundary language as the summary. It exports only a fresh Level Calibration Summary whose results match the current inputs, and is the preferred way to expose detailed per-level mapping data before adding dense on-screen lookup tables.
_Avoid_: automatic persistence, hidden cache write, measured calibration certificate

**Educational Simulation Boundary**:
The rule that FeCIM results are model estimates for education and design exploration, not validated device or silicon measurements.
_Avoid_: hardware proof, measured advantage, demonstrated device result

**ISPP Convergence Policy Module**:
A UI-neutral module that owns shared ISPP write-verify convergence decisions such as bounds recovery, overshoot handling, guard acceptance, and convergence receipts. It is consumed by waveform-based Module 1 adapters and L-K physics adapters so bug fixes stay local.
_Avoid_: duplicated writer heuristics, GUI-specific convergence logic, hidden controller tweak

**ISPP Convergence Receipt**:
A structured result from the ISPP Convergence Policy Module that explains what convergence rule was applied, what bounds or status changed, and whether a full-range reset was used. It supports debugging and regression tests without exposing every writer implementation detail.
_Avoid_: ad-hoc log-only decision, unexplained bounds mutation, silent convergence shortcut

## Example dialogue

Developer: "Can we delete this Legacy Fyne Surface panel now?"
Domain expert: "Only after the Default UI Surface has Migration Parity for its controls, plots, exports, and boundary notices."

Developer: "The screenshot looks better; is that parity?"
Domain expert: "No. Screenshot readability helps, but Migration Parity means the same user-observable behavior is present in the Default UI Surface."

Developer: "Can the Module 1 plot claim measured HZO behavior?"
Domain expert: "No. Keep the Educational Simulation Boundary visible unless a validated measurement citation is attached."
