# Hysteresis Overshoot Fix - Learnings

## Problem Identified
The calibrated field approach (applying large E-fields proportional to target level) caused overshoot because ferroelectric switching is highly nonlinear. Once E > Ec, multiple hysterons switch simultaneously, causing the polarization level to jump past the target.

## Solution: Incremental Pulse Write with Feedback
Implemented "program-and-verify" approach used in real ferroelectric memory:

### Key Algorithm Changes
1. **Small Incremental Pulses**: Instead of one large calibrated field, apply pulses just above Ec
   - Far from target (gap > 5 levels): E = Ec × 1.3
   - Close to target (gap ≤ 5 levels): E = Ec × 1.1
   - At target (gap ≤ 0): E = 0

2. **Continuous Feedback Loop**: Check current level during WRITE phase and adjust pulse strength dynamically
   ```go
   gap := targetLevel - currentLevel
   if gap <= 0 {
       writeE = 0  // Stop immediately when reached
   } else if gap <= 5 {
       writeE = Ec * 1.1  // Reduce pulse strength
   }
   ```

3. **Immediate Transition**: Move to HOLD phase as soon as within ±1 level of target
   - No time-based requirements
   - React immediately to level convergence
   - Prevents overshoot from continued pulsing

### Physics Principle
Real ferroelectric memory programming uses iterative "write-verify" cycles:
1. Apply small write pulse
2. Verify current level
3. Apply another pulse if needed
4. Stop when target reached

This avoids overshoot because you stop as soon as the target is reached, rather than blindly applying a large field that switches too many domains.

## Files Modified
- `<local-path>`
  - Manual mode (lines ~43-104): Incremental pulse approach for click-to-level
  - WriteReadDemo mode (lines ~157-192): Same approach for automated write/read cycles

## Testing
- Build: ✅ Success
- Unit tests: ✅ All pass (10 tests in module1-hysteresis)
- No regressions introduced

## Technical Insights
1. **Nonlinear Switching**: Ferroelectric switching is NOT proportional to field strength above Ec. Small changes in E above Ec can cause large jumps in P.

2. **Domain Dynamics**: The Preisach model represents distribution of switching thresholds (hysterons). When E exceeds many hysteron thresholds simultaneously, they all switch at once.

3. **Analog Precision**: For 30-level analog memory, ±1 level tolerance (3.3% error) is acceptable for practical use.

4. **Feedback Control**: Essential for reliable analog state programming. Open-loop (calibrated field) fails due to nonlinearity; closed-loop (feedback) succeeds.

## Next Steps
1. Test GUI behavior with different target level sequences
2. Verify improvement in success rate for WriteReadDemo mode
3. Consider adding configurable pulse strength parameters for fine-tuning
4. Document in main project docs if improvement confirmed by user
