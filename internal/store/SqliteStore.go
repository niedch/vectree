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
// This function maintains parent-child relationships based on heading levels within each document
func (s *SqliteStore) InsertChunks(ctx context.Context, chunks []Chunk) (int, error) {
	count := 0
	
	// Track the most recent document ID at each level for each source document
	// parentStacks[documentId][level] = database document ID
	parentStacks := make(map[string]map[int]int64)
	
	for _, chunk := range chunks {
		// Get or create parent stack for this source document
		parentStack, exists := parentStacks[chunk.DocumentId]
		if !exists {
			parentStack = make(map[int]int64)
			parentStacks[chunk.DocumentId] = parentStack
		}
		
		// Find the parent: the most recent document with a level less than current level
		// within the same source document
		var parentId *int
		for level := chunk.Level - 1; level >= 1; level-- {
			if docId, exists := parentStack[level]; exists {
				parentIdInt := int(docId)
				parentId = &parentIdInt
				break
			}
		}
		
		// Create document and embedding objects
		doc := datastore.Document{
			Document: chunk.Text,
			Level:    chunk.Level,
			ParentId: parentId,
		}
		
		emb := datastore.Embedding{
			Embedding: chunk.Vector,
		}
		
		// Insert the document with its embedding
		docId, err := s.datastore.InsertDocument(ctx, doc, emb)
		if err != nil {
			return count, fmt.Errorf("failed to insert chunk: %w", err)
		}
		
		// Update the parent stack for this level in this source document
		parentStack[chunk.Level] = docId
		
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

