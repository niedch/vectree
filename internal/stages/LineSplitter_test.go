package stages

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLineSplitter_Run(t *testing.T) {
	splitter := NewLineSplitter()
	in := make(chan Document, 1)
	in <- Document{Content: "hello\nworld", Source: "file://test.md"}
	close(in)

	out := splitter.Run(context.Background(), in)

	var lines []string
	for line := range out {
		lines = append(lines, line.Text)
	}

	expectedLines := []string{"hello", "world"}
	assert.Equal(t, expectedLines, lines)
}

func TestLineSplitter_Run_EmptyLines(t *testing.T) {
	splitter := NewLineSplitter()
	in := make(chan Document, 1)
	in <- Document{Content: "hello\n\nworld", Source: "file://test.md"}
	close(in)

	out := splitter.Run(context.Background(), in)

	var lines []string
	for line := range out {
		lines = append(lines, line.Text)
	}

	expectedLines := []string{"hello", "world"}
	assert.Equal(t, expectedLines, lines)
}

func TestLineSplitter_Run_EmptyString(t *testing.T) {
	splitter := NewLineSplitter()
	in := make(chan Document, 1)
	in <- Document{Content: "", Source: "file://test.md"}
	close(in)

	out := splitter.Run(context.Background(), in)

	var lines []string
	for line := range out {
		lines = append(lines, line.Text)
	}

	assert.Empty(t, lines)
}
