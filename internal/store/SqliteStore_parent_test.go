package store

import (
	"context"
	"testing"

	"github.com/niedch/tree-rag/internal/conf"
	"github.com/niedch/tree-rag/internal/datastore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParentChildRelationships verifies that parent-child relationships
// are correctly established based on heading levels
func TestParentChildRelationships(t *testing.T) {
	// Create an in-memory database with migrations
	config := &conf.Config{
		Database: conf.Database{
			ConnectionString: ":memory:",
		},
	}
	db, err := datastore.OpenConnection(config)
	require.NoError(t, err)
	defer db.Close()

	ds := datastore.NewSqliteDatastore(db)
	store := NewSqliteStore(ds)

	ctx := context.Background()

	// Create test chunks representing a document structure:
	// # Main Title (Level 1) - should have no parent
	// ## Section A (Level 2) - parent should be Main Title
	// ### Subsection A1 (Level 3) - parent should be Section A
	// ## Section B (Level 2) - parent should be Main Title
	chunks := []Chunk{
		{
			Text:       "# Main Title\nIntroduction content",
			Vector:     make([]float32, 768),
			Level:      1,
			DocumentId: "doc1",
		},
		{
			Text:       "## Section A\nSection A content",
			Vector:     make([]float32, 768),
			Level:      2,
			DocumentId: "doc1",
		},
		{
			Text:       "### Subsection A1\nSubsection content",
			Vector:     make([]float32, 768),
			Level:      3,
			DocumentId: "doc1",
		},
		{
			Text:       "## Section B\nSection B content",
			Vector:     make([]float32, 768),
			Level:      2,
			DocumentId: "doc1",
		},
	}

	// Insert chunks
	count, err := store.InsertChunks(ctx, chunks)
	require.NoError(t, err)
	assert.Equal(t, 4, count)

	// Verify parent relationships
	// Document 1 (Main Title) should have no parent
	doc1, err := ds.GetDocument(ctx, 1)
	require.NoError(t, err)
	assert.Nil(t, doc1.ParentId, "Main Title should have no parent")
	assert.Equal(t, 1, doc1.Level)

	// Document 2 (Section A) should have parent = 1
	doc2, err := ds.GetDocument(ctx, 2)
	require.NoError(t, err)
	require.NotNil(t, doc2.ParentId, "Section A should have a parent")
	assert.Equal(t, 1, *doc2.ParentId, "Section A parent should be Main Title")
	assert.Equal(t, 2, doc2.Level)

	// Document 3 (Subsection A1) should have parent = 2
	doc3, err := ds.GetDocument(ctx, 3)
	require.NoError(t, err)
	require.NotNil(t, doc3.ParentId, "Subsection A1 should have a parent")
	assert.Equal(t, 2, *doc3.ParentId, "Subsection A1 parent should be Section A")
	assert.Equal(t, 3, doc3.Level)

	// Document 4 (Section B) should have parent = 1
	doc4, err := ds.GetDocument(ctx, 4)
	require.NoError(t, err)
	require.NotNil(t, doc4.ParentId, "Section B should have a parent")
	assert.Equal(t, 1, *doc4.ParentId, "Section B parent should be Main Title")
	assert.Equal(t, 2, doc4.Level)
}

// TestMultipleDocumentsParentIsolation verifies that parent relationships
// are isolated per source document
func TestMultipleDocumentsParentIsolation(t *testing.T) {
	// Create an in-memory database with migrations
	config := &conf.Config{
		Database: conf.Database{
			ConnectionString: ":memory:",
		},
	}
	db, err := datastore.OpenConnection(config)
	require.NoError(t, err)
	defer db.Close()

	ds := datastore.NewSqliteDatastore(db)
	store := NewSqliteStore(ds)

	ctx := context.Background()

	// Create chunks from two different documents
	// Document 1:
	//   # Title 1 (Level 1)
	//   ## Section 1A (Level 2)
	// Document 2:
	//   # Title 2 (Level 1)
	//   ## Section 2A (Level 2)
	chunks := []Chunk{
		{
			Text:       "# Title 1",
			Vector:     make([]float32, 768),
			Level:      1,
			DocumentId: "doc1",
		},
		{
			Text:       "## Section 1A",
			Vector:     make([]float32, 768),
			Level:      2,
			DocumentId: "doc1",
		},
		{
			Text:       "# Title 2",
			Vector:     make([]float32, 768),
			Level:      1,
			DocumentId: "doc2",
		},
		{
			Text:       "## Section 2A",
			Vector:     make([]float32, 768),
			Level:      2,
			DocumentId: "doc2",
		},
	}

	// Insert chunks
	count, err := store.InsertChunks(ctx, chunks)
	require.NoError(t, err)
	assert.Equal(t, 4, count)

	// Verify parent relationships
	// Document 1 (Title 1) - no parent
	doc1, err := ds.GetDocument(ctx, 1)
	require.NoError(t, err)
	assert.Nil(t, doc1.ParentId)

	// Document 2 (Section 1A) - parent should be doc 1
	doc2, err := ds.GetDocument(ctx, 2)
	require.NoError(t, err)
	require.NotNil(t, doc2.ParentId)
	assert.Equal(t, 1, *doc2.ParentId)

	// Document 3 (Title 2) - no parent (different source document)
	doc3, err := ds.GetDocument(ctx, 3)
	require.NoError(t, err)
	assert.Nil(t, doc3.ParentId, "Title 2 should have no parent (different source document)")

	// Document 4 (Section 2A) - parent should be doc 3, NOT doc 1
	doc4, err := ds.GetDocument(ctx, 4)
	require.NoError(t, err)
	require.NotNil(t, doc4.ParentId)
	assert.Equal(t, 3, *doc4.ParentId, "Section 2A parent should be Title 2 (doc 3), not Title 1")
}

// TestGetParentDocument verifies the GetParentDocument function
func TestGetParentDocument(t *testing.T) {
	// Create an in-memory database with migrations
	config := &conf.Config{
		Database: conf.Database{
			ConnectionString: ":memory:",
		},
	}
	db, err := datastore.OpenConnection(config)
	require.NoError(t, err)
	defer db.Close()

	ds := datastore.NewSqliteDatastore(db)
	store := NewSqliteStore(ds)

	ctx := context.Background()

	// Create test chunks
	chunks := []Chunk{
		{
			Text:       "# Main Title",
			Vector:     make([]float32, 768),
			Level:      1,
			DocumentId: "doc1",
		},
		{
			Text:       "## Section A",
			Vector:     make([]float32, 768),
			Level:      2,
			DocumentId: "doc1",
		},
	}

	// Insert chunks
	_, err = store.InsertChunks(ctx, chunks)
	require.NoError(t, err)

	// Get parent of document 2 (Section A)
	parent, err := ds.GetParentDocument(ctx, 2)
	require.NoError(t, err)
	assert.Equal(t, 1, parent.Id)
	assert.Equal(t, "# Main Title", parent.Document)
	assert.Equal(t, 1, parent.Level)

	// Try to get parent of document 1 (should fail - no parent)
	_, err = ds.GetParentDocument(ctx, 1)
	assert.Error(t, err, "Should error when trying to get parent of root document")
}
