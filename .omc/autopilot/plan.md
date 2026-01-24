# FeCIM Module 5 Implementation Plan

## Phase 1: P0 Critical Data Fixes

### TODO 1.1: Replace Mythic AI with IBM Analog AI
**File**: <local-path>
**Line**: 201
**Change**: Replace Mythic AI competitor entry with IBM Analog AI

### TODO 1.2: Replace Mythic AI in liveslide.go  
**File**: <local-path>
**Line**: 412
**Change**: Update educational panel text from "Mythic AI: Not scalable" to "IBM Analog AI: Research only"

### TODO 1.3: Update market timeline struct
**File**: <local-path>
**Lines**: 18-31
**Change**: Update MarketSegment struct from Y2025/Y2030 to Y2024/Y2026/Y2030

### TODO 1.4: Update MarketOpportunityChart to use new fields
**File**: <local-path>
**Change**: Update all Y2025 references to Y2024, add Y2026 bar, update citation to 2026

## Phase 2: P1 Visual Differentiation

### TODO 2.1: Add IsEstimated field to Architecture struct
**File**: <local-path>
**Line**: After line 30
**Change**: Add `IsEstimated bool` field

### TODO 2.2: Mark FeCIM as estimated
**File**: <local-path>
**Line**: 107
**Change**: Add `IsEstimated: true` to FeCIMChip()

### TODO 2.3: Add amber color constant
**File**: <local-path>
**Line**: After imports
**Change**: Add `var estimatedColor = color.RGBA{255, 191, 0, 255}`

### TODO 2.4: Update FeCIM label with amber indicator
**File**: <local-path>
**Lines**: 128-136
**Change**: Add amber asterisk to FeCIM label

### TODO 2.5: Add legend for estimated indicator
**File**: <local-path>
**Lines**: 143-152
**Change**: Add "* = Estimated (TRL 4)" note in amber

### TODO 2.6: Update Competitor struct
**File**: <local-path>
**Lines**: 186-195
**Change**: Add `IsEstimated bool` field to Competitor struct

### TODO 2.7: Mark FeCIM/IBM as estimated in competitors
**File**: <local-path>
**Lines**: 197-202
**Change**: Set IsEstimated: true for FeCIM and IBM Analog AI

### TODO 2.8: Render estimated indicator in CompetitiveMatrix
**File**: <local-path>
**Lines**: 232-245
**Change**: Show amber energy values for estimated competitors

## Phase 3: P2 Add GUI Tests

### TODO 3.1: Create market_test.go
**File**: <local-path> (NEW)
**Tests**: TestMarketDataNoMythic, TestMarketDataHasIBM, TestMarketSegmentYears, TestCompetitorFeCIMHighlighted

### TODO 3.2: Create hero_test.go  
**File**: <local-path> (NEW)
**Tests**: TestEnergyConstants, TestEnergyRatios, TestNewAnimatedEnergyRace, TestEstimatedColorDefined

## Verification Checkpoints

After Phase 1:
- go build ./module5-comparison/... succeeds
- grep -r "Mythic" module5-comparison/ returns no matches

After Phase 2:
- go build ./module5-comparison/... succeeds
- Visual: amber indicators visible in UI

After Phase 3:
- go test ./module5-comparison/... passes (30+ tests)
- All new tests pass

## Success Criteria
1. go build ./... succeeds
2. go test ./... passes
3. No Mythic AI references
4. Market shows 2024/2026/2030
5. Amber indicators on estimated values

PLANNING_COMPLETE
