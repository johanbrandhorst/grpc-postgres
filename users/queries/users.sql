-- name: GetUser :one
SELECT * FROM users
WHERE id = $1;

-- name: AddUser :one
INSERT INTO users (
  role
) VALUES (
  $1
)
RETURNING *;

-- name: DeleteUser :one
DELETE FROM users
WHERE id = $1
RETURNING *;
