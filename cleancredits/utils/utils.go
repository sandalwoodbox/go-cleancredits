package utils

import "math"

func ClampInt(n, min, max int) int {
	return int(math.Min(math.Max(float64(n), float64(min)), float64(max)))
}
