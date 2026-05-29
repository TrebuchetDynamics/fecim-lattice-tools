# TODOS — Module 1 Hysteresis Completion Checklist

## Physics Model Hardening

- [x] Preisach Everett product-form corrected (non-negative guarantee)
- [x] maxHysteresisLoopPoints bound added (1M)
- [x] maxDiscreteStateCount bound added (1M)
- [x] isValidPreisachMaterial checks added (finite, coercive representable)
- [x] Preisach simulation numeric stability: finite checks, no NaN/Inf propagation
- [x] Landau-Khalatnikov ODE numeric stability: dt bounded, finite damping
- [ ] **[OPEN]** Preisach-NLS hybrid: still missing Everett NLS product-form integration
- [ ] **[OPEN]** Preisach Delta tuning needs golden ensemble sweep validation

## ISPP Write Controller

- [x] Guard band logic for overshoot direction flip prevention
- [x] guardActive flag tracks protection state
- [x] guardSign correction direction storage
- [x] guardCount limited to prevent false ACCEPT
- [x] overshootLimit = 30 triggers SUCCESS (not FAIL)
- [x] writeAttempts default 10 → now adjustable
- [x] ACCEPT ±1 guard interaction: `guardActive` skips ACCEPT (line 575: `&& !guardActive`)
- [x] Bounds collapse after overshoot: direction-aware widening with `minBracketWidthFrac` (lines 630-656)
- [x] Guard-band flattening: `maxGuardPulses=2` (line 490), `guardSign` clamped (line 780)
- [x] Write attempt limit: `MaxRetries` field configurable, `ForceResetLimit` for high-variation materials

## Simulation Engine

- [x] Multicell IR-drop array physics with sneak path compensation
- [x] State machine: APPLY→WAIT→VERIFY→loop
- [x] Headless CLI calibration runner
- [x] TUI text interface with keyboard controls
- [x] Waveform generation (sine, triangle, sawtooth)
- [x] Hysteresis polarization loop visualization (P-E plot)
- [x] State history ring buffer with configurable size

## GUI / Rendering

- [ ] **[OPEN]** Fyne legacy thread safety: `fyne.Do()` wrapper needed for all widget updates
- [ ] **[OPEN]** Gogpu default path should replace Fyne as primary UI surface
- [ ] **[OPEN]** Plot margin responsiveness: should resize with window
- [x] Headless headless regression test suite runs via CLI

## Validation

- [x] Golden regression data match (Preisach Everett product-form correct)
- [x] Physics validation tests pass against literature benchmarks
- [ ] **[WARN]** FECIM_UPDATE_PHYSICS_GOLDEN=1 must be set to regenerate golden files
- [ ] **[OPEN]** Writer stress test included in CI? It's offline only now
- [ ] **[OPEN]** Ensemble tests (all 9 materials × 2 engines) should pass full cycles

## Documentation

- [x] CONTEXT.md updated with current state and behavioral rules
- [x] AGENTS.md written for ISPP guard interactions
- [x] diagnostics_remanent_staircase.md written for Everett negative fix history
- [x] ADR 0001: Level Calibration Workflow UI Architecture (docs/adr/0001-level-calibration-workflow-ui-architecture.md)
- [x] ADR 0002: ISPP Convergence Guard Logic (docs/adr/0002-ispp-convergence-guard-logic.md)
- [ ] **[TODO]** docs/3-develop/api-reference.md needs updated writing-guide tables
- [ ] **[TODO]** module1-hysteresis/pkg/controller/stress_test.go move into CI pipeline

## Integration

- [ ] **[OPEN]** Module 3-MNIST should interface with Preisach → charge injection mapping
- [ ] **[OPEN]** Module 6-EDA export uses S-P-E curves from Module 1
- [ ] **[OPEN]** Crossbar Module 2 depends on Preisach non-ideality calculations

## Immediate Next Steps

1. **Fix ACCEPT ±1 guard issue** — skip ACCEPT when guardActive=true
2. **Fix bounds collapse after overshoot** — widen to at least [0,targetV]
3. **Move writer_stress_test.go into CI** — add to controller test package
4. **Create ADR for Level Calibration Workflow** — document new workflow design
5. **Update API reference docs** — reflect current writer.go/preisach.go structures

## Skills Required

- **TDD (test-driven development)** — Needed for guard-logic fixes
- **Fyne (legacy GUI)** — Thread safety documentation (if still maintained)
- **gogpu (UI)** — Default interface path migration
- **Preisach physics** — Everett product-form understanding
- **ISPP controller** — Guard band binary search stress convergence
- **Physics golden regression** — One-time regeneration for new Everett form