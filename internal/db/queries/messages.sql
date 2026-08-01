-- name: GetMessage :one
SELECT *
FROM message
WHERE id = $1 LIMIT 1;

-- name: LoadMessages :many
SELECT *
FROM message
ORDER BY trigger, response;

-- name: CreateMessage :one
INSERT INTO message (enabled, sender, trigger, response)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateMessage :one
UPDATE message SET
    enabled = $2,
    sender = $3,
    trigger = $4,
    response = $5
WHERE id = $1
RETURNING *;

-- name: DeleteMessage :exec
DELETE FROM message
WHERE id = $1;
