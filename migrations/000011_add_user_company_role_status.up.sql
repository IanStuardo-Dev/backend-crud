ALTER TABLE users
    ADD COLUMN IF NOT EXISTS company_id BIGINT REFERENCES companies(id),
    ADD COLUMN IF NOT EXISTS role VARCHAR(32),
    ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS default_branch_id BIGINT REFERENCES branches(id);

UPDATE users
SET company_id = COALESCE(company_id, 1),
    role = COALESCE(role, 'company_admin'),
    is_active = TRUE,
    default_branch_id = COALESCE(default_branch_id, 1)
WHERE role IS NULL
   OR company_id IS NULL
   OR default_branch_id IS NULL;

ALTER TABLE users
    ALTER COLUMN role SET NOT NULL;

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_role_valid_chk;

ALTER TABLE users
    ADD CONSTRAINT users_role_valid_chk
        CHECK (role IN ('super_admin', 'company_admin', 'inventory_manager', 'sales_user'));

CREATE INDEX IF NOT EXISTS idx_users_company_role
    ON users (company_id, role);
