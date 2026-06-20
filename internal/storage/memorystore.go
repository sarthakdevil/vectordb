package storage

type Vector struct {
	ID   int
	Data []float32
}

func (v Vector) Size() int {
	return len(v.Data)
}

type VectorStore struct {
	vectors []Vector
	nextID  int
}

func NewVectorStore() *VectorStore {
	return &VectorStore{}
}

func (vs *VectorStore) Add(data []float32) Vector {
	v := Vector{
		ID:   vs.nextID,
		Data: data,
	}

	vs.vectors = append(vs.vectors, v)
	vs.nextID++

	return v
}

func (vs *VectorStore) DeleteByID(id int) bool {
	for i, v := range vs.vectors {
		if v.ID == id {
			vs.vectors = append(vs.vectors[:i], vs.vectors[i+1:]...)
			return true
		}
	}

	return false
}

func (vs *VectorStore) FindByID(id int) (Vector, bool) {
	for _, v := range vs.vectors {
		if v.ID == id {
			return v, true
		}
	}

	return Vector{}, false
}

func (vs *VectorStore) All() []Vector {
	result := make([]Vector, len(vs.vectors))
	copy(result, vs.vectors)
	return result
}
