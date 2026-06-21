package flat

import (
	"vectordb/internal/distances"
	"vectordb/internal/indexes"
)

type FlatIndex struct {
	vectors   []indexes.Vector
	nextID    int
	newVector func(id int, data []float32) indexes.Vector
}

func NewFlatIndex(newVector func(id int, data []float32) indexes.Vector) *FlatIndex {
	return &FlatIndex{newVector: newVector}
}

func (fi *FlatIndex) Add(data []float32) indexes.Vector {
	v := fi.newVector(fi.nextID, data)
	fi.vectors = append(fi.vectors, v)
	fi.nextID++

	return v
}

func (fi *FlatIndex) DeleteByID(id int) bool {
	for i, v := range fi.vectors {
		if v.GetID() == id {
			fi.vectors = append(fi.vectors[:i], fi.vectors[i+1:]...)
			return true
		}
	}

	return false
}

func (fi *FlatIndex) FindByID(id int) (indexes.Vector, bool) {
	for _, v := range fi.vectors {
		if v.GetID() == id {
			return v, true
		}
	}

	var zero indexes.Vector
	return zero, false
}

func (fi *FlatIndex) All() []indexes.Vector {
	result := make([]indexes.Vector, len(fi.vectors))
	copy(result, fi.vectors)
	return result
}

func (fi *FlatIndex) SearchByExactData(data []float32, searchtype string) (indexes.Vector, bool) {
	const threshold float32 = 0.8

	for _, vector := range fi.vectors {
		vectorData := vector.GetData()
		if len(vectorData) != len(data) {
			continue
		}

		switch searchtype {
		case "cosine":
			if distances.CosineSimilarity(vectorData, data) >= threshold {
				return vector, true
			}
		case "distance":
			if distances.L2Distance(vectorData, data) <= threshold {
				return vector, true
			}
		default:
			if distances.ExactMatch(vectorData, data) {
				return vector, true
			}
		}
	}

	var zero indexes.Vector
	return zero, false
}
