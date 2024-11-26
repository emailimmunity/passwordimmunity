-- Create payments table
CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    status VARCHAR(20) NOT NULL,
    payment_method VARCHAR(50) NOT NULL,
    mollie_id VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payments_mollie_id ON payments(mollie_id);
CREATE INDEX idx_payments_organization_id ON payments(organization_id);

-- Create licenses table
CREATE TABLE IF NOT EXISTS licenses (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL,
    type VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    valid_until TIMESTAMP WITH TIME ZONE NOT NULL,
    features TEXT[] DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);

CREATE INDEX idx_licenses_organization_id ON licenses(organization_id);
CREATE INDEX idx_licenses_valid_until ON licenses(valid_until);

-- Add license_id to payments table
ALTER TABLE payments ADD COLUMN license_id UUID REFERENCES licenses(id);
CREATE INDEX idx_payments_license_id ON payments(license_id);
