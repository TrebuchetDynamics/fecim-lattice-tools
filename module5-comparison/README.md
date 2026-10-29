# Module 5: Comparison

An educational assumption workbench, not a measured device benchmark or investment forecast. Go owns calculations, source evidence, display/speech values and exports; React provides controls and semantic tables.

## Browser workflow

Build either profile using the [shared web instructions](../web/README.md), then open `/modules/comparison`:

- Choose MNIST, ResNet-50, BERT-Base, GPT-2 or LLM-70B; enter **inferences per second** (positive, finite, at most 1e9).
- Choose conservative, baseline or optimistic assumptions and Technical Review or Presentation (summary view without the raw assumptions table; full source text remains available). **Calculate** applies these edits; tables and exports retain the last successful result until then. The separate sensitivity-output selection applies on **Rank sensitivity** and is preserved by Calculate. Reset restores GPT-2, baseline, 10000 inf/s and Technical Review.
- Inspect energy, electricity-only monthly cost, assumed latency/TCO ratios, source evidence, market and fabrication context. Compare the selected scenario with baseline, or rank actual minus/plus 10% one-input sensitivities. Ties use input-name order.
- The separate **architecture-demonstration** table uses the original TOPS/TDP/memory-bandwidth protocol, batch 1. Its results do not replace scenario per-MAC assumptions. Neither protocol aggregates new Crossbar/MNIST/Circuits screen results.
- Hosted calculations are ephemeral Go/WASM Worker requests. No scientific upload, persistence, clipboard operation or download is provided. One scientific computation may be active per tab. A calculation is a bounded synchronous Go operation, not an asynchronous study/job. Failed initialization closes its acquired Worker/session; Retry creates a fresh connection.
- **Local export:** build the local profile and matching native application, start `fecim-lattice-tools serve --root DIRECTORY`, and open its launch link. Create/open a project in Projects before entering Comparison. Use Refresh opened projects if a project was opened after visiting Comparison, select its export project, and activate Export JSON or Export CSV. Immutable content-named files are written beneath that project's `exports/comparison/`; the browser cannot supply a path. A static local-profile preview does not grant an authenticated API or filesystem capability.

Browser tests are automated candidate checks, **not** Orca speech qualification. Real Orca + Firefox/Linux acceptance for both profiles remains **NOT RUN**.

## Dimensional identities and artifact versions

The workload MAC counts preserve the native calculator protocol: MNIST 101632, ResNet-50 4e9, BERT-Base 11e9, GPT-2 35e9, LLM-70B 140e12. These are educational model inputs, not device measurements or a parameter-count substitution.

- J/inference = MACs/inference × fJ/MAC × 1e-15.
- Daily inferences = inferences/second × 86400.
- Daily kWh = J/inference × daily inferences / 3.6e6.
- Annual electricity savings in **million USD** = (GPU daily kWh − FeCIM daily kWh) × 365 × assumed USD/kWh / 1e6.
- The existing monthly electricity-only protocol explicitly assumes **730 hours/month** and **$0.10/kWh**. Scenario TCO assumptions are a separate USD/month input, not an electricity measurement. No payback period is asserted without a capital-cost denominator.

For example, 1e9 MACs at 1e6 fJ/MAC is 1 J/inference. At 2 inf/s this is 172800 inferences/day, 0.048 kWh/day, 2 W, and $0.146/month under that monthly protocol.

New local and Fyne JSON/CSV artifacts declare `fecim-comparison/2` and model `scenario-estimates/2`. Version 2 fixes the former export's missing daily-rate conversion, incorrect J→kWh `/3600`, million-USD `/1000`, and inconsistent workload lookup. Old unversioned exports are **not silently reinterpreted or migrated**. Canonical JSON excludes clocks, transport/session IDs and presentation mode; unavailable CO2 is null with a reason. CSV records versions, units, assumptions and unavailable-CO2 text alongside the same canonical values.

