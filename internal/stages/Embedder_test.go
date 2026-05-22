package stages

import (
	"context"
	"testing"

	"github.com/niedch/vectree/internal/ai"
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
	in := make(chan []Section)
	out := embedder.Run(context.Background(), in)

	// Test data
	batch1 := []Section{
		{Text: "text1", Level: 1},
		{Text: "text2", Level: 2},
	}
	texts1 := []string{"text1", "text2"}
	embeddings1 := [][]float32{{1.0, 2.0}, {3.0, 4.0}}

	batch2 := []Section{
		{Text: "text3", Level: 1},
		{Text: "text4", Level: 3},
	}
	texts2 := []string{"text3", "text4"}
	embeddings2 := [][]float32{{5.0, 6.0}, {7.0, 8.0}}

	// Set up mock expectations
	mockModel.On("GenerateEmbeddings", mock.Anything, texts1).Return(embeddings1, nil)
	mockModel.On("GenerateEmbeddings", mock.Anything, texts2).Return(embeddings2, nil)

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
	assert.Equal(t, 1, results[0].Level)

	assert.Equal(t, "text2", results[1].Chunk)
	assert.Equal(t, embeddings1[1], results[1].Vector)
	assert.Equal(t, 2, results[1].Level)

	assert.Equal(t, "text3", results[2].Chunk)
	assert.Equal(t, embeddings2[0], results[2].Vector)
	assert.Equal(t, 1, results[2].Level)

	assert.Equal(t, "text4", results[3].Chunk)
	assert.Equal(t, embeddings2[1], results[3].Vector)
	assert.Equal(t, 3, results[3].Level)

	// Assert that the mock's expectations were met
	mockModel.AssertExpectations(t)
}
