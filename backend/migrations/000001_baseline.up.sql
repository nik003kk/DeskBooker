-- DB-010: Establish the initial migration version.
-- Business tables are intentionally deferred to future database model issues.
-- The migration tool creates and manages its own schema_migrations version table.
SELECT 1;
