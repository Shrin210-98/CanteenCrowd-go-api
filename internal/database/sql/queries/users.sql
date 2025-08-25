-- name: CreateUser :one
INSERT INTO users (username, email, password_hash, user_type) VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users 
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetUserByUsername :one
SELECT * FROM users 
WHERE username = $1 AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT * FROM users 
WHERE email = $1 AND deleted_at IS NULL;

-- name: UpdateUser :one
UPDATE users 
SET 
    username = $2,
    email = $3,
    last_login = $4,
    login_attempts = $5,
    is_locked = $6,
    locked_until = $7,
    password_reset_token = $8,
    token_expiry = $9,
    must_change_password = $10,
    email_verified = $11,
    mfa_enabled = $12,
    last_password_change = $13,
    user_type = $14,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdatePassword :one
UPDATE users 
SET 
    password_hash = $2,
    last_password_change = NOW(),
    must_change_password = $3,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateLoginAttempts :one
UPDATE users 
SET 
    login_attempts = $2,
    is_locked = $3,
    locked_until = $4,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SetPasswordResetToken :one
UPDATE users 
SET 
    password_reset_token = $2,
    token_expiry = $3,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: ClearPasswordResetToken :one
UPDATE users 
SET 
    password_reset_token = NULL,
    token_expiry = NULL,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: VerifyEmail :one
UPDATE users 
SET 
    email_verified = TRUE,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteUser :exec
UPDATE users 
SET 
    deleted_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users 
WHERE deleted_at IS NULL 
ORDER BY created_at DESC;

-- name: ListUsersByType :many
SELECT * FROM users 
WHERE user_type = $1 AND deleted_at IS NULL 
ORDER BY created_at DESC;

-- name: GetUserByResetToken :one
SELECT * FROM users 
WHERE password_reset_token = $1 AND deleted_at IS NULL;

-- name: CheckUsernameExists :one
SELECT EXISTS(
    SELECT 1 FROM users 
    WHERE username = $1 AND deleted_at IS NULL
);

-- name: CheckEmailExists :one
SELECT EXISTS(
    SELECT 1 FROM users 
    WHERE email = $1 AND deleted_at IS NULL
);