package mcpserver

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/niedch/vectree/internal/ai"
	"github.com/niedch/vectree/internal/conf"
	"github.com/niedch/vectree/internal/datastore"
	"github.com/niedch/vectree/internal/mcptemplate"
)

type handler struct {
	querier  datastore.Querier
	embedder ai.EmbeddingModel
	config   *conf.Config
}

func (h *handler) searchDocumentation(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	searchString, err := request.RequireString("search-string")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	emb, err := h.embedder.GenerateEmbedding(ctx, searchString)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	docs, err := h.querier.SearchSimilarEmbeddings(ctx, emb, h.config.Retrieval.SimilarityResults)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	resultString := mcptemplate.BuildResponseString(docs)

	return mcp.NewToolResultText(resultString), nil
}

func (h *handler) getParentContext(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	documentId, err := request.RequireInt("document-id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	doc, err := h.querier.GetDocument(ctx, documentId)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Document not found: %v", err)), nil
	}

	if doc.ParentId == nil {
		return mcp.NewToolResultText(fmt.Sprintf(
			"Document %d is a root document (Level %d heading) and has no parent context.\n\n"+
				"**Document Content:**\n\n%s",
			doc.Id, doc.Level, doc.Document)), nil
	}

	parentDoc, err := h.querier.GetParentDocument(ctx, documentId)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Could not find parent document: %v", err)), nil
	}

	parentDocs := []datastore.DocumentWithEmbedding{
		{
			Document:       *parentDoc,
			Embedding:      nil,
			EmbeddingRowid: 0,
		},
	}

	resultString := mcptemplate.BuildResponseString(parentDocs)

	return mcp.NewToolResultText(resultString), nil
}

func registerTools(s *server.MCPServer, deps Dependencies) {
	h := &handler{
		querier:  deps.Querier,
		embedder: deps.Embedder,
		config:   deps.Config,
	}

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

	s.AddTool(researchTool, h.searchDocumentation)

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

	s.AddTool(contextTool, h.getParentContext)
}
