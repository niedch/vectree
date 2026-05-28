package visualize

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/niedch/vectree/internal/ai"
	"github.com/niedch/vectree/internal/datastore"
)

//go:embed templates/index.html
var templatesFS embed.FS

type Server struct {
	ds       datastore.Querier
	embedder ai.EmbeddingModel
	limit    int
	pcaModel *PCAModel
	mu       sync.RWMutex
}

func NewServer(ds datastore.Querier, embedder ai.EmbeddingModel, limit int) *Server {
	return &Server{
		ds:       ds,
		embedder: embedder,
		limit:    limit,
	}
}

type EmbeddingResponseItem struct {
	Id       string  `json:"id"`
	ParentId *string `json:"parent_id,omitempty"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Z        float64 `json:"z"`
	Text     string  `json:"text"`
}

type DocumentDetailResponse struct {
	Id       string  `json:"id"`
	Content  string  `json:"content"`
	Level    int     `json:"level"`
	ParentId *string `json:"parent_id,omitempty"`
}

type ProjectPromptRequest struct {
	Prompt string `json:"prompt"`
}

type NearestDocument struct {
	Id       string  `json:"id"`
	ParentId *string `json:"parent_id,omitempty"`
	Text     string  `json:"text"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Z        float64 `json:"z"`
}

type ProjectPromptResponse struct {
	X                float64           `json:"x"`
	Y                float64           `json:"y"`
	Z                float64           `json:"z"`
	Text             string            `json:"text"`
	NearestDocuments []NearestDocument `json:"nearest_documents"`
}

func (s *Server) Start(addr string) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/embeddings", s.handleAPIEmbeddings)
	mux.HandleFunc("/api/project-prompt", s.handleProjectPrompt)
	mux.HandleFunc("/api/document/", s.handleGetDocument)

	return http.ListenAndServe(addr, mux)
}

func intPtrToString(v *int) *string {
	if v == nil {
		return nil
	}
	s := strconv.Itoa(*v)
	return &s
}

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

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/index.html" {
		if r.URL.Path != "/favicon.ico" {
			http.NotFound(w, r)
		}
		return
	}

	content, err := templatesFS.ReadFile("templates/index.html")
	if err != nil {
		http.Error(w, "Failed to load index.html", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(content)
}

const pageSize = 500

func (s *Server) handleAPIEmbeddings(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	total, err := s.ds.CountDocumentEmbeddings(ctx)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if total == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}

	fetchLimit := total
	if s.limit > 0 && s.limit < total {
		fetchLimit = s.limit
	}

	var allRecords []datastore.DocumentWithEmbedding
	for offset := 0; offset < fetchLimit; offset += pageSize {
		remaining := fetchLimit - offset
		pageLimit := pageSize
		if remaining < pageLimit {
			pageLimit = remaining
		}

		page, err := s.ds.GetDocumentEmbeddingsPage(ctx, pageLimit, offset)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		allRecords = append(allRecords, page...)
	}

	var vectors [][]float32
	for i := range allRecords {
		vectors = append(vectors, allRecords[i].Embedding)
	}

	reduced, model, err := ReduceTo3D(vectors)
	if err != nil {
		http.Error(w, "Reduction failed", http.StatusInternalServerError)
		return
	}

	s.mu.Lock()
	s.pcaModel = model
	s.mu.Unlock()

	var response []EmbeddingResponseItem
	for i, point := range reduced {
		text := allRecords[i].Document.Document
		if len(text) > 200 {
			text = text[:197] + "..."
		}

		response = append(response, EmbeddingResponseItem{
			Id:       strconv.Itoa(allRecords[i].Document.Id),
			ParentId: intPtrToString(allRecords[i].Document.ParentId),
			X:        point[0],
			Y:        point[1],
			Z:        point[2],
			Text:     text,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

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

	nearestDocs, err := s.ds.SearchSimilarEmbeddings(ctx, emb, 3)
	if err != nil {
		nearestDocs = nil
	}

	var nearest []NearestDocument
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

	resp := ProjectPromptResponse{
		X:    point[0],
		Y:    point[1],
		Z:    point[2],
		Text: fmt.Sprintf("PROMPT: %s", req.Prompt),
		NearestDocuments: nearest,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
