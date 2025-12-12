package stages

import (
	"context"
	"log"

	"broadcom.com/vertex-ingestor/internal/store"
)

type Store struct {
	datastore store.Datastore
}

func NewStore(datastore store.Datastore) *Store {
	return &Store{
		datastore: datastore,
	}
}

func (s *Store) Run(ctx context.Context, in <-chan []*EmbedderOut) <-chan any {
	out := make(chan any)

	go func() {
		defer close(out)
		err := s.datastore.Initialize(ctx)
		if err != nil {
			log.Fatalln(err)
		}

		count := 0
		for emb := range in {
			embedderOut := emb

			chunks := convertEmbedderOutToChunks(embedderOut)
			result_Count, err := s.datastore.InsertChunks(ctx, chunks)
			if err != nil {
				log.Fatal(err)
			}

			count += result_Count
		}

		log.Println("Stored", count, "embeddings")
	}()

	return out
}

func convertEmbedderOutToChunks(anyArr []*EmbedderOut) []store.Chunk {
	chunks := make([]store.Chunk, len(anyArr))

	for idx, embedding := range anyArr {
		embedderOut := embedding
		if embedderOut == nil {
			log.Printf("Embedding is nil")
			continue
		}

		chunks[idx] = store.Chunk{
			Text:   embedderOut.Chunk,
			Vector: embedderOut.Vector,
		}
	}

	return chunks
}
