package cmd

import (
	"context"
	"sync"

	"broadcom.com/vertex-ingestor/internal/ai"
	"broadcom.com/vertex-ingestor/internal/conf"
	"broadcom.com/vertex-ingestor/internal/datastore"
	"broadcom.com/vertex-ingestor/internal/pipeline"
	"broadcom.com/vertex-ingestor/internal/stages"
	"broadcom.com/vertex-ingestor/internal/store"
	"github.com/spf13/cobra"
)

const (
	TOC_URL = "https://techdocs.broadcom.com/us/en/ca-enterprise-software/valueops/connectall/4-0/jcr:content.toc.html"
)

var ingestCmd = &cobra.Command{
	Use:   "ingest",
	Short: "Ingests documents, creates vector embeddings, and stores them in the vector database.",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		config := conf.Load()

		embedder := ai.NewGeminiEmbedder("embedding-001")
		embedder.Initialize(ctx)

		db, err := datastore.OpenConnection(config)
		if err != nil {
			panic(err)
		}

		ds := datastore.NewSqliteDatastore(db)
		store := store.NewSqliteStore(ds)

		var wg sync.WaitGroup

		RunDocumentationPipelineAsync(ctx, &wg, embedder, store)
		RunMarkdownPipelineAsync(ctx, &wg, embedder, store)

		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(ingestCmd)
}

func RunMarkdownPipelineAsync(ctx context.Context, wg *sync.WaitGroup, embedder ai.EmbeddingModel, store store.Datastore) {
	wg.Add(1)

	go func() {
		p1 := pipeline.New(stages.NewDirLoader("../connectall"))
		p2 := pipeline.AddStage(p1, stages.NewNodeModulesFilter())
		p3 := pipeline.AddStage(p2, stages.NewFileLoader())
		p4 := pipeline.AddStage(p3, stages.NewHeaderSplitter())
		p5 := pipeline.AddStage(p4, stages.NewBatcher[string](64))
		p6 := pipeline.AddStage(p5, stages.NewEmbedder(embedder, 8))
		p7 := pipeline.AddStage(p6, stages.NewBatcher[*stages.EmbedderOut](64))
		p8 := pipeline.AddStage(p7, stages.NewStore(store))
		p8.Run(ctx)

		wg.Done()
	}()
}

func RunDocumentationPipelineAsync(ctx context.Context, wg *sync.WaitGroup, embedder ai.EmbeddingModel, store store.Datastore) {
	wg.Add(1)
	go func() {
		p1 := pipeline.New(stages.NewIndexLoader(TOC_URL))
		p2 := pipeline.AddStage(p1, stages.NewDebugStage())
		p3 := pipeline.AddStage(p2, stages.NewContentLoader(10))
		p4 := pipeline.AddStage(p3, stages.NewBatcher[string](64))
		p5 := pipeline.AddStage(p4, stages.NewEmbedder(embedder, 8))
		p6 := pipeline.AddStage(p5, stages.NewBatcher[*stages.EmbedderOut](64))
		p7 := pipeline.AddStage(p6, stages.NewStore(store))
		p7.Run(ctx)

		wg.Done()
	}()
}
