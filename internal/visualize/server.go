package visualize

import (
	"context"
	"embed"
	"net/http"
	"strconv"

	"github.com/niedch/vectree/internal/ai"
	"github.com/niedch/vectree/internal/datastore"
)

//go:embed templates/index.html
var templatesFS embed.FS

func NewServer(ds datastore.Querier, embedder ai.EmbeddingModel, limit int) *Server {
	return &Server{
		ds:       ds,
		embedder: embedder,
		limit:    limit,
	}
}

func (s *Server) Start(addr string) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/embeddings", s.handleAPIEmbeddings)
	mux.HandleFunc("/api/project-prompt", s.handleProjectPrompt)
	mux.HandleFunc("/api/document/", s.handleGetDocument)

	return http.ListenAndServe(addr, mux)
}

func (s *Server) fetchEmbeddings(ctx context.Context) ([]datastore.DocumentWithEmbedding, error) {
	total, err := s.ds.CountDocumentEmbeddings(ctx)
	if err != nil {
		return nil, err
	}

	if total == 0 {
		return nil, nil
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
			return nil, err
		}
		allRecords = append(allRecords, page...)
	}

	return allRecords, nil
}

func (s *Server) computeResponseItems(records []datastore.DocumentWithEmbedding, reduced [][3]float64) []EmbeddingResponseItem {
	response := make([]EmbeddingResponseItem, len(reduced))
	for i, point := range reduced {
		response[i] = EmbeddingResponseItem{
			Id:       strconv.Itoa(records[i].Document.Id),
			ParentId: intPtrToString(records[i].Document.ParentId),
			X:        point[0],
			Y:        point[1],
			Z:        point[2],
			Text:     truncateText(records[i].Document.Document, 200),
		}
	}
	return response
}
