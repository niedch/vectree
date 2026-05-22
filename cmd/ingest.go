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

const (
	TOC_URL = "https://techdocs.broadcom.com/us/en/ca-enterprise-software/valueops/connectall/4-0/jcr:content.toc.html"
)

var ingestCmd = &cobra.Command{
	Use:   "ingest",
	Short: "Ingest ConnectAll documentation and generate vector embeddings",
	Long: `The ingest command processes ConnectAll documentation from multiple sources and stores vector embeddings.

This command runs two parallel ingestion pipelines:

1. Web Documentation Pipeline:
   - Fetches the table of contents from Broadcom TechDocs (techdocs.broadcom.com)
   - Downloads all linked documentation pages
   - Extracts content from HTML <main> tags
   - Converts HTML to Markdown format

2. Local Markdown Pipeline:
   - Scans local markdown files in the ../connectall directory
   - Filters out node_modules and other irrelevant files
   - Loads markdown file content

Both pipelines then:
   - Split documents using Markdown AST-based header splitting
   - Create overlapping sections with parent context for better retrieval
   - Batch documents for efficient processing (configurable batch size)
   - Generate embeddings using Google Gemini embedding model
   - Store chunks with embeddings in SQLite vector database
   - Maintain parent-child relationships between document sections

Configuration:
- Embedding model: Configurable via config (default: text-embedding-004)
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

		pipelines, err := pipeline.NewPipelineBuilder(config).BuildAll()
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
