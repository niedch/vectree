package stages

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"runtime"
	"strings"

	"github.com/niedch/vectree/internal/mdast"
)

type Section struct {
	Text       string
	Level      int
	ParentId   *int
	DocumentId string
	Source     string
}

type MdAstSplitter struct {
	workers int
}

func NewMdAstSplitter() *MdAstSplitter {
	return &MdAstSplitter{workers: runtime.NumCPU()}
}

func (s MdAstSplitter) Run(ctx context.Context, in <-chan Document) <-chan Section {
	return WorkerPoolStage(ctx, in, s.workers, func(ctx context.Context, doc Document, out chan<- Section) error {
		return s.processDocument(ctx, doc, out)
	})
}

func (s MdAstSplitter) processDocument(ctx context.Context, doc Document, out chan<- Section) error {
	hash := sha256.Sum256([]byte(doc.Content))
	documentId := hex.EncodeToString(hash[:8])

	docNode := mdast.ParseMarkdown(doc.Content)
	children := docNode.Children()

	var sb strings.Builder
	sb.Grow(len(doc.Content))

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
		case out <- Section{
			Text:       sectionText,
			Level:      heading.Level,
			DocumentId: documentId,
			Source:     doc.Source,
		}:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}
