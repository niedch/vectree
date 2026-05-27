package promptloader

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFile_BasicPrompt(t *testing.T) {
	p, err := loadFile("testdata/basic.prompt")
	require.NoError(t, err)
	assert.Equal(t, "greeter", p.Name)
	assert.Equal(t, "A friendly greeting prompt", p.Description)
	assert.Len(t, p.Arguments, 2)

	assert.Equal(t, "name", p.Arguments[0].Name)
	assert.True(t, p.Arguments[0].Required)
	assert.Equal(t, "string", p.Arguments[0].Description)

	assert.Equal(t, "language", p.Arguments[1].Name)
	assert.False(t, p.Arguments[1].Required)
	assert.Equal(t, "string", p.Arguments[1].Description)
}

func TestLoadFile_NoFrontmatter(t *testing.T) {
	p, err := loadFile("testdata/notext.prompt")
	require.NoError(t, err)
	assert.Equal(t, "notext", p.Name)
	assert.Empty(t, p.Description)
	assert.Empty(t, p.Arguments)
}

func TestLoadFile_OnlyName(t *testing.T) {
	p, err := loadFile("testdata/onlyname.prompt")
	require.NoError(t, err)
	assert.Equal(t, "simple", p.Name)
	assert.Empty(t, p.Description)
	assert.Empty(t, p.Arguments)
}

func TestRenderPrompt(t *testing.T) {
	source := "Hello {{name}}, welcome to {{place}}!"
	result, err := RenderPrompt(source, map[string]string{
		"name":  "Alice",
		"place": "Wonderland",
	})
	require.NoError(t, err)
	assert.Equal(t, "Hello Alice, welcome to Wonderland!", result)
}

func TestRenderPrompt_NoVars(t *testing.T) {
	result, err := RenderPrompt("static text", nil)
	require.NoError(t, err)
	assert.Equal(t, "static text", result)
}

func TestRenderPrompt_MissingVar(t *testing.T) {
	source := "Hello {{name}}!"
	result, err := RenderPrompt(source, map[string]string{})
	require.NoError(t, err)
	assert.Equal(t, "Hello !", result)
}

func TestRenderPrompt_SomeMissing(t *testing.T) {
	source := "Hello {{name}} at {{place}}!"
	result, err := RenderPrompt(source, map[string]string{"name": "Alice"})
	require.NoError(t, err)
	assert.Equal(t, "Hello Alice at !", result)
}
