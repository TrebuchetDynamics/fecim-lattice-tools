# FeCIM Visualizer Improvement Specification

**Date:** 2026-01-24
**Focus:** Module 5 (Comparison) + App-wide improvements
**Status:** Ready for Implementation

---

## Requirements Analysis Summary

### Functional Requirements
1. Fix outdated competitor data (Mythic AI bankrupt 2023)
2. Update market timeline (Y2025 is now historical)
3. Improve visual differentiation for unverified values
4. Add test coverage for GUI components
5. Update citations to current year

### Non-Functional Requirements
1. Don't break existing functionality
2. All tests must pass before and after changes
3. Maintain 30 FPS animation performance
4. Keep investor-pitch credibility

### Implicit Requirements
1. Dr. Tour/Dr. Shin critique documentation
2. Screenshot evidence of improvements
3. Physics accuracy validation

---

## Critical Fixes (Prioritized)

| Priority | Issue | File | Action |
|----------|-------|------|--------|
| P0 | Mythic AI bankrupt | market.go:201 | Replace with active competitor |
| P0 | Y2025 outdated | market.go:21-30 | Update to Y2024/Y2026/Y2030 |
| P1 | Unverified values same color | hero.go, widgets.go | Add amber visual indicator |
| P1 | IsEstimated flag missing | architecture.go | Add field, propagate to UI |
| P2 | No GUI tests | pkg/gui/ | Add market_test.go, hero_test.go |
| P3 | Citation year wrong | market.go:129 | Update to 2026 |

---

## Files to Modify

module5-comparison/pkg/
├── comparison/
│   └── architecture.go      # Add IsEstimated field
└── gui/
    ├── market.go            # P0: Competitors + Market data
    ├── hero.go              # P1: Visual differentiation
    ├── widgets.go           # P1: Visual differentiation
    ├── liveslide.go         # Update Mythic reference
    ├── market_test.go       # NEW: Data validation tests
    └── hero_test.go         # NEW: Animation state tests

---

## Replacement Competitors for Mythic AI

Options (all active in 2026):
1. Syntiant - neuromorphic, commercial products
2. IBM Analog AI - research phase, similar CIM tech
3. Cerebras - wafer-scale, different segment but relevant

Recommended: IBM Analog AI (most similar technology approach)

---

## Implementation Order

1. P0: Data correctness (Mythic AI, market timeline)
2. P1: Visual differentiation for estimates
3. P2: Test coverage
4. P3: Polish (citations, performance)

---

## Verification Checklist

- [ ] go build succeeds
- [ ] go test ./... passes (26+ tests)
- [ ] App launches, all tabs render
- [ ] No Mythic AI references
- [ ] Market data shows 2024/2026/2030
- [ ] Unverified values visually distinct
- [ ] Screenshots captured before/after

---

EXPANSION_COMPLETE
