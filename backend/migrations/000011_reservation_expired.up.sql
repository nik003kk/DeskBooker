-- EXPIRED is historical and is intentionally absent from the active-overlap
-- exclusion constraint introduced by migration 10.
ALTER TABLE public.reservations DROP CONSTRAINT reservations_status_check;
ALTER TABLE public.reservations ADD CONSTRAINT reservations_status_check
    CHECK (status IN ('PENDING', 'CONFIRMED', 'CHECKED_IN', 'EXPIRED', 'CANCELLED', 'COMPLETED'));
CREATE INDEX reservations_no_show_due_idx
    ON public.reservations (starts_at, id)
    WHERE status = 'CONFIRMED' AND checked_in_at IS NULL;
