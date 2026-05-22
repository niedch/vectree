package pipeline

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/niedch/vectree/internal/ai"
	"github.com/niedch/vectree/internal/stages"
	"github.com/niedch/vectree/internal/store"
)

var benchDataDir string

func TestMain(m *testing.M) {
	var err error
	benchDataDir, err = os.MkdirTemp("", "benchmark-pipeline-")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(benchDataDir)

	createTempMarkdownFiles(benchDataDir, 1000)

	os.Exit(m.Run())
}

// createTempMarkdownFiles creates a directory and fills it with markdown files
// until it reaches the desired size in MB.
func createTempMarkdownFiles(dir string, totalSizeMB int) {
	const mb = 1024 * 1024
	const fileSize = 1 * mb
	totalSize := totalSizeMB * mb

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
			fmt.Fprintf(&contentBuilder, sectionTemplate,
				sectionNum, sectionNum, sectionNum, sectionNum, sectionNum,
				sectionNum, sectionNum, sectionNum, sectionNum, sectionNum)
		}

		fileName := filepath.Join(dir, fmt.Sprintf("benchmark_file_%d.md", i))
		err := os.WriteFile(fileName, []byte(contentBuilder.String()), 0644)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkPipeline(b *testing.B) {
	benchmarkModel := ai.NewMockBenchmarkModel()
	benchmarkStore := store.NewMockBenchmarkStore()

	p := NewPipeline()
	p.AddStage(TypedStage(stages.NewDirLoader(benchDataDir)))
	p.AddStage(TypedStage(stages.NewIndexFileFilter()))
	p.AddStage(TypedStage(stages.NewFileLoader()))
	p.AddStage(TypedStage(stages.NewMdAstSplitter()))
	p.AddStage(TypedStage(stages.NewBatcher[stages.SectionWithLevel](64)))
	p.AddStage(TypedStage(stages.NewEmbedder(benchmarkModel, 8)))
	p.AddStage(TypedStage(stages.NewBatcher[*stages.EmbedderOut](32)))
	p.AddStage(TypedStage(stages.NewStoreStage(benchmarkStore)))

	b.ResetTimer()
	for b.Loop() {
		out := p.Run(context.Background())
		for range out {
		}
	}
}
