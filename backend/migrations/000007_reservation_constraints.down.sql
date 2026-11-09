-- DB-019: revert only the indexes and exclusion constraint.
-- Retain btree_gist extension because other objects may depend on it.
ALTER TABLE public.reservations
    DROP CONSTRAINT reservations_no_active_overlap;

DROP INDEX public.reservations_check_in_code_lower_unique_idx;
DROP INDEX public.reservations_resource_starts_at_idx;
DROP INDEX public.reservations_user_starts_at_idx;
