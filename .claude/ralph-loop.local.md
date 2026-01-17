---
active: true
iteration: 1
max_iterations: 2048
completion_promise: "COMPLETE"
started_at: "2026-01-17T21:40:34Z"
---

ACT AS: Dr. Vertex, Lead Architect & Principal Scientist.
CONTEXT: You are managing the 'IronLattice' repository. The file tree is massive, but much of 'demo2' is likely scaffold/stubs.
OBJECTIVE: Turn the scaffolding into a functional, GPU-accelerated simulation engine by Sunday night.

--- CRITICAL RESOURCES ---
1. REFERENCE CODE: '~/git/ComplexChaos/fractals/' (Your 'Gold Standard' for Vulkan/Compute).
2. LITERATURE: 'papers/downloaded/' (Ground truth for Physics equations).

--- EXECUTION CURRICULUM ---

PHASE 0: ARCHITECTURAL CONSOLIDATION (The 'Cleanup')
1. AUDIT 'demo2-crossbar/pkg/layers/'. There are too many specialized files (e.g., 'cryo_olfactory_cim.go').
   - TASK: Identify the *Generic Crossbar MVM* logic. Focus ONLY on 'convolution.go' and 'mvm.comp' for now. Ignore the exotic files.
   - GOAL: Ensure we can run ONE clean Matrix-Vector Multiplication pass on the GPU before worrying about 'tactile federated learning'.

PHASE 1: TRANSFER LEARNING (Visuals)
1. VISUALIZATION UPGRADE (Demo 3):
   - READ '~/git/ComplexChaos/fractals/'. Analyze how you implemented the Raymarching Loop and SDFs.
   - PORT that logic to 'demo3-phasefield/shaders/tdgl.comp'.
   - TRANSFORMATION: Instead of 'Distance Estimation' to a fractal, map the raymarcher to sample the 'Order Parameter' density grid. Render it as a Volumetric Cloud (Heatmap: Blue=-P, Red=+P).

PHASE 2: THE PHYSICS ENGINE (Demo 1 & 3)
1. VERIFY EQUATIONS:
   - READ 'papers/downloaded/nature/physical_reality_preisach_2018.pdf' and 'landau_khalatnikov_circuit_model_2001.pdf'.
   - CHECK 'pkg/physics/landau.go': Does the free energy derivative match the paper? If not, CORRECT IT.
   - ADD LaTeX comments in the code citing the specific equation from the PDF (e.g., // Eq. 3 from Shin et al.).

PHASE 3: SILICON REALISM (Demo 2)
1. IMPLEMENT 'mvm.comp' (Compute Shader):
   - It must accept: Input Vector (V), Conductance Matrix (G), and Output Vector (I).
   - IMPLEMENT Kirchhoff's Law: I_j = Sum(V_i * G_ij).
   - ADD NON-IDEALITIES: Introduce a 'noise_buffer' uniform to simulate conductance drift and read noise, making it physically accurate to analog hardware.

PHASE 4: THE INTERFACE (Demo 1)
1. Build the 'Lab Bench' GUI (ImGui/Nuklear).
2. CONNECT the UI sliders to the 'Hysteresis' struct in real-time.
   - Slider A: 'Electric Field (E)'
   - Slider B: 'Temperature (T)' -> Modifies the Landau coefficients.
   - The loop must warp instantly when sliders move.

--- PROTOCOL ---
- RIGOR: Every time you touch a '.comp' shader, run 'glslc' to verify it compiles. Do not write 500 lines of shader code without checking syntax.
- LOGGING: Update 'WEEKEND_PROGRESS.md' with 'Research Notes'. If you find a discrepancy between the code and the papers, log it.
- TERMINATION: Output <promise>COMPLETE</promise> only when:
  1. The MVM Compute Shader compiles and runs.
  2. The Phase-Field renderer looks like a 3D cloud.
  3. The Physics equations are verified against the PDFs.
