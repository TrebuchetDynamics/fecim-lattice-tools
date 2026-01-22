# FeCIM Tool Demo: Video Script for Dr. Tour & Team

**Title:** "FeCIM Visualization Suite - Private Demo"

**Duration:** 8 minutes

**Audience:** Dr. Tour, Dr. Jaeho Shin, Tawfik Jarjour

**Tone:** Professional, direct, no fluff

---

## OPENING (0:00 - 0:30)

### [Screen: Your MNIST demo ready. You draw a "7". It recognizes it.]

**YOU:**
> "Dr. Tour, Dr. Shin, Mr. Jarjour—"

> "I watched your COSM presentation. 30 discrete states. 87% MNIST. Compute-in-memory. I built a tool to visualize and demonstrate all of it."

### [Beat. Show the "7" result.]

**YOU:**
> "This is what I want to show you. Eight minutes."

---

## MODULE 1: HYSTERESIS (0:30 - 1:45)

### [Switch to Module 1: P-E curve]

**YOU:**
> "You said: 'It's not 0-1-0-1. It's got 30 discrete states.'"

> "This is the hysteresis visualization."

### [Trace the P-E curve in real-time]

**YOU:**
> "Preisach model. As I sweep the electric field, the polarization follows the hysteresis loop. When I stop—"

### [Stop. Polarization holds.]

**YOU:**
> "—it remembers. Non-volatile."

### [Show 30 levels bar]

**YOU:**
> "30 stable states. Each one is a programmable conductance level. That's your 4.9 bits per cell."

### [Show material selector briefly]

**YOU:**
> "I've parameterized for different ferroelectric behaviors. Your superlattice would need calibration with real data—but the framework is here."

---

## MODULE 2: CROSSBAR MVM (1:45 - 3:30)

### [Switch to Module 2: Crossbar visualization]

**YOU:**
> "You said: 'The same device does the memory and the computation.'"

> "This is that computation."

### [Show crossbar grid, cells colored by conductance]

**YOU:**
> "Each intersection is a FeFET cell. Color indicates conductance—blue is low, red is high. These are your weights."

### [Apply input voltages. Show currents summing.]

**YOU:**
> "I apply input voltages to the columns. Ohm's Law at each cell: current equals voltage times conductance. Kirchhoff's Law at each row: currents sum."

### [Show matrix equation appearing]

**YOU:**
> "That's matrix-vector multiplication. One analog step. No data movement."

### [Switch to IR Drop tab]

**YOU:**
> "But you know it's not ideal. Here's IR drop."

### [Show voltage gradient heatmap]

**YOU:**
> "Wire resistance causes voltage to drop across the array. Corner cells see different voltages than edge cells. Error accumulates."

### [Switch to Sneak Path tab]

**YOU:**
> "Sneak paths. Current flows backwards through unselected cells."

### [Show interference visualization]

**YOU:**
> "This is why selector devices matter. I can toggle different on/off ratios to show the impact."

### [Quick toggle: 10:1 vs 1000:1]

**YOU:**
> "Your engineers know this. Now they can show it to investors in 30 seconds."

---

## MODULE 3: MNIST - THE FLAGSHIP (3:30 - 5:15)

### [Switch to Module 3: MNIST demo]

**YOU:**
> "You reported 87% on MNIST with 88% theoretical max."

> "This matches that."

### [Show the neural network architecture]

**YOU:**
> "Two-layer network. 784 inputs—28×28 pixels. 128 hidden neurons. 10 outputs. The weights are quantized to 30 levels."

### [Draw a clear "3"]

**YOU:**
> "I draw a 3."

### [Watch it compute. Result appears.]

**YOU:**
> "87% confidence. Correct."

### [Draw "7", "5", "9" quickly. All correct.]

**YOU:**
> "7. 5. 9."

### [Toggle FP32 vs CIM comparison]

**YOU:**
> "Here's the key for investor demos. Side-by-side comparison. Full precision floating point versus your 30-level quantization."

### [Show where they differ]

**YOU:**
> "When they match—quantization isn't hurting you. When they differ—this shows why 87% instead of 99%. It's honest. It builds credibility."

### [Show failure mode preset]

**YOU:**
> "I added presets. 'Ideal.' 'Noisy.' 'Broken ADC.' Investors can see what degrades accuracy and why your 87% is impressive—not easy."

---

## MODULE 4: PERIPHERAL CIRCUITS (5:15 - 6:30)

### [Switch to Module 4]

**YOU:**
> "A crossbar alone isn't a chip. You need DACs, ADCs, TIAs. This shows the full system."

### [Show WRITE mode]

**YOU:**
> "Write mode. I select a cell, choose a target level—say 22 out of 30."

### [Show voltage calculation]

