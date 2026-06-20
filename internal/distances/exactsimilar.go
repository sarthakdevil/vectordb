package similarity
import "vectordb/internal/storage"

func findExactMatch(vectors []storage.Vector, data []float32) (storage.Vector, bool) {
	for _, vector := range vectors {
		if len(vector.Data) != len(data) {
			continue
		}

		matched := true
		for i := range data {
			if vector.Data[i] != data[i] {
				matched = false
				break
			}
		}

		if matched {
			return vector, true
		}
	}

	return storage.Vector{}, false
}