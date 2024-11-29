-- Drop trigger first
DROP TRIGGER IF EXISTS update_feature_activations_updated_at ON feature_activations;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_feature_activations_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_feature_activations_org_id;
DROP INDEX IF EXISTS idx_feature_activations_expires;

-- Drop table
DROP TABLE IF EXISTS feature_activations;
