# FeCIM Lattice Tools CLI Reference

> **Note:** This file was previously located at `docs/CLI.md`. It has moved to `docs/guides/cli-reference.md`.

This document describes the command-line interface for all FeCIM Lattice Tools modules.

## Top-level launcher flags (`cmd/fecim-lattice-tools`)

The primary binary exposes these flags:

| Flag               | Description                                                                                         | Default       |
| ------------------ | --------------------------------------------------------------------------------------------------- | ------------- |
| `--logger`         | Enable file logging (or pass level token)                                                           | `false`       |
| `--verbosity`      | Log level (`off\|info\|debug\|trace`) when logger is enabled                                        | `info`        |
| `--calibrate`      | Run hysteresis calibration and exit                                                                 | `false`       |
| `--material`       | Material to calibrate (`all` or explicit name)                                                      | `all`         |
| `--force`          | Force recalibration even if cache exists                                                            | `false`       |
| `--verify`         | Verify calibration quality after calibration                                                        | `false`       |
| `--list-materials` | Print available materials and exit                                                                  | `false`       |
| `--mode`           | Run headless mode and exit (for example `hysteresis`)                                               | empty         |
| `--engine`         | Headless hysteresis engine: `preisach` or `lk`                                                      | `preisach`    |
| `--screenshot-dir` | Output directory for GUI screenshots                                                                | `screenshots` |
| `--recording-dir`  | Output directory for GUI recordings                                                                 | `recordings`  |
| `--module`         | Initial module (`home`, `hysteresis`, `crossbar`, `mnist`, `circuits`, `comparison`, `eda`, `docs`) | `home`        |

Use `fecim-lattice-tools --help` for the canonical, runtime-generated list.

## Common CLI Flags

These flags are available only on supporting subcommands, not on every command. Check the specific subcommand's `--help` before use. The root launcher, installed `hysteresis` subcommand and workbench commands do not inherit this table.

| Flag                | Description                                              |
| ------------------- | -------------------------------------------------------- |
| `--json`            | Output results as JSON instead of human-readable text    |
| `-q, --quiet`       | Suppress informational output (only show results/errors) |
| `-c, --config FILE` | Load configuration from YAML or JSON file                |
| `--batch FILE`      | Process multiple items from a batch file                 |
| `-o, --output FILE` | Write output to file (default: stdout)                   |
| `-h, --help`        | Show help message                                        |

## Configuration Files

Only subcommands that advertise `--config` in their help accept configuration files. Accepted formats and keys depend on that subcommand. The root launcher, installed `hysteresis` subcommand and workbench commands do not support this common configuration-file interface.

### Example JSON Config (mnist.json)

```json
{
  "hidden_size": 128,
  "noise_level": 0.02,
  "epochs": 10,
  "levels": [8, 16, 24, 31]
}
```

## Batch Processing

Only subcommands that advertise a batch-file option accept batch files. The root launcher, installed `hysteresis` subcommand and workbench commands do not support `--batch FILE`. Other uses of `--batch`, such as an inference batch size, do not imply file input.

The following are illustrative file formats, not a Hysteresis material-batch interface. Consult the supporting subcommand's help for its accepted format:

1. **Line-based** (one item per line, `#` for comments):

```
superlattice
fecim
cryogenic
# This is a comment
hzo32
```

2. **JSON array**:

```json
["superlattice", "fecim", "cryogenic", "hzo32"]
```

---

## Module 1: Hysteresis

Visualize ferroelectric hysteresis or run headless simulation diagnostics. Material presets and discrete levels are educational simulation parameters, not validated device measurements.

### Basic Usage

```bash
fecim-lattice-tools hysteresis [options]
```

Without options (or with `hysteresis gui`), the installed binary opens the legacy Fyne Hysteresis view. Desktop startup requires a CGO-enabled build. Help, material listing and headless diagnostics also work with `CGO_ENABLED=0`.

### Options

| Flag               | Description                                  | Default  |
| ------------------ | -------------------------------------------- | -------- |
| `--engine NAME`    | Headless engine: `preisach` or `lk`            | preisach |
| `--headless`       | Run headless hysteresis diagnostics           | false    |
| `--list-materials` | Print material names as plain text and exit    | false    |
| `-h, --help`       | Show help and exit                            | -        |

These are the installed subcommand's options. It has no material, frequency, TUI, Vulkan, config, batch, JSON or output-file switches. `--engine` selects the headless engine, not the GUI engine.

