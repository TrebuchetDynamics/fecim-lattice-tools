package arraysim

import "testing"

func TestPatternMarchCMinus_4x4_ReadWriteAndStuckAtCoverage(t *testing.T) {
	const n = 4
	state := makeIntGrid(n, n)

	// ⇕(w0)
	for r := 0; r < n; r++ {
		for c := 0; c < n; c++ {
			writeCell(t, state, r, c, 0)
		}
	}
	assertAllValue(t, state, 0)

	// ⇑(r0,w1)
	for r := 0; r < n; r++ {
		for c := 0; c < n; c++ {
			readCellEquals(t, state, r, c, 0)
			writeCell(t, state, r, c, 1)
			readCellEquals(t, state, r, c, 1)
		}
	}

	// ⇑(r1,w0)
	for r := 0; r < n; r++ {
		for c := 0; c < n; c++ {
			readCellEquals(t, state, r, c, 1)
			writeCell(t, state, r, c, 0)
			readCellEquals(t, state, r, c, 0)
		}
	}

	// ⇓(r0,w1)
	for r := n - 1; r >= 0; r-- {
		for c := n - 1; c >= 0; c-- {
			readCellEquals(t, state, r, c, 0)
			writeCell(t, state, r, c, 1)
			readCellEquals(t, state, r, c, 1)
		}
	}

	// ⇓(r1,w0)
	for r := n - 1; r >= 0; r-- {
		for c := n - 1; c >= 0; c-- {
			readCellEquals(t, state, r, c, 1)
			writeCell(t, state, r, c, 0)
			readCellEquals(t, state, r, c, 0)
		}
	}

	// ⇕(r0)
	for r := 0; r < n; r++ {
		for c := 0; c < n; c++ {
			readCellEquals(t, state, r, c, 0)
		}
	}

	assertAllValue(t, state, 0)

	assertMarchCMinusDetectsStuckAt(t, n)
}

func assertMarchCMinusDetectsStuckAt(t *testing.T, n int) {
	t.Helper()
	detected := 0
	for r := 0; r < n; r++ {
		for c := 0; c < n; c++ {
			if marchCMinusDetectsFault(n, fault{r: r, c: c, stuckAt: 0}) {
				detected++
			}
			if marchCMinusDetectsFault(n, fault{r: r, c: c, stuckAt: 1}) {
				detected++
			}
		}
	}
	want := 2 * n * n
	if detected != want {
		t.Fatalf("stuck-at coverage incomplete: detected=%d want=%d", detected, want)
	}
}

type fault struct {
	r, c    int
	stuckAt int
}

func marchCMinusDetectsFault(n int, f fault) bool {
	state := makeIntGrid(n, n)
	write := func(r, c, v int) {
		if r == f.r && c == f.c {
			state[r][c] = f.stuckAt
			return
		}
		state[r][c] = v
	}
	read := func(r, c int) int { return state[r][c] }

	for r := 0; r < n; r++ {
		for c := 0; c < n; c++ {
			write(r, c, 0)
		}
	}
	for r := 0; r < n; r++ {
		for c := 0; c < n; c++ {
			if read(r, c) != 0 {
				return true
			}
			write(r, c, 1)
		}
	}
	for r := 0; r < n; r++ {
		for c := 0; c < n; c++ {
			if read(r, c) != 1 {
				return true
			}
			write(r, c, 0)
		}
	}
	for r := n - 1; r >= 0; r-- {
		for c := n - 1; c >= 0; c-- {
			if read(r, c) != 0 {
				return true
			}
			write(r, c, 1)
		}
	}
	for r := n - 1; r >= 0; r-- {
		for c := n - 1; c >= 0; c-- {
			if read(r, c) != 1 {
				return true
			}
			write(r, c, 0)
		}
	}
	for r := 0; r < n; r++ {
		for c := 0; c < n; c++ {
			if read(r, c) != 0 {
				return true
			}
		}
	}
	return false
}

func writeCell(t *testing.T, state [][]int, r, c, v int) {
	t.Helper()
	before := cloneIntGrid(state)
	state[r][c] = v
	for i := range state {
		for j := range state[i] {
			if i == r && j == c {
				continue
			}
			if state[i][j] != before[i][j] {
				t.Fatalf("write at (%d,%d) changed non-target (%d,%d)", r, c, i, j)
			}
		}
	}
}

func readCellEquals(t *testing.T, state [][]int, r, c, want int) {
	t.Helper()
	if got := state[r][c]; got != want {
		t.Fatalf("read mismatch at (%d,%d): got=%d want=%d", r, c, got, want)
	}
}

func assertAllValue(t *testing.T, state [][]int, want int) {
	t.Helper()
	for r := range state {
		for c := range state[r] {
			if state[r][c] != want {
				t.Fatalf("state mismatch at (%d,%d): got=%d want=%d", r, c, state[r][c], want)
			}
		}
	}
}

func makeIntGrid(rows, cols int) [][]int {
	out := make([][]int, rows)
	for r := 0; r < rows; r++ {
		out[r] = make([]int, cols)
	}
	return out
}

func cloneIntGrid(src [][]int) [][]int {
	out := make([][]int, len(src))
	for i := range src {
		out[i] = append([]int(nil), src[i]...)
	}
	return out
}
