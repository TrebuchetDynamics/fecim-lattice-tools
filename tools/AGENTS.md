<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-13 | Updated: 2026-02-13 -->

# tools/ — External Tool Integrations and Wrappers

**Purpose:** Integration points and wrappers for external open-source tools (Heracles, CrossSim, ngspice, Icarus Verilog, Verilator, OpenROAD, OpenLane2). These tools are OPTIONAL; FeCIM works without them.

**Status:** Production
**Stability:** High (stable external interfaces)
**Optional:** Yes—all external tools are optional dependencies

## Key Files

| File | Purpose | Tool | Version |
|------|---------|------|---------|
| `external/README.md` | Tool inventory, installation instructions, version pins | Reference | — |
| `external/heracles_wrapper.go` | Wrapper for Heracles HZO compact model | Heracles | v0.4.0 |
| `external/crosssim_wrapper.go` | Wrapper for CrossSim crossbar simulator | CrossSim | v3.1.0 |
| `external/spice_wrapper.sh` | SPICE netlist validation and simulation | ngspice | 42 |
| `external/verilog_wrapper.sh` | Verilog syntax checking and simulation | iverilog + Verilator | 12.0 / 5.028 |
| `external/openroad_wrapper.sh` | Place-and-route via OpenROAD | OpenROAD | v2.0-2025.02 |
| `external/openlane_wrapper.sh` | End-to-end RTL→GDS flow via OpenLane2 | OpenLane2 | v2.1.2 |

## Subdirectories

| Directory | Purpose | Contents |
|-----------|---------|----------|
| `external/` | External tool wrappers and integration scripts | Bash/Go wrapper files, README |

## For AI Agents

### Working in This Directory

**When adding a new external tool:**
1. Create wrapper in `external/` (naming: `<tool>_wrapper.sh` or `<tool>_wrapper.go`)
2. Document tool purpose, version pin, and installation method in `external/README.md`
3. Add error handling for missing tool (graceful fallback)
4. Add tests in `validation/external/<tool>_test.go` (optional, but recommended)
5. Update tool inventory table in `external/README.md`

**When updating a tool version:**
1. Update version pin in `external/README.md`
2. Test with new version: run integration tests
3. Document breaking changes in commit message
4. Update CI/CD pipeline versions if applicable

**When integrating tool output:**
1. Parse tool output format (JSON, CSV, or text)
2. Add parser function to wrapper (e.g., `parseHeracelesOutput()`)
3. Map tool output to FeCIM data structures
4. Add tests for parser

**When calling external tools from Go code:**
1. Use `os/exec` package with timeout
2. Capture stdout/stderr separately
3. Check exit code
4. Document expected output format
5. Add graceful fallback if tool is missing

### Testing Requirements

**Tool wrappers must degrade gracefully:**
- If tool is not installed, wrapper should return error (not panic)
- FeCIM functionality should not break if optional tool is missing
- Core simulation runs without external tools

**Integration tests are optional but recommended:**
```bash
# If external tool is installed, run integration test
if command -v heracles &> /dev/null; then
    go test -v ./validation/external/... -run TestHeracles
fi
```

**Parser tests must handle:**
- Valid tool output
- Malformed output
- Empty output
- Partial results

**Version compatibility must be tested:**
- Document minimum and maximum tested versions
- Test with version pin from `external/README.md`
- Document version upgrade path

### Common Patterns

**Wrapping a shell tool in Go** (see `external/heracles_wrapper.go`):
```go
func RunHeracles(configPath string) (result HeracelesResult, err error) {
    cmd := exec.Command("heracles", "-config", configPath)
    output, err := cmd.CombinedOutput()
    if err != nil {
        // Tool not found or execution failed
        return HeracelesResult{}, fmt.Errorf("heracles failed: %w", err)
    }
    
    // Parse output
    result, err = parseHeracelesOutput(string(output))
    return result, err
}
```

**Checking if tool is installed** (in wrappers):
```bash
if ! command -v heracles &> /dev/null; then
    echo "Warning: Heracles not installed. Skipping comparison."
    exit 0
fi
```

**Timeout handling** (in Go wrappers):
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
cmd := exec.CommandContext(ctx, "tool", "args...")
```

**Capturing stderr separately** (in Go wrappers):
```go
var stdout, stderr bytes.Buffer
cmd.Stdout = &stdout
cmd.Stderr = &stderr
err := cmd.Run()
if err != nil {
    log.Printf("Tool error: %s\n", stderr.String())
}
```

## Dependencies

### Required (FeCIM runs without these)
- None—all external tools are optional

### Optional (Install as needed)

| Tool | Purpose | Version | Install |
|------|---------|---------|---------|
| Heracles | HZO compact model reference | v0.4.0 | `go build` from source |
| CrossSim | Crossbar array simulator | v3.1.0 | `pip install cross-sim==3.1.0` |
| ngspice | SPICE circuit simulator | 42 | `apt install ngspice` |
| Icarus Verilog | Verilog simulator | 12.0 | `apt install iverilog` |
| Verilator | Verilog lint/sim | 5.028 | `apt install verilator` |
| OpenROAD | Place-and-route | v2.0-2025.02 | Build from source |
| OpenLane2 | RTL→GDS flow | v2.1.2 | `pip install openlane==2.1.2` |
| Python scientific stack | Data analysis | numpy 2.1.3, scipy 1.14.1 | `pip install numpy scipy` |

### Internal
- `validation/` — Uses tool wrappers for comparison
- `shared/export/` — May export to SPICE format

## MANUAL

**Tool Inventory Reference:**
See `external/README.md` for complete table with:
- Tool name and purpose
- Version pin (explicit)
- License
- Installation method
- FeCIM use-case

**Installing Optional Tools (Examples):**

```bash
# Heracles (compile from source)
cd /tmp
git clone https://github.com/byu-vdl/heracles.git
cd heracles
go build -o heracles ./cmd
sudo mv heracles /usr/local/bin/

