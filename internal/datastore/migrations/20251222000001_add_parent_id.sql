-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

ALTER TABLE document ADD COLUMN parent_id INTEGER DEFAULT NULL;

-- Add foreign key constraint (self-referencing)
-- Note: SQLite doesn't enforce foreign keys by default, but we define it for documentation
CREATE INDEX IF NOT EXISTS idx_document_parent_id ON document(parent_id);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

-- SQLite doesn't support DROP COLUMN directly, so we need to recreate the table
CREATE TABLE document_backup (
  id INTEGER PRIMARY KEY,
  document TEXT,
  level INTEGER DEFAULT 0
);

INSERT INTO document_backup (id, document, level)
SELECT id, document, level FROM document;

DROP TABLE document;

ALTER TABLE document_backup RENAME TO document;
