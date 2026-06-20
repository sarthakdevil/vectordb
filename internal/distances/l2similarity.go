package similarity
import (
	"math"
	"vectordb/internal/storage"
)

func L2similarity(vectors []storage.Vector, data []float32) (storage.Vector, bool) {
	const threshold float32 = 0.8

	for _, vector := range vectors {
		if len(vector.Data) != len(data) {
			continue
		}

		var distSum float32 = 0.0
		for i := range data {
			diff := vector.Data[i] - data[i]
			distSum += diff * diff
		}

		distance := float32(math.Sqrt(float64(distSum)))

		if distance <= threshold {
			return vector, true
		}
	}
	return storage.Vector{}, false
}