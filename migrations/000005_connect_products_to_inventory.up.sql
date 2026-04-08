ALTER TABLE products
    ADD COLUMN company_id BIGINT,
    ADD COLUMN branch_id BIGINT;

UPDATE products
SET company_id = 1,
    branch_id = 1
WHERE company_id IS NULL
   OR branch_id IS NULL;

ALTER TABLE products
    ALTER COLUMN company_id SET NOT NULL,
    ALTER COLUMN branch_id SET NOT NULL;

ALTER TABLE products
    DROP CONSTRAINT IF EXISTS products_sku_key;

ALTER TABLE products
    ADD CONSTRAINT products_company_id_fkey
        FOREIGN KEY (company_id) REFERENCES companies(id),
    ADD CONSTRAINT products_company_branch_fkey
        FOREIGN KEY (company_id, branch_id) REFERENCES branches(company_id, id);

CREATE UNIQUE INDEX idx_products_branch_sku_unique ON products (branch_id, sku);
CREATE INDEX idx_products_company_branch ON products (company_id, branch_id);
