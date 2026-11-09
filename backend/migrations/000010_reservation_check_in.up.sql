-- Keep checked-in reservations occupied in the same atomic schema change.
ALTER TABLE public.reservations ADD COLUMN checked_in_at TIMESTAMPTZ;
ALTER TABLE public.reservations DROP CONSTRAINT reservations_status_check;
ALTER TABLE public.reservations ADD CONSTRAINT reservations_status_check
    CHECK (status IN ('PENDING', 'CONFIRMED', 'CHECKED_IN', 'CANCELLED', 'COMPLETED'));
ALTER TABLE public.reservations ADD CONSTRAINT reservations_check_in_state_check
    CHECK ((status <> 'CHECKED_IN' OR checked_in_at IS NOT NULL)
       AND (checked_in_at IS NULL OR status IN ('CHECKED_IN', 'COMPLETED')));
ALTER TABLE public.reservations DROP CONSTRAINT reservations_no_active_overlap;
ALTER TABLE public.reservations ADD CONSTRAINT reservations_no_active_overlap
    EXCLUDE USING gist (
        resource_id WITH =,
        tstzrange(starts_at, ends_at, '[)') WITH &&
    ) WHERE (status IN ('PENDING', 'CONFIRMED', 'CHECKED_IN'));