Source: `cmd/fecim-lattice-tools/subcommands.go::runHysteresisSubcommand`, `main.go::runRoot` and `mode.go::runHysteresisMode` in the same directory.

### Help and Available Materials

Show the runtime-generated help:

```bash
fecim-lattice-tools hysteresis --help
```

List material names:

```bash
fecim-lattice-tools hysteresis --list-materials
```

The listing is plain text, not JSON or a table of material properties:

```text
Available materials:
  - HZO (Si-doped, Park 2015 midpoint)
  - FeCIM HZO
  - FeCIM HZO (TARGET)
  - Literature Superlattice
  - Cryogenic HZO
  - HZO Standard (32 states)
  - HZO FTJ (140 states)
  - AlScN (8-16 states)
  - α-In₂Se₃
```

### Headless Examples

Run diagnostics with the default Preisach engine or opt into Landau-Khalatnikov (L-K):

```bash
fecim-lattice-tools hysteresis --headless
fecim-lattice-tools hysteresis --headless --engine lk
```

Headless runs do not produce static ASCII plots or JSON material data. Console diagnostic logging is off by default. Runs attempt to write Hysteresis CSV logs separately (see retention below).

For headless runs, optional `FECIM_MATERIAL` selects a preset. When unset, the default is `FeCIMMaterial`. For example:

```bash
FECIM_MATERIAL=superlattice fecim-lattice-tools hysteresis --headless --engine preisach
```

The root launcher's `--material` selects a material for `--calibrate` only. It does not select the material for headless simulation.

For plain-text console diagnostics and file logging, use `fecim-lattice-tools --logger --mode hysteresis --engine lk`. This root invocation uses the same headless runner and `FECIM_MATERIAL` selection.

---

## Module 2: Crossbar

Crossbar array matrix-vector multiplication (MVM) visualization.

### Basic Usage

```bash
fecim-lattice-tools crossbar inference [options]
```

### Options

| Flag                   | Description              | Default |
| ---------------------- | ------------------------ | ------- |
| `--size N`             | Array size (NxN)         | 64      |
| `--layers N`           | Neural network layers    | 3       |
| `--batch N`            | Inference batch size     | 1       |
| `--noise LEVEL`        | Device noise level (0-1) | 0.02    |
| `--adc BITS`           | ADC resolution           | 6       |
| `--seed N`             | Random seed (0=time)     | 1       |
| `--no-color`           | Disable colored output   | false   |
| `--benchmark`          | Run benchmark            | false   |
| `--show-array`         | Show array state         | false   |
| `--show-mvm`           | Show MVM operation       | false   |
| `--show-irdrop`        | Show IR drop analysis    | false   |
| `--show-sneak`         | Show sneak path analysis | false   |
| `--show-nonidealities` | Show all effects         | false   |

### Examples

Run MVM with JSON output:

```bash
fecim-lattice-tools crossbar inference --show-mvm --json --size 32
```

Using config file:

```bash
fecim-lattice-tools crossbar inference --config crossbar.yaml
```

---

## Module 3: MNIST

MNIST digit recognition on ferroelectric crossbar arrays.

### Basic Usage

```bash
fecim-lattice-tools mnist cli [options]
```

### Options

| Flag                   | Description                      | Default |
| ---------------------- | -------------------------------- | ------- |
| `--train`              | Train the network                | false   |
| `--evaluate`           | Evaluate on test set             | false   |
| `--interactive`        | Interactive mode                 | false   |
| `--epochs N`           | Training epochs                  | 5       |
| `--hidden N`           | Hidden layer size                | 128     |
| `--noise LEVEL`        | Device noise (0-1)               | 0.02    |
| `--load FILE`          | Load weights from file           | -       |
| `--save FILE`          | Save weights to file             | -       |
| `--core-eval`          | Dual-mode evaluation (FP vs CIM) | false   |
| `--core-samples N`     | Samples for core-eval            | 1000    |
| `--core-levels LIST`   | Comma-separated levels           | -       |
| `--export-levels LIST` | Export quantized weights         | -       |

### Examples

Evaluate with JSON output:

```bash
fecim-lattice-tools mnist cli --evaluate --json --load weights.json
```

Core evaluation with level sweep:

```bash
fecim-lattice-tools mnist cli --core-eval --core-levels 8,16,24,31 --json
```

Using config file:

```bash
fecim-lattice-tools mnist cli --config mnist.yaml --evaluate
```

---

## Module 4: Circuits

Peripheral circuits visualization (DAC, ADC, TIA, Charge Pump).

### Basic Usage

