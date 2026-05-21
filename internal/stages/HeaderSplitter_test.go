package stages

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeaderSplitter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name: "No headers",
			input: `This is a text without headers.
It has multiple lines.`,
			expected: []string{`This is a text without headers.
It has multiple lines.`},
		},
		{
			name: "Simple headers",
			input: `
# Title
This is the first section.
## Subtitle 1
This is the first subsection.
## Subtitle 2
This is the second subsection.`,
			expected: []string{`# Title
This is the first section.`,
				`## Subtitle 1
This is the first subsection.`,
				`## Subtitle 2
This is the second subsection.`},
		},
		{
			name: "Text before should not be ingested first header",
			input: `Some text before the first header.
# Header 1
Text in section 1.`,
			expected: []string{`# Header 1
Text in section 1.`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()

			in := make(chan string, 1)
			in <- tt.input
			close(in)

			splitter := NewHeaderSplitter()
			out := splitter.Run(ctx, in)

			var results []string
			for res := range out {
				results = append(results, res)
			}

			assert.Equal(t, tt.expected, results)
		})
	}
}
