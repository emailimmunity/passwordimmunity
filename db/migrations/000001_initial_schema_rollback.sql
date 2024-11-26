-- Rollback initial schema migration

-- Drop indexes
DROP INDEX IF EXISTS idx_user_organizations_organization_id;
DROP INDEX IF EXISTS idx_user_organizations_user_id;
DROP INDEX IF EXISTS idx_audit_logs_organization_id;
DROP INDEX IF EXISTS idx_audit_logs_user_id;
DROP INDEX IF EXISTS idx_vault_items_organization_id;
DROP INDEX IF EXISTS idx_vault_items_user_id;
DROP INDEX IF EXISTS idx_users_email;

-- Drop tables in reverse order of creation to handle dependencies
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS vault_items;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS user_organizations;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS organizations;
DROP TABLE IF EXISTS users;

-- Drop UUID extension
DROP EXTENSION IF EXISTS "uuid-ossp";
