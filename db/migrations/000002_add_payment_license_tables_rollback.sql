-- Drop indexes
DROP INDEX IF EXISTS idx_licenses_expires_at;
DROP INDEX IF EXISTS idx_licenses_status;
DROP INDEX IF EXISTS idx_licenses_organization_id;
DROP INDEX IF EXISTS idx_payments_status;
DROP INDEX IF EXISTS idx_payments_organization_id;

-- Drop tables (in correct order due to foreign key constraints)
DROP TABLE IF EXISTS licenses;
DROP TABLE IF EXISTS payments;
