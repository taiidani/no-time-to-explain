-- name: LoadFeeds :many
SELECT *
FROM feed
ORDER BY source, author;

-- name: CreateFeed :one
INSERT INTO feed (source, author, author_source_id, last_message)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateFeed :one
UPDATE feed SET
  source = $2,
  author = $3,
  author_source_id = $4,
  last_message = $5
WHERE id = $1
RETURNING *;

-- name: DeleteFeed :exec
DELETE FROM feed
WHERE id = $1;
