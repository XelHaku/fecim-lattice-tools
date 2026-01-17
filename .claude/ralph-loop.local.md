---
active: true
iteration: 1
max_iterations: 512
completion_promise: "COMPLETE"
started_at: "2026-01-17T21:11:50Z"
---

You are the Lead Engineer for IronLattice. Your goal is to advance Demos 1, 2, and 3 from 'Planned' to 'Functional'.

Adhere to this STRICT execution plan:

PHASE 1: RESEARCH & CONTEXT
1. Read the PDFs in 'papers/' and the markdown files in 'docs/' to understand the physics (Preisach, Hysteresis, CIM).
2. If you lack knowledge on specific Go-Vulkan bindings or compute shaders, search for documentation on 'github.com/bbredesen/go-vk' or standard Vulkan specs.

PHASE 2: DEMO 1 (HYSTERESIS GRAPHICS)
1. The physics is done. Now implement the Vulkan rendering pipeline in 'demo1-hysteresis/pkg/render'.
2. Create the vertex and fragment shaders to visualize the P-E loop.
3. Ensure 'go build ./demo1-hysteresis/cmd/hysteresis' compiles successfully.

PHASE 3: DEMO 2 (CROSSBAR MVM)
1. Implement the matrix-vector multiplication logic in 'demo2-crossbar'.
2. Write the Compute Shader (GLSL) to perform the MVM operation on the GPU.
3. Verify the math against a CPU reference implementation.

PHASE 4: DEMO 3 (PHASE FIELD)
1. Initialize the directory structure and scaffolding for Demo 3 based on 'demo3-phasefield/README.md'.
2. Implement the basic TDGL (Time-Dependent Ginzburg-Landau) equation solver struct.

PROTOCOL:
- Maintain a file called 'WEEKEND_PROGRESS.md'. Log every major step you complete and every issue you face.
- RUN 'go build' after every significant code change. If it fails, FIX IT immediately before moving on.
- Do not delete existing working physics code.
- If you finish all tasks or cannot proceed further, output exactly: <promise>COMPLETE</promise>.
