<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-13 | Updated: 2026-02-13 -->

# config/ — Configuration Files and Defaults

**Purpose:** Centralized YAML/JSON configuration for physics models, crossbar arrays, peripherals, materials, simulation parameters, and EDA flows. All defaults are documented here.

**Status:** Production
**Stability:** High (config schema is versioned)
**Validation:** All configs validated via `validation/configvalidator/`

## Key Files

| File | Purpose | Key Sections |
|------|---------|--------------|
| `materials.yaml` | Material parameters: HZO, AlScN, cryogenic variants | alpha, beta, gamma, Pr, Ec, thickness |
| `crossbar.yaml` | Crossbar array config: dimensions, conductance range, non-idealities | 128×128 default, 30-level quantization |
| `calibration.yaml` | Calibration fitting defaults | Temperature-dependent parameters |
| `constants.yaml` | Physical constants and unit conversions | ε₀, µ₀, Boltzmann constant, etc. |
| `energy.yaml` | Power and energy model defaults | Voltage levels, timing parameters |
| `preisach.yaml` | Preisach model config: hysteron distribution, grid size | Distribution type, bounds |
| `simulation.yaml` | Simulation parameters: solver tolerances, time step limits | ODE step sizes, convergence criteria |
| `timing.yaml` | Circuit timing: pulse widths, delays, sampling times | In ns or µs |
| `training.yaml` | Neural network training defaults | Learning rate, batch size, epochs |
| `benchmarks.yaml` | Benchmark suite configuration | Problem sizes, iteration counts |
| `physics/` | Physics configuration subdirectory | Material-specific overrides |

## Subdirectories

| Directory | Purpose | Contents |
|-----------|---------|----------|
| `physics/` | Material-specific physics configs | Landau coefficients, Preisach params per material |

## Key Configuration Parameters

### materials.yaml
```yaml
materials:
  hzo:
    alpha: [coefficient]        # Landau energy coefficient
    beta: [coefficient]         # Negative for ferroelectric
    gamma: [coefficient]        # Positive for stabilization
    pr_uc_cm2: 20               # Remanent polarization
    ec_kv_cm: 150               # Coercive field
    thickness_nm: 10            # Physical thickness
    temperature_k: 300          # Default temp
  alscn:
    [similar structure]
  cryogenic:
    temperature_k: 77           # Liquid nitrogen
```

### crossbar.yaml
```yaml
crossbar:
  default_rows: 128             # Default array size
  default_cols: 128
  quantization_levels: 30       # Discrete conductance levels
  conductance_min_s: 1.0e-6     # Gmin (1 µS)
  conductance_max_s: 100.0e-6   # Gmax (100 µS)
  device_variation: 0.05        # σ = 5% device-to-device
  read_noise: 0.02              # σ = 2%
  write_noise: 0.03             # σ = 3%
```

### simulation.yaml
```yaml
simulation:
  landau_rk4_step: 1.0e-9       # ODE solver step (1 ns)
  ispp_max_iterations: 1000     # Write-verify loop max
  ispp_tolerance_level: 1       # Accept ±1 level
  solver_tol_relative: 1.0e-6   # Relative error tolerance
  solver_tol_absolute: 1.0e-8   # Absolute error tolerance
```

## For AI Agents

### Working in This Directory

**When adding new config parameters:**
1. Add to appropriate YAML file (materials.yaml, crossbar.yaml, etc.)
2. Document the parameter with units and valid range
3. Add reference to literature/measurement source
4. Update validation rules in `validation/configvalidator/`
5. Add tests to `config_validation_test.go`

