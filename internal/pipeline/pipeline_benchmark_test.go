package pipeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/niedch/tree-rag/internal/ai"
	"github.com/niedch/tree-rag/internal/stages"
	"github.com/niedch/tree-rag/internal/store"
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

	// Create realistic hierarchical markdown content with nested sections
	// This will demonstrate the AST splitter's duplication feature
	sectionTemplate := `# Main Section %d

This is the introduction to section %d with some context and overview information that will be duplicated across child sections.

## Subsection %d.1

Detailed content for subsection %d.1. This includes technical information, code examples, and explanations that are specific to this subsection.

### Deep Subsection %d.1.1

Even more detailed content at the third level. This demonstrates deep nesting and how the AST splitter handles multiple levels of hierarchy.

## Subsection %d.2

More detailed content for subsection %d.2. This section covers different aspects and includes additional examples and use cases.

### Deep Subsection %d.2.1

Additional deep content that shows how nested sections work in practice.

## Subsection %d.3

Final subsection with concluding remarks and additional information.

`
	
	numFiles := totalSize / fileSize
	for i := range numFiles {
		// Generate content by repeating the section template
		var contentBuilder strings.Builder
		sectionsPerFile := fileSize / len(fmt.Sprintf(sectionTemplate, i, i, i, i, i, i, i, i, i, i))
		for j := range sectionsPerFile {
			sectionNum := i*sectionsPerFile + j
			contentBuilder.WriteString(fmt.Sprintf(sectionTemplate, 
				sectionNum, sectionNum, sectionNum, sectionNum, sectionNum, 
				sectionNum, sectionNum, sectionNum, sectionNum, sectionNum))
		}
		
		fileName := filepath.Join(dir, fmt.Sprintf("benchmark_file_%d.md", i))
		err := os.WriteFile(fileName, []byte(contentBuilder.String()), 0644)
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
		p.AddStage(TypedStage(stages.NewBatcher[stages.SectionWithLevel](64)))
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
