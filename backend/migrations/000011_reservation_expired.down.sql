DROP INDEX public.reservations_no_show_due_idx;
-- Preserve historical rows on downgrade using the previous nonblocking state.
UPDATE public.reservations SET status='CANCELLED' WHERE status='EXPIRED';
ALTER TABLE public.reservations DROP CONSTRAINT reservations_status_check;
ALTER TABLE public.reservations ADD CONSTRAINT reservations_status_check
    CHECK (status IN ('PENDING', 'CONFIRMED', 'CHECKED_IN', 'CANCELLED', 'COMPLETED'));
