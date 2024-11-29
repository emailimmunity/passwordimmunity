CREATE TABLE feature_activations (
    id SERIAL PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    feature_id VARCHAR(255),
    bundle_id VARCHAR(255),
    expires_at TIMESTAMP NOT NULL,
    payment_id VARCHAR(255) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT feature_or_bundle_required CHECK (
        (feature_id IS NOT NULL AND bundle_id IS NULL) OR
        (feature_id IS NULL AND bundle_id IS NOT NULL)
    ),
    CONSTRAINT valid_status CHECK (status IN ('active', 'expired', 'cancelled')),
    CONSTRAINT valid_currency CHECK (currency IN ('EUR', 'USD', 'GBP'))
);

CREATE INDEX idx_feature_activations_org_feature ON feature_activations(organization_id, feature_id);
CREATE INDEX idx_feature_activations_org_bundle ON feature_activations(organization_id, bundle_id);
CREATE INDEX idx_feature_activations_expires_at ON feature_activations(expires_at);

-- Add trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_feature_activations_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_feature_activations_updated_at
    BEFORE UPDATE ON feature_activations
    FOR EACH ROW
    EXECUTE FUNCTION update_feature_activations_updated_at();
