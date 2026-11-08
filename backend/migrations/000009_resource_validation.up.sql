-- DB-035: enforce nonnegative floor for every resource in PostgreSQL.
ALTER TABLE public.resources
    ADD CONSTRAINT resources_floor_nonnegative_check
    CHECK (floor IS NULL OR floor >= 0);
