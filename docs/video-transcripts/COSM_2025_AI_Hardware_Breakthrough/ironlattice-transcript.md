# Ferroelectric CIM Presentation Transcript
## Dr. external research group - First Public Presentation

**Source:** Jesus and Science YouTube Channel  
**Date:** November 2024 (references Nov 3rd WSJ article)  
**Context:** First public disclosure of Ferroelectric CIM technology

> **Note:** This is an archival transcript. Statements are quoted from the source and are **not** verified or endorsed by this project.

---

## Opening Remarks

So what I'm going to introduce today is something that this is the first time I'm speaking about it publicly. It's a new design for AI computation and it addresses one of the biggest challenges that we've been hearing about which is the energy requirements and to really lower the energy. Not all the details are going to be in here but enough for you to get the concept. The patents still have not gone public. So we're still not revealing the details of this.

This is an effort that's been going on in my group for about a year and a half. The story behind it is when Nvidia really just burst on the scene about a year and a half ago, I brought in one of the guys in my group who works in device architectures and I said, "Look, we've got to get in on this. We've got to do something in AI." But we're not a software group. We've got to do it differently. I want you to design a new hardware system. 

And he said, "That's very hard." 

I said, "That's exactly why I'm asking you to do this because we do it as President Kennedy said, we're doing it because it is hard." 

And after 6 months he came up with something.

---

## Core Technology: Compute-in-Memory

This is ultra low power compute in memory. So what happens now is when you do computation you actually move information from the logic to the memory back and forth. You do a logic and then you move it back and you bus it back and forth. And this takes a lot of energy to do this. 

And this is a compute in memory system where the same device does the memory and the computation and it's based on a **ferroelectric super lattice** and it's **capital light**. There's no exotic materials in here. There's no graphene in here, no exotic materials. It works on a standard CMOS line and can translate just like that.

---

## Market Context

So if we look at the exponential growth in AI that we've been talking about and you see what's happening here and companies are investing billions of dollars and I just heard one recent company is $150 billion committed to their data centers. The thing is that they're often not scalable. They're not CMOS compatible and they're not cost effective. So we had to address this.

---

## Core Innovation

And so the idea is that we would enable this and we'd come up with a core innovation. It's a **CMOS ready ferroelectric layer**. Works on a normal processing line, normal fabrication line. It's a proprietary super lattice. It solves the historical reliability challenges. There's these extensive write cycles and data retention even at extreme operating temperatures. We've demonstrated it's the ultimate building block and we'd truly like to do this compute computation in memory all in one device.

---

## Market Introduction Strategy

Our way of thinking about how we would introduce this to the market is to just look at it initially as a memory system. So if you look at this:

### vs. NAND Flash
- **10 million times lower** read write energy than NAND flash
- **Million times faster** than NAND flash  
- **90% voltage reduction** than NAND flash

### vs. DRAM
- **Nonvolatile** - zero refresh cycles
- **Thousand times lower** read write energy

### Phased Introduction

So what we would think of doing is first switching out—if this were a data center system—we would initially switch out the NAND systems. We could start making money that way without having to rewrite software without having to do anything much here and then we would start switching out the DRAM again moving into that market and then do the full compute in memory. That's what we have thinking about how we would do this and introduce it.

You could do the same thing even in a smartphone—again just initially switching out the NAND, the flash memory, and then moving into the DRAM system and then moving into the compute. Because if there are companies that are going directly to try to do the compute and memory, compute and memory, but it's hard to get this introduced into the market that way. So that's the plan of doing it this way.

And so if you look at the growth of NAND flash, you know, it's a healthy growth on NAND flash, healthy growth on DRAM, and of course then the AI semiconductor market. So huge areas of growth here that we would be able to capitalize upon that we would try to do.

---

## De-Risking Strategy

We want to derisk this again—this pathway to just have this drop in CMOS compatible proven materials standard equipment. There's no unusual tools that we even need for this. We're building this straight on our university system with all the standard tools. 

**We are Technology Readiness Level TRL4** which is just a component validation in a lab environment. We have not yet moved this into a foundry. We're talking with several of the classic foundries to be able to do this. 

Our strategy is to be **capital light**—would be a **fabless business model**, tier one R&D partners and we'd secure US production to follow.

---

## Technical Performance Data

### Endurance
This is some of the endurance. I've intentionally kept off here the scale, but you can see that this is a logarithmic scale here. And these are the cycle numbers, the endurance, and the retention.

### Electrical Performance
And we look at the electrical performance, the stability. So these are the polarization curves that we've got here and how we've been able to do this computation and trying to extrapolate this out. **We still have to get this up to the required 10^12 cycles** and eventually hopefully even higher than that.

### Compute-in-Memory Demonstration
We've done the compute in memory. We've put this on the **MNIST system**. So we can read these handwritten numbers. **We're at 87% validation here**. 

It's got **30 discrete states**. So we have all these intermediate states. So it's not 0-1-0-1. And we have 30 discrete states that we can access in this post synaptic current. We can do it current in time, current and pulse. And theoretical is 88% is the theoretical maximum. We're at 87% being able to read these handwritten numbers.

---

## Competitive Positioning

