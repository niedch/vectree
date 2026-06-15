package stages

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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

// Run implements the Stage interface. It splits the input document by markdown headers
// and sends each chunk to the output channel. This implementation is optimized for memory efficiency.
func (s *HeaderSplitter) Run(ctx context.Context, in <-chan Document) <-chan Section {
	out := make(chan Section)

	go func() {
		defer close(out)

		for doc := range in {
			content := doc.Content
			if content == "" {
				continue
			}

			hash := sha256.Sum256([]byte(content))
			documentId := hex.EncodeToString(hash[:8])

			headerIndices := re.FindAllStringIndex(content, -1)

			// If there are no headers, send the whole content.
			if len(headerIndices) == 0 {
				select {
				case out <- Section{Text: content, Level: 0, DocumentId: documentId, Source: doc.Source}:
				case <-ctx.Done():
					return
				}
				continue
			}

			// Send content between headers.
			for i := 0; i < len(headerIndices)-1; i++ {
				output := strings.TrimSpace(content[headerIndices[i][0]:headerIndices[i+1][0]])
				if output != "" {
					headerLine := content[headerIndices[i][0]:headerIndices[i][1]]
					level := countHeaderLevel(headerLine)
					select {
					case out <- Section{Text: output, Level: level, DocumentId: documentId, Source: doc.Source}:
					case <-ctx.Done():
						return
					}
				}
			}

			// Send the content from the last header to the end.
			output := strings.TrimSpace(content[headerIndices[len(headerIndices)-1][0]:])
			if output != "" {
				headerLine := content[headerIndices[len(headerIndices)-1][0]:headerIndices[len(headerIndices)-1][1]]
				level := countHeaderLevel(headerLine)
				select {
				case out <- Section{Text: output, Level: level, DocumentId: documentId, Source: doc.Source}:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return out
}

func countHeaderLevel(headerLine string) int {
	level := 0
	for _, c := range headerLine {
		if c == '#' {
			level++
		} else {
			break
		}
	}
	return level
}
