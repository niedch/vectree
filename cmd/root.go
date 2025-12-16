package cmd

import (
	"os"
	"strings"

	"broadcom.com/vertex-ingestor/internal/datastore"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "rag-vec-search",
	Short: "A CLI tool for ingesting documents, generating embeddings, and searching a vector database.",
	Long: `rag-vec-search is a command-line tool that provides functionalities for managing a vector search system.

It allows you to:
- Generate vector embeddings for the documents using a generative AI model.
- Store the embeddings in a Weaviate vector database.
- Search the vector database with a given query to retrieve relevant document chunks.
- Generate a response from a prompt enhanced with the search results.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func buildResponseString(docs []datastore.DocumentWithEmbedding) string {
	sb := strings.Builder{}
	for _, doc := range docs {
		sb.WriteString(doc.Document.Document)
		sb.WriteString("\n")
	}
	return sb.String()
}
