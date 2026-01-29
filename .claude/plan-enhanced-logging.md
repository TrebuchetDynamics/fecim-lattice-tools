# Enhanced Logging Implementation Plan

## HANDOFF: Planner -> TDD-Guide

### Objective
Increase logging capabilities across all modules to capture every input, calculation, and error for easier debugging. All logs should be written to the `logs/` folder.

### Current State Analysis
- **Existing infrastructure**: `shared/logging/logging.go` provides a solid foundation with:
  - Verbosity levels (Off, Info, Debug, Trace)
  - Thread-safe shared log file
  - Per-module prefixes
  - UI event logging (Button, ValueChange, SliderChange, etc.)

### What's Missing
1. **Input logging** - Function inputs not systematically logged
2. **Calculation logging** - Intermediate and final calculation results
3. **Error logging** - Errors need structured logging with context
4. **Module coverage** - Not all modules use logging consistently

### Implementation Plan

## Phase 1: Enhance Core Logging Infrastructure

### 1.1 Add New Logging Methods to `shared/logging/logging.go`

```go
// Calculation logs at TRACE level - for physics/math calculations
func (l *Logger) Calculation(funcName string, inputs map[string]interface{}, result interface{})

// Input logs at TRACE level - for function entry with parameters
func (l *Logger) Input(funcName string, params map[string]interface{})

// Output logs at TRACE level - for function return values
func (l *Logger) Output(funcName string, result interface{})

// Error logs at INFO level (always logged) - for errors with context
func (l *Logger) Error(err error, context string, details map[string]interface{})

// ErrorContext logs an error with operation context
func (l *Logger) ErrorContext(operation string, err error)
```

### 1.2 Add Structured Logging Helpers

```go
// LogFunc is a helper to log function entry/exit
func (l *Logger) LogFunc(funcName string, inputs map[string]interface{}) func(result interface{})

// Usage:
// defer l.LogFunc("MVM", map[string]interface{}{"rows": 8, "input": vector})(result)
```

## Phase 2: Add Logging to Each Module

### Module 1: Hysteresis (`module1-hysteresis/`)

Files to update:
- `pkg/ferroelectric/preisach.go` - Log Preisach model calculations
- `pkg/ferroelectric/preisach_advanced.go` - Log advanced physics
- `pkg/ferroelectric/material.go` - Log material property calculations
- `pkg/simulation/engine.go` - Log simulation steps
- `pkg/gui/gui.go` - Already has logger, enhance coverage

Key functions to instrument:
- `PreisachModel.Update(E float64)` - Log E field input, P result
- `PreisachModel.GetHysteresisLoop()` - Log loop generation
- `HZOMaterial.CoerciveVoltage()` - Log voltage calculations
- `Engine.Step()` - Log each simulation step

### Module 2: Crossbar (`module2-crossbar/`)

Files to update:
- `pkg/crossbar/array.go` - Log MVM/VMM operations, quantization
- `pkg/crossbar/nonidealities.go` - Log IR drop, sneak path analysis
- `pkg/crossbar/drift.go` - Log drift calculations
- `pkg/crossbar/irdrop.go` - Log IR drop simulations
- `pkg/crossbar/sneakpath.go` - Log sneak path analysis
- `pkg/gui/app.go` - Already has logger, enhance

Key functions to instrument:
- `Array.MVM(input)` - Log input vector, weight matrix, output
- `Array.ProgramWeight()` - Log weight programming
- `QuantizeTo30Levels()` - Log quantization
- `AnalyzeIRDrop()` - Log analysis parameters and results
- `AnalyzeSneakPaths()` - Log sneak path calculations

### Module 3: MNIST (`module3-mnist/`)

Files to update:
- `pkg/core/network.go` - Log inference operations
- `pkg/core/quantize.go` - Log quantization operations
- `pkg/gui/app.go` - Already has logger
- `pkg/gui/dualmode.go` - Already has logger

Key functions to instrument:
- `DualModeNetwork.Infer()` - Log input image, predictions
- `forwardFP()` / `forwardCIM()` - Log layer activations
- `QuantizeWeights()` - Log weight quantization stats

### Module 4: Circuits (`module4-circuits/`)

Files to update:
- `pkg/peripherals/dac.go` - Log DAC conversions
- `pkg/peripherals/adc.go` - Log ADC conversions
- `pkg/peripherals/tia.go` - Log TIA calculations
- `pkg/peripherals/chargepump.go` - Log charge pump operations
- `pkg/peripherals/analysis.go` - Log timing/power analysis
- `pkg/gui/app.go` - Add logger initialization

Key functions to instrument:
- `DAC.Convert()` - Log digital input, analog output
- `ADC.Convert()` - Log analog input, digital output
- `TIA.Convert()` - Log current to voltage conversion
- `ChargePump.ActualOutputVoltage()` - Log voltage generation
- `AnalyzeTiming()` - Log timing results
- `AnalyzePower()` - Log power breakdown

### Module 5: Comparison (`module5-comparison/`)

Files to update:
- `pkg/comparison/architecture.go` - Log comparison calculations
- `pkg/gui/app.go` - Already has logger

Key functions to instrument:
- `CompareArchitectures()` - Log comparison parameters
- `CalculateEfficiency()` - Log efficiency metrics
- `RunInference()` - Log benchmark results
- `ScaleToDataCenter()` - Log scaling calculations

### Module 6: EDA (`module6-eda/`)

Files to update:
- `pkg/compiler/compiler.go` - Log compilation steps
- `pkg/export/csv.go` - Log export operations
- `pkg/export/json.go` - Log export operations
- `pkg/export/spice.go` - Log SPICE generation
- Entry points already use global logging

Key functions to instrument:
- `Compile()` - Log compilation config and results
- `ExportCSV/JSON/SPICE()` - Log export operations

## Phase 3: Testing Strategy

### Test File Locations
Each module should have logging tests:
- `shared/logging/logging_test.go` - Core logging tests (exists, enhance)
- Create integration tests verifying logs are written

### Test Categories
1. **Unit tests** - Verify new logging methods work
2. **Integration tests** - Verify logs appear in files
3. **Verbosity tests** - Verify level filtering works

## Phase 4: Documentation

Update `docs/development/scriptReference.md`:
- Add logging function reference
- Add debugging with logs guide

## Files Changed Summary

### New/Modified in `shared/logging/`
- `logging.go` - Add Calculation, Input, Output, Error methods

### Module Files to Add Loggers
- ~25 files across 6 modules need logger initialization and calls

## Implementation Order

1. Enhance `shared/logging/logging.go` with new methods
2. Add tests for new methods
3. Instrument Module 2 (Crossbar) - core physics
4. Instrument Module 1 (Hysteresis) - physics model
5. Instrument Module 4 (Circuits) - peripherals
6. Instrument Module 3 (MNIST) - neural network
7. Instrument Module 5 (Comparison) - calculations
8. Instrument Module 6 (EDA) - compiler

## Open Questions
None - requirements are clear.

## Recommendations
- Set default verbosity to `VerbosityInfo` so important events are logged
- Add `--verbosity trace` flag documentation for deep debugging
- Consider adding JSON-formatted logs option for machine parsing (future)
