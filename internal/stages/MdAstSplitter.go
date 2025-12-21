package stages

import (
	"broadcom.com/vertex-ingestor/internal/mdast"
	"context"
)

type MdAstSplitter struct{}

func NewMdAstSplitter() *MdAstSplitter {
	return &MdAstSplitter{}
}

func (s MdAstSplitter) Run(ctx context.Context, in <-chan string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for doc := range in {

			docNode := mdast.ParseMarkdown(doc)
			output := mdast.PrintAST(docNode, 4)

			select {
			case out <- output:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}
