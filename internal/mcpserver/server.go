package mcpserver

import (
	"github.com/mark3labs/mcp-go/server"
	"github.com/niedch/vectree/internal/ai"
	"github.com/niedch/vectree/internal/conf"
	"github.com/niedch/vectree/internal/datastore"
)

type Dependencies struct {
	Config   *conf.Config
	Querier  datastore.Querier
	Embedder ai.EmbeddingModel
}

func New(deps Dependencies) *server.MCPServer {
	s := server.NewMCPServer(
		"Documentation RAG",
		"1.0.0",
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)

	registerTools(s, deps)
	registerPrompts(s, deps.Config.Prompts.Path)

	return s
}
