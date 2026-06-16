package stages

// FileRef carries a local filesystem path and its associated source URI.
// DirLoader emits FileRef entries and FileLoader consumes them.
type FileRef struct {
	Path   string // Local filesystem path to read the file
	Source string // Source URI (file://, https://github.com/..., etc.)
}
