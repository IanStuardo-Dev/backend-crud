DROP INDEX IF EXISTS idx_products_company_branch;
DROP INDEX IF EXISTS idx_products_branch_sku_unique;

ALTER TABLE products
    DROP CONSTRAINT IF EXISTS products_company_branch_fkey,
    DROP CONSTRAINT IF EXISTS products_company_id_fkey;

ALTER TABLE products
    DROP COLUMN IF EXISTS branch_id,
    DROP COLUMN IF EXISTS company_id;

ALTER TABLE products
    ADD CONSTRAINT products_sku_key UNIQUE (sku);
