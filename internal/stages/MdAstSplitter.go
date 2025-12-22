package stages

import (
	"broadcom.com/vertex-ingestor/internal/mdast"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

type SectionWithLevel struct {
	Text       string
	Level      int
	ParentId   *int
	DocumentId string // Hash or identifier to track which document this section belongs to
}

type MdAstSplitter struct{}

func NewMdAstSplitter() *MdAstSplitter {
	return &MdAstSplitter{}
}

func (s MdAstSplitter) Run(ctx context.Context, in <-chan string) <-chan SectionWithLevel {
	out := make(chan SectionWithLevel)
	go func() {
		defer close(out)
		for doc := range in {
			// Generate a unique document ID based on content hash
			hash := sha256.Sum256([]byte(doc))
			documentId := hex.EncodeToString(hash[:8]) // Use first 8 bytes for brevity

			docNode := mdast.ParseMarkdown(doc)

			if !extractAllSections(ctx, docNode, documentId, out) {
				return
			}
		}
	}()
	return out
}

// extractAllSections outputs a section for EVERY heading in the document
// Each section includes the heading, its content, and all subheadings with their content
func extractAllSections(ctx context.Context, docNode *mdast.DocumentNode, documentId string, out chan<- SectionWithLevel) bool {
	children := docNode.Children()

	for i := range children {
		heading, isHeading := children[i].(*mdast.HeadingNode)
		if !isHeading {
			continue
		}

		// Build a section for this heading
		var sb strings.Builder
		sb.WriteString(heading.ToMarkdown())

		// Collect all following content until we hit a heading of equal or higher level
		j := i + 1
		for j < len(children) {
			nextNode := children[j]

			// Check if it's a heading
			if nextHeading, isNextHeading := nextNode.(*mdast.HeadingNode); isNextHeading {
				// If it's a heading of equal or higher level (lower or equal number), stop
				if nextHeading.Level <= heading.Level {
					break
				}
				// Otherwise, it's a subheading - include it
				sb.WriteString(nextHeading.ToMarkdown())
			} else {
				// It's a paragraph or other content - include it
				sb.WriteString(nextNode.ToMarkdown())
			}

			j++
		}

		// Output the complete section for this heading with level information
		section := SectionWithLevel{
			Text:       sb.String(),
			Level:      heading.Level,
			DocumentId: documentId,
		}
		select {
		case out <- section:
		case <-ctx.Done():
			return false
		}
	}

	return true
}
