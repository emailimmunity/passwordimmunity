-- Create feature_activations table
CREATE TABLE IF NOT EXISTS feature_activations (
    feature_id VARCHAR(255) NOT NULL,
    organization_id VARCHAR(255) NOT NULL,
    activated_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (feature_id, organization_id)
);

-- Add indexes for common queries
CREATE INDEX IF NOT EXISTS idx_feature_activations_org_id ON feature_activations(organization_id);
CREATE INDEX IF NOT EXISTS idx_feature_activations_expires ON feature_activations(expires_at);

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
