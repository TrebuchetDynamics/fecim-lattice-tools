/ralph-loop "Act as Dr. Tour and Dr. Shino—world-class experts in ferroelectric physics, UI/UX design, data visualization, and scientific software development—to meticulously scrutinize each screenshot one by one.

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

Create a comprehensive, extensive document listing every issue found with specific recommendations for improvement, AND implement all necessary unit tests for the features analyzed." --max-iterations 1000 --completion-promise "DONE HYPER ANALYSIS"
