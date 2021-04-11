-- name: CreateUser :one
INSERT INTO users (
    username,
    fullname,
    email,
    hashed_password
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE username = $1 LIMIT 1;