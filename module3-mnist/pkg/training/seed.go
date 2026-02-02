package training

import (
	"math/rand"
	"os"
	"strconv"
)

// Seed global RNG for reproducible training when FECIM_DEBUG_SEED is set.
func init() {
	seedStr := os.Getenv("FECIM_DEBUG_SEED")
	if seedStr == "" {
		return
	}
	seed, err := strconv.ParseInt(seedStr, 10, 64)
	if err != nil {
		return
	}
	rand.Seed(seed)
}
