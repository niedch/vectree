/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"net/http"

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
		http.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Hello")
		})

		log.Fatal(http.ListenAndServe(":8080", nil))
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}
