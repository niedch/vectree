package stages

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"runtime"
	"strings"

	"github.com/niedch/vectree/internal/mdast"
)

type SectionWithLevel struct {
	Text       string
	Level      int
	ParentId   *int
	DocumentId string
}

type MdAstSplitter struct {
	workers int
}

func NewMdAstSplitter() *MdAstSplitter {
	return &MdAstSplitter{workers: runtime.NumCPU()}
}

func (s MdAstSplitter) Run(ctx context.Context, in <-chan string) <-chan SectionWithLevel {
	return WorkerPoolStage(ctx, in, s.workers, s.processDocument)
}

func (s MdAstSplitter) processDocument(ctx context.Context, doc string, out chan<- SectionWithLevel) error {
	hash := sha256.Sum256([]byte(doc))
	documentId := hex.EncodeToString(hash[:8])

	docNode := mdast.ParseMarkdown(doc)
	children := docNode.Children()

	var sb strings.Builder
	sb.Grow(len(doc))

	for i := range children {
		heading, isHeading := children[i].(*mdast.HeadingNode)
		if !isHeading {
			continue
		}

		sb.Reset()
		mdast.WriteMarkdown(heading, &sb)

		j := i + 1
		for j < len(children) {
			nextNode := children[j]

			if nextHeading, isNextHeading := nextNode.(*mdast.HeadingNode); isNextHeading {
				if nextHeading.Level <= heading.Level {
					break
				}
				mdast.WriteMarkdown(nextHeading, &sb)
			} else {
				mdast.WriteMarkdown(nextNode, &sb)
			}

			j++
		}

		sectionText := sb.String()

		select {
		case out <- SectionWithLevel{
			Text:       sectionText,
			Level:      heading.Level,
			DocumentId: documentId,
		}:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}
