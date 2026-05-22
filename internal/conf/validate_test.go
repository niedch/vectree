package conf

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateSuccess(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")
	f, err := os.CreateTemp(t.TempDir(), "*.toml")
	require.NoError(t, err)
	defer f.Close()

	config_str := `
[sources.docs]
type = "http"
url = "https://docs.example.com/"

[ai]
provider = "gemini"
embedding_model = "text_embedding_004"
`
	_, err = f.WriteString(config_str)
	require.NoError(t, err)

	_, err = loadCustomFile(f.Name())
	require.NoError(t, err)
}

func TestInvalidSourceType(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "*.toml")
	require.NoError(t, err)
	defer f.Close()

	config_str := `
[sources.tech-docs]
type = "invalidtype"
`
	_, err = f.WriteString(config_str)
	require.NoError(t, err)

	_, err = loadCustomFile(f.Name())
	require.Error(t, err)
}

func TestHttpNeedToHaveValidUrl(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "*.toml")
	require.NoError(t, err)
	defer f.Close()

	config_str := `
[sources.tech-docs]
type = "http"
url = "asdfasdf"
	`
	_, err = f.WriteString(config_str)
	require.NoError(t, err)

	_, err = loadCustomFile(f.Name())
	require.Error(t, err)
}

func TestInvalidAIProvider(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "*.toml")
	require.NoError(t, err)
	defer f.Close()

	config_str := `
[sources.tech-docs]
type = "http"
url = "https://example.com"

[ai]
provider = "invalid"
`
	_, err = f.WriteString(config_str)
	require.NoError(t, err)

	_, err = loadCustomFile(f.Name())
	require.Error(t, err)
}

func TestGeminiProviderRequiresApiKey(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")
	f, err := os.CreateTemp(t.TempDir(), "*.toml")
	require.NoError(t, err)
	defer f.Close()

	config_str := `
[sources.tech-docs]
type = "http"
url = "https://example.com"

[ai]
provider = "gemini"
embedding_model = "text_embedding_004"
`
	_, err = f.WriteString(config_str)
	require.NoError(t, err)

	_, err = loadCustomFile(f.Name())
	require.Error(t, err)
}

func TestOllamaProviderRequiresUrl(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "*.toml")
	require.NoError(t, err)
	defer f.Close()

	config_str := `
[sources.tech-docs]
type = "http"
url = "https://example.com"

[ai]
provider = "ollama"
`
	_, err = f.WriteString(config_str)
	require.NoError(t, err)

	_, err = loadCustomFile(f.Name())
	require.Error(t, err)
}

func TestOllamaProviderWithUrl(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "*.toml")
	require.NoError(t, err)
	defer f.Close()

	config_str := `
[sources.tech-docs]
type = "http"
url = "https://example.com"

[ai]
provider = "ollama"
url = "http://localhost:11434"
embedding_model = "some-model"
`
	_, err = f.WriteString(config_str)
	require.NoError(t, err)

	_, err = loadCustomFile(f.Name())
	require.NoError(t, err)
}

func TestAIInvalidUrl(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-key")
	f, err := os.CreateTemp(t.TempDir(), "*.toml")
	require.NoError(t, err)
	defer f.Close()

	config_str := `
[sources.tech-docs]
type = "http"
url = "https://example.com"

[ai]
provider = "openai"
url = "not-a-valid-url"
`
	_, err = f.WriteString(config_str)
	require.NoError(t, err)

	_, err = loadCustomFile(f.Name())
	require.Error(t, err)
}

func TestMarkdownNeedsToPointToAFolder(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")
	f, err := os.CreateTemp(t.TempDir(), "*.toml")
	require.NoError(t, err)
	defer f.Close()

	config_str := `
[sources.tech-docs]
type = "markdown"
location = "internal/conf/"

[ai]
provider = "gemini"
embedding_model = "text_embedding_004"
	`
	_, err = f.WriteString(config_str)
	require.NoError(t, err)

	cfg, err := loadCustomFile(f.Name())
	require.NoError(t, err)

	geminiProviderConfig := cfg.AI.AsGeminiProviderConfig()
	assert.Equal(t, DEFAULT_VERTEX_SIZE, geminiProviderConfig.VertexSize)
}