And if you look at where we are on this, this is Ferroelectric CIM. I think we really can address well:
- The read/write energy
- The read/write speed  
- The endurance cycles
- The operating voltage density
- CMOS scalability

And if we compare it to other things that are out there I think we're doing very well.

### Weebit Nano - Previous Success
Actually this company Weebit—this is another memory that came out of my lab. We started this company in 2015 out of my lab and it's selling now on the market. It's got three big customers and this is based on a ReRAM system. 

I told them at the time early on I said we've got to start looking at this for neuromorphic computation and they said no no neuromorphic is not really important and it turns out to be the most important thing right now. But anyway I think that we've got good standards here.

### Market Disruption Potential
And then if we look at the DRAM market disruption—again we're doing quite well on the DRAM disruption. And I think that we can handle that quite well. 

And then again this AI compute market disruptions—if we look at Ferroelectric CIM and where we compare to Google and Intel and Mythic AI, **we're actually in discussions with one of these companies right now** in all restricted access discussions and **with a second large company in the country** as well.

---

## Response to Industry Trends: Wafer-Scale Computing

So, George [Gilder] just came out with this article in the Wall Street Journal just on November 3rd. He said, "The microchip era is about to end." So, how could I come here and not at least address this? Because he says, you know, the microchips are coming to an end. So, could we ride with this? 

So, we actually studied quite in depth what he was looking at. So, to quote him in this article, he says the microchip era is about to end. And he talks about these **three fundamental stress factors**:
1. The current chip size limits
2. The lithographical constraints on the building of these
3. The memory bottleneck  
4. The packaging complexity

He went on and he says—but I think this is how our ferroelectric memory and this compute in memory device will **support full wafer integrated circuits** while simultaneously solving these core problems. So again we think we can unlock these and our bottom line here is after this analysis of looking at what George has said is:

**There may be significant hurdles to achieve true wafer scale integrated circuits. However if the industry does move in that direction this technology vector can be enhanced and possibly even enabled by our ferroelectric memory and CIM (compute-in-memory) devices.**

---

## Team

This is the group here that we've got going. Every time we ask somebody who's deeply involved in the industry to look at this and tell us what they think they want to come on as an adviser. And so that's what we've got.

### Core Team
- **J Shin** - Device engineer in my lab that came up with this idea
- **Tophik Jarur** - One of my former students. He's been with Accenture for 13 years working with the largest manufacturers in the world, many of them. And so he's going to take this on.

### Advisers
- **Fred** - Been like 35 or 40 years. He was with IBM and he's really helped us to boost up our patent portfolio. He was a key adviser on the patent side at IBM. So he's gone through all our patents and then we filed more patents to boost it up.
- **John** - Has been a big investor in computing systems
- **Two gentlemen** - Deeply involved in the semiconductor industry and foundry experts

---

## Funding Status

**We haven't raised a penny to date. We've taken no money** because we are just trying to see what would be the best strategy for going forward. It's not that we're opposed to taking money. We just haven't done anything with it because we really want to move with the best strategy.

---

## Summary

This sort of summarizes where we are on this advance and how we're trying to bring this forward. Again, it's this—there's no more busing information back and forth to memory. **It's all done in the same device. Computation memory in the same device** and **this could lower the requirements in a data center by 80 to 90% of the energy requirements**. 

So you think about the amount of energy that goes into these energy centers and we hear again and again how it's going to suck up all of our electricity. Well, if we stick with the same devices, that may be the case. But there could be a way to change out devices and to change technology behind the devices and then address the energy problem in that way by changing out the devices.

With that, I will open it up for questions if there are any.

---

## PRESENTATION SLIDES ANALYSIS

### Slide 1: Title & Overview

**Ferroelectric CIM enables ultra-low-power (ULP) AI computing**

Key Points:
- ⚡ Ultra low power AI through compute in memory
- 🔬 Ferroelectric superlattice device unites storage and compute
- 💰 Capital light, fab friendly path using standard materials and techniques

**Branding:**
- Ferroelectric CIM logo features a lattice/grid pattern design
- Prepared for: CSM (Christian Stewardship Ministries)
- Source: jesusandscience.org

---

### Slide 2: Problem Statement

**Situation: Exponential growth of AI and related energy consumption**

**AI Compute Performance Requirements:**
- Doubling every 2 months now
- Growth trajectory from 2012 → 2024 showing exponential curve
- Key milestones labeled: AlexNet, GPT-2, GPT-3, GPT-4

**Market Failures:**
> "Companies have invested $Bs in solutions and have fallen short"

Three critical failures:
1. **Not Scalable**
2. **Not CMOS Compatible**
3. **Not Cost Effective**

*Source: http://ai.science.org/doi/10.1126/science.aeh0600*

---

### Slide 3: Solution Architecture

**Our Solution: How we will enable a breakthrough in Computing**
> "We've engineered a single device that acts as both ultra fast, non-volatile memory and an intelligent logic gate."

#### The Ferroelectric CIM Device Architecture

**Two Device Structures Shown:**

1. **FTJ Structure (2-terminal):**
   - Ferroelectric layer
   - **Proprietary** superlattice at core
   - Compact 2-terminal design

2. **Fe FET Structure (3-terminal):**
   - Ferroelectric layer
   - **Proprietary** superlattice
   - 3-terminal transistor configuration
   - More complex but higher functionality

