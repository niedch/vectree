package stages

import (
	"context"
	"strings"
)

type LineSplitter struct{}

func NewLineSplitter() *LineSplitter {
	return &LineSplitter{}
}

func (s LineSplitter) Run(ctx context.Context, in <-chan string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for doc := range in {

			lines := strings.Split(doc, "\n")

			for _, line := range lines {

				if len(line) == 0 {
					continue
				}

				select {
				case out <- line:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
