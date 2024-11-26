ALTER TABLE payments
ADD COLUMN license_type VARCHAR(50) NOT NULL DEFAULT 'free',
ADD COLUMN period VARCHAR(20) NOT NULL DEFAULT 'monthly';

-- Add indexes for common queries
CREATE INDEX idx_payments_license_type ON payments(license_type);
CREATE INDEX idx_payments_period ON payments(period);
