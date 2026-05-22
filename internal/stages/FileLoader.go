package stages

import (
	"context"
	"log"
	"os"
)

type FileLoader struct {
	sourceName string
	dirPath    string
}

func NewFileLoader(sourceName string) *FileLoader {
	return &FileLoader{sourceName: sourceName}
}

func (l FileLoader) Run(ctx context.Context, in <-chan string) <-chan string {
	out := make(chan string)

	go func() {
		defer close(out)
		fileCount := 0
		var totalSize uint64 = 0

		for filePath := range in {
			data, err := os.ReadFile(filePath)
			if err != nil {
				return
			}

			if len(data) == 0 {
				continue
			}

			fileCount++
			totalSize += uint64(len(data))

			select {
			case out <- string(data):
			case <-ctx.Done():
				return
			}
		}

		log.Printf("[%s] Loaded %d files, total size: %d bytes (%.2f MB)\n",
			l.sourceName, fileCount, totalSize, float64(totalSize)/(1024*1024))
	}()

	return out
}
