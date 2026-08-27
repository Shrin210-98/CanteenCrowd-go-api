-- ============ TENANT QUERIES ============

-- name: CreateTenant :one
INSERT INTO tenants (name, slug, plan, status, settings) 
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetTenantByID :one
SELECT * FROM tenants WHERE id = $1;

-- name: GetTenantBySlug :one
SELECT * FROM tenants WHERE slug = $1;

-- name: UpdateTenant :one
UPDATE tenants 
SET 
    name = $2,
    slug = $3,
    plan = $4,
    status = $5,
    settings = $6,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: SoftDeleteTenant :exec
UPDATE tenants SET 
    status = 'inactive',
    updated_at = NOW()
WHERE id = $1;

-- name: ListTenants :many
SELECT * FROM tenants WHERE status = 'active' 
ORDER BY created_at DESC;

-- name: ListAllTenants :many
SELECT * FROM tenants ORDER BY created_at DESC;

-- name: ListTenantsByPlan :many
SELECT * FROM tenants WHERE plan = $1 AND status = 'active'
ORDER BY created_at DESC;

-- name: CheckTenantSlugExists :one
SELECT EXISTS(SELECT 1 FROM tenants WHERE slug = $1);

-- name: CountTenants :one
SELECT COUNT(*) FROM tenants WHERE status = 'active';

-- name: GetTenantStats :one
SELECT 
    t.*,
    (SELECT COUNT(*) FROM users u WHERE u.tenant_id = t.id AND u.deleted_at IS NULL) as user_count,
    (SELECT COUNT(*) FROM roles r WHERE r.tenant_id = t.id) as role_count,
    (SELECT COUNT(*) FROM user_profiles up WHERE up.tenant_id = t.id) as profile_count
FROM tenants t
WHERE t.id = $1;

-- ============ USER QUERIES ============

-- name: CreateUser :one
INSERT INTO users (tenant_id, username, email, password_hash, user_type) 
VALUES ($1, $2, $3, $4, $5)
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

-- name: GetUserByTenantAndID :one
SELECT * FROM users 
WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL;

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
    tenant_id = $15,
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
WHERE id = $1 AND tenant_id = $2;

-- name: ListUsers :many
SELECT * FROM users 
WHERE deleted_at IS NULL 
ORDER BY created_at DESC;

-- name: ListUsersByTenant :many
SELECT * FROM users 
WHERE tenant_id = $1 AND deleted_at IS NULL 
ORDER BY created_at DESC;

-- name: ListUsersByType :many
SELECT * FROM users 
WHERE user_type = $1 AND deleted_at IS NULL 
ORDER BY created_at DESC;

-- name: ListUsersByTenantAndType :many
SELECT * FROM users 
WHERE tenant_id = $1 AND user_type = $2 AND deleted_at IS NULL 
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

-- name: CheckUsernameExistsInTenant :one
SELECT EXISTS(
    SELECT 1 FROM users 
    WHERE username = $1 AND tenant_id = $2 AND deleted_at IS NULL
);

-- name: CheckEmailExistsInTenant :one
SELECT EXISTS(
    SELECT 1 FROM users 
    WHERE email = $1 AND tenant_id = $2 AND deleted_at IS NULL
);

-- name: GetUsersByTenantID :many
SELECT u.* 
FROM users u
WHERE u.tenant_id = $1 AND u.deleted_at IS NULL
ORDER BY u.created_at DESC;

-- name: GetStaffUsersByTenant :many
SELECT u.* 
FROM users u
WHERE u.tenant_id = $1 
  AND u.user_type IN ('staff', 'guest')
  AND u.deleted_at IS NULL
ORDER BY u.created_at DESC;

-- name: CountUsersByTenant :one
SELECT COUNT(*) 
FROM users 
WHERE tenant_id = $1 AND deleted_at IS NULL;

-- name: CountActiveUsersByTenant :one
SELECT COUNT(*) 
FROM users 
WHERE tenant_id = $1 
  AND is_locked = FALSE 
  AND deleted_at IS NULL;

-- name: CountUsersByType :one
SELECT COUNT(*) 
FROM users 
WHERE user_type = $1 AND deleted_at IS NULL;

-- ============ USER PROFILE QUERIES ============

-- name: CreateUserProfile :one
INSERT INTO user_profiles (user_id, tenant_id, full_name, phone, avatar_url, timezone, permissions) 
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetUserProfile :one
SELECT * FROM user_profiles 
WHERE user_id = $1;

-- name: GetUserProfileByTenant :one
SELECT * FROM user_profiles 
WHERE user_id = $1 AND tenant_id = $2;

-- name: UpdateUserProfile :one
UPDATE user_profiles 
SET 
    full_name = $2,
    phone = $3,
    avatar_url = $4,
    timezone = $5,
    permissions = $6,
    updated_at = NOW()
WHERE user_id = $1
RETURNING *;

-- name: UpdateUserProfileByTenant :one
UPDATE user_profiles 
SET 
    full_name = $2,
    phone = $3,
    avatar_url = $4,
    timezone = $5,
    permissions = $6,
    updated_at = NOW()
WHERE user_id = $1 AND tenant_id = $7
RETURNING *;

-- name: DeleteUserProfile :exec
DELETE FROM user_profiles 
WHERE user_id = $1;

-- name: ListUserProfilesByTenant :many
SELECT * FROM user_profiles 
WHERE tenant_id = $1
ORDER BY created_at DESC;

-- name: CountProfilesByTenant :one
SELECT COUNT(*) 
FROM user_profiles 
WHERE tenant_id = $1;

-- ============ USER TOKEN QUERIES ============

-- name: CreateUserToken :one
INSERT INTO user_tokens (user_id, tenant_id, token, type, expires_at) 
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUserToken :one
SELECT * FROM user_tokens 
WHERE token = $1 AND used_at IS NULL AND expires_at > NOW();

