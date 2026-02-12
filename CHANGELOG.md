# Changelog

All notable changes to this project are documented here.

## [2026-02-11] Massive TODO Sprint

This sprint closed a large TODO backlog across physics fidelity, validation, reliability, UX, accessibility, docs, and performance.

### Major Highlights

- **Physics model rigor expanded**
  - Added/validated Landau-Khalatnikov and Preisach ISPP/WRD behaviors.
  - Improved convergence diagnostics, staircase/remanent behavior checks, and headless evidence output.
  - Tightened model sign/unit handling and non-finite protections.

- **Module 4 circuit realism upgrades**
  - Completed major Tier-A/Tier-B solver and dispatch improvements.
  - Strengthened sense-chain wiring and READ-path observability.
  - Added ISPP coupling paths and circuit-level reporting/overlays.

- **Cross-module validation and coverage push**
  - Added broad regression coverage for critical physics paths.
  - Expanded tests for crossbar/comparison/help/themes/accessibility and UI-engine sync.
  - Increased parity checks across CLI/GUI and headless workflows.

- **Performance and stability work**
  - Reduced hot-path allocations (physics, quantization, inference pipeline).
  - Added material construction caching and bounded map/allocation safety fixes.
  - Resolved key concurrency/race hazards in shared UI/progress managers.

- **Error handling and safety hardening**
  - Replaced panic-prone paths with explicit error handling.
  - Hardened CLI/IO boundaries and renderer loop behavior.
  - Closed critical audit findings across MNIST/GUI/EDA paths.

- **UX and accessibility improvements**
  - Added keyboard shortcuts, docs search affordances, and clearer labels.
  - Improved DPI/layout resilience and readability consistency.
  - Added accessibility registry/announcements, text alternatives, reduced-motion support.

- **Documentation and honesty/audit alignment**
  - Fixed large sets of broken internal docs links.
  - Closed documentation gaps and marked heuristic/citation-needed areas clearly.
  - Synced module docs with implemented behavior and acceptance criteria.

- **EDA and architecture progress**
  - Added foundations for force-directed placement and Manhattan routing.
  - Expanded module6 validation tests and export/CLI/GUI parity checks.
  - Added multicell hysteresis API + MVM sneak trace reporting architecture work.

### Representative Commits

- `cb6c703` Implement ARCH-2 multicell API and ARCH-3 MVM sneak tracing
- `fb59aae` Implement ARCH-4 training foundation and VK-4 Vulkan cleanup
- `1b00cce` Add peripheral PVT INL/DNL model, corner analysis, and ISPP cycle trail
- `5364137` Add headless WRD/ISPP Preisach+LK regression suites with JSON summaries
- `b6faaf1` docs: fix 112 broken internal markdown links in docs
- `b09f4da` race-safety: fix top concurrency hazards in shared UI/progress managers
- `898782e` audit: fix critical error-handling gaps across mnist/gui/eda
- `c198bef` perf: reduce hot-path allocations in physics and quantization
- `75c6b02` perf: cache material construction
- `868dd72` perf: reduce inference pipeline allocations (PERF-02)

### Sprint Outcome

- Significant reduction of high-priority TODO items.
- Better alignment between documented claims and model behavior.
- Stronger test/race baseline and improved maintainability for next development cycle.
