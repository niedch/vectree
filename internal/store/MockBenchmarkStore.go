package store

import (
	"context"
	"log"
)

type MockBenchmarkStore struct {
}

func NewMockBenchmarkStore() *MockBenchmarkStore {
	return &MockBenchmarkStore{}
}

func (dc *MockBenchmarkStore) Initialize(ctx context.Context) error {
	return nil
}

func (dc *MockBenchmarkStore) InsertChunks(ctx context.Context, chunks []Chunk) (int, error) {
	log.Printf("Inserting %d Chunks", len(chunks))
	return len(chunks), nil
}

func (dc *MockBenchmarkStore) SearchIndex(ctx context.Context, searchVector []float32) ([]SearchResult, error) {
	return []SearchResult{}, nil
}

