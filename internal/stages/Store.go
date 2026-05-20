package stages

import (
	"context"
	"fmt"
	"log"

	"github.com/niedch/tree-rag/internal/store"
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
		batchCount := 0
		minChunkSize := int(^uint(0) >> 1) // max int
		maxChunkSize := 0
		var totalChunkSize uint64 = 0

		for emb := range in {
			embedderOut := emb
			batchCount++

			batchSize, chunks := convertEmbedderOutToChunks(embedderOut)
			resultCount, err := s.datastore.InsertChunks(ctx, chunks)
			if err != nil {
				log.Fatal(err)
			}

			size += uint64(batchSize)
			count += resultCount

			// Track chunk statistics
			for _, chunk := range chunks {
				chunkLen := len(chunk.Text)
				totalChunkSize += uint64(chunkLen)
				if chunkLen < minChunkSize {
					minChunkSize = chunkLen
				}
				if chunkLen > maxChunkSize {
					maxChunkSize = chunkLen
				}
			}

			// Send a signal to output channel to keep pipeline alive
			select {
			case out <- count:
			case <-ctx.Done():
				return
			}
		}

		avgChunkSize := float64(0)
		if count > 0 {
			avgChunkSize = float64(totalChunkSize) / float64(count)
		}

		log.Printf("Stored %d embeddings in %d batches\n", count, batchCount)
		log.Printf("Total text size: %s (%d bytes)\n", formatBytes(size), size)
		log.Printf("Chunk size stats - Min: %d, Max: %d, Avg: %.2f bytes\n", minChunkSize, maxChunkSize, avgChunkSize)
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

		// Calculate text size in bytes
		textSize := len(embedderOut.Chunk)
		batchSize += textSize

		chunks[idx] = store.Chunk{
			Text:       embedderOut.Chunk,
			Vector:     embedderOut.Vector,
			Level:      embedderOut.Level,
			ParentId:   embedderOut.ParentId,
			DocumentId: embedderOut.DocumentId,
		}
	}

	return batchSize, chunks
}
