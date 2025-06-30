-- name: GetUserById :one
SELECT * FROM users
WHERE id = $1
LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1
LIMIT 1;


-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = $1
LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (
    username,
    display_name,
    email,
    password_hash
  )
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET display_name = COALESCE($2, display_name),
  email = COALESCE($3, email),
  email_verified = COALESCE($4, email_verified),
  password_hash = COALESCE($5, password_hash),
  last_login_at = COALESCE($6, last_login_at)
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: IndexUsers :many
SELECT * FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdatePasswordResetToken :one
UPDATE users
SET reset_password_token = $2,
  reset_password_expires = $3
WHERE id = $1
RETURNING *;

-- name: GetUserByResetToken :one
SELECT * FROM users
WHERE reset_password_token = $1
  AND reset_password_expires > NOW()
LIMIT 1;

-- name: UpdateLoginAttempts :one
UPDATE users
SET failed_login_attempts = $2,
  account_locked = $3
WHERE id = $1
RETURNING *;

-- name: UpdateLastLoginAt :one
UPDATE users
SET last_login_at = $2
WHERE id = $1
RETURNING *;

-- name: VerifyEmail :one
UPDATE users
SET email_verified = true
WHERE id = $1
RETURNING *;