-- name: GetUserTokenByType :one
SELECT * FROM user_tokens 
WHERE user_id = $1 AND type = $2 AND used_at IS NULL AND expires_at > NOW()
ORDER BY created_at DESC
LIMIT 1;

-- name: MarkTokenUsed :exec
UPDATE user_tokens 
SET used_at = NOW()
WHERE id = $1;

-- name: DeleteUserTokens :exec
DELETE FROM user_tokens 
WHERE user_id = $1;

-- name: DeleteUserTokensByType :exec
DELETE FROM user_tokens 
WHERE user_id = $1 AND type = $2;

-- name: DeleteExpiredTokens :exec
DELETE FROM user_tokens 
WHERE expires_at < NOW();

-- ============ REFRESH TOKEN QUERIES ============

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (user_id, tenant_id, token, expires_at) 
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens 
WHERE token = $1 AND expires_at > NOW();

-- name: DeleteRefreshToken :exec
DELETE FROM refresh_tokens 
WHERE token = $1;

-- name: DeleteUserRefreshTokens :exec
DELETE FROM refresh_tokens 
WHERE user_id = $1;

-- name: DeleteTenantRefreshTokens :exec
DELETE FROM refresh_tokens 
WHERE tenant_id = $1;

-- name: DeleteExpiredRefreshTokens :exec
DELETE FROM refresh_tokens 
WHERE expires_at < NOW();

-- ============ ROLE QUERIES ============

-- name: CreateRole :one
INSERT INTO roles (tenant_id, name, description, permission_template, is_default) 
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetRoleByID :one
SELECT * FROM roles 
WHERE id = $1;

-- name: GetRoleByTenantAndName :one
SELECT * FROM roles 
WHERE tenant_id = $1 AND name = $2;

-- name: UpdateRole :one
UPDATE roles 
SET 
    name = $2,
    description = $3,
    permission_template = $4,
    is_default = $5,
    updated_at = NOW()
WHERE id = $1 AND tenant_id = $6
RETURNING *;

-- name: DeleteRole :exec
DELETE FROM roles 
WHERE id = $1 AND tenant_id = $2;

-- name: ListRolesByTenant :many
SELECT * FROM roles 
WHERE tenant_id = $1
ORDER BY created_at DESC;

-- name: ListDefaultRolesByTenant :many
SELECT * FROM roles 
WHERE tenant_id = $1 AND is_default = TRUE
ORDER BY created_at DESC;

-- name: CountRolesByTenant :one
SELECT COUNT(*) 
FROM roles 
WHERE tenant_id = $1;

-- ============ JOIN QUERIES ============

-- name: GetUserWithProfile :one
SELECT 
    u.*,
    p.full_name,
    p.phone,
    p.avatar_url,
    p.timezone,
    p.permissions
FROM users u
LEFT JOIN user_profiles p ON u.id = p.user_id
WHERE u.id = $1 AND u.deleted_at IS NULL;

-- name: GetUserWithTenant :one
SELECT 
    u.*,
    t.name as tenant_name,
    t.slug as tenant_slug,
    t.plan as tenant_plan,
    t.status as tenant_status
FROM users u
JOIN tenants t ON u.tenant_id = t.id
WHERE u.id = $1 AND u.deleted_at IS NULL;

-- name: ListUsersWithProfilesByTenant :many
SELECT 
    u.*,
    p.full_name,
    p.phone,
    p.avatar_url,
    p.timezone
FROM users u
LEFT JOIN user_profiles p ON u.id = p.user_id
WHERE u.tenant_id = $1 AND u.deleted_at IS NULL
ORDER BY u.created_at DESC;

-- name: GetTenantWithStats :one
SELECT 
    t.*,
    COUNT(DISTINCT u.id) as total_users,
    COUNT(DISTINCT CASE WHEN u.deleted_at IS NULL THEN u.id END) as active_users,
    COUNT(DISTINCT r.id) as total_roles
FROM tenants t
LEFT JOIN users u ON u.tenant_id = t.id
LEFT JOIN roles r ON r.tenant_id = t.id
WHERE t.id = $1
GROUP BY t.id;

-- ============ USER MANAGEMENT QUERIES ============

-- name: ListUsersWithFilters :many
SELECT 
    u.*,
    e.id as employee_id,
    e.first_name,
    e.last_name,
    d.name as department_name,
    p.title as position_title
FROM users u
LEFT JOIN employees e ON u.id = e.user_id
LEFT JOIN departments d ON e.department_id = d.id
LEFT JOIN positions p ON e.position_id = p.id
WHERE 
    u.tenant_id = @tenant_id
AND 
    u.deleted_at IS NULL
AND
    (u.username ILIKE '%' || @search::text || '%' 
     OR u.email ILIKE '%' || @search::text || '%'
     OR e.first_name ILIKE '%' || @search::text || '%'
     OR e.last_name ILIKE '%' || @search::text || '%'
     OR @search::text = '')
AND
    (u.user_type = @user_type OR @user_type = '')
ORDER BY u.created_at DESC
LIMIT @limit_count OFFSET @offset_count;

-- name: CountUsersWithFilters :one
SELECT COUNT(*) FROM users u
LEFT JOIN employees e ON u.id = e.user_id
WHERE 
    u.tenant_id = @tenant_id
AND 
    u.deleted_at IS NULL
AND
    (u.username ILIKE '%' || @search::text || '%' 
     OR u.email ILIKE '%' || @search::text || '%'
     OR e.first_name ILIKE '%' || @search::text || '%'
     OR e.last_name ILIKE '%' || @search::text || '%'
     OR @search::text = '')
AND
    (u.user_type = @user_type OR @user_type = '');