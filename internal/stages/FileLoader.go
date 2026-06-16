package stages

import (
	"context"
	"log"
	"os"
)

type FileLoader struct {
	sourceName string
}

func NewFileLoader(sourceName string) *FileLoader {
	return &FileLoader{sourceName: sourceName}
}

func (l FileLoader) Run(ctx context.Context, in <-chan FileRef) <-chan Document {
	out := make(chan Document)

	go func() {
		defer close(out)
		fileCount := 0
		var totalSize uint64 = 0

		for ref := range in {
			data, err := os.ReadFile(ref.Path)
			if err != nil {
				return
			}

			if len(data) == 0 {
				continue
			}

			fileCount++
			totalSize += uint64(len(data))

			select {
			case out <- Document{Content: string(data), Source: ref.Source}:
			case <-ctx.Done():
				return
			}
		}

		log.Printf("[%s] Loaded %d files, total size: %d bytes (%.2f MB)\n",
			l.sourceName, fileCount, totalSize, float64(totalSize)/(1024*1024))
	}()

	return out
}