**When modifying material parameters:**
1. Only modify `materials.yaml` for calibration or new materials
2. Create new material entry (don't overwrite existing)
3. Test via `shared/physics/material_test.go`
4. Run regression tests: `FECIM_UPDATE_PHYSICS_GOLDEN=1 go test ./validation/...`
5. Commit golden data changes with explanation

**When changing simulation defaults:**
1. Modify `simulation.yaml`
2. Understand impact on solver convergence (see `shared/physics/landau_nls_test.go`)
3. Run full test suite to check for regressions
4. Document rationale in commit message

**When adding per-material overrides:**
1. Create YAML file in `physics/defaults/` directory
2. Name as `<material>_physics.yaml`
3. Override only necessary fields (others fall back to `simulation.yaml`)
4. Load via `config/physics.go` → `LoadPhysicsConfig(material)`

### Testing Requirements

**All configs must pass validation:**
```bash
go run ./validation/configvalidator/cmd/validate/ -r config/
```
Exit code must be 0 (all valid).

**Material configs must pass physics tests:**
```bash
go test -v ./shared/physics/ -run TestMaterial
```

**Simulation defaults must pass regression:**
```bash
go test -v ./validation/ -run TestPhysicsRegression
```

**Configuration schema must be consistent:**
- Each YAML file has documented structure
- Types are validated (int, float, string, array)
- Ranges are enforced (e.g., 1-1000 K for temperature)

### Common Patterns

**Material parameter usage** (in `shared/physics/material.go`):
```go
matConfig := config.LoadMaterial("hzo")  // reads from config/materials.yaml
solver := physics.NewLKSolver(matConfig.Alpha, matConfig.Beta, matConfig.Gamma)
```

**Simulation parameter usage** (in module apps):
```go
simConfig := config.LoadSimulation()
rk4Step := simConfig.LandauRK4Step    // e.g., 1.0e-9 s
isppMaxIter := simConfig.ISPPMaxIterations  // e.g., 1000
```

**Per-material overrides** (in `config/physics/defaults/`):
```yaml
# hzo_physics.yaml (material-specific solver tweaks)
landau_rk4_step: 0.5e-9    # Smaller step for this material
ispp_max_iterations: 1500  # More aggressive convergence
```

**Crossbar non-idealities** (in `module2-crossbar/`):
```go
cfg := config.LoadCrossbar()
variation := cfg.DeviceVariation  // 0.05 = 5% σ
readNoise := cfg.ReadNoise        // 0.02 = 2%
```

## Dependencies

### Internal
- `validation/configvalidator/` — Config validation rules and CLI
- `shared/physics/` — Physics engines that consume these configs
- All modules read from `config/` directory

### External
- `gopkg.in/yaml.v2` — YAML parsing
- Standard library: `encoding/json`, `io/ioutil`

## MANUAL

**Config File Locations:**
- Application runtime: loads from `config/` relative to binary
- Docker/CI: may mount as ConfigMap
- Override via env: `FECIM_CONFIG_PATH=/path/to/config`

**Material Parameter Calibration:**
1. Measure or extract from literature (P-E curves)
2. Add to `materials.yaml` under new material name
3. Run calibration: `go run ./cmd/fecim-lattice-tools mode=calibration material=<name>`
4. Extract α, β, γ from P-E curve fitting
5. Regenerate golden data: `FECIM_UPDATE_PHYSICS_GOLDEN=1 go test ./validation/...`

**Crossbar Quantization:**
Default is 30 levels (see `crossbar.yaml`):
```yaml
quantization_levels: 30  # Maps conductance to 0-29 levels
```
To change: modify `quantization_levels` field and regenerate golden data.

**Simulation Solver Tuning:**
ODE step size affects accuracy vs. performance:
```yaml
simulation:
  landau_rk4_step: 1.0e-9     # 1 ns: more accurate, slower
  # vs.
  landau_rk4_step: 5.0e-9     # 5 ns: faster, less accurate
```
Benchmark with `go test -bench=. ./validation/benchmarks/...`

**Configuration Schema Versioning:**
Each YAML file implicitly has a version:
- Fields are added, never removed
- Default values are provided for missing fields
- Breaking changes require documentation in `docs/development/CONFIG_MIGRATION.md`

**Adding New Physics Constants:**
1. Add to `constants.yaml` with units
2. Reference source/publication
3. Update any dependent calculations in `shared/physics/units.go`
4. Test with `shared/physics/units_test.go`

**Validation Rules Reference:**
See `validation/configvalidator/README.md` for detailed rules:
- Material names: non-empty string
- Temperature: 1-1000 K
- Conductance: must satisfy 0 < g_min < g_max
- Array size: 1-4096 rows/cols

**Config Precedence (if multiple sources exist):**
1. Environment variable (highest priority)
2. Command-line flag
3. File in current directory
4. `config/` relative to binary
5. Hard-coded defaults (lowest priority)

**Testing Config Changes:**
```bash
# Validate syntax
go run ./validation/configvalidator/cmd/validate/ config/materials.yaml

# Test impact on physics
go test -v ./shared/physics/ -run TestMaterial

# Full regression (slow, regenerates golden data)
FECIM_UPDATE_PHYSICS_GOLDEN=1 go test ./validation/...

# Run benchmarks to detect performance regressions
go test -bench=. ./validation/benchmarks/...
```

**Material Library Maintenance:**
- HZO variants are in `materials.yaml` (standard, SI-doped, cryogenic, FTJ)
- AlScN variants are documented separately
- Each material entry includes publication reference
- Update via pull request with calibration explanation

