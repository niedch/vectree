package visualize

import (
	"sync"

	"github.com/niedch/vectree/internal/ai"
	"github.com/niedch/vectree/internal/datastore"
)

const pageSize = 500

type Server struct {
	ds       datastore.Querier
	embedder ai.EmbeddingModel
	limit    int
	pcaModel *PCAModel
	mu       sync.RWMutex
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
