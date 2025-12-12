package stages

import (
	"context"
	"io/fs"
	"log"
	"path/filepath"
	"strings"
)

type DirLoader struct {
	dirPath string
}

func NewDirLoader(dirPath string) *DirLoader {
	return &DirLoader{
		dirPath: dirPath,
	}
}

func (l DirLoader) Run(ctx context.Context, _ <-chan any) <-chan string {
	out := make(chan string)

	go func() {
		defer close(out)

		err := filepath.Walk(l.dirPath, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if !info.IsDir() && strings.HasSuffix(info.Name(), ".md") {
				log.Println("Found File: ", path)
				select {
				case out <- path:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return nil
		})

		if err != nil {
			return
		}
	}()

	return out
}
