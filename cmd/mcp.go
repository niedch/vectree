package cmd

import (
	"context"
	"fmt"
	"log"

	"broadcom.com/vertex-ingestor/internal/ai"
	"broadcom.com/vertex-ingestor/internal/conf"
	"broadcom.com/vertex-ingestor/internal/datastore"
	"broadcom.com/vertex-ingestor/internal/mcptemplate"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

// mcpCmd represents the mcp command
var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start the MCP server for ConnectAll documentation search",
	Long: `Start the Model Context Protocol (MCP) server that provides AI-powered 
search capabilities for ConnectAll documentation.

The MCP server exposes a 'search-documentation' tool that uses semantic search
with embeddings to find relevant documentation based on natural language queries.
The server communicates via stdio and can be integrated with MCP-compatible clients.

Example:
  connectall-doc-rag mcp`,
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

		model := ai.NewGeminiEmbedder(config.GEMINI_API_KEY, config.AI.EmbeddingModel)

		// Add documentation search tool
		researchTool := mcp.NewTool("search-documentation",
			mcp.WithDescription(`Search ConnectAll documentation including official user guides and internal developer documentation. 
Use this tool to find information about ConnectAll features, configuration, API usage, troubleshooting, and development guidelines. 
Performs semantic search to find the most relevant documentation sections.`),
			mcp.WithString("search-string",
				mcp.Required(),
				mcp.Description(`A natural language query describing what you want to know about ConnectAll. 
Examples: 'How to configure authentication', 
					'API endpoints for integration', 
					'troubleshooting connection errors', 
					'developer setup guide'`),
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

			docs, err := ds.SearchSimilarEmbeddings(ctx, emb, config.Retrieval.SimilarityResults)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

		resultString := mcptemplate.BuildResponseString(docs)

		return mcp.NewToolResultText(resultString), nil
	})

	// Add tool to get parent context for a document
	contextTool := mcp.NewTool("get-parent-context",
		mcp.WithDescription(`Get the parent context for a specific document. When you find a relevant document 
in search results, you can use this tool to get its parent document for broader context. 
For example, if you find a document at level 3, you can request its parent at level 2 to understand 
the broader topic or section it belongs to.`),
		mcp.WithNumber("document-id",
			mcp.Required(),
			mcp.Description("The ID of the document whose parent context you want to retrieve"),
		),
	)

	s.AddTool(contextTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		documentId, err := request.RequireInt("document-id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// First, get the requested document to check if it's a root document
		doc, err := ds.GetDocument(ctx, documentId)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Document not found: %v", err)), nil
		}

		// Check if this document is a root document (no parent)
		if doc.ParentId == nil {
			return mcp.NewToolResultText(fmt.Sprintf(
				"Document %d is a root document (Level %d heading) and has no parent context.\n\n"+
					"**Document Content:**\n\n%s",
				doc.Id, doc.Level, doc.Document)), nil
		}

		// Get the parent document
		parentDoc, err := ds.GetParentDocument(ctx, documentId)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Could not find parent document: %v", err)), nil
		}

		// Wrap the parent document in a DocumentWithEmbedding slice for template consistency
		// Note: We don't need the embedding for display, so we can leave it empty
		parentDocs := []datastore.DocumentWithEmbedding{
			{
				Document: *parentDoc,
				Embedding: nil,
				EmbeddingRowid: 0,
			},
		}

		// Use the same template as search results for consistency
		resultString := mcptemplate.BuildResponseString(parentDocs)

		return mcp.NewToolResultText(resultString), nil
	})

	// Add prompts to guide LLM usage
		s.AddPrompt(mcp.NewPrompt("connectall-help",
			mcp.WithPromptDescription("Get help with ConnectAll features, configuration, or troubleshooting"),
			mcp.WithArgument("topic",
				mcp.ArgumentDescription("The specific topic or area you need help with (e.g., authentication, integrations, API)"),
				mcp.RequiredArgument(),
			),
		), func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			topic := request.Params.Arguments["topic"]

			message := fmt.Sprintf(`I need help with ConnectAll. Please search the documentation for information about: %s

Use the search-documentation tool to find relevant information from both official user guides and internal developer documentation.`, topic)

			return mcp.NewGetPromptResult("",
				[]mcp.PromptMessage{
					mcp.NewPromptMessage(
						mcp.RoleUser,
						mcp.NewTextContent(message),
					),
				},
			), nil
		})

		s.AddPrompt(mcp.NewPrompt("connectall-troubleshoot",
			mcp.WithPromptDescription("Troubleshoot ConnectAll issues by searching documentation for solutions"),
			mcp.WithArgument("issue",
				mcp.ArgumentDescription("Description of the issue or error you're experiencing"),
				mcp.RequiredArgument(),
			),
		), func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			issue := request.Params.Arguments["issue"]

			message := fmt.Sprintf(`I'm experiencing an issue with ConnectAll: %s

Please search the documentation for troubleshooting information, common solutions, and relevant configuration details.`, issue)

			return mcp.NewGetPromptResult("",
				[]mcp.PromptMessage{
					mcp.NewPromptMessage(
						mcp.RoleUser,
						mcp.NewTextContent(message),
					),
				},
			), nil
		})

		s.AddPrompt(mcp.NewPrompt("connectall-develop",
			mcp.WithPromptDescription("Find developer documentation for building with ConnectAll"),
			mcp.WithArgument("dev-topic",
				mcp.ArgumentDescription("What you want to develop or integrate (e.g., custom adapter, API integration, plugin)"),
				mcp.RequiredArgument(),
			),
		), func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			devTopic := request.Params.Arguments["dev-topic"]

			message := fmt.Sprintf(`I'm developing with ConnectAll and need information about: %s

Please search the internal developer documentation and API guides for relevant information and best practices.`, devTopic)

			return mcp.NewGetPromptResult("",
				[]mcp.PromptMessage{
					mcp.NewPromptMessage(
						mcp.RoleUser,
						mcp.NewTextContent(message),
					),
				},
			), nil
		})

		// Start the server
		if err := server.ServeStdio(s); err != nil {
			fmt.Printf("Server error: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}
