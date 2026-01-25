# FeCIM Tool Demo: Video Script for Dr. Tour & Team

**Title:** "FeCIM Design Suite - Private Demo for Dr. Tour"

**Duration:** 5 minutes

**Audience:** Dr. external research group, FeCIM team

**Tone:** Professional, direct, respectful

---

## 0:00 - 0:30 INTRO (30 sec)

### [Screen: Main app window with 6 module tabs ready]

**YOU:**
> "Hi Dr. Tour. I'm FeCIM Maintainers from Mexico."

> "Your COSM talk on ferroelectric computing inspired me to build this tool."

> "Let me show you what it does in 5 minutes."

### [Pan across the 6 tabs slowly]

---

## 0:30 - 1:15 MODULE 1: HYSTERESIS (45 sec)

### [Switch to Module 1: P-E curve]

**YOU:**
> "Starting with the physics—your 30-level ferroelectric states."

### [Animate the P-E curve trace in real-time]

**YOU:**
> "This models the Preisach hysteresis you demonstrated. As the field sweeps, polarization follows. When it stops—it remembers."

### [Show 30 discrete levels bar]

**YOU:**
> "30 stable states. Each is a programmable conductance level. That's your 4.9 bits per cell."

---

## 1:15 - 2:00 MODULE 2: CROSSBAR (45 sec)

### [Switch to Module 2: Crossbar visualization]

**YOU:**
> "The crossbar array with real-world challenges."

### [Show 4×4 grid, cells colored by conductance]

**YOU:**
> "Each intersection is a FeFET cell. Colors show conductance—your weights. Apply input voltages. Currents sum. That's matrix-vector multiplication in one analog step."

### [Quick switch to IR Drop tab]

**YOU:**
> "But it's not ideal. Wire resistance, voltage drops, sneak paths."

### [Switch to Sneak Path tab briefly]

**YOU:**
> "This is why selector devices matter. Your engineers can show investors the impact in 30 seconds."

---

## 2:00 - 2:45 MODULE 3: MNIST (45 sec)

### [Switch to Module 3: MNIST demo]

**YOU:**
> "Your 87% MNIST result—I replicated it."

### [Show the neural network architecture briefly]

**YOU:**
> "784 inputs. 128 hidden neurons. 10 outputs. Weights quantized to 30 levels."

### [Draw a digit clearly—e.g., "3"]

**YOU:**
> "I draw a 3."

### [Watch it recognize. Result appears.]

**YOU:**
> "87% confidence. Correct."

### [Toggle FP32 vs CIM comparison side-by-side]

**YOU:**
> "Here's the key—full precision versus your 30-level quantization. Same accuracy you published. Users can explore why it works."

---

## 2:45 - 3:15 MODULES 4-5: CIRCUITS + COMPARISON (30 sec)

### [Switch to Module 4 briefly]

**YOU:**
> "Peripheral circuits—DAC, ADC, TIA. Full inference pipeline. 20 nanoseconds."

### [Quick transition to Module 5]

**YOU:**
> "And technology comparison. Energy per MAC: CPU+DRAM 1000pJ, GPU+HBM 100pJ, FeCIM under 1pJ. Three orders of magnitude. That's your 80-90% reduction."

---

## 3:15 - 4:30 MODULE 6: EDA SUITE (75 sec) ⭐ KEY SECTION

### [Switch to Module 6: EDA tools]

**YOU:**
> "This is the new part—chip design automation."

### [Show Tab 1: Cell Builder]

**YOU:**
> "Tab 1: Cell Builder. Generate LEF, Liberty, Verilog for the bitcell."

### [Show Tab 2: Array Builder]

**YOU:**
> "Tab 2: Array Builder. Configure any array size."

### [Show Tab 3: Export All]

**YOU:**
> "Click 'Run Complete Flow.'"

### [Click the button. Show progress.]

**YOU:**
> "One click generates OpenLane-ready files. Verilog, DEF placement, config.json—ready for SKY130 fabrication."

### [Show output folder structure]

