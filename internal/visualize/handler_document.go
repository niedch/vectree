package visualize

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
)

func (s *Server) handleGetDocument(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/api/document/"):]
	if idStr == "" {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	doc, err := s.ds.GetDocument(ctx, id)
	if err != nil {
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	resp := DocumentDetailResponse{
		Id:       strconv.Itoa(doc.Id),
		Content:  doc.Document,
		Level:    doc.Level,
		ParentId: intPtrToString(doc.ParentId),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
