DROP INDEX IF EXISTS idx_branches_company_coordinates;
DROP INDEX IF EXISTS idx_branches_company_region_city;
DROP INDEX IF EXISTS idx_branches_company_active;

ALTER TABLE branches
    DROP CONSTRAINT IF EXISTS branches_coordinates_pair_chk,
    DROP CONSTRAINT IF EXISTS branches_longitude_range_chk,
    DROP CONSTRAINT IF EXISTS branches_latitude_range_chk;

ALTER TABLE branches
    DROP COLUMN IF EXISTS opening_hours,
    DROP COLUMN IF EXISTS is_active,
    DROP COLUMN IF EXISTS region,
    DROP COLUMN IF EXISTS city,
    DROP COLUMN IF EXISTS longitude,
    DROP COLUMN IF EXISTS latitude;