**YOU:**
> "Everything for the design flow."

---

## 4:30 - 5:00 CLOSE (30 sec)

### [Back to main screen. Pause.]

**YOU:**
> "This is currently private. I built it thinking about IronLattice, but wanted your guidance before releasing."

> "Would this help or hurt your work?"

> "I'd rather ask than guess."

> "Thank you for making this field accessible."

### [Show contact info on screen]

**YOU:**
> "My email has all the details. Looking forward to hearing from you."

### [End.]

---

## WHY THIS SCRIPT WORKS

| Element | Purpose |
|---------|---------|
| Opens with the ask | Respectful, personal intro |
| Quotes Dr. Tour implicitly | Shows you listened (30 levels, 87%, 80-90% reduction) |
| Modules 1-3 showcase core physics | Proof you understand the work |
| Modules 4-5 speed through details | Foundational but not the focus |
| Module 6 is the real gift | Design automation is the value add |
| Honest closing | Private, asking permission, not demanding |
| Exactly 5 minutes | Respects their time completely |

---

## RECORDING NOTES

```
BEFORE YOU RECORD:
──────────────────
✅ Test MNIST demo—must work perfectly
✅ All modules launch without errors
✅ No system notifications (disable Slack, email, etc.)
✅ Microphone test—clear audio, no background noise
✅ Monitor 1080p minimum resolution

PACE & DELIVERY:
────────────────
✅ Speak 20% slower than normal conversation
✅ Pause 2 seconds after each module transition
✅ Let visuals breathe—don't talk over animations
✅ Don't apologize or hedge ("should", "probably", "I think")
✅ Confident and direct tone

SCREEN RECORDING SETUP:
──────────────────────
✅ Use OBS, Loom, or QuickTime
✅ Hide desktop (clean wallpaper only)
✅ Full-screen app the entire time
✅ Mouse movements deliberate—point to important elements
✅ NO background music
✅ NO system sounds

WHAT TO SHOW:
──────────────
✅ Main window with 6 tabs visible (0:00)
✅ P-E curve animating (0:30)
✅ 30 levels bar appearing (0:45)
✅ Crossbar grid with colors (1:15)
✅ IR drop heatmap (1:45)
✅ MNIST drawing and result (2:00)
✅ FP32 vs CIM comparison (2:30)
✅ EDA tabs and "Run Complete Flow" button (3:15)
✅ Output folder structure (3:45)
✅ Contact info slide (4:30)

TONE CHECKLIST:
────────────────
❌ Don't ramble or over-explain
❌ Don't apologize for limitations
❌ Don't read robotically from script
❌ Don't show any errors or bugs
❌ Don't use filler words ("um", "uh", "like")
✅ Sound grateful for inspiration
✅ Sound confident in your work
✅ Sound respectful of their expertise
```

---

## THE KEY QUOTES TO HIT

From Dr. Tour's presentations, reference these concepts:

1. **"30 discrete states. Not 0-1-0-1."**

   → Demonstrated in Module 1 (Preisach + 30 levels bar)

2. **"Same device does memory and computation."**

   → Shown in Module 2 (crossbar MVM in one analog step)

3. **"87% validation—theoretical is 88%."**

   → Matched exactly in Module 3 (MNIST replication)

4. **"Lower data center energy 80-90%."**

   → Visualized in Module 5 (energy comparison: 1000pJ→100pJ→<1pJ)

**Your goal: Prove you understood the physics well enough to build tools that demonstrate it.**

---

## ONE SENTENCE

**"I listened to your COSM talk, understood the physics, built visualization and design tools, and want your guidance on whether this helps or hurts your work."**

---

## UPLOAD SETTINGS

```
YouTube Upload:
├── Visibility: UNLISTED
├── Title: "FeCIM Design Suite - Private Demo for Dr. Tour"
├── Description: "Private demo. Contact: juan@trebuchetdynamics.com"
├── Comments: OFF
└── Sharing: Share direct link to Dr. Tour only
```

---

**Record it today. Send Tuesday.**