# CrossSim (Python package)
pip install cross-sim==3.1.0

# ngspice (system package)
sudo apt-get install ngspice

# Icarus Verilog
sudo apt-get install iverilog

# Verilator
sudo apt-get install verilator

# OpenROAD (complex; see https://github.com/The-OpenROAD-Project)
git clone https://github.com/The-OpenROAD-Project/OpenROAD.git
cd OpenROAD && cmake . && make -j$(nproc)

# OpenLane2 (Python package)
pip install openlane==2.1.2

# Python scientific stack
pip install numpy==2.1.3 scipy==1.14.1
```

**Heracles Integration:**
Heracles is the reference HZO compact model:
```go
// Load HZO calibration
cal := physics.LoadCalibration("data/calibrations/fecim_hzo.json")

// Compare with Heracles
heracelesResult, err := heracles_wrapper.RunHeracles("data/calibrations/fecim_hzo.json")
if err == nil {
    // Compute comparison metrics
    ks := validation.KSTest(simulated, heracelesResult.PolarizationCurve)
    if ks.PValue > 0.05 {
        log.Printf("✓ Matches Heracles (p=%.3f)", ks.PValue)
    }
}
```

**CrossSim Integration:**
CrossSim simulates crossbar array behavior at architecture level:
```go
// Export crossbar config
crossbarConfig := module2.GenerateConfig(128, 128)

// Run CrossSim
crosssimResult, err := crosssim_wrapper.RunCrossSim(crossbarConfig)
if err == nil {
    // Compare MVM throughput
    simThroughput := module2.MeasureThroughput(...)
    crosssimThroughput := crosssimResult.Throughput
    ratio := simThroughput / crosssimThroughput
    log.Printf("FeCIM/CrossSim throughput ratio: %.2f", ratio)
}
```

**SPICE Netlist Validation:**
Export peripheral circuits to SPICE and validate:
```go
// Export ADC circuit to SPICE
netlist, err := peripherals.ExportADCtoSPICE(adcConfig)
if err != nil {
    log.Fatal(err)
}

// Validate with ngspice
valid, err := spice_wrapper.ValidateSPICE(netlist)
if valid {
    log.Println("✓ SPICE netlist is valid")
}
```

**Verilog Syntax Checking:**
Validate generated Verilog before synthesis:
```bash
# Syntax check
iverilog -c data/fecim_crossbar_128x128.v -o /tmp/check.vvp

# Lint with Verilator
verilator --lint-only data/fecim_crossbar_128x128.v
```

**OpenROAD Place-and-Route:**
Generate layout for crossbar:
```bash
# Synthesize and P&R
openroad -python scripts/synthesis.tcl \
    -output data/fecim_crossbar_64x64_openroad.def
```

**OpenLane2 Full Flow:**
Automated RTL→GDS:
```bash
# Run OpenLane2 flow
openlane ./data/fecim_crossbar_config.tcl
# Outputs: GDS, DEF, reports in results/
```

**Tool Version Compatibility Matrix:**
(From `external/README.md`)
- Heracles v0.4.0: Tested with FeCIM v1.2+
- CrossSim v3.1.0: Tested with FeCIM v1.0+
- ngspice 42: Tested with all versions
- OpenLane2 v2.1.2: Tested with FeCIM v1.2+

**Checking Tool Availability in CI:**
```yaml
# In GitHub Actions workflow
- name: Install optional tools
  run: |
    pip install cross-sim==3.1.0 || echo "CrossSim optional"
    apt-get install -y ngspice || echo "ngspice optional"

- name: Run integration tests
  run: go test ./validation/external/... || echo "Tool tests skipped"
```

**Graceful Tool Degradation:**
All wrappers follow this pattern:
1. Check if tool is installed
2. If missing, return "tool not available" error (not fatal)
3. Caller decides whether to fail or skip test
4. FeCIM core always works without external tools

**Adding New Tool Integration:**
1. Write wrapper in `external/<tool>_wrapper.sh` or `.go`
2. Document in `external/README.md` table
3. Add optional integration test (can skip if tool missing)
4. Update CI/CD if tool should be installed
5. Document expected output format
6. Add timeout handling (prevents hanging)

**Tool Output Parsing:**
Wrappers expose parsed output:
```go
type HeracelesResult struct {
    LoopPoints       []Point
    RemantPol        float64
    CoerciveField    float64
    SaturationPol    float64
    ComparisonErr    float64
}
```

**Performance Profiling with External Tools:**
Use wrappers to profile FeCIM components:
```bash
# Profile L-K solver against Heracles
time ./heracles -material hzo > heracles_timing.txt
time go test -bench=Landau ./validation/... > fecim_timing.txt

# Compare timings
echo "Heracles: $(grep real heracles_timing.txt)"
echo "FeCIM: $(grep real fecim_timing.txt)"
```

**CI/CD Integration:**
Optional tool runs are wrapped:
```yaml
- name: Compare with external tools
  continue-on-error: true  # Don't fail CI if tools missing
  run: go test ./validation/external/...
```

