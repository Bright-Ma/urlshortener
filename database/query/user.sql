-- name: GetUserByEmail :one
SELECT id, username, password_hash, email
FROM users
WHERE email = $1;

-- name: IsEmailAvaliable :one
SELECT NOT EXISTS (
    SELECT 1 from users
    WHERE email = $1
);

-- name: CreateUser :one
INSERT INTO users (
    username,
    email,
    password_hash
) VALUES (
    $1, $2, $3
) RETURNING id,email;

-- name: UpdatePasswordByEmail :one
UPDATE users
SET password_hash = $1, updated_at = $2
WHERE email = $3
RETURNING id, username, email;
