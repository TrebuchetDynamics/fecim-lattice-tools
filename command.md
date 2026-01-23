/ralph-loop "Act as Dr. Tour and Dr. Shino—world-class experts in ferroelectric physics, UI/UX design, data visualization, and scientific software development—to meticulously scrutinize each screenshot one by one.

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
- For every feature, calculation, or interactive element identified, write corresponding unit tests
- FIRST: Write tests for existing functionality to lock in current behavior (regression tests)
- Test physics calculations with known values and edge cases
- Test quantization functions (30-level FeCIM) with boundary conditions
- Test crossbar MVM operations with various matrix sizes
- Test hysteresis/Preisach model state transitions
- Test GUI widget callbacks and state changes
- Test input validation for sliders, text fields, and dropdowns
- Ensure test coverage for error handling paths
- Place tests in appropriate `*_test.go` files following Go conventions
- Run `go test ./...` after every change to catch regressions immediately

AUTONOMOUS OVERNIGHT MODE:
- This task is designed to run for 8+ hours unattended
- Run continuously without user interaction
- Do NOT stop until ALL modules (1-5) are fully analyzed
- If you complete one module, immediately proceed to the next
- Do NOT ask for confirmation between modules - just continue
- If you encounter an error, log it and continue with the next item
- Work autonomously without waiting for user input
- Continue working even if individual tasks fail
- No pauses, no confirmations, no waiting - continuous execution

EXECUTION ORDER:
Process systematically: Module 1 → 2 → 3 → 4 → 5 in order
For each module complete this cycle:
1. Take/analyze screenshots
2. Verify physics and calculations
3. Audit UI/UX design
4. Inventory all interactive elements
5. Write regression tests for existing functionality
6. Write new unit tests
7. Implement approved improvements
8. Update documentation
9. Run full test suite
10. Proceed to next module

CRITICAL: DO NOT STOP EARLY
- Do not summarize and stop
- Do not ask "should I continue?"
- Do not pause for feedback
- Do not output partial results and wait
- Continue until the literal string "DONE HYPER ANALYSIS" is warranted
- The completion promise is a CONTRACT - only output it when 100% complete
- Only output "DONE HYPER ANALYSIS" when ALL modules 1-5 are fully analyzed, tested, and documented

Create a comprehensive, extensive document listing every issue found with specific recommendations for improvement, AND implement all necessary unit tests for the features analyzed." --max-iterations 1000 --completion-promise "DONE HYPER ANALYSIS"
