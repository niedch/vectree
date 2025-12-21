package cmd

import (
	"context"
	"log"
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
	Short: "Ingests ConnectAll documentation and generates vector embeddings",
	Long: `The ingest command processes ConnectAll documentation from multiple sources and stores vector embeddings.

This command runs two parallel ingestion pipelines:

1. Documentation Pipeline:
   - Fetches the table of contents from Broadcom TechDocs
   - Downloads all linked documentation pages
   - Processes HTML content into text chunks
   - Generates embeddings for each chunk

2. Markdown Pipeline:
   - Scans local markdown files in the ../connectall directory
	- Header-based splitting for semantic coherence
   - Filters out node_modules and other irrelevant files
   - Splits documents using header-based chunking
   - Generates embeddings for each section

Both pipelines use:
- Google Gemini embedding-001 model for vector generation
- Batch processing (64 chunks per batch) for efficiency
- Parallel embedding generation (8 concurrent workers)
- SQLite database for persistent storage

The process runs asynchronously and waits for both pipelines to complete.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		config := conf.Load()

		log.Println(config)
		embedder := ai.NewGeminiEmbedder(config.GEMINI_API_KEY, config.AI.EmbeddingModel)
		embedder.Initialize(ctx)

		db, err := datastore.OpenConnection(config)
		if err != nil {
			panic(err)
		}

		ds := datastore.NewSqliteDatastore(db)
		store := store.NewSqliteStore(ds)

		var wg sync.WaitGroup

		// RunDocumentationPipelineAsync(ctx, &wg, config, embedder, store)
		RunMarkdownPipelineAsync(ctx, &wg, config, embedder, store)

		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(ingestCmd)
}

func RunMarkdownPipelineAsync(ctx context.Context, wg *sync.WaitGroup, config *conf.Config, embedder ai.EmbeddingModel, store store.Datastore) {
	wg.Add(1)

	go func() {
		p1 := pipeline.New(stages.NewDirLoader("../connectall"))
		p2 := pipeline.AddStage(p1, stages.NewNodeModulesFilter())
		p3 := pipeline.AddStage(p2, stages.NewFileLoader())
		p4 := pipeline.AddStage(p3, stages.NewHeaderSplitter())
		p5 := pipeline.AddStage(p4, stages.NewBatcher[string](config.Pipeline.EmbedderBatchSize))
		p6 := pipeline.AddStage(p5, stages.NewEmbedder(embedder, config.Pipeline.EmbedderWorkers))
		p7 := pipeline.AddStage(p6, stages.NewBatcher[*stages.EmbedderOut](config.Pipeline.StoreBatchSize))
		p8 := pipeline.AddStage(p7, stages.NewStore(store))
		p8.Run(ctx)

		wg.Done()
	}()
}

func RunDocumentationPipelineAsync(ctx context.Context, wg *sync.WaitGroup, config *conf.Config, embedder ai.EmbeddingModel, store store.Datastore) {
	wg.Add(1)

	go func() {
		p1 := pipeline.New(stages.NewDocTocLoader(TOC_URL))
		p2 := pipeline.AddStage(p1, stages.NewDebugStage())
		p3 := pipeline.AddStage(p2, stages.NewContentLoader(config.Pipeline.DocuLoaderWorkers))
		p4 := pipeline.AddStage(p3, stages.NewBatcher[string](config.Pipeline.EmbedderBatchSize))
		p5 := pipeline.AddStage(p4, stages.NewEmbedder(embedder, config.Pipeline.EmbedderWorkers))
		p6 := pipeline.AddStage(p5, stages.NewBatcher[*stages.EmbedderOut](config.Pipeline.StoreBatchSize))
		p7 := pipeline.AddStage(p6, stages.NewStore(store))
		p7.Run(ctx)

		wg.Done()
	}()
}
