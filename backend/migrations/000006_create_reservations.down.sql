-- DB-018 rollback: deletes all reservations; preserves users and resources.
-- No CASCADE, so later dependent application tables are not deleted silently.
DROP TABLE public.reservations;
