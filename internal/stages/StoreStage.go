package stages

import (
	"context"
	"fmt"
	"log"

	"github.com/niedch/vectree/internal/store"
)

type StoreStage struct {
	datastore store.Datastore
}

func NewStoreStage(datastore store.Datastore) *StoreStage {
	return &StoreStage{
		datastore: datastore,
	}
}

func (s *StoreStage) Run(ctx context.Context, in <-chan []*EmbedderOut) <-chan any {
	out := make(chan any)

	go func() {
		defer close(out)

		stats := newChunkStats()

		for emb := range in {
			batchSize, chunks := convertEmbedderOutToChunks(emb)
			resultCount, err := s.datastore.InsertChunks(ctx, chunks)
			if err != nil {
				log.Fatal(err)
			}

			stats.addBatch(batchSize, resultCount)
			for _, chunk := range chunks {
				stats.observeChunk(chunk.Text)
			}

			select {
			case out <- stats.count:
			case <-ctx.Done():
				return
			}
		}

		stats.log()
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
			Source:     embedderOut.Source,
		}
	}

	return batchSize, chunks
}

type chunkStats struct {
	count          int
	totalSize      uint64
	totalChunkSize uint64
	minChunk       int
	maxChunk       int
	batches        int
}

func newChunkStats() chunkStats {
	return chunkStats{minChunk: int(^uint(0) >> 1)}
}

func (s *chunkStats) addBatch(batchSize int, numResults int) {
	s.batches++
	s.totalSize += uint64(batchSize)
	s.count += numResults
}

func (s *chunkStats) observeChunk(text string) {
	n := len(text)
	s.totalChunkSize += uint64(n)
	if n < s.minChunk {
		s.minChunk = n
	}
	if n > s.maxChunk {
		s.maxChunk = n
	}
}

func (s *chunkStats) log() {
	avg := float64(0)
	if s.count > 0 {
		avg = float64(s.totalChunkSize) / float64(s.count)
	}
	log.Printf("Stored %d embeddings in %d batches\n", s.count, s.batches)
	log.Printf("Total text size: %s (%d bytes)\n", formatBytes(s.totalSize), s.totalSize)
	log.Printf("Chunk size stats - Min: %d, Max: %d, Avg: %.2f bytes\n", s.minChunk, s.maxChunk, avg)
}