**YOU:**
> "The tool calculates programming voltage. 4.2 volts. 50 nanosecond pulse. The ferroelectric switches."

### [Show READ mode briefly]

**YOU:**
> "Read mode. Low voltage—stays below threshold. Doesn't disturb the cell. Current flows through TIA to ADC. Digital output."

### [Show COMPUTE mode]

**YOU:**
> "Compute mode. This is inference. Digital inputs convert through DACs. Voltages hit the crossbar. Currents sum. TIAs convert to voltage. ADCs digitize. Full pipeline. 20 nanoseconds."

### [Show timing diagram]

**YOU:**
> "Timing diagram. DAC settling: 5ns. Crossbar: 5ns. ADC: 10ns. Total: 20ns."

> "A GPU doing the same matrix multiply takes 500 nanoseconds minimum. Plus memory bandwidth."

---

## MODULE 5: COMPARISON (6:30 - 7:30)

### [Switch to Module 5]

**YOU:**
> "You said: 'This could lower data center energy by 80 to 90%.'"

> "This is how I visualize that claim."

### [Show energy comparison bar chart]

**YOU:**
> "Energy per MAC operation. CPU plus DRAM: 1000 picojoules. GPU plus HBM: 100 picojoules. Ferroelectric CIM: under 1 picojoule."

> "Three orders of magnitude. That's your 80-90%."

### [Show competitive matrix]

**YOU:**
> "Competitive comparison. NAND, ReRAM, PCM, MRAM. Only FeCIM has checkmarks across every category."

### [Point to each row]

**YOU:**
> "Energy. Speed. Endurance. Voltage. Density. CMOS compatibility."

> "This is your investor slide. Interactive. They can ask 'what about ReRAM?' and you show them."

### [Show market size briefly]

**YOU:**
> "Market context. NAND: 80 billion. DRAM: 140 billion. AI semiconductors: 160 billion growing to 400 billion by 2030."

> "Your phased entry strategy—NAND first, then DRAM, then full CIM—this visualizes that path."

---

## CLOSING (7:30 - 8:00)

### [Back to MNIST demo. Draw one more digit. It works.]

**YOU:**
> "Dr. Tour, you said you're at TRL 4. Lab validation. You're talking to foundries. You're preparing for investor discussions."

> "This tool exists to help with that."

### [Beat.]

**YOU:**
> "Five modules. Education. Investor demos. Foundry conversations. Engineer onboarding."

> "Module 6—design automation—is early. I know I might be naive about the complexities of real production flows."

> "But I'm working to make it an insightful tool for design exploration. And I'm ready to learn what you actually need."

### [Beat.]

**YOU:**
> "The models are based on published physics and your public presentation. They'd need calibration with your actual device data to be accurate."

> "I built the framework. You have the measurements."

### [Show contact info]

**YOU:**
> "Private repo. If you want access, send GitHub usernames. If you want something specific built, tell me what you need."

> "FeCIM Maintainers. Monterrey, Mexico."

### [End.]

---

## WHY THIS SCRIPT WORKS

| Element | Purpose |
|---------|---------|
| Opens with MNIST demo | Hook—shows it works immediately |
| Quotes Dr. Tour directly | Shows you listened, builds connection |
| Module by module | Organized, easy to follow |
| Acknowledges limitations | "Needs calibration"—honest, builds trust |
| Ties to their needs | Investors, foundries, engineers |
| Ends with clear offer | "Send usernames"—specific call to action |
| 8 minutes exactly | Respects their time |

---

## RECORDING NOTES

```
PACE:
─────
- Speak slower than normal
- Pause after each module transition
- Let the visuals breathe
- Don't rush the MNIST demo

VISUALS:
────────
- Clean desktop
- Full screen app
- Zoom in on important parts
- Mouse guides attention

TONE:
─────
- Professional, not salesy
- Confident, not arrogant
- Helpful, not desperate
- Direct, not rambling

TEST BEFORE RECORDING:
──────────────────────
- MNIST demo works perfectly
- All modules launch without errors
- No notifications pop up
- Audio is clean
```

---

## THE KEY QUOTES TO HIT

From Dr. Tour's talk, reference these:

1. > "It's got 30 discrete states. So it's not 0-1-0-1."
   
   → Your Module 1 shows this

2. > "The same device does the memory and the computation."
   
   → Your Module 2 shows this

3. > "We're at 87% validation here... theoretical is 88%."
   
   → Your Module 3 matches this

4. > "This could lower the requirements in a data center by 80 to 90%."
   
   → Your Module 5 visualizes this

**You're not explaining FeCIM to them. They invented it.**

**You're showing them YOU understood it well enough to build tools for it.**

---

## ONE SENTENCE

**"I watched your presentation, I understood the physics, I built the visualization—here's how it can help you close deals."**

---

**Record it. Send it. See what happens.** 🦁
