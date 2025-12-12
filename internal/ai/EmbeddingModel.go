package ai

import "context"

type EmbeddingModel interface {
	ModelId() string
	Initialize(ctx context.Context)

	GenerateEmbeddings(ctx context.Context, text []string) ([][]float32, error)
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
}
