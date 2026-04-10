CREATE TABLE product_neighbor_feedback (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    branch_id BIGINT NOT NULL,
    source_product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    suggested_product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action VARCHAR(32) NOT NULL,
    note TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT product_neighbor_feedback_action_chk
        CHECK (action IN ('accepted', 'rejected', 'ignored')),
    CONSTRAINT product_neighbor_feedback_distinct_products_chk
        CHECK (source_product_id <> suggested_product_id),
    CONSTRAINT product_neighbor_feedback_branch_fk
        FOREIGN KEY (company_id, branch_id) REFERENCES branches(company_id, id) ON DELETE CASCADE,
    CONSTRAINT product_neighbor_feedback_unique_feedback
        UNIQUE (company_id, branch_id, user_id, source_product_id, suggested_product_id)
);

CREATE INDEX idx_product_neighbor_feedback_source_created_at
    ON product_neighbor_feedback (source_product_id, created_at DESC);

CREATE INDEX idx_product_neighbor_feedback_suggested_created_at
    ON product_neighbor_feedback (suggested_product_id, created_at DESC);

CREATE INDEX idx_product_neighbor_feedback_branch_action_created_at
    ON product_neighbor_feedback (branch_id, action, created_at DESC);
