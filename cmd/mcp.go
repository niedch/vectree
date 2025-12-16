/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log"

	"broadcom.com/vertex-ingestor/internal/ai"
	"broadcom.com/vertex-ingestor/internal/conf"
	"broadcom.com/vertex-ingestor/internal/datastore"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

// mcpCmd represents the mcp command
var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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

func init() {
	rootCmd.AddCommand(mcpCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// mcpCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// mcpCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
