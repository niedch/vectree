package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"broadcom.com/vertex-ingestor/internal/ai"
	"broadcom.com/vertex-ingestor/internal/conf"
	"broadcom.com/vertex-ingestor/internal/datastore"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "rag-vec-search",
	Short: "A CLI tool for ingesting documents, generating embeddings, and searching a vector database.",
	Long: `rag-vec-search is a command-line tool that provides functionalities for managing a vector search system.

It allows you to:
- Generate vector embeddings for the documents using a generative AI model.
- Store the embeddings in a Weaviate vector database.
- Search the vector database with a given query to retrieve relevant document chunks.
- Generate a response from a prompt enhanced with the search results.`,
	Run: func(cmd *cobra.Command, args []string) {
		s := server.NewMCPServer(
			"ConnectAll documentation",
			"1.0.0",
			server.WithToolCapabilities(false),
			server.WithRecovery(),
		)

		config := conf.Load()
		db, err := datastore.OpenConnection(config)
		if err != nil {
			log.Fatalln(err)
		}
		ds := datastore.NewSqliteDatastore(db)

		model := ai.NewGeminiEmbedder("embedding-001")

		// Add a calculator tool
		researchTool := mcp.NewTool("search-documentation",
			mcp.WithDescription("Allows you to search the connectall documentation for Relevant Information for you."),
			mcp.WithString("search-string",
				mcp.Required(),
				mcp.Description("The search-string that you want to search for!"),
			),
		)

		s.AddTool(researchTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			searchString, err := request.RequireString("search-string")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			model.Initialize(ctx)

			emb, err := model.GenerateEmbedding(ctx, searchString)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			docs, err := ds.SearchSimilarEmbeddings(ctx, emb, 3)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			resultString := buildResponseString(docs)

			return mcp.NewToolResultText(resultString), nil
		})

		// Start the server
		if err := server.ServeStdio(s); err != nil {
			fmt.Printf("Server error: %v\n", err)
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func buildResponseString(docs []datastore.DocumentWithEmbedding) string {
	sb := strings.Builder{}
	for _, doc := range docs {
		sb.WriteString(doc.Document.Document)
		sb.WriteString("\n")
	}
	return sb.String()
}
