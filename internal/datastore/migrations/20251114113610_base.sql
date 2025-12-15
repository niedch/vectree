-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS document (
  id INTEGER PRIMARY KEY,
  document TEXT
);

CREATE VIRTUAL TABLE IF NOT EXISTS embedding USING vec0(
  embedding FLOAT[768]
);

-- Mapping table to link documents to embeddings
CREATE TABLE IF NOT EXISTS document_embedding (
  document_id INTEGER NOT NULL,
  embedding_rowid INTEGER NOT NULL,
  PRIMARY KEY (document_id, embedding_rowid),
  FOREIGN KEY (document_id) REFERENCES document(id) ON DELETE CASCADE
);

-- Index for faster lookups by embedding_rowid
CREATE INDEX IF NOT EXISTS idx_embedding_rowid ON document_embedding(embedding_rowid);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

DROP INDEX IF EXISTS idx_embedding_rowid;
DROP TABLE IF EXISTS document_embedding;
DROP TABLE IF EXISTS embedding;
DROP TABLE IF EXISTS document;
