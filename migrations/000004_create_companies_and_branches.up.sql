CREATE TABLE companies (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(150) NOT NULL,
    legal_name VARCHAR(255) NOT NULL DEFAULT '',
    tax_id VARCHAR(64) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_companies_tax_id_unique
    ON companies (tax_id)
    WHERE tax_id <> '';

CREATE TABLE branches (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    code VARCHAR(64) NOT NULL,
    name VARCHAR(150) NOT NULL,
    address TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT branches_company_code_key UNIQUE (company_id, code),
    CONSTRAINT branches_company_id_id_key UNIQUE (company_id, id)
);

INSERT INTO companies (id, name, legal_name, tax_id, created_at, updated_at)
VALUES (1, 'Default Company', 'Default Company', '', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

INSERT INTO branches (id, company_id, code, name, address, created_at, updated_at)
VALUES (1, 1, 'MAIN', 'Main Branch', '', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('companies', 'id'), GREATEST((SELECT MAX(id) FROM companies), 1));
SELECT setval(pg_get_serial_sequence('branches', 'id'), GREATEST((SELECT MAX(id) FROM branches), 1));