#### Core Innovations

**Core Innovation #1: The CMOS-Ready Ferroelectric Layer using standard materials**

**Core Innovation #2: Proprietary Superlattice**
- Solves historical reliability challenges and stabilizes the ferroelectric materials
- Extensive write cycles and data retention even at extreme operating temperatures

**The Ultimate Building Block for Next-Generation AI & Data Centers:**
- ✅ Unified Memory that is fast, dense, and persistent
- ✅ True Compute-in-Memory

*Source: Ferroelectric CIM*

---

### Slide 4: Data Center Use Case (Dell PowerEdge)

**Solution in Practice: Data Center Transformation (Dell PowerEdge)**

> "Hyperscale, super-efficient data centers can replace typical, power-hungry server implementations. Below is just one of millions of servers."

#### PowerEdge XE7740
Actual AI Ultra Training and Computing Server

**Current Configuration:**
- **NAND:** Present in storage (122 TB in this model)
- **DRAM:** DRAM (4TB) and GPUs
- **Processor:** 2 Intel Xeon Processors

#### Ferroelectric CIM Benefits

**Increase Performance and Efficiency (NAND)**
- 10,000,000 × lower read / write energy
- 1,000,000× faster than NAND flash (10ns switching)
- 90% voltage reduction (2-3 V vs. 10-20 V)

**Standby power reduction (DRAM)**
- Zero refresh cycles
- 1,000 × lower read / write energy

**Future AI-Compute Replacement**
- Merging compute and storage leads to a large reduction in system power

> "Ferroelectric CIM will enable lower power consumption"

*Source: Dell Computing user sheet [08]*

---

### Slide 5: Mobile Use Case (iPhone 16 Pro)

**Solution in Practice: Revolutionizing Edge AI (Apple iPhone 16 Pro)**

> "Ferroelectric CIM can fundamentally extend device battery life and enable higher AI compute workloads on the edge"

#### iPhone 16 Pro with Edge AI Capabilities

**Internal board layout showing:**
- Main A18 Pro chip
- Memory and storage components highlighted

**Current Configuration:**
- **NAND:** Present in storage
- **DRAM:** Apple A18 Pro processor combines the functions of the DRAM, Processor, and GPU but they remain 3 distinct units
- **Processor:** (Integrated in A18)

#### Ferroelectric CIM Benefits (Same as Data Center)

**Storage power savings (NAND)**
- 10,000,000 × lower read / write energy
- 1,000,000× faster than NAND flash (10ns switching)
- 90% voltage reduction (2-3 V vs. 10-20 V)

**Standby power reduction (DRAM)**
- Zero refresh cycles
- 1,000 × lower read / write energy

**Future AI-Compute Replacement**
- Merging compute and storage leads to a large reduction in system power

> "Ferroelectric CIM will enable lower power consumption"

*Source: iFixit [03]*

---

### Slide 6: Market Landscape

**Market Landscape: $350B+ disruption opportunity today**

> "Our target markets are projected to expand from a combined $153B in 2025 to over $711B by 2030"

#### Three Target Markets with Growth Projections (2025-2030):

**1. NAND Flash Market Growth (2025-2030)**
- 2025: 78.3B → 2030: 98.1B
- High Estimate ($B USD): 87.7B → 110.5B
- Steady growth trajectory
- *Sources: Analysis based on Q2 2025 NAND Market reports and World Semiconductor Trade Statistics (WSTS)*
- *Note: Projections model a 4.5% CAGR from 2025*

**2. DRAM Market Growth (2025-2030)**
- 2025: 142.7B → 2030: 219.5B  
- High Estimate ($B USD): 174.5B → 268.1B
- Strong growth driven by AI/ML workloads
- *Sources: Consensus estimates from IHS 2025 reports by Fortune Business Insights and Verified Market Research*
- *Note: Projections model a 9.0% CAGR (High)/8.5% (Mid) from 2025*

**3. AI Semiconductor Market Growth (2025-2030)**
- 2025: 163.1B → 2030: 402.6B
- High Estimate ($B USD): 223.3B → 551.1B
- Explosive growth reflecting AI boom
- *Sources: Forecasts derived from IHS 2025 AI sector reports and VisiontlR's Internal modeling (July 2025) as based on the industry-agnostic trajectory through 2032*

**Total Combined Market:** $153B (2025) → $711B (2030)

---

### Slide 7: De-Risking Production

**De-Risking Production: A Low-Risk Path to Mass Production**

> "Our manufacturing strategy is to not expend infrastructure and to operate fabless business like Nvidia"

#### The Technology: A 'Drop-In' Solution

**Key Advantages:**
- ✅ **CMOS Compatible**
- ✅ **Proven Materials**
- ✅ **Standard Equipment:** Leverages existing ALD & PVD tools

**Current Status:**
- Our technology lab is at Technology Readiness Level 4 (TRL 4) which is component validation in lab environment

#### The Strategy: A Capital-Light Model

**Business Approach:**
- 💼 **'Fabless' Business Model**
- 🤝 **Tier-1 R&D Partners**
- 🇺🇸 **Secure U.S. Production to follow**

*Image: Modern semiconductor fabrication facility (cleanroom)*

---

### Slide 8: Electrical Performance - Switching Speed

