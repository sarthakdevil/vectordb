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

func (vs *VectorStore) ensureIndex() {
	if vs.index == nil {
		vs.index = flatindex.NewFlatIndex(newVector)
	}
}

func (vs *VectorStore) Add(data []float32, text string) Vector {
	vs.ensureIndex()
	return vs.index.Add(data, text).(Vector)
}

func (vs *VectorStore) DeleteByID(id int) bool {
	vs.ensureIndex()
	return vs.index.DeleteByID(id)
}

func (vs *VectorStore) FindByID(id int) (Vector, bool) {
	vs.ensureIndex()
	v, ok := vs.index.FindByID(id)
	if !ok {
		return Vector{}, false
	}

	return v.(Vector), true
}

func (vs *VectorStore) All() []Vector {
	vs.ensureIndex()
	vectors := vs.index.All()
	result := make([]Vector, len(vectors))
	for i, vector := range vectors {
		result[i] = vector.(Vector)
	}

	return result
}
