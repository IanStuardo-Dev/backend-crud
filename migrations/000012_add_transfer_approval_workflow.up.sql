ALTER TABLE inventory_transfers
    DROP CONSTRAINT IF EXISTS inventory_transfers_status_check;

ALTER TABLE inventory_transfers
    ADD COLUMN supervisor_user_id BIGINT REFERENCES users(id),
    ADD COLUMN approved_by_user_id BIGINT REFERENCES users(id),
    ADD COLUMN dispatched_by_user_id BIGINT REFERENCES users(id),
    ADD COLUMN received_by_user_id BIGINT REFERENCES users(id),
    ADD COLUMN cancelled_by_user_id BIGINT REFERENCES users(id),
    ADD COLUMN approved_at TIMESTAMPTZ,
    ADD COLUMN dispatched_at TIMESTAMPTZ,
    ADD COLUMN received_at TIMESTAMPTZ,
    ADD COLUMN cancelled_at TIMESTAMPTZ;

UPDATE inventory_transfers
SET
    status = 'received',
    supervisor_user_id = COALESCE(completed_by_user_id, requested_by_user_id),
    approved_by_user_id = completed_by_user_id,
    dispatched_by_user_id = completed_by_user_id,
    received_by_user_id = completed_by_user_id,
    approved_at = created_at,
    dispatched_at = created_at,
    received_at = COALESCE(completed_at, created_at)
WHERE status = 'completed';

ALTER TABLE inventory_transfers
    ALTER COLUMN supervisor_user_id SET NOT NULL;

ALTER TABLE inventory_transfers
    DROP COLUMN completed_by_user_id,
    DROP COLUMN completed_at;

ALTER TABLE inventory_transfers
    ADD CONSTRAINT inventory_transfers_status_check
        CHECK (status IN ('pending_approval', 'approved', 'in_transit', 'received', 'cancelled'));

CREATE INDEX idx_inventory_transfers_supervisor_status
    ON inventory_transfers (supervisor_user_id, status, created_at DESC);

ALTER TABLE inventory_movements
    DROP CONSTRAINT IF EXISTS inventory_movements_movement_type_check;

ALTER TABLE inventory_movements
    ADD CONSTRAINT inventory_movements_movement_type_check
        CHECK (movement_type IN ('sale', 'transfer_out', 'transfer_in', 'transfer_return'));
