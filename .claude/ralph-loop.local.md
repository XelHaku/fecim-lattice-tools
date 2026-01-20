---
active: true
iteration: 1
max_iterations: 0
completion_promise: null
started_at: "2026-01-20T00:51:15Z"
---

## IronLattice Demo Upgrade - Live Slide Pattern

MISSION: Upgrade demos to 'live presentation slides' following Demo 1 pattern.

REFERENCE FILES:
- PLAYBOOK.md (project handbook)
- ironlattice-transcript.md (Dr. Tour's presentation)
- TODO.md (task tracking)

CURRENT PRIORITY: Demo 2 (Crossbar MVM) → Demo 3 (MNIST) → Demo 4 (Circuits)

LIVE SLIDE CHECKLIST (per demo):
□ 1. Visual Anchor (hero element)
□ 2. Main Plot (the science)
□ 3. 30-Level Indicator
□ 4. Controls (mode, speed, parameters)
□ 5. Educational Panel ('What You're Seeing' - dynamic, phase-aware)
□ 6. Operation Log (timestamped actions)
□ 7. Mode Indicator (IDLE/WRITE/READ/COMPUTE)
□ 8. Input/Output Displays (vectors, values)
□ 9. Dr. Tour Quote
□ 10. Status Bar

DEMO MODES:
□ Auto Mode (self-running)
□ Demo Mode (step-by-step educational)
□ Manual Mode (full control)

DEMO 2 SPECIFIC:
- Visual Anchor: Crossbar grid with animated current flow
- Input Vector: V₀, V₁, V₂... (editable bars)
- Output Vector: I₀, I₁, I₂... (result bars)
- Educational: '1. Voltages applied → 2. I = G × V → 3. Output currents'
- Key Stat: 'N² multiplications in 1 cycle'
- Quote: 'Compute in memory where the same device does memory and computation'

DEMO 3 SPECIFIC:
- Visual Anchor: Drawing canvas + big prediction number
- Layer Flow: Input (784) → Hidden (128) → Output (10)
- Educational: '1. Pixels → 2. MVM × 2 → 3. Prediction'
- Key Stat: '87% accuracy (88% theoretical max)'
- Quote: 'We're at 87% validation here'

DEMO 4 SPECIFIC:
- Visual Anchor: Signal flow diagram (DAC → Cell → ADC)
- Timing Diagram: Animated write/read cycle
- Educational: '1. Digital → 2. Voltage → 3. Program → 4. Read'
- Key Stat: 'Standard CMOS compatible'
- Quote: 'Works on a standard CMOS line'

WORKFLOW:
1. User shows screenshot or describes current state
2. Claude evaluates against checklist
3. Claude gives ONE specific next action
4. User implements
5. Repeat until checklist complete

SUCCESS CRITERIA:
- Demo runs without errors
- All checklist items checked
- Self-explanatory (can run without presenter talking)
- Follows Demo 1's live slide pattern

When complete, output: DEMO [N] LIVE SLIDE COMPLETE
