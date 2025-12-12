package stages

import (
	"context"
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

		for filePath := range in {
			data, err := os.ReadFile(filePath)
			if err != nil {
				return
			}

			if (len(data) == 0) {
				continue;
			}

			select {
			case out <- string(data):
			case <-ctx.Done():
				return
			}
		}
	}()

	return out
}
