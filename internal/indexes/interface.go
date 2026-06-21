package indexes

// Vector is the interface that all vector types must implement
type Vector interface {
	GetID() int
	GetData() []float32
}
