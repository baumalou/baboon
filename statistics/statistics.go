package statistics

import (
	"github.com/montanaflynn/stats"
)

// Min function
func Min(d map[int]float64) float64 {
	a, _ := stats.Min(d)
	return a
}

func Max() {
	a, _ = stats.Max(d)
	return a
}
