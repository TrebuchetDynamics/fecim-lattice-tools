## 2026-01-29: Relaxation Compensation Implementation

### Pattern: Level-Dependent Compensation Initialization
- Used parabolic profile for initial compensation values
- Formula: `0.05 * 4 * normalizedPos * (1 - normalizedPos)`
- Rationale: Middle levels experience more relaxation drift than edge levels
- Peak compensation: 5% at mid-level, tapering to 0% at edges

### Pattern: Compensation as Multiplicative Factor
- Applied as: `writeE = baseE * (1.0 + compensation)`
- Allows positive and negative compensation (bounds: -5% to +25%)
- More intuitive than additive compensation for E-field scaling

### Pattern: Symmetric Learning
- 1% adjustment per retry in error-reducing direction
- Ascending: error > 0 (overshot) → decrease, error < 0 (undershot) → increase
- Descending: error < 0 (went too far) → decrease, error > 0 (didn't go far enough) → increase
- Simple, robust, and converges quickly

### Pattern: Full Persistence Integration
- Added fields to both TempCalibration and CalibrationData structs
- Integrated with temperature interpolation system
- Compensation values interpolated between temperature calibrations
- JSON tags for backward compatibility

### Successful Approach
- Build-first verification: ensured code compiles before detailed testing
- Incremental implementation: data structures → initialization → application → learning → persistence
- Defensive bounds checking: prevents array access panics during compensation application

### Convention: Logging Format
- Used format: "RELAX_UP[idx]: old → new (err=...)"
- Only logs when compensation actually changes
- Consistent with existing "CALIB_UP[idx]" logging pattern
