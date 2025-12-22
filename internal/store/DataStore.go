package store

import "context"

type SearchResult struct {
	Chunk    string  `db:"chunk"`
	Distance float64 `db:"distance"`
}

type Chunk struct {
	Text   string
	Vector []float32
	Level  int
}

type Datastore interface {
	Initialize(ctx context.Context) error

	InsertChunks(ctx context.Context, chunks []Chunk) (int, error)
	SearchIndex(ctx context.Context, searchToken []float32) ([]SearchResult, error)
}
