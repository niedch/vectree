/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/niedch/vectree/internal/conf"
	"github.com/niedch/vectree/internal/pipeline"
	"github.com/niedch/vectree/internal/stages"
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
  vectree ingestDebug`,
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
		ingestionPipeline.AddStage(pipeline.TypedStage(stages.NewMdAstSplitter()))

		outDir := "output"
		if err := os.MkdirAll(outDir, 0755); err != nil {
			log.Fatal("Error creating output directory: ", err)
		}

		out := ingestionPipeline.Run(ctx)
		i := 0
		for output := range out {
			var content string
			switch v := output.(type) {
			case string:
				content = v
			case stages.SectionWithLevel:
				content = v.Text
			default:
				continue
			}
			filename := filepath.Join(outDir, fmt.Sprintf("output_%d.md", i))
			if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
				log.Printf("Error writing %s: %v", filename, err)
			}
			i++
		}

		log.Printf("Ingested: %d Documents", i)
	},
}

func init() {
	rootCmd.AddCommand(ingestDebugCmd)
}
