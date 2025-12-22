package stages

import (
	"broadcom.com/vertex-ingestor/internal/mdast"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
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
		docCount := 0
		var totalInputSize uint64 = 0
		var totalOutputSize uint64 = 0
		sectionCount := 0

		for doc := range in {
			docCount++
			inputSize := len(doc)
			totalInputSize += uint64(inputSize)

			// Generate a unique document ID based on content hash
			hash := sha256.Sum256([]byte(doc))
			documentId := hex.EncodeToString(hash[:8]) // Use first 8 bytes for brevity

			docNode := mdast.ParseMarkdown(doc)

			if !extractAllSections(ctx, docNode, documentId, out, &sectionCount, &totalOutputSize) {
				return
			}

			if docCount%100 == 0 {
				log.Printf("MdAstSplitter: Processed %d documents, %d sections so far. Input: %d bytes, Output: %d bytes\n", 
					docCount, sectionCount, totalInputSize, totalOutputSize)
			}
		}

		log.Printf("MdAstSplitter: Total processed %d documents into %d sections\n", docCount, sectionCount)
		log.Printf("MdAstSplitter: Input size: %d bytes (%.2f MB), Output size: %d bytes (%.2f MB)\n", 
			totalInputSize, float64(totalInputSize)/(1024*1024), 
			totalOutputSize, float64(totalOutputSize)/(1024*1024))
		log.Printf("MdAstSplitter: Expansion ratio: %.2fx\n", float64(totalOutputSize)/float64(totalInputSize))
	}()
	return out
}

// extractAllSections outputs a section for EVERY heading in the document
// Each section includes the heading, its content, and all subheadings with their content
func extractAllSections(ctx context.Context, docNode *mdast.DocumentNode, documentId string, out chan<- SectionWithLevel, sectionCount *int, totalOutputSize *uint64) bool {
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
		sectionText := sb.String()
		section := SectionWithLevel{
			Text:       sectionText,
			Level:      heading.Level,
			DocumentId: documentId,
		}
		
		*sectionCount++
		*totalOutputSize += uint64(len(sectionText))
		
		select {
		case out <- section:
		case <-ctx.Done():
			return false
		}
	}

	return true
}
