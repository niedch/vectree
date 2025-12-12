-- name: GetChunk :one
SELECT * FROM chunks
WHERE id = ? LIMIT 1;


-- name: CheckVersion :one
SELECT vec_version()
