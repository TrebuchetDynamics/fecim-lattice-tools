package arraysim

import "math/rand"

func GenerateCheckerboard(rows, cols, quantLevels int) [][]int {
	pattern := makePattern(rows, cols)
	high := maxLevel(quantLevels)
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			if (r+c)%2 == 0 {
				pattern[r][c] = high
			}
		}
	}
	return pattern
}

func GenerateAllOnes(rows, cols, quantLevels int) [][]int {
	pattern := makePattern(rows, cols)
	high := maxLevel(quantLevels)
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			pattern[r][c] = high
		}
	}
	return pattern
}

func GenerateAllZeros(rows, cols int) [][]int {
	return makePattern(rows, cols)
}

func GenerateWalkingOnes(rows, cols, position, quantLevels int) [][]int {
	pattern := makePattern(rows, cols)
	r, c, ok := decodePosition(rows, cols, position)
	if !ok {
		return pattern
	}
	pattern[r][c] = maxLevel(quantLevels)
	return pattern
}

func GenerateWalkingZeros(rows, cols, position, quantLevels int) [][]int {
	pattern := GenerateAllOnes(rows, cols, quantLevels)
	r, c, ok := decodePosition(rows, cols, position)
	if !ok {
		return pattern
	}
	pattern[r][c] = 0
	return pattern
}

func GenerateDiagonal(rows, cols, quantLevels int) [][]int {
	pattern := makePattern(rows, cols)
	high := maxLevel(quantLevels)
	if rows <= 0 || cols <= 0 || high == 0 {
		return pattern
	}
	span := rows - 1
	if cols-1 > span {
		span = cols - 1
	}
	if span <= 0 {
		pattern[0][0] = high
		return pattern
	}
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			idx := r
			if c > idx {
				idx = c
			}
			pattern[r][c] = int(float64(high) * float64(idx) / float64(span))
		}
	}
	return pattern
}

func GenerateRowStripe(rows, cols, quantLevels int) [][]int {
	pattern := makePattern(rows, cols)
	high := maxLevel(quantLevels)
	for r := 0; r < rows; r++ {
		if r%2 != 0 {
			continue
		}
		for c := 0; c < cols; c++ {
			pattern[r][c] = high
		}
	}
	return pattern
}

func GenerateRandom(rows, cols, quantLevels int, seed int64) [][]int {
	pattern := makePattern(rows, cols)
	high := maxLevel(quantLevels)
	if high == 0 {
		return pattern
	}
	rng := rand.New(rand.NewSource(seed))
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			pattern[r][c] = rng.Intn(high + 1)
		}
	}
	return pattern
}

func makePattern(rows, cols int) [][]int {
	if rows < 0 {
		rows = 0
	}
	if cols < 0 {
		cols = 0
	}
	out := make([][]int, rows)
	for r := 0; r < rows; r++ {
		out[r] = make([]int, cols)
	}
	return out
}

func maxLevel(quantLevels int) int {
	if quantLevels <= 1 {
		return 0
	}
	return quantLevels - 1
}

func decodePosition(rows, cols, position int) (int, int, bool) {
	total := rows * cols
	if rows <= 0 || cols <= 0 || position < 0 || position >= total {
		return 0, 0, false
	}
	return position / cols, position % cols, true
}
