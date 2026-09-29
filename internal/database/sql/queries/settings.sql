-- name: GetDepartmentsByNames :many
SELECT id, name FROM departments
WHERE tenant_id = sqlc.arg(tenant_id)
  AND name = ANY(sqlc.arg(names)::text[])
  AND is_active = TRUE;

-- name: GetPositionsByTitles :many
SELECT id, title FROM positions
WHERE tenant_id = sqlc.arg(tenant_id)
  AND title = ANY(sqlc.arg(titles)::text[])
  AND is_active = TRUE;

-- name: GetRolesByNames :many
SELECT id, name FROM roles
WHERE tenant_id = sqlc.arg(tenant_id)
  AND name = ANY(sqlc.arg(names)::text[]);

-- name: GetExistingEmails :many
SELECT email FROM users
WHERE tenant_id = sqlc.arg(tenant_id)
  AND email = ANY(sqlc.arg(emails)::text[])
  AND deleted_at IS NULL;

-- name: GetExistingUsernames :many
SELECT username FROM users
WHERE tenant_id = sqlc.arg(tenant_id)
  AND username = ANY(sqlc.arg(usernames)::text[])
  AND deleted_at IS NULL;

-- name: SetMustChangePassword :exec
UPDATE users
SET must_change_password = TRUE,
    last_password_change = NOW(),
    updated_at = NOW()
WHERE id = sqlc.arg(id) AND tenant_id = sqlc.arg(tenant_id);