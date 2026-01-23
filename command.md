/ralph-loop "Act as Dr. Tour and Dr. Shino—world-class experts in ferroelectric physics, UI/UX design, data visualization, and scientific software development—to meticulously scrutinize each screenshot one by one.

QUICK START - DO THIS FIRST:
1. Run: ls <local-path> to see all 67 screenshots
2. Create HYPER_ANALYSIS_REPORT.md with initial header
3. Begin Module 1 analysis immediately

SCOPE: Focus ONLY on Modules 1-5 for now:
- Module 1: Hysteresis (P-E curve visualization, Preisach model)
- Module 2: Crossbar+ (MVM + Non-Idealities: Ideal, IR Drop, Sneak Paths, Drift)
- Module 3: MNIST (Neural network digit recognition, FP vs CIM dual mode)
- Module 4: Circuits (DAC/ADC/TIA peripheral design)
- Module 5: Comparison (Technology comparison and technical briefing)
Do NOT analyze Module 6 (EDA) or any other modules at this time.

AVAILABLE RESOURCES:
You have full access to the following documentation for reference and updates:
- <local-path> (main project documentation)
- <local-path> (full documentation directory)
- <local-path> (project guidelines and conventions)
- <local-path> (ALL screenshots to analyze)
Use these resources to:
- Understand module specifications, physics constants, and design patterns
- Cross-reference claimed values against documented specifications
- Update documentation when improvements are made
- Ensure changes align with project conventions and coding standards
- Reference Dr. Tour's quotes and verified specifications

For each screenshot, analyze and document:

Physics & Calculations:
- Verify all equations, constants, and formulas for correctness
- Check unit consistency (µC/cm², MV/cm, etc.)
- Validate numerical values against published research
- Ensure the 30-level FeCIM quantization is accurately represented
- Confirm hysteresis curves, P-E relationships, and Preisach model behavior

UI/UX Design:
- Layout balance, spacing, and visual hierarchy
- Color scheme consistency and accessibility (contrast ratios)
- Typography: font choices, sizes, readability, and alignment
- Widget placement and intuitive flow
- Responsive behavior and screen utilization

Proposed UI Improvements (Detailed):
- For each screen, propose specific UI enhancements with thorough explanations
- Describe WHY each improvement matters (user experience, clarity, professionalism)
- Provide exact specifications: pixel values, color hex codes, font sizes, padding/margins
- Suggest better widget arrangements with mockup descriptions or ASCII layouts
- Recommend improved labeling, tooltips, and contextual help text
- Propose animation/transition improvements with timing details
- Identify opportunities for visual feedback (hover states, click feedback, progress indicators)
- Suggest information density optimizations (what to show/hide, collapsible sections)
- Recommend accessibility improvements (keyboard navigation, screen reader support, high contrast)
- Prioritize improvements as: Critical, High, Medium, Low impact

Data Visualization:
- Axis labels, units, and scales on all charts/graphs
- Legend clarity and placement
- Heatmap color gradients and value mappings
- Animation smoothness and timing
- Visual accuracy of crossbar arrays and neural network representations

Technical Quality:
- Identify any visual bugs, glitches, or rendering issues
- Check for truncated text, overlapping elements, or misalignment
- Evaluate loading states and error handling displays

Interactive Elements Inventory:
- List ALL interactive inputs visible in each screenshot (buttons, sliders, dropdowns, checkboxes, text fields, tabs, toggles, etc.)
- For each input, note its purpose and verify it appears functional
- Check "Read Cells", "Program", "Reset", and similar action buttons
- Validate slider ranges and step values make sense for the parameter
- Ensure dropdowns have appropriate options visible
- Flag any inputs that appear broken, unresponsive, or poorly labeled

Educational Clarity:
- Are concepts explained clearly for the target audience?
- Do tooltips and labels enhance understanding?
- Is the relationship between memory and computation evident?

Regression & Stability (DO NOT BREAK EXISTING FUNCTIONALITY):
- CRITICAL: Document all currently working features BEFORE making any changes
- Run existing tests and ensure they all pass before modifications
- For each proposed change, verify it does not break existing behavior
- If a change might affect other components, trace all dependencies first
- Create backup/snapshot of current state before implementing fixes
- Test all existing workflows end-to-end after each modification
- If something breaks, revert immediately and reassess the approach
- Maintain backwards compatibility for any public APIs or interfaces

