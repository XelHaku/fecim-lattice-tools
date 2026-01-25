# Email Improvement Plan: Dr. Tour Outreach

## Context

**Original Request:** Improve the email to Dr. Tour about FeCIM Design Suite
**Output Location:** `docs/videos/email.md`
**Target Send Date:** Tuesday 9am Houston time

---

## Analysis: What's Working Well (KEEP)

### 1. Strategic Frame: "Help or Hurt" Dilemma
The honest framing of "would this help or hurt you?" is **excellent**. It:
- Shows respect for IronLattice's business interests
- Positions Juan as an ally, not a competitor
- Creates a genuine decision point (not just "look what I built")

### 2. Clear Call to Action
"Send me your GitHub username" is low-friction and specific. Keep it.

### 3. Personal Touch (P.S. about faith/science)
This is memorable and shows Juan watched the whole presentation. Keep the P.S.

### 4. Contact Information
Complete and professional. Keep as-is.

### 5. Subject Line
"FeCIM Design Suite - Open-Source EDA with OpenLane Integration" is descriptive and hits keywords Dr. Tour's team would recognize.

---

## Analysis: What Needs Improvement

### Issue 1: Opening is Weak - "inspired me to build something"
**Problem:** Passive, vague, doesn't hook immediately.
**Better:** Lead with the concrete value proposition in the first sentence.

### Issue 2: "15,000 lines of production-quality Go code" - Wrong Metric
**Problem:** Line count doesn't matter to a scientist/entrepreneur. Sounds like bragging.
**Better:** Describe what the tool DOES, not its size.

### Issue 3: Demo Link Placement - Buried After Intro
**Problem:** Busy professors skim. Demo should be visible immediately.
**Better:** Demo link in the first 3 lines, not after a paragraph of explanation.

### Issue 4: "WHAT IT GENERATES" Section - Too Technical
**Problem:** Lists file formats (DEF, LEF, Liberty) without explaining value.
**Better:** Translate to outcomes: "Design a chip in hours instead of weeks."

### Issue 5: Scenario A/B Section - Too Long
**Problem:** Forces Dr. Tour to read 8 lines analyzing his own business.
**Better:** Ask the question directly in 2 lines. He knows the implications.

### Issue 6: "I'm not looking for funding or a job" - Unnecessary Disclaimer
**Problem:** Mentioning it creates doubt. ("Why is he bringing this up?")
**Better:** Remove entirely. Actions speak louder.

### Issue 7: Two Repository URLs - Confusing
**Problem:** Current draft has both `github.com/your-org/fecim-lattice-tools` and mentions "repository is currently private" - slightly redundant.
**Better:** Single, clear URL with access instructions.

### Issue 8: Length - Slightly Too Long
**Problem:** ~350 words. Should be under 250 for a first cold email.
**Better:** Cut to essentials. Every sentence must earn its place.

---

## Concrete Changes

### Change 1: Rewrite Opening (Hook First)
**Before:**
```
Your COSM presentation on ferroelectric computing inspired me to build
something, and I need your guidance before deciding whether to release it.
```

**After:**
```
I built an open-source EDA suite for ferroelectric compute-in-memory
chip design. Before releasing it publicly, I wanted your input.
```
*Why:* Leads with the concrete deliverable, not with "inspiration."
*Note:* Using "an" instead of "the first" avoids unverifiable claims - the statement is compelling without needing superlative validation.

### Change 2: Move Demo Link to Top
**Before:** Demo link buried after second paragraph.

**After:** Demo link in paragraph 2, right after the hook.
```
5-MINUTE DEMO: [YouTube link]
```
*Why:* If he only reads 30 seconds, he sees the demo link.

### Change 3: Replace Line Count with Value Statement
**Before:**
```
It's about 15,000 lines of production-quality Go code across 6 modules.
```

**After:**
```
Six modules covering hysteresis physics, crossbar simulation, MNIST inference,
peripheral circuits, and chip design automation.
```
*Why:* Describes functionality, not vanity metrics.

### Change 4: Simplify "What It Generates" to Business Outcome
**Before:** Lists Verilog, DEF, Liberty, LEF, SPICE...

**After:**
```
WHAT IT DOES:
Design a FeCIM array, click export, get OpenLane-ready files.
Weeks of manual work reduced to minutes.
```
*Why:* Dr. Tour cares about outcomes, not file formats.

### Change 5: Collapse Scenario A/B to One Question
**Before:** 8 lines explaining two scenarios.

**After:**
```
MY QUESTION:
Would open-source FeCIM design tools help IronLattice build an ecosystem,
or would they give competitors too easy a path? I'd rather ask than guess.
```
*Why:* He understands the trade-off. Don't over-explain.

### Change 6: Remove "Not Looking for Funding" Line
**Delete entirely.** His first thought won't be "is this guy asking for money?" - so don't plant that thought.

### Change 7: Simplify Access Instructions
**Before:** Two paragraphs about repository and access.

