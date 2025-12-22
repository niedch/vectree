package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "connectall-doc-rag",
	Short: "A document ingestion and RAG (Retrieval-Augmented Generation) system for ConnectAll documentation",
	Long: `connectall-doc-rag is a command-line tool for building and serving a RAG system for ConnectAll documentation.

This tool provides two main capabilities:

1. Document Ingestion (ingest command):
   - Crawls ConnectAll documentation from multiple sources
   - Splits documents into semantic chunks using header-based splitting
   - Generates vector embeddings using Google's Gemini embedding model
   - Stores embeddings in a SQLite vector database for efficient retrieval

2. MCP Server (mcp command):
   - Exposes documentation search via the Model Context Protocol (MCP)
   - Enables AI assistants to search and retrieve relevant documentation
   - Performs semantic search using vector similarity
   - Returns contextually relevant documentation chunks

The system uses a pipeline architecture for efficient parallel processing of documents
and supports both web-based documentation crawling and local markdown file ingestion.`,
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
