# Experimental Data Package

This directory contains literature-anchored calibration and validation datasets for FeCIM/HfO2 model validation.

## Directory Layout

- `hzo/pe-loops/` — P-E loop metrics (Pr, Ec, loop-shape points)
- `hzo/switching-time/` — field/pulse dependent switching kinetics
- `hzo/endurance/` — endurance-cycle retention of switching window
- `hzo/retention/` — retention vs time/temperature
- `hfo2/pe-loops/` — undoped/doped HfO2 family P-E references
- `crossbar/read-margin/` — experimentally anchored read-margin trends
- `crossbar/ir-drop/` — experimentally anchored IR-drop trends

## JSON Schema (matches `validation.LiteratureDataset`)

```json
{
  "reference": {
    "doi": "string (required)",
    "authors": "string (required)",
    "year": "integer (required)",
    "title": "string (required)",
    "journal": "string (required)",
    "figure": "string (recommended; figure source)",
    "table": "string (optional)",
    "validated_at": "RFC3339 timestamp (required by Go time.Time parser)"
  },
  "data_points": [
    {
      "value": "number (required)",
      "unit": "string (required)",
      "uncertainty": "number (required, 1-sigma or extraction uncertainty)",
      "conditions": {
        "<numeric_condition_key>": "number"
      }
    }
  ],
  "metadata": {
    "dataset_type": "string",
    "notes": "string",
    "extraction_method": "string",
    "source_path": "string"
  }
}
```

### Required Reference Fields

At minimum, each dataset must include:
- DOI
- authors
- year
- title
- journal
- at least one of: figure/table (prefer both when available)

## Data Extraction Methodology

1. Prefer values explicitly reported in paper text/tables.
2. If values are figure-derived, digitize from the cited panel and report extraction uncertainty.
3. Record waveform/material context in `conditions` (e.g., frequency_hz, thickness_nm, temperature_k, pulse_mv_cm).
4. Keep assumptions in `metadata.notes` and `metadata.extraction_method`.
5. Do not mix multiple papers in one JSON file.

## Unit Conventions

Use these units for validation interoperability:
- Electric field `E`: `MV/cm`
- Polarization `P`: `uC/cm2`
- Time: `ns`
- Frequency: `Hz`
- Temperature: `K`

## Uncertainty Reporting Guidance

- Use numeric `uncertainty` for every data point.
- If directly reported by paper, store that value.
- If estimated from figure extraction, include conservative uncertainty reflecting axis resolution + digitization noise.
- Suggested defaults when unstated:
  - P (digitized): ±1.0 to ±2.0 `uC/cm2`
  - E (digitized): ±0.03 to ±0.08 `MV/cm`
  - Time (pulse switching): ±5% to ±15% of value
- Keep uncertainty basis in metadata (`uncertainty_basis`).
