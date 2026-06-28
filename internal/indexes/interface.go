package indexes

// Vector is the interface that all vector types must implement
type Vector interface {
	GetID() int
	GetData() []float32
	GetText() string
}

// Index is the interface that all index implementations must implement
type Index interface {
	Add(data []float32, text string) Vector
	DeleteByID(id int) bool
	FindByID(id int) (Vector, bool)
	All() []Vector
	SearchByExactData(data []float32, searchtype string) (Vector, bool)
}
