# Cell Geometry Decision: FeCIM Bit Cell for OpenLane Integration

## Decision Summary

**Selected Option: A - Placeholder Cell (0.46um x 2.72um) for Simulation-Only**

## Rationale

### Why Placeholder Cell?

1. **TRL 4 Reality**: Per Dr. Tour's presentation, FeCIM technology is at Technology Readiness Level 4. Designing actual transistor-level cells would imply manufacturing readiness that doesn't exist.

2. **Educational Focus**: The FeCIM Visualizer is an educational tool demonstrating compute-in-memory concepts. A placeholder cell allows:
   - Valid OpenLane flow testing
   - DEF/LEF syntax validation
   - Array-level architecture exploration
   - Without pretending to have fabrication-ready designs

3. **SKY130 Compatibility Gap**: The SKY130 PDK lacks ferroelectric materials. Any "real" FeCIM cell in SKY130 would be a fiction. A clearly-labeled placeholder is more honest.

4. **Rapid Iteration**: Placeholder cells allow quick experimentation with array architectures (passive vs 1T1R) without getting stuck in cell-level design.

## Cell Specifications

```
Cell Name:    fecim_bit
Technology:   SKY130 (placeholder)
Width:        0.46 um (0.46 = 2 * SKY130 site width of 0.23um)
Height:       2.72 um (matches SKY130 standard cell row height)
Pins:
  - WL (input): Word Line connection (left edge)
  - BL (inout): Bit Line connection (top edge)
```

### Dimension Rationale

- **Height (2.72um)**: Standard SKY130 row height, ensuring compatibility with row-based placement
- **Width (0.46um)**: 2x site width (0.23um), minimum for a cell with 2 pins while maintaining routing track access

## Future Considerations

When FeCIM technology advances to TRL 7+:
1. Replace stub LEF with characterized library cell
2. Add timing/power models (.lib file)
3. Include actual transistor-level GDS layout

## Implementation Files

- `cells/fecim_bit.stub.lef` - Abstract LEF with pin definitions only
- No GDS file (placeholder has no physical implementation)
- No .lib file (no timing characterization for placeholder)

## Sign-off

This decision aligns with:
- Dr. Tour's emphasis on demonstrating the *concept* of 30-level FeCIM
- Educational visualization goals of the project
- Practical OpenLane integration requirements

---
*Document created for FeCIM Lattice Generator Phase 0*
*Date: 2026-01-23*
