package stages

import (
	"context"
	"testing"

	"broadcom.com/vertex-ingestor/internal/ai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEmbedder_Run(t *testing.T) {
	// Create a mock embedding model
	mockModel := ai.NewMockEmbeddingModel(t)

	// Create an Embedder stage with the mock model
	embedder := Embedder{
		Model:   mockModel,
		Workers: 1,
	}

	// Create input and output channels
	in := make(chan []string)
	out := embedder.Run(context.Background(), in)

	// Test data
	batch1 := []string{"text1", "text2"}
	embeddings1 := [][]float32{{1.0, 2.0}, {3.0, 4.0}}

	batch2 := []string{"text3", "text4"}
	embeddings2 := [][]float32{{5.0, 6.0}, {7.0, 8.0}}

	// Set up mock expectations
	mockModel.On("GenerateEmbeddings", mock.Anything, batch1).Return(embeddings1, nil)
	mockModel.On("GenerateEmbeddings", mock.Anything, batch2).Return(embeddings2, nil)

	// Send test data to the input channel
	go func() {
		in <- batch1
		in <- batch2
		close(in)
	}()

	// Collect results from the output channel
	var results []*EmbedderOut
	for res := range out {
		results = append(results, res)
	}

	// Assert the results
	assert.Len(t, results, 4)

	assert.Equal(t, "text1", results[0].Chunk)
	assert.Equal(t, embeddings1[0], results[0].Vector)

	assert.Equal(t, "text2", results[1].Chunk)
	assert.Equal(t, embeddings1[1], results[1].Vector)

	assert.Equal(t, "text3", results[2].Chunk)
	assert.Equal(t, embeddings2[0], results[2].Vector)

	assert.Equal(t, "text4", results[3].Chunk)
	assert.Equal(t, embeddings2[1], results[3].Vector)

	// Assert that the mock's expectations were met
	mockModel.AssertExpectations(t)
}