-- Drop indexes first
DROP INDEX IF EXISTS idx_payments_license_id;
DROP INDEX IF EXISTS idx_licenses_valid_until;
DROP INDEX IF EXISTS idx_licenses_organization_id;
DROP INDEX IF EXISTS idx_payments_organization_id;
DROP INDEX IF EXISTS idx_payments_mollie_id;

-- Drop foreign key constraints
ALTER TABLE payments DROP COLUMN IF EXISTS license_id;

-- Drop tables
DROP TABLE IF EXISTS licenses;
DROP TABLE IF EXISTS payments;
