/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	_ "github.com/mattn/go-sqlite3"

	"broadcom.com/vertex-ingestor/internal/datastore"
	"github.com/spf13/cobra"
)

// sqliteCmd represents the sqlite command
var sqliteCmd = &cobra.Command{
	Use:   "sqlite",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("sqlite called")
		ctx := context.Background()

		sqlite_vec.Auto()
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			panic(err)
		}

		queries := datastore.New(db)
		check_version, err := queries.CheckVersion(ctx)
		if err != nil {
			panic(err)
		}

		log.Println(check_version)
	},
}

func init() {
	rootCmd.AddCommand(sqliteCmd)
}