```bash
fecim-lattice-tools circuits cli [options]
```

### Options

| Flag            | Description              | Default |
| --------------- | ------------------------ | ------- |
| `--dac`         | Show DAC details         | false   |
| `--adc`         | Show ADC details         | false   |
| `--tia`         | Show TIA details         | false   |
| `--pump`        | Show Charge Pump details | false   |
| `--all`         | Show all circuits        | false   |
| `--linearity`   | Show INL/DNL analysis    | false   |
| `--timing`      | Show timing diagrams     | false   |
| `--power`       | Show power breakdown     | false   |
| `--ispp`        | Run ISPP demo            | false   |
| `--level N`     | Demo level (0-29)        | 15      |
| `--logger`      | Enable file logging      | false   |
| `--verbosity N` | Log verbosity (0-3)      | 2       |

### Examples

Get all circuit specs as JSON:

```bash
fecim-lattice-tools circuits cli --all --json
```

Show specific circuits with JSON output:

```bash
fecim-lattice-tools circuits cli --dac --adc --json
```

Using config file:

```bash
fecim-lattice-tools circuits cli --config circuits.yaml --all
```

---

## Module 5: Comparison

Architecture comparison (CPU vs GPU vs FeCIM).

### Basic Usage

```bash
fecim-lattice-tools comparison cli [options]
```

### Options

| Flag              | Description                 | Default |
| ----------------- | --------------------------- | ------- |
| `--all`           | Show all comparisons        | true    |
| `--specs`         | Show architecture specs     | false   |
| `--inference`     | Show inference comparison   | false   |
| `--datacenter`    | Show datacenter comparison  | false   |
| `--advantages`    | Show FeCIM advantages       | false   |
| `--workload NAME` | Workload type               | mnist   |
| `--throughput N`  | Target throughput (inf/sec) | 10000   |
| `--no-color`      | Disable colored output      | false   |

### Workloads

- `mnist` - MNIST digit recognition
- `resnet` - ResNet-50 image classification
- `bert` - BERT-base NLP
- `gpt2` - GPT-2 language model
- `llm` - Large language model

### Examples

Compare architectures with JSON output:

```bash
fecim-lattice-tools comparison cli --all --json --workload resnet
```

Specific comparisons:

```bash
fecim-lattice-tools comparison cli --specs --advantages --json
```

Using config file:

```bash
fecim-lattice-tools comparison cli --config comparison.yaml
```

---

## Module 6: EDA

EDA design generation and export.

### Basic Usage

```bash
fecim-lattice-tools eda cli [options]
```

### Options

| Flag            | Description                             | Default     |
| --------------- | --------------------------------------- | ----------- |
| `--mode MODE`   | Operation mode (storage/memory/compute) | compute     |
| `--input FILE`  | Input weights JSON                      | -           |
| `--output DIR`  | Output directory                        | data        |
| `--name NAME`   | Design name                             | fecim_array |
| `--rows N`      | Array rows                              | 128         |
| `--cols N`      | Array columns                           | 128         |
| `--levels N`    | Conductance levels (2-30)               | 30          |
| `--tech NAME`   | Technology (SKY130/GF180MCU/IHP_SG13G2) | SKY130      |
| `--arch TYPE`   | Architecture (passive/1T1R)             | passive     |
| `--vdd V`       | Supply voltage                          | 1.8         |
| `--gmin US`     | Min conductance (μS)                    | 10.0        |
| `--gmax US`     | Max conductance (μS)                    | 100.0       |
| `--export-json` | Export JSON mapping                     | true        |
| `--csv`         | Export CSV                              | true        |
| `--spice`       | Export SPICE netlist                    | true        |
| `--verilog`     | Export Verilog                          | true        |
| `--def`         | Export DEF placement                    | true        |
| `--json-output` | Output results as JSON                  | false       |
| `--quiet`       | Suppress output                         | false       |
| `--config FILE` | Config file                             | -           |

### Examples

Generate design with JSON result:

```bash
fecim-lattice-tools eda cli --rows 64 --cols 64 --json-output
```

Compute mode with weights:

```bash
fecim-lattice-tools eda cli --mode compute --input weights.json --json-output
```

Using config file:

```bash
fecim-lattice-tools eda cli --config eda.yaml --json-output
```

---

## JSON Output Examples

### Hysteresis: No JSON Output

The installed `hysteresis` subcommand has no JSON output option or material-info JSON schema. Use `hysteresis --list-materials` for plain-text names, not machine-readable material properties.

