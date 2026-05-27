package cmd

import (
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/server"
	"github.com/niedch/vectree/internal/ai"
	"github.com/niedch/vectree/internal/conf"
	"github.com/niedch/vectree/internal/datastore"
	"github.com/niedch/vectree/internal/mcpserver"
	"github.com/spf13/cobra"
)

// mcpCmd represents the mcp command
var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start the MCP server for documentation search",
	Long: `Start the Model Context Protocol (MCP) server that provides AI-powered 
search capabilities for ingested documentation.

The MCP server exposes tools and prompts for searching and navigating documentation:

Tools:
   1. search-documentation
      - Performs semantic search across all ingested documentation
      - Uses vector similarity to find relevant content
      - Returns ranked results with document IDs and metadata
      - Accepts natural language queries

   2. get-parent-context
      - Retrieves the parent document for a given document ID
      - Useful for understanding broader context of search results
      - Returns the parent section/heading that contains the document

Prompts:
   - Prompts are loaded from the configured dotprompt library directory
   - Each .prompt file in the library becomes a prompt with its defined
     arguments and description
   - Documentation prompts guide LLM usage of the search tools

The server communicates via stdio and can be integrated with MCP-compatible 
clients like Claude Desktop, Zed, or other AI assistants.

Configuration:
- AI provider configured in config.toml (gemini, openai, or ollama)
- API keys loaded from environment variables (GEMINI_API_KEY, OPENAI_API_KEY)
- Uses the same embedding model as ingestion for query encoding
- Similarity results count configurable via config

Example:
  vectree mcp`,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := conf.Load()
		if err != nil {
			log.Fatal("Error loading config: ", err)
		}

		db, err := datastore.OpenConnection(config)
		if err != nil {
			log.Fatalln(err)
		}
		ds := datastore.NewSqliteDatastore(db)

		model, err := ai.NewGeminiEmbedder(cmd.Context(), config.AI.AsGeminiProviderConfig())
		if err != nil {
			log.Fatal("Error initializing embedding model: ", err)
		}

		s := mcpserver.New(mcpserver.Dependencies{
			Config:   config,
			Querier:  ds,
			Embedder: model,
		})

		if err := server.ServeStdio(s); err != nil {
			fmt.Printf("Server error: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}
