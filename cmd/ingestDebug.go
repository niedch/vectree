/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"

	"broadcom.com/vertex-ingestor/internal/pipeline"
	"broadcom.com/vertex-ingestor/internal/stages"
	"github.com/spf13/cobra"
)

// ingestDebugCmd represents the ingestDebug command
var ingestDebugCmd = &cobra.Command{
	Use:   "ingestDebug",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		p := pipeline.NewPipeline()
		p.AddStage(pipeline.TypedStage(stages.NewDirLoader("./test_data/embedder")))
		p.AddStage(pipeline.TypedStage(stages.NewDebugStage()))

		out := p.Run(ctx)
		for range out {
			// Pipeline execution happens as we consume the output
		}
	},
}

func init() {
	rootCmd.AddCommand(ingestDebugCmd)
}
