-- name: ShowAuthor :one
SELECT *
FROM authors
WHERE id = $1
LIMIT 1;
-- name: CreateAuthor :one
INSERT INTO authors (name)
VALUES ($1)
RETURNING *;
-- name: IndexAuthors :many
SELECT *
FROM authors
ORDER BY created_at DESC;
