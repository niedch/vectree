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

		p1 := pipeline.New(stages.NewDirLoader("./test_data/embedder"))
		p2 := pipeline.AddStage(p1, stages.NewNodeModulesFilter())
		p3 := pipeline.AddStage(p2, stages.NewFileLoader())
		p4 := pipeline.AddStage(p3, stages.NewMdAstSplitter())
		p5 := pipeline.AddStage(p4, stages.NewDebugStage())

		p5.Run(ctx);
	},
}

func init() {
	rootCmd.AddCommand(ingestDebugCmd)
}
