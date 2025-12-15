package cmd

import (
	"context"
	"log"

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

		ctx := context.Background()
		db, err := datastore.OpenConnection(config)
		if err != nil {
			panic(err)
		}
		defer db.Close()

		var vecVersion string
		err = db.QueryRow("select vec_version()").Scan(&vecVersion)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("vec_version=%s\n", vecVersion)

		document := datastore.Document{Document: "Test Insert"}
		embedding := datastore.Embedding{Embedding: []float32{0.1, 0.2, 0.3}}

		ds := datastore.NewSqliteDatastore(db)
		docId, err := ds.InsertDocument(ctx, document, embedding)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Inserted document with ID: %d\n", docId)
	},
}

func init() {
	rootCmd.AddCommand(sqliteCmd)
}
