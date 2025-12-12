package cmd

import (
	"context"

	"broadcom.com/vertex-ingestor/internal/ai"
	"broadcom.com/vertex-ingestor/internal/pipeline"
	"broadcom.com/vertex-ingestor/internal/stages"
	"broadcom.com/vertex-ingestor/internal/store"
	"github.com/spf13/cobra"
)

var ingestCmd = &cobra.Command{
	Use:   "ingest",
	Short: "Ingests documents, creates vector embeddings, and stores them in the vector database.",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		embedder := ai.NewGeminiEmbedder("embedding-001")
		embedder.Initialize(ctx)

		store := store.NewMockBenchmarkStore()

		p1 := pipeline.New(stages.NewDirLoader("."))
		p2 := pipeline.AddStage(p1, stages.NewIndexFileFilter())
		p3 := pipeline.AddStage(p2, stages.NewFileLoader())
		p4 := pipeline.AddStage(p3, stages.NewHeaderSplitter())
		p5 := pipeline.AddStage(p4, stages.NewBatcher[string](32))
		p6 := pipeline.AddStage(p5, stages.NewEmbedder(embedder, 4))
		p7 := pipeline.AddStage(p6, stages.NewBatcher[*stages.EmbedderOut](16))
		p8 := pipeline.AddStage(p7, stages.NewStore(store))
		p8.Run(ctx)
	},
}

func init() {
	rootCmd.AddCommand(ingestCmd)
}
