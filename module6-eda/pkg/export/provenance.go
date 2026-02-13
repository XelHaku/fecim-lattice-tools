package export

const provenanceWarningText = "NOT suitable for signoff or tapeout. Educational/research use only."

const characterizationProvenanceBlockC = `/* CHARACTERIZATION PROVENANCE
 * CHAR_STATUS: ESTIMATED_SCALAR
 * WARNING: This Liberty file is generated from analytical models,
 * NOT from silicon measurement or SPICE post-layout simulation.
 * Timing values are estimated from Landau-Khalatnikov transient simulation.
 * NOT suitable for signoff or tapeout. Educational/research use only.
 * For production characterization, use Cadence Liberate or Synopsys SiliconSmart
 * with measured device data.
 */
`

const characterizationProvenanceBlockHash = `# CHARACTERIZATION PROVENANCE
# CHAR_STATUS: ESTIMATED_SCALAR
# WARNING: This Liberty file is generated from analytical models,
# NOT from silicon measurement or SPICE post-layout simulation.
# Timing values are estimated from Landau-Khalatnikov transient simulation.
# NOT suitable for signoff or tapeout. Educational/research use only.
# For production characterization, use Cadence Liberate or Synopsys SiliconSmart
# with measured device data.
`

const characterizationProvenanceBlockSlash = `// CHARACTERIZATION PROVENANCE
// CHAR_STATUS: ESTIMATED_SCALAR
// WARNING: This Liberty file is generated from analytical models,
// NOT from silicon measurement or SPICE post-layout simulation.
// Timing values are estimated from Landau-Khalatnikov transient simulation.
// NOT suitable for signoff or tapeout. Educational/research use only.
// For production characterization, use Cadence Liberate or Synopsys SiliconSmart
// with measured device data.
`

const characterizationProvenanceBlockSpice = `* CHARACTERIZATION PROVENANCE
* CHAR_STATUS: ESTIMATED_SCALAR
* WARNING: This Liberty file is generated from analytical models,
* NOT from silicon measurement or SPICE post-layout simulation.
* Timing values are estimated from Landau-Khalatnikov transient simulation.
* NOT suitable for signoff or tapeout. Educational/research use only.
* For production characterization, use Cadence Liberate or Synopsys SiliconSmart
* with measured device data.
`
