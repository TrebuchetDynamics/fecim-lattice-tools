package crossbar

import core "fecim-lattice-tools/shared/crossbar"

// EffectiveConductanceMatrix is normalized in the native core. Physical table
// cells must use the physical mapping API, including variation, in siemens.
func physicalConductances(a *core.Array) [][]float64 {
	out := make([][]float64, a.Rows())
	for i := range out {
		out[i] = make([]float64, a.Cols())
		for j := range out[i] {
			out[i][j] = a.GetPhysicalConductanceForCell(i, j)
		}
	}
	return out
}
