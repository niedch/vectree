package visualize

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/niedch/vectree/internal/datastore"
)

func (s *Server) handleAPIEmbeddings(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	records, err := s.fetchEmbeddings(ctx)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if len(records) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}

	vectors := extractVectors(records)

	reduced, model, err := ReduceTo3D(vectors)
	if err != nil {
		http.Error(w, "Reduction failed", http.StatusInternalServerError)
		return
	}

	s.mu.Lock()
	s.pcaModel = model
	s.mu.Unlock()

	response := s.computeResponseItems(records, reduced)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func extractVectors(records []datastore.DocumentWithEmbedding) [][]float32 {
	vectors := make([][]float32, len(records))
	for i := range records {
		vectors[i] = records[i].Embedding
	}
	return vectors
}
