-- 000002_update_user_type.down.sql
BEGIN;

-- Reverse order: Drop new constraint first
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_user_type_check;

-- Revert data back
UPDATE users SET user_type = 'tenant' WHERE user_type = 'tenant_owner';

-- Restore old constraint
ALTER TABLE users ADD CONSTRAINT users_user_type_check 
    CHECK (user_type IN ('tenant', 'staff', 'guest'));

COMMIT;