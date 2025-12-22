package stages

import (
	"context"
	"log"

	"broadcom.com/vertex-ingestor/internal/ai"
)

type EmbedderOut struct {
	Chunk      string
	Vector     []float32
	Level      int
	ParentId   *int
	DocumentId string
}

type Embedder struct {
	Model   ai.EmbeddingModel
	Workers int
}

func NewEmbedder(model ai.EmbeddingModel, workers int) *Embedder {
	return &Embedder{
		Model:   model,
		Workers: workers,
	}
}

func (e Embedder) Run(ctx context.Context, in <-chan []SectionWithLevel) <-chan *EmbedderOut {
	return WorkerPoolStage(ctx, in, e.Workers, func(ctx context.Context, batch []SectionWithLevel, out chan<- *EmbedderOut) error {
		log.Println("Generating Embeddings for batch", len(batch))
		
		// Extract text from sections for embedding
		texts := make([]string, len(batch))
		for i, section := range batch {
			texts[i] = section.Text
		}
		
		embs, err := e.Model.GenerateEmbeddings(ctx, texts)
		if err != nil {
			log.Println(err)
			return err
		}

		for idx, emb := range embs {
			section := batch[idx]
			vector := emb

			outputItem := &EmbedderOut{
				Chunk:      section.Text,
				Vector:     vector,
				Level:      section.Level,
				ParentId:   section.ParentId,
				DocumentId: section.DocumentId,
			}

			select {
			case out <- outputItem:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		
		return nil
	})
}


