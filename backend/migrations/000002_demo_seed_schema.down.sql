-- DB-011: remove ONLY the isolated development fixtures and their schema.
-- WARNING: this deletes demo_seed data. Do not add production data here.
DROP SCHEMA demo_seed CASCADE;
