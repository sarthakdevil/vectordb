package ivf

import (
	"math"
	"math/rand"
	"vectordb/internal/distances"
	"vectordb/internal/indexes"
)

// IVFIndex is a hybrid index that:
// - Starts with flat indexing for small datasets (< 100 vectors)
// - After 100 vectors, uses k-means clustering to create inverted files
// - Rebuilds the clusters every 100 additions
type IVFIndex struct {
	vectors     []indexes.Vector
	newVector   func(id int, data []float32) indexes.Vector
	nextID      int
	centroids   [][]float32      
	clusters    map[int][]indexes.Vector
	addCount    int                      
	numClusters int                      
}

func NewIVFIndex(newVector func(id int, data []float32) indexes.Vector) *IVFIndex {
	return &IVFIndex{
		newVector:   newVector,
		clusters:    make(map[int][]indexes.Vector),
		numClusters: 10,
	}
}

func (idx *IVFIndex) Add(data []float32) indexes.Vector {
	v := idx.newVector(idx.nextID, data)
	idx.vectors = append(idx.vectors, v)
	idx.nextID++
	idx.addCount++

	if len(idx.vectors) == 100 {
		idx.rebuildClusters()
	} else if len(idx.vectors) > 100 && idx.addCount >= 100 {
		idx.rebuildClusters()
		idx.addCount = 0
	}

	return v
}

func (idx *IVFIndex) DeleteByID(id int) bool {
	for i, v := range idx.vectors {
		if v.GetID() == id {
			idx.vectors = append(idx.vectors[:i], idx.vectors[i+1:]...)
			// Rebuild clusters if we have enough vectors
			if len(idx.vectors) >= 100 {
				idx.rebuildClusters()
			}
			return true
		}
	}
	return false
}

func (idx *IVFIndex) FindByID(id int) (indexes.Vector, bool) {
	for _, v := range idx.vectors {
		if v.GetID() == id {
			return v, true
		}
	}
	var zero indexes.Vector
	return zero, false
}

func (idx *IVFIndex) All() []indexes.Vector {
	result := make([]indexes.Vector, len(idx.vectors))
	copy(result, idx.vectors)
	return result
}

func (idx *IVFIndex) SearchByExactData(data []float32, searchtype string) (indexes.Vector, bool) {
	const threshold float32 = 0.8

	if len(idx.centroids) == 0 {
		return idx.flatSearch(data, searchtype, threshold)
	}

	nearestClusterID := idx.findNearestCluster(data)
	clusterVectors := idx.clusters[nearestClusterID]

	for _, vector := range clusterVectors {
		vectorData := vector.GetData()
		if len(vectorData) != len(data) {
			continue
		}

		if idx.matchesSearchType(vectorData, data, searchtype, threshold) {
			return vector, true
		}
	}

	var zero indexes.Vector
	return zero, false
}

func (idx *IVFIndex) rebuildClusters() {
	if len(idx.vectors) < 100 {
		idx.centroids = nil
		idx.clusters = make(map[int][]indexes.Vector)
		return
	}

	idx.centroids = make([][]float32, idx.numClusters)
	for i := 0; i < idx.numClusters; i++ {
		randIdx := rand.Intn(len(idx.vectors))
		centroid := make([]float32, len(idx.vectors[randIdx].GetData()))
		copy(centroid, idx.vectors[randIdx].GetData())
		idx.centroids[i] = centroid
	}
	
	idx.clusters = make(map[int][]indexes.Vector)
	for i := 0; i < idx.numClusters; i++ {
		idx.clusters[i] = []indexes.Vector{}
	}

	for _, vector := range idx.vectors {
		clusterID := idx.findNearestClusterCentroid(vector.GetData())
		idx.clusters[clusterID] = append(idx.clusters[clusterID], vector)
	}
}

func (idx *IVFIndex) findNearestCluster(data []float32) int {
	return idx.findNearestClusterCentroid(data)
}

func (idx *IVFIndex) findNearestClusterCentroid(data []float32) int {
	if len(idx.centroids) == 0 {
		return 0
	}

	nearestID := 0
	minDist := float32(math.Inf(1))

	for i, centroid := range idx.centroids {
		dist := distances.L2Distance(centroid, data)
		if dist < minDist {
			minDist = dist
			nearestID = i
		}
	}

	return nearestID
}

func (idx *IVFIndex) flatSearch(data []float32, searchtype string, threshold float32) (indexes.Vector, bool) {
	for _, vector := range idx.vectors {
		vectorData := vector.GetData()
		if len(vectorData) != len(data) {
			continue
		}

		if idx.matchesSearchType(vectorData, data, searchtype, threshold) {
			return vector, true
		}
	}

	var zero indexes.Vector
	return zero, false
}

func (idx *IVFIndex) matchesSearchType(vectorData, data []float32, searchtype string, threshold float32) bool {
	switch searchtype {
	case "cosine":
		return distances.CosineSimilarity(vectorData, data) >= threshold
	case "distance":
		return distances.L2Distance(vectorData, data) <= threshold
	default:
		return distances.ExactMatch(vectorData, data)
	}
}
