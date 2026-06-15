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

func makeFileRefs(paths ...string) []FileRef {
	refs := make([]FileRef, len(paths))
	for i, p := range paths {
		refs[i] = FileRef{Path: p, Source: "file://" + p}
	}
	return refs
}

func TestNodeModulesFilter_Run(t *testing.T) {
	tests := []struct {
		name     string
		input    []FileRef
		expected []FileRef
	}{
		{
			name: "No node_modules files",
			input: makeFileRefs(
				"/path/to/file1.md",
				"/path/to/file2.go",
				"/path/to/src/file3.js",
			),
			expected: makeFileRefs(
				"/path/to/file1.md",
				"/path/to/file2.go",
				"/path/to/src/file3.js",
			),
		},
		{
			name: "Filter node_modules directory",
			input: makeFileRefs(
				"/path/to/file1.md",
				"/path/to/node_modules/package.json",
				"/path/to/file2.go",
			),
			expected: makeFileRefs(
				"/path/to/file1.md",
				"/path/to/file2.go",
			),
		},
		{
			name: "Filter nested node_modules",
			input: makeFileRefs(
				"/path/to/file1.md",
				"/path/to/node_modules/lib/index.js",
				"/path/to/src/node_modules/package.json",
				"/path/to/file2.go",
			),
			expected: makeFileRefs(
				"/path/to/file1.md",
				"/path/to/file2.go",
			),
		},
		{
			name: "All files in node_modules",
			input: makeFileRefs(
				"/path/to/node_modules/file1.md",
				"/path/to/node_modules/file2.go",
				"/path/to/node_modules/src/file3.js",
			),
			expected: makeFileRefs(),
		},
		{
			name:     "Empty input",
			input:    makeFileRefs(),
			expected: makeFileRefs(),
		},
		{
			name: "File with node_modules in name but not in path",
			input: makeFileRefs(
				"/path/to/my_node_modules_backup.md",
				"/path/to/node_modules_info.txt",
			),
			expected: makeFileRefs(),
		},
		{
			name: "Mixed valid and node_modules files",
			input: makeFileRefs(
				"/project/src/main.go",
				"/project/node_modules/react/index.js",
				"/project/docs/README.md",
				"/project/frontend/node_modules/vue/dist/vue.js",
				"/project/backend/server.go",
			),
			expected: makeFileRefs(
				"/project/src/main.go",
				"/project/docs/README.md",
				"/project/backend/server.go",
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewNodeModulesFilter()
			ctx := context.Background()

			in := make(chan FileRef, len(tt.input))
			for _, ref := range tt.input {
				in <- ref
			}
			close(in)

			out := filter.Run(ctx, in)

			results := []FileRef{}
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
	input := makeFileRefs(
		"/path/to/Node_Modules/file1.js",
		"/path/to/NODE_MODULES/file2.js",
		"/path/to/node_modules/file3.js",
	)

	in := make(chan FileRef, len(input))
	for _, ref := range input {
		in <- ref
	}
	close(in)

	out := filter.Run(ctx, in)

	var results []FileRef
	for result := range out {
		results = append(results, result)
	}

	// Current implementation is case-sensitive, so only lowercase "node_modules" is filtered
	assert.Len(t, results, 2)
	assert.Equal(t, "/path/to/Node_Modules/file1.js", results[0].Path)
	assert.Equal(t, "/path/to/NODE_MODULES/file2.js", results[1].Path)
}
