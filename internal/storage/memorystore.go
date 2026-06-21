package storage

import (
	"vectordb/internal/indexes"
	flatindex "vectordb/internal/indexes/flat"
)

type Vector struct {
	ID   int
	Data []float32
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

type VectorStore struct {
	flatindex *flatindex.FlatIndex
}

func newVector(id int, data []float32) indexes.Vector {
	return Vector{ID: id, Data: data}
}

func NewVectorStore() *VectorStore {
	return &VectorStore{flatindex: flatindex.NewFlatIndex(newVector)}
}

func (vs *VectorStore) ensureIndex() {
	if vs.flatindex == nil {
		vs.flatindex = flatindex.NewFlatIndex(newVector)
	}
}

func (vs *VectorStore) Add(data []float32) Vector {
	vs.ensureIndex()
	return vs.flatindex.Add(data).(Vector)
}

func (vs *VectorStore) DeleteByID(id int) bool {
	vs.ensureIndex()
	return vs.flatindex.DeleteByID(id)
}

func (vs *VectorStore) FindByID(id int) (Vector, bool) {
	vs.ensureIndex()
	v, ok := vs.flatindex.FindByID(id)
	if !ok {
		return Vector{}, false
	}

	return v.(Vector), true
}

func (vs *VectorStore) All() []Vector {
	vs.ensureIndex()
	vectors := vs.flatindex.All()
	result := make([]Vector, len(vectors))
	for i, vector := range vectors {
		result[i] = vector.(Vector)
	}

	return result
}
