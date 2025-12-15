package store

import (
	"context"
	"fmt"

	"broadcom.com/vertex-ingestor/internal/datastore"
)

type SqliteStore struct {
	datastore *datastore.SqliteDatastore
}

func NewSqliteStore(datastore *datastore.SqliteDatastore) *SqliteStore {
	return &SqliteStore{datastore: datastore}
}

// Initialize is a no-op for SqliteStore since the datastore is already initialized
func (s *SqliteStore) Initialize(ctx context.Context) error {
	return nil
}

// InsertChunks inserts multiple chunks (text + embeddings) into the database
// Returns the number of successfully inserted chunks
func (s *SqliteStore) InsertChunks(ctx context.Context, chunks []Chunk) (int, error) {
	count := 0
	
	for _, chunk := range chunks {
		// Create document and embedding objects
		doc := datastore.Document{
			Document: chunk.Text,
		}
		
		emb := datastore.Embedding{
			Embedding: chunk.Vector,
		}
		
		// Insert the document with its embedding
		_, err := s.datastore.InsertDocument(ctx, doc, emb)
		if err != nil {
			return count, fmt.Errorf("failed to insert chunk: %w", err)
		}
		
		count++
	}
	
	return count, nil
}

// SearchIndex searches for similar embeddings and returns matching documents
func (s *SqliteStore) SearchIndex(ctx context.Context, searchVector []float32) ([]SearchResult, error) {
	// Default limit for search results
	const defaultLimit = 10
	
	// Search for similar embeddings
	results, err := s.datastore.SearchSimilarEmbeddings(ctx, searchVector, defaultLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to search embeddings: %w", err)
	}
	
	// Convert DocumentWithEmbedding results to SearchResult format
	searchResults := make([]SearchResult, len(results))
	for i, result := range results {
		searchResults[i] = SearchResult{
			Chunk: result.Document.Document,
			// Note: Distance is not returned by SearchSimilarEmbeddings
			// If needed, it could be added to the datastore query
			Distance: 0.0,
		}
	}
	
	return searchResults, nil
}