**Ferroelectric CIM Electrical performance (Switching speed)**

#### Endurance Testing

**Graph showing:**
- Ferroelectric polarization (P in a.u.) vs Cycle number
- Logarithmic scale: 10⁰ to 10² cycles
- Clear cyan/blue states showing dual polarization levels
- Stable switching over extended cycles

#### Retention Testing

**Graph showing:**
- Polarization retention over time (0-1000 seconds)
- Two distinct states (positive/negative) maintained
- Minimal degradation over 1000s duration

**Pulse condition: 100, 10, 1 μs, 100, and 10 ns.**

✓ Ferroelectric memory device for endurance and retention characteristics.
  (Wake-up → Stable operation → Fatigue)

*Confidential | Copyright ©*

---

### Slide 9: Electrical Performance - Stability

**Ferroelectric CIM Electrical performance (Stability)**

#### Polarization vs. Electric Field (P-E) Curves

**Left Graph:**
- Shows classic ferroelectric hysteresis behavior
- Multiple curves for different retention times (10⁴, 10⁵, 10⁶, 10⁷ s)
- Demonstrates stable ferroelectric behavior over time

**Key Finding:**
> "Polarization vs. Electric Field (PE) curves substantiate ferroelectric behavior and cyclability (wake-up → stable operation)"

> "Stable operation for 10⁴ s (retention) and 10⁹ cycles (endurance)."

> "Fully CMOS-compatible fabrication procedure without exotic materials."

#### Retention Characteristics

**Bottom Left Graph:**
- ON-OFF ratio > 10⁵
- Shows clear separation between ON and OFF states
- Retention demonstrated over 10 years

**Bottom Right Graph - Endurance:**
- ON-OFF ratio > 10⁵ maintained
- Tested across 10⁰ to 10⁹ cycles
- Consistent performance throughout

**Comparison:**
> "Compared to NAND flash and DRAM, it has non-volatile characteristics, fast operation speed, and no refresh process needed."

---

### Slide 10: Compute-in-Memory Applications

**Ferroelectric CIM Compute-in-memory applications**

#### Post-Synaptic Current Demonstrations

**1. Post-synaptic current/time**
- Shows potentiation behavior (0.1 ms pulse)
- PSC (a.u.) increasing over time (0-160 ms)

**2. Post-synaptic current/pulse width**
- Variable pulse widths: 0.1 ms, 0.01 ms
- PSC response varies with pulse duration
- Time scale: 0-20 ms

**3. 30 Intermediate States**
- Graph showing Vg = 0.1 V
- 30 distinct, stable conductance levels
- PSC (State) from 0-60
- Demonstrates long-term potentiation and depression (LTP and LTD)

#### MNIST Handwritten Digit Recognition

**Schematic Diagram:**
- Shows neural network architecture
- Input pattern → Weight matrix → Output neurons
- Includes visual of handwritten "5"
- Bar graphs showing classification output

