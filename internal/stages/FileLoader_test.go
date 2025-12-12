package stages

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileLoader_Run(t *testing.T) {
	loader := NewFileLoader()
	in := make(chan string, 1)
	in <- "../../test_data/file_loader/Markdown.md"
	close(in)

	out := loader.Run(context.Background(), in)

	var foundContent string
	for data := range out {
		foundContent = data
	}
	assert.Equal(t, "# Test Markdown File\n", foundContent)

}

func TestFileLoader_Run_NonExistentFile(t *testing.T) {
	loader := NewFileLoader()
	in := make(chan string, 1)
	in <- "../../test_data/file_loader/NonExsiting.md"
	close(in)

	out := loader.Run(context.Background(), in)

	var foundContent []string
	for data := range out {
		foundContent = append(foundContent, data)
	}

	assert.Empty(t, foundContent)
}

func TestFileLoader_Run_EmptyFile(t *testing.T) {
	loader := NewFileLoader()
	in := make(chan string, 1)
	in <- "../../test_data/file_loader/EmptyMarkdown.md"
	close(in)

	out := loader.Run(context.Background(), in)

	var foundContent []string
	for data := range out {
		foundContent = append(foundContent, data)
	}

	assert.Empty(t, foundContent)
}