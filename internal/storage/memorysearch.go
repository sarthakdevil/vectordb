package storage

func (vs *VectorStore) SearchByID(id int) (Vector, bool) {
	return vs.FindByID(id)
}

func (vs *VectorStore) SearchByExactData(data []float32, searchtype string) (Vector, bool) {
	vs.ensureIndex()
	v, ok := vs.flatindex.SearchByExactData(data, searchtype)
	if !ok {
		return Vector{}, false
	}

	return v.(Vector), true
}
