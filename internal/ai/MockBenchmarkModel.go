package ai

import (
	"context"
)

type MockBenchmarkModel struct {
}

func NewMockBenchmarkModel() *MockBenchmarkModel {
	return &MockBenchmarkModel{}
}

func (ai *MockBenchmarkModel) ModelId() string {
	return "MockModel"
}

func (ai *MockBenchmarkModel) Initialize(ctx context.Context) {

}

func (ai *MockBenchmarkModel) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	return []float32{0.1, 0.2, 0.3}, nil

}

func (ai *MockBenchmarkModel) GenerateEmbeddings(ctx context.Context, text []string) ([][]float32, error) {
	embeddings := make([][]float32, len(text))
	for i := range text {
		embeddings[i] = []float32{0.1, 0.2, 0.3}
	}
	return embeddings, nil
}
