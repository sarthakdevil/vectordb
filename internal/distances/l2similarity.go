package distances

import "math"

func L2Distance(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return float32(math.Inf(1))
	}

	var distSum float32
	for i := range a {
		diff := a[i] - b[i]
		distSum += diff * diff
	}

	return float32(math.Sqrt(float64(distSum)))
}