**After:**
```
Repository is private. Reply with your GitHub username and I'll add you immediately.
github.com/your-org/fecim-lattice-tools
```
*Why:* One clear action.

### Change 8: Tighten P.S.
**Before:**
```
P.S. - Your integration of faith and science in that talk resonated beyond
just the technology. That perspective is rare and appreciated.
```

**After:**
```
P.S. The faith perspective in your COSM talk was memorable. Rare in this field.
```
*Why:* Same sentiment, fewer words. More punch.

---

## Final Improved Email

```
TO: tour@rice.edu

SUBJECT: Question Before Open-Sourcing FeCIM Design Tools


Dr. Tour,

I built an open-source EDA suite for ferroelectric compute-in-memory chip
design. Before releasing it publicly, I wanted your input.

5-MINUTE DEMO: [unlisted YouTube link]

WHAT IT DOES:
- Hysteresis physics (Preisach model, 30 discrete analog states)
- Crossbar simulation (MVM with IR drop, sneak paths, device variation)
- MNIST inference demo (matching your reported 87% hardware accuracy)
- Chip design automation: design an array, click export, get OpenLane-ready
  files (Verilog, DEF, SPICE)

MY QUESTION:
Would open-source FeCIM design tools help IronLattice build an ecosystem,
or give competitors too easy a path? I'd rather ask than guess.

PROPOSED NEXT STEP:
Reply with your GitHub username(s) and I'll grant private access immediately.
We can discuss whether to open-source, keep private, or collaborate.

github.com/your-org/fecim-lattice-tools

Best regards,

FeCIM Maintainers
Monterrey, Mexico
juan@trebuchetdynamics.com
+52 812 193 7470

P.S. The faith perspective in your COSM talk was memorable. Rare in this field.
```

### CC Strategy Decision

**Removed CC to jaeho-shin@rice.edu.** Rationale:
- Cold emails should go directly to the decision-maker
- CCing a colleague without established relationship can appear presumptuous
- If Dr. Tour wants to loop in collaborators, he will
- Avoids any perception of "going around" someone

### Subject Line Change

**Before:** "FeCIM Design Suite - Open-Source EDA with OpenLane Integration"
**After:** "Question Before Open-Sourcing FeCIM Design Tools"

*Why:* The new subject line:
- Signals a decision is being requested (not just "look what I built")
- Creates curiosity and mild urgency
- Shorter and more direct
- Matches the email's actual purpose (asking for input)

### Closing Line Change

**Removed:** "No pressure to respond. But I built this specifically for FeCIM and wanted to check with you first."

*Why:* This line:
- Weakens the ask ("no pressure" = permission to ignore)
- Is redundant (the email already explains why Juan is reaching out)
- The P.S. provides a softer close without undermining the ask

---

## Word Count Comparison

| Version | Words |
|---------|-------|
| Original | ~350 |
| Improved | ~180 |

**Reduction: 49%** - Much more scannable for a busy professor.

---

## Key Improvements Summary

| Element | Before | After |
|---------|--------|-------|
| Hook | Vague "inspired me" | Concrete "I built an..." |
| Claim | "the first" (unverifiable) | "an" (still compelling) |
| Subject line | Descriptive | Decision-focused question |
| CC | jaeho-shin@rice.edu | Removed (direct to decision-maker) |
| Demo link | Buried | Line 4 |
| Technical proof | Line count (15k) | Feature list |
| Business question | 8 lines, two scenarios | 2 lines, direct question |
| Disclaimer | "Not looking for funding" | Removed |
| Closing | "No pressure to respond" | Removed (P.S. softens naturally) |
| Access instructions | Two paragraphs | Two lines |
| Total length | 350 words | ~180 words |

---

## Task List

- [ ] **TODO-1:** Save final email to `docs/videos/email.md`
- [ ] **TODO-2:** Verify YouTube demo link placeholder is clear
- [ ] **TODO-3:** Confirm send timing (Tuesday 9am Houston = 8am CST)

---

## Success Criteria

1. Email is under 250 words (currently ~180)
2. Demo link appears in first 5 lines
3. Single clear call-to-action (GitHub username)
4. "Help or hurt" question preserved but tightened
5. P.S. preserved but shortened
6. No mention of funding/jobs
7. No unverifiable claims ("the first" replaced with "an")
8. Subject line communicates the decision being requested
9. No CC that could appear presumptuous
10. No weak closing that gives permission to ignore

---

## Revision History

### Iteration 1 (Critic Feedback)

**MUST FIX (Addressed):**
- Changed "the first open-source EDA suite" to "an open-source EDA suite" - avoids credibility risk if other tools exist

**SHOULD CONSIDER (Addressed):**
1. **CC strategy:** Removed CC to jaeho-shin@rice.edu - direct to decision-maker is cleaner
2. **Closing line:** Removed "No pressure to respond" - weakened the ask unnecessarily
3. **Subject line:** Changed to "Question Before Open-Sourcing FeCIM Design Tools" - signals decision requested

---

PLAN_READY: .omc/plans/email-improvement.md
