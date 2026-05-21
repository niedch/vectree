package stages

import (
	"context"
	"log"
	"os"
)

type FileLoader struct {
	dirPath string
}

func NewFileLoader() *FileLoader {
	return &FileLoader{}
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

		log.Printf("FileLoader: Loaded %d files, total size: %d bytes (%.2f MB)\n",
			fileCount, totalSize, float64(totalSize)/(1024*1024))
	}()

	return out
}
