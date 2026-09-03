DROP TRIGGER IF EXISTS update_roles_updated_at ON roles;
DROP TRIGGER IF EXISTS update_user_profiles_updated_at ON user_profiles;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_tenants_updated_at ON tenants;

DROP INDEX IF EXISTS idx_refresh_tokens_tenant;
DROP INDEX IF EXISTS idx_user_tokens_token;
DROP INDEX IF EXISTS idx_user_tokens_tenant;
DROP INDEX IF EXISTS idx_user_profiles_tenant;
DROP INDEX IF EXISTS idx_users_role_id;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_tenant_type;
DROP INDEX IF EXISTS idx_users_tenant_id;
DROP INDEX IF EXISTS idx_roles_tenant;

DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS user_tokens;
DROP TABLE IF EXISTS user_profiles;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS tenants;