### Circuits Result

```json
{
  "dac": {
    "bits": 5,
    "levels": 30,
    "vref_low_v": -1.5,
    "vref_high_v": 1.5,
    "resolution_v": 0.1,
    "energy_fj": 10.5
  },
  "adc": {
    "bits": 5,
    "levels": 30,
    "conversion_time_ns": 100,
    "enob": 4.5,
    "energy_fj": 15.2
  }
}
```

### Comparison Result

```json
{
  "workload": "mnist",
  "target_throughput": 10000,
  "architectures": [
    {
      "name": "CPU+DRAM",
      "tdp_watts": 65,
      "tops_per_watt": 0.1,
      "latency_ms": 5.0,
      "energy_mj": 10.0,
      "throughput_infs": 200
    }
  ],
  "advantages": {
    "vs_cpu_energy_reduction": 100,
    "vs_cpu_cost_reduction": 10,
    "vs_gpu_power_reduction": 50,
    "vs_gpu_area_reduction": 20
  }
}
```

### EDA Result

```json
{
  "design_name": "fecim_array",
  "mode": "compute",
  "rows": 128,
  "cols": 128,
  "total_cells": 16384,
  "active_cells": 16384,
  "area_mm2": 0.25,
  "power_mw": 10.5,
  "throughput_gops": 100.0,
  "technology": "SKY130",
  "output_files": [
    "data/fecim_array_design.json",
    "data/fecim_array_cells.csv",
    "data/fecim_array.sp",
    "data/fecim_array.v",
    "data/fecim_array.def"
  ]
}
```

---

## Scripting Examples

### Shell Loop for Headless Diagnostics

Use a shell loop and redirection, not Hysteresis batch or output-file flags. Each root invocation below enables logging and runs diagnostics for one preset. Text capture does not redirect the separate CSV or diagnostic log files.

```bash
#!/bin/bash
mkdir -p results
materials=(superlattice fecim cryogenic hzo32 ftj140 alscn)

for mat in "${materials[@]}"; do
    echo "Processing $mat..."
    FECIM_MATERIAL="$mat" fecim-lattice-tools --logger --mode hysteresis \
        > "results/${mat}.txt" 2>&1
done
```

### JSON Processing with jq

Use `jq` only with commands that produce JSON. Hysteresis help, material listing and diagnostic output are plain text.

```bash
# Extract specific field from circuits
fecim-lattice-tools circuits cli --all --json | jq '.dac.energy_fj'

# Filter comparison results
fecim-lattice-tools comparison cli --json | \
    jq '.architectures[] | select(.name | contains("FeCIM"))'
```

### Pipeline Example

```bash
# Generate design, validate, and report
fecim-lattice-tools eda cli --rows 64 --cols 64 --json-output | \
    jq '{cells: .total_cells, area: .area_mm2, power: .power_mw}'
```

---

## Research and Design Workbench

Use the checked-in project bundle or initialize a new one:

```bash
fecim-lattice-tools project init ./my-study
fecim-lattice-tools project validate ./my-study
fecim-lattice-tools experiment run ./my-study -workers 2
fecim-lattice-tools report generate ./my-study
```

Export one successful immutable run without rerunning it:

```bash
fecim-lattice-tools export ./my-study -run <64-character-run-id>
```

Exports are written to `./my-study/exports/<run-id>/`; the manifest records hashes and the EDA technology assumptions. Existing exports are not overwritten. See the [workbench guide](research-design-workbench.md).

## Exit Codes

| Code | Meaning           |
| ---- | ----------------- |
| 0    | Success           |
| 1    | General error     |
| 2    | Invalid arguments |

---

## Hysteresis CSV log retention

New `hysteresis-*.csv` data logs are subject to conservative retention: 180 days, 100 files, and 2 GiB total, pruning at most once per day. Active logs and files newer than one day are not removed. On first use, a baseline marker protects every existing hysteresis CSV in that directory; this deliberately preserves pre-upgrade user data, even when it exceeds the limits. Other application logs are not affected.

## Environment Variables

| Variable           | Description                           |
| ------------------ | ------------------------------------- |
| `FECIM_LOG_LEVEL`  | Logging level (debug/info/warn/error) |
| `FECIM_CONFIG_DIR` | Default config directory              |

---

## See Also

- [EDA Tools Guide](../modules/eda/tools.md) - OpenLane integration
- [API Documentation](../internals/api-reference.md) - Programmatic access
- [Module 6 EDA README](../modules/eda/README.md) - EDA module overview
