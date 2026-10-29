package crossbar

import (
	"fecim-lattice-tools/workbench/moduleapi"
	"fmt"
	"math"
	"math/rand"
)

// The native stack tab is an illustrative layer-index gradient with +/-0.15
// noise, not an inter-layer electrical solver. Preserve that distinction and
// generate only the selected layer from its own seed: 512 layers never allocate
// 512 arrays. All cell values and labels are Go-owned, keyboard-readable data.
func (p *Port) stackTable() moduleapi.TablePage {
	rng := rand.New(rand.NewSource(p.cfg.Seed + int64(p.layer)))
	base := float64(p.layer) / math.Max(float64(p.layers-1), 1)
	t := moduleapi.TablePage{ID: "stack-cells", Caption: fmt.Sprintf("Illustrative stack layer %d of %d — page %d; seeded gradient, not measured or electrically coupled", p.layer, p.layers, p.page), Page: p.page, Total: (p.cfg.Rows*p.cfg.Cols + 31) / 32, Columns: []string{"Row", "Column", "Illustrative normalized state"}, Rows: [][]moduleapi.NumericCell{}}
	for index := 0; index < p.cfg.Rows*p.cfg.Cols && index < p.page*32; index++ {
		v := math.Max(0, math.Min(1, base+(rng.Float64()-0.5)*0.3))
		if index < (p.page-1)*32 {
			continue
		}
		t.Rows = append(t.Rows, []moduleapi.NumericCell{cell(float64(index/p.cfg.Cols), "count"), cell(float64(index%p.cfg.Cols), "count"), cell(v, "ratio")})
	}
	return t
}
