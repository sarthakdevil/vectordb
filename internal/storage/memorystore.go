package storage

import (
	"vectordb/internal/indexes"
	flatindex "vectordb/internal/indexes/flat"
	ivfindex "vectordb/internal/indexes/ivf"
)

type Vector struct {
	ID   int
	Data []float32
	Text string
}

func (v Vector) Size() int {
	return len(v.Data)
}

func (v Vector) GetID() int {
	return v.ID
}

func (v Vector) GetData() []float32 {
	return v.Data
}

func (v Vector) GetText() string {
	return v.Text
}

type VectorStore struct {
	index indexes.Index
}

func newVector(id int, data []float32, text string) indexes.Vector {
	return Vector{ID: id, Data: data, Text: text}
}

func NewVectorStore(indexType string) *VectorStore {
	switch indexType {
	case "flat":
		return &VectorStore{index: flatindex.NewFlatIndex(newVector)}
	default:
		return &VectorStore{index: ivfindex.NewIVFIndex(newVector)}
	}
}

func (vs *VectorStore) Add(data []float32, text string) Vector {
	return vs.index.Add(data, text).(Vector)
}

func (vs *VectorStore) DeleteByID(id int) bool {
	return vs.index.DeleteByID(id)
}

func (vs *VectorStore) FindByID(id int) (Vector, bool) {
	v, ok := vs.index.FindByID(id)
	if !ok {
		return Vector{}, false
	}

	return v.(Vector), true
}

func (vs *VectorStore) All() []Vector {
	vectors := vs.index.All()
	result := make([]Vector, len(vectors))
	for i, vector := range vectors {
		result[i] = vector.(Vector)
	}

	return result
}
