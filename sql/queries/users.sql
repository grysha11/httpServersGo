-- name: CreateUser :one
INSERT INTO users (email, hashed_password)
VALUES (
    $1,
    $2
)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: UpdateUserByID :one
UPDATE users
SET email = $2,
    hashed_password = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteUsers :exec
DELETE FROM users;