-- name: ShowMeow :one
SELECT *
FROM meows
WHERE id = $1
LIMIT 1;
-- name: CreateMeow :one
INSERT INTO meows (content)
VALUES ($1)
RETURNING *;
-- name: IndexMeows :many
SELECT *
FROM meows
ORDER BY created_at DESC;
