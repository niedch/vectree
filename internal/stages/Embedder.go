package stages

import (
	"context"
	"log"

	"broadcom.com/vertex-ingestor/internal/ai"
)

type EmbedderOut struct {
	Chunk  string
	Vector []float32
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

func (e Embedder) Run(ctx context.Context, in <-chan []string) <-chan *EmbedderOut {
	return WorkerPoolStage(ctx, in, e.Workers, func(ctx context.Context, batch []string, out chan<- *EmbedderOut) error {
		log.Println("Generating Embeddings for batch", len(batch))
		embs, err := e.Model.GenerateEmbeddings(ctx, batch)
		if err != nil {
			log.Println(err)
			return err
		}

		for idx, emb := range embs {
			chunk := batch[idx]
			vector := emb

			outputItem := &EmbedderOut{
				Chunk:  chunk,
				Vector: vector,
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


