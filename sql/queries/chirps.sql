-- name: CreateChirp :one
INSERT INTO chirps (body, user_id)
VALUES (
    $1,
    $2
)
RETURNING *;

-- name: DeleteChirps :exec
DELETE FROM chirps;

-- name: GetAllChirps :many
SELECT * FROM chirps;

-- name: GetChirpByID :one
SELECT * FROM chirps
WHERE id = $1 LIMIT 1;