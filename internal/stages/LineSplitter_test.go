package stages

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLineSplitter_Run(t *testing.T) {
	splitter := NewLineSplitter()
	in := make(chan string, 1)
	in <- "hello\nworld"
	close(in)

	out := splitter.Run(context.Background(), in)

	var lines []string
	for line := range out {
		lines = append(lines, line)
	}

	expectedLines := []string{"hello", "world"}
	assert.Equal(t, expectedLines, lines)
}

func TestLineSplitter_Run_EmptyLines(t *testing.T) {
	splitter := NewLineSplitter()
	in := make(chan string, 1)
	in <- "hello\n\nworld"
	close(in)

	out := splitter.Run(context.Background(), in)

	var lines []string
	for line := range out {
		lines = append(lines, line)
	}

	expectedLines := []string{"hello", "world"}
	assert.Equal(t, expectedLines, lines)
}

func TestLineSplitter_Run_EmptyString(t *testing.T) {
	splitter := NewLineSplitter()
	in := make(chan string, 1)
	in <- ""
	close(in)

	out := splitter.Run(context.Background(), in)

	var lines []string
	for line := range out {
		lines = append(lines, line)
	}

	assert.Empty(t, lines)
}
