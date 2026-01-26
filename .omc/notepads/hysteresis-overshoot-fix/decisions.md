# Hysteresis Overshoot Fix - Design Decisions

## Decision 1: Incremental Pulse Strength Levels
**Chosen**: Two-tier pulse strength based on gap to target
- Far from target (gap > 5): Ec × 1.3
- Close to target (gap ≤ 5): Ec × 1.1

**Rationale**:
- Ec × 1.3 provides faster convergence when far from target
- Ec × 1.1 provides precise control near target to avoid overshoot
- Both values are just above Ec threshold, ensuring switching occurs but at controlled rate

**Alternatives Considered**:
- Continuous proportional control: More complex, may still overshoot
- Single fixed pulse (Ec × 1.1): Too slow when far from target
- Three+ tiers: Added complexity without significant benefit

## Decision 2: Immediate Transition on Target Reached
**Chosen**: Transition to HOLD phase as soon as `abs(currentLevel - targetLevel) <= 1`

**Rationale**:
- Prevents overshoot from continued pulsing after target reached
- ±1 level tolerance (3.3% error) is acceptable for 30-level analog memory
- Mimics real ferroelectric memory "program-and-verify" behavior

**Alternatives Considered**:
- Wait for field stability + time duration: Risks overshoot, slower
- Exact level match (error = 0): Too strict for analog memory, may never converge
- ±2 level tolerance: Too loose, reduces effective bit density

## Decision 3: Continuous Feedback During WRITE Phase
**Chosen**: Recalculate writeE every simulation frame based on current level

**Rationale**:
- Allows dynamic adjustment as level approaches target
- Reduces pulse strength automatically when close to target
- Essential for avoiding overshoot in nonlinear switching regime

**Alternatives Considered**:
- Fixed field throughout WRITE phase: Original approach, caused overshoot
- Periodic updates (e.g., every 100ms): Misses rapid level changes
- Only check at end of WRITE: Too late to prevent overshoot

## Decision 4: Apply Same Fix to Both Modes
**Chosen**: Use incremental pulse approach in both Manual and WriteReadDemo modes

**Rationale**:
- Consistent behavior across UI
- Both modes suffered from same overshoot problem
- Code reuse and maintainability

## Decision 5: Gap-Based Pulse Selection
**Chosen**: Use `gap = abs(targetLevel - currentLevel)` to determine pulse strength

**Rationale**:
- Simple, intuitive metric
- Works symmetrically for both UP and DOWN transitions
- Easy to tune (threshold of 5 levels is ~17% of range)

**Alternatives Considered**:
- Percentage-based gap: More complex calculation, no clear benefit
- Distance in polarization units: Less intuitive for level-based UI
- Time-based ramping: Doesn't account for switching dynamics

## Implementation Notes
- Changes are backwards-compatible: No API changes
- Minimal performance impact: Simple arithmetic per frame
- Self-documenting: Variable names (gap, writeE) clearly indicate purpose
- Thread-safe: All updates within existing mutex locks
