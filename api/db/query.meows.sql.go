// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: query.meows.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createMeow = `-- name: CreateMeow :one
INSERT INTO meows (content)
VALUES ($1)
RETURNING id, user_id, content, created_at
`

func (q *Queries) CreateMeow(ctx context.Context, content string) (Meow, error) {
	row := q.db.QueryRow(ctx, createMeow, content)
	var i Meow
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Content,
		&i.CreatedAt,
	)
	return i, err
}

const indexMeows = `-- name: IndexMeows :many
SELECT id, user_id, content, created_at
FROM meows
ORDER BY created_at DESC
`

func (q *Queries) IndexMeows(ctx context.Context) ([]Meow, error) {
	rows, err := q.db.Query(ctx, indexMeows)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Meow
	for rows.Next() {
		var i Meow
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Content,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const showMeow = `-- name: ShowMeow :one
SELECT id, user_id, content, created_at
FROM meows
WHERE id = $1
LIMIT 1
`

func (q *Queries) ShowMeow(ctx context.Context, id pgtype.UUID) (Meow, error) {
	row := q.db.QueryRow(ctx, showMeow, id)
	var i Meow
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Content,
		&i.CreatedAt,
	)
	return i, err
}
