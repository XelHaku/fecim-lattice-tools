/ralph-loop:/ralph-loop

---

## Purpose

Keep iterating on Demos 2, 3, 4 to match Demo 1's "Live Slide" standard.

---

## The Command

When you type `/ralph-loop`, I will:

```
1. Ask: "Which demo? (2/3/4)"
2. Ask: "Screenshot or code?"
3. Evaluate against the Live Slide checklist
4. Give specific next action
5. Repeat until done
```

---

## Live Slide Checklist (Per Demo)

```
┌─────────────────────────────────────────────────────────┐
│  LIVE SLIDE CHECKLIST                                   │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  □ 1. Visual Anchor (hero element)                      │
│  □ 2. Main Plot/Visualization                           │
│  □ 3. 30-Level Indicator                                │
│  □ 4. Controls (material, mode, speed)                  │
│  □ 5. Educational Panel ("What You're Seeing")          │
│  □ 6. Operation Log (live actions)                      │
│  □ 7. Mode Indicator (WRITE/READ/COMPUTE)               │
│  □ 8. Key Stats ("Why This Matters")                    │
│  □ 9. Dr. Tour Quote                                    │
│  □ 10. Status Bar                                       │
│                                                         │
│  Demo Modes:                                            │
│  □ Auto Mode (self-running)                             │
│  □ Demo Mode (step-by-step)                             │
│  □ Manual Mode (full control)                           │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

---

## Demo-Specific Requirements

### Demo 2: Crossbar MVM

```
Visual Anchor:    Crossbar grid with animated current flow
Main Plot:        Heat map (conductance / IR drop / sneak)
Educational:      "1. Voltages applied → 2. Current flows → 3. Output = MVM"
Key Stat:         "16 multiplications in 1 cycle"
Quote:            "Compute in memory where the same device does memory and computation"
```

### Demo 3: MNIST

```
Visual Anchor:    Drawing canvas + big prediction number
Main Plot:        Layer activations flowing
Educational:      "1. 784 pixels → 2. 100K multiplications → 3. Prediction"
Key Stat:         "87% accuracy (88% theoretical max)"
Quote:            "We're at 87% validation here"
```

### Demo 4: Peripheral Circuits

```
Visual Anchor:    Signal flow diagram (DAC → Cell → ADC)
Main Plot:        Timing diagram + linearity plot
Educational:      "1. Digital → 2. Voltage → 3. Program → 4. Read back"
Key Stat:         "All standard CMOS components"
Quote:            "Works on a standard CMOS line"
```

---

## Usage

```
You:     /ralph-loop
Claude:  Which demo? (2/3/4)
You:     2
Claude:  Screenshot or describe current state?
You:     [uploads screenshot]
Claude:  ✅ Has: [list]
         ❌ Missing: [list]
         → Next action: [specific task]
You:     [does it]
You:     /ralph-loop
...repeat...
```

---

## Ready?

Type `/ralph-loop` to start.

🔥