CREATE TABLE sales (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT NOT NULL REFERENCES companies(id),
    branch_id BIGINT NOT NULL,
    created_by_user_id INTEGER NOT NULL REFERENCES users(id),
    total_amount_cents BIGINT NOT NULL CHECK (total_amount_cents >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT sales_company_branch_fkey
        FOREIGN KEY (company_id, branch_id) REFERENCES branches(company_id, id)
);

CREATE TABLE sale_items (
    id BIGSERIAL PRIMARY KEY,
    sale_id BIGINT NOT NULL REFERENCES sales(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL REFERENCES products(id),
    product_sku VARCHAR(64) NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price_cents BIGINT NOT NULL CHECK (unit_price_cents >= 0),
    subtotal_cents BIGINT NOT NULL CHECK (subtotal_cents >= 0)
);

CREATE INDEX idx_sale_items_sale_id ON sale_items (sale_id);

CREATE TABLE inventory_movements (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT NOT NULL REFERENCES companies(id),
    branch_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL REFERENCES products(id),
    sale_id BIGINT REFERENCES sales(id) ON DELETE SET NULL,
    movement_type VARCHAR(32) NOT NULL CHECK (movement_type IN ('sale')),
    quantity_delta INTEGER NOT NULL CHECK (quantity_delta <> 0),
    stock_after INTEGER NOT NULL CHECK (stock_after >= 0),
    created_by_user_id INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT inventory_movements_company_branch_fkey
        FOREIGN KEY (company_id, branch_id) REFERENCES branches(company_id, id)
);

CREATE INDEX idx_inventory_movements_product_created_at
    ON inventory_movements (product_id, created_at DESC);
