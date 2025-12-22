package pipeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"broadcom.com/vertex-ingestor/internal/ai"
	"broadcom.com/vertex-ingestor/internal/stages"
	"broadcom.com/vertex-ingestor/internal/store"
	"github.com/stretchr/testify/require"
)

// createTempMarkdownFiles creates a directory and fills it with markdown files
// until it reaches the desired size in MB.
func createTempMarkdownFiles(b *testing.B, dir string, totalSizeMB int) {
	b.Helper()
	const mb = 1024 * 1024
	const fileSize = 1 * mb
	totalSize := totalSizeMB * mb

	err := os.MkdirAll(dir, 0755)
	require.NoError(b, err)

	content := strings.Repeat("# Markdown Content\n\nThis is a sample markdown file for benchmarking.\n", fileSize/69) // 69 bytes per line

	numFiles := totalSize / fileSize
	for i := range numFiles {
		fileName := filepath.Join(dir, fmt.Sprintf("benchmark_file_%d.md", i))
		err := os.WriteFile(fileName, []byte(content), 0644)
		require.NoError(b, err)
	}
}

func BenchmarkPipeline(b *testing.B) {
	// Create a temporary directory for test data
	tmpDir, err := os.MkdirTemp("", "benchmark-pipeline-")
	require.NoError(b, err)
	b.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	// Generate 500MB of markdown files
	createTempMarkdownFiles(b, tmpDir, 1000)

	// Sprintf in Mock implementation takes a lot of memory
	// // --- Mock Services ---
	// mockEmbedder := new(ai.MockEmbeddingModel)
	// mockEmbedder.On("GenerateEmbeddings", mock.Anything, mock.Anything).Return([][]float32{{0.1, 0.2, 0.3}}, nil)

	benchmarkModel := ai.NewMockBenchmarkModel()

	// Sprintf in Mock implementation takes a lot of memory
	// mockStore := new(store.MockDatastore)
	// mockStore.On("Initialize", mock.Anything).Return(nil)
	// mockStore.On("InsertChunks", mock.Anything, mock.AnythingOfType("[]store.Chunk")).Return(32, nil)

	benchmarkStore := store.NewMockBenchmarkStore();

	b.ResetTimer()
	for b.Loop() {
		p := NewPipeline()
		p.AddStage(TypedStage(stages.NewDirLoader(tmpDir)))
		p.AddStage(TypedStage(stages.NewIndexFileFilter()))
		p.AddStage(TypedStage(stages.NewFileLoader()))
		p.AddStage(TypedStage(stages.NewMdAstSplitter()))
		p.AddStage(TypedStage(stages.NewBatcher[string](64)))
		p.AddStage(TypedStage(stages.NewEmbedder(benchmarkModel, 8)))
		p.AddStage(TypedStage(stages.NewBatcher[*stages.EmbedderOut](32)))
		p.AddStage(TypedStage(stages.NewStore(benchmarkStore)))

		// --- Run Pipeline ---
		out := p.Run(context.Background())
		for range out {
			// Pipeline execution happens as we consume the output
		}
	}
}
