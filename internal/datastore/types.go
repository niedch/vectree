package datastore

type Document struct {
	Id       int    `db:"id"`
	Document string `db:"document"`
	Level    int    `db:"level"`
	ParentId *int   `db:"parent_id"` // Nullable parent reference
	Source   string `db:"source"`
}

type Embedding struct {
	Rowid     int64     `db:"rowid"`
	Embedding []float32 `db:"embedding"`
}

type DocumentEmbedding struct {
	DocumentId     int   `db:"document_id"`
	EmbeddingRowid int64 `db:"embedding_rowid"`
}

type DocumentWithEmbedding struct {
	Document
	Embedding      []float32 `db:"embedding"`
	EmbeddingRowid int64     `db:"embedding_rowid"`
}
