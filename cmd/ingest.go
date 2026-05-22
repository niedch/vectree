package cmd

import (
	"context"
	"log"

	"github.com/niedch/vectree/internal/ai"
	"github.com/niedch/vectree/internal/conf"
	"github.com/niedch/vectree/internal/datastore"
	"github.com/niedch/vectree/internal/pipeline"
	"github.com/niedch/vectree/internal/stages"
	"github.com/niedch/vectree/internal/store"
	"github.com/spf13/cobra"
)

var ingestCmd = &cobra.Command{
	Use:   "ingest",
	Short: "Ingest documentation from configured sources and generate vector embeddings",
	Long: `The ingest command processes documentation from all sources defined in config.toml and stores vector embeddings.

Sources are defined in the [sources] section of config.toml:

1. Web Documentation Pipeline (type = "http"):
   - Crawls the configured URL recursively up to max_depth
   - Extracts content using the configured CSS selector
   - Converts HTML to Markdown format
   - Parallel crawling with configurable worker count

2. Local Markdown Pipeline (type = "markdown"):
   - Scans local markdown files in the configured location directory
   - Filters out node_modules and other irrelevant files
   - Loads markdown file content

All pipelines then:
   - Split documents using Markdown AST-based header splitting
   - Create overlapping sections with parent context for better retrieval
   - Batch documents for efficient processing (configurable batch size)
   - Generate embeddings using the configured AI provider
   - Store chunks with embeddings in SQLite vector database
   - Maintain parent-child relationships between document sections

Configuration:
- Embedding model: Configurable via config (default: text-embedding-004)
- AI provider: gemini, openai, or ollama (configurable)
- Batch size: Configurable for both embedding and storage stages
- Workers: Parallel embedding generation (default: 8 workers)
- Database: SQLite with vec0 extension for vector similarity search

Example:
  vectree ingest`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		config, err := conf.Load()
		if err != nil {
			log.Fatal("Error loading config: ", err)
		}

		embedder, err := ai.NewGeminiEmbedder(ctx, config.AI.AsGeminiProviderConfig())
		if err != nil {
			log.Fatal("Error initializing embedding model: ", err)
		}

		db, err := datastore.OpenConnection(config)
		if err != nil {
			panic(err)
		}

		ds := datastore.NewSqliteDatastore(db)
		store := store.NewSqliteStore(ds)

		pipelines, err := pipeline.NewPipelineBuilder(config).BuildSources()
		if err != nil {
			log.Fatalf("Failed to build Pipeline: %e", err)
		}

		main := pipeline.NewPipeline()
		main.AddStage(pipeline.MergePipelines(pipelines))
		main.AddStage(pipeline.TypedStage(stages.NewMdAstSplitter()))
		main.AddStage(pipeline.TypedStage(stages.NewBatcher[stages.SectionWithLevel](config.Pipeline.EmbedderBatchSize)))
		main.AddStage(pipeline.TypedStage(stages.NewEmbedder(embedder, config.Pipeline.EmbedderWorkers)))
		main.AddStage(pipeline.TypedStage(stages.NewBatcher[*stages.EmbedderOut](config.Pipeline.StoreBatchSize)))
		main.AddStage(pipeline.TypedStage(stages.NewStoreStage(store)))

		out := main.Run(ctx)
		for range out {
		}
	},
}

func init() {
	rootCmd.AddCommand(ingestCmd)
}
