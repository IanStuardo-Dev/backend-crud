CREATE TABLE branch_inventory (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT NOT NULL REFERENCES companies(id),
    branch_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    stock_on_hand INTEGER NOT NULL DEFAULT 0 CHECK (stock_on_hand >= 0),
    reserved_stock INTEGER NOT NULL DEFAULT 0 CHECK (reserved_stock >= 0 AND reserved_stock <= stock_on_hand),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT branch_inventory_company_branch_fkey
        FOREIGN KEY (company_id, branch_id) REFERENCES branches(company_id, id),
    CONSTRAINT branch_inventory_branch_product_key UNIQUE (branch_id, product_id)
);

INSERT INTO branch_inventory (company_id, branch_id, product_id, stock_on_hand, reserved_stock, created_at, updated_at)
SELECT p.company_id, p.branch_id, p.id, p.stock, 0, NOW(), NOW()
FROM products p
ON CONFLICT (branch_id, product_id) DO NOTHING;

CREATE INDEX idx_branch_inventory_company_branch
    ON branch_inventory (company_id, branch_id);

CREATE INDEX idx_branch_inventory_product
    ON branch_inventory (product_id);

CREATE TABLE inventory_transfers (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT NOT NULL REFERENCES companies(id),
    origin_branch_id BIGINT NOT NULL,
    destination_branch_id BIGINT NOT NULL,
    status VARCHAR(32) NOT NULL CHECK (status IN ('completed')),
    requested_by_user_id INTEGER NOT NULL REFERENCES users(id),
    completed_by_user_id INTEGER REFERENCES users(id),
    note TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    CONSTRAINT inventory_transfers_company_origin_fkey
        FOREIGN KEY (company_id, origin_branch_id) REFERENCES branches(company_id, id),
    CONSTRAINT inventory_transfers_company_destination_fkey
        FOREIGN KEY (company_id, destination_branch_id) REFERENCES branches(company_id, id),
    CONSTRAINT inventory_transfers_distinct_branches_chk
        CHECK (origin_branch_id <> destination_branch_id)
);

CREATE INDEX idx_inventory_transfers_company_created_at
    ON inventory_transfers (company_id, created_at DESC);

CREATE TABLE inventory_transfer_items (
    id BIGSERIAL PRIMARY KEY,
    transfer_id BIGINT NOT NULL REFERENCES inventory_transfers(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL REFERENCES products(id),
    quantity INTEGER NOT NULL CHECK (quantity > 0)
);

CREATE INDEX idx_inventory_transfer_items_transfer_id
    ON inventory_transfer_items (transfer_id);

ALTER TABLE inventory_movements
    ADD COLUMN IF NOT EXISTS transfer_id BIGINT REFERENCES inventory_transfers(id) ON DELETE SET NULL;

ALTER TABLE inventory_movements
    DROP CONSTRAINT IF EXISTS inventory_movements_movement_type_check;

ALTER TABLE inventory_movements
    ADD CONSTRAINT inventory_movements_movement_type_check
        CHECK (movement_type IN ('sale', 'transfer_out', 'transfer_in'));
