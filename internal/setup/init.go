package setup

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed project/*
var projectFiles embed.FS

func InitProject() error {
	return fs.WalkDir(projectFiles, "project", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel("project", path)
		if relPath == "." {
			return nil
		}

		if d.IsDir() {
			return os.MkdirAll(relPath, 0755)
		}

		data, err := projectFiles.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(relPath, data, 0644)
	})
}

