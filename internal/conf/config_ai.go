package conf

const (
	OPENAI_PROVIDER AIProvider = "openai"
	GEMINI_PROVIDER AIProvider = "gemini"
	OLLAMA_PROVIDER AIProvider = "ollama"
)
const (
	DEFAULT_GEMINI_VERTEX_SIZE = 768
)

type AIProvider string

type AIBase struct {
	Provider AIProvider
}

type AI struct {
	EmbeddingModel string     `koanf:"embedding_model"`
	VertexSize     int        `koanf:"vertex_size"`
	Provider       AIProvider `koanf:"provider"`
	URL            string     `koanf:"url"`
	GeminiAPIKey   string     `koanf:"gemini_api_key"`
	OpenAIAPIKey   string     `koanf:"openai_api_key"`
}

type GeminiProviderConfig struct {
	AIBase
	EmbeddingModel string `validate:"required"`
	VertexSize     int
	APIKey         string `validate:"required"`
}

func (a AI) AsGeminiProviderConfig() *GeminiProviderConfig {
	if a.Provider != GEMINI_PROVIDER {
		return nil
	}

	vertexSize := a.VertexSize
	if vertexSize == 0 {
		vertexSize = DEFAULT_GEMINI_VERTEX_SIZE
	}

	return &GeminiProviderConfig{
		AIBase:         AIBase{Provider: a.Provider},
		EmbeddingModel: a.EmbeddingModel,
		VertexSize:     vertexSize,
		APIKey:         a.GeminiAPIKey,
	}
}

type OpenAIProviderConfig struct {
	AIBase
	EmbeddingModel string `validate:"required"`
	URL            string `validate:"omitempty,url"`
	APIKey         string `validate:"required"`
}

func (a AI) AsOpenAIProviderConfig() *OpenAIProviderConfig {
	if a.Provider != OPENAI_PROVIDER {
		return nil
	}
	return &OpenAIProviderConfig{
		AIBase:         AIBase{Provider: a.Provider},
		EmbeddingModel: a.EmbeddingModel,
		URL:            a.URL,
		APIKey:         a.OpenAIAPIKey,
	}
}

type OllamaProviderConfig struct {
	AIBase
	EmbeddingModel string `validate:"required"`
	URL            string `validate:"required,url"`
}

func (a AI) AsOllamaProviderConfig() *OllamaProviderConfig {
	if a.Provider != OLLAMA_PROVIDER {
		return nil
	}
	return &OllamaProviderConfig{
		AIBase:         AIBase{Provider: a.Provider},
		EmbeddingModel: a.EmbeddingModel,
		URL:            a.URL,
	}
}

func (p AIProvider) String() string {
	return string(p)
}
