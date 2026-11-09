CREATE TABLE public.reservation_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reservation_id UUID NOT NULL REFERENCES public.reservations(id),
    event_type TEXT NOT NULL CHECK (event_type = 'NO_SHOW_EXPIRED'),
    previous_status TEXT NOT NULL CHECK (previous_status = 'CONFIRMED'),
    new_status TEXT NOT NULL CHECK (new_status = 'EXPIRED'),
    occurred_at TIMESTAMPTZ NOT NULL,
    message TEXT NOT NULL,
    CONSTRAINT reservation_audit_log_event_unique UNIQUE (reservation_id, event_type)
);
CREATE INDEX reservation_audit_log_occurred_at_idx
    ON public.reservation_audit_log (occurred_at, id);
