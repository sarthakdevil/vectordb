package similarity

import (
	"math"
	"vectordb/internal/storage"
)

func CosineSimilarity(vectors []storage.Vector, data []float32) (storage.Vector, bool) {
	const threshold float32 = 0.8

	for _, vector := range vectors {
		if len(vector.Data) != len(data) {
			continue
		}

		var dotProduct float32
		var normA float32
		var normB float32

		for i := range data {
			dotProduct += vector.Data[i] * data[i]
			normA += vector.Data[i] * vector.Data[i]
			normB += data[i] * data[i]
		}

		if normA == 0 || normB == 0 {
			continue
		}

		similarity := dotProduct /
			(float32(math.Sqrt(float64(normA))) *
				float32(math.Sqrt(float64(normB))))

		if similarity >= threshold {
			return vector, true
		}
	}

	return storage.Vector{}, false
}