-- Downgrade preserves reservations and restores their previous active status.
ALTER TABLE public.reservations DROP CONSTRAINT reservations_check_in_state_check;
UPDATE public.reservations SET status = 'CONFIRMED' WHERE status = 'CHECKED_IN';
ALTER TABLE public.reservations DROP CONSTRAINT reservations_no_active_overlap;
ALTER TABLE public.reservations ADD CONSTRAINT reservations_no_active_overlap
    EXCLUDE USING gist (
        resource_id WITH =,
        tstzrange(starts_at, ends_at, '[)') WITH &&
    ) WHERE (status IN ('PENDING', 'CONFIRMED'));
ALTER TABLE public.reservations DROP CONSTRAINT reservations_status_check;
ALTER TABLE public.reservations ADD CONSTRAINT reservations_status_check
    CHECK (status IN ('PENDING', 'CONFIRMED', 'CANCELLED', 'COMPLETED'));
ALTER TABLE public.reservations DROP COLUMN checked_in_at;
