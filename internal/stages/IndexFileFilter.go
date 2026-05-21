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

func (l IndexFileFilter) Run(ctx context.Context, in <-chan string) <-chan string {
	out := make(chan string)

	go func() {
		defer close(out)

		for filePath := range in {
			if strings.Contains(filePath, "!") {
				log.Println("Ignoring File", filePath)
				continue
			}

			select {
			case out <- filePath:
			case <-ctx.Done():
				return
			}
		}
	}()

	return out
}
