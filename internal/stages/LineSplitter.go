package stages

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

type LineSplitter struct{}

func NewLineSplitter() *LineSplitter {
	return &LineSplitter{}
}

func (s LineSplitter) Run(ctx context.Context, in <-chan string) <-chan Section {
	out := make(chan Section)
	go func() {
		defer close(out)
		for doc := range in {
			hash := sha256.Sum256([]byte(doc))
			documentId := hex.EncodeToString(hash[:8])

			lines := strings.SplitSeq(doc, "\n")

			for line := range lines {
				if len(line) == 0 {
					continue
				}

				select {
				case out <- Section{Text: line, Level: 0, DocumentId: documentId}:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
