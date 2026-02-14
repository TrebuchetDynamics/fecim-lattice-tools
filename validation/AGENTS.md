<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-13 | Updated: 2026-02-13 -->

# validation/ — Statistical Validation and Config Checking

**Purpose:** Statistical comparison of simulation outputs against literature data, configuration file validation, regression testing, and integration test harnesses. Ensures physics accuracy and configuration integrity.

**Status:** Production
**Stability:** High (mature validation framework)
**Test Coverage:** 90%+ (comprehensive test suites)

## Key Files

| File | Purpose | Key Type/Function |
|------|---------|-------------------|
| `literature.go` | Load and parse published experimental P-E curves, endurance, retention | `LiteratureDataset` |
| `statistics.go` | Statistical comparison: KS test, chi-square, RMSE, MAE, correlation | `KSTest()`, `ChiSquare()`, `RMSE()` |
| `interfaces.go` | Shared validation interfaces | `Validator`, `ComparableResult` |
| `readiness_report.go` | Generate release readiness reports | `ReadinessReport` |
| `physics_regression_test.go` | Regression testing against golden physics data | Validates P-E loops, endurance |
| `literature_validation_test.go` | Compare simulation to published literature | KS tests vs. experimental datasets |
| `montecarlo_validation_test.go` | Monte Carlo ensemble validation | Process variation coverage |
| `integration_test.go` | Cross-module integration validation | Module1 + Module2 consistency |
| `m1_m4_physics_consistency_test.go` | Module1 (hysteresis) ↔ Module4 (circuits) physics consistency | ISPP engine alignment |
| `determinism_test.go` | Verify deterministic physics (no floating-point jitter) | Seed consistency checks |
| `config_validation_test.go` | YAML/JSON config structural validation | Config schema enforcement |
| `validation_test.go` | Core statistical function tests | Unit tests for KS, chi-square, etc. |

## Subdirectories

| Directory | Purpose | Key Exports |
|-----------|---------|-------------|
| `configvalidator/` | Rule-based config validation engine | Validator CLI tool + library |
| `benchmarks/` | Performance benchmarks for physics and crossbar | `benchmark_suite.go` |
| `calibration/` | Parameter calibration for materials | Calibration routines |
| `comparator/` | Result comparison tools: CSV diff, JSON merge | Comparative analysis |
| `external/` | External tool integrations (Heracles, CrossSim, ngspice) | Tool wrapper scripts |
| `heracles/` | Heracles compact model integration | HZO comparison harness |
| `integration/` | Integration test suites (crossbar + hysteresis) | Multi-module test drivers |
| `testdata/` | Test data: regression baselines, literature data, example outputs | Golden data + fixtures |

## Key Subdirectory Files

### `configvalidator/`
| File | Purpose |
|------|---------|
| `validator.go` | Core validation engine with rule registration |
| `preisach.go` | Preisach config validation (hysteron count, distribution) |
| `calibration.go` | Calibration config validation (field ranges, temps) |
| `weight_matrix.go` | Weight matrix validation (dimensions, NaN/Inf checks) |
| `array_design.go` | Array design validation (rows/cols, architecture) |
| `openlane.go` | OpenLane flow config validation (paths, required fields) |
| `cmd/validate/main.go` | Standalone CLI tool |
| `README.md` | CLI documentation and validation rules |

### `testdata/`
| Subdirectory | Contents |
|---|---|
| `physics_regression/` | Golden P-E loops, endurance curves, regression baselines (regenerated via `FECIM_UPDATE_PHYSICS_GOLDEN=1`) |
| `ispp_regression/` | ISPP write-verify convergence data for 9 materials |
| `literature/` | Published experimental datasets (HZO, AlScN, cryogenic) |

## For AI Agents

### Working in This Directory

**When adding validation:**
1. Use `statistics.go` functions (KSTest, ChiSquare, RMSE) for distribution comparison
2. Add literature datasets to `literature.go` using standard parsing
3. Create test in `validation_test.go` or dedicated `*_validation_test.go`
4. Ensure test passes with tolerance thresholds (see comments in test files)

**When validating configuration:**
1. Add validation rules to `configvalidator/validator.go`
2. Implement type-specific validation (e.g., `preisach.go`, `array_design.go`)
3. Test with `configvalidator/cmd/validate/` CLI tool
4. Add test cases to `configvalidator/validator_test.go`

**When comparing external tools:**
1. Use `external/` wrappers for Heracles, CrossSim, ngspice
2. Store comparison results in `testdata/` for reproducibility
3. Add comparator logic to `comparator/` if needed
4. Document tool version pins in `tools/external/README.md`

**When running integration tests:**
1. Module1 + Module2 tests live in `integration/`
2. Use `integration_test.go` for cross-module consistency
3. Physics consistency between Module1 and Module4 in `m1_m4_physics_consistency_test.go`

**When adding benchmarks:**
1. Define benchmark suite in `benchmarks/benchmark_suite.go`
2. Benchmark naming: `Benchmark<What><Size>` (e.g., `BenchmarkLKSolver128x128`)
3. Run with `go test -bench=. ./validation/benchmarks/...`

### Testing Requirements

**Regression tests must pass:**
- `physics_regression_test.go` compares outputs to golden data in `testdata/physics_regression/`
- Golden data is regenerated ONLY with `FECIM_UPDATE_PHYSICS_GOLDEN=1` after verified physics changes
- Typical tolerances: ≤5% relative error for Pr, Ec; ≤2% for saturation level

