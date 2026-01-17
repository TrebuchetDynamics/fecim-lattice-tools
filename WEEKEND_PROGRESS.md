# IronLattice Weekend Progress Log

## Session Started: 2026-01-17

### Current Status Assessment

**Demo 1 (Hysteresis):** Physics complete, rendering pipeline needs implementation
**Demo 2 (Crossbar MVM):** Structure exists but main.go has wrong import paths
**Demo 3 (Phase Field):** Only README.md and PHYSICS.md exist, no code structure

---

## Progress Log

### Entry 1: Initial Assessment

**Completed:**
- Read through all demo README.md files
- Reviewed existing Go code structure
- Identified physics models already implemented (Preisach model, materials)
- Reviewed existing shader files (hysteresis.vert/frag, mvm.comp)

**Issues Found:**
1. Demo 2 main.go uses wrong import paths (`github.com/ironlattice/vis/demo2-inference/pkg/...`)
2. No go.sum file - dependencies not resolved
3. Demo 3 has no code structure at all

**Next Steps:**
1. Fix import paths in demo2-crossbar/cmd/inference/main.go
2. Add required dependencies to go.mod
3. Create Demo 3 directory structure and TDGL solver scaffold

---