## Scientific limitations

- Baseline CPU/DRAM, GPU/HBM and FeCIM energies are **1000, 100 and 1 pJ/MAC**, respectively, with the unchanged scenario alternatives in `pkg/comparison/scenario_profiles.go`. These are simulation assumptions, **not validated performance, energy multipliers or silicon specifications**. Datasheet descriptions are source context, not proof of these aggregate assumptions.
- Scenario CO2 reduction is **not available**: the old calculation multiplied currency by kg/kWh. No physical energy-duration input supports that scenario claim. Emissions-intensity presets are assumptions, never calibrated evidence. The separate native architecture demo retains its explicitly assumed physical energy protocol; it does not validate scenario CO2.
- Source-confidence tags are qualitative bookkeeping, **not statistical confidence intervals**. The former `abs(value) × (1 − confidence)` bounds are not presented or exported as uncertainty. Sensitivity shows actual ±10% input perturbations, with unavailable/zero denominators withheld.
- Market/fabrication numbers are inherited illustrative context, not verified forecasts, additive addressable-market proof, a manufacturing quotation, or a readiness certification. Analog multilevel yield and memory integration remain risks. No new fitting, coefficient tuning, literature promotion, MNIST accuracy or retention/endurance claim is made here.

## Implementation and checks

- `pkg/comparison/`: scenario presets/evidence, dimensional calculations, versioned serializers, and the separate original `architecture.go` demonstration/CLI model.
- `pkg/gui/`: deprecated Fyne compatibility; delegates scenario calculations/evidence and versioned exports to the shared model. Launch defaults are unchanged.
- `../workbench/modules/comparison/`: revision-checked module port, bounded Worker envelope (16 KiB), Go presentation, native rooted exports.
- `../shared/viewmodel/comparison/`: preserves the legacy architecture snapshot and supplies the actual web port.
- `../web/src/modules/comparison/`: discovered Screen, profile-specific transports and WASM entry; `../cmd/fecim-comparison-wasm/` provides a native/JS production entry for parity.

```bash
go test ./workbench/modules/comparison ./shared/viewmodel/comparison ./module5-comparison/pkg/comparison -count=1
env CGO_ENABLED=1 xvfb-run -a go test ./module5-comparison/pkg/gui -count=1
env CGO_ENABLED=1 xvfb-run -a go test ./workbench/server -run '^TestModuleTransportContract$' -count=1
# Set FECIM_FIREFOX_EXECUTABLE if Firefox is installed outside Playwright's default cache:
env FECIM_WASM_PARITY=1 CGO_ENABLED=1 go test ./validation/wasm -run 'TestModuleProductionParity|TestExampleResultsJSONNativeWASM' -count=1 -v
cd web
npm test && npm run typecheck && npm run build:hosted && npm run build:local
npm run test:e2e -- test/e2e/comparison.spec.ts
FECIM_WEB_PROFILE=local npm run test:e2e -- test/e2e/comparison.spec.ts
```

`TestComparisonWebUnits` checks hand-computable dimensions and artifact identities. `TestComparisonWebWorkflow` covers every workload/scenario and invalid/unavailable states. The production parity case compares 19 requests' original Go scientific artifacts and complete Go replies (including architecture, diff and sensitivity numeric tables) across native, Node and Firefox Worker; the original example-report gate is unchanged. The transport contract exercises Comparison's advertised native session, canonical values, stale-revision rejection and close lifecycle. Feature browser tests exercise actual controls/tables/evidence/exports, hosted isolation and cold-route budgets. These checks do not establish human accessibility, hardware validation or release qualification.

See the [module learning guide](../docs/modules/comparison/README.md), [honesty audit](../docs/research/honesty-audit.md) and [approved M5 roadmap](../docs/internals/superpowers/plans/2026-10-28-all-modules-web-roadmap.md#m5-comparison--web-comparison) for context and normative boundaries.
