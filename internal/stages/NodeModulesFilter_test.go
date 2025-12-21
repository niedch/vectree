package stages

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewNodeModulesFilter(t *testing.T) {
	filter := NewNodeModulesFilter()
	assert.NotNil(t, filter)
}

func TestNodeModulesFilter_Run(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name: "No node_modules files",
			input: []string{
				"/path/to/file1.md",
				"/path/to/file2.go",
				"/path/to/src/file3.js",
			},
			expected: []string{
				"/path/to/file1.md",
				"/path/to/file2.go",
				"/path/to/src/file3.js",
			},
		},
		{
			name: "Filter node_modules directory",
			input: []string{
				"/path/to/file1.md",
				"/path/to/node_modules/package.json",
				"/path/to/file2.go",
			},
			expected: []string{
				"/path/to/file1.md",
				"/path/to/file2.go",
			},
		},
		{
			name: "Filter nested node_modules",
			input: []string{
				"/path/to/file1.md",
				"/path/to/node_modules/lib/index.js",
				"/path/to/src/node_modules/package.json",
				"/path/to/file2.go",
			},
			expected: []string{
				"/path/to/file1.md",
				"/path/to/file2.go",
			},
		},
		{
			name: "All files in node_modules",
			input: []string{
				"/path/to/node_modules/file1.md",
				"/path/to/node_modules/file2.go",
				"/path/to/node_modules/src/file3.js",
			},
			expected: []string{},
		},
		{
			name:     "Empty input",
			input:    []string{},
			expected: []string{},
		},
		{
			name: "File with node_modules in name but not in path",
			input: []string{
				"/path/to/my_node_modules_backup.md",
				"/path/to/node_modules_info.txt",
			},
			expected: []string{},
		},
		{
			name: "Mixed valid and node_modules files",
			input: []string{
				"/project/src/main.go",
				"/project/node_modules/react/index.js",
				"/project/docs/README.md",
				"/project/frontend/node_modules/vue/dist/vue.js",
				"/project/backend/server.go",
			},
			expected: []string{
				"/project/src/main.go",
				"/project/docs/README.md",
				"/project/backend/server.go",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewNodeModulesFilter()
			ctx := context.Background()

			in := make(chan string, len(tt.input))
			for _, path := range tt.input {
				in <- path
			}
			close(in)

			out := filter.Run(ctx, in)

			results := []string{}
			for result := range out {
				results = append(results, result)
			}

			assert.Equal(t, tt.expected, results)
		})
	}
}

func TestNodeModulesFilter_CaseInsensitive(t *testing.T) {
	filter := NewNodeModulesFilter()
	ctx := context.Background()

	// Note: The current implementation is case-sensitive
	// This test documents the current behavior
	input := []string{
		"/path/to/Node_Modules/file1.js",
		"/path/to/NODE_MODULES/file2.js",
		"/path/to/node_modules/file3.js",
	}

	in := make(chan string, len(input))
	for _, path := range input {
		in <- path
	}
	close(in)

	out := filter.Run(ctx, in)

	var results []string
	for result := range out {
		results = append(results, result)
	}

	// Current implementation is case-sensitive, so only lowercase "node_modules" is filtered
	assert.Len(t, results, 2)
	assert.Contains(t, results, "/path/to/Node_Modules/file1.js")
	assert.Contains(t, results, "/path/to/NODE_MODULES/file2.js")
	assert.NotContains(t, results, "/path/to/node_modules/file3.js")
}
