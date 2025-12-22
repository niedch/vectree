package stages

import (
	"context"
	"fmt"
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
		var size uint64 = 0

		for emb := range in {
			embedderOut := emb

			batchSize, chunks := convertEmbedderOutToChunks(embedderOut)
			resultCount, err := s.datastore.InsertChunks(ctx, chunks)
			if err != nil {
				log.Fatal(err)
			}

			size += uint64(batchSize)
			count += resultCount

			// Send a signal to output channel to keep pipeline alive
			select {
			case out <- count:
			case <-ctx.Done():
				return
			}
		}

		log.Printf("Stored %d embeddings\n", count)
		log.Printf("BatchSize: %s\n", formatBytes(size))
	}()

	return out
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func convertEmbedderOutToChunks(anyArr []*EmbedderOut) (int, []store.Chunk) {
	chunks := make([]store.Chunk, len(anyArr))
	batchSize := 0

	for idx, embedding := range anyArr {
		embedderOut := embedding
		if embedderOut == nil {
			log.Printf("Embedding is nil")
			continue
		}

		batchSize += len(embedderOut.Chunk)

		chunks[idx] = store.Chunk{
			Text:   embedderOut.Chunk,
			Vector: embedderOut.Vector,
		}
	}

	return batchSize, chunks
}
