package stages

import (
	"context"
	"regexp"
	"strings"
)

var (
	re = regexp.MustCompile(`(?m)^#+\s.*$`)
)

// HeaderSplitter splits text based on markdown headers.
type HeaderSplitter struct{}

// NewHeaderSplitter creates a new HeaderSplitter stage.
func NewHeaderSplitter() *HeaderSplitter {
	return &HeaderSplitter{}
}

// Run implements the Stage interface. It splits the input string by markdown headers
// and sends each chunk to the output channel. This implementation is optimized for memory efficiency.
func (s *HeaderSplitter) Run(ctx context.Context, in <-chan string) <-chan string {
	out := make(chan string)

	go func() {
		defer close(out)

		for content := range in {
			if content == "" {
				continue
			}

			headerIndices := re.FindAllStringIndex(content, -1)

			// If there are no headers, send the whole content.
			if len(headerIndices) == 0 {
				select {
				case out <- content:
				case <-ctx.Done():
					return
				}
				continue
			}

			// Send content between headers.
			for i := 0; i < len(headerIndices)-1; i++ {
				output := strings.TrimSpace(content[headerIndices[i][0]:headerIndices[i+1][0]])
				if output != "" {
					select {
					case out <- output:
					case <-ctx.Done():
						return
					}
				}
			}

			// Send the content from the last header to the end.
			output := strings.TrimSpace(content[headerIndices[len(headerIndices)-1][0]:])
			if output != "" {
				select {
				case out <- output:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return out
}
