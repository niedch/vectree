package setup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitProject(t *testing.T) {
	tmpDir := t.TempDir()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	if err := InitProject(); err != nil {
		t.Fatal(err)
	}

	expected := []string{
		"Dockerfile",
		"config.toml",
		filepath.Join("prompts", "documentation-help.prompt"),
		filepath.Join("prompts", "documentation-develop.prompt"),
	}

	for _, path := range expected {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected %s to exist: %v", path, err)
		}
	}

	if _, err := os.Stat("prompts"); err != nil {
		t.Errorf("expected prompts dir to exist: %v", err)
	}
}
