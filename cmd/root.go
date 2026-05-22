package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "vectree",
	Short: "Document ingestion and RAG system for ConnectAll documentation",
	Long: `vectree is a command-line tool for building and serving a RAG (Retrieval-Augmented Generation) 
system for ConnectAll documentation.

This tool provides three main commands:

1. ingest - Document Ingestion Pipeline
   - Crawls ConnectAll documentation from Broadcom TechDocs
   - Loads local markdown files from the ../connectall directory
   - Splits documents using Markdown AST-based header splitting
   - Generates vector embeddings using Google's Gemini embedding model
   - Stores embeddings in SQLite database with vec0 extension
   - Maintains hierarchical document relationships (parent-child)

2. mcp - Model Context Protocol Server
   - Exposes documentation search via the Model Context Protocol (MCP)
   - Provides 'search-documentation' tool for semantic search
   - Provides 'get-parent-context' tool for retrieving parent sections
   - Includes helpful prompts for common documentation tasks
   - Communicates via stdio for integration with AI assistants

3. ingestDebug - Development/Debug Command
   - Tests the document loading pipeline in isolation
   - Useful for debugging file loading and processing issues
   - Does not perform embedding generation or storage

Architecture:
- Pipeline-based processing with typed stages
- Parallel processing with configurable workers
- Batch processing for efficient embedding generation
- Vector similarity search using SQLite vec0 extension
- Hierarchical document structure with parent-child relationships

Use 'vectree [command] --help' for more information about a command.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Root command flags can be added here if needed
}
