package store

import (
	"context"

	"broadcom.com/vertex-ingestor/internal/datastore"
)

type SqliteStore struct {
	datastore  *datastore.SqliteDatastore
}

func NewSqliteStore(datastore *datastore.SqliteDatastore) *SqliteStore {
	return &SqliteStore{datastore: datastore}
}

func (dc *SqliteStore) Initialize(ctx context.Context) error {
	// Unimplemented
	return nil
}

func (dc *SqliteStore) InsertChunks(ctx context.Context, chunks []Chunk) (int, error) {
	return 0, nil
	
}

func (dc *SqliteStore) SearchIndex(ctx context.Context, searchVector []float32) ([]SearchResult, error) {
	return nil, nil
}