**Results Graph:**
- Accuracy (%) vs Epoch (#)
- Red line: Experiment
- Black dashed: Ideal case
- Achieves ~60% accuracy

**Performance:**
> "MNIST pattern (hand-written digit) recognition simulation."

> "Estimated pattern recognition accuracy is ~87% (88% theoretical maximum)."

---

### Slide 11: Memory Market Competitive Analysis

**Memory Market Disruption: Outperforming NAND and Competition**

> "NAND Flash has been optimized for density and non-volatility but suffers from significant limitations in speed, power consumption, and endurance. Emerging technologies like Resistive RAM (ReRAM) and Phase-Change Memory (PCRAM) attempt to address this."

#### Competitive Comparison Matrix

| Feature | **Ferroelectric CIM** | **3D NAND Flash** | **ReRAM (e.g., Weebit)** | **PCRAM (e.g., IBM)** |
|---------|----------------|-------------------|--------------------------|----------------------|
| **Write/Read Energy** | ✅ | ❌ | ✅ | 🟡 |
| **Write/Read Speed** | ✅ | ❌ | ✅ | ❌ |
| **Endurance (Cycles)** | ✅ | ❌ | ❌ | 🟡 |
| **Operating Voltage** | ✅ | ❌ | ✅ | ❌ |
| **Density (Multi-Level)** | ✅ | ✅ | ❌ | ✅ |
| **CMOS Scalability** | ✅ | ✅ | 🟡 | 🟡 |

**Legend:**
- ✅ = Technology has these characteristics (green checkmark)
- 🟡 = Technology is working towards this capability (yellow checkmark)
- ❌ = Does not have or struggles with this feature

**Ferroelectric CIM Advantage:** Only technology with green checkmarks across ALL categories

---

### Slide 12: AI Compute Market Competitive Analysis

**AI Compute Market Disruption: Enabling True Ultra-Low-Power AI**

> "Today's AI chips force a trade-off between power-hungry designs and impractical exotic architectures. Ferroelectric CIM offers the high-path around exotic, impractical exotic architectures, without compromise on efficiency"

#### Competitive Comparison Matrix

| Feature | **Ferroelectric CIM** | **Google TPU v6e** | **Intel Loihi 2** | **Mythic AI** |
|---------|----------------|-------------------|------------------|--------------|
| **Architecture** | ✅ | ✅ | ✅ | ✅ |
| **Memory / Logic Integration** | ✅ | ❌ | 🟡 | 🟡 |
| **Power Efficiency** | ✅ | ❌ | ✅ | ✅ |
| **CMOS Compatibility** | ✅ | ✅ | 🟡 | ❌ |
| **Scalability & Precision** | ✅ | ✅ | ❌ | ❌ |

**Legend:**
- ✅ = Technology has these characteristics (green checkmark)
- 🟡 = Technology is working towards this capability (yellow checkmark)
- ❌ = Limitation or does not have this feature

**Key Differentiators:**
- **Memory/Logic Integration:** Ferroelectric CIM is the only solution with TRUE compute-in-memory (green)
- **CMOS Compatibility:** Only Ferroelectric CIM and Google TPU are fully CMOS compatible
- **All-around Excellence:** Ferroelectric CIM is the only technology with green across ALL categories

---

### Slide 13: George Gilder WSJ Article

**Ferroelectric CIM - Industry Context**

**Wall Street Journal Opinion Article:**

### **The Microchip Era Is About to End**

> "The future is in wafers. Data centers will be the size of a box, not vast energy-hogging structures."

**By George Gilder**  
**Nov 3, 2025 1:34 pm ET**

**Significance:** This influential Wall Street Journal opinion piece by technology futurist George Gilder frames the broader industry context for Ferroelectric CIM's technology, arguing that the traditional microchip paradigm is reaching its limits and wafer-scale integration is the future.

---

### Slide 14: Wafer-Scale Computing Analysis (Part 1)

**Ferroelectric CIM - Full Wafer processing – "*The End of the Microchip*"**

**In the Wall Street Journal article by George Gilder, "The Microchip Era Is About to End," several technological and practical factors are identified, including three fundamental drivers for wafer scale circuitry:**

#### 1. Current Chip Size Limits – The Reticle Constraint → Wafer scale imaging

**Challenges:**
- Lithography can pattern only up to a fixed maximum chip size.
- Ultra-fast AI workloads can't run on a single chip. Requires thousands to millions of GPUs/chiplets connected in parallel. With heat loads that push power component density.
- Current complexity reduces efficiency and increases cost, heat, and latency.
- Emerging Lithography capabilities may enable larger up to wafer sized circuitry

#### 2. Memory Bottleneck – Wafer-Scale Integration Improves Efficiency

**Benefits:**
- Larger, denser chips shorten the distance between compute and memory.
- Less data movement means far lower energy loss.
- Higher locality boosts the efficiency of AI training and inference.

#### 3. Packaging Complexity & Energy Consumption – Improved by Wafer-Scale Integration

**Advantages:**
- Fewer chips reduce multi-chip packaging complexity, heat, and design overhead.
- Long-distance signaling wastes power and lowers efficiency.
- Wafer-scale integration simplifies communication and reduces energy loss.

---

### Slide 15: Wafer-Scale Computing Analysis (Part 2)

**Ferroelectric CIM - Full Wafer processing – "*The End of the Microchip*"**

**How our ferroelectric memory & CIM devices support full wafer ICs --- while simultaneously solving the core problems that limit today's chips**

#### Unlocking Large ICs and Dense Placements

**Integrating Memory and Computation – Turning today's small, discrete units into one unified system**

- **CIM architectures perform MAC operations inside the memory cells.**
- **Minimizes inter-chip communication and avoids massive interconnect networks.**
- **While Starting as Fully CMOS-compatible, manufacturable in existing foundries (e.g., TSMC, Samsung).**

#### Lower Operating Heat Outputs Dramatically Enable Denser Implementations

**Thermal Advantages:**
- **Lower heat output reduces cooling requirements and system power.** Enabling larger or denser circuits.
- **Components can be placed much closer together without thermal throttling.**
- **Higher density improves performance, lowers energy per operation, and reduces cost.**

---

### Slide 16: Wafer-Scale Computing Analysis (Part 3)

**Ferroelectric CIM - Full Wafer processing – "*The End of the Microchip*"**

**How our ferroelectric memory & CIM devices support full wafer ICs --- while simultaneously solving the core problems that limit today's chips**

#### While Solving Today's Core Problems

**Combining Nonvolatility and High Speed**
- **Provides nonvolatile data storage with nanosecond-scale switching.**
- **Delivers DRAM-level speed with NAND-like persistence.**
- **Fewer data transfers → major reductions in power consumption for AI data centers.**

**Simplifying the Memory Hierarchy**
- **Today's hierarchy: L1/L2 → DRAM → HBM → SSD → HDD adds latency and energy overhead.**
- **Ferroelectric memories (FeFET, FTJ) can sit directly next to compute units.**
- **Enables in-place computation without frequent data movement.**

#### Bottom Line Conclusion

**Yellow Highlighted Text:**
> "Bottom Line: There may be significant hurdles to achieving true wafer-scale integrated circuits. However, if the industry moves in that direction, that technology vector can be considerably enhanced—and possibly even enabled—by our ferroelectric memory and CIM devices."

---

### Slide 17: Team Overview

**Meet the Team**

> "We brought together the top semiconductor, academic, and business minds to get Ferroelectric CIM launched"

#### Core Team (Left Side)

**Dr. external research group**
- Founder & Chief Scientist
- T.T. & W.F. Chao Endowed Chair in Chemistry
- Materials Science & Nanoengineering
- external research institution

**Dr. Jaeho Shin**
- Co-Founder & CTO
- Ferroelectric Superlattice R&D
- external research institution

**Tawfik Jarjour**
- CEO, Semi Industry, Data Centers & AI Hardware
- Business leadership

#### Advisors (Right Side)

**Dr. Fred Flitsch**
- Advisor, Foundry & Process Expertise
- Former IBM, extensive semiconductor experience

**John Jaggers**
- Advisor, Growth and Strategy, VC
- Managing Partner
- Investment and strategic guidance

**Dr. Yoontak Hwuang**
- Advisor, Semiconductor Commercialization
- Industry expertise

**Sandeep Davé**
- Foundry Expert
- Semiconductor manufacturing expertise

**Team Composition:**
- 3 Core team members (Founder, Co-Founder/CTO, CEO)
- 4 Strategic advisors covering foundry, investment, commercialization, and manufacturing
- Strong mix of academic excellence (external research institution) and industry experience (IBM, semiconductors, VC)

---

## Q&A Session

### Question 1: Electromagnetic Interference and Manufacturing Scale

**Q:** Are you subject to electromagnetic interference on any of this? And what is the scale going down 2 nanometer boundaries on manufacturing—what is the scale that your process uses?

**A:** Right. So our process right now is being done strictly in the university. So it's not very small scale. So, this is what we're talking about with the foundries—that we would go, we would start working our way down with the commercial foundries. 

As far as electromagnetic interference or some electrical pulse or something, we have no idea. I just don't know. I don't want to give a premature answer on this. That's certainly an important consideration, but I just don't have an answer for that. 

And again, **we're just at TRL 4 and remember commercialization is 9. So we have a lot of steps we still need to go.**

### Question 2: Applicability and National Security

**Q:** Um Dr. this is really exciting. I just have one question to kind of clarify this somewhat. I was wondering if you think that your innovations and this technology can be applicable to all the data centers that are being built now, including like maybe Elon Musk's Colossus 1 and 2—is this perhaps applicable to everyone and can you keep this confined to the United States or Western countries and keep it from being stolen by China?

**A:** Well, it is applicable to all of them. Yes. And like I said, we'd like to have this stage introduction—just start replacing our flash memory, which flash memory, you know, burst on the scene in about 2001, something like that. So, it's been around for about 25 years. So, it's about time. And which is really amazing technology—you mean that you could have electronic memory that's nonvolatile with no moving parts and just by putting in a deep trench capacitor which was really really amazing when it came out. 

But and then we would begin to switch out DRAM and phase this in that way so that it'd be an easy transition. That's the hope. But yes, **it would apply to all the data centers. This would work for that.**

As far as **keeping it in the United States**, I mean that's what we want to do. We certainly want to keep this in the United States, but you know, we are in a university and I have a group of all sorts of people. I mean, the FBI has come to me and said, you know, we're concerned—do you have any students that might be spies? And I'm like, I don't know. You're the guy—you're supposed to tell me. 

So, they came back again. I said, fear not. You don't have to worry. I asked all my guys. I said, are you spies? They all said no. So, we're good to go. 

I mean, I don't know what more I could do. If they're a spy, you stop them at the border. I mean, that's your job, not my job. So, you know, we do the best we can, but I don't do any classified work in my laboratory. Because if you really do classified work, you can't do it in a university. You have to have a special facility. The nearest facility to me is NASA. And if you do classified work, every time I even need to use the restroom, you have to take out the hard drive, put it in a locked cabinet, go to the restroom, come back and put the thing back in. And this just doesn't work at my age. I mean, it's just not the way to go. 

And so, we don't do classified work. And the way we succeed is we share information between all of us. We sit in group meetings for hours and they want everybody to contribute to this. So I just remember how we beat the Soviet Union. We beat the Soviet Union not by keeping things absolutely hardened and concealed. We beat the Soviet Union by sharing information between us. They kept things secret and we beat the pants off of them and it was all based around silicon technology and that's because we had an open system of sharing and communicating and that's the way we succeed.

---

## Key Technical Claims Summary

| Metric | Comparison | Performance |
|--------|-----------|-------------|
| Energy vs NAND | Read/Write Energy | 10,000,000× lower |
| Speed vs NAND | Operation Speed | 1,000,000× faster |
| Voltage vs NAND | Operating Voltage | 90% reduction |
| Energy vs DRAM | Read/Write Energy | 1,000× lower |
| Refresh vs DRAM | Refresh Cycles | Zero (nonvolatile) |
| States | Analog Levels | 30 discrete states |
| MNIST Accuracy | Handwriting Recognition | 87% (vs 88% theoretical max) |
| Data Center Energy | Overall Reduction | 80-90% reduction potential |
| Endurance Target | Write Cycles | 10^12 cycles (in progress) |
| Current Status | Technology Readiness | TRL 4 (lab validation) |
| Manufacturing | CMOS Compatibility | Standard fab equipment |

---

## COMPREHENSIVE KEY POINTS SUMMARY

### 🔬 Core Technology

**Technology Name:** Ferroelectric CIM  
**Architecture:** Compute-in-Memory (CIM)  
**Core Innovation:** Ferroelectric super lattice that performs both memory storage AND computation in the same device

**Key Principle:**
- Eliminates the "von Neumann bottleneck" (moving data between separate memory and processor)
- Memory cells perform matrix-vector multiplication using physics (Ohm's Law)
- Computation happens where data is stored—no data movement required

---

### 🧪 Materials & Manufacturing

**Primary Material:** Ferroelectric super lattice (proprietary composition)
- **NOT exotic materials** - No graphene or specialized compounds
- **CMOS compatible** - Works on standard semiconductor fabrication lines
- **Standard equipment** - No unusual tools required
- **University fabrication** - Currently built at external research institution with standard tools

**Manufacturing Advantages:**
- Drop-in replacement for existing fab processes
- Capital light approach
- Scalable with current infrastructure
- De-risked manufacturing pathway

---

### ⚡ Performance Metrics

#### vs. NAND Flash Memory
- **Energy:** 10,000,000× (10 million times) lower read/write energy
- **Speed:** 1,000,000× (1 million times) faster
- **Voltage:** 90% voltage reduction
- **Non-volatile:** Data retained without power

#### vs. DRAM
- **Energy:** 1,000× (thousand times) lower read/write energy  
- **Refresh:** Zero refresh cycles (non-volatile vs DRAM's constant refresh)
- **Retention:** Maintains data without power

#### Data Center Impact
- **80-90% reduction** in total energy requirements
- Addresses AI/ML energy crisis
- Applicable to all data centers (confirmed: Tesla Colossus, etc.)

---

### 🎯 Memory Capabilities

**30 Discrete Analog States**
- Not binary (0/1) but 30 distinct polarization levels
- Stores ~5 bits per cell (vs 1 bit for binary)
- Direct analog weight storage for neural networks
- Post-synaptic current modulation (current in time, current in pulse)

**Endurance & Reliability:**
- **Target:** 10^12 (1 trillion) write cycles
- **Current Status:** Demonstrated extensive write cycles (exact number not disclosed)
- **Data Retention:** Maintained even at extreme operating temperatures
- Solves historical ferroelectric reliability challenges

---

### 🧠 AI/ML Demonstration

**MNIST Handwritten Digit Recognition:**
- **87% accuracy** on physical hardware
- **88% theoretical maximum** for this task
- Demonstrates real compute-in-memory functionality
- Not a simulation—actual ferroelectric synaptic device

**Capability:**
- Matrix-vector multiplication in analog domain
- Parallel computation across entire crossbar array
- Single-step operation vs millions of sequential CPU operations

---

### 👥 Team & Leadership

#### Core Team
- **Dr. external research group** - Principal Investigator, external research institution
- **J Shin** - Device engineer who developed the core technology (6 months of focused R&D)
- **Tophik Jarur** - Business lead, former student, 13 years at Accenture working with major manufacturers

#### Advisory Board
- **Fred** - 35-40 years experience, former IBM patent expert, bolstered Ferroelectric CIM patent portfolio
- **John** - Major investor in computing systems
- **Two unnamed advisers** - Semiconductor industry and foundry experts

**Recognition:** "Every time we ask somebody who's deeply involved in the industry to look at this and tell us what they think, they want to come on as an adviser."

---

### 📜 Intellectual Property

**Patent Status:**
- **Multiple patents filed** (exact number not disclosed)
- **NOT yet public** - Still pending publication
- Strengthened by IBM patent expert (Fred)
- Filed additional patents after initial review
- Details intentionally withheld until patents publish

**Protection Strategy:**
- Proprietary super lattice composition
- Process patents for fabrication
- Device architecture patents
- Securing US-based IP

---

### 🚀 Development Timeline

**Project History:**
- **Start Date:** ~18 months before presentation (circa early 2023)
- **Catalyst:** Nvidia's breakthrough emergence in AI
- **Development Time:** 6 months from concept to prototype
- **Motivation:** "We do it because it is hard" (JFK quote)
- **Approach:** Hardware-focused (not software), doing AI differently

**Current Status:**
- **TRL 4:** Technology Readiness Level 4 - Component validation in lab environment
- **Path to TRL 9:** Commercialization requires 5 more readiness levels
- **Academic Setting:** external research institution lab fabrication
- **Next Steps:** Transition to commercial foundry

---

### 💼 Business Strategy

#### Market Introduction - Phased Approach

**Phase 1: NAND Flash Replacement**
- Drop-in replacement for existing flash memory
- Immediate revenue potential
- No software changes required
- Easy market entry
- Target: Data centers and smartphones

**Phase 2: DRAM Replacement**  
- Replace volatile memory with non-volatile alternative
- Eliminate refresh power consumption
- Expand market penetration

**Phase 3: Full Compute-in-Memory**
- Deploy complete CIM architecture
- Maximum energy savings (80-90%)
- Full AI acceleration capabilities

**Rationale:** "Companies going directly to compute-in-memory find it hard to get introduced into the market that way."

#### Business Model
- **Fabless design** - No fab ownership, license/partner model
- **Capital light** - Minimal infrastructure investment
- **Tier 1 R&D partners** - Collaborate with major semiconductor companies
- **US production focus** - Secure domestic manufacturing

#### Funding Status
- **$0 raised to date** - Intentionally bootstrapped
- **Not opposed to funding** - Waiting for optimal strategy
- **Strategic approach** - Prioritizing best path forward over quick capital

---

### 🏢 Market Opportunities

**Target Markets:**
1. **NAND Flash Market** - Healthy growth trajectory, 25-year-old technology ripe for disruption
2. **DRAM Market** - Large established market with energy inefficiencies
3. **AI Semiconductor Market** - Explosive growth, massive energy demands

**Market Context:**
- Companies investing $150+ billion in data centers
- Existing solutions "not scalable, not CMOS compatible, not cost effective"
- Flash memory dominant since ~2001 - "it's about time" for replacement

---

### 🤝 Industry Engagement

**Active Discussions:**
- **Major Company #1:** Under restricted access (Google/Intel/Mythic AI mentioned as comparisons)
- **Major Company #2:** Large US company, restricted access
- **Multiple foundries:** "Several classic foundries" in discussions

**Previous Success - Weebit Nano:**
- Memory company spun out of Tour lab in 2015
- Based on ReRAM technology
- Currently selling on market with 3 major customers
- Validates Tour lab's commercialization track record
- Lesson learned: Should have pursued neuromorphic computing earlier

---

### 🌐 Industry Trends & Positioning

**Wafer-Scale Computing (George Gilder, WSJ Nov 3, 2024):**
- Article: "The Microchip Era is About to End"
- Three fundamental stress factors:
  1. Chip size limits
  2. Lithographical constraints  
  3. Memory bottleneck
  4. Packaging complexity

**Ferroelectric CIM Response:**
- Ferroelectric CIM can support full wafer-scale integration
- Solves core problems Gilder identified
- "If industry moves to wafer scale, this technology vector can be enhanced and possibly enabled by our devices"

**Competitive Position:**
- Benchmarked against Google, Intel, Mythic AI
- "We're doing very well" in comparisons
- Unique: CMOS compatibility + 30 states + proven in hardware

---

### 🛡️ National Security & IP Protection

**US Focus:**
- Goal: Keep technology in US/Western countries
- Reality: University research environment with international students
- FBI inquiries about potential espionage
- Tour's philosophy: "We beat the Soviet Union by sharing information, not by keeping secrets"

**Approach:**
- No classified work in lab (requires special facilities)
- Open collaboration within research group
- Patent protection as primary IP defense
- Pragmatic about university research constraints

**Quote:** "We beat the Soviet Union not by keeping things absolutely hardened and concealed. We beat the Soviet Union by sharing information between us."

---

### ⚠️ Technical Challenges & Open Questions

**Acknowledged Challenges:**
1. **Endurance:** Need to reach 10^12 cycles (currently in progress)
2. **Electromagnetic Interference:** "We have no idea" - not yet tested
3. **Manufacturing Scale:** Currently university-scale, need foundry transition
4. **Node Size:** Not at commercial nanometer scales yet
5. **TRL Progress:** TRL 4 → TRL 9 requires significant work

**Unknowns:**
- Specific fab node compatibility (2nm, 3nm, etc.)
- Production yield rates
- Cost per device at scale
- Full system integration requirements

---

### 🎯 Value Proposition Summary

**Why Ferroelectric CIM Matters:**

1. **Energy Crisis Solution:** AI is consuming unsustainable amounts of electricity; 80-90% reduction addresses existential constraint
2. **CMOS Compatible:** Works with existing $trillions of fab infrastructure
3. **Multiple Revenue Streams:** Can enter market incrementally (NAND → DRAM → CIM)
4. **Proven Hardware:** 87% MNIST accuracy demonstrates real-world functionality
5. **Capital Efficient:** Fabless model minimizes investment requirements
6. **Track Record:** Team has successfully commercialized memory technology before (Weebit)

**The Grand Vision:**
> "There's no more busing information back and forth to memory. It's all done in the same device. Computation memory in the same device."

**Impact:**
> "This could lower the requirements in a data center by 80 to 90% of the energy requirements... There could be a way to change out devices and to change technology behind the devices and then address the energy problem in that way."

---

### 📊 Quick Reference: Critical Specs

| Category | Specification |
|----------|--------------|
| **Technology** | Ferroelectric super lattice compute-in-memory |
| **States** | 30 discrete analog levels |
| **Energy (vs NAND)** | 10,000,000× reduction |
| **Energy (vs DRAM)** | 1,000× reduction |
| **Speed (vs NAND)** | 1,000,000× faster |
| **Endurance Goal** | 10^12 write cycles |
| **MNIST Accuracy** | 87% (hardware tested) |
| **AI Energy Savings** | 80-90% for data centers |
| **Volatility** | Non-volatile (no refresh) |
| **CMOS Compatible** | Yes - standard fabs |
| **Exotic Materials** | No - proven semiconductors |
| **TRL** | 4 (lab validation) |
| **Patents** | Multiple filed, not yet public |
| **Funding Raised** | $0 (strategic choice) |
| **Team Origin** | external research institution, Dr. Tour lab |
| **Development Time** | 18 months (6 months to prototype) |
| **Market Entry** | Phase 1: NAND replacement |
| **Business Model** | Fabless, tier-1 partnerships |
