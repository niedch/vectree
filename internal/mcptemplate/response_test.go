package mcptemplate

import (
	"testing"

	"github.com/niedch/tree-rag/internal/datastore"
	"github.com/stretchr/testify/assert"
)

func TestBuildResponseString(t *testing.T) {
	// Create test documents
	docs := []datastore.DocumentWithEmbedding{
		{
			Document: datastore.Document{
				Id:       1,
				Document: "This is the first test document.",
				Level:    1,
			},
			Embedding:      []float32{0.1, 0.2, 0.3},
			EmbeddingRowid: 100,
		},
		{
			Document: datastore.Document{
				Id:       2,
				Document: "This is the second test document.",
				Level:    2,
			},
			Embedding:      []float32{0.4, 0.5, 0.6},
			EmbeddingRowid: 101,
		},
	}

	result := BuildResponseString(docs)

	// Verify the result contains expected elements
	expectedElements := []string{
		"## Document Result 1",
		"## Document Result 2",
		"**Metadata:**",
		"**Document ID:** 1",
		"**Document ID:** 2",
		"**Level:** 1",
		"**Level:** 2",
		"This is the first test document.",
		"This is the second test document.",
		"---", // Separator between documents
	}

	for _, expected := range expectedElements {
		assert.Contains(t, result, expected, "Expected result to contain %q", expected)
	}
}

func TestBuildResponseStringEmpty(t *testing.T) {
	// Test with empty slice
	docs := []datastore.DocumentWithEmbedding{}
	result := BuildResponseString(docs)

	// Should still contain the tip even with no documents
	assert.Contains(t, result, "💡 Tip:", "Expected result to contain tip")
	
	// Should not contain any document results
	assert.NotContains(t, result, "## Document Result", "Expected no document results for empty input")
}

func TestBuildResponseStringSingleDocument(t *testing.T) {
	// Test with single document
	docs := []datastore.DocumentWithEmbedding{
		{
			Document: datastore.Document{
				Id:       42,
				Document: "Single document content.",
				Level:    3,
			},
			Embedding:      []float32{0.7, 0.8, 0.9},
			EmbeddingRowid: 200,
		},
	}

	result := BuildResponseString(docs)

	// Should contain the tip separator and final tip
	assert.Contains(t, result, "💡 Tip:", "Result should contain the tip at the end")

	// Should contain the document content and metadata
	expectedElements := []string{
		"## Document Result 1",
		"**Document ID:** 42",
		"**Level:** 3",
		"Single document content.",
	}

	for _, expected := range expectedElements {
		assert.Contains(t, result, expected, "Expected result to contain %q", expected)
	}
}
