package visualize

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func (s *Server) handleProjectPrompt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ProjectPromptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Prompt == "" {
		http.Error(w, "Prompt is required", http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	pcaModel := s.pcaModel
	s.mu.RUnlock()

	if pcaModel == nil {
		http.Error(w, "PCA model not yet initialized. Please load the main visualization first.", http.StatusServiceUnavailable)
		return
	}

	ctx := context.Background()

	emb, err := s.embedder.GenerateEmbedding(ctx, req.Prompt)
	if err != nil {
		http.Error(w, "Embedding generation failed", http.StatusInternalServerError)
		return
	}

	point, err := pcaModel.Project(emb)
	if err != nil {
		http.Error(w, "Projection failed", http.StatusInternalServerError)
		return
	}

	nearest := s.fetchNearestDocuments(ctx, emb, pcaModel)

	resp := ProjectPromptResponse{
		X:                point[0],
		Y:                point[1],
		Z:                point[2],
		Text:             fmt.Sprintf("PROMPT: %s", req.Prompt),
		NearestDocuments: nearest,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) fetchNearestDocuments(ctx context.Context, emb []float32, pcaModel *PCAModel) []NearestDocument {
	nearestDocs, err := s.ds.SearchSimilarEmbeddings(ctx, emb, 3)
	if err != nil {
		return nil
	}

	nearest := make([]NearestDocument, 0, len(nearestDocs))
	for _, doc := range nearestDocs {
		neighborPoint, err := pcaModel.Project(doc.Embedding)
		if err != nil {
			continue
		}

		nearest = append(nearest, NearestDocument{
			Id:       strconv.Itoa(doc.Document.Id),
			ParentId: intPtrToString(doc.Document.ParentId),
			Text:     doc.Document.Document,
			X:        neighborPoint[0],
			Y:        neighborPoint[1],
			Z:        neighborPoint[2],
		})
	}
	return nearest
}
