# Zero to ASIC: Silicon Design with Skywater-PDK

**Speaker:** Matt Venn (with Mohammed Kassem)

---

## Introduction

[Music] Thanks for the introduction. So yeah, I'm super happy to be here. It's quite a funny experience being part of a remote conference, but I'm really glad that they've done this. I want to say thanks to the Hackaday team for extending out the kind of 20-minute demo slot into something that I can fit this into because it's a bit of a huge subject.

I'm really a learning-by-doing kind of person. So when I decided I wanted to learn a bit more about ASIC stuff, I just kind of jumped into doing it—doing the VLSI courses, doing Magic stuff. And then there was this big announcement by Google and Skywater for this open source PDK.

One of the things I'm going to really try hard to do is to not confuse everyone with loads of jargon. So this really is a Zero to ASIC course. So let's start at the top. ASIC is a Custom Integrated Circuit. So it's lots of tiny little switches packed into a silicon chip and then put onto a tiny little chip and then you put that on your board.

The aims of this kind of one-hour demo presentation are to give you a high-level understanding of what it takes to make a custom digital chip, why you might want to do that, see some of the tools in action for simple design, familiarity with the jargon to kind of take away this slightly unapproachable aspect of ASIC stuff, and where to find help.

## Historical Context

When I was getting ready with all of this stuff, it was really blowing my mind what a crazy field this is. It's full of huge and tiny numbers. We've got tiny feature sizes in the order of nanometers. So we've got meters; a thousand times smaller is millimeters; a thousand times smaller is micrometers or microns; a thousand times smaller than that is nanometers. And nanometers is the scale that we're operating on here.

We've got huge setup costs, which is one of the things that's held it out of the approachability for makers, for example, or the Hackaday community. The tiny switches used inside modern chips—MOSFETs—they are the most manufactured thing that has ever been made: 1.3 times 10 to the 22 according to Wikipedia, which is totally mind-blowing.

And yeah, it's magical in a way. Computers think using etchings and poison sand and measure time using vibrating crystals. So if you're looking for magic, you found it. I like that quote.

## What is a Chip?

Digital computers need switches. It used to be mechanics, and then it was relays, and then it was valves, and then it was transistors. The first transistor was 1954.

An integrated circuit is when you put multiple switches together or multiple parts of a component on one bit of silicon. It's a way of making things smaller. We see over the years how things just got more and more exponentially complex, and now we've got millions or billions of transistors or switches on a chip.

We've got a chip here and the top's been etched off. You can see the gold bond wires that break out this little chip of silicon in the middle with all the patterns that actually do the magical work. One interesting thing here is that this chip is so tiny that it has to be put in this big block of plastic and then bonded with these bond wires for us to actually get those signals out to the real world.

Now we're switching from the optical microscope to the scanning electron microscope. We start to see some of the internal structure here. These thicker lines are going to be power distribution, and then we start to see all the fine interconnect wires that connect everything together. Underneath those tiny wires, we've got some other structures with some other wires—these slightly thicker ones that are more power distribution—and then underneath that, we've actually got the switches, the MOSFETs.

## The MOSFET

A MOSFET is a type of transistor. So these small switches that we need to build up the chip, we use a MOSFET these days, which stands for Metal Oxide Silicon Field Effect Transistor.

If you've used MOSFETs in your own projects maybe for driving motors or H-bridges, they kind of look like big packages. But we have a tiny little chip of silicon, and on the cross-section, it looks a bit like this capacitor formed between the gate and the body. The oxide layer (silicon dioxide) is non-conductive. Then we have a metal layer and a semiconductor layer. When you put an electric field across here, it changes the electrochemical properties of this channel region and allows a current to flow from the source to the drain.

Why MOSFETs? They're easy to manufacture in volume. We've got this super cool way of making hundreds and millions of them all in one go. It's easy to change the size, so you can make a big one for power switching or tiny ones for calculation/digital computers.

## CMOS and Inverters

Let's have a look at how we could use a MOSFET to build one of the most basic structures of digital electronics: an inverter.

If the gate of the MOSFET is zero (not turned on), the pull-up resistor pulls the output high. If we turn on the input, it shorts to ground, and the output becomes low. One problem with NMOS (using a resistor) is efficiency because when it's pulled down, the resistor uses power. So we put an N and a P channel together (CMOS - Complementary MOS). When the P is on, it's pulled up; when the N is on, it's pulled down. This efficiency is the big reason for CMOS.

## Making a Chip (Lithography)

How do you actually make a chip? Fundamentally, you draw squares that represent where metal, silicon dioxide isolation, or doping (N/P type) needs to be. It used to be done by hand—cutting red tape. But you can do it on anything; you could do it in Inkscape.

You end up with stuff in a **GDSII** file format, and that's what you send to the fab (factory). "Taping out" comes from when they took the tape reels out of the computers and carried them to the factory.

