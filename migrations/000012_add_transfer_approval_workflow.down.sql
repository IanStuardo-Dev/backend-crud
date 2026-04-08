ALTER TABLE inventory_movements
    DROP CONSTRAINT IF EXISTS inventory_movements_movement_type_check;

ALTER TABLE inventory_movements
    ADD CONSTRAINT inventory_movements_movement_type_check
        CHECK (movement_type IN ('sale', 'transfer_out', 'transfer_in'));

DROP INDEX IF EXISTS idx_inventory_transfers_supervisor_status;

ALTER TABLE inventory_transfers
    DROP CONSTRAINT IF EXISTS inventory_transfers_status_check;

ALTER TABLE inventory_transfers
    ADD COLUMN completed_by_user_id BIGINT REFERENCES users(id),
    ADD COLUMN completed_at TIMESTAMPTZ;

UPDATE inventory_transfers
SET
    status = CASE
        WHEN status = 'received' THEN 'completed'
        ELSE 'completed'
    END,
    completed_by_user_id = COALESCE(received_by_user_id, dispatched_by_user_id, approved_by_user_id, supervisor_user_id, requested_by_user_id),
    completed_at = COALESCE(received_at, dispatched_at, approved_at, cancelled_at, created_at);

ALTER TABLE inventory_transfers
    DROP COLUMN supervisor_user_id,
    DROP COLUMN approved_by_user_id,
    DROP COLUMN dispatched_by_user_id,
    DROP COLUMN received_by_user_id,
    DROP COLUMN cancelled_by_user_id,
    DROP COLUMN approved_at,
    DROP COLUMN dispatched_at,
    DROP COLUMN received_at,
    DROP COLUMN cancelled_at;

ALTER TABLE inventory_transfers
    ADD CONSTRAINT inventory_transfers_status_check
        CHECK (status IN ('completed'));
