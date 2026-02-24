-- name: GetUserByEmail :one
SELECT id, hash_password
FROM users
WHERE email = $1
LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (
    email,
    hash_password
) VALUES (
    $1, $2
) RETURNING id;
