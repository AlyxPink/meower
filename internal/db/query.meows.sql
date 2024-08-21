-- name: ShowMeow :one
SELECT *
FROM meows
WHERE id = $1
LIMIT 1;
-- name: CreateMeow :one
INSERT INTO meows (name)
VALUES ($1)
RETURNING *;
-- name: IndexMeows :many
SELECT *
FROM meows
ORDER BY created_at DESC;
