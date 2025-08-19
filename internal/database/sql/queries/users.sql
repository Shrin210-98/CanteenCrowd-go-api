-- name: CreateUser :one
INSERT INTO users (
    employee_id, username, email, password_hash
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- name: UpdateUser :one
UPDATE users
SET 
    username = $2,
    email = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: IncrementLoginAttempts :exec
UPDATE users 
SET login_attempts = login_attempts + 1,
    is_locked = CASE WHEN login_attempts + 1 >= 5 THEN true ELSE false END,
    locked_until = CASE WHEN login_attempts + 1 >= 5 THEN NOW() + INTERVAL '30 minutes' ELSE NULL END
WHERE id = $1;

-- name: ResetLoginAttempts :exec
UPDATE users 
SET login_attempts = 0,
    is_locked = false,
    locked_until = NULL,
    last_login = NOW()
WHERE id = $1;

-- name: SetPasswordResetToken :exec
UPDATE users 
SET password_reset_token = $1,
    token_expiry = $2
WHERE email = $3;

-- name: UpdatePassword :exec
UPDATE users 
SET password_hash = $1,
    password_reset_token = NULL,
    token_expiry = NULL,
    must_change_password = $2,
    updated_at = NOW()
WHERE id = $3;