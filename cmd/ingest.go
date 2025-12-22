package cmd

import (
	"context"

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

		embedder := ai.NewGeminiEmbedder(config.GEMINI_API_KEY, config.AI.EmbeddingModel)
		embedder.Initialize(ctx)

		db, err := datastore.OpenConnection(config)
		if err != nil {
			panic(err)
		}

		ds := datastore.NewSqliteDatastore(db)
		store := store.NewSqliteStore(ds)

		markdownFilesP := pipeline.NewPipeline()
		markdownFilesP.AddStage(pipeline.TypedStage(stages.NewDirLoader("../connectall")))
		markdownFilesP.AddStage(pipeline.TypedStage(stages.NewNodeModulesFilter()))
		markdownFilesP.AddStage(pipeline.TypedStage(stages.NewFileLoader()))

		docuP := pipeline.NewPipeline()
		docuP.AddStage(pipeline.TypedStage(stages.NewDocTocLoader(TOC_URL)))
		docuP.AddStage(pipeline.TypedStage(stages.NewDebugStage()))
		docuP.AddStage(pipeline.TypedStage(stages.NewContentLoader(config.Pipeline.DocuLoaderWorkers)))

		main := pipeline.NewPipeline()
		main.AddStage(pipeline.MergePipelines(markdownFilesP, docuP))
		main.AddStage(pipeline.TypedStage(stages.NewMdAstSplitter()))
		main.AddStage(pipeline.TypedStage(stages.NewBatcher[string](config.Pipeline.EmbedderBatchSize)))
		main.AddStage(pipeline.TypedStage(stages.NewEmbedder(embedder, config.Pipeline.EmbedderWorkers)))
		main.AddStage(pipeline.TypedStage(stages.NewBatcher[*stages.EmbedderOut](config.Pipeline.StoreBatchSize)))
		main.AddStage(pipeline.TypedStage(stages.NewStore(store)))

		out := main.Run(ctx)
		for range out { }
	},
}

func init() {
	rootCmd.AddCommand(ingestCmd)
}
