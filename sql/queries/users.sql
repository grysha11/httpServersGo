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

-- name: DeleteUsers :exec
DELETE FROM users;