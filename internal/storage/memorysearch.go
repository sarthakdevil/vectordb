package storage

import "vectordb/internal/distances"

func (vs *VectorStore) SearchByID(id int) (Vector, bool) {
	return vs.FindByID(id)
}

func (vs *VectorStore) SearchByExactData(data []float32, searchtype string) (Vector, bool) {
	switch searchtype {
	case "cosine":
		return distances.Cosine(vs.All(), data)
	case "distance":
		return distances.Distance(vs.All(), data)
	default:
		return distances.L2(vs.All(), data)
	}
}
