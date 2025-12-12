package stages

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndexFileFilter(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "No ignored files",
			input:    []string{"/path/to/file1.md", "/path/to/file2.go"},
			expected: []string{"/path/to/file1.md", "/path/to/file2.go"},
		},
		{
			name:     "One ignored file",
			input:    []string{"/path/to/file1.md", "/path/to/!ignore.txt"},
			expected: []string{"/path/to/file1.md"},
		},
		{
			name:     "All ignored files",
			input:    []string{"/path/to/!file1.md", "/path/to/!file2.go"},
			expected: nil,
		},
		{
			name:     "Empty input",
			input:    []string{},
			expected: nil,
		},
		{
			name:     "Mixed valid and invalid files",
			input:    []string{"file1", "file!2", "file3", "!file4"},
			expected: []string{"file1", "file3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			in := make(chan string, len(tt.input))
			for _, path := range tt.input {
				in <- path
			}
			close(in)

			filter := NewIndexFileFilter()
			out := filter.Run(ctx, in)

			var results []string
			for res := range out {
				results = append(results, res)
			}

			assert.Equal(t, tt.expected, results)
		})
	}
}
