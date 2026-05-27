package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/niedch/vectree/internal/ai"
	"github.com/niedch/vectree/internal/conf"
	"github.com/niedch/vectree/internal/datastore"
	"github.com/niedch/vectree/internal/mcptemplate"
	"github.com/niedch/vectree/internal/prompt"
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
		s := server.NewMCPServer(
			"Documentation RAG",
			"1.0.0",
			server.WithToolCapabilities(false),
			server.WithRecovery(),
		)

		config, err := conf.Load()
		if err != nil {
			log.Fatal("Error loading config: ", err)
		}

		db, err := datastore.OpenConnection(config)
		if err != nil {
			log.Fatalln(err)
		}
		ds := datastore.NewSqliteDatastore(db)

		model, err := ai.NewGeminiEmbedder(context.Background(), config.AI.AsGeminiProviderConfig())
		if err != nil {
			log.Fatal("Error initializing embedding model: ", err)
		}

		// Add documentation search tool
		researchTool := mcp.NewTool("search-documentation",
			mcp.WithDescription(`Search ingested documentation including user guides and developer documentation. 
Use this tool to find information about features, configuration, API usage, troubleshooting, and development guidelines. 
Performs semantic search to find the most relevant documentation sections.`),
			mcp.WithString("search-string",
				mcp.Required(),
				mcp.Description(`A natural language query describing what you want to know. 
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
					Document:       *parentDoc,
					Embedding:      nil,
					EmbeddingRowid: 0,
				},
			}

			// Use the same template as search results for consistency
			resultString := mcptemplate.BuildResponseString(parentDocs)

			return mcp.NewToolResultText(resultString), nil
		})

		// Load custom prompts from dotprompt library
		if config.Prompts.Path != "" {
			customPrompts, err := prompt.LoadDir(config.Prompts.Path)
			if err != nil {
				log.Fatal("Error loading prompts: ", err)
			}

			for _, p := range customPrompts {
				desc := p.Description
				if desc == "" {
					desc = fmt.Sprintf("Custom prompt: %s", p.Name)
				}

				opts := []mcp.PromptOption{
					mcp.WithPromptDescription(desc),
				}

				for _, arg := range p.Arguments {
					argOpts := []mcp.ArgumentOption{}
					if arg.Description != "" {
						argOpts = append(argOpts, mcp.ArgumentDescription(arg.Description))
					}

					if arg.Required {
						argOpts = append(argOpts, mcp.RequiredArgument())
					}

					opts = append(opts, mcp.WithArgument(arg.Name, argOpts...))
				}

				s.AddPrompt(mcp.NewPrompt(p.Name, opts...), func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
					return mcp.NewGetPromptResult("",
						[]mcp.PromptMessage{
							mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(p.Source)),
						},
					), nil
				})
			}
		}

		// Start the server
		if err := server.ServeStdio(s); err != nil {
			fmt.Printf("Server error: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}