Unit Test Requirements:
- Focus on TESTABLE LOGIC ONLY (skip GUI widget callbacks - Fyne GUI is hard to unit test)
- FIRST: Write tests for existing functionality to lock in current behavior (regression tests)
- Test physics calculations with known values and edge cases
- Test quantization functions (30-level FeCIM) with boundary conditions
- Test crossbar MVM operations with various matrix sizes
- Test hysteresis/Preisach model state transitions
- Test core package functions in pkg/core/, pkg/crossbar/, pkg/ferroelectric/
- Place tests in appropriate `*_test.go` files following Go conventions
- Run `go test ./...` after changes - if tests fail, log the failure and continue (don't block)

WORKLOAD ESTIMATE:
- 67 screenshots to analyze across 5 modules
- ~10-step cycle per module = 50+ major operations
- Estimated iterations needed: 500-800
- Buffer for errors/retries: 200 iterations
- Total allocated: 1000 iterations (should be sufficient)

AUTONOMOUS OVERNIGHT MODE:
- This task is designed to run for 8+ hours unattended
- Run continuously without user interaction
- Do NOT stop until ALL modules (1-5) are fully analyzed
- If you complete one module, immediately proceed to the next
- Do NOT ask for confirmation between modules - just continue
- Work autonomously without waiting for user input
- No pauses, no confirmations, no waiting - continuous execution

ERROR RECOVERY (NEVER STOP ON ERRORS):
- If a screenshot fails to load: log it, skip it, continue to next screenshot
- If a test fails: log the failure, continue to next task
- If a file write fails: retry once, then log and continue
- If code analysis is unclear: document uncertainty and move on
- If you cannot find expected files: log missing files and continue
- NEVER let any single failure stop the entire process
- Log all errors to HYPER_ANALYSIS_REPORT.md under "## Errors Encountered"

CHECKPOINTING (SAVE PROGRESS FREQUENTLY):
- After completing EACH screenshot analysis, append findings to HYPER_ANALYSIS_REPORT.md immediately
- After completing EACH module, write a "## Module X COMPLETE" section
- This ensures partial progress is saved even if something fails later

OUTPUT LOCATION:
Write ALL findings to: <local-path>
- Continuously append to this file as you work
- Use clear markdown sections for each module
- Include timestamps for each major section completed

EXECUTION ORDER:
Process systematically: Module 1 → 2 → 3 → 4 → 5 in order

For each module complete this SIMPLIFIED cycle:
1. LIST all screenshots for this module (ls screenshots/*moduleXX*)
2. ANALYZE each screenshot one by one:
   - Read the image
   - Document physics/UI/interactive elements findings
   - Append to HYPER_ANALYSIS_REPORT.md immediately
3. WRITE tests for core logic (skip GUI tests)
4. IMPLEMENT only SAFE fixes:
   - Typo corrections
   - Obvious spacing/alignment issues
   - DO NOT touch algorithms, physics, or architecture
   - When in doubt, DOCUMENT instead of fixing
5. Mark module COMPLETE in report and proceed

SUCCESS CRITERIA (what "DONE" means):
✓ All 67 screenshots analyzed and documented
✓ HYPER_ANALYSIS_REPORT.md contains findings for all 5 modules
✓ At least one new test file per module with testable logic
✓ All existing tests still pass (`go test ./...`)
✓ Each module has a "COMPLETE" marker in the report

CRITICAL: DO NOT STOP EARLY
- Do not summarize and stop
- Do not ask "should I continue?"
- Do not pause for feedback
- Do not output partial results and wait
- Continue until the literal string "DONE HYPER ANALYSIS" is warranted
- The completion promise is a CONTRACT - only output it when 100% complete
- Only output "DONE HYPER ANALYSIS" when ALL modules 1-5 are fully analyzed, tested, and documented

FALLBACK IF APPROACHING ITERATION LIMIT:
- If you sense you're running low on iterations, prioritize in this order:
  1. HIGHEST: Complete HYPER_ANALYSIS_REPORT.md with all screenshot analyses
  2. HIGH: Write "## Module X COMPLETE" markers for finished modules
  3. MEDIUM: Write at least basic tests for core physics functions
  4. LOW: Implement safe fixes
- A complete analysis document WITHOUT fixes is SUCCESS
- An incomplete document WITH fixes is FAILURE

Create a comprehensive, extensive document listing every issue found with specific recommendations for improvement, AND implement all necessary unit tests for the features analyzed." --max-iterations 1000 --completion-promise "DONE HYPER ANALYSIS"
