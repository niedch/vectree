/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/niedch/tree-rag/internal/conf"
	"github.com/niedch/tree-rag/internal/pipeline"
	"github.com/spf13/cobra"
)

// ingestDebugCmd represents the ingestDebug command
var ingestDebugCmd = &cobra.Command{
	Use:   "ingestDebug",
	Short: "Debug command to test the document loading pipeline",
	Long: `The ingestDebug command is a development/debugging tool that tests the document loading pipeline.

This command runs a simplified pipeline that:
- Loads markdown files from the test_data/embedder directory
- Outputs debug information about the loaded files
- Does not perform embedding generation or database storage

This is useful for:
- Testing the file loading stage in isolation
- Debugging document processing issues
- Verifying file paths and content loading

Example:
  connectall-doc-rag ingestDebug`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		config, err := conf.Load()
		if err != nil {
			log.Fatal("Error loading config: ", err)
		}

		pipelines, err := pipeline.NewPipelineBuilder(config).BuildAll()
		if err != nil {
			log.Fatalf("Failed to build Pipeline: %e", err)
		}

		ingestionPipeline := pipeline.NewPipeline()
		ingestionPipeline.AddStage(pipeline.MergePipelines(pipelines))

		out := ingestionPipeline.Run(ctx)
		for i := range out {
			fmt.Printf("%s", i)
		}

		// p := pipeline.NewPipeline()
		// p.AddStage(pipeline.TypedStage(stages.NewDirLoader("./test_data/embedder")))
		// p.AddStage(pipeline.TypedStage(stages.NewDebugStage()))
		//
		// out := p.Run(ctx)
		// for range out {
		// 	// Pipeline execution happens as we consume the output
		// }
	},
}

func init() {
	rootCmd.AddCommand(ingestDebugCmd)
}
