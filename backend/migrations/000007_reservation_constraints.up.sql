-- DB-019: enforce safe reservation relationships and lookup performance.
-- PostgreSQL btree_gist supplies GiST equality for UUID resource_id.
-- This extension is trusted and available on the PostgreSQL 17 image.
CREATE EXTENSION IF NOT EXISTS btree_gist;

-- User reservation history, newest-first and by time window.
CREATE INDEX reservations_user_starts_at_idx
    ON public.reservations (user_id, starts_at DESC);

-- Resource schedule and date-range queries.
CREATE INDEX reservations_resource_starts_at_idx
    ON public.reservations (resource_id, starts_at);

-- Check-in codes may be absent; present codes identify one reservation.
-- Case-insensitive matching avoids issuing equivalent codes.
CREATE UNIQUE INDEX reservations_check_in_code_lower_unique_idx
    ON public.reservations (LOWER(check_in_code))
    WHERE check_in_code IS NOT NULL;

-- [) means two reservations may meet at a boundary without overlapping.
-- Cancelled and completed reservations no longer block future bookings.
-- The database exclusion constraint is concurrency-safe, unlike a
-- check-then-insert in application code.
ALTER TABLE public.reservations
    ADD CONSTRAINT reservations_no_active_overlap
    EXCLUDE USING gist (
        resource_id WITH =,
        tstzrange(starts_at, ends_at, '[)') WITH &&
    )
    WHERE (status IN ('PENDING', 'CONFIRMED'));
