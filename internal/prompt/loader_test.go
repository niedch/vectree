package prompt

import (
	"testing"

	"github.com/google/dotprompt/go/dotprompt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDir_BasicPrompt(t *testing.T) {
	prompts, err := LoadDir("testdata")
	require.NoError(t, err)

	var greeter *Prompt
	for i := range prompts {
		if prompts[i].Name == "greeter" {
			greeter = &prompts[i]
			break
		}
	}
	require.NotNil(t, greeter, "expected to find prompt named 'greeter'")
	assert.Equal(t, "A friendly greeting prompt", greeter.Description)
	assert.Len(t, greeter.Arguments, 2)

	argsByName := make(map[string]Argument)
	for _, a := range greeter.Arguments {
		argsByName[a.Name] = a
	}

	assert.Equal(t, "string", argsByName["name"].Description)
	assert.True(t, argsByName["name"].Required)
	assert.Equal(t, "string", argsByName["language"].Description)
	assert.False(t, argsByName["language"].Required)
}

func TestLoadDir_NoFrontmatter(t *testing.T) {
	prompts, err := LoadDir("testdata")
	require.NoError(t, err)

	var notext *Prompt
	for i := range prompts {
		if prompts[i].Name == "notext" {
			notext = &prompts[i]
			break
		}
	}
	require.NotNil(t, notext, "expected to find prompt named 'notext'")
	assert.Empty(t, notext.Description)
	assert.Empty(t, notext.Arguments)
}

func TestLoadDir_OnlyName(t *testing.T) {
	prompts, err := LoadDir("testdata")
	require.NoError(t, err)

	var simple *Prompt
	for i := range prompts {
		if prompts[i].Name == "simple" {
			simple = &prompts[i]
			break
		}
	}
	require.NotNil(t, simple, "expected to find prompt named 'simple'")
	assert.Empty(t, simple.Description)
	assert.Empty(t, simple.Arguments)
}

func TestLoadDir_SourcePreserved(t *testing.T) {
	prompts, err := LoadDir("testdata")
	require.NoError(t, err)

	var greeter *Prompt
	for i := range prompts {
		if prompts[i].Name == "greeter" {
			greeter = &prompts[i]
			break
		}
	}
	require.NotNil(t, greeter, "expected to find prompt named 'greeter'")
	assert.Contains(t, greeter.Source, "Hello {{name}}!")
	assert.Contains(t, greeter.Source, "{{#if language}}")
}

func TestLoadFromSource_BasicPrompt(t *testing.T) {
	dp := dotprompt.NewDotprompt(nil)

	source := `---
name: test-prompt
description: A test prompt
input:
  schema:
    query: string
    limit?: number
---
Search for: {{query}}{{#if limit}} (max {{limit}} results){{/if}}`

	p, err := LoadFromSource(dp, source, "fallback-name")
	require.NoError(t, err)

	assert.Equal(t, "test-prompt", p.Name)
	assert.Equal(t, "A test prompt", p.Description)
	assert.Len(t, p.Arguments, 2)

	argsByName := make(map[string]Argument)
	for _, a := range p.Arguments {
		argsByName[a.Name] = a
	}
	assert.True(t, argsByName["query"].Required)
	assert.Equal(t, "string", argsByName["query"].Description)
	assert.False(t, argsByName["limit"].Required)
	assert.Equal(t, "number", argsByName["limit"].Description)

	assert.Contains(t, p.Source, "Search for: {{query}}")
}

func TestLoadFromSource_NoFrontmatter(t *testing.T) {
	dp := dotprompt.NewDotprompt(nil)

	p, err := LoadFromSource(dp, "just plain text", "plain")
	require.NoError(t, err)

	assert.Equal(t, "plain", p.Name)
	assert.Empty(t, p.Description)
	assert.Empty(t, p.Arguments)
	assert.Equal(t, "just plain text", p.Source)
}

func TestLoadFromSource_InlineSchemaWithCommaDescription(t *testing.T) {
	dp := dotprompt.NewDotprompt(nil)

	source := `---
name: research-topic
description: Search documentation for a topic
input:
  schema:
    dev-topic: string,The technical topic, API, or feature to research
---
Search for {{dev-topic}}`

	p, err := LoadFromSource(dp, source, "fallback")
	require.NoError(t, err)

	assert.Equal(t, "research-topic", p.Name)
	assert.Equal(t, "Search documentation for a topic", p.Description)
	assert.Len(t, p.Arguments, 1)

	assert.Equal(t, "dev-topic", p.Arguments[0].Name)
	assert.True(t, p.Arguments[0].Required)
	assert.Equal(t, "The technical topic, API, or feature to research", p.Arguments[0].Description)
}

func TestLoadFromSource_NameFromFrontmatterOverrides(t *testing.T) {
	dp := dotprompt.NewDotprompt(nil)

	source := `---
name: overridden-name
---
Hello`
	p, err := LoadFromSource(dp, source, "original-name")
	require.NoError(t, err)

	assert.Equal(t, "overridden-name", p.Name)
}