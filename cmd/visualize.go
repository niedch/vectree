package cmd

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"

	"github.com/niedch/vectree/internal/ai"
	"github.com/niedch/vectree/internal/conf"
	"github.com/niedch/vectree/internal/datastore"
	"github.com/niedch/vectree/internal/visualize"
	"github.com/spf13/cobra"
)

var (
	visualizePort  int
	visualizeLimit int
)

var visualizeCmd = &cobra.Command{
	Use:   "visualize",
	Short: "Start the 3D embedding visualization server",
	Long: `Start a local web server to visualize ingested document embeddings in 3D using PCA.

This command:
- Loads all embeddings from the SQLite database
- Reduces them to 3D using Principal Component Analysis (PCA)
- Serves a web UI with an interactive Plotly.js 3D scatter plot
- Allows projecting natural language prompts into the embedding space

The visualization server provides:
  - GET  /  - Web UI with 3D scatter plot
  - GET  /api/embeddings  - All embeddings reduced to 3D
  - POST /api/project-prompt  - Project a prompt into the 3D space
  - GET  /api/document/{id}  - Get full document content

Example:
  vectree visualize --port 8090 --limit 1000`,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := conf.Load()
		if err != nil {
			log.Fatal("Error loading config: ", err)
		}

		db, err := datastore.OpenConnection(config)
		if err != nil {
			log.Fatalln(err)
		}
		ds := datastore.NewSqliteDatastore(db)

		model, err := ai.NewGeminiEmbedder(cmd.Context(), config.AI.AsGeminiProviderConfig())
		if err != nil {
			log.Fatal("Error initializing embedding model: ", err)
		}

		server := visualize.NewServer(ds, model, visualizeLimit)
		addr := fmt.Sprintf(":%d", visualizePort)

		url := fmt.Sprintf("http://localhost%s", addr)
		fmt.Printf("Visualization server starting on %s\n", url)
		go openBrowser(url)
		if err := server.Start(addr); err != nil {
			log.Fatal("Server failed: ", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(visualizeCmd)

	visualizeCmd.Flags().IntVarP(&visualizePort, "port", "p", 8090, "Port to run the visualization server on")
	visualizeCmd.Flags().IntVarP(&visualizeLimit, "limit", "l", 1000, "Maximum number of embeddings to visualize")
}


func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	cmd.Start()
}

