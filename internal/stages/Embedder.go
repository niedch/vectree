package stages

import (
	"context"
	"log"
	"sync"

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
	if e.Workers <= 0 {
		e.Workers = 1
	}
	out := make(chan *EmbedderOut)
	var wg sync.WaitGroup
	wg.Add(e.Workers)

	worker := func() {
		defer wg.Done()

		for batch := range in {
			log.Println("Generating Embeddings for batch", len(batch))
			embs, err := e.Model.GenerateEmbeddings(ctx, batch)
			if err != nil {
				log.Fatalln(err)
				continue
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
					return
				}
			}
		}
	}

	for i := 0; i < e.Workers; i++ {
		go worker()
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}


