-- 000002_update_user_type.up.sql
BEGIN;

-- Step 1: Drop old constraint first
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_user_type_check;

-- Step 2: Update existing data
UPDATE users SET user_type = 'tenant_owner' WHERE user_type = 'tenant';

-- Step 3: Add new constraint
ALTER TABLE users ADD CONSTRAINT users_user_type_check 
    CHECK (user_type IN ('tenant_owner', 'staff', 'guest'));

COMMIT;