ALTER TABLE branches
    ADD COLUMN IF NOT EXISTS latitude DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS longitude DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS city VARCHAR(150) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS region VARCHAR(150) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS opening_hours JSONB NOT NULL DEFAULT '{}'::jsonb;

ALTER TABLE branches
    DROP CONSTRAINT IF EXISTS branches_latitude_range_chk,
    DROP CONSTRAINT IF EXISTS branches_longitude_range_chk,
    DROP CONSTRAINT IF EXISTS branches_coordinates_pair_chk;

ALTER TABLE branches
    ADD CONSTRAINT branches_latitude_range_chk
        CHECK (latitude IS NULL OR (latitude >= -90 AND latitude <= 90)),
    ADD CONSTRAINT branches_longitude_range_chk
        CHECK (longitude IS NULL OR (longitude >= -180 AND longitude <= 180)),
    ADD CONSTRAINT branches_coordinates_pair_chk
        CHECK (
            (latitude IS NULL AND longitude IS NULL)
            OR
            (latitude IS NOT NULL AND longitude IS NOT NULL)
        );

CREATE INDEX IF NOT EXISTS idx_branches_company_active
    ON branches (company_id, is_active);

CREATE INDEX IF NOT EXISTS idx_branches_company_region_city
    ON branches (company_id, region, city);

CREATE INDEX IF NOT EXISTS idx_branches_company_coordinates
    ON branches (company_id, latitude, longitude)
    WHERE latitude IS NOT NULL AND longitude IS NOT NULL;
