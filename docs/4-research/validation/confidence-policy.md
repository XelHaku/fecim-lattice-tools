# External Validation Confidence Policy

This policy defines how FeCIM internal model outputs are compared against external tools and references.

## Agreement bands

For **quantitative** comparisons, compute relative error against a reference:

- **Green**: within **5%**
- **Yellow**: **>5% to 20%**
- **Red**: **>20%**

Use the same bands for each metric below, with metric-specific notes.

| Metric | Comparison type | Quantitative method | Notes |
|---|---|---|---|
| Remanent polarization (Pr) | Quantitative | `|Pr_model - Pr_ref| / |Pr_ref|` | Compare at matched pulse/relax conditions and temperature. |
| Coercive field (Ec) | Quantitative | `|Ec_model - Ec_ref| / |Ec_ref|` | Compare extraction method consistently (loop midpoint/criterion). |
| Switching time | Quantitative | `|t_sw_model - t_sw_ref| / |t_sw_ref|` | Compare at same voltage/field and target polarization fraction. |
| Read margin | Quantitative | `|RM_model - RM_ref| / |RM_ref|` | Use common definition (e.g., normalized ΔI or ΔV sense margin). |
| Energy per write | Quantitative | `|Ew_model - Ew_ref| / |Ew_ref|` | Include same pulse train and peripheral assumptions. |

## Trend-only vs quantitative comparisons

Some external comparisons are not numerically equivalent because model abstractions differ.

### Quantitative allowed

- Same metric definition
- Same operating point (material, geometry, temperature, bias, timing)
- Compatible extraction method

### Trend-only only

Use trend-only (directional agreement) when any of the below is true:

- Different physical abstraction level (compact model vs array-level approximations)
- Missing matched operating conditions
- Different extraction/measurement definitions not reconcilable
- Sparse external output (shape/ordering only, no absolute scale)

Trend-only outcomes should be reported as:

- **Match**: direction/ordering matches
- **Partial match**: mostly matches with noted deviations
- **Mismatch**: trend direction or ordering disagrees

## Reporting requirements

For each comparator run, report:

1. tool + version
2. metric definition
3. operating point and assumptions
4. numeric delta and band (for quantitative)
5. trend verdict (when trend-only)
6. baseline file identifier used
