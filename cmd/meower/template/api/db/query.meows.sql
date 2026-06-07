-- Meows are the worked-example resource. Reads LEFT JOIN users so the timeline
-- can show the author (older meows may have a NULL user_id, hence LEFT JOIN).
-- sqlc names the joined columns; the handler maps them onto the proto Meow.

-- name: ShowMeow :one
SELECT m.id, m.user_id, m.content, m.created_at,
  u.username AS author_username,
  u.display_name AS author_display_name
FROM meows m
  LEFT JOIN users u ON u.id = m.user_id
WHERE m.id = $1
LIMIT 1;

-- name: CreateMeow :one
INSERT INTO meows (user_id, content)
VALUES ($1, $2)
RETURNING id, user_id, content, created_at;

-- name: IndexMeows :many
SELECT m.id, m.user_id, m.content, m.created_at,
  u.username AS author_username,
  u.display_name AS author_display_name
FROM meows m
  LEFT JOIN users u ON u.id = m.user_id
ORDER BY m.created_at DESC;

-- name: UpdateMeow :one
UPDATE meows
SET content = $2
WHERE id = $1
RETURNING id, user_id, content, created_at;

-- name: DeleteMeow :exec
DELETE FROM meows
WHERE id = $1;
