package stages

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDirLoader(t *testing.T) {
	dirPath := "test_dir"
	loader := NewDirLoader(dirPath)
	assert.Equal(t, dirPath, loader.dirPath)
}

func TestDirLoader_Run(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_dir_loader")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	err = os.MkdirAll(filepath.Join(tempDir, "sub"), 0755)
	assert.NoError(t, err)
	_, err = os.Create(filepath.Join(tempDir, "file1.md"))
	assert.NoError(t, err)
	_, err = os.Create(filepath.Join(tempDir, "file2.txt"))
	assert.NoError(t, err)
	_, err = os.Create(filepath.Join(tempDir, "sub", "file3.md"))
	assert.NoError(t, err)

	loader := NewDirLoader(tempDir)
	in := make(chan any)
	out := loader.Run(context.Background(), in)

	var foundFiles []string
	for file := range out {
		foundFiles = append(foundFiles, file)
	}

	expectedFiles := []string{
		filepath.Join(tempDir, "file1.md"),
		filepath.Join(tempDir, "sub", "file3.md"),
	}

	// Sort slices for consistent comparison
	sort.Strings(foundFiles)
	sort.Strings(expectedFiles)

	assert.Equal(t, expectedFiles, foundFiles)
}

func TestDirLoader_Run_NonExistentDir(t *testing.T) {
	loader := NewDirLoader("non_existent_dir")
	in := make(chan any)
	out := loader.Run(context.Background(), in)

	var foundFiles []string
	for file := range out {
		foundFiles = append(foundFiles, file)
	}

	assert.Empty(t, foundFiles)
}

func TestDirLoader_Run_ContextCancellation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_dir_loader_cancel")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	_, err = os.Create(filepath.Join(tempDir, "file1.md"))
	assert.NoError(t, err)

	loader := NewDirLoader(tempDir)
	in := make(chan any)
	ctx, cancel := context.WithCancel(context.Background())
	out := loader.Run(ctx, in)

	// Cancel the context immediately
	cancel()

	var foundFiles []string
	for file := range out {
		foundFiles = append(foundFiles, file)
	}

	assert.Empty(t, foundFiles)
}
