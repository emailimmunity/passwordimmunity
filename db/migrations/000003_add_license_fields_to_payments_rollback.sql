DROP INDEX IF EXISTS idx_payments_period;
DROP INDEX IF EXISTS idx_payments_license_type;

ALTER TABLE payments
DROP COLUMN IF EXISTS period,
DROP COLUMN IF EXISTS license_type;
