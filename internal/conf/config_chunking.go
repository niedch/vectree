package conf

const (
	MDAST_STRATEGY  ChunkingStrategy = "mdast"
	HEADER_STRATEGY ChunkingStrategy = "header"
	LINE_STRATEGY   ChunkingStrategy = "line"
)

type ChunkingStrategy string

type Chunking struct {
	Strategy ChunkingStrategy `koanf:"strategy"`
}
