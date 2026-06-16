package stages

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndexFileFilter(t *testing.T) {
	tests := []struct {
		name     string
		input    []FileRef
		expected []FileRef
	}{
		{
			name:     "No ignored files",
			input:    makeFileRefs("/path/to/file1.md", "/path/to/file2.go"),
			expected: makeFileRefs("/path/to/file1.md", "/path/to/file2.go"),
		},
		{
			name:     "One ignored file",
			input:    makeFileRefs("/path/to/file1.md", "/path/to/!ignore.txt"),
			expected: makeFileRefs("/path/to/file1.md"),
		},
		{
			name:     "All ignored files",
			input:    makeFileRefs("/path/to/!file1.md", "/path/to/!file2.go"),
			expected: nil,
		},
		{
			name:     "Empty input",
			input:    makeFileRefs(),
			expected: nil,
		},
		{
			name:     "Mixed valid and invalid files",
			input:    makeFileRefs("file1", "file!2", "file3", "!file4"),
			expected: makeFileRefs("file1", "file3"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()

			in := make(chan FileRef, len(tt.input))
			for _, ref := range tt.input {
				in <- ref
			}
			close(in)

			filter := NewIndexFileFilter()
			out := filter.Run(ctx, in)

			var results []FileRef
			for res := range out {
				results = append(results, res)
			}

			assert.Equal(t, tt.expected, results)
		})
	}
}
