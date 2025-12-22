-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

ALTER TABLE document ADD COLUMN level INTEGER DEFAULT 0;

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

-- SQLite doesn't support DROP COLUMN directly, so we need to recreate the table
CREATE TABLE document_backup (
  id INTEGER PRIMARY KEY,
  document TEXT
);

INSERT INTO document_backup (id, document)
SELECT id, document FROM document;

DROP TABLE document;

ALTER TABLE document_backup RENAME TO document;