A quick walkthrough of the process:
1. Get doped silicon ready.
2. Put photoresist on top, bake it.
3. Put a mask on and expose with UV light.
4. Develop/wash away the holes.
5. Etch/cut away or bombard with dopants to change chemical structure, or grow silicon dioxide/polysilicon/metal.

You end up with these amazing 3D structures.

## Demo: Designing with Magic

The tool I'm going to use is **Magic**. It's a very old tool but still used a lot today in the open source flow.

If I wanted to make an N-channel MOSFET, I draw where the N-type dopants are going to be implanted, then draw the gate (polysilicon). Where they overlap forms an N-type transistor. I connect them with local interconnects.

Magic has a **DRC** (Design Rule Check). It tells you if anything breaks a rule (e.g., specific to the Skywater 130 PDK).

I can also ask Magic to `extract`. It looks at how everything is connected and extracts a digital netlist. Then I can do `x2spice` (with parasitic capacitance modeling) to create a SPICE file and `make sim` to simulate it. I can see the input/output waveforms and measure switching time (e.g., 1 nanosecond).

## The Skywater PDK

The Skywater 130nm process has been made available on an open source license. Previously you needed an restricted access. Now we can use this **PDK (Process Design Kit)** which contains:
- DRC rules
- Behavioral models (SPICE)
- **Standard Cells**

I would never normally want to draw an inverter by hand; I would use a Standard Cell. Using **KLayout**, we can see the standard cells (e.g., high density library). There are about 150 of them. They fit together like Legos with power at the top and ground at the bottom.

## OpenLane Flow

OpenLane describes the flow from high-level design to GDSII.

1. **Synthesis (Yosys)**: Takes Verilog (HDL) and turns it into specific cells (AND/OR gates, flip-flops).
2. **OpenROAD**:
    - Floorplanning: How big the chip is, where pins go.
    - Placement: putting cells in optimal places.
    - Clock Tree Synthesis (CTS): Special handling for the clock signal.
    - Routing: Connecting everything.
    - Antenna insertion: Protecting gates from manufacturing static.
3. **Signoff**: DRC, LVS (Layout vs Schematic), Parasitic Extraction.

### Demo: Digital Counter

I'll demonstrate simulating a 7-segment second counter. We have a 16MHz clock. We count to 16 million to get one second, then increment a digit counter (0-9), then decode that to 7-segment LEDs.

In Verilog:
- `reg [23:0] second_counter;` (24 bits to count to 16M)
- `always @(posedge clk)` block for sequential logic.
- Combinatorial logic for the 7-segment decoder.

Running `flow.tcl` in OpenLane creates the GDSII. We can look at the intermediate steps:
- **Floorplan**: Pins and chip size.
- **Placement**: Cells arranged but messy.
- **Detailed Placement**: Everything lined up.
- **Routing**: Wiring it all together.

## Simulation and Verification

We are making physical hardware. We can't change it after it's made, and there's a 2-3 month wait. So we need to be sure.
- **Simulation**: Uses tools like **CocoTB** to simulate waveforms and check logic.
- **Formal Verification**: Mathematically proving the design works.

## Google Skywater Shuttle

Every few months, a factory takes a group order called a **Multi-Project Wafer (MPW)**. Usually expensive ($10k-$15k). Google is sponsoring a **free shuttle** for open source designs.

- Your design goes into a "user project area" on the Caravel harness.
- The harness provides a RISC-V processor, RAM, IO, Logic Analyzer to support your design.
- It's packaged in a wafer-level chip-scale package (cheap and high performance).

**Resources:**
- FOSSi Foundation Dial-ups
- Slack channel
- Kunal's VLSI Udemy course
- Tiny Tapeout / Zero to ASIC course

---

## Q&A with Mohammed Kassem

**Q: Where are we at with the IOs?**
**Mohammed:** The IOs are finished but not released yet. They will be out very soon (hopefully tomorrow). If you use the shuttle harness, you don't need to worry about IOs as much, but for scratch designs, you will.

**Q: What analog blocks are available?**
**Mohammed:** Currently just the lowest level (transistors) and digital standard cells. There is a vacuum for analog blocks (amplifiers, PLLs, etc.). We encourage the community to build them.

**Q: Can I use Magic for parasitic extraction?**
**Mohammed:** Yes. There is also an effort for KLayout extraction rules.

**Q: Are there open source FPGAs?**
**Mohammed:** Yes, there are efforts. Some slots on the shuttle will be filled with eFPGAs built from standard cells.

**Q: Can standard cells be shapes other than rectangles?**
**Mohammed:** Generally yes, as long as you follow design rules, but usually they fit the grid.

**Q: Risk-V 64-bit?**
**Matt:** If you can fit 10 PicoRV32s, you can probably fit a bigger core.
**Mohammed:** It depends on the architecture, but we've synthesized up to 900k gates.

**Q: Timeline?**
**Mohammed:** 76 days for wafer out, plus packaging. Roughly 2.5 months.

[Music]