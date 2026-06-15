package stages

import (
	"context"
	"io/fs"
	"path/filepath"
	"strings"
)

type DirLoader struct {
	dirPath      string
	sourcePrefix string
}

func NewDirLoader(dirPath string) *DirLoader {
	return &DirLoader{dirPath: dirPath}
}

func NewDirLoaderWithSource(dirPath, sourcePrefix string) *DirLoader {
	return &DirLoader{dirPath: dirPath, sourcePrefix: sourcePrefix}
}

func (l DirLoader) Run(ctx context.Context, in <-chan any) <-chan FileRef {
	out := make(chan FileRef)

	walk := func(root string) {
		filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if !info.IsDir() && strings.HasSuffix(info.Name(), ".md") {
				source := path
				if l.sourcePrefix != "" {
					rel, err := filepath.Rel(root, path)
					if err == nil {
						source = l.sourcePrefix + rel
					}
				} else {
					source = "file://" + path
				}
				select {
				case out <- FileRef{Path: path, Source: source}:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return nil
		})
	}

	go func() {
		defer close(out)

		if l.dirPath != "" {
			walk(l.dirPath)
		} else {
			for raw := range in {
				dir, ok := raw.(string)
				if !ok {
					continue
				}
				walk(dir)
			}
		}
	}()

	return out
}
