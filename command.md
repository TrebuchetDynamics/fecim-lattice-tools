/ralph-loop:ralph-loop "ACT AS: Dr. Vertex, Lead Architect & Principal Scientist.
CONTEXT: You are building 'IronLattice-vis' - an interactive GPU-accelerated visualization of ferroelectric compute-in-memory technology for Dr. external research group's IronLattice startup. Deadline: 2 weeks.

OBJECTIVE: Build 3 working demos that visualize ferroelectric CIM technology.

--- THE VISION ---
We are visualizing the technology that will replace GPUs for AI computation.
- 30 discrete polarization levels (not just 0/1)
- Square hysteresis loops (IronLattice's key innovation)
- Compute-in-memory (no Von Neumann bottleneck)
- Target: 87% MNIST accuracy (matching Dr. Tour's results)

--- CRITICAL RESOURCES ---
1. REFERENCE CODE: '~/git/ComplexChaos/fractals/' (Vulkan/Compute gold standard)
2. LITERATURE: 'papers/downloaded/' (Physics equations)
3. EXISTING TOOLS: CrossSim, Preisachmodel, ferro_scripts (reference implementations)

--- BUILD ORDER ---

PHASE 1: DEMO 1 - SINGLE FERROELECTRIC CELL (Days 1-3)
Goal: User controls electric field, sees polarization response with 30 discrete levels.

1. PHYSICS ENGINE:
   - IMPLEMENT Preisach model in 'demo1-hysteresis/pkg/ferroelectric/preisach.go'
     * Array of hysterons with (alpha, beta) thresholds
     * P = Σ hysteron states
   - IMPLEMENT Landau-Khalatnikov in 'demo1-hysteresis/pkg/ferroelectric/landau.go'
     * Free energy: F(P) = αP² + βP⁴ - E·P
     * Dynamics: dP/dt = -γ · dF/dP
   - VERIFY equations against papers. Add LaTeX comments citing sources.

2. 30 DISCRETE LEVELS:
   - Map polarization range [-Pr, +Pr] to 30 levels
   - Visualize as color gradient (level 1 = blue, level 30 = red)
   - Show current level number on screen

3. VISUALIZATION:
   - Real-time P-E hysteresis curve (X = Electric Field, Y = Polarization)
   - Animated dot tracing the curve as E changes
   - Show square loop characteristic (IronLattice's advantage)

4. INTERACTIVITY:
   - Slider: Electric Field (E) from -Emax to +Emax
   - Slider: Sweep speed
   - Button: Auto-sweep (trace full loop)
   - Display: Current P value, current level (1-30)

PHASE 2: DEMO 2 - CROSSBAR ARRAY (Days 4-6)
Goal: Visualize matrix-vector multiplication happening in memory.

1. CROSSBAR STRUCTURE:
   - IMPLEMENT 'demo2-crossbar/pkg/crossbar/array.go'
     * N×M grid of ferroelectric cells
     * Each cell holds conductance G derived from polarization level
     * G = G_min + (level/30) × (G_max - G_min)

2. MVM COMPUTE SHADER:
   - IMPLEMENT 'demo2-crossbar/shaders/mvm.comp'
     * Input: Voltage vector V[N]
     * Weights: Conductance matrix G[N×M] (30 levels each)
     * Output: Current vector I[M]
     * Kirchhoff's Law: I_j = Σ(V_i × G_ij)
   - ADD non-idealities: conductance noise, IR drop (optional)

3. VISUALIZATION:
   - Grid of cells, color = level (1-30)
   - Animate voltage flowing in from left (horizontal lines)
   - Animate current flowing down columns (vertical lines)
   - Brightness = signal magnitude
   - Show input vector, weight matrix, output vector

4. INTERACTIVITY:
   - Set input voltage vector manually
   - Click cell to change its level (1-30)
   - Watch MVM result update in real-time

PHASE 3: DEMO 3 - MNIST NEURAL NETWORK (Days 7-10)
Goal: Draw digit, watch computation flow through crossbars, see prediction.

1. NETWORK STRUCTURE:
   - IMPLEMENT 'demo3-mnist/pkg/network/network.go'
     * Layer 1: 784×128 crossbar (input → hidden)
     * ReLU activation
     * Layer 2: 128×10 crossbar (hidden → output)
     * Softmax → prediction

2. PRETRAINED WEIGHTS:
   - LOAD weights from file (quantized to 30 levels)
   - OR train simple network, quantize weights to 30 levels
   - Target: 87% accuracy on MNIST test set

3. VISUALIZATION:
   - Drawing canvas: 28×28 pixels
   - Layer 1 crossbar: show 784 inputs flowing in, 128 outputs
   - Layer 2 crossbar: show 128 inputs flowing in, 10 outputs
   - Output neurons: 10 bars showing activation
   - Prediction: highlight winning digit

4. INTERACTIVITY:
   - Mouse draw digit on canvas
   - 'Classify' button runs inference
   - Watch computation animate through both layers
   - Show prediction with confidence percentage

--- TECH STACK ---
- Language: Go
- Graphics: Vulkan
- Shaders: GLSL (SPIR-V) - compile with glslc
- UI: ImGui or Nuklear for sliders/buttons

--- PROTOCOL ---
1. RIGOR: Run 'glslc' after every shader edit. No 500-line shaders without compile check.
2. MODULARITY: Each demo should run independently.
3. PHYSICS FIRST: Get the math right before making it pretty.
4. LOGGING: Update 'PROGRESS.md' daily with completed tasks.
5. TEST: Each demo must have at least one working test case.

--- SUCCESS CRITERIA ---
DEMO 1 COMPLETE when:
- [ ] P-E curve animates in real-time
- [ ] 30 levels visually distinct
- [ ] Preisach OR Landau model working
- [ ] Interactive sliders control E

DEMO 2 COMPLETE when:
- [ ] MVM compute shader compiles and runs
- [ ] Voltage/current flow animated
- [ ] Can manually set cell levels
- [ ] Output vector updates in real-time

DEMO 3 COMPLETE when:
- [ ] Can draw digit with mouse
- [ ] Inference runs through both crossbars
- [ ] Prediction displays correctly
- [ ] Achieves >80% accuracy on test digits

--- TERMINATION ---
Output <promise>COMPLETE</promise> only when:
1. All 3 demos run independently
2. Video recording shows full workflow
3. README documents how to run each demo
4. Code is clean enough to share with Dr. external research group" --max-iterations 2048 --completion-promise "COMPLETE"