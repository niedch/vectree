-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS document (
  id INTEGER PRIMARY KEY,
  document TEXT
);

-- Virtual table for vector embeddings using sqlite-vec
-- vec0 only supports a single embedding column
-- rowid is automatically created by SQLite for all tables
CREATE VIRTUAL TABLE IF NOT EXISTS embeddings USING vec0(
  embedding FLOAT[768]
);

-- Mapping table to link embeddings to documents
-- The rowid from embeddings table will be stored here
CREATE TABLE IF NOT EXISTS embeddings (
  rowid INTEGER PRIMARY KEY,
  document_id INTEGER NOT NULL,
  embedding BLOB,
  distance REAL,
  FOREIGN KEY (document_id) REFERENCES document(id)
);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

DROP TABLE IF EXISTS embeddings;
DROP TABLE IF EXISTS document;
