package stages

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileLoader_Run(t *testing.T) {
	loader := NewFileLoader("Test")
	in := make(chan FileRef, 1)
	in <- FileRef{Path: "../../test_data/file_loader/Markdown.md", Source: "file://test.md"}
	close(in)

	out := loader.Run(context.Background(), in)

	var foundContent string
	var foundSource string
	for doc := range out {
		foundContent = doc.Content
		foundSource = doc.Source
	}
	assert.Equal(t, "# Test Markdown File\n", foundContent)
	assert.Equal(t, "file://test.md", foundSource)
}

func TestFileLoader_Run_NonExistentFile(t *testing.T) {
	loader := NewFileLoader("Test")
	in := make(chan FileRef, 1)
	in <- FileRef{Path: "../../test_data/file_loader/NonExsiting.md", Source: "file://nonexistent.md"}
	close(in)

	out := loader.Run(context.Background(), in)

	var foundContent []Document
	for doc := range out {
		foundContent = append(foundContent, doc)
	}

	assert.Empty(t, foundContent)
}

func TestFileLoader_Run_EmptyFile(t *testing.T) {
	loader := NewFileLoader("Test")
	in := make(chan FileRef, 1)
	in <- FileRef{Path: "../../test_data/file_loader/EmptyMarkdown.md", Source: "file://empty.md"}
	close(in)

	out := loader.Run(context.Background(), in)

	var foundContent []Document
	for doc := range out {
		foundContent = append(foundContent, doc)
	}

	assert.Empty(t, foundContent)
}
