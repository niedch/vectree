-- name: GetDocument :one
SELECT * FROM document
WHERE id = ? LIMIT 1;

-- name: InsertDocument :one
INSERT INTO document (
  document
) VALUES (
  ?
)
RETURNING *;

-- name: ListDocuments :many
SELECT * FROM document
ORDER BY id;

-- name: InsertEmbedding :one
INSERT INTO embeddings (
  embedding
) VALUES (
  ?
)
RETURNING rowid;

-- name: GetEmbeddingWithDocument :one
SELECT e.rowid, e.embedding, e.document_id, d.document
FROM embeddings e
JOIN embeddings ON e.rowid = em.rowid
JOIN document d ON e.document_id = d.id
WHERE e.rowid = ? LIMIT 1;

-- name: SearchEmbedding :many
SELECT 
  rowid,
  distance
FROM embeddings
WHERE embedding MATCH ?
ORDER BY distance
LIMIT ?;
