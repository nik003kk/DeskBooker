-- DB-015: rollback the canonical users table.
-- WARNING: deleting this table permanently deletes all application users.
-- Intentionally no CASCADE: refuse rollback if future tables reference users.
DROP TABLE public.users;
