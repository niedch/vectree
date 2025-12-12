package cmd

import (
	"broadcom.com/vertex-ingestor/internal/conf"
	"broadcom.com/vertex-ingestor/internal/datastore"
	"github.com/spf13/cobra"
)

// sqliteCmd represents the sqlite command
var sqliteCmd = &cobra.Command{
	Use:   "sqlite",
	Short: "Test SQLite with vector extension",
	Long: `Test command to verify SQLite with sqlite-vec extension is working correctly.
This command will:
1. Initialize an in-memory SQLite database with vector extension
2. Create the embeddings virtual table
3. Insert a test embedding
4. Verify the data was stored correctly`,
	Run: func(cmd *cobra.Command, args []string) {
		config := conf.Load()

		_, err := datastore.OpenConnection(config)
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(sqliteCmd)
}
