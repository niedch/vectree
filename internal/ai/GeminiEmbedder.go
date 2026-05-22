package ai

import (
	"context"
	"fmt"
	"log"

	"github.com/niedch/vectree/internal/conf"
	"google.golang.org/genai"
)

type GeminiEmbedder struct {
	embeddingModel string
	apikey         string
	Client         *genai.Client
}

func NewGeminiEmbedder(ctx context.Context, cfg *conf.GeminiProviderConfig) (*GeminiEmbedder, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: cfg.APIKey})
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client: %w", err)
	}

	return &GeminiEmbedder{
		embeddingModel: cfg.EmbeddingModel,
		apikey:         cfg.APIKey,
		Client:         client,
	}, nil
}

func (ai *GeminiEmbedder) ModelId() string {
	return ai.embeddingModel
}

func (ai *GeminiEmbedder) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	contents := []*genai.Content{
		genai.NewContentFromText(text, genai.RoleUser),
	}

	result := ai.callEmbedding(ctx, contents)
	return result.Embeddings[0].Values, nil
}

func (ai *GeminiEmbedder) GenerateEmbeddings(ctx context.Context, text []string) ([][]float32, error) {
	contents := ai.GetContents(text)
	result := ai.callEmbedding(ctx, contents)

	embeddings := make([][]float32, len(result.Embeddings))
	for i, ai := range result.Embeddings {
		embeddings[i] = ai.Values
	}

	return embeddings, nil
}

func (ai *GeminiEmbedder) callEmbedding(ctx context.Context, contents []*genai.Content) *genai.EmbedContentResponse {
	result, err := ai.Client.Models.EmbedContent(ctx,
		ai.embeddingModel,
		contents,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}
	return result
}

func (ai *GeminiEmbedder) GetContents(texts []string) []*genai.Content {
	contents := make([]*genai.Content, len(texts))
	for idx, text := range texts {
		contents[idx] = genai.NewContentFromText(text, genai.RoleUser)
	}
	return contents
}
