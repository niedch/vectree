package stages

import (
	"context"
	"io/fs"
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

func (l DirLoader) Run(ctx context.Context, in <-chan any) <-chan string {
	out := make(chan string)

	go func() {
		defer close(out)

		walk := func(dir string) {
			err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if ctx.Err() != nil {
					return ctx.Err()
				}
				if !info.IsDir() && strings.HasSuffix(info.Name(), ".md") {
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
		}

		if l.dirPath != "" {
			walk(l.dirPath)
		}

		for raw := range in {
			dir, ok := raw.(string)
			if !ok {
				continue
			}
			walk(dir)
		}
	}()

	return out
}
