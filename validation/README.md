# Validation Package

Statistical validation, literature comparison, and configuration verification for FeCIM simulation outputs. Ensures simulation results stay within tolerance of published experimental data and that configuration files are well-formed.

## Overview

The `validation/` package provides two layers of quality assurance: (1) statistical comparison of simulation outputs against literature datasets, and (2) structural validation of YAML/JSON configuration files used across all modules. It supports regression testing to catch physics drift between releases.

## Package Structure

### Root — Statistical Validation

- **literature.go** — Literature dataset loading and parsing (published P-E curves, endurance data, retention measurements)
- **statistics.go** — Statistical comparison functions: Kolmogorov-Smirnov test, chi-square test, RMSE, MAE, relative error, correlation coefficients
- **interfaces.go** — Shared interfaces for validation backends

### `configvalidator/` — Configuration Validation

- **validator.go** — Core validation engine with rule registration
- **preisach.go** — Preisach model config validation (hysteron count, distribution bounds)
- **calibration.go** — Calibration config validation (field ranges, temperature limits)
- **weight_matrix.go** — Weight matrix validation (dimensions, value ranges, NaN/Inf checks)
- **array_design.go** — Array design config validation (rows/cols, architecture, PDK)
- **openlane.go** — OpenLane flow config validation (paths, required fields)
- **cmd/validate/main.go** — Standalone CLI validation tool

## Key Types and Functions

| Type / Function | Package | Description |
|---|---|---|
| `LiteratureDataset` | `validation` | Parsed experimental data from publications |
| `KSTest` | `validation` | Kolmogorov-Smirnov two-sample test |
| `ChiSquare` | `validation` | Chi-square goodness-of-fit test |
| `RMSE`, `MAE` | `validation` | Root-mean-square and mean-absolute error |
| `Validator` | `configvalidator` | Rule-based config validation engine |
| `ValidatePreisach` | `configvalidator` | Preisach config rules |
| `ValidateWeightMatrix` | `configvalidator` | Weight matrix integrity checks |
| `ValidateArrayDesign` | `configvalidator` | Array config rules |

## Testing

```bash
# Run all validation tests
go test ./validation/...

# With race detector
go test -race ./validation/...

# Specific suites
go test -v ./validation/ -run TestStatistics
go test -v ./validation/ -run TestPhysicsRegression
go test -v ./validation/ -run TestIntegration
go test -v ./validation/configvalidator/...

# Standalone config validator CLI
go run ./validation/configvalidator/cmd/validate/ -config path/to/config.yaml
```

Key test suites:
- `validation_test.go` — Core statistical functions
- `physics_regression_test.go` — Physics output regression against golden values
- `integration_test.go` — End-to-end validation pipelines
- `integration/` — Cross-module integration tests (crossbar + hysteresis)
- `configvalidator/` — Config rule coverage

## Physics Context

**Regression baselines:** Golden P-E loops and key metrics (P_r, E_c, 2P_r) are stored in `testdata/` directories. Each release must reproduce these within defined tolerances (typically ≤5% relative error for P_r and E_c).

**Literature comparison:** Published experimental data from HZO and AlScN papers are parsed and compared against simulation output using KS tests (p > 0.05 threshold) and RMSE metrics.

**Statistical tests:**
- **Kolmogorov-Smirnov:** Distribution shape comparison between simulated and experimental P-E curves
- **Chi-square:** Goodness-of-fit for discrete level distributions
- **RMSE/MAE:** Point-by-point error quantification

## Related Documentation

- `docs/4-research/validation/` — Validation results and boundary docs
- `docs/4-research/physics-validation.md` — Physics validation criteria
- `docs/3-develop/testing/TESTING.md` — Testing conventions
- `docs/4-research/honesty-audit.md` — Claims verification
