ALTER TABLE inventory_movements
    DROP CONSTRAINT IF EXISTS inventory_movements_movement_type_check;

ALTER TABLE inventory_movements
    ADD CONSTRAINT inventory_movements_movement_type_check
        CHECK (movement_type IN ('sale'));

ALTER TABLE inventory_movements
    DROP COLUMN IF EXISTS transfer_id;

DROP INDEX IF EXISTS idx_inventory_transfer_items_transfer_id;
DROP TABLE IF EXISTS inventory_transfer_items;

DROP INDEX IF EXISTS idx_inventory_transfers_company_created_at;
DROP TABLE IF EXISTS inventory_transfers;

DROP INDEX IF EXISTS idx_branch_inventory_product;
DROP INDEX IF EXISTS idx_branch_inventory_company_branch;
DROP TABLE IF EXISTS branch_inventory;
