DROP INDEX IF EXISTS idx_users_company_role;

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_role_valid_chk;

ALTER TABLE users
    DROP COLUMN IF EXISTS default_branch_id,
    DROP COLUMN IF EXISTS is_active,
    DROP COLUMN IF EXISTS role,
    DROP COLUMN IF EXISTS company_id;
