package stages

import (
	"context"
	"log"
	"strings"
)

type IndexFileFilter struct {
	dirPath string
}

func NewIndexFileFilter() *IndexFileFilter {
	return &IndexFileFilter{}
}

func (l IndexFileFilter) Run(ctx context.Context, in <-chan FileRef) <-chan FileRef {
	out := make(chan FileRef)

	go func() {
		defer close(out)

		for ref := range in {
			if strings.Contains(ref.Path, "!") {
				log.Println("Ignoring File", ref.Path)
				continue
			}

			select {
			case out <- ref:
			case <-ctx.Done():
				return
			}
		}
	}()

	return out
}