**Literature validation must pass:**
- KS test p-value > 0.05 for distribution shape comparison
- RMSE < 10% of max P for point-wise error
- Tests compare simulated P-E curves to published datasets

**Integration tests must pass:**
- `integration_test.go` validates Module1-Module2 consistency
- `m1_m4_physics_consistency_test.go` ensures ISPP engine alignment
- `determinism_test.go` verifies reproducible results with fixed seed

**Config validation must pass:**
- All YAML/JSON in `config/` and `data/` must validate
- CLI tool: `go run ./validation/configvalidator/cmd/validate/ -r config/`
- Exit code 0 = all valid, 1 = one or more invalid

**Benchmark baselines:**
- Benchmarks run on CI for performance regression detection
- Baseline stored in `benchmarks/bench_baseline.txt` (committed to repo)
- Track throughput (cells/s) and latency (µs/cell) across releases

### Common Patterns

**Statistical comparison** (in `statistics.go`):
```
SimulatedData, ExperimentalData → KSTest() → p-value (p > 0.05 = pass)
OR: RMSE(simulated, experimental) < threshold
```

**Regression golden data** (in `testdata/physics_regression/`):
```
JSON files: {"material": "hzo", "loops": [...], "metrics": {"pr_uc": ..., "ec_kv": ...}}
Regenerate via: FECIM_UPDATE_PHYSICS_GOLDEN=1 go test ./validation/...
```

**Config validation** (in `configvalidator/`):
```
ConfigFile → detectType() → Validator.Validate() → Result{Valid, Errors, Warnings}
Supported types: calibration, preisach_state, array_design, weight_matrix, openlane
```

**Integration test** (in `integration_test.go`):
```
Load Module1 config → Run hysteresis simulation → Extract P-E curve
Load Module2 config → Run crossbar with same curve → Compare outputs
```

**Literature dataset** (in `literature.go`):
```
LiteratureDataset struct holds published P-E curves, metadata
ParseLiteratureJSON() → []LiteratureDataset
Compare with simulated via KSTest or RMSE
```

## Dependencies

### Internal
- `shared/physics/` — Core physics engines (Preisach, Landau-K, ISPP, material params)
- `shared/export/` — Export pipeline for validation outputs
- `shared/logging/` — Logging infrastructure
- `module1-hysteresis/` — Physics-based hysteresis simulation
- `module2-crossbar/` — Crossbar array and MVM
- `module4-circuits/` — Circuit simulation (ISPP engine, peripherals)

### External
- `encoding/json`, `encoding/csv` — Data parsing
- `math/stats` — Statistical functions (KS test implementation)
- Standard library: `testing`, `flag`, `os`, `io`

## MANUAL

**Regenerating Physics Golden Data:**
After VERIFIED physics changes (architectural fixes or bug fixes):
```bash
FECIM_UPDATE_PHYSICS_GOLDEN=1 go test ./validation/...
git add validation/testdata/physics_regression/*.json
git commit -m "physics: update golden data (reason)"
```
This regenerates P-E loops, endurance curves, and regression baselines.

**Running Full Validation Suite:**
```bash
go test -v ./validation/...           # All validation tests
go test -race ./validation/...        # With race detector
go test -bench=. ./validation/...     # Include benchmarks
```

**Validation Config Files:**
Test with CLI tool:
```bash
go build -o bin/validate ./validation/configvalidator/cmd/validate
./bin/validate -r config/              # Recursively validate all JSON/YAML
./bin/validate -w data/                # Show warnings too
./bin/validate -s config.json          # Summary-only (for CI)
```

**Literature Datasets:**
- HZO: `testdata/literature/hzo_*.json`
- AlScN: `testdata/literature/alscn_*.json`
- Cryogenic: `testdata/literature/cryo_*.json`
- Format: `[{field: E_kvcm, polarization_uccm2, ...}, ...]`
- Add new: implement `ParseLiteratureXXX()` in `literature.go`

**Benchmark Baseline:**
Track performance regressions:
```bash
go test -bench=. ./validation/benchmarks/... -benchmem > benchmarks/bench_baseline.txt
git add benchmarks/bench_baseline.txt
```

**External Tool Integration:**
- Heracles: `external/heracles_wrapper.go` calls HZO compact model
- CrossSim: `external/crosssim_wrapper.go` calls crossbar simulator
- ngspice: `external/spice_wrapper.sh` validates exported netlists
- All tools are OPTIONAL; validation runs without them

**Regression Test Golden Data Locations:**
- Physics (P-E loops): `testdata/physics_regression/preisach_loop_*.json`
- ISPP convergence: `testdata/ispp_regression/*_convergence.json`
- Endurance: `testdata/physics_regression/*_endurance.json`
- Update via: `FECIM_UPDATE_PHYSICS_GOLDEN=1 go test ./validation/...`

**Determinism Verification:**
Run `determinism_test.go` to ensure bit-exact reproducibility:
```bash
go test -v ./validation/ -run TestDeterminism
```
This seeds RNG, runs simulation, checks output matches previous run.

**Config Validator Extensibility:**
To add a new config type:
1. Add const in `configvalidator/validator.go`
2. Update `detectConfigType()` with indicator fields
3. Create `<type>.go` with validation rules
4. Add tests to `validator_test.go`
5. Update README with new type docs

**Statistics Test Thresholds:**
- KS test p-value: typically > 0.05 (accept) or < 0.01 (reject)
- RMSE: typically < 10% of max value
- Relative error: typically < 5% for key metrics (Pr, Ec)
- Tolerances are comment-documented in each test

