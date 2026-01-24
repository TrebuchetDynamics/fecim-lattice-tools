# Module 6: FeCIM Array Builder for OpenLane

**Purpose:** Generate OpenLane-compatible EDA files for FeCIM crossbar arrays
**Status:** Educational/Research Tool (Work In Progress)
**Last Updated:** 2026-01-24

---

## What This Module Actually Does

Module 6 is an **array builder** that generates EDA file formats compatible with the open-source OpenLane RTL-to-GDSII flow. It does NOT:
- Compile neural network weights (that's conceptual, not implemented)
- Generate validated FeFET device models
- Produce fabrication-ready designs

### Capabilities (Implemented)

| Tab | Function | Output |
|-----|----------|--------|
| 1. Cell Builder | Define FeCIM bitcell dimensions | LEF, Liberty (.lib), Verilog |
| 2. Array Builder | Configure crossbar array size | Array parameters |
| 3. Verilog Export | Generate array netlist | Verilog module |
| 4. DEF Export | Generate placement file | DEF with cell instances |
| 5. Validation | Syntax checking | Yosys validation results |
| 6. Learn | OpenLane tutorial | Educational content |
| 7. Export All | Batch export | All files for OpenLane |

### What Gets Generated

```
output/
├── fecim_bitcell.lef       # Cell abstract (dimensions, pins)
├── fecim_bitcell.lib       # Timing library (PLACEHOLDER VALUES)
├── fecim_bitcell.v         # Behavioral Verilog (pass-through only)
├── fecim_array_NxM.v       # Array instantiation
├── fecim_array_NxM.def     # Placement definition
└── config.json             # OpenLane configuration
```

---

## Critical Disclaimers

### 1. Placeholder Timing Values

All Liberty (.lib) timing parameters are **placeholders**, not characterized values:

```
rise_time: 0.1 ns      ← PLACEHOLDER (not from simulation)
fall_time: 0.1 ns      ← PLACEHOLDER
input_cap: 0.002 pF    ← PLACEHOLDER
leakage:   0.001 nW    ← PLACEHOLDER
```

**Real FeFET characterization requires:**
- SPICE simulation with validated Verilog-A models [[1]](#ref1)
- Silicon measurements from test chips
- Liberty characterization flow (e.g., Liberate, OpenSTA)

### 2. Behavioral Model Limitations

The generated Verilog is a **pass-through model** only:

```verilog
// Generated code - does NOT model FeFET physics
module fecim_bitcell (input WL, input BL, output Q);
  assign Q = WL & BL;  // Simplified logic, NOT real behavior
endmodule
```

**What it doesn't model:**
- Polarization states (Pr, Ps)
- Hysteresis (Preisach model)
- 30-level analog states
- Retention, endurance, drift

### 3. No Physical Layout

LEF defines an **abstract view** only (bounding box + pins). There is no:
- Actual transistor layout
- DRC-clean geometry
- LVS-verifiable netlist

**Real FeFET layout requires:**
- Magic VLSI or KLayout for physical design [[2]](#ref2)
- Custom FeFET process layers (not in standard PDKs)
- DRC/LVS verification with foundry rules

### 4. Cell Dimensions

Default dimensions based on SKY130 standard cells [[3]](#ref3):

| Parameter | Value | Source |
|-----------|-------|--------|
| Cell Width | 0.46 μm | SKY130 unithd site width |
| Cell Height | 2.72 μm | SKY130 standard cell height |
| Site Name | unithd | SKY130 LEF specification |

**Note:** Actual FeFET cells may require different dimensions depending on device structure and process.

---

## How OpenLane Integration Works

### The Goal

Treat a FeCIM crossbar array as a "hard macro" that can be placed alongside standard digital logic in an OpenLane flow.

### The Reality

1. **Standard digital logic** (controllers, I/O) → Synthesized by Yosys, placed by OpenROAD
2. **FeCIM array** → Manually instantiated as a macro (our generated files)
3. **Interface** → Connected via standard cell wrappers

### Required Files for OpenLane

```tcl
# config.tcl for OpenLane
set ::env(PDK) "sky130A"
set ::env(EXTRA_LEFS) "$::env(DESIGN_DIR)/lef/fecim_bitcell.lef"
set ::env(EXTRA_LIBS) "$::env(DESIGN_DIR)/lib/fecim_bitcell.lib"
set ::env(FP_DEF_TEMPLATE) "$::env(DESIGN_DIR)/def/fecim_array.def"
```

**References:** OpenLane v2.0 configuration [[4]](#ref4)

---

## Verified File Format Compliance

### LEF 5.8 Compliance

Generated LEF follows the LEF/DEF 5.8 specification [[5]](#ref5):

```lef
VERSION 5.8 ;
MACRO fecim_bitcell
  CLASS CORE ;
  ORIGIN 0 0 ;
  SIZE 0.460 BY 2.720 ;
  PIN WL
    DIRECTION INPUT ;
    PORT LAYER met1 ; RECT 0 1.2 0.14 1.48 ; END
  END WL
  ...
END fecim_bitcell
```

### Liberty Compliance

Generated Liberty follows Synopsys Liberty format [[6]](#ref6):

```liberty
library(fecim_bitcell_lib) {
  cell(fecim_bitcell) {
    pin(WL) { direction : input; capacitance : 0.002; }
    /* PLACEHOLDER TIMING - NOT CHARACTERIZED */
  }
}
```

### DEF Compliance

Generated DEF follows DEF 5.8 specification [[5]](#ref5):

```def
VERSION 5.8 ;
DESIGN fecim_array_4x4 ;
UNITS DISTANCE MICRONS 1000 ;
COMPONENTS 16 ;
  - cell_0_0 fecim_bitcell + PLACED ( 0 0 ) N ;
  ...
END COMPONENTS
```

---

## What Would Make This Production-Ready

### Required for Real Fabrication

| Requirement | Current State | What's Needed |
|-------------|---------------|---------------|
| FeFET device model | None | Verilog-A with validated parameters |
| Timing characterization | Placeholder | SPICE simulation + Liberty flow |
| Physical layout | Abstract only | Magic/KLayout with DRC-clean geometry |
| Process support | Standard CMOS | FeFET-capable PDK or post-processing |
| Verification | Syntax only | Full DRC/LVS/timing sign-off |

### Closest Path to Silicon

1. **IHP SG13G2 PDK** - Has RRAM support, closest to CIM [[7]](#ref7)
2. **Custom Verilog-A models** - Preisach-based FeFET simulation [[1]](#ref1)
3. **Post-CMOS processing** - Deposit ferroelectric layers on finished CMOS

---

## References

<a name="ref1"></a>
**[1] FeFET Compact Modeling**
- "Temperature and Variability-Aware FeFET Model," *Solid-State Electronics*, 2024
- Preisach-based Verilog-A model with temperature and variability effects
- [DOI](https://www.sciencedirect.com/science/article/abs/pii/S0038110124001035)

<a name="ref2"></a>
**[2] Open-Source Layout Tools**
- Magic VLSI: [opencircuitdesign.com/magic](http://opencircuitdesign.com/magic/)
- KLayout: [klayout.de](https://www.klayout.de/)
- GDSFactory: [gdsfactory.github.io](https://gdsfactory.github.io/gdsfactory/)

<a name="ref3"></a>
**[3] SkyWater SKY130 PDK**
- Standard cell library specifications
- Cell height: 2.72 μm, site width: 0.46 μm
- [skywater-pdk.readthedocs.io](https://skywater-pdk.readthedocs.io/)
- [github.com/google/skywater-pdk](https://github.com/google/skywater-pdk)

<a name="ref4"></a>
**[4] OpenLane Documentation**
- "OpenLANE: The Open-Source Digital ASIC Implementation Flow," WOSET 2020
- [openlane.readthedocs.io](https://openlane.readthedocs.io/)
- [Paper PDF](https://woset-workshop.github.io/PDFs/2020/a21.pdf)

<a name="ref5"></a>
**[5] LEF/DEF 5.8 Specification**
- Si2/OpenAccess Coalition standard
- Library Exchange Format and Design Exchange Format
- [si2.org](https://si2.org/)

<a name="ref6"></a>
**[6] Liberty Timing Format**
- Synopsys standard for timing characterization
- Industry-standard cell library format
- [Synopsys Liberty Documentation](https://www.synopsys.com/)

<a name="ref7"></a>
**[7] IHP Open Source PDK**
- 130nm BiCMOS process with RRAM support
- OpenROAD flow supported
- [github.com/IHP-GmbH/IHP-Open-PDK](https://github.com/IHP-GmbH/IHP-Open-PDK)

---

## Related Documentation

| Document | Purpose |
|----------|---------|
| [REFERENCES.md](./REFERENCES.md) | Full scientific reference list |
| [SKY130.md](./SKY130.md) | SKY130 PDK integration guide |
| [eda.opensource.md](./eda.opensource.md) | Open-source EDA ecosystem overview |
| [eda.research.md](./eda.research.md) | Research paper collection |
| [plan-demo6.md](./plan-demo6.md) | Implementation plan with disclaimers |

---

## Summary

**Module 6 is an educational tool** that demonstrates how FeCIM arrays could integrate with open-source EDA flows. It generates syntactically valid files but:

- Uses **placeholder** timing values
- Provides **abstract** cell representations only
- Does **not** model actual FeFET physics
- Is **not** validated for fabrication

For production use, consult the references above and work with foundry partners who support ferroelectric processes.

---

*This document aims to be honest about capabilities and limitations. All claims are backed by references.*
