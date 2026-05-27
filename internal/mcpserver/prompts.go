package mcpserver

import (
	"context"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/niedch/vectree/internal/prompt"
)

func registerPrompts(s *server.MCPServer, promptsDir string) {
	if promptsDir == "" {
		return
	}

	customPrompts, err := prompt.LoadDir(promptsDir)
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
